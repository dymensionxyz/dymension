package utransfer

import "encoding/json"

func CreateMemo(eibcFee string, fulfillHook []byte) string {

	m := map[string]map[string]any{
		"eibc": map[string]any{
			"fee": eibcFee,
		},
	}
	if len(fulfillHook) > 0 {
		m["eibc"]["fulfill_hook"] = fulfillHook
	}

	eibcJson, _ := json.Marshal(m)

	return string(eibcJson)

}
