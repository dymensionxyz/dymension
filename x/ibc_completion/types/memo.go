package types

import "encoding/json"

type Memo struct {
	// can be nil
	OnCompletionHook []byte `json:"on_completion,omitempty"`
}

func MakeMemo(onCompletionHook []byte) (string, error) {
	m := Memo{
		OnCompletionHook: onCompletionHook,
	}
	bz, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
