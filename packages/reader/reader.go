package filereader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/TheAldriguzDev/aldriguz_aio/packages/log"
)

type Folders []string

var folders = Folders{"proxies", "accounts", "orders", "tasks", "data", "settings"}

type Files []string

var files = Folders{"settings"}

type SettingsData struct {
	License   string `json:"license"`
	TwoCap    string `json:"2Captcha"`
	SuccessWh string `json:"success_webhook"`
	FailedWh  string `json:"failed_webhook"`
}

var location string = "File Manager"

func GetDir() bool {
	for _, folderName := range folders {
		folder, err := os.Open("./" + folderName)
		if err != nil {
			log.Error("Folder: "+folderName+" not found!", "Reader")

			folderCreation := CreateFolder(folderName)
			if folderCreation {
				log.Success("Folder: "+folderName+" has been created!", "Reader")
			} else {
				log.Error("Error while creating "+folderName+" folder!", "Reader")
			}
		}
		defer folder.Close()
	}

	return true
}

func CreateFolder(folderName string) bool {
	err := os.Mkdir(folderName, 0755)
	if err != nil {
		return false
	}

	return true
}

func GetFiles() error {
	if _, err := os.Stat("./settings/settings.json"); err != nil {
		log.Error("Cannot find settings.json file!", "Reader")
		err := CreateFile()
		time.Sleep(2 * time.Second)
		if err != nil {
			return err
		} else {
			return nil
		}
	}

	return nil
}

func CreateFile() error {
	settData := SettingsData{
		License:   "",
		TwoCap:    "",
		SuccessWh: "",
		FailedWh:  "",
	}

	jsonData, err := json.MarshalIndent(settData, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./settings/settings.json", jsonData, 0644)
	if err != nil {
		return err
	} else {
		log.Success("Settings.json file has been created, please fill it!", "Reader")
		time.Sleep(10 * time.Second)
		os.Exit(0)
		return nil
	}
}

func ReadSettings() (SettingsData, error) {
	settingsData := SettingsData{}

	content, err := ioutil.ReadFile("./settings/settings.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &settingsData)
	if err != nil {
		panic(err)
	}

	if settingsData.License == "" {
		return settingsData, fmt.Errorf("License expected but received null")
	}

	return settingsData, nil
}

func ReadFolder(path string, fileType string) (string, error) {
	folder, err := os.Open("./" + path)

	if err != nil {
		return "", nil
	}

	defer folder.Close()

	fileList, err := folder.ReadDir(0)

	if err != nil {
		return "", err
	}

	var filesName []string

	for _, file := range fileList {
		filesName = append(filesName, file.Name())
	}

	if len(filesName) == 0 {
		log.Error("No files has been found! Closing the bot...", "Reader")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}

	var selection int

	for {
		log.Log(">>> Please select a "+fileType+" file", log.Colors.DarkGrey, "Reader")
		counter := 0

		for _, elem := range filesName {
			log.Log(">>> "+strconv.Itoa(counter)+" - "+elem, log.Colors.DarkGrey, "Reader")
			counter++
		}

		fmt.Printf(log.Colors.DarkGrey+"\r[%s] [%s] >>> "+log.Colors.Reset, location, log.GetTime())

		if _, err := fmt.Scan(&selection); err != nil || selection <= -1 || selection > len(filesName) {
			log.Error("Wrong input, please retry!", "Reader")
			time.Sleep(3 * time.Second)
			log.ClearConsole()
			log.PrintLogo()
			log.Info("Aldriguz AIO files manager", "Reader")
			fmt.Println("")
			continue
		}
		break
	}

	return filesName[selection], nil
}
