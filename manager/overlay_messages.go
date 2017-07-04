package manager

import (
	"encoding/json"
	"log"
)

type Message interface {
	ToJson() []byte
}

type Hello struct {
	Type string `json:"type"`
}

func (m *Hello) ToJson() []byte {
	result, err := json.Marshal(&m)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return result
}

func NewHello() Message {
	m := Hello{}
	m.Type = "hello"

	return &m
}

type Follower struct {
	Type    string `json:"type"`
	Service string `json:"service"`
	Name    string `json:"name"`
}

func (m *Follower) ToJson() []byte {
	result, err := json.Marshal(&m)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return result
}

func NewFollower(service string, name string) Message {
	m := Follower{}
	m.Type = "follower"
	m.Service = service
	m.Name = name

	return &m
}

type Subscriber struct {
	Type    string `json:"type"`
	Service string `json:"service"`
	Name    string `json:"name"`
	Tier    string `json:"tier"`
	Months  string `json:"months"`
}

func (m *Subscriber) ToJson() []byte {
	result, err := json.Marshal(&m)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return result
}

func NewSubscriber(service string, name string, tier string, months string) Message {
	m := Subscriber{}
	m.Type = "subscriber"
	m.Service = service
	m.Name = name
	m.Tier = tier
	m.Months = months

	return &m
}

type Host struct {
	Type    string `json:"type"`
	Service string `json:"service"`
	Name    string `json:"name"`
}

func (m *Host) ToJson() []byte {
	result, err := json.Marshal(&m)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return result
}

type Bits struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Bits    int    `json:"bits"`
	Message string `json:"message"`
}

func (m *Bits) ToJson() []byte {
	result, err := json.Marshal(&m)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return result
}

func NewHost(service string, name string) Message {
	m := Host{}
	m.Type = "host"
	m.Service = service
	m.Name = name

	return &m
}

func NewBits(name string, bits int, message string) Message {
	m := Bits{}
	m.Type = "bits"
	m.Name = name
	m.Bits = bits
	m.Message = message

	return &m
}
