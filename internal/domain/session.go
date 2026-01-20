package domain

import "time"

type Session struct {
	SessionID  string `json:"sessionId"`
	ServerCode string `json:"serverCode"`
	ClientIP   string `json:"ip"`
	StartTime  int64  `json:"startTime"`
	EndTime    int64  `json:"endTime,omitempty"`
	Connected  bool   `json:"connected"`
	PeerConfig string `json:"-"`
	ClientKey  string `json:"-"`
}

type VPNStatus struct {
	Connected bool   `json:"connected"`
	Server    string `json:"server,omitempty"`
	Duration  int64  `json:"duration,omitempty"`
	IP        string `json:"ip,omitempty"`
}

type SpeedTest struct {
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`
	Latency  int     `json:"latency"`
}

type ConnectRequest struct {
	ServerCode string `json:"serverCode"`
}

type DisconnectRequest struct {
	SessionID string `json:"sessionId"`
}

func NewSession(serverCode, clientIP, peerConfig, clientKey string) *Session {
	return &Session{
		SessionID:  generateSessionID(),
		ServerCode: serverCode,
		ClientIP:   clientIP,
		StartTime:  time.Now().Unix(),
		Connected:  true,
		PeerConfig: peerConfig,
		ClientKey:  clientKey,
	}
}

func generateSessionID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
