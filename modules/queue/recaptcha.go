package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	captchasolver "github.com/TheAldriguzDev/tienne_aio/packages/captchaSolver"
	"github.com/TheAldriguzDev/tienne_aio/packages/log"

	http "github.com/bogdanfinn/fhttp"
)

type RecapData struct {
	ChallengeDetails string `json:"challengeDetails"`
	SessionID        string `json:"sessionId"`
	SiteKey          string `json:"siteKey"`
}

func (queue *queueSession) GetRecaptcha(queueInfo QueueInfo) (RecapData, error) {
	var recapInfo RecapData
	recaptchaUrl := fmt.Sprintf("https://%s/challengeapi/recaptchainvisible/challenge/", queueInfo.ReCaptchaHost)

	req, err := http.NewRequest(http.MethodPost, recaptchaUrl, nil)

	if err != nil {
		return recapInfo, err
	}

	req.Header = http.Header{
		"Host":                           {queueInfo.ReCaptchaHost},
		"sec-ch-ua":                      {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"x-queueit-challange-eventid":    {queueInfo.EventID},
		"x-queueit-challange-customerid": {queueInfo.CustomerID},
		"sec-ch-ua-mobile":               {"?0"},
		"user-agent":                     {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
		"x-queueit-challange-hash":       {queueInfo.ChallengeApiChecksumHash},
		"x-queueit-challange-reason":     {"1"},
		"sec-ch-ua-platform":             {"\"Windows\""},
		"accept":                         {"*/*"},
		"origin":                         {fmt.Sprintf("https://%s", queueInfo.ReCaptchaHost)},
		"sec-fetch-site":                 {"same-origin"},
		"sec-fetch-mode":                 {"cors"},
		"sec-fetch-dest":                 {"empty"},
		"referer":                        {queue.QueueUrl},
		"accept-language":                {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
		"Content-Type":                   {"application/x-www-form-urlencoded"},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return recapInfo, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return recapInfo, fmt.Errorf("Get recaptcha challenge bad status code! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return recapInfo, err
	}

	err = json.Unmarshal(body, &recapInfo)

	return recapInfo, nil
}

// Now we can submit the recaptcha

type RecapStats struct {
	UserAgent      string `json:"userAgent"`
	Screen         string `json:"screen"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browserVersion"`
	IsMobile       bool   `json:"isMobile"`
	OS             string `json:"os"`
	OSVersion      string `json:"osVersion"`
	CookiesEnabled bool   `json:"cookiesEnabled"`
	Tries          int    `json:"tries"`
	Duration       int    `json:"duration"`
}

// ChallengeData represents the main JSON object.
type RecaptchaPayload struct {
	ChallengeType    string     `json:"challengeType"`
	SessionID        string     `json:"sessionId"`
	ChallengeDetails string     `json:"challengeDetails"`
	Solution         string     `json:"solution"`
	Stats            RecapStats `json:"stats"`
	CustomerID       string     `json:"customerId"`
	EventID          string     `json:"eventId"`
	Version          int        `json:"version"`
}

// Response

type RecapSessInfo struct {
	SessionID     string `json:"sessionId"`
	Timestamp     string `json:"timestamp"`
	Checksum      string `json:"checksum"`
	SourceIP      string `json:"sourceIp"`
	ChallengeType string `json:"challengeType"`
	Version       int    `json:"version"`
	CustomerID    string `json:"customerId"`
	WaitingRoomID string `json:"waitingRoomId"`
}

// Struct principale
type SubmitRecapRes struct {
	IsVerified bool        `json:"isVerified"`
	Timestamp  string      `json:"timestamp"`
	Session    PowSessInfo `json:"sessionInfo"`
}

func (queue *queueSession) SubmitRecaptchaInv(queueInfo QueueInfo, recapData RecapData) (SubmitRecapRes, error) {
	var recapRes SubmitRecapRes
	log.Warn("Solving captcha...", location)

	// Add user api key
	client := captchasolver.Create2CapClient("ae5f40ecb8180449e9cbfce2c7c438b0")

	recapInfo := captchasolver.RecapStruct{
		SiteKey: recapData.SiteKey,
		Url:     queue.QueueUrl,
	}

	recapValue, err := captchasolver.SolveQueueItRecap(recapInfo, client)

	if err != nil {
		return recapRes, err
	}

	log.Success("Captcha has been solved!", location)

	recapPayload := RecaptchaPayload{
		ChallengeType:    "recaptcha-invisible",
		SessionID:        recapData.SessionID,
		ChallengeDetails: recapData.ChallengeDetails,
		Solution:         recapValue,
		Stats: RecapStats{
			Duration:       1697,
			Tries:          1,
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
			Screen:         "1739 x 979",
			Browser:        "Chrome",
			BrowserVersion: "115.0.0.0",
			IsMobile:       false,
			OS:             "Windows",
			OSVersion:      "10",
			CookiesEnabled: true,
		},
		CustomerID: queueInfo.CustomerID,
		EventID:    queueInfo.EventID,
		Version:    6,
	}

	payload, err := json.Marshal(recapPayload)

	if err != nil {
		return recapRes, err
	}

	recapUrl := fmt.Sprintf("https://%s/challengeapi/verify", queueInfo.ReCaptchaHost)

	req, err := http.NewRequest(http.MethodPost, recapUrl, bytes.NewBuffer(payload))

	if err != nil {
		return recapRes, err
	}

	req.Header = http.Header{
		"Host":               {queueInfo.ReCaptchaHost},
		"sec-ch-ua":          {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"accept":             {"application/json, text/javascript, */*; q=0.01"},
		"content-type":       {"application/json"},
		"x-requested-with":   {"XMLHttpRequest"},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
		"sec-ch-ua-platform": {"\"Windows\""},
		"origin":             {fmt.Sprintf("https://%s", queueInfo.ReCaptchaHost)},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-dest":     {"empty"},
		"referer":            {queue.QueueUrl},
		"accept-language":    {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return recapRes, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return recapRes, fmt.Errorf("Submit recaptcha bad status code! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return recapRes, err
	}

	err = json.Unmarshal(body, &recapRes)

	if err != nil {
		return recapRes, err
	}

	return recapRes, nil
}
