package inventory

import (
	"github.com/lietu/stream-manager/database"
	"gopkg.in/mgo.v2/bson"
	"log"
	"testing"
	"time"
)

func TestCancellation(t *testing.T) {
	database.SetTestMode()
	db := database.GetDB()
	i := Item{}
	i.Id = bson.NewObjectId()
	i.Name = "Test item"
	i.Description = "Test description"
	i.Available = 2
	i.Reserved = 0
	i.NotAvailableUntil = "1985-01-01T00:00:00.00Z"
	i.Save()

	defer func() {
		db.C(ITEM_COLLECTION).RemoveId(i.Id)
	}()

	userId := bson.NewObjectId()

	inv := newInventory()
	inv.reservationDuration = time.Millisecond * 10
	go inv.Run()
	defer func() {
		inv.Stop()
	}()

	rr := inv.MakeReservation(i.Id, userId, 1)
	if !rr.Ok {
		log.Panicf("Reservation failed: %s", rr.Msg)
	}

	defer func() {
		db.C(RESERVATION_COLLECTION).RemoveId(rr.Id)
	}()

	i2 := inv.GetItem(i.Id)

	if i2.Reserved != 1 {
		log.Panic("Item did not get reserved")
	}

	if i2.Available != i.Available-1 {
		log.Panic("Item availability did not change")
	}

	time.Sleep(inv.reservationDuration * 2)

	i3 := inv.GetItem(i.Id)

	if i3.Reserved != 0 {
		log.Panic("Reservation did not get cancelled")
	}

	if i3.Available != i.Available {
		log.Panic("Item availability did not get reset")
	}
}
