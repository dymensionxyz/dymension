package ibctesting_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

/*
Friday:
Got the basic hook structure up:
A memo can contain a hook name and data
It will end up in x/forward
which is connected to the warp route keeper, so it can initiate sends

Next sensible things:
- Need to ideally have a test that takes a transfer with the memo, and ensures the forwarder gets called
 (can use a dummy hook)
-
*/

type forwardSuite struct {
	eibcSuite
}

func TestForwardSuite(t *testing.T) {
	suite.Run(t, new(forwardSuite))
}

func (s *forwardSuite) SetupTest() {
	s.eibcSuite.SetupTest()
}

func (s *forwardSuite) TestForward() {
	s.T().Log("running test forward!")
}
