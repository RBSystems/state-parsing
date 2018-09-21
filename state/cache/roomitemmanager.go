package cache

import (
	"time"

	"github.com/byuoitav/common/nerr"
	sd "github.com/byuoitav/state-parser/state/statedefinition"
)

/*
Room Item Manager handles managing access to a single room in a cache. Changes to the room are submitted via the IncomingWriteChan and reads are submitted via the IncomingReadChan.
*/
type RoomItemManager struct {
	WriteRequests chan RoomTransactionRequest //channel to buffer changes to the room.
	ReadRequests  chan chan sd.StaticRoom
}

//RoomTransactionRequest is submitted to read/write a the room being managed by this manager
type RoomTransactionRequest struct {
	ResponseChan chan RoomTransactionResponse
	MergeRoom    sd.StaticRoom
}

type RoomTransactionResponse struct {
	Changes bool          //if the Transaction Request resulted in changes
	NewRoom sd.StaticRoom //the updated room with the changes included in the Transaction request included
	Error   *nerr.E       //if there were errors
}

func GetNewRoomManager(id string) RoomItemManager {
	a := RoomItemManager{
		WriteRequests: make(chan RoomTransactionRequest, 100),
		ReadRequests:  make(chan chan sd.StaticRoom, 100),
	}

	room := sd.StaticRoom{RoomID: id, UpdateTimes: make(map[string]time.Time)}
	go StartRoomManager(a, room)
	return a
}

func StartRoomManager(m RoomItemManager, room sd.StaticRoom) {

	var merged sd.StaticRoom
	var changes bool
	var err *nerr.E

	for {
		select {
		case write := <-m.WriteRequests:

			if write.MergeRoom.RoomID != room.RoomID {
				write.ResponseChan <- RoomTransactionResponse{Error: nerr.Create("Can't chagne the ID of a room", "invalid-operation"), NewRoom: room, Changes: false}

			}
			_, merged, changes, err = sd.CompareRooms(room, write.MergeRoom)

			if err != nil {
				write.ResponseChan <- RoomTransactionResponse{Error: err, Changes: false}
				continue
			}

			if changes {
				//only reassign if we have to
				room = merged
			}

			write.ResponseChan <- RoomTransactionResponse{Error: err, NewRoom: room, Changes: changes}

		case read := <-m.ReadRequests:
			//just send it back
			read <- room
		}
	}
}
