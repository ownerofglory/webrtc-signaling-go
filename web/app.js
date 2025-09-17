const ws = new WebSocket("wss://" + location.host + "/webrtc-signaling/ws");

let pc = null;
let dc = null;
let pendingOffer = null;

let remoteDescSet = false;
let candidateQueue = [];

async function getRTCConfig() {
    try {
        const res = await fetch("/webrtc-signaling/api/rtc-config");
        if (!res.ok) {
            throw new Error("Failed to fetch RTC config: " + res.status);
        }
        const cfg = await res.json();
        console.log("Fetched RTC config:", cfg);
        return cfg;
    } catch (err) {
        console.error("Error fetching RTC config:", err);
        return { iceServers: [{ urls: "stun:stun.l.google.com:19302" }] }; // fallback
    }
}

ws.onmessage = async (ev) => {
    const m = JSON.parse(ev.data);

    if (m.from && (m.signal === null || m.signal === undefined) && (!m.to || m.to === "")) {
        document.getElementById("myId").innerText = "Your ID: " + m.from;
        log("Your ID is " + m.from);
        return;
    }

    if (!m.signal) return;

    if (!pc) initPeerConnection();

    if (m.signal.type === "offer") {
        pendingOffer = m;
        showIncomingCallModal(m.from);
        log("Incoming call from " + m.from);
    }
    else if (m.signal.type === "answer") {
        hideCallingModal();
        console.log("Applying remote answer:", m.signal);
        await pc.setRemoteDescription({
            type: m.signal.type,
            sdp: m.signal.sdp
        });
        remoteDescSet = true;
        await flushCandidateQueue();
        log("Set remote answer");
    }
    else if (m.signal.candidate) {
        const candidate = {
            candidate: m.signal.candidate,
            sdpMid: m.signal.sdpMid,
            sdpMLineIndex: m.signal.sdpMLineIndex
        };
        if (remoteDescSet) {
            try {
                if (candidate.candidate && candidate.candidate !== "") {
                    await pc.addIceCandidate(candidate);
                    console.log("Applied ICE candidate immediately:", candidate);
                }
            } catch (err) {
                console.error("Error adding candidate:", err, candidate);
            }
        } else {
            candidateQueue.push(candidate);
            console.log("Queued candidate until remote description is set:", candidate);
        }
    }
};

async function initPeerConnection() {
    const rtcConfig = await getRTCConfig();
    pc = new RTCPeerConnection(rtcConfig);

    pc.onicecandidate = (e) => {
        const peerId = document.getElementById("peerId").value;
        if (e.candidate && peerId) {
            ws.send(JSON.stringify({
                signal: {
                    candidate: e.candidate.candidate,
                    sdpMid: e.candidate.sdpMid,
                    sdpMLineIndex: e.candidate.sdpMLineIndex
                },
                to: peerId
            }));
        }
    };

    pc.oniceconnectionstatechange = () => {
        console.log("ICE state:", pc.iceConnectionState);
        log("ICE state: " + pc.iceConnectionState);
    };

    pc.ontrack = (e) => {
        document.getElementById("remote").srcObject = e.streams[0];
    };

    pc.ondatachannel = (e) => {
        dc = e.channel;
        dc.onopen = () => log("Data channel open");
        dc.onmessage = (ev) => log("Peer: " + ev.data);
    };
}

async function startCall() {
    const peerId = document.getElementById("peerId").value.trim();
    if (!peerId) {
        alert("Enter a peer ID first");
        return;
    }

    if (!pc) initPeerConnection();

    const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
    document.getElementById("local").srcObject = stream;
    stream.getTracks().forEach(track => pc.addTrack(track, stream));

    dc = pc.createDataChannel("chat");
    dc.onopen = () => log("Data channel open");
    dc.onmessage = (ev) => log("Peer: " + ev.data);

    showCallingModal(peerId, "Dialing");

    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    await waitForIceGathering(pc);

    ws.send(JSON.stringify({
        signal: {
            type: pc.localDescription.type,
            sdp: pc.localDescription.sdp
        },
        to: peerId
    }));
    showCallingModal(peerId, "Connecting with ");
    log("Offer sent to " + peerId);
}

