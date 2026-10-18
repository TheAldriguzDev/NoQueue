package captchasolver

import (
	"time"

	"github.com/TheAldriguzDev/aldriguz_aio/packages/log"

	api2captcha "github.com/2captcha/2captcha-go"
)

var location string = "Captcha Solver"

// 2captcha solver
func Create2CapClient(apiKey string) *api2captcha.Client {
	client := api2captcha.NewClient(apiKey)

	return client
}

type RecapStruct struct {
	SiteKey string
	Url     string
}

func SolveQueueItRecap(recapInfo RecapStruct, client *api2captcha.Client) (string, error) {
	captcha := api2captcha.ReCaptcha{
		SiteKey:   recapInfo.SiteKey,
		Url:       recapInfo.Url,
		Invisible: true,
		Action:    "verify",
	}

	id, err := client.Send(captcha.ToRequest())
	if err != nil {
		return "", err
	}

	var capStatus bool = false
	var capResult string = ""

	for !capStatus {
		code, err := client.GetResult(id)
		if err != nil {
			log.Error(err.Error(), location)
			return "", err
		}

		if code == nil {
			log.General("Waiting for captcha...", location)
			time.Sleep(5 * time.Second)
			continue
		}

		capResult = *code
		capStatus = true
	}

	return capResult, nil
}

// func SolveQueueItRecap2(recapInfo RecapStruct, client *api2captcha.Client) (string, error) {
// 	captcha := api2captcha.ReCaptcha{
// 		SiteKey: recapInfo.SiteKey,
// 		Url:     recapInfo.Url,
// 		Version: "v3",
// 		Action:  "verify",
// 		Score:   0.7,
// 	}

// 	id, err := client.Send(captcha.ToRequest())
// 	if err != nil {
// 		return "", err
// 	}

// 	var capStatus bool = false
// 	var capResult string = ""

// 	for !capStatus {
// 		code, err := client.GetResult(id)
// 		if err != nil {
// 			log.Error(err.Error(), location)
// 			return "", err
// 		}

// 		if code == nil {
// 			log.General("Waiting for captcha...", location)
// 			time.Sleep(5 * time.Second)
// 			continue
// 		}

// 		capResult = *code
// 		capStatus = true
// 	}

// 	return capResult, nil
// }
