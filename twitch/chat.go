package twitch

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kyleterry/tenyks/irc"
	"log"
	"net/url"
	"regexp"
	"time"
	"strings"
)

// Regular Expression to match any and all of the following:
//
// {user} is now hosting you.
// {user} is now hosting you for {count} viewers.
// {user} is now hosting you for up to {count} viewers.
// {user} is now auto hosting you.
// {user} is now auto hosting you for {count} viewers.
// {user} is now auto hosting you for up to {count} viewers.
//
// (Fuck you Twitch for such incredible amount of inconsistency)
var hostRe = regexp.MustCompile("^(?P<user>[a-zA-Z0-9_-]+) is now (?:hosting you|auto hosting you)(?: for(?: up to)? [0-9]+ viewers)?\\.$")

func (tss *TwitchStreamService) reconnectToChat() {
	log.Print("Reconnecting to Twitch chat in 5 seconds")
	go func() {
		// Reconnect after a while
		time.Sleep(time.Second * 5)
		tss.connectToChat()
	}()
}

func (tss *TwitchStreamService) connectToChat() {
	u := url.URL{Scheme: "wss", Host: "irc-ws.chat.twitch.tv", Path: "/"}
	log.Printf("Connecting to Twitch chat at %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// TODO: Better handling
		log.Printf("Connecting to Twitch chat failed: %s", err)
		tss.reconnectToChat()
		return
	}

	tss.chat = c

	go func() {
		defer tss.chat.Close()

		log.Print("Connected to Twitch chat.")
		username := tss.manager.Config.Twitch.Username

		// Tell the server that we know IRC
		tss.sendChat(fmt.Sprintf("PASS oauth:%s", tss.manager.Config.Twitch.OAuthToken))
		tss.sendChat(fmt.Sprintf("NICK %s", username))
		tss.sendChat(fmt.Sprintf("JOIN #%s", username))

		// And specifically the Twitch proprietary things (gives us HOSTTARGET etc.)
		// https://dev.twitch.tv/docs/v5/guides/irc/#twitch-irc-capability-commands
		tss.sendChat("CAP REQ :twitch.tv/tags")
		tss.sendChat("CAP REQ :twitch.tv/commands")

		separator := []byte("\r\n")
		for tss.running {
			_, message, err := c.ReadMessage()

			if err != nil {
				log.Printf("Chat socket error: %s", err)
				tss.reconnectToChat()
				return
			}

			for _, line := range bytes.Split(message, separator) {
				if len(line) > 0 {
					tss.handleChatMessage(string(line))
				}
			}
		}
	}()
}

func (tss *TwitchStreamService) sendChat(msg string) {
	log_msg := msg
	if len(msg) >= 12 && msg[:11] == "PASS oauth:" {
		log_msg = "PASS oauth:..."
	}
	log.Printf("CHAT> %s", log_msg)
	tss.chat.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Parse the Twitch "tags" capability stuff out of the IRC line so the IRC message handler understands the messages
func parseTags(input string) (tags map[string]string, line string) {
	line = input

	if line[:8] == "@badges=" {
		index := strings.Index(line, " :")

		if index < 0 {
			return
		}

		tagdata := line[1:index]
		line = line[index+1:]

		tags = map[string]string{}
		for _, v := range strings.Split(tagdata, ";") {
			data := strings.SplitN(v, "=", 2)
			key, value := data[0], data[1]
			tags[key] = value
		}
	}

	return
}

func subMap(plan string) string {
	if plan == "1000" {
		return "$4.99"
	} else if plan == "2000" {
		return "$9.99"
	} else if plan == "3000" {
		return "$24.99"
	}

	return plan
}

func (tss *TwitchStreamService) handleChatMessage(line string) {
	tags, line := parseTags(line)
	msg := irc.ParseMessage(line)

	if msg == nil {
		log.Printf("Failed to parse chat message: %s", line)
	} else if msg.Command == "PING" {
		log.Printf("CHAT< %s", msg.RawMsg)
		tss.sendChat(fmt.Sprintf("PONG :%s", msg.Trail))
	} else if msg.Command == "USERNOTICE" {
		// TODO: Think about moving this to the PubSub system similar to bits
		tier := subMap(tags["msg-param-sub-plan"])
		login := tags["login"]
		months := tags["msg-param-months"]

		log.Printf("%s did a %s sub for %s months in a row", login, tier, months)
		tss.manager.SendSubscriberNotification("twitch", login, tier, months)
	} else if msg.Command == "PRIVMSG" {
		if msg.Nick == "jtv" {
			tss.handleJTVMessage(msg)
		} else {
			log.Printf("<%s> %s", msg.Nick, msg.Trail)
		}
	} else if msg.Command == "HOSTTARGET" {
		log.Printf("CHAT< %s", msg.RawMsg)
		// Your channel is hosting someone else

		// msg.Trail is either
		// {target} {viewerCount}
		// - {viewerCount}
	} else {
		log.Printf("Unknown chat message: %s", msg.RawMsg)
	}
}

func getHostedBy(message string) string {
	if matches := hostRe.FindStringSubmatch(message); len(matches) > 0 {
		return matches[1]
	}
	return ""
}

func (tss *TwitchStreamService) handleJTVMessage(msg *irc.Message) {
	hosted_by := getHostedBy(msg.Trail)
	if hosted_by != "" {
		log.Printf("Being hosted by %s", hosted_by)
		tss.manager.SendHostNotification("twitch", hosted_by)
	} else {
		log.Printf("<JTV> %s", msg.Trail)
	}
}
