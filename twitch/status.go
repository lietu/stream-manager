package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/lietu/stream-manager/utils"
	"log"
)

func (tss *TwitchStreamService) updateStreamStatus() {
	path := fmt.Sprintf("streams/%s", tss.channelId)
	res, err := tss.kraken(path)
	if err != nil {
		log.Printf("Failed to fetch stream status via Kraken API: %s", err)
		return
	}
	defer res.Body.Close()

	ksr := StreamResponse{}
	err = json.NewDecoder(res.Body).Decode(&ksr)
	if err != nil {
		log.Printf("Failed to fetch stream status, JSON parsing failed: %s", err)
		log.Print(res.Body)
		return
	}

	if ksr.Stream == nil {
		utils.StatusLog("twitch_stream_status", fmt.Sprintf("Twitch stream %s is OFFLINE", tss.manager.Config.Twitch.Username))
		tss.live = false
		return
	}

	tss.live = true

	utils.StatusLog("twitch_stream_status", fmt.Sprintf("Twitch stream live since %s with %d viewers", ksr.Stream.CreatedAt, ksr.Stream.Viewers))
}
