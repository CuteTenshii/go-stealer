package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DiscordClient struct {
	Token string
}

type DiscordUser struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone"`
	Username    string  `json:"username"`
	GlobalName  *string `json:"global_name"`
	Bio         *string `json:"bio"`
	Avatar      *string `json:"avatar"`
	MFAEnabled  bool    `json:"mfa_enabled"`
	Flags       int     `json:"flags"`
	PremiumType int     `json:"premium_type"`
}

func apiRequest(method, url, token string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("https://discord.com/api/v9/%s", url), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status code %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *DiscordClient) UserInfo() (*DiscordAccountResult, error) {
	res, err := apiRequest("GET", "users/@me", c.Token, nil)
	if err != nil {
		return nil, err
	}

	var user DiscordUser
	err = json.Unmarshal(res, &user)
	if err != nil {
		return nil, err
	}

	account := &DiscordAccountResult{
		ID:         user.ID,
		Username:   user.Username,
		GlobalName: user.GlobalName,
		Email:      user.Email,
		Phone:      user.Phone,
		Bio:        user.Bio,
		Token:      c.Token,
		HasMFA:     user.MFAEnabled,
		Flags:      user.Flags,
		NitroType:  user.PremiumType,
	}
	if user.GlobalName != nil {
		account.GlobalName = user.GlobalName
	}
	if user.Avatar != nil {
		account.AvatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", user.ID, *user.Avatar)
	} else {
		account.AvatarURL = "https://cdn.discordapp.com/embed/avatars/0.png"
	}
	return account, nil
}
