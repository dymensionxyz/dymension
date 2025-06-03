package types

import (
	"encoding/binary"
)

var _ binary.ByteOrder

const (
	ModuleName = "kas"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

const KeyProcessedWithdrawals = "pws"
