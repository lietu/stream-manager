package database

import (
	"github.com/lietu/stream-manager/config"
	"gopkg.in/mgo.v2"
	"log"
)

var _session *mgo.Session
var activeDB = "stream_manager"

func ConfigureDB(config *config.Config) {
	log.Printf("Connecting to DB at %s", config.MongoHosts)
	s, err := mgo.Dial(config.MongoHosts)
	if err != nil {
		log.Fatalf("Failed to connect to mongo on %s: %s", config.MongoHosts, err)
	}
	_session = s
}

func SetTestMode() {
	ConfigureDB(config.GetTestConfig())
	activeDB = "stream_manager_test"
}

func GetDB() *mgo.Database {
	if _session == nil {
		log.Panic("Attempt to GetDB before ConfigureDB")
	}

	return _session.DB(activeDB)
}
