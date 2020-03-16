package bot

type signaturesMap map[string]map[string]string

type deviceKeys struct {
	Algorithms []string           `json:"algorithms"`
	DeviceID   string             `json:"device_id"`
	Keys       map[string]string  `json:"keys"`
	UserID     string             `json:"user_id"`
	Signatures *signaturesMap     `json:"signatures,omitempty"`
	Unsigned   *map[string]string `json:"unsigned,omitempty"`
}

type uploadDeviceKeysReq struct {
	DeviceKeys deviceKeys `json:"device_keys"`
}

type oneTimeKeysReqMap struct {
	Key        string         `json:"key"`
	Signatures *signaturesMap `json:"signatures,omitempty"`
}

type uploadOneTimeKeysReq struct {
	OneTimeKeys map[string]oneTimeKeysReqMap `json:"one_time_keys"`
}
