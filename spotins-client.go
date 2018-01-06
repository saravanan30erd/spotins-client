package main

import (
	"bytes"
	"encoding/json"
	"github.com/ashwanthkumar/slack-go-webhook"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	// Check the instance(spot) termination status in every 5 seconds
	secs = 5
)

func slackNotify(message string, quit chan int) {
	host, _ := os.Hostname()
	attachment1 := slack.Attachment{}
	attachment1.AddField(slack.Field{Title: "Host", Value: host}).AddField(slack.Field{Title: "Message", Value: message})
	payload := slack.Payload{
		Text:        "Notification from Spot Instances",
		Username:    host,
		IconEmoji:   ":red_circle:",
		Attachments: []slack.Attachment{attachment1},
	}
	slackUrl := os.Getenv("SLACK_URL")
	err := slack.Send(slackUrl, "", payload)
	if len(err) == 0 {
		quit <- 1
	}
}

func getStat(quit chan int) {
	resp, _ := http.Get("http://169.254.169.254/latest/meta-data/spot/termination-time")
	if resp.StatusCode == 200 {
		instanceID, _ := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
		data, _ := ioutil.ReadAll(instanceID.Body)
		insID := map[string]string{"InstanceID": string(data)}
		jsonData, _ := json.Marshal(insID)
		req, _ := http.NewRequest("POST", "http://127.0.0.1:5000/instance/notify", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: time.Second * 3}
		_, err := client.Do(req)
		if err != nil {
			//slack notification
			slackNotify("Spot Instance termination notification failed", quit)
		} else {
			quit <- 1
		}
	}
}

func main() {
	quit := make(chan int, 1)
	timer := time.NewTicker(time.Duration(secs) * time.Second)
	for {
		select {
		case <-timer.C:
			go getStat(quit)
		case <-quit:
			return
		}
	}
}
