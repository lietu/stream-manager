package lametric

type requestFrame struct {
	Text  string `json:"text"`
	Icon  string `json:"icon"`
	Index int `json:"index"`
}

type request struct {
	Frames []requestFrame `json:"frames"`
}
