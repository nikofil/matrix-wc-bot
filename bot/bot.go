package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/matrix-org/gomatrix"
	"github.com/nikofil/matrix-wc-bot/bot/olm"
)

// WCBot stores a list of messages for each room
type WCBot struct {
	client   *gomatrix.Client
	deviceID string
	roomMsgs map[string][]string
	acc      olm.Account
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
		olm.CreateNewAccount(),
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

func (bot *WCBot) processEncrypted(evt *gomatrix.Event) {
	fmt.Println(evt)
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
				bot.msgToRoom(room, fmt.Sprintf("Found [%s] in this room %d times!", search, cnt))
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

	bot.client.Syncer.(*gomatrix.DefaultSyncer).OnEventType("m.room.encrypted", func(evt *gomatrix.Event) {
		go bot.processEncrypted(evt)
	})

	if err = bot.prepareKeys(); err != nil {
		return err
	}

	return bot.client.Sync()
}

func (bot *WCBot) prepareKeys() error {
	pickleFileName := "keys-" + bot.deviceID + ".pickle"
	keyUploadURL := bot.client.BuildURL("keys", "upload")
	resp := make(map[string]map[string]int)

	if pickled, err := ioutil.ReadFile(pickleFileName); err == nil {
		bot.acc = olm.AccountFromPickle("1", string(pickled))
	} else {
		keys := bot.genDeviceKeys()
		fmt.Println("keys", keys)

		algorithms := []string{"m.olm.v1.curve25519-aes-sha2", "m.megolm.v1.aes-sha2"}

		dkeys := deviceKeys{
			UserID:     bot.client.UserID,
			DeviceID:   bot.deviceID,
			Algorithms: algorithms,
			Keys:       keys,
		}

		signatures, err := bot.signKeys(interface{}(dkeys))
		if err != nil {
			return err
		}

		dkeys.Signatures = signatures
		request := uploadDeviceKeysReq{
			DeviceKeys: dkeys,
		}

		fmt.Println("req keys", keys)
		fmt.Println("req sigs", signatures)
		fmt.Println("req algs", algorithms)
		fmt.Println("req ids", bot.client.UserID, bot.deviceID)
		err = bot.client.MakeRequest("POST", keyUploadURL, request, &resp)
		if err != nil {
			return err
		}
		fmt.Println("resp", resp)
	}

	fmt.Println("max otks", bot.acc.GetMaxNumberOfOneTimeKeys())

	otkRequest := bot.genOneTimeKeys(10)
	fmt.Println("otk", otkRequest)
	err := bot.client.MakeRequest("POST", keyUploadURL, otkRequest, &resp)
	if err != nil {
		return err
	}
	bot.acc.MarkKeysAsPublished()
	fmt.Println("resp", resp)

	pickled := bot.acc.Pickle("1")
	if err = ioutil.WriteFile("keys-"+bot.deviceID+".pickle", []byte(pickled), 0644); err != nil {
		return err
	}

	return nil
}

func (bot *WCBot) genDeviceKeys() map[string]string {
	keysMap := make(map[string]string)
	keysRes := make(map[string]string)

	keys := bot.acc.GetIdentityKeys()
	json.Unmarshal([]byte(keys), &keysMap)

	for algo, keyVal := range keysMap {
		keysRes[bot.deviceID+":"+algo] = keyVal
	}

	return keysRes
}

func (bot *WCBot) genOneTimeKeys(num int) uploadOneTimeKeysReq {
	oneTimeKeysMap := make(map[string]map[string]string)
	resMap := make(map[string]oneTimeKeysReqMap)

	bot.acc.GenerateOneTimeKeys(num)
	oneTimeKeys := bot.acc.GetOneTimeKeys()
	json.Unmarshal([]byte(oneTimeKeys), &oneTimeKeysMap)

	for algo, keys := range oneTimeKeysMap {
		for keyID, keyVal := range keys {
			otkMap := oneTimeKeysReqMap{Key: keyVal}
			signatures, err := bot.signKeys(interface{}(otkMap))
			if err == nil {
				otkMap.Signatures = signatures
				resMap["signed_"+algo+":"+keyID] = otkMap
			}
		}
	}
	return uploadOneTimeKeysReq{resMap}
}

func (bot *WCBot) signKeys(keys interface{}) (*signaturesMap, error) {
	dkeysBytes, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}

	signatures := signaturesMap(map[string]map[string]string{
		bot.client.UserID: map[string]string{"ed25519:" + bot.deviceID: bot.acc.Sign(string(dkeysBytes))},
	})

	return &signatures, nil
}
