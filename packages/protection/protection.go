package protection

import (
	"os"
	"strings"
	"time"

	"github.com/TheAldriguzDev/aldriguz_aio/packages/discord"
	"github.com/TheAldriguzDev/aldriguz_aio/packages/log"

	"github.com/mitchellh/go-ps"
)

// We should also add a temp ban to the key, maybe connecting the backend!

type BadApplications []string

var BadApp = BadApplications{
	"fiddler",
	"tcpview",
	"smsniff",
	"socketsniff",
	"charles",
	"mitmweb",
	"mitmdump",
	"burpsuite",
	"burp",
	"fiddle",
	"ghidra",
	"wireshark",
	"ilspy",
	"insomnia",
}

func GetProcess(licenseKey string, userName string) {
	for {
		processes, err := ps.Processes()
		if err != nil {
			panic(err)
		}

		for _, proc := range processes {
			name := proc.Executable()

			for _, appName := range BadApp {
				if strings.Contains(strings.ToLower(name), strings.ToLower(appName)) {
					log.Error("Crack attempt found!", "Protection")
					NotifyCrack(appName, licenseKey, userName)
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func NotifyCrack(appName string, userLicense string, userName string) {
	webhookURL := os.Getenv("ALDRIGUZ_ALERT_WEBHOOK")
	if webhookURL == "" {
		log.Warn("ALDRIGUZ_ALERT_WEBHOOK is not configured", "Protection")
		return
	}
	whData := discord.DiscordParams{
		Embeds: []discord.Embed{
			{
				Author: discord.Author{
					Name: "Aldriguz AIO",
					//URL:     "https://twitter.com/lollipop_ext",
				},
				Title: "Crack attemp found!",
				Color: 16776975,
				Footer: discord.Footer{
					Text: "Aldriguz AIO",
				},
				Fields: []discord.Field{
					{
						Name:   "Application",
						Value:  appName, // Add the discord user here
						Inline: true,
					},
					{
						Name:   "User name",
						Value:  userName, // Add the discord user here
						Inline: true,
					},
					{
						Name:   "License",
						Value:  userLicense, // Add the discord user here
						Inline: true,
					},
				},
			},
		},
	}

	err := discord.PostWebhook(webhookURL, whData)

	if err != nil {
		panic(err)
	} else {
		log.Success("Alert has been sent!", "Protection")
		os.Exit(0)
	}
}
