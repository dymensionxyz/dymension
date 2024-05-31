package types

type Memo struct {
	Data struct {
		SkipDelay bool `json:"skip_delay"`
	} `json:"delayedack"` // namespace
}
