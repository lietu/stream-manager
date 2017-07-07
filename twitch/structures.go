package twitch

type User struct {
	Id          string `json:"_id"`
	Bio         string `json:"bio"`
	CreatedAt   string `json:"created_at"`
	DisplayName string `json:"display_name"`
	Logo        string `json:"logo"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	UpdatedAt   string `json:"updated_at"`
}

type Follow struct {
	CreatedAt     string `json:"created_at"`
	Notifications bool   `json:"notifications"`
	User          *User  `json:"user"`
}

type Subscription struct {
	CreatedAt     string `json:"created_at"`
	SubPlan		  string `json:"sub_plan"`
	SubPlanName   string `json:"sub_plan_name"`
	User          *User  `json:"user"`
}

type Preview struct {
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
	Template string `json:"template"`
}

type Channel struct {
	Mature                       bool   `json:"mature"`
	Status                       string `json:"status"`
	BroadcasterLanguage          string `json:"broadcaster_language"`
	DisplayName                  string `json:"display_name"`
	Game                         string `json:"game"`
	Language                     string `json:"language"`
	Id                           int    `json:"_id"`
	Name                         string `json:"name"`
	CreatedAt                    string `json:"created_at"`
	UpdatedAt                    string `json:"updated_at"`
	Partner                      bool   `json:"partner"`
	Logo                         string `json:"logo"`
	VideoBanner                  string `json:"video_banner"`
	ProfileBanner                string `json:"profile_banner"`
	ProfileBannerBackgroundColor string `json:"profile_banner_background_color"`
	Url                          string `json:"url"`
	Views                        int    `json:"views"`
	Followers                    int    `json:"followers"`
}

type Stream struct {
	Id          int      `json:"_id"`
	Game        string   `json:"game"`
	Viewers     int      `json:"viewers"`
	VideoHeight int      `json:"video_height"`
	AverageFPS  int      `json:"average_fps"`
	Delay       float32  `json:"delay"`
	CreatedAt   string   `json:"created_at"`
	IsPlaylist  bool     `json:"is_playlist"`
	Preview     *Preview `json:"preview"`
	Channel     *Channel `json:"channel"`
}

type UsersResponse struct {
	Total int     `json:"_total"`
	Users []*User `json:"users"`
}

type FollowsResponse struct {
	Cursor  string    `json:"_cursor"`
	Total   int       `json:"_total"`
	Follows []*Follow `json:"follows"`
}

type SubscriptionsResponse struct {
	Cursor  string    `json:"_cursor"`
	Total   int       `json:"_total"`
	Subscriptions []*Subscription `json:"subscriptions"`
}

type StreamResponse struct {
	Stream *Stream `json:"stream"`
}

type PubSubIncoming struct {
	Type string `json:"type"`
}

type PubSubResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
	Nonce string `json:"nonce"`
}

type PubSubMessageData struct {
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type PubSubMessage struct {
	Type string            `json:"type"`
	Data PubSubMessageData `json:"data"`
}

type BadgeEntitlement struct {
	NewVersion      int `json:"new_version"`
	PreviousVersion int `json:"previous_version"`
}

/*
{
	"data": {
		"user_name":       "dallasnchains",
		"channel_name":    "dallas",
		"user_id":         "129454141",
		"channel_id":      "44322889",
		"time":            "2017-02-09T13:23:58.168Z",
		"chat_message":    "cheer10000 New badge hype!",
		"bits_used":       10000,
		"total_bits_used": 25000,
		"context":         "cheer",
		"badge_entitlement": {
			"new_version":      25000,
			"previous_version": 10000
		}
	},
	"version":      "1.0",
	"message_type": "bits_event",
	"message_id":   "8145728a4-35f0-4cf7-9dc0-f2ef24de1eb6"
}
*/

type PubSubBitsData struct {
	BadgeEntitlement BadgeEntitlement `json:"badge_entitlement"`
	BitsUsed         int              `json:"bits_used"`
	ChannelId        string           `json:"channel_id"`
	ChannelName      string           `json:"channel_name"`
	ChatMessage      string           `json:"chat_message"`
	Context          string           `json:"context"`
	Time             string           `json:"time"`
	TotalBitsUsed    int              `json:"total_bits_used"`
	UserId           string           `json:"user_id"`
	UserName         string           `json:"user_name"`
}

type PubSubBits struct {
	Data        PubSubBitsData `json:"data"`
	MessageId   string         `json:"message_id"`
	MessageType string         `json:"message_type"`
	Version     string         `json:"version"`
}

type BitsImage map[string]map[string]string

type BitsTier struct {
	Color   string               `json:"color"`
	Id      string               `json:"id"`
	Images  map[string]BitsImage `json:"images"`
	MinBits int                  `json:"min_bits"`
}

type BitsAction struct {
	Backgrounds []string   `json:"backgrounds"`
	Prefix      string     `json:"prefix"`
	Scales      []string   `json:"scales"`
	States      []string   `json:"states"`
	Tiers       []BitsTier `json:"tiers"`
}

type BitsActions struct {
	Type    string `json:"type"`
	Actions []BitsAction `json:"actions"`
}
