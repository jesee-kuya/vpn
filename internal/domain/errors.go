package domain

import "errors"

var (
	ErrServerNotFound   = errors.New("server not found")
	ErrSessionNotFound  = errors.New("session not found")
	ErrInvalidRequest   = errors.New("invalid request")
	ErrWireGuardFailed  = errors.New("wireguard operation failed")
	ErrNotConnected     = errors.New("not connected")
	ErrAlreadyConnected = errors.New("already connected")
)
