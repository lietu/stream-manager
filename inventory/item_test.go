package inventory

import (
	"log"
	"testing"
)

func TestAvailability(t *testing.T) {
	i := Item{}
	i.Available = 1
	i.NotAvailableUntil = "1985-04-12T23:20:50.52Z"

	if !i.IsAvailable(1) {
		log.Panicf("Item not available even though it should be after %s", i.NotAvailableUntil)
	}

	c := 2
	if i.IsAvailable(c) {
		log.Panicf("%d item(s) available even though there should only be %d of them", c, i.Available)
	}

	i.NotAvailableUntil = "2099-01-01T00:00:00.00Z"
	if i.IsAvailable(1) {
		log.Panicf("Item available even though it should be unavailable until %s", i.NotAvailableUntil)
	}

	log.Print("TestAvailability complete")
}
