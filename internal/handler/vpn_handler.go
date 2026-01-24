package handler

import (
	"net/http"

	"p2nova-vpn/internal/domain"
)

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	var req domain.ConnectRequest
	if err := DecodeJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request")
		return
	}

	session, err := h.vpnService.Connect(req.ServerCode)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, http.StatusOK, map[string]interface{}{
		"sessionId": session.SessionID,
		"ip":        session.ClientIP,
		"startTime": session.StartTime,
		"config":    session.PeerConfig,
	})
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	var req domain.DisconnectRequest
	if err := DecodeJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.vpnService.Disconnect(req.SessionID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.vpnService.GetStatus()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, http.StatusOK, status)
}

func (h *Handler) GetSpeed(w http.ResponseWriter, r *http.Request) {
	speed := h.vpnService.GetSpeed()
	SuccessResponse(w, http.StatusOK, speed)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	SuccessResponse(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"app":    "p2Nova VPN",
	})
}
