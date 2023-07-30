package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/TheAldriguzDev/tienne_aio/packages/encrypt"
	"github.com/TheAldriguzDev/tienne_aio/packages/log"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
)

var location string = "Whop"

type WhopData struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Hwid string `json:"hwid"`
}

func LoginClient(userKey string) (WhopResponse, error) {
	client := http.Client{}
	var userData WhopResponse

	secretKeyBytes, err := base64.StdEncoding.DecodeString(encrypt.SecretKey)
	if err != nil {
		panic(err)
	}

	secret_bearer := "X9l16e83DgkrysoBRdGdI5O04VIpHMOMjuStCYe4HvIxbrKeWEhLt2nQW8km/VD6ol/huDvdfjFpG0KjZcpglA=="
	encryptedData, err := base64.StdEncoding.DecodeString(secret_bearer)
	if err != nil {
		panic(err)
	}

	// Decifra i dati
	decryptedData, err := encrypt.DecryptAES(secretKeyBytes, encryptedData)
	if err != nil {
		panic(err)
	}

	// Converte il byte slice decifrato in una stringa
	bearer := string(decryptedData)

	if userKey == "" {
		return userData, fmt.Errorf("No user key has been found!")
	}

	whopPayload := WhopData{
		Metadata: Metadata{
			Hwid: GenerateHWID(),
		},
	}

	jsonData, err := json.Marshal(whopPayload)

	if err != nil {
		return userData, err
	}

	loginUrl := fmt.Sprintf("https://api.whop.com/api/v2/memberships/%s/validate_license", userKey)
	req, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		return userData, err
	}

	req.Header = http.Header{
		"Authorization": {`Bearer ` + bearer},
	}

	res, err := client.Do(req)

	if err != nil {
		return userData, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return userData, err
	}

	if res.StatusCode != 201 && res.StatusCode == 400 {
		return userData, fmt.Errorf("This license key is already binded! Please reset your data!")
	}

	if res.StatusCode != 201 {
		return userData, fmt.Errorf("An error occured while checking the license key! [INVALID]")
	}

	err = json.Unmarshal(body, &userData)

	if err != nil {
		return userData, err
	}

	if userData.Valid != true {
		return userData, fmt.Errorf("This license key is not valid!")
	}

	return userData, nil
}

func GenerateHWID() string {
	id, err := machineid.ID()
	if err != nil {
		log.Error("Error while getting the machine ID!", location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
		return ""
	}

	return id
}
