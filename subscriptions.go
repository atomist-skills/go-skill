package skill

import (
	"olympos.io/encoding/edn"
)

func (e *EventContextSubscription) GetResultInMapForm() []map[edn.Keyword]edn.RawMessage {
	return decode[[]map[edn.Keyword]edn.RawMessage](e.Result)
}

func (e *EventContextSubscription) GetResultInListForm() [][]edn.RawMessage {
	return decode[[][]edn.RawMessage](e.Result)
}

func decode[P interface{}](event edn.RawMessage) P {
	ednboby, _ := edn.Marshal(event)
	var decoded P
	edn.Unmarshal(ednboby, &decoded)
	return decoded
}
