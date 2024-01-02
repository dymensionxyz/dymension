package types

import proto "github.com/gogo/protobuf/proto"

//FIXME: should be probably genrated by proto

type Operation int32

const (
	Operation_write  Operation = 0
	Operation_read   Operation = 1
	Operation_delete Operation = 2
)

var Operation_name = map[int32]string{
	0: "WRITE",
	1: "READ",
	2: "DELETE",
}

var Operation_value = map[string]int32{
	"WRITE":  0,
	"READ":   1,
	"DELETE": 2,
}

func (x Operation) String() string {
	return proto.EnumName(Operation_name, int32(x))
}
