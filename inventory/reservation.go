package inventory

import (
	"gopkg.in/mgo.v2/bson"
)

var RESERVATION_COLLECTION = "inventory_reservations"

type ReservationResult struct {
	Ok  bool
	Msg string
	Id  *bson.ObjectId
}

type ReservationRequest struct {
	ItemId bson.ObjectId
	UserId bson.ObjectId
	Count  int
	Result chan *ReservationResult
}

type Reservation struct {
	Id     bson.ObjectId `bson:"_id,omitempty"`
	ItemId bson.ObjectId
	UserId bson.ObjectId
	Count  int
}

func NewReservationRequest(itemId bson.ObjectId, userId bson.ObjectId, count int) *ReservationRequest {
	r := ReservationRequest{}
	r.ItemId = itemId
	r.UserId = userId
	r.Count = count
	r.Result = make(chan *ReservationResult)

	return &r
}
