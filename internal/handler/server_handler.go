package handler

import (
	"net/http"
)

func (h *Handler) GetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.serverService.ListServers()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(w, http.StatusOK, servers)
}

func (h *Handler) SelectServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerCode string `json:"serverCode"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Invalid request")
		return
	}

	server, err := h.serverService.GetServer(req.ServerCode)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	SuccessResponse(w, http.StatusOK, server)
}
