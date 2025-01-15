package utest

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IsErrSuite struct {
	suite.Suite
}

func TestIsErr(t *testing.T) {
	suite.Run(t, new(IsErrSuite))
}

func (s *IsErrSuite) TestSimple() {
	errA := errors.New("a")
	errB := fmt.Errorf("b: %w", errA)
	IsErr(s.Require(), errA, errB)
}

func (s *IsErrSuite) TestSimpleFail() {
	s.T().Skip() // it will fail (as it should), so skip
	errA := errors.New("a")
	errB := errors.New("b")
	IsErr(s.Require(), errA, errB)
}
