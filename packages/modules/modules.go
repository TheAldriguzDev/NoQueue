package modules

import (
	"fmt"
	"os"
	"time"

	"github.com/TheAldriguzDev/tienne_aio/packages/log"
)

type AvailableSites []string

var Sites = AvailableSites{"Queue IT", "Footlocker", "Nike Order"}

var location string = "Modules helper"

func SiteSelector() string {
	var selection int
	for {
		log.ClearConsole()
		log.PrintLogo()

		fmt.Println("")

		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] Please select an action: \n"+log.Colors.Reset, location, log.GetTime())
		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] >>> 1 - Queue IT \n"+log.Colors.Reset, location, log.GetTime())
		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] >>> 2 - Footlocker EU \n"+log.Colors.Reset, location, log.GetTime())
		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] >>> 3 - Nike Order tracker \n"+log.Colors.Reset, location, log.GetTime())
		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] >>> 0 - Exit \n"+log.Colors.Reset, location, log.GetTime())

		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] >>> "+log.Colors.Reset, location, log.GetTime())

		if _, err := fmt.Scan(&selection); err != nil || selection < 0 || selection > 4 {
			fmt.Printf(log.Colors.Red+"\r[%s] [%s] Wrong input, please retry! \n"+log.Colors.Reset, location, log.GetTime())
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	if selection != 0 {
		return Sites[selection-1]
	} else {
		os.Exit(0)
		return ""
	}
}
