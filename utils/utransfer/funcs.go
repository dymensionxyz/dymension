package utransfer

import "encoding/json"

// TODO: need to check determinism hazards
func CreateMemo(eibcFee string, onComplete []byte) string {

	m := map[string]map[string]any{
		"eibc": map[string]any{
			"fee": eibcFee,
		},
	}
	if len(onComplete) > 0 {
		m["eibc"]["on_completion"] = onComplete
	}

	eibcJson, _ := json.Marshal(m)

	return string(eibcJson)

}
