package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ownerofglory/webrtc-signaling-go/config"
	"io"
	"log/slog"
	"net/http"
)

const (
	GetRTCConfigPath = basePath + "/rtc-config"

	CloudFlareURL = "https://rtc.live.cloudflare.com/v1/turn/keys"
	DefaultTTL    = 60 * 60
)

type CloudFlareConfigRequest struct {
	TTL int `json:"ttl"`
}

type CloudFlareConfig struct {
	ICEServers []struct {
		URLs       []string `json:"urls"`
		Username   string   `json:"username,omitempty"`
		Credential string   `json:"credential,omitempty"`
	} `json:"iceServers"`
	TTL *int `json:"ttl"`
}

type rtcConfigHandler struct {
	turnKey      string
	turnAPIToken string
	client       *http.Client
}

func NewRTCConfigHandler(cfg *config.WebRTCSignalingAppConfig) *rtcConfigHandler {
	return &rtcConfigHandler{
		turnKey:      cfg.TURNKey,
		turnAPIToken: cfg.TURNAPIToken,
		client:       &http.Client{},
	}
}

func (h *rtcConfigHandler) HandleGetRTCConfig(rw http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/%s/credentials/generate-ice-servers", CloudFlareURL, h.turnKey)
	reqBody := &CloudFlareConfigRequest{
		TTL: DefaultTTL,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		slog.Error("Error marshalling request body", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	clientReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		slog.Error("Error creating request", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	clientReq.Header.Add("Authorization", "Bearer "+h.turnAPIToken)
	clientReq.Header.Add("Content-Type", "application/json")

	res, err := h.client.Do(clientReq)
	if err != nil {
		slog.Error("Error getting rtc config", "error", err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	respPayload, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Error reading response body", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
	}

	cloudFlareConfig := CloudFlareConfig{}
	err = json.Unmarshal(respPayload, &cloudFlareConfig)
	if err != nil {
		slog.Error("Error unmarshalling response body", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	cloudFlareConfig.TTL = &reqBody.TTL

	respPayload, err = json.Marshal(cloudFlareConfig)
	if err != nil {
		slog.Error("Error marshalling response body", "error", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write(respPayload)
}
