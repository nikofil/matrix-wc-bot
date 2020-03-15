package bot

type deviceKeys struct {
	Algorithms []string                     `json:"algorithms"`
	DeviceID   string                       `json:"device_id"`
	Keys       map[string]string            `json:"keys"`
	Signatures map[string]map[string]string `json:"signatures,omitempty"`
	UserID     string                       `json:"user_id"`
}

type uploadKeysReq struct {
	DeviceKeys  deviceKeys        `json:"device_keys,omitempty"`
	OneTimeKeys map[string]string `json:"one_time_keys,omitempty"`
}
