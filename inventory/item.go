package inventory

import (
	"encoding/json"
	"github.com/lietu/stream-manager/database"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

var ITEM_COLLECTION = "inventory_items"

type Item struct {
	Id                bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	Available         int           `json:"available"`
	Reserved          int           `json:"reserved"`
	PriceCents        int           `json:"priceCents"`
	Interval          string        `json:"interval"`
	LastPurchase      string        `json:"-"`
	NotAvailableUntil string        `json:"notAvailableUntil"`
}

func (i *Item) IsAvailable(count int) bool {
	blockedUntil := i.GetBlockedUntil()
	now := time.Now()

	if now.Before(blockedUntil) {
		return false
	}

	if i.Available < count {
		return false
	}

	return true
}

func (i *Item) GetBlockedUntil() time.Time {
	t := time.Time{}

	if i.NotAvailableUntil != "" {
		_t, err := time.Parse(time.RFC3339, i.NotAvailableUntil)
		if err != nil {
			log.Printf("Failed to parse Item.NotAvailableUntil %s: %s", i.NotAvailableUntil, err)
			return t
		}
		t = _t
	}

	return t
}

func (i *Item) getDuration() time.Duration {
	d, err := time.ParseDuration(i.Interval)
	if err != nil {
		log.Printf("Failed to parse duration %s: %s", i.Interval, err)
	}
	return d
}

func (i *Item) reserve(inv *Inventory, userId bson.ObjectId, count int) (result bool, msg string, reservationId *bson.ObjectId) {
	result = false

	if !i.IsAvailable(count) {
		msg = "Item not available in the desired amount"
		return
	}

	i.Available -= count
	i.Reserved += count

	reservationId = inv.saveReservation(i.Id, userId, count)
	if reservationId == nil {
		msg = "Failed to reserve the desired amount"
		return
	}

	i.Save()

	result = true
	msg = "Item reserved"
	return
}

func (i *Item) CompleteOrder(req *CompleteOrderRequest) *CompleteOrderResult {
	db := database.GetDB()
	cor := &CompleteOrderResult{}
	cor.Ok = false
	cor.OrderId = nil

	r := Reservation{}
	err := db.C(RESERVATION_COLLECTION).FindId(req.ReservationId).One(&r)
	if err != nil {
		log.Printf("Could not find reservation %s", req.ReservationId.Hex())
		if !i.IsAvailable(req.Count) {
			log.Printf("Item %s is not available in desired quantity anymore", i.Id.Hex())
			return cor
		}
	}

	o := Order{}
	o.Id = bson.NewObjectId()
	o.UserId = req.UserId
	o.ItemId = req.ItemId
	o.Count = req.Count
	o.TotalPriceCents = i.PriceCents * o.Count
	o.Timestamp = time.Now().Format(time.RFC3339)

	if o.Save() {
		cor.OrderId = &o.Id
		cor.Ok = true

		i.Reserved -= o.Count
		i.NotAvailableUntil = time.Now().Add(i.getDuration()).Format(time.RFC3339)
		i.Save()
	}

	return cor
}

func (i *Item) GetPushData() []byte {
	data, err := json.Marshal(i)
	if err != nil {
		log.Printf("Failed to get push data for item %s: %s", i.Id.Hex(), err)
	}
	return data
}

func (i *Item) Save() {
	db := database.GetDB()
	_, err := db.C(ITEM_COLLECTION).UpsertId(i.Id, i)
	if err != nil {
		log.Printf("Failed to update item %s: %s", i.Id, err)
	}
}
