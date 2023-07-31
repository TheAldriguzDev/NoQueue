package queue

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TheAldriguzDev/tienne_aio/packages/discord"
	"github.com/TheAldriguzDev/tienne_aio/packages/log"
	proxyManager "github.com/TheAldriguzDev/tienne_aio/packages/proxy"
	filereader "github.com/TheAldriguzDev/tienne_aio/packages/reader"

	http "github.com/bogdanfinn/fhttp"

	tls_client "github.com/bogdanfinn/tls-client"
)

var location string = "Queue"

func InitTls(proxy string) (queueSession, error) {
	proxy = strings.ReplaceAll(proxy, "\r", "")
	proxyData := strings.Split(strings.TrimSpace(proxy), ":")
	proxyUrl := fmt.Sprintf("http://%s:%s@%s:%s", proxyData[2], proxyData[3], proxyData[0], proxyData[1])

	jarOptions := []tls_client.CookieJarOption{}

	jar := tls_client.NewCookieJar(jarOptions...)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.MeshIos2),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithProxyUrl(proxyUrl),
		tls_client.WithInsecureSkipVerify(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	if err != nil {
		fmt.Println("Error while creating the tls client!")
		return queueSession{}, err
	}

	return queueSession{
		httpClient: client,
	}, nil
}

func (queue *queueSession) RotateProxy() error {
	proxie, err := proxyManager.RandomProxy(queue.ProxyArr)
	if err != nil {
		return err
	}
	proxie = strings.ReplaceAll(proxie, "\r", "")
	proxyData := strings.Split(strings.TrimSpace(proxie), ":")
	if len(proxyData) < 4 {
		return fmt.Errorf("The proxy format is not correct! %s", proxie)
	}
	proxyUrl := fmt.Sprintf("http://%s:%s@%s:%s", proxyData[2], proxyData[3], proxyData[0], proxyData[1])
	if err != nil {
		return err
	} else {
		err = queue.httpClient.SetProxy(proxyUrl)
		if err != nil {
			return err
		}
		queue.Proxy = proxie
	}
	return nil
}

