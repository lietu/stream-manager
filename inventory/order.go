package inventory

import (
	"fmt"
	"github.com/lietu/stream-manager/database"
	"gopkg.in/mgo.v2/bson"
	"log"
)

var ORDER_COLLECTION = "orders"

type CompleteOrderRequest struct {
	ItemId        bson.ObjectId
	UserId        bson.ObjectId
	Count         int
	ReservationId bson.ObjectId
	Result        chan *CompleteOrderResult
}

type CompleteOrderResult struct {
	Ok      bool
	OrderId *bson.ObjectId
}

type Order struct {
	Id              bson.ObjectId `bson:"_id,omitempty"`
	UserId          bson.ObjectId
	ItemId          bson.ObjectId
	Count           int
	TotalPriceCents int
	Timestamp       string
}

func (o *Order) Save() bool {
	db := database.GetDB()
	_, err := db.C(ORDER_COLLECTION).UpsertId(o.Id, o)
	if err != nil {
		log.Printf("Failed to update item %s: %s", o.Id, err)
		return false
	}
	return true
}

func LoadOrder(orderId bson.ObjectId) *Order {
	db := database.GetDB()
	order := &Order{}
	err := db.C(ORDER_COLLECTION).FindId(orderId).One(&order)
	if err != nil {
		fmt.Printf("Failed to get order %s from DB: %s", orderId.Hex(), err)
		return nil
	}
	return order
}

func NewFinishOrderRequest(reservationId bson.ObjectId, itemId bson.ObjectId, userId bson.ObjectId, count int) *CompleteOrderRequest {
	req := CompleteOrderRequest{}
	req.ReservationId = reservationId
	req.ItemId = itemId
	req.UserId = userId
	req.Count = count
	req.Result = make(chan *CompleteOrderResult)

	return &req
}
