package queue

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"crypto/sha256"
	"encoding/hex"
	"strconv"

	http "github.com/bogdanfinn/fhttp"
)

// The proof-of-work user ID is read from the queue page metadata.

// Structs
type Parameters struct {
	Type       string `json:"type"`
	Input      string `json:"input"`
	Runs       int    `json:"runs"`
	Complexity int    `json:"complexity"`
	ZeroCount  int    `json:"zeroCount"`
}

// Challenge represents the main JSON data structure.
type PowChallenge struct {
	Function         string     `json:"function"`
	SessionID        string     `json:"sessionId"`
	Meta             string     `json:"meta"`
	Parameters       Parameters `json:"parameters"`
	ChallengeDetails string     `json:"challengeDetails"`
}

func (queue *queueSession) GetPowChallenge(queueInfo QueueInfo) (PowChallenge, error) {

	var powRes PowChallenge
	powUrl := fmt.Sprintf("https://%s/challengeapi/pow/challenge/%s", queueInfo.ProofOfWorkHost, queueInfo.UserId)
	req, err := http.NewRequest(http.MethodPost, powUrl, nil)

	if err != nil {
		return powRes, err
	}

	req.Header = http.Header{
		"Host":                           {queueInfo.ProofOfWorkHost},
		"accept":                         {"*/*"},
		"accept-language":                {"en-US,en;q=0.9"},
		"origin":                         {queueInfo.ProofOfWorkHost},
		"powtag-customerid":              {queueInfo.CustomerID},
		"powtag-eventid":                 {queueInfo.EventID},
		"powtag-hash":                    {queueInfo.ChallengeApiChecksumHash},
		"powtag-userid":                  {queueInfo.UserId},
		"referer":                        {queue.QueueUrl},
		"sec-ch-ua":                      {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"sec-ch-ua-mobile":               {"?0"},
		"sec-ch-ua-platform":             {"\"Windows\""},
		"sec-fetch-dest":                 {"empty"},
		"sec-fetch-mode":                 {"cors"},
		"sec-fetch-site":                 {"same-origin"},
		"user-agent":                     {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
		"x-queueit-challange-customerid": {queueInfo.CustomerID},
		"x-queueit-challange-eventid":    {queueInfo.EventID},
		"x-queueit-challange-hash":       {queueInfo.ChallengeApiChecksumHash},
		"x-queueit-challange-reason":     {"1"},
		"x-queueit-challange-userid":     {queueInfo.UserId},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return powRes, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return powRes, fmt.Errorf("Get pow challenge bad status code! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return powRes, err
	}

	err = json.Unmarshal(body, &powRes)

	if err != nil {
		return powRes, err
	}

	return powRes, nil
}

type PowPayLoad struct {
	ChallengeType    string     `json:"challengeType"`
	SessionID        string     `json:"sessionId"`
	ChallengeDetails string     `json:"challengeDetails"`
	Solution         string     `json:"solution"`
	Stats            QueueStats `json:"stats"`
	CustomerID       string     `json:"customerId"`
	EventID          string     `json:"eventId"`
	Version          int        `json:"version"`
}

type QueueStats struct {
	Duration       int    `json:"duration"`
	Tries          int    `json:"tries"`
	UserAgent      string `json:"userAgent"`
	Screen         string `json:"screen"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browserVersion"`
	IsMobile       bool   `json:"isMobile"`
	OS             string `json:"os"`
	OSVersion      string `json:"osVersion"`
	CookiesEnabled bool   `json:"cookiesEnabled"`
}

// Pow response

type PowSessInfo struct {
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
type SubmitPowRes struct {
	IsVerified bool        `json:"isVerified"`
	Timestamp  string      `json:"timestamp"`
	Session    PowSessInfo `json:"sessionInfo"`
}

func (queue *queueSession) SolvePow(qInfo QueueInfo, powInfo PowChallenge) (SubmitPowRes, error) {
	var powRes SubmitPowRes
	solution, err := GetHash(powInfo.Parameters.Input, powInfo.Parameters.Complexity, powInfo.Parameters.Runs)
	if err != nil {
		return powRes, err
	}

	powSolution, err := EncodeHash(solution)
	if err != nil {
		return powRes, err
	}

	powChallengePayLoad := PowPayLoad{
		ChallengeType:    "proofofwork",
		SessionID:        powInfo.SessionID,
		ChallengeDetails: powInfo.ChallengeDetails,
		Solution:         powSolution,
		Stats: QueueStats{
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
		CustomerID: qInfo.CustomerID,
		EventID:    qInfo.EventID,
		Version:    6,
	}

	jsonData, err := json.Marshal(powChallengePayLoad)

	if err != nil {
		return powRes, err
	}

	powUrl := fmt.Sprintf("https://%s/challengeapi/verify", qInfo.ProofOfWorkHost)

	req, err := http.NewRequest(http.MethodPost, powUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		return powRes, err
	}

	req.Header = http.Header{
		"Host":               {qInfo.ProofOfWorkHost},
		"sec-ch-ua":          {"\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Google Chrome\";v=\"114\""},
		"accept":             {"application/json, text/javascript, */*; q=0.01"},
		"content-type":       {"application/json"},
		"x-requested-with":   {"XMLHttpRequest"},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"},
		"sec-ch-ua-platform": {"\"Windows\""},
		"origin":             {fmt.Sprintf("https://%s", qInfo.ProofOfWorkHost)},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-dest":     {"empty"},
		"referer":            {queue.QueueUrl},
		"accept-language":    {"it-IT,it;q=0.9,en-IT;q=0.8,en;q=0.7,si-LK;q=0.6,si;q=0.5,en-US;q=0.4"},
	}

	res, err := queue.httpClient.Do(req)

	if err != nil {
		return powRes, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return powRes, fmt.Errorf("Submit pow challenge bas status code! [%v]", res.StatusCode)
	}

	if err != nil {
		return powRes, err
	}

	err = json.Unmarshal(body, &powRes)

	if err != nil {
		return powRes, err
	}

	return powRes, nil
}

type Solution struct {
	Postfix int64  `json:"postfix"`
	Hash    string `json:"hash"`
}

func GetHash(input string, complexity, runs int) ([]Solution, error) {
	solutions := make([]Solution, runs)
	var postFix int64
	hash := sha256.New()
	var hashBytes [sha256.Size]byte

	inputBytes := []byte(input)
	postFixBytes := make([]byte, 0, 20) // all 64-bit integers will fit

	for i := 0; i < runs; i++ {
		for {
			hash.Reset()
			hash.Write(inputBytes)
			postFixBytes = strconv.AppendInt(postFixBytes[:0], postFix, 10)
			hash.Write(postFixBytes)
			hash.Sum(hashBytes[:0])

			if checkZeroPrefix(hashBytes, complexity) {
				solutions[i] = Solution{
					Postfix: postFix,
					Hash:    hex.EncodeToString(hashBytes[:]),
				}
				postFix++
				break
			}

			postFix++
		}
	}

	return solutions, nil
}

// checkZeroPrefix checks if the first characters are all zeros in hexadecimal presentation
func checkZeroPrefix(data [sha256.Size]byte, count int) bool {
	byteCount := (count + 1) / 2

	for i := 0; i < byteCount-1; i++ {
		if data[i] != 0x00 {
			return false
		}
	}

	if count%2 == 0 {
		if data[byteCount-1] != 0x00 {
			return false
		}
	} else {
		if (data[byteCount-1] & 0xf0) != 0x00 {
			return false
		}
	}

	return true
}

type HashItem struct {
	Hash    string `json:"hash"`
	Postfix int    `json:"postfix"`
}

type HashChallenge struct {
	Hash []HashItem `json:"hash"`
	Type string     `json:"type"`
}

func EncodeHash(solutions []Solution) (string, error) {
	var hashList []HashItem

	for _, elem := range solutions {
		hashList = append(hashList, HashItem{
			Hash:    elem.Hash,
			Postfix: int(elem.Postfix),
		})
	}

	powJson := HashChallenge{
		Hash: hashList,
		Type: "HashChallenge",
	}

	jsonData, err := json.Marshal(powJson)
	if err != nil {
		return "", err
	}

	// Encode the JSON data to base64
	encodedData := base64.StdEncoding.EncodeToString(jsonData)

	return encodedData, nil
}
