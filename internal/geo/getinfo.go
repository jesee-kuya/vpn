package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GeoInfo struct {
	Country string `json:"country"` // ISO code, e.g., KE
	City    string `json:"city"`
	Region  string `json:"region"`
	IP      string `json:"ip"`
}

func GetServerGeo(ip string) (*GeoInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://ipinfo.io/%s/json", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info GeoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}
