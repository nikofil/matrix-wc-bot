package bot

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// MySyncer is our custon handler for /sync responses
type MySyncer struct {
	mautrix.DefaultSyncer
	callback func(*event.Event, id.RoomID)
	first    bool
}

// ProcessResponse processes a /sync response
func (s *MySyncer) ProcessResponse(res *mautrix.RespSync, since string) (err error) {
	if s.first {
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
