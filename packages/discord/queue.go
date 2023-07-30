package discord

import (
	"sync"
	"time"

	"github.com/TheAldriguzDev/tienne_aio/packages/log"
)

var location string = "Discord"

type WebhookQueue struct {
	queue      []*DiscordParamsWithRetries
	maxRetries int
	sync.Mutex
}

type DiscordParamsWithRetries struct {
	Msg     *DiscordParams
	Retries int
}

// Create the queue
func StartQueue(maxRetries int) *WebhookQueue {
	return &WebhookQueue{
		maxRetries: maxRetries,
	}
}

// Enqueue function
func (wq *WebhookQueue) Enqueue(msg *DiscordParams) {
	wq.Lock()
	defer wq.Unlock()
	wq.queue = append(wq.queue, &DiscordParamsWithRetries{
		Msg:     msg,
		Retries: 0,
	})
}

// Dequeue function
func (wq *WebhookQueue) Dequeue() *DiscordParamsWithRetries {
	wq.Lock()
	defer wq.Unlock()
	if len(wq.queue) == 0 {
		return nil
	}
	msg := wq.queue[0]
	wq.queue = wq.queue[1:]
	return msg
}

func (wq *WebhookQueue) Run() {
	for {
		msg := wq.Dequeue()
		if msg == nil {
			// If no webhook are in queue, we wait a small amount of time
			time.Sleep(500 * time.Millisecond)
			continue
		}

		err := PostWebhook(msg.Msg.URL, *msg.Msg)
		if err != nil {
			log.Error("Error while sending the webhook! Retrying...", location)
			msg.Retries++
			if msg.Retries < wq.maxRetries {
				wq.Enqueue(msg.Msg)
			} else {
				log.Error("Send webhook max retry reached!", location)
			}
		} else {
			log.Success("Webhook has been sent!", location)
			// We wait for two second to proceed the queue
			time.Sleep(750 * time.Millisecond)
		}
	}
}