async function answerCall() {
    if (!pendingOffer) {
        alert("No incoming offer to answer");
        return;
    }

    const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
    document.getElementById("local").srcObject = stream;
    stream.getTracks().forEach(track => pc.addTrack(track, stream));

    console.log("Applying remote offer:", pendingOffer.signal);
    await pc.setRemoteDescription({
        type: pendingOffer.signal.type,
        sdp: pendingOffer.signal.sdp
    });
    remoteDescSet = true;

    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    await waitForIceGathering(pc);

    ws.send(JSON.stringify({
        signal: {
            type: pc.localDescription.type,
            sdp: pc.localDescription.sdp
        },
        to: pendingOffer.from
    }));
    log("Answered call from " + pendingOffer.from);

    await flushCandidateQueue();
    document.getElementById("answerBtn").disabled = true;
    pendingOffer = null;
}

async function flushCandidateQueue() {
    for (const c of candidateQueue) {
        try {
            if (c.candidate && c.candidate !== "") {
                await pc.addIceCandidate(c);
                console.log("Flushed queued candidate:", c);
            }
        } catch (err) {
            console.error("Error flushing candidate:", err, c);
        }
    }
    candidateQueue = [];
}

function waitForIceGathering(pc) {
    return new Promise(resolve => {
        if (pc.iceGatheringState === "complete") {
            resolve();
        } else {
            pc.onicegatheringstatechange = () => {
                if (pc.iceGatheringState === "complete") {
                    resolve();
                }
            };
        }
    });
}

function copyMyId() {
    const text = document.getElementById("myId").innerText.replace("Your ID: ", "");
    navigator.clipboard.writeText(text).then(() => {
        log("Copied your ID to clipboard: " + text);
    }).catch(err => {
        console.error("Failed to copy ID:", err);
    });
}

document.getElementById("acceptBtn").onclick = async () => {
    document.getElementById("incomingCallModal").classList.add("hidden");
    await answerCall();
};

document.getElementById("rejectBtn").onclick = () => {
    document.getElementById("incomingCallModal").classList.add("hidden");
    pendingOffer = null;
    log("Call rejected");
    // (Optional) Send a "reject" message to caller if you want them to know
};

document.getElementById("acceptBtn").onclick = async () => {
    hideIncomingCallModal();
    await answerCall();
};

document.getElementById("rejectBtn").onclick = () => {
    hideIncomingCallModal();
    pendingOffer = null;
    log("Call rejected");
};

document.getElementById("cancelCallBtn").onclick = () => {
    hideCallingModal();
    pendingOffer = null;
    pc && pc.close();
    pc = null;
    log("Call cancelled");
};

function showIncomingCallModal(callerId) {
    const modal = document.getElementById("incomingCallModal");
    const box = document.getElementById("incomingCallBox");

    document.getElementById("callerId").innerText = callerId;
    modal.classList.remove("hidden");

    requestAnimationFrame(() => {
        box.classList.remove("scale-95", "opacity-0");
        box.classList.add("scale-100", "opacity-100");
    });
}

function hideIncomingCallModal() {
    const modal = document.getElementById("incomingCallModal");
    const box = document.getElementById("incomingCallBox");

    box.classList.remove("scale-100", "opacity-100");
    box.classList.add("scale-95", "opacity-0");

    setTimeout(() => modal.classList.add("hidden"), 300);
}

function showCallingModal(calleeId, callStatus) {
    const modal = document.getElementById("callingModal");
    const box = document.getElementById("callingBox");

    document.getElementById("calleeId").innerText = calleeId;
    document.getElementById("callStatus").innerText = callStatus;
    modal.classList.remove("hidden");

    requestAnimationFrame(() => {
        box.classList.remove("scale-95", "opacity-0");
        box.classList.add("scale-100", "opacity-100");
    });
}

function hideCallingModal() {
    const modal = document.getElementById("callingModal");
    const box = document.getElementById("callingBox");

    box.classList.remove("scale-100", "opacity-100");
    box.classList.add("scale-95", "opacity-0");

    setTimeout(() => modal.classList.add("hidden"), 300);
}

function log(txt) {
    document.getElementById("log").innerHTML += "<p>" + txt + "</p>";
}