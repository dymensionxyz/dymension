package types

import (
	"time"
)

const (
	LockMemoName        = "bridge_and_lock"
	DefaultLockDuration = 60 * time.Second
)

type LockMemo struct {
	ToLock bool `json:"to_lock"`
}

func (m *LockMemo) Validate() error {
	return nil
}
