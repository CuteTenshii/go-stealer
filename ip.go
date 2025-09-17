package main

import (
	"encoding/json"
	"net/http"
)

type IPInfo struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func GetIPInfo() (*IPInfo, error) {
	res, err := http.Get("https://ipinfo.io/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var ipInfo IPInfo
	err = json.NewDecoder(res.Body).Decode(&ipInfo)

	if err != nil {
		return nil, err
	}
	return &ipInfo, nil
}
