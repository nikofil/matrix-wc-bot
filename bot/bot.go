package bot

import (
	"fmt"
	"strings"

	"github.com/matrix-org/gomatrix"
)

// WCBot stores a list of messages for each room
type WCBot struct {
	client    *gomatrix.Client
	deviceID  string
	roomMsgs  map[string][]string
}

// NewWCBot creates a new WCBot and logs in
func NewWCBot(serverURL, username, password, deviceID string) (*WCBot, error) {
	client, err := gomatrix.NewClient(serverURL, "", "")
	if err != nil {
		return nil, err
	}

	bot := WCBot{
		client,
		deviceID,
		make(map[string][]string),
	}

	login, err := client.Login(&gomatrix.ReqLogin{
		User:     username,
		Password: password,
		Type:     "m.login.password",
		DeviceID: deviceID})
	if err != nil {
		return nil, err
	}
	fmt.Println("Logged in as", login.UserID, ", device", login.DeviceID)
	client.UserID = login.UserID
	client.AccessToken = login.AccessToken

	return &bot, nil
}

func (bot *WCBot) msgToRoom(roomID, msg string) error {
	_, err := bot.client.SendMessageEvent(roomID, "m.room.message", map[string]string{"msgtype": "m.text", "body": msg})
	return err
}

func (bot *WCBot) processMsg(evt *gomatrix.Event) {
	room := evt.RoomID
	if body, ok := evt.Body(); ok {
		if strings.HasPrefix(body, "!wc ") {
			search := strings.TrimPrefix(body, "!wc ")
			if roomMsgs, ok := bot.roomMsgs[room]; ok {
				fmt.Printf(" ! Looking for [%s] in room %s\n", search, room)
				cnt := 0
				for _, msg := range roomMsgs {
					if strings.Contains(msg, search) {
						cnt++
					}
				}
				fmt.Println(bot.msgToRoom(room, fmt.Sprintf("Found [%s] in this room %d times!", search, cnt)))
			} else {
				fmt.Printf(" ! No messages found for room [%s]\n", room)
			}
		}
		fmt.Printf(" + Msg [%s] in room %s\n", body, room)
		bot.roomMsgs[room] = append(bot.roomMsgs[room], body)
	}
}

// Run starts the bot and listens for messages
func (bot *WCBot) Run() error {
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

	bot.client.Syncer.(*gomatrix.DefaultSyncer).OnEventType("m.room.message", func(evt *gomatrix.Event) {
		go bot.processMsg(evt)
	})

	return bot.client.Sync()
}
