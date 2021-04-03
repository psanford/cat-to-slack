package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

func main() {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("Must set SLACK_WEBHOOK_URL in environment")
	}

	r := bufio.NewReader(os.Stdin)

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		fmt.Println(line)

		err = slack.PostWebhook(webhookURL, &slack.WebhookMessage{
			Text: line,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
