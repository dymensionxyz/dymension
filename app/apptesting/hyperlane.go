package apptesting

import (
	"github.com/dymensionxyz/dymension/v3/utils/uhyp"
)

/*
TODO: this is all a big wip
*/

func (s *KeeperTestHelper) SetupHyperlane() {

	server := uhyp.NewServer(&s.App.HyperCoreKeeper, &s.App.HyperCoreKeeper.PostDispatchKeeper, &s.App.HyperCoreKeeper.IsmKeeper)
	owner := Alice

	mailboxId, err := server.CreateDefaultMailbox(s.Ctx, owner)
	s.NoError(err)
	_ = mailboxId

}
