package manager

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type OverlayClient struct {
	connection *websocket.Conn
	request    *http.Request
	manager    *Manager
	connected  bool
	mutex      *sync.Mutex
}

type overlayClientList []*OverlayClient

var connectedOverlayClients = overlayClientList{}
var addOverlayClientCh = make(chan *OverlayClient)
var removeOverlayClientCh = make(chan *OverlayClient)
var getOverlayClientsCh = make(chan chan []*OverlayClient)
var exitOverlayClientHandlerCh = make(chan bool)

// Makes client list handling synchronous
func manageOverlayClientChannels() {
	log.Print("Managing overlay client channels")
	running := true
	for running {
		select {
		case c := <-addOverlayClientCh:
			connectedOverlayClients = append(connectedOverlayClients, c)

		case remove := <-removeOverlayClientCh:
			overlayClients := overlayClientList{}
			for _, c := range connectedOverlayClients {
				if remove != c {
					overlayClients = append(overlayClients, c)
				}
			}
			connectedOverlayClients = overlayClients

		case get := <-getOverlayClientsCh:
			get <- connectedOverlayClients

		case <-exitOverlayClientHandlerCh:
			running = false
		}
	}
	log.Print("Done managing overlay client channels")
}

func getOverlayClients() []*OverlayClient {
	ch := make(chan []*OverlayClient)
	getOverlayClientsCh <- ch
	return <-ch
}

func SendMessageToAllOverlays(message Message) {
	data := message.ToJson()

	log.Printf("Sending to all overlays: %s", data)

	for _, c := range getOverlayClients() {
		c.Send(data)
	}
}

func (c *OverlayClient) SendHello() {
	c.SendMessage(NewHello())
	for _, ss := range c.manager.streamServices {
		ss.WelcomeOverlayClient(c)
	}
}

func (c *OverlayClient) SendMessage(message Message) {
	data := message.ToJson()
	c.Send(data)
}

func (c *OverlayClient) Send(data []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		c.connection.WriteMessage(websocket.TextMessage, data)
	}
}

func (c *OverlayClient) GetRemoteAddr() string {
	return c.request.RemoteAddr
}

func (c *OverlayClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		removeOverlayClientCh <- c
		c.connected = false
		c.connection.Close()
	}
}

func (c *OverlayClient) Handle() {
	c.SendHello()
	addOverlayClientCh <- c
	for {
		_, message, err := c.connection.ReadMessage()

		if err != nil {
			if err.Error() == "websocket: close 1001 " {
				log.Printf("Client from %s disconnected", c.GetRemoteAddr())
			} else {
				log.Printf("Client from %s error: %s", c.GetRemoteAddr(), err.Error())
			}

			c.Close()

			break
		} else {
			log.Printf("Overlay client trying to send unexpected message: %s", message)
			c.Close()
		}
	}
}

// Handler for /events listeners
var upgrader = websocket.Upgrader{}
func (m *Manager) overlayEventHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	log.Printf("Connection from %s", r.RemoteAddr)

	if err != nil {
		log.Print("upgrade: ", err)
		return
	}

	c := newOverlayClient(conn, r, m)
	defer c.Close()
	c.Handle()
}

func newOverlayClient(conn *websocket.Conn, r *http.Request, m *Manager) *OverlayClient {
	c := OverlayClient{}
	c.connection = conn
	c.request = r
	c.connected = true
	c.manager = m
	c.mutex = &sync.Mutex{}

	return &c
}
