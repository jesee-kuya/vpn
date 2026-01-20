package handler

import (
	"p2nova-vpn/internal/service"
)

type Handler struct {
	vpnService    *service.VPNService
	serverService *service.ServerService
}

func NewHandler(vpnService *service.VPNService, serverService *service.ServerService) *Handler {
	return &Handler{
		vpnService:    vpnService,
		serverService: serverService,
	}
}
