package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/psanford/backoff"
	"github.com/slack-go/slack"
)

var prefix = flag.String("prefix", "", "Line prefix")
var configPath = flag.String("config", "", "Path to config file")

func main() {
	flag.Parse()

	if *configPath != "" {
		daemonMode()
	} else {
		singleMode()
	}
}

func daemonMode() {
	conf := loadConfig(*configPath)

	for _, w := range conf.Watch {
		go watchFileLoop(w)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}

func watchFileLoop(w Watch) {
	minBackoff := 5 * time.Second
	maxBackoff := 300 * time.Second
	boff := backoff.New(minBackoff, maxBackoff)
	for {
		startTime := time.Now()
		err := watchFile(w)
		log.Printf("watch %s err: %s", w.Path, err)
		if time.Since(startTime) > maxBackoff*5 {
			boff.Reset()
		}
		wait := boff.Next()
		log.Printf("waiting %s", wait)
		time.Sleep(wait)
	}
}

func watchFile(w Watch) error {
	var (
		in       io.Reader
		tailMode bool
	)

	if w.Path == "" {
		panic("path must be set.")
	} else if w.Path == "-" {
		in = os.Stdin
	} else {
		f, err := os.Open(w.Path)
		if err != nil {
			return err
		}
		f.Seek(0, io.SeekEnd)
		tailMode = true
		in = f
	}

	r := bufio.NewReader(in)

	if w.Prefix != "" {
		w.Prefix = strings.TrimSpace(w.Prefix)
	}

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			if tailMode {
				time.Sleep(5 * time.Second)
				continue
			} else {
				return err
			}
		} else if err != nil {
			return err
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if w.Prefix != "" {
			line = w.Prefix + " " + line
		}

		fmt.Println(line)

		err = slack.PostWebhook(w.HookUrl, &slack.WebhookMessage{
			Text: line,
		})
		if err != nil {
			return err
		}
	}
}

func singleMode() {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("Must set SLACK_WEBHOOK_URL in environment")
	}

	w := Watch{
		Path:    "-",
		Prefix:  *prefix,
		HookUrl: webhookURL,
	}

	err := watchFile(w)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
}
