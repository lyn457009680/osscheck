package scheduler

import (
	"context"
	"osscheck/engine"
)

type Scheduler interface {
	Submit(engine.Request)
	WorkerReady(chan engine.Request)
	Start(context.Context)
}
