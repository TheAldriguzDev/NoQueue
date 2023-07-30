package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func PostWebhook(url string, msg DiscordParams) error {
	// Check if the webhook url exist
	if url == "" {
		return fmt.Errorf("Expected discord webhook, received null")
	}

	// Convert the message to a json
	jsonData, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	// Create the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set request header
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("Error while loggin the webhook: [%v]", resp.StatusCode)
	}

	defer resp.Body.Close()
	return nil
}
