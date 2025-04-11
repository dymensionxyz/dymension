package utransfer

import "encoding/json"

// TODO: need to check determinism hazards
func CreateMemo(eibcFee string, fulfillHook []byte) string {

	m := map[string]map[string]any{
		"eibc": map[string]any{
			"fee": eibcFee,
		},
	}
	if len(fulfillHook) > 0 {
		m["eibc"]["on_fulfill"] = fulfillHook
	}

	eibcJson, _ := json.Marshal(m)

	return string(eibcJson)

}
