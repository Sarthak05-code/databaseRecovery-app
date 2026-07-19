package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackMessage struct {
	Text string `json:"text"`
}

func SendSlackNotification(webhookURL, text string) error {
	if webhookURL == "" {
		return nil
	}

	msg := SlackMessage{Text: text}
	buf := new(bytes.Buffer)

	json.NewEncoder(buf).Encode(msg)

	resp, err := http.Post(webhookURL, "application/json", buf)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack status code : %d", resp.StatusCode)
	}
	return nil
}
