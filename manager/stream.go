package manager

import "log"

type StreamService interface {
	Start()
	Stop()
	WelcomeOverlayClient(client *OverlayClient)
}

type StreamServiceConstructor func(*Manager) StreamService

var registeredServices = map[string]StreamServiceConstructor{}

func RegisterStreamService(name string, constructor StreamServiceConstructor) {
	log.Printf("Registering stream service %s", name)
	registeredServices[name] = constructor
}

func GetStreamService(name string, manager *Manager) StreamService {
	ss := registeredServices[name](manager)
	return ss
}
