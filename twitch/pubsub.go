package twitch

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"strings"
	"time"
)

func (tss *TwitchStreamService) reconnectToPubSub() {
	if tss.pubsub != nil {
		tss.pubsub.Close()
	}
	tss.pubsub = nil
	log.Print("Reconnecting to Twitch PubSub in 1 second")
	go func() {
		// Reconnect after a while
		time.Sleep(time.Second * 1)
		tss.connectToPubSub()
	}()
}

func (tss *TwitchStreamService) connectToPubSub() {
	u := url.URL{Scheme: "wss", Host: "pubsub-edge.twitch.tv", Path: "/"}
	log.Printf("Connecting to Twitch PubSub at %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// TODO: Better handling
		log.Printf("Connecting to Twitch PubSub failed: %s", err)
		tss.reconnectToPubSub()
		return
	}

	tss.pubsub = c

	ping := true
	go func() {
		for ping {
			tss.sendPubSubMessage(map[string]interface{}{
				"type": "PING",
			})
			time.Sleep(time.Minute * 4)
		}
	}()

	go func() {
		defer tss.pubsub.Close()
		defer func() {
			ping = false
		}()

		log.Print("Connected to Twitch PubSub.")

		for _, channel := range tss.pubsubTopics {
			tss.listenPubSubTopic(channel)
		}

		for tss.running {
			_, message, err := c.ReadMessage()

			if err != nil {
				log.Printf("PubSub socket error: %s", err)
				tss.reconnectToPubSub()
				return
			}

			tss.handlePubSubMessage(message)
		}
	}()
}

func (tss *TwitchStreamService) listenPubSubTopic(topic string) {
	if tss.pubsub == nil {
		for _, t := range tss.pubsubTopics {
			if t == topic {
				return
			}
		}

		tss.pubsubTopics = append(tss.pubsubTopics, topic)
		return
	}

	data := map[string]interface{}{
		"topics":     []string{topic},
		"auth_token": tss.manager.Config.Twitch.OAuthToken,
	}

	request := map[string]interface{}{
		"type": "LISTEN",
		"data": data,
	}

	tss.sendPubSubMessage(request)
}

func (tss *TwitchStreamService) sendPubSubMessage(msg map[string]interface{}) {
	jsonOut, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	data, _ := msg["data"].(map[string]interface{})
	if _, ok := data["auth_token"]; ok {
		data["auth_token"] = "..."
	}

	jsonPrintable, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	log.Printf("PUBSUB> %s", jsonPrintable)
	tss.pubsub.WriteMessage(websocket.TextMessage, []byte(jsonOut))
}

func (tss *TwitchStreamService) handlePubSubMessage(message []byte) {
	log.Printf("PUBSUB< %s", strings.TrimSpace(string(message)))

	i := PubSubIncoming{}
	json.Unmarshal(message, &i)

	if i.Type == "RESPONSE" {
		r := PubSubResponse{}
		json.Unmarshal(message, &r)
		tss.handlePubSubResponse(r)
	} else if i.Type == "RECONNECT" {
		tss.reconnectToPubSub()
	} else if i.Type == "MESSAGE" {
		m := PubSubMessage{}
		json.Unmarshal(message, &m)
		if m.Data.Topic[:23] == "channel-bits-events-v1." {
			b := PubSubBits{}
			json.Unmarshal([]byte(m.Data.Message), &b)
			tss.handleBits(b.Data)
		}
	}
}

func (tss *TwitchStreamService) handlePubSubResponse(response PubSubResponse) {
	if response.Error != "" {
		if response.Error == "ERR_SERVER" {
			tss.reconnectToPubSub()
		} else {
			log.Panicf("Got error from Twitch PubSub: %#v", response)
		}
	} else {
		log.Print("Twitch PubSub acknowledged message.")
	}
}

func (tss *TwitchStreamService) handleBits(bits PubSubBitsData) {
	log.Printf("Got %d bit(s) from %s", bits.BitsUsed, bits.UserName)
	tss.manager.SendBitsNotification(bits.UserName, bits.BitsUsed, bits.ChatMessage)
}
