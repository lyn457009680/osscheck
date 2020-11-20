package scheduler

import (
	"context"
	"github.com/cihub/seelog"
	"os"
	"osscheck/request"
	"time"
)

type QueuedScheduler struct {
	requestChan chan request.Request
	workerChan  chan chan request.Request
}

func (s *QueuedScheduler) Submit(r request.Request) {
	s.requestChan <- r
}

func (s *QueuedScheduler) WorkerReady(w chan request.Request) {
	s.workerChan <- w
}

func (s *QueuedScheduler) Start(OverCtx context.Context) {
	s.workerChan = make(chan chan request.Request)
	s.requestChan = make(chan request.Request)
	go func() {
		var requestQ []request.Request
		var workerQ []chan request.Request
		timeContinue := time.Now().Unix()
		for {
			var activeRequest request.Request
			var activeWorker chan request.Request
			if len(requestQ) > 0 && len(workerQ) > 0 {
				activeRequest = requestQ[0]
				activeWorker = workerQ[0]
			}
			select {
			case r := <-s.requestChan:
				requestQ = append(requestQ, r)
				timeContinue = time.Now().Unix()
			case w := <-s.workerChan:
				workerQ = append(workerQ, w)
			case activeWorker <- activeRequest:
				workerQ = workerQ[1:]
				requestQ = requestQ[1:]
			case <-OverCtx.Done():
				seelog.Errorf("检测到未包含资源文件程序退出")
				time.Sleep(1 * time.Second)
				os.Exit(1)
			default:
				nowTime := time.Now().Unix()
				if nowTime-timeContinue > 10 {
					seelog.Errorf("无请求队列程序退出!!")
					time.Sleep(1 * time.Second)
					os.Exit(1)
				}
			}
		}
	}()
}
