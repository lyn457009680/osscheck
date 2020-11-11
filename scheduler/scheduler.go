package scheduler

import (
	"context"
	"osscheck/request"
)

type Scheduler interface {
	Submit(request.Request)
	WorkerReady(chan request.Request)
	Start(context.Context)
}
