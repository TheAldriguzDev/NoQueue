package queue

import tls_client "github.com/bogdanfinn/tls-client"

type queueSession struct {
	httpClient       tls_client.HttpClient
	ProxyArr         []string `json:"proxy"`
	ProxyFile        string
	UserAgent        string
	Proxy            string
	QueueUrl         string
	SoftBlockUrl     string
	EncodedChallenge string
	id               string
}

// queue struct

type Language struct {
	Text  string `json:"Text"`
	Value string `json:"Value"`
}

// Supporting struct for the 'challenges' field
type Challenge struct {
	Name         string `json:"name"`
	HasAnimation bool   `json:"hasAnimation"`
}

// The main struct
type QueueInfo struct {
	CustomerID                               string      `json:"customerId"`
	EventID                                  string      `json:"eventId"`
	RequireKey                               bool        `json:"requireKey"`
	IsQueueitEnqueueTokenRequired            bool        `json:"isQueueitEnqueueTokenRequired"`
	QueueID                                  string      `json:"queueId"`
	TargetURL                                string      `json:"targetUrl"`
	CustomURLParams                          string      `json:"customUrlParams"`
	Culture                                  string      `json:"culture"`
	Languages                                []Language  `json:"languages"`
	Layout                                   string      `json:"layout"`
	LayoutVersion                            int         `json:"layoutVersion"`
	SDKVersion                               string      `json:"sdkVersion"`
	IsBeforeOrIdle                           bool        `json:"isBeforeOrIdle"`
	IsIdle                                   bool        `json:"isIdle"`
	ChallengeVerifyEndpoint                  string      `json:"challengeVerifyEndpoint"`
	ShowChallenges                           bool        `json:"showChallenges"`
	ChallengesIssuedByReason                 int         `json:"challengesIssuedByReason"`
	Challenges                               []Challenge `json:"challenges"`
	IsPreview                                bool        `json:"isPreview"`
	IsWebPushNotificationEnabled             bool        `json:"isWebPushNotificationEnabled"`
	WebPushPublicKey                         string      `json:"webPushPublicKey"`
	ProofOfWorkSource                        string      `json:"proofOfWorkSource"`
	ProofOfWorkHost                          string      `json:"proofOfWorkHost"`
	BotDetectSource                          string      `json:"botDetectSource"`
	BotDetectHost                            string      `json:"botDetectHost"`
	ReCaptchaSource                          string      `json:"reCaptchaSource"`
	ReCaptchaHost                            string      `json:"reCaptchaHost"`
	InviteOnlyEmailVerificationSource        string      `json:"inviteOnlyEmailVerificationSource"`
	InviteOnlyEmailVerificationHost          string      `json:"inviteOnlyEmailVerificationHost"`
	ShowInviteOnlyEmailVerificationChallenge bool        `json:"showInviteOnlyEmailVerificationChallenge"`
	IsCustomerUsingEnqueueTokenFormatV1      bool        `json:"isCustomerUsingEnqueueTokenFormatV1"`
	ChallengeApiChecksumHash                 string      `json:"challengeApiChecksumHash"`
	UserId                                   string      `json:"userId"`
	EncodedChallenge                         string      `json:"encodedChallenge"`
}
