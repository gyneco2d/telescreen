package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type Config struct {
	BotUserToken string `json:"botUserToken"`
	AnnounceChannelID string `json:"announceChannelID"`
}

type Request struct {
	Token string `json:"token"`
	Challenge string `json:"challenge"`
	EventType string `json:"type"`
}

var (
	botId	string
	botName	string
)

func run(api *slack.Client, config Config) int {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()

		var reqBody Request
		json.Unmarshal([]byte(body), &reqBody)

		eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: reqBody.Token}))
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("[Error] ", e)
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
			log.Println("URLVerification")
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				userInfo, err := api.GetUserInfo(ev.User)
				if err != nil {
					fmt.Println(err)
				}

				name := userInfo.Profile.DisplayName
				if name == "" {
					name = userInfo.Profile.RealName
				}
				ts, err := ev.EventTimeStamp.Float64()
				if err != nil {
					fmt.Println(err)
				}
				sec, dec := math.Modf(ts)
				timestamp := time.Unix(int64(sec), int64(dec*(1e9)))
				const layout = "2006-01-02 15:04:05"
				text := "[" + timestamp.Format(layout) + "] " + name + ": " + ev.Text
				api.PostMessage(config.AnnounceChannelID, slack.MsgOptionText(text, false))
			}
		}
	})
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)

	return 0
}

func main() {
	jsonString, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	config := new(Config)
	err = json.Unmarshal(jsonString, config)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	api := slack.New(config.BotUserToken)
	os.Exit(run(api, *config))
}
