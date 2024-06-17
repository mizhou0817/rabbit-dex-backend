package model

// One struct for decision making below the flow
// Fields can be duplicated from context
type MatchingMeta struct {
	Device     string `msgpack:"device" json:"device"`
	IsApi      bool   `msgpack:"is_api" json:"is_api"`
	ExchangeId string `msgpack:"exchange_id" json:"exchange_id"`
	IsPm       bool   `msgpack:"is_pm" json:"is_pm"`
}

func (m *MatchingMeta) SetPm(isPm bool) {
	m.IsPm = isPm
}

func (m *MatchingMeta) SetDevice(device string) {
	m.Device = device
}

func (m *MatchingMeta) SetApi(isApi bool) {
	m.IsApi = isApi
}

func (m *MatchingMeta) SetEid(eid string) {
	m.ExchangeId = eid
}
