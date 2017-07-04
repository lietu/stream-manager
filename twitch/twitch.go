package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lietu/stream-manager/manager"
	"log"
	"net/http"
	"time"
)

type InitCallback func()

type TwitchStreamService struct {
	bitsActions  BitsActions
	manager      *manager.Manager
	live         bool
	channelId    string
	running      bool
	exited       chan bool
	chat         *websocket.Conn
	pubsub       *websocket.Conn
	pubsubTopics []string
}

var initCallbacks = []InitCallback{}

func RegisterInitFunc(callback InitCallback) {
	initCallbacks = append(initCallbacks, callback)
}

func (tss *TwitchStreamService) Init() {
	for _, c := range initCallbacks {
		c()
	}
}

func (tss *TwitchStreamService) Start() {
	tss.Init()
	log.Print("Starting TwitchStreamService")
	tss.running = true

	tss.fetchChannelId()
	tss.fetchBitsActions()

	go func() {
		log.Print("Running Twitch stream service.")
		tss.connectToChat()

		tss.listenPubSubTopic("channel-bits-events-v1." + tss.channelId)
		tss.connectToPubSub()

		statusCheck := time.NewTicker(time.Second * 10)
		followCheck := time.NewTicker(time.Second * 3)
		defer followCheck.Stop()
		defer statusCheck.Stop()

		for tss.running {
			select {
			case <-statusCheck.C:
				tss.updateStreamStatus()
			case <-followCheck.C:
				//if tss.live {
				tss.checkFollowers()
				//}

			}
		}

		tss.exited <- true
	}()
}

func (tss *TwitchStreamService) Stop() {
	log.Print("Stopping TwitchStreamService")
	tss.running = false
	<-tss.exited
}

func (tss *TwitchStreamService) WelcomeOverlayClient(c *manager.OverlayClient) {
	data, err := json.Marshal(tss.bitsActions)
	if err != nil {
		panic(err)
	}
	c.Send(data)
}

func (tss *TwitchStreamService) fetchChannelId() {
	username := tss.manager.Config.Twitch.Username
	path := fmt.Sprintf("users/?login=%s", username)
	res, err := tss.kraken(path)
	if err != nil {
		// TODO: Better handling plz
		panic(err)
	}
	defer res.Body.Close()

	kur := UsersResponse{}
	err = json.NewDecoder(res.Body).Decode(&kur)
	if err != nil {
		// TODO: Better handling plz
		panic(err)
	}

	tss.channelId = kur.Users[0].Id

	log.Printf("Determined Twitch user id for %s to be %s", username, tss.channelId)
}

func (tss *TwitchStreamService) fetchBitsActions() {
	path := fmt.Sprintf("bits/actions?channel_id=%s", tss.channelId)
	res, err := tss.kraken(path)
	if err != nil {
		// TODO: Better handling plz
		panic(err)
	}
	defer res.Body.Close()
	bitsActions := BitsActions{}
	err = json.NewDecoder(res.Body).Decode(&bitsActions)
	if err != nil {
		// TODO: Better handling plz
		panic(err)
	}

	bitsActions.Type = "bits_actions"
	tss.bitsActions = bitsActions

	log.Print("Determined Twitch Bits Actions")
}

func (tss *TwitchStreamService) kraken(path string) (*http.Response, error) {
	c := &http.Client{}

	url := fmt.Sprintf("https://api.twitch.tv/kraken/%s", path)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Set("Client-ID", tss.manager.Config.Twitch.ClientId)
	// TODO: Consider sending 'Authorization: OAuth ...'

	return c.Do(req)
}

func NewTwitchStreamService(manager *manager.Manager) manager.StreamService {
	tss := &TwitchStreamService{}
	tss.manager = manager
	tss.running = false
	tss.live = false
	tss.exited = make(chan bool)
	tss.pubsubTopics = []string{}
	return tss
}

func init() {
	manager.RegisterStreamService("twitch", NewTwitchStreamService)
}
