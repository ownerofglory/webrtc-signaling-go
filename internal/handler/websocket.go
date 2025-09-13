package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ownerofglory/webrtc-signaling-go/config"
	"log/slog"
	"net/http"
	"sync"
)

const WSPath = basePathWS

type (
	wsHandler struct {
		upgrader *websocket.Upgrader
	}

	webRTCClientID string

	webRTCSignalingMessageType string
	webrtcSignalingMessageSDP  string

	webRTCSignalingMessage struct {
		MessageType webRTCSignalingMessageType `json:"type"`
		SDP         webrtcSignalingMessageSDP  `json:"sdp"`
	}

	webRTCClientMessage struct {
		SignalingMessage *webRTCSignalingMessage `json:"signal"`
		ReceiverPeerID   webRTCClientID          `json:"to"`
		OriginPeerID     webRTCClientID          `json:"from"`
	}

	webRTCClientConn struct {
		conn          *websocket.Conn
		id            webRTCClientID
		connCloseOnce sync.Once
		readCh        chan *webRTCClientMessage
		writeCh       chan *webRTCClientMessage
	}
)

const (
	webrtcOffer     webRTCSignalingMessageType = "offer"
	webrtcAnswer    webRTCSignalingMessageType = "answer"
	webrtcCandidate webRTCSignalingMessageType = "candidate"
)

var (
	webRTCConnections = make(map[webRTCClientID]*webRTCClientConn)
	connMx            sync.RWMutex
)

func NewWSHandler(conf *config.WebRTCSignalingAppConfig) *wsHandler {
	return &wsHandler{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				//u, err := url.Parse(r.Header.Get("Origin"))
				//if err != nil {
				//	return false
				//}
				//return u.Host == conf.AllowedOrigin
				return true
			},
		},
	}
}

func (h *wsHandler) HandleWS(rw http.ResponseWriter, req *http.Request) {
	conn, err := h.upgrader.Upgrade(rw, req, nil)
	if err != nil {
		slog.Error("Error when upgrading to websocket", "err", err.Error())
		return
	}

	clientUUID, err := uuid.NewV7()
	if err != nil {
		slog.Error("Error when generate client ID", "err", err.Error())
		return
	}

	clientID := webRTCClientID(clientUUID.String())
	clientConn := &webRTCClientConn{
		conn:    conn,
		id:      clientID,
		readCh:  make(chan *webRTCClientMessage),
		writeCh: make(chan *webRTCClientMessage),
	}

	connMx.Lock()
	webRTCConnections[clientID] = clientConn
	connMx.Unlock()

	slog.Info("Client connected", "clientID", clientID)

	clientConn.serve()
}

func (c *webRTCClientConn) serve() {
	go c.readRoutine()
	go c.writeRoutine()
	go c.processMessage()
}

func (c *webRTCClientConn) close() {
	c.connCloseOnce.Do(func() {
		connMx.Lock()
		defer connMx.Unlock()

		slog.Debug("Closing WebRTC client", "clientId", c.id)

		close(c.readCh)
		close(c.writeCh)
		err := c.conn.Close()
		if err != nil {
			slog.Error("Error when close websocket connection", "err", err.Error())
			return
		}

		delete(webRTCConnections, c.id)
	})
}

func (c *webRTCClientConn) readRoutine() {
	defer c.close()
	for {
		mType, payload, err := c.conn.ReadMessage()
		if err != nil {
			slog.Error("Error when reading websocket message", "err", err.Error())
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				break
			}
			continue
		}

		if mType == websocket.TextMessage {
			var m webRTCClientMessage
			err := json.Unmarshal(payload, &m)
			if err != nil {
				slog.Error("Error when unmarshalling message", "err", err.Error())
				continue
			}
			c.readCh <- &m
		}
	}
}

func (c *webRTCClientConn) writeRoutine() {
	defer c.close()
	for m := range c.writeCh {
		payload, err := json.Marshal(m)
		if err != nil {
			slog.Error("Error when marshalling message", "err", err.Error())
			break
		}

		err = c.conn.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			slog.Error("Error when sending message", "err", err.Error())
			return
		}
	}
}

func (c *webRTCClientConn) processMessage() {
	for m := range c.readCh {
		func() {
			connMx.RLock()
			defer connMx.RUnlock()

			recepient := webRTCConnections[m.ReceiverPeerID]
			if recepient == nil {
				slog.Error("Error when receiving message from client", "clientId", c.id)
				return
			}

			m.OriginPeerID = c.id

			recepient.writeCh <- m
		}()

	}
}
