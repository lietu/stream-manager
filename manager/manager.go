package manager

import (
	"fmt"
	"github.com/braintree/manners"
	"github.com/lietu/stream-manager/config"
	"github.com/lietu/stream-manager/database"
	"github.com/lietu/stream-manager/inventory"
	"github.com/lietu/stream-manager/storage"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"github.com/lietu/stream-manager/lametric"
	"sync"
	"time"
)

type Manager struct {
	Config         *config.Config
	Inventory      *inventory.Inventory
	streamServices []StreamService
	lametric       *lametric.LaMetric
	WebUIAddress   string
	WebUI		   *http.ServeMux
}

const testOverlays = false

func random(min, max int) int {
	return rand.Intn(max - min) + min
}

func (m *Manager) Start() {
	log.Print("Starting Stream Manager...")

	// Listen to CTRL+C
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	// Load config, connect to DB
	m.Config = config.LoadConfig()
	database.ConfigureDB(m.Config)

	// Start the inventory subsystem
	m.Inventory = inventory.Start(m)

	// Set up stream services (Twitch)
	for _, name := range m.Config.StreamServices {
		ss := GetStreamService(name, m)
		m.streamServices = append(m.streamServices, ss)
	}

	// Initialize LaMetric usage
	m.lametric = lametric.New(m.Config.LaMetric)

	// And the WebUI mux so stream services can register their handlers
	m.WebUIAddress = fmt.Sprintf("%s:%d", m.Config.WebUI.Host, m.Config.WebUI.Port)
	m.WebUI = http.NewServeMux()

	// Start all the stream subsystems
	log.Print("Starting stream services")
	for _, ss := range m.streamServices {
		ss.Start()
	}

	// Set up WebUI HTTP server
	webUIServer := manners.NewWithServer(&http.Server{
		Addr: m.WebUIAddress,
		Handler: m.WebUI,
	})

	// Set up the overlay HTTP handlers
	overlayAddress := fmt.Sprintf("%s:%d", m.Config.Overlay.Host, m.Config.Overlay.Port)
	overlayServeMux := storage.ConfigureOverlayHTTP(m.Config)
	overlayServeMux.HandleFunc("/events", m.overlayEventHandler)
	overlayServer := manners.NewWithServer(&http.Server{
		Addr:    overlayAddress,
		Handler: overlayServeMux,
	})

	// Make sure we have something to manage the overlay client connections
	go manageOverlayClientChannels()

	// Run the HTTP servers
	go func() {
		log.Printf("Starting WebUI server at %s", m.WebUIAddress)
		err := webUIServer.ListenAndServe()

		if err != nil {
			log.Printf("WebUI server exited with an error: %s", err)
		} else {
			log.Print("WebUI server exited cleanly")
		}
	}()

	go func() {
		log.Printf("Starting overlay server at %s", overlayAddress)
		err := overlayServer.ListenAndServe()

		if err != nil {
			log.Printf("Overlay server exited with an error: %s", err)
		} else {
			log.Print("Overlay server exited cleanly")
		}
	}()

	// Generate bunch of notifications if we want to test the overlays
	if testOverlays {
		wait := time.Millisecond * 500
		rand.Seed(time.Now().Unix())
		go func() {
			for {
				/*
				<-time.After(wait)
				m.SendHostNotification("twitch", "lieturd")

				<-time.After(wait)
				m.SendBitsNotification("liepoop", 110, "cheer10 cheer100")

				<-time.After(wait)
				m.SendHostNotification("twitch", "lietu")

				<-time.After(wait)
				m.SendSubscriberNotification("twitch", "lietu", "$4.99", "1")
				*/

				<-time.After(wait)
				r := random(1, 100)
				if r < 3 {
					SendMessageToAllOverlays(NewSubscriber("twitch", "lietu", "$4.99", "1"))
					continue
				} else if r < 30 {
					SendMessageToAllOverlays(NewHost("twitch", "lietu"))
					continue
				} else if r < 60 {
					SendMessageToAllOverlays(NewFollower("twitch", "lietu"))
					continue
				}

				SendMessageToAllOverlays(NewBits("liepoop", 110, "cheer10 cheer100"))
			}
		}()
	}

	// Wait for CTRL+C
	log.Println("Waiting for signal")
	<-exit

	// Tell client connection handler to stop
	exitOverlayClientHandlerCh <- true

	// TODO: Blocking close maybe?
	log.Println("Closing overlay server")
	overlayServer.Close()
	webUIServer.Close()

	log.Printf("Closing %d stream services", len(m.streamServices))
	for _, ss := range m.streamServices {
		ss.Stop()
	}
}

func (m *Manager) SendToFrontend(msgType string, data []byte) {
	// Who even knows what this was supposed to do at this point...
}

func (m *Manager) SendFollowerNotification(service string, name string) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		SendMessageToAllOverlays(NewFollower(service, name))
		wg.Done()
	}()
	go func() {
		m.lametric.Follower(service, name)
		wg.Done()
	}()

	wg.Wait()
}

func (m *Manager) SendSubscriberNotification(service string, name string, tier string, months string) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		SendMessageToAllOverlays(NewSubscriber(service, name, tier, months))
		wg.Done()
	}()
	go func() {
		m.lametric.Subscriber(service, name, tier, months)
		wg.Done()
	}()

	wg.Wait()
}

func (m *Manager) SendBitsNotification(name string, bits int, message string) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		go SendMessageToAllOverlays(NewBits(name, bits, message))
		wg.Done()
	}()
	go func() {
		go m.lametric.Bits(name, bits, message)
		wg.Done()
	}()

	wg.Wait()
}

func (m *Manager) SendHostNotification(service string, name string) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		go SendMessageToAllOverlays(NewHost(service, name))
		wg.Done()
	}()
	go func() {
		go m.lametric.Host(service, name)
		wg.Done()
	}()

	wg.Wait()
}

func NewManager() *Manager {
	m := &Manager{}
	return m
}
