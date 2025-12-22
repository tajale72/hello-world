package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type IPInfoResponse struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Org      string `json:"org"`
	Loc      string `json:"loc"` // "lat,lng"
	Timezone string `json:"timezone"`
}

type GeoInfo struct {
	Country string
	Region  string
	City    string
	ISP     string
	Lat     string
	Lng     string
}

func LookupGeo(ip string) (GeoInfo, error) {
	token := "adf713f5b27ce8"
	if token == "" {
		return GeoInfo{}, fmt.Errorf("IPINFO_TOKEN not set")
	}
	req := http.Request{
		Header: make(http.Header),
		URL:    &url.URL{Scheme: "https", Host: "ipinfo.io", Path: fmt.Sprintf("/%s", ip)},
	}
	req.Header.Set("Authorization", "Bearer "+token)

	req.Method = http.MethodGet

	client := &http.Client{}
	resp, err := client.Do(&req)
	if err != nil {
		return GeoInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GeoInfo{}, fmt.Errorf("geo lookup failed: %s", resp.Status)
	}

	var r IPInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return GeoInfo{}, err
	}

	lat, lng := "", ""
	if r.Loc != "" {
		fmt.Sscanf(r.Loc, "%[^,],%s", &lat, &lng)
	}

	return GeoInfo{
		Country: r.Country,
		Region:  r.Region,
		City:    r.City,
		ISP:     r.Org,
		Lat:     lat,
		Lng:     lng,
	}, nil
}
