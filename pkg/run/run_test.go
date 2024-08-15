package run

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type impl struct{}

func (i impl) Stop(_ context.Context) error {
	return nil
}

func (i impl) Start(_ context.Context) error {
	return nil
}

func TestStartable(t *testing.T) {
	imp := impl{}
	require.Implements(t, (*Startable)(nil), imp)
}

func TestStopable(t *testing.T) {
	imp := impl{}
	require.Implements(t, (*Stoppable)(nil), imp)
}
