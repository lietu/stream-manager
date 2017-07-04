package twitch

import (
	"testing"
	"github.com/kyleterry/tenyks/irc"
)

func TestHostPattern(t *testing.T) {
	strings := []string{
		"lietu2 is now hosting you.",
		"lietu2 is now hosting you for 123 viewers.",
		"lietu2 is now hosting you for up to 321 viewers.",
		"lietu2 is now auto hosting you.",
		"lietu2 is now auto hosting you for 1 viewers.",
		"lietu2 is now auto hosting you for up to 99999 viewers.",
	}

	for _, s := range strings {
		hosted_by := getHostedBy(s)

		if hosted_by != "lietu2" {
			t.Errorf("Did not match host message: %s", s)
		}
	}
}

func TestEmote(t *testing.T) {
	line := "@badges=subscriber/0,bits/100;color=;display-name=lie2;emote-only=1;emotes=217363:0-6;id=d17a8b43-3cd9-45f0-bb9e-e4011471861d;mod=0;room-id=30496684;sent-ts=1499167443515;subscriber=1;tmi-sent-ts=1499167442511;turbo=0;user-id=105047668;user-type= :lie2!lie2@lie2.tmi.twitch.tv PRIVMSG #lietu :lietuYO"
	tags, line := parseTags(line)
	msg := irc.ParseMessage(line)

	if tags["user-id"] != "105047668" {
		t.Errorf("Expected user-id 105047668, got: %s", tags["user-id"])
	}

	if msg.Command != "PRIVMSG" {
		t.Errorf("Expected PRIVMSG, got: %s", msg.Command)
	}
}

func TestSubscription(t *testing.T) {
	line := "@badges=subscriber/1;color=red;display-name=lietu;msg-id=1;msg-param-months=1;msg-param-sub-plan=3000;msg-param-sub-plan-name=$24.99;room-id=1;login=lietu;subscriber=1;system-msg=Lietu just subscribed :tmi.twitch.tv USERNOTICE #lietu :hola como estas"
	tags, line := parseTags(line)
	msg := irc.ParseMessage(line)
	sub := subMap(tags["msg-param-sub-plan"])

	if tags["display-name"] != "lietu" {
		t.Errorf("Expected display-name lietu but got: %s", tags["display-name"])
	}

	if msg.Command != "USERNOTICE" {
		t.Errorf("Expected USERNOTICE command but got: %s", msg.Command)
	}

	if sub != "$24.99" {
		t.Errorf("Expected $24.99 sub but got: %s", sub)
	}
}
