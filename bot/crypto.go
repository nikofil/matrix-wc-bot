package bot

import (
	"fmt"

	"maunium.net/go/mautrix/id"
)

type cryptoLog struct{}

func (_ cryptoLog) Error(message string, args ...interface{}) {
	fmt.Printf(message, args)
}

func (_ cryptoLog) Warn(message string, args ...interface{}) {
	fmt.Printf(message, args)
}

func (_ cryptoLog) Debug(message string, args ...interface{}) {
	fmt.Printf(message, args)
}

func (_ cryptoLog) Trace(message string, args ...interface{}) {
	fmt.Printf(message, args)
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