package inventory

import (
	"fmt"
	"github.com/lietu/stream-manager/database"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

var RESERVATION_DURATION = time.Minute * 5

type StreamManager interface {
	SendToFrontend(msgType string, data []byte)
}

type Inventory struct {
	db                  *mgo.Database
	reserveCn           chan *ReservationRequest
	cancelReservationCn chan *bson.ObjectId
	completeOrderCn     chan *CompleteOrderRequest
	stopCn              chan bool
	reservationDuration time.Duration
	streamManager       StreamManager
}

func (inv *Inventory) reservationTimeout(reservationId *bson.ObjectId, timeout time.Duration) {
	time.Sleep(timeout)
	inv.cancelReservationCn <- reservationId
}

func (inv *Inventory) Run() {
	inv.SetupIndexes()
	inv.PrintList()

	for {
		select {
		case r := <-inv.reserveCn:
			log.Printf("Got a reservation request for item %s from user %s", r.ItemId.Hex(), r.UserId.Hex())
			item := inv.GetItem(r.ItemId)
			result := false
			msg := ""
			var reservationId *bson.ObjectId = nil

			if item != nil {
				result, msg, reservationId = item.reserve(inv, r.UserId, r.Count)

				if result {
					go inv.reservationTimeout(reservationId, inv.reservationDuration)

					r.Result <- &ReservationResult{
						result,
						msg,
						reservationId,
					}

					inv.pushItem(item)
					continue
				}
			} else {
				result = false
				msg = "Item not found"
			}

			r.Result <- &ReservationResult{
				result,
				msg,
				nil,
			}
		case reservationId := <-inv.cancelReservationCn:
			log.Printf("Got a reservation cancellation request for reservation %s", reservationId.Hex())
			inv.cancelReservation(reservationId)
		case req := <-inv.completeOrderCn:
			item := inv.GetItem(req.ItemId)
			res := item.CompleteOrder(req)
			if res.Ok {
				inv.DeleteReservation(&req.ReservationId)
				inv.pushItem(item)
				inv.reportOrder(res.OrderId)
			}
			req.Result <- res
		case <-inv.stopCn:
			log.Print("Stopping Inventory")
			return
		}
	}

	log.Panic("Quit Inventory loop")
}

func (inv *Inventory) Stop() {
	log.Print("Attempting to stop inventory")
	inv.stopCn <- true
}

func (inv *Inventory) CompleteOrder(req *CompleteOrderRequest) *CompleteOrderResult {
	inv.completeOrderCn <- req
	return <-req.Result
}

func (inv *Inventory) pushItem(i *Item) {
	if inv.streamManager != nil {
		data := i.GetPushData()
		if data != nil {
			inv.streamManager.SendToFrontend("item", data)
		}
	}
}

func (inv *Inventory) reportOrder(orderId *bson.ObjectId) {
	o := LoadOrder(*orderId)
	i := inv.GetItem(o.ItemId)
	log.Printf("Order %s finished for %dx %s (%s)", o.Id.Hex(), o.Count, i.Name, i.Id.Hex())
}

func (inv *Inventory) MakeReservation(itemId bson.ObjectId, userId bson.ObjectId, count int) *ReservationResult {
	rr := NewReservationRequest(itemId, userId, count)
	log.Printf("Sending reservation request for item %s", itemId.Hex())
	inv.reserveCn <- rr
	log.Printf("Waiting for response to reservation request for item %s", itemId.Hex())
	return <-rr.Result
}

func (inv *Inventory) saveReservation(itemId bson.ObjectId, userId bson.ObjectId, count int) *bson.ObjectId {
	r := Reservation{}
	r.Id = bson.NewObjectId()
	r.ItemId = itemId
	r.UserId = userId
	r.Count = count

	err := inv.db.C(RESERVATION_COLLECTION).Insert(r)
	if err != nil {
		log.Printf("Failed to save reservation: %s", err)
		return nil
	}

	return &r.Id
}

func (inv *Inventory) cancelReservation(id *bson.ObjectId) {
	r := Reservation{}
	err := inv.db.C(RESERVATION_COLLECTION).FindId(*id).One(&r)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Printf("No need to cancel with reservation %s, it's already gone.", id.Hex())
		} else {
			log.Printf("Failed to cancel reservation %s: %s", id.Hex(), err)
		}
		return
	}

	item := inv.GetItem(r.ItemId)
	item.Available += r.Count
	item.Reserved -= r.Count
	item.Save()

	inv.DeleteReservation(id)
}

func (inv *Inventory) DeleteReservation(id *bson.ObjectId) {
	err := inv.db.C(RESERVATION_COLLECTION).RemoveId(*id)
	if err != nil {
		log.Printf("Failed to delete reservation %s: %s", id.Hex(), err)
	}
}

func (inv *Inventory) PrintList() {
	items := inv.GetList()
	log.Printf("Inventory consists of %d items", len(items))
	for _, item := range items {
		log.Printf(" - %s (%s): %d available", item.Name, item.Description, item.Available)
	}
}

func (inv *Inventory) GetList() []*Item {
	items := []*Item{}
	err := inv.db.C(ITEM_COLLECTION).Find(nil).All(&items)
	if err != nil {
		fmt.Printf("Failed to get items from DB: %s", err)
	}
	return items
}

func (inv *Inventory) GetItem(itemId bson.ObjectId) *Item {
	item := &Item{}
	err := inv.db.C(ITEM_COLLECTION).FindId(itemId).One(&item)
	if err != nil {
		fmt.Printf("Failed to get item %s from DB: %s", itemId.Hex(), err)
		return nil
	}
	return item
}

func (inv *Inventory) SetupIndexes() {

}

func Start(sm StreamManager) *Inventory {
	manager := newInventory()
	manager.streamManager = sm
	go manager.Run()
	return manager
}

func newInventory() *Inventory {
	i := Inventory{}
	i.db = database.GetDB()
	i.reserveCn = make(chan *ReservationRequest)
	i.cancelReservationCn = make(chan *bson.ObjectId)
	i.completeOrderCn = make(chan *CompleteOrderRequest)
	i.stopCn = make(chan bool)
	i.reservationDuration = RESERVATION_DURATION
	i.streamManager = nil

	return &i
}
