package bot

import (
	"fmt"

	"github.com/matrix-org/gomatrix"
)

// WCBot stores a list of messages for each room
type WCBot struct {
	client   *gomatrix.Client
	roomMsgs map[string][]string
}

// NewWCBot creates a new WCBot and logs in
func NewWCBot(serverURL, username, password string) (*WCBot, error) {
	client, err := gomatrix.NewClient(serverURL, "", "")
	if err != nil {
		return nil, err
	}

	bot := WCBot{client, make(map[string][]string)}

	login, err := client.Login(&gomatrix.ReqLogin{User: username, Password: password, Type: "m.login.password"})
	if err != nil {
		return nil, err
	}
	fmt.Println("Logged in as", login.UserID, "device", login.DeviceID)
	client.UserID = login.UserID
	client.AccessToken = login.AccessToken

	return &bot, nil
}

// Run starts the bot and listens for messages
func (bot *WCBot) Run() error {
	bot.client.Syncer.(*gomatrix.DefaultSyncer).OnEventType("m.room.message", func(evt *gomatrix.Event) {
		fmt.Println("got msg", evt)
	})

	joined, err := bot.client.JoinedRooms()
	if err != nil {
		return err
	}
	for _, room := range joined.JoinedRooms {
		if msgs, err := bot.client.Messages(room, "", "", 'b', 1000); err == nil {
			for _, chunk := range msgs.Chunk {
				if body, ok := chunk.Body(); ok {
					bot.roomMsgs[room] = append(bot.roomMsgs[room], body)
				}
			}
		} else {
			return err
		}
	}

	for room, msgs := range bot.roomMsgs {
		fmt.Printf("%s - %d msgs\n", room, len(msgs))
	}

	fmt.Println("Waiting for syncs")
	// x.SendMessageEvent("!hmNFHDNUWDkHPeIyXb:localhost", "m.room.message", map[string]string{"msgtype": "m.text", "body": "HELLO!!"})
	return bot.client.Sync()
}
