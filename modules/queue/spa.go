package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

type SpaPayload struct {
	TargetURL               string `json:"targetUrl"`
	CustomURLParams         string `json:"customUrlParams"`
	LayoutVersion           int64  `json:"layoutVersion"`
	LayoutName              string `json:"layoutName"`
	IsClientReadyToRedirect bool   `json:"isClientRedayToRedirect"`
	IsBeforeOrIdle          bool   `json:"isBeforeOrIdle"`
}

type SpaResponse struct {
	RedirectURL        string `json:"redirectUrl"`
	IsRedirectToTarget bool   `json:"isRedirectToTarget"`
}

func (queue *queueSession) CheckQueueStatus(queueInfo QueueInfo, enqueue EnqueueResponse) (string, error) {
	var spaUrl string
	if queueInfo.EncodedChallenge != "" {
		spaUrl = fmt.Sprintf("https://%s/spa-api/queue/%s/%s/%s/status?cid=%s&l=%s&scv=%s", url.QueryEscape(queueInfo.ProofOfWorkHost), url.QueryEscape(queueInfo.CustomerID), url.QueryEscape(queueInfo.EventID), url.QueryEscape(enqueue.QueueID), url.QueryEscape(queueInfo.Culture), url.QueryEscape(queueInfo.Layout), queueInfo.EncodedChallenge)
	} else {
		spaUrl = fmt.Sprintf("https://%s/spa-api/queue/%s/%s/%s/status?cid=%s&l=%s", url.QueryEscape(queueInfo.ProofOfWorkHost), url.QueryEscape(queueInfo.CustomerID), url.QueryEscape(queueInfo.EventID), url.QueryEscape(enqueue.QueueID), url.QueryEscape(queueInfo.Culture), url.QueryEscape(queueInfo.Layout))
	}
	spaPayload := SpaPayload{
		TargetURL:               queueInfo.TargetURL,
		CustomURLParams:         queueInfo.CustomURLParams,
		LayoutVersion:           int64(queueInfo.LayoutVersion),
		LayoutName:              queueInfo.Layout,
		IsClientReadyToRedirect: true,
		IsBeforeOrIdle:          false,
	}

	jsonPayload, err := json.Marshal(spaPayload)

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, spaUrl, bytes.NewBuffer(jsonPayload))

	req.Header = http.Header{
		"authority":          {fmt.Sprintf("%s", queueInfo.ProofOfWorkHost)},
		"sec-ch-ua":          {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"accept":             {"application/json, text/javascript, */*; q=0.01"},
		"content-type":       {"application/json"},
		"x-requested-with":   {"XMLHttpRequest"},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"},
		"sec-ch-ua-platform": {"\"macOS\""},
		"origin":             {fmt.Sprintf("https://%s", queueInfo.ProofOfWorkHost)},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-dest":     {"empty"},
		"referer":            {queue.QueueUrl},
		"accept-language":    {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
		http.HeaderOrderKey: {
			"authority",
			"accept",
			"accept-encoding",
			"accept-language",
			"content-type",
			"origin",
			"referer",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"user-agent",
			"x-requested-with",
		},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		switch res.StatusCode {
		case 302:
			return res.Header.Get("Location"), fmt.Errorf("Got a redirect while waiting in queue!")
		default:
			return "", fmt.Errorf("Check queue bad status code! [%v]", res.StatusCode)
		}
	}

	if err != nil {
		return "", err
	}

	if !strings.Contains(string(body), "redirectUrl") {
		return "", fmt.Errorf("Still in queue...")
	}

	var spaRes SpaResponse

	err = json.Unmarshal(body, &spaRes)

	if err != nil {
		return "", err
	}

	return spaRes.RedirectURL, nil

}
