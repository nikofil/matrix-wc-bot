package bot

import (
	"encoding/json"
	"fmt"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// WCBot stores a list of messages for each room
type WCBot struct {
	client     *mautrix.Client
	deviceID   string
	roomMsgs   map[string][]string
	cryptoMach *crypto.OlmMachine
}

// NewWCBot creates a new WCBot and logs in
func NewWCBot(serverURL, username, password, deviceID string) (*WCBot, error) {
	client, err := mautrix.NewClient(serverURL, "", "")
	if err != nil {
		return nil, err
	}

	gobStore, err := crypto.NewGobStore("cryptoStore.gob")
	if err != nil {
		return nil, err
	}

	roomStats := &roomCache{}

	bot := WCBot{
		client,
		deviceID,
		make(map[string][]string),
		crypto.NewOlmMachine(client, cryptoLog{}, gobStore, roomStats),
	}

	login, err := client.Login(&mautrix.ReqLogin{
		Identifier: mautrix.UserIdentifier{
			Type: "m.id.user",
			User: username,
		},
		Password:                 password,
		Type:                     "m.login.password",
		DeviceID:                 id.DeviceID(deviceID),
		InitialDeviceDisplayName: "bot dev " + deviceID})
	if err != nil {
		return nil, err
	}
	fmt.Println("Logged in as", login.UserID, ", device", login.DeviceID)
	client.UserID = login.UserID
	client.AccessToken = login.AccessToken
	client.DeviceID = id.DeviceID(deviceID)

	// client.UploadKeys(&mautrix.ReqUploadKeys{})

	bot.cryptoMach.Load()

	client.Syncer = &MySyncer{*client.Syncer.(*mautrix.DefaultSyncer), bot.processMsg, bot.client, roomStats.SetRoomState, bot.cryptoMach, true}

	return &bot, nil
}

func (bot *WCBot) msgToRoom(roomID id.RoomID, msg string) (err error) {
	fmt.Println(" + Sending", msg, "to", roomID)
	if bot.cryptoMach.StateStore.IsEncrypted(roomID) {
		members := bot.cryptoMach.StateStore.(*roomCache).RoomMembers(roomID)
		bot.cryptoMach.ShareGroupSession(roomID, members)
		json, err := json.Marshal(map[string]string{"msgtype": "m.text", "body": msg})
		if err != nil {
			return err
		}
		content := event.Content{}
		content.UnmarshalJSON(json)
		content.ParseRaw(event.EventMessage)
		enc, err := bot.cryptoMach.EncryptMegolmEvent(roomID, event.EventMessage, content)
		if err != nil {
			return err
		}
		_, err = bot.client.SendMessageEvent(roomID, event.EventEncrypted, enc)
	} else {
		_, err = bot.client.SendMessageEvent(
			roomID,
			event.EventMessage,
			map[string]string{"msgtype": "m.text", "body": msg})
	}
	return
}

func (bot *WCBot) processMsg(evt *event.Event, room id.RoomID) {
	if err := evt.Content.ParseRaw(event.EventMessage); err == nil {
		if evt.Type == event.EventEncrypted {
			evt.RoomID = room
			evt.Content.ParseRaw(event.EventEncrypted)
			evt, err = bot.cryptoMach.DecryptMegolmEvent(evt)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		body := evt.Content.AsMessage().Body
		if strings.HasPrefix(body, "!wc ") {
			search := strings.TrimPrefix(body, "!wc ")
			if roomMsgs, ok := bot.roomMsgs[room.String()]; ok {
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
		fmt.Printf(" + New msg [%s] in room %s\n", body, room)
		bot.roomMsgs[room.String()] = append(bot.roomMsgs[room.String()], body)
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
				if chunk.Type == event.EventEncrypted {
					var err error
					chunk.Content.ParseRaw(event.EventEncrypted)
					chunk, err = bot.cryptoMach.DecryptMegolmEvent(chunk)
					if err != nil {
						fmt.Println("Error decrypting msg:", err)
						continue
					}
				}
				if body := chunk.Content.AsMessage().Body; body != "" {
					fmt.Printf(" + Read existing msg [%s]\n", body)
					bot.roomMsgs[room.String()] = append(bot.roomMsgs[room.String()], body)
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

	return bot.client.Sync()
}
