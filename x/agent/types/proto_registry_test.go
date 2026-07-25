package types_test

import (
	"strings"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	_ "github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// Gogo builds each file descriptor eagerly when its generated file's init runs,
// and inits run in lexical file-name order. A sibling import whose .pb.go sorts
// after the importer is baked in as an empty placeholder forever, which makes
// AutoCLI's dynamicpb rendering silently drop the nested fields (this is why
// the feedback messages live in feedback.proto, which sorts before genesis and
// query). This guards the module's proto file naming against that regression.
func TestQueryDescriptorImportsResolve(t *testing.T) {
	for _, path := range []string{
		"dymensionxyz/dymension/agent/genesis.proto",
		"dymensionxyz/dymension/agent/query.proto",
	} {
		fd, err := gogoproto.HybridResolver.FindFileByPath(path)
		require.NoError(t, err)
		for i := 0; i < fd.Imports().Len(); i++ {
			imp := fd.Imports().Get(i)
			if strings.HasPrefix(imp.Path(), "dymensionxyz/") {
				require.False(t, imp.IsPlaceholder(), "%s: import %s is a placeholder", path, imp.Path())
			}
		}
	}
}
