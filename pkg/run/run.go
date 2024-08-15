package run

import "context"

type Startable interface {
	Start(ctx context.Context) error
}

type Stoppable interface {
	Stop(ctx context.Context) error
}
