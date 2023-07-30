package main

import (
	"fmt"
	"os"
	"time"

	"github.com/TheAldriguzDev/tienne_aio/modules/queue"
	"github.com/TheAldriguzDev/tienne_aio/packages/auth"
	"github.com/TheAldriguzDev/tienne_aio/packages/log"
	"github.com/TheAldriguzDev/tienne_aio/packages/modules"
	filereader "github.com/TheAldriguzDev/tienne_aio/packages/reader"
)

var location = "Dashboard"

var errorTimeOut = time.Duration(10)

func main() {
	log.PrintLogo()

	folderCheck := filereader.GetDir()

	if !folderCheck {
		log.Error("Error while checking the folders!", location)
		time.Sleep(errorTimeOut * time.Second)
		os.Exit(0)
	}

	settingsErr := filereader.GetFiles()

	if settingsErr != nil {
		log.Error("No settings.json has been found!", location)
		time.Sleep(errorTimeOut * time.Second)
		os.Exit(0)
	}

	settingData, settingsErr := filereader.ReadSettings()

	if settingsErr != nil {
		log.Error("No license key has been found!", location)
		time.Sleep(errorTimeOut * time.Second)
		os.Exit(0)
	}

	/*
		When we are here we have checked that:
			- All the folders exist
			- The settings.json file exist
			- The user license key exist

		So we can proceed and start the bot
	*/

	userData, err := auth.LoginClient(settingData.License)

	// We start checking if the user want to spoof our request
	/*
		REMEMBER TO ENABLE IT
	*/
	//go protection.GetProcess(settingData.License, userData.Discord.Username)

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(errorTimeOut * time.Second)
		os.Exit(0)
	} else {
		log.Success("Successfully logged in!", location)
	}

	// Now we can start the bot!

	//log.ClearConsole()
	//log.PrintLogo()

	fmt.Println("")
	log.Info("Welcome to Tienne AIO, "+userData.Discord.Username, location)

	// Now we have to select the site

	choosenSite := modules.SiteSelector()

	log.ClearConsole()
	log.PrintLogo()
	fmt.Println("")

	switch choosenSite {
	case "Queue IT":
		log.Info("Queue IT module", location)

		taskFile, err := filereader.ReadFolder("tasks", "tasks")
		if err != nil {
			panic(err)
		}

		queue.InitQueue(taskFile, settingData.SuccessWh, settingData.FailedWh)
		//skatedeluxe.InitialiseTasks(taskFile, settingData.SuccessWh, settingData.FailedWh)
		break
	case "Footlocker":
		log.Info("Footlocker EU module", location)

		// taskFile, err := filereader.ReadFolder("tasks", "tasks")
		// if err != nil {
		// 	panic(err)
		// }

		//footlocker.InitTasks(taskFile, settingData.SuccessWh, settingData.FailedWh)
		break
	case "Nike Order":
		log.Info("Nike order tracker", location)
		break
	default:
		log.Error("The selected module does not exist", location)
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}
}

func getCurrentTimeWithMilliseconds() string {
	now := time.Now()
	milliseconds := now.Nanosecond() / 1e6
	return fmt.Sprintf("%s.%03d", now.Format("2006-01-02 15:04:05"), milliseconds)
}
