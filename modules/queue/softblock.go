package queue

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"

	http "github.com/bogdanfinn/fhttp"
	"github.com/otiai10/gosseract/v2"
)

type SoftBlock struct {
	Key              string `json:"key"`
	ImageBase64      string `json:"imageBase64"`
	Meta             string `json:"meta"`
	SessionID        string `json:"sessionId"`
	ChallengeDetails string `json:"challengeDetails"`
	Response         string `json:"response"`
}

// Funzione per rilevare il testo nell'immagine utilizzando l'OCR
func getTextFromImage(img image.Image) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	// Conversione dell'immagine in formato TIFF (OCR supporta TIFF)
	tempFile, err := os.CreateTemp("", "captcha_*.tiff")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())
	if err := jpeg.Encode(tempFile, img, nil); err != nil {
		return "", err
	}

	// Imposta l'immagine sorgente per l'OCR e ottieni il testo rilevato
	client.SetImage(tempFile.Name())
	client.SetWhitelist("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	text, err := client.Text()
	if err != nil {
		return "", err
	}

	return text, nil
}

func decodeBase64Image(base64Str string) (image.Image, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (queue *queueSession) GetCaptcha(queueInfo QueueInfo) (SoftBlock, error) {
	var softBlockInfo SoftBlock

	captchaUrl := fmt.Sprintf("https://%s/challengeapi/queueitcaptcha/challenge/%s", queueInfo.BotDetectHost, queueInfo.Culture)
	req, err := http.NewRequest(http.MethodPost, captchaUrl, nil)

	if err != nil {
		return softBlockInfo, err
	}

	req.Header = http.Header{
		"authority":                      {queueInfo.BotDetectHost},
		"accept":                         {"*/*"},
		"accept-language":                {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
		"origin":                         {fmt.Sprintf("https://%s", queueInfo.BotDetectHost)},
		"referer":                        {queue.QueueUrl},
		"sec-ch-ua":                      {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"sec-ch-ua-mobile":               {"?0"},
		"sec-ch-ua-platform":             {"\"macOS\""},
		"sec-fetch-dest":                 {"empty"},
		"sec-fetch-mode":                 {"cors"},
		"sec-fetch-site":                 {"same-origin"},
		"user-agent":                     {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"},
		"x-queueit-challange-customerid": {queueInfo.CustomerID},
		"x-queueit-challange-eventid":    {queueInfo.EventID},
		"x-queueit-challange-hash":       {queueInfo.ChallengeApiChecksumHash},
		"x-queueit-challange-reason":     {"0"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return softBlockInfo, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return softBlockInfo, fmt.Errorf("Get captcha bad status code! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return softBlockInfo, err
	}

	err = json.Unmarshal(body, &softBlockInfo)
	if err != nil {
		return softBlockInfo, err
	}

	return softBlockInfo, nil
}

type Stats struct {
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

type SoftBlockChallenge struct {
	ChallengeType    string `json:"challengeType"`
	SessionID        string `json:"sessionId"`
	ChallengeDetails string `json:"challengeDetails"`
	Solution         string `json:"solution"`
	Stats            Stats  `json:"stats"`
	CustomerID       string `json:"customerId"`
	EventID          string `json:"eventId"`
	Version          int    `json:"version"`
}

// Challenge response

type SessionInfo struct {
	SessionID     string `json:"sessionId"`
	Timestamp     string `json:"timestamp"`
	Checksum      string `json:"checksum"`
	SourceIP      string `json:"sourceIp"`
	ChallengeType string `json:"challengeType"`
	Version       int    `json:"version"`
	CustomerID    string `json:"customerId"`
	WaitingRoomID string `json:"waitingRoomId"`
}

type SoftChallengeResponse struct {
	IsVerified  bool        `json:"isVerified"`
	Timestamp   string      `json:"timestamp"`
	SessionInfo SessionInfo `json:"sessionInfo"`
}

// To do:
/*
- Add correct retries
- Control flow when captcha are not solved correctly
*/

func (queue *queueSession) SolveSoftChallenge(softPayload SoftBlock, queueInfo QueueInfo) (SoftChallengeResponse, error) {
	verifyChallenge := fmt.Sprintf("https://%s/challengeapi/verify", queueInfo.BotDetectHost)

	var softResp SoftChallengeResponse

	challengePayload := SoftBlockChallenge{
		ChallengeType:    "botdetect",
		SessionID:        softPayload.SessionID,
		ChallengeDetails: softPayload.ChallengeDetails,
		Solution:         softPayload.Response,
		Stats: Stats{
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
			Screen:         "1739 x 979",
			Browser:        "Chrome",
			BrowserVersion: "115.0.0.0",
			IsMobile:       false,
			OS:             "Windows",
			OSVersion:      "10",
			CookiesEnabled: true,
			Tries:          1, // Modify with correct tries
			Duration:       13012,
		},
		CustomerID: queueInfo.CustomerID,
		EventID:    queueInfo.EventID,
		Version:    6,
	}

	payload, err := json.Marshal(challengePayload)

	if err != nil {
		return softResp, err
	}

	req, err := http.NewRequest(http.MethodPost, verifyChallenge, bytes.NewBuffer(payload))

	if err != nil {
		return softResp, err
	}

	req.Header = http.Header{
		"Accept":           {"application/json, text/javascript, */*; q=0.01"},
		"Accept-Language":  {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
		"Connection":       {"keep-alive"},
		"Content-Type":     {"application/json"},
		"origin":           {fmt.Sprintf("https://%s", queueInfo.ProofOfWorkHost)},
		"referer":          {fmt.Sprintf("https://%s/softblock/?c=%s&e=%s&cid=%s&rticr=0", queueInfo.ProofOfWorkHost, queueInfo.CustomerID, queueInfo.EventID, queueInfo.Culture)},
		"User-Agent":       {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"},
		"X-Requested-With": {"XMLHttpRequest"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return softResp, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return softResp, fmt.Errorf("Softblock challenge bad status code! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return softResp, err
	}

	err = json.Unmarshal(body, &softResp)

	if err != nil {
		return softResp, err
	}
	return softResp, nil
}
