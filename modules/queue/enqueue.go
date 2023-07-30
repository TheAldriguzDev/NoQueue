package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	http "github.com/bogdanfinn/fhttp"
)

type EnqueuePayload struct {
	LayoutName      string `json:"layoutName"`
	CustomUrlParams string `json:"customUrlParams"`
	TargetUrl       string `json:"targetUrl"`
	Referrer        string `json:"Referrer"`
}

type EnqueueResponse struct {
	QueueID                      string `json:"queueId"`
	RedirectURL                  string `json:"redirectUrl"`
	ChallengeFailed              bool   `json:"challengeFailed"`
	MissingCustomDataKey         bool   `json:"missingCustomDataKey"`
	CustomDataUniqueKeyViolation bool   `json:"customDataUniqueKeyViolation"`
	InvalidQueueitEnqueueToken   bool   `json:"invalidQueueitEnqueueToken"`
	ServerIsBusy                 bool   `json:"serverIsBusy"`
}

func (queue *queueSession) Enqueue(queueInfo QueueInfo) (EnqueueResponse, error) {
	var enqueueRes EnqueueResponse

	enqueueUrl := fmt.Sprintf("https://%s/spa-api/queue/%s/%s/enqueue?cid=%s", queueInfo.ProofOfWorkHost, queueInfo.CustomerID, queueInfo.EventID, queueInfo.Culture)

	enqueuePayload := EnqueuePayload{
		LayoutName:      queueInfo.Layout,
		CustomUrlParams: queueInfo.CustomURLParams,
		TargetUrl:       queueInfo.TargetURL,
		Referrer:        queue.QueueUrl,
	}

	jsonPayload, err := json.Marshal(enqueuePayload)
	if err != nil {
		return enqueueRes, err
	}

	req, err := http.NewRequest(http.MethodPost, enqueueUrl, bytes.NewBuffer(jsonPayload))

	if err != nil {
		return enqueueRes, err
	}

	req.Header = http.Header{
		"Host":                     {fmt.Sprintf("%s", queueInfo.ProofOfWorkHost)},
		"sec-ch-ua":                {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"sec-ch-ua-mobile":         {"?0"},
		"user-agent":               {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"},
		"content-type":             {"application/json"},
		"accept":                   {"application/json, text/javascript, */*; q=0.01"},
		"x-queueit-qpage-referral": {queue.QueueUrl},
		"x-requested-with":         {"XMLHttpRequest"},
		"sec-ch-ua-platform":       {"\"macOS\""},
		"origin":                   {fmt.Sprintf("https://%s", queueInfo.ProofOfWorkHost)},
		"sec-fetch-site":           {"same-origin"},
		"sec-fetch-mode":           {"cors"},
		"sec-fetch-dest":           {"empty"},
		"referer":                  {queue.QueueUrl},
		"accept-language":          {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return enqueueRes, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return enqueueRes, fmt.Errorf("Enqueue bad status code! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return enqueueRes, err
	}

	err = json.Unmarshal(body, &enqueueRes)

	if err != nil {
		return enqueueRes, err
	}

	return enqueueRes, nil
}

type ChallengeSession struct {
	SessionId     string `json:"sessionId"`
	Timestamp     string `json:"timestamp"`
	Checksum      string `json:"checksum"`
	SourceIP      string `json:"sourceIp"`
	ChallengeType string `json:"challengeType"`
	Version       int    `json:"version"`
	CustomerId    string `json:"customerId"`
	WaitingRoomId string `json:"waitingRoomId"`
}

type EnqueueChallengeData struct {
	ChallengeSessions []ChallengeSession `json:"challengeSessions"`
	LayoutName        string             `json:"layoutName"`
	CustomURLParams   string             `json:"customUrlParams"`
	TargetURL         string             `json:"targetUrl"`
	Referrer          string             `json:"Referrer"`
}

func (queue *queueSession) EnqueueChallenge(queueInfo QueueInfo, enqueueData EnqueueChallengeData) (EnqueueResponse, error) {
	var enqueueRes EnqueueResponse

	var enqueueUrl string
	if queueInfo.EncodedChallenge != "" {
		enqueueUrl = fmt.Sprintf("https://%s/spa-api/queue/%s/%s/enqueue?cid=%s&scv=%s", url.QueryEscape(queueInfo.ProofOfWorkHost), url.QueryEscape(queueInfo.CustomerID), url.QueryEscape(queueInfo.EventID), url.QueryEscape(queueInfo.Culture), queueInfo.EncodedChallenge)
	} else {
		enqueueUrl = fmt.Sprintf("https://%s/spa-api/queue/%s/%s/enqueue?cid=%s", url.QueryEscape(queueInfo.ProofOfWorkHost), url.QueryEscape(queueInfo.CustomerID), url.QueryEscape(queueInfo.EventID), url.QueryEscape(queueInfo.Culture))
	}

	//enqueueUrl := fmt.Sprintf("https://%s/spa-api/queue/%s/%s/enqueue?cid=%s", queueInfo.ProofOfWorkHost, queueInfo.CustomerID, queueInfo.EventID, queueInfo.Culture)

	jsonPayload, err := json.Marshal(enqueueData)
	if err != nil {
		return enqueueRes, err
	}

	req, err := http.NewRequest(http.MethodPost, enqueueUrl, bytes.NewBuffer(jsonPayload))

	if err != nil {
		return enqueueRes, err
	}

	req.Header = http.Header{
		"Host":                     {fmt.Sprintf("%s", queueInfo.ProofOfWorkHost)},
		"sec-ch-ua":                {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"sec-ch-ua-mobile":         {"?0"},
		"user-agent":               {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"},
		"content-type":             {"application/json"},
		"accept":                   {"application/json, text/javascript, */*; q=0.01"},
		"x-queueit-qpage-referral": {queue.QueueUrl},
		"x-requested-with":         {"XMLHttpRequest"},
		"sec-ch-ua-platform":       {"\"macOS\""},
		"origin":                   {fmt.Sprintf("https://%s", queueInfo.ProofOfWorkHost)},
		"sec-fetch-site":           {"same-origin"},
		"sec-fetch-mode":           {"cors"},
		"sec-fetch-dest":           {"empty"},
		"referer":                  {queue.QueueUrl},
		"accept-language":          {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return enqueueRes, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		switch res.StatusCode {
		case 302:
			return enqueueRes, fmt.Errorf("Softblock on enqueue")
		default:
			return enqueueRes, fmt.Errorf("Enqueue bad status code! [%v]", res.StatusCode)
		}
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return enqueueRes, err
	}

	err = json.Unmarshal(body, &enqueueRes)

	if err != nil {
		return enqueueRes, err
	}

	if enqueueRes.QueueID == "" {
		return enqueueRes, fmt.Errorf("Queue ID empty!")
	}

	return enqueueRes, nil
}
