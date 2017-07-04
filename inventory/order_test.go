package inventory

import (
	"github.com/lietu/stream-manager/database"
	"gopkg.in/mgo.v2/bson"
	"log"
	"testing"
	"time"
)

func TestCompleteOrder(t *testing.T) {
	database.SetTestMode()
	db := database.GetDB()
	i := Item{}
	i.Id = bson.NewObjectId()
	i.Name = "Test item"
	i.Description = "Test description"
	i.Available = 2
	i.Reserved = 0
	i.NotAvailableUntil = "1985-01-01T00:00:00.00Z"
	i.Interval = "15m"
	i.Save()

	defer func() {
		db.C(ITEM_COLLECTION).RemoveId(i.Id)
	}()

	userId := bson.NewObjectId()

	inv := newInventory()
	go inv.Run()
	defer func() {
		inv.Stop()
	}()

	count := 1
	rr := inv.MakeReservation(i.Id, userId, count)
	if !rr.Ok {
		log.Panicf("Reservation failed: %s", rr.Msg)
	}

	defer func() {
		db.C(RESERVATION_COLLECTION).RemoveId(rr.Id)
	}()

	req := NewFinishOrderRequest(*rr.Id, i.Id, userId, count)
	res := inv.CompleteOrder(req)

	if res.Ok != true {
		log.Panic("Could not finish order")
	}

	defer func() {
		db.C(ORDER_COLLECTION).RemoveId(res.OrderId)
	}()

	i2 := inv.GetItem(i.Id)
	if i2.IsAvailable(1) {
		log.Panic("Item did not get blocked properly after order completion")
	}

	blockedUntil := i2.GetBlockedUntil()
	dmin, _ := time.ParseDuration("14m")
	dmax, _ := time.ParseDuration("16m")
	min := time.Now().Add(dmin)
	max := time.Now().Add(dmax)
	if blockedUntil.Before(min) || blockedUntil.After(max) {
		log.Panicf("Item should be blocked until %s to %s, but is really blocked until %s", min, max, blockedUntil)
	}

	if i2.Available != i.Available-count {
		log.Panic("Item did not get used up properly after order completion")
	}

	if i2.Reserved > 0 {
		log.Panic("Item reservation was not cleared properly after order completion")
	}
}