func InitQueue(taskFile string, successWh string, failedWh string, apiKey string) {

	webhookQueue := discord.StartQueue(10)
	go webhookQueue.Run()

	proxyFile, err := filereader.ReadFolder("proxies", "proxies")

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	proxyArr, err := proxyManager.ReadProxyFile(proxyFile)

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	file, err := os.Open("./tasks/" + taskFile)

	if err != nil {
		log.Error(err.Error(), location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	defer file.Close()
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()

	if err != nil {
		log.Error("Error while reading the csv file!", location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	// We check if the csv file is empty, and in negative case we remove the first line that contains the table informations
	if len(rows) > 0 {
		// We remove the first line of the csv table
		rows = rows[1:]
	} else {
		log.Error("The csv is empty!", location)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}

	var wg sync.WaitGroup

	var i int = 0
	for _, row := range rows {
		i++
		delay, err := strconv.Atoi(row[2])
		if err != nil {
			log.Error("Error while converting the delay!", location)
			time.Sleep(5 * time.Second)
			return
		}
		wg.Add(1)
		go QueueModule(&wg, proxyArr, webhookQueue, successWh, row, log.PrettyId(i), delay, apiKey)
	}

	wg.Wait()

	time.Sleep(30 * time.Second)
	log.Success("All task has been completed!", location)
}

func QueueModule(wg *sync.WaitGroup, proxyArr []string, webhookQueue *discord.WebhookQueue, successWebhook string, row []string, id string, delay int, apiKey string) {
	defer wg.Done()

	proxie, err := proxyManager.RandomProxy(proxyArr)
	module, err := InitTls(proxie)
	module.Proxy = proxie
	module.apiKey = apiKey
	module.id = id
	if err != nil {
		log.ErrorTask(err.Error(), location, id)
		time.Sleep(time.Duration(delay) * time.Second)
		os.Exit(0)
	}

	module.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	module.ProxyArr = proxyArr
	log.SuccessTask("Queue initialiased!", location, id)

	// queueUrl, err := module.GetRedirect()

	// if err != nil {
	// 	log.Error(err.Error(), location)
	// 	os.Exit(0)
	// }

	module.QueueUrl = row[0]

	// Now we can fetch all the queue info + challenges:
	queueInfo, err, redirectUrl := module.GetQueueChallenge()
	if err != nil {
		if err.Error() == "Softblock found!" {
			log.WarnTask("Softblock has been found!", location, id)
			// Here we need to solve softblock challenge

			// Here we need to add a softblock url field
			module.SoftBlockUrl = fmt.Sprintf("https://footlocker.queue-it.net%s", redirectUrl)

			queueInfo, err, redirectUrl = module.GetQueueChallenge()
			if err != nil {
				log.ErrorTask(err.Error(), location, id)
				return
			}

			// Now we can solve the challenge
			response := module.SoftBlockStep(queueInfo)
			if response.IsVerified != true {
				log.ErrorTask("Error while solving the queue it captcha!", location, id)
			} else {
				softChallenge := ChallengeSession{
					SessionId:     response.SessionInfo.SessionID,
					Timestamp:     response.SessionInfo.Timestamp,
					Checksum:      response.SessionInfo.Checksum,
					SourceIP:      response.SessionInfo.SourceIP,
					ChallengeType: response.SessionInfo.ChallengeType,
					Version:       response.SessionInfo.Version,
					CustomerId:    response.SessionInfo.CustomerID,
					WaitingRoomId: response.SessionInfo.WaitingRoomID,
				}

				jsonData, err := json.Marshal(softChallenge)
				if err != nil {
					log.ErrorTask(err.Error(), location, id)
					return
				}
				encodedParams := url.QueryEscape(string(jsonData))
				module.EncodedChallenge = encodedParams
			}

			// Than we had to retry with the correct info like in the status to get the other challenges

			module.SoftBlockUrl = ""
			queueInfo, err, redirectUrl = module.GetQueueChallenge()
			if err != nil {
				log.ErrorTask(err.Error(), location, id)
				return
			}

			// Here we have solved the challenge

		} else {
			log.ErrorTask(err.Error(), location, id)
			return
		}
	}

	var challengeList []string

	if queueInfo.CustomerID == "" || queueInfo.EventID == "" || queueInfo.QueueID == "" || queueInfo.TargetURL == "" || queueInfo.Culture == "" || queueInfo.Layout == "" {
		log.ErrorTask("Queue information not found!", location, id)
		return
	}

	for _, challenge := range queueInfo.Challenges {
		challengeList = append(challengeList, challenge.Name)
	}

	if len(challengeList) == 0 {
		log.WarnTask("No challenges has been found!", location, id)

		enqueueRes, err := module.Enqueue(queueInfo)
		if err != nil {
			log.ErrorTask(err.Error(), location, id)
			return
		}

		log.SuccessTask("Queue has been joined!", location, id)

		var queueStatus bool = false
		var queueRedirect string = ""

		for !queueStatus {
			redirectUrl, err := module.CheckQueueStatus(queueInfo, enqueueRes)
			if err != nil {
				if err.Error() == "Still in queue..." {
					log.GeneralTask("Waiting in queue...", location, id)
				} else {
					log.ErrorTask("Error while waiting in queue!", location, id)
					log.ErrorTask(err.Error(), location, id)
				}
				time.Sleep(time.Duration(delay) * time.Second)
				continue
			}

			if redirectUrl == "" {
				log.ErrorTask("No redirect URL has been found!", location, id)
				time.Sleep(time.Duration(module.Delay) * time.Second)
				continue
			}

			queueStatus = true
			queueRedirect = redirectUrl
		}

		log.SuccessTask("Queue has been passed!", location, id)

		module.SendLog(queueInfo, queueRedirect, webhookQueue)
		return
	}

	var enqueuePayLoad EnqueueChallengeData

	enqueuePayLoad.LayoutName = queueInfo.Layout
	enqueuePayLoad.CustomURLParams = queueInfo.CustomURLParams
	enqueuePayLoad.TargetURL = queueInfo.TargetURL
	enqueuePayLoad.Referrer = module.QueueUrl

	for _, challenge := range challengeList {
		log.WarnTask("Solving challenge: "+challenge, location, id)

		switch challenge {
		case "BotDetect":
			response := module.SoftBlockStep(queueInfo)
			enqueuePayLoad.ChallengeSessions = append(enqueuePayLoad.ChallengeSessions, ChallengeSession{
				SessionId:     response.SessionInfo.SessionID,
				Timestamp:     response.SessionInfo.Timestamp,
				Checksum:      response.SessionInfo.Checksum,
				SourceIP:      response.SessionInfo.SourceIP,
				ChallengeType: response.SessionInfo.ChallengeType,
				Version:       response.SessionInfo.Version,
				CustomerId:    response.SessionInfo.CustomerID,
				WaitingRoomId: response.SessionInfo.WaitingRoomID,
			})
			break
		case "ProofOfWork":
			response := module.PowStep(queueInfo)
			enqueuePayLoad.ChallengeSessions = append(enqueuePayLoad.ChallengeSessions, ChallengeSession{
				SessionId:     response.Session.SessionID,
				Timestamp:     response.Session.Timestamp,
				Checksum:      response.Session.Checksum,
				SourceIP:      response.Session.SourceIP,
				ChallengeType: response.Session.ChallengeType,
				Version:       response.Session.Version,
				CustomerId:    response.Session.CustomerID,
				WaitingRoomId: response.Session.WaitingRoomID,
			})
			break
		case "RecaptchaInvisible":
			response := module.RecapInvisStep(queueInfo)
			enqueuePayLoad.ChallengeSessions = append(enqueuePayLoad.ChallengeSessions, ChallengeSession{
				SessionId:     response.Session.SessionID,
				Timestamp:     response.Session.Timestamp,
				Checksum:      response.Session.Checksum,
				SourceIP:      response.Session.SourceIP,
				ChallengeType: response.Session.ChallengeType,
				Version:       response.Session.Version,
				CustomerId:    response.Session.CustomerID,
				WaitingRoomId: response.Session.WaitingRoomID,
			})
			break
		default:
			log.ErrorTask("Challenge not supported yet!", location, id)
		}
	}

	log.SuccessTask("All challenges has been solved!", location, id)

	enqueueRes, err := module.EnqueueChallenge(queueInfo, enqueuePayLoad)
	if err != nil {
		if err.Error() == "Softblock on enqueue" {
			// We need to solve softblock here
			log.WarnTask("Found a softblock challenge while enqueuing!", location, id)
			response := module.SoftBlockStep(queueInfo)

			if response.IsVerified != true {
				log.ErrorTask("Error while solving the queue it captcha!", location, id)
			} else {
				softChallenge := ChallengeSession{
					SessionId:     response.SessionInfo.SessionID,
					Timestamp:     response.SessionInfo.Timestamp,
					Checksum:      response.SessionInfo.Checksum,
					SourceIP:      response.SessionInfo.SourceIP,
					ChallengeType: response.SessionInfo.ChallengeType,
					Version:       response.SessionInfo.Version,
					CustomerId:    response.SessionInfo.CustomerID,
					WaitingRoomId: response.SessionInfo.WaitingRoomID,
				}

				jsonData, err := json.Marshal(softChallenge)
				if err != nil {
					log.ErrorTask(err.Error(), location, id)
					return
				}
				encodedParams := url.QueryEscape(string(jsonData))
				queueInfo.EncodedChallenge = encodedParams

				enqueueRes, err = module.EnqueueChallenge(queueInfo, enqueuePayLoad)
				if err != nil {
					log.ErrorTask(err.Error(), location, id)
					return
				}
			}

		} else {
			log.ErrorTask(err.Error(), location, id)
			return
		}
	}

	log.SuccessTask("Queue has been joined!", location, id)
	queueInfo.EncodedChallenge = ""

	var queueStatus bool = false
	var queueRedirect string = ""

	for !queueStatus {
		redirectUrl, err := module.CheckQueueStatus(queueInfo, enqueueRes)
		if err != nil {
			if err.Error() == "Still in queue..." {
				log.GeneralTask("Waiting in queue...", location, id)
			} else {
				if err.Error() == "Got a redirect while waiting in queue!" {
					if strings.Contains(redirectUrl, "softblock") {
						log.WarnTask("Found a softblock challenge!", location, id)
						response := module.SoftBlockStep(queueInfo)

						if response.IsVerified != true {
							log.ErrorTask("Error while solving the queue it captcha!", location, id)
						} else {
							softChallenge := ChallengeSession{
								SessionId:     response.SessionInfo.SessionID,
								Timestamp:     response.SessionInfo.Timestamp,
								Checksum:      response.SessionInfo.Checksum,
								SourceIP:      response.SessionInfo.SourceIP,
								ChallengeType: response.SessionInfo.ChallengeType,
								Version:       response.SessionInfo.Version,
								CustomerId:    response.SessionInfo.CustomerID,
								WaitingRoomId: response.SessionInfo.WaitingRoomID,
							}

							jsonData, err := json.Marshal(softChallenge)
							if err != nil {
								log.ErrorTask(err.Error(), location, id)
								return
							}
							encodedParams := url.QueryEscape(string(jsonData))
							queueInfo.EncodedChallenge = encodedParams
						}
					}
				} else {
					log.ErrorTask("Error while waiting in queue!", location, id)
					log.ErrorTask(err.Error(), location, id)
				}
			}
			time.Sleep(time.Duration(module.Delay) * time.Second)
			continue
		}

		if redirectUrl == "" {
			log.ErrorTask("No redirect URL has been found!", location, id)
			time.Sleep(time.Duration(module.Delay) * time.Second)
			continue
		}

		queueStatus = true
		queueRedirect = redirectUrl
	}

	log.SuccessTask("Queue has been passed!", location, id)

	module.SendLog(queueInfo, queueRedirect, webhookQueue)
	return
}

func (queue *queueSession) GetRedirect() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://footlocker.queue-it.net/?c=footlocker&e=cxcdtest02&t=https%3A%2F%2Fwww.footlocker.com%2Fsearch%3Fquery%3D6H053P85&cid=en-EN", nil)

	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-language":           {"en-US,en;q=0.9"},
		"sec-ch-ua":                 {`"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"`},
		"sec-ch-ua-mobile":          {`?0`},
		"sec-ch-ua-platform":        {`"Windows"`},
		"sec-fetch-dest":            {`document`},
		"sec-fetch-mode":            {`navigate`},
		"sec-fetch-site":            {`none`},
		"sec-fetch-user":            {`?1`},
		"upgrade-insecure-requests": {`1`},
		"user-agent":                {`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36`},
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
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 302 {
		return "", fmt.Errorf("No redirect has been found!")
	}

	redirectUrl := res.Header.Get("Location")

	splitted := strings.Split(redirectUrl, "://")

	return "https://" + splitted[1], nil
}

func (queue *queueSession) GetQueueChallenge() (QueueInfo, error, string) {
	var queueInfo QueueInfo

	var url string

	// Checking if the queue has some url
	if queue.SoftBlockUrl != "" {
		url = queue.SoftBlockUrl
	} else {
		if queue.EncodedChallenge != "" {
			url = queue.QueueUrl + fmt.Sprintf("&scv=%s", queue.EncodedChallenge)
		} else {
			url = queue.QueueUrl
		}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return queueInfo, err, ""
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-language":           {"en-US,en;q=0.9"},
		"sec-ch-ua":                 {`"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"`},
		"sec-ch-ua-mobile":          {`?0`},
		"sec-ch-ua-platform":        {`"Windows"`},
		"sec-fetch-dest":            {`document`},
		"sec-fetch-mode":            {`navigate`},
		"sec-fetch-site":            {`none`},
		"sec-fetch-user":            {`?1`},
		"upgrade-insecure-requests": {`1`},
		"user-agent":                {`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36`},
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
		return queueInfo, err, ""
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode == 302 {
		if strings.Contains(res.Header.Get("Location"), "softblock") {
			return queueInfo, fmt.Errorf("Softblock found!"), res.Header.Get("Location")
		} else {
			return queueInfo, fmt.Errorf("Unknow redirect found!"), res.Header.Get("Location")
		}
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return queueInfo, err, ""
	}

	queueInfo, err = GetQueueInfo(string(body))
	if err != nil {
		return queueInfo, err, ""
	}

	return queueInfo, nil, ""
}

// General regex function
func GetQueueInfo(htmlCode string) (QueueInfo, error) {
	queueInfo := QueueInfo{}

	fields := map[string]string{
		"customerId":                      `customerId:\s*'([^']+)'`,
		"eventId":                         `eventId:\s*'([^']+)'`,
		"queueId":                         `queueId:\s*'([^']+)'`,
		"targetUrl":                       `targetUrl:\s*decodeURIComponent\('([^']+)'\)`,
		"customUrlParams":                 `customUrlParams:\s*'([^']+)'`,
		"culture":                         `culture:\s*'([^']+)'`,
		"layout":                          `layout:\s*'([^']+)'`,
		"layoutVersion":                   `layoutVersion:\s*(\d+)`,
		"challengeVerifyEndpoint":         `challengeVerifyEndpoint:\s*'([^']+)'`,
		"challenges":                      `challenges:\s*(\[[^\]]+\])`,
		"challengeApiChecksumHash":        `challengeApiChecksumHash:\s*'([^']+)'`,
		"proofOfWorkHost":                 `proofOfWorkHost:\s*'([^']+)'`,
		"botDetectHost":                   `botDetectHost:\s*'([^']+)'`,
		"reCaptchaHost":                   `reCaptchaHost:\s*'([^']+)'`,
		"inviteOnlyEmailVerificationHost": `inviteOnlyEmailVerificationHost:\s*'([^']+)'`,
		"userId":                          `data-userid="([^"]+)"`,
	}

	for field, regexPattern := range fields {
		re := regexp.MustCompile(regexPattern)
		matches := re.FindStringSubmatch(htmlCode)

		if len(matches) >= 2 {
			value := matches[1]
			if field == "challenges" {
				var challenges []Challenge
				err := json.Unmarshal([]byte(value), &challenges)
				if err == nil {
					queueInfo.Challenges = challenges
				}
			} else if field == "layoutVersion" {
				queueInfo.LayoutVersion = ConvertToInt(value)
			} else if field == "targetUrl" {
				decodedValue, err := url.QueryUnescape(value)
				if err == nil {
					queueInfo.TargetURL = decodedValue
				}
			} else if field == "customerId" {
				queueInfo.CustomerID = value
			} else if field == "eventId" {
				queueInfo.EventID = value
			} else if field == "queueId" {
				queueInfo.QueueID = value
			} else if field == "customUrlParams" {
				queueInfo.CustomURLParams = value
			} else if field == "culture" {
				queueInfo.Culture = value
			} else if field == "layout" {
				queueInfo.Layout = value
			} else if field == "challengeVerifyEndpoint" {
				queueInfo.ChallengeVerifyEndpoint = value
			} else if field == "challengeApiChecksumHash" {
				queueInfo.ChallengeApiChecksumHash = value
			} else if field == "proofOfWorkHost" {
				queueInfo.ProofOfWorkHost = value
			} else if field == "botDetectHost" {
				queueInfo.BotDetectHost = value
			} else if field == "reCaptchaHost" {
				queueInfo.ReCaptchaHost = value
			} else if field == "inviteOnlyEmailVerificationHost" {
				queueInfo.InviteOnlyEmailVerificationHost = value
			} else if field == "userId" {
				queueInfo.UserId = value // Store the data-userid value in the struct field
			}
		}
	}

	return queueInfo, nil
}

func ConvertToInt(s string) int {
	var num int
	_, err := fmt.Sscanf(s, "%d", &num)
	if err != nil {
		return 0
	}
	return num
}

func (queue *queueSession) SendLog(qInfo QueueInfo, url string, whQueue *discord.WebhookQueue) {
	whData := &discord.DiscordParams{
		URL: "https://discord.com/api/webhooks/1133840581290893412/L1otnU4Sgb3LkQfTgRNoP1f7_U5TR4DVgd010TMf06zb__hUjsWFh09gUjiciS3obEhi",
		Embeds: []discord.Embed{
			{
				Author: discord.Author{
					Name:    "Hypemedia CLI",
					IconURL: "https://media.discordapp.net/attachments/906886120502075403/1095465707690524753/hypemedialov2.jpg?width=1170&height=1170",
				},
				Title: "Queue IT link ready!",
				Color: 16776975,
				Url:   url,
				Footer: discord.Footer{
					Text: "Hypemedia Monitors",
				},
				Fields: []discord.Field{
					{
						Name:   "Customer ID",
						Value:  qInfo.CustomerID,
						Inline: true,
					},
					{
						Name:   "Event ID",
						Value:  qInfo.EventID,
						Inline: true,
					},
				},
				Thumbnail: discord.Thumbnail{
					URL: "https://awsmp-logos.s3.amazonaws.com/e75e9234-d0c1-4847-8fb0-19e6523d2e86/18ea1379605ef7e783c2f8b9f55beaee.png",
				},
			},
		},
	}

	// err := discord.PostWebhook("https://discord.com/api/webhooks/1133840581290893412/L1otnU4Sgb3LkQfTgRNoP1f7_U5TR4DVgd010TMf06zb__hUjsWFh09gUjiciS3obEhi", *whData)
	// if err != nil {
	// 	log.Error("An error occured while sending the webhook!", location)
	// } else {
	// 	log.Success("Webhook sent", location)
	// 	return
	// }

	whQueue.Enqueue(whData)
}

// Add retry handler
func (queue *queueSession) SoftBlockStep(queueInfo QueueInfo) SoftChallengeResponse {
	var softBlockInfo SoftBlock
	var captchaStatus bool = false
	var captchaResult string = ""

	var softResponse SoftChallengeResponse

	for !captchaStatus {
		log.General("Getting captcha...", location)
		captchaInfo, err := queue.GetCaptcha(queueInfo)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		log.Warn("Solving captcha...", location)
		tempCaptcha, err := SolveCaptcha(captchaInfo.ImageBase64)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		if strings.Contains(captchaResult, " ") {
			log.Error("Error while solving the captcha!", location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		if len(tempCaptcha) < 6 {
			log.Error("Error while solving the captcha!", location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		captchaResult = tempCaptcha
		softBlockInfo = captchaInfo
		softBlockInfo.Response = tempCaptcha

		log.Success("Captcha has been solved: "+captchaResult, location)

		solveRes, err := queue.SolveSoftChallenge(softBlockInfo, queueInfo)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		if solveRes.IsVerified != true {
			log.Error("Error while submitting the captcha [Bad Captcha]!", location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		captchaStatus = true
		softResponse = solveRes
	}

	log.Success("SoftBlock challenge has been solved!", location)
	return softResponse
}

func (queue *queueSession) PowStep(queueInfo QueueInfo) SubmitPowRes {
	var getPowStatus bool = false
	var powChallenge PowChallenge

	for !getPowStatus {
		powChallengeTemp, err := queue.GetPowChallenge(queueInfo)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		getPowStatus = true
		powChallenge = powChallengeTemp
	}

	log.General("Got pow challenge. Solving...", location)

	var submitPowStatus bool = false
	var powResult SubmitPowRes

	for !submitPowStatus {
		powResTemp, err := queue.SolvePow(queueInfo, powChallenge)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		if powResTemp.IsVerified != true {
			log.Error("Error while solving pow challenge!", location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}
		submitPowStatus = true
		powResult = powResTemp
	}

	log.Success("Pow challenge has been solved!", location)
	return powResult
}

// Funziona come deve
func (queue *queueSession) RecapInvisStep(queueInfo QueueInfo) SubmitRecapRes {
	var recapStatus bool = false

	var recapRes SubmitRecapRes

	for !recapStatus {
		log.General("Getting recaptcha...", location)
		recapInfo, err := queue.GetRecaptcha(queueInfo)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		tempCaptcha, err := queue.SubmitRecaptchaInv(queueInfo, recapInfo)
		if err != nil {
			log.Error(err.Error(), location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		if tempCaptcha.IsVerified != true {
			log.Error("Error while solving recaptcha! [NOT VERIFIED]", location)
			time.Sleep(time.Duration(queue.Delay) * time.Second)
			continue
		}

		recapStatus = true
		recapRes = tempCaptcha
	}

	log.Success("Recaptcha challenge has been solved!", location)
	return recapRes
}

// Other

// Not useful
func (queue *queueSession) FetchRedirect() (string, error) {
	req, err := http.NewRequest(http.MethodGet, queue.QueueUrl, nil)

	if err != nil {
		return "", err
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-language":           {"en-US,en;q=0.9"},
		"sec-ch-ua":                 {`"Not?A_Brand";v="8", "Chromium";v="108", "Google Chrome";v="108"`},
		"sec-ch-ua-mobile":          {`?0`},
		"sec-ch-ua-platform":        {`"Windows"`},
		"sec-fetch-dest":            {`document`},
		"sec-fetch-mode":            {`navigate`},
		"sec-fetch-site":            {`none`},
		"sec-fetch-user":            {`?1`},
		"upgrade-insecure-requests": {`1`},
		"user-agent":                {`Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36`},
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
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode == 302 {
		return "", fmt.Errorf("Got another redirect! Stopping...")
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("Fetch redirect unknown status! [%v]", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	href, err := GetHref(string(body))
	if err != nil {
		return "", err
	}

	return href, nil
}

// Not useful
func GetHref(body string) (string, error) {
	regex := regexp.MustCompile(`document\.location\.href\s*=\s*decodeURIComponent\('([^']+)'\)`)
	matches := regex.FindStringSubmatch(body)

	if len(matches) == 2 {
		encodedUrl := matches[1]
		decodedUrl, err := url.QueryUnescape(encodedUrl)
		if err != nil {
			return "", err
		}
		return decodedUrl, nil
	} else {
		return "", fmt.Errorf("Href not found in the redirect!")
	}
}
