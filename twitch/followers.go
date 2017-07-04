package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/lietu/stream-manager/database"
	"github.com/lietu/stream-manager/utils"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

var lastFollowTime = time.Time{}

func (tss *TwitchStreamService) checkFollowers() {
	follows := tss.getLatestFollows()
	maxFollowTime := lastFollowTime

	for _, f := range follows {
		followTime, err := time.Parse(time.RFC3339Nano, f.CreatedAt)

		if err != nil {
			log.Printf("Failed to parse time %s", f.CreatedAt)
			continue
		}

		if followTime.After(lastFollowTime) {
			log.Printf("New follow at %s: %s", f.CreatedAt, f.User.Name)
			tss.manager.SendFollowerNotification("twitch", f.User.Name)

			if followTime.After(maxFollowTime) {
				maxFollowTime = followTime
			}
		}
	}

	if lastFollowTime != maxFollowTime {
		updateLastFollowTime(maxFollowTime)
	}
}

func (tss *TwitchStreamService) getLatestFollows() []*Follow {
	path := fmt.Sprintf("channels/%s/follows?limit=5", tss.channelId)
	res, err := tss.kraken(path)
	if err != nil {
		log.Printf("Failed to fetch latest follows via Kraken API: %s", err)
		return []*Follow{}
	}
	defer res.Body.Close()

	kfr := FollowsResponse{}
	err = json.NewDecoder(res.Body).Decode(&kfr)
	if err != nil {
		log.Printf("Failed to fetch latest follows, JSON error: %s", err)
		log.Print(res.Body)
		return []*Follow{}
	}

	utils.StatusLog("twitch_follower_count", fmt.Sprintf("You have %d followers on Twitch", kfr.Total))
	return kfr.Follows
}

func readLastFollowTime() {
	db := database.GetDB()
	coll := db.C("twitch_stored_values")

	result := StoredLastFollow{}

	err := coll.Find(bson.M{"key": "last_follow_time"}).One(&result)
	if err != nil {
		log.Printf("Failed to get last follow time from DB: %s", err)
		return
	}

	t, err := time.Parse(time.RFC3339Nano, result.Value)
	if err != nil {
		log.Printf("Failed to parse stored last follow time %s as a time: %s", result.Value, err)
		return
	}

	lastFollowTime = t
}

func updateLastFollowTime(value time.Time) {
	db := database.GetDB()
	coll := db.C("twitch_stored_values")

	stored := StoredLastFollow{}
	stored.Key = "last_follow_time"
	stored.Value = value.Format(time.RFC3339)
	_, err := coll.Upsert(bson.M{"key": stored.Key}, stored)
	if err != nil {
		log.Printf("Failed to store last follow time to DB: %s", err)
	}

	lastFollowTime = value
}

func init() {
	RegisterInitFunc(readLastFollowTime)
}
