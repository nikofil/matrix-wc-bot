package bot

import (
	"fmt"

	"maunium.net/go/mautrix/id"
)

type cryptoLog struct{}

func (cryptoLog) Error(message string, args ...interface{}) {
	fmt.Printf("[ERROR] "+message+"\n", args...)
}

func (cryptoLog) Warn(message string, args ...interface{}) {
	fmt.Printf("[Warn] "+message+"\n", args...)
}

func (cryptoLog) Debug(message string, args ...interface{}) {
	fmt.Printf("[debug] "+message+"\n", args...)
}

func (cryptoLog) Trace(message string, args ...interface{}) {
	fmt.Printf("[trace] "+message+"\n", args...)
}

type roomCache struct {
	roomMembers map[id.RoomID][]id.UserID
	roomIsEnc   map[id.RoomID]bool
}

func (rc *roomCache) SetRoomState(roomMemb map[id.RoomID][]id.UserID, roomEnc map[id.RoomID]bool) {
	rc.roomMembers = roomMemb
	rc.roomIsEnc = roomEnc
}

func (rc *roomCache) IsEncrypted(rid id.RoomID) bool {
	if res, ok := rc.roomIsEnc[rid]; ok {
		return res
	}
	return false
}

func (rc *roomCache) FindSharedRooms(uid id.UserID) []id.RoomID {
	rooms := make([]id.RoomID, 0)
	for rid, membs := range rc.roomMembers {
		for _, memb := range membs {
			if memb == uid {
				rooms = append(rooms, rid)
				break
			}
		}
	}
	return rooms
}

func (rc *roomCache) RoomMembers(rid id.RoomID) []id.UserID {
	return rc.roomMembers[rid]
}
