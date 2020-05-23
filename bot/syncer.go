package bot

import (
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// MySyncer is our custon handler for /sync responses
type MySyncer struct {
	mautrix.DefaultSyncer
	callback     func(*event.Event, id.RoomID)
	client       *mautrix.Client
	setRoomState func(map[id.RoomID][]id.UserID, map[id.RoomID]bool)
	first        bool
}

// ProcessResponse processes a /sync response
func (s *MySyncer) ProcessResponse(res *mautrix.RespSync, since string) (err error) {
	if s.first {
		roomMembers := make(map[id.RoomID][]id.UserID)
		roomIsEnc := make(map[id.RoomID]bool)
		for i, v := range res.Rooms.Join {
			isEnc := false
			mids := make([]id.UserID, 0)
			members, err := s.client.Members(i, mautrix.ReqMembers{At: v.Timeline.PrevBatch})
			if err != nil {
				return err
			}
			for _, e := range members.Chunk {
				if e.Type == event.StateMember {
					e.Content.ParseRaw(event.StateMember)
					if e.Content.AsMember().Membership == event.MembershipJoin {
						mids = append(mids, e.Sender)
					}
				}
			}
			if s.client.StateEvent(i, event.StateEncryption, "", nil) == nil {
				isEnc = true
			}
			fmt.Println("Neb is in room", i, "along with", mids)
			roomMembers[i] = mids
			if isEnc {
				fmt.Println("Room", i, "is encrypted!")
			}
			roomIsEnc[i] = isEnc
			s.setRoomState(roomMembers, roomIsEnc)
		}
		// ignore first sync results
		s.first = false
		return nil
	}
	for roomID, roomData := range res.Rooms.Join {
		for _, evt := range roomData.Timeline.Events {
			go s.callback(evt, roomID)
		}
	}
	return nil
}
