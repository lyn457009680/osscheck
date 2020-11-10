package scheduler

import "osscheck/engine"

type Scheduler interface {
	Submit(engine.Request)
	WorkerReady(chan engine.Request)
	Start()
}
