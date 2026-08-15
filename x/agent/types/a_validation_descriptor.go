package types

import gogoproto "github.com/cosmos/gogoproto/proto"

// validation.proto sorts after generated importers. Register its descriptor
// first so gogo does not permanently bake placeholder imports into them.
func init() {
	descriptor, _ := (&ValidationRequest{}).Descriptor()
	gogoproto.RegisterFile("dymensionxyz/dymension/agent/validation.proto", descriptor)
}
