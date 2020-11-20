package engine

import (
	"context"
	"github.com/cihub/seelog"
	"osscheck/config"
	"osscheck/fetcher"
	"osscheck/request"
	"osscheck/scheduler"
	"strings"
	"sync"
)

type ConcurrentEngine struct {
	Scheduler   scheduler.QueuedScheduler
	WorkerCount int
}

func (e *ConcurrentEngine) Run(seeds []request.Request) {
	out := make(chan request.ParseResult, e.WorkerCount)
	Overcontext, cancelFunc := context.WithCancel(context.Background())
	e.Scheduler.Start(Overcontext)
	for i := 0; i < e.WorkerCount; i++ {
		in := make(chan request.Request)
		go func() {
			for {
				e.Scheduler.WorkerReady(in)
				request := <-in
				parseResult, err := work(request)
				if err != nil {
					seelog.Errorf("错误%v", err)
				}
				out <- parseResult
			}
		}()
	}
	for _, r := range seeds {
		e.Scheduler.Submit(r)
	}
	itemCount := 0
	for parseResult := range out {
		for _, item := range parseResult.Items {
			itemCount++
			seelog.Infof("get item %v %v", itemCount, item)
			if !strings.Contains(item.(string), config.CHECKURL) {
				seelog.Errorf("资源文件 %s 不包含检测地址 %s", item.(string), config.CHECKURL)
				if false {
					cancelFunc()
				}
			}
		}
		for _, request := range parseResult.Requests {
			e.Scheduler.Submit(request)
		}
	}
}

var requestCount = &sync.Map{}

func work(r request.Request) (request.ParseResult, error) {
	seelog.Infof("%s:fetch url:%s", r.Url, r.DeviceType)
	body, err := fetcher.Fetcher(r.Url, r.DeviceType)
	if err != nil {
		cInt := 0
		count, ok := requestCount.Load(r.Url)
		if ok {
			cInt = count.(int)
		}
		if cInt < config.REQUEST_ERROR_NUMBER {
			seelog.Warnf("请求%v失败,错误信息为%v,错误次数为%v,回到队列中", r.Url, err, cInt)
			requestCount.Store(r.Url, cInt+1)
			return request.ParseResult{
				Requests: []request.Request{r},
			}, err
		}
		seelog.Warnf("请求%v失败,错误信息为%v,错误次数为%v,放弃请求", r.Url, err, cInt)
		return request.ParseResult{}, err
	}
	seelog.Infof("fetch url :%s finish ", r.Url)
	var parseResult request.ParseResult
	if body != nil {
		parseResult = r.ParserFunc(body, r.DeviceType)
	}
	return parseResult, err
}
