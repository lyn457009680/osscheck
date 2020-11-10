package engine

import (
	"fmt"
	"github.com/cihub/seelog"
	"log"
	"osscheck/config"
	"osscheck/fetcher"
	"sync"
)

type Scheduler interface {
	Submit(Request)
	WorkerChan() chan Request
	ReadyNotifier
	Run()
}
type ReadyNotifier interface {
	WorkerReady(chan Request)
}
type ConcurrentEngine struct {
	Scheduler   Scheduler
	WorkerCount int

}
func (e *ConcurrentEngine) Run(seeds ...Request) {
	out := make(chan ParseResult,e.WorkerCount)
	e.Scheduler.Run()
	for i := 0; i < e.WorkerCount; i++ {
		createWorker(e.Scheduler.WorkerChan(), out,e.Scheduler)
	}
	for _, r := range seeds {
		e.Scheduler.Submit(r)
	}
	itemCount := 0
	for parseResult := range out {
		for _, item := range parseResult.Items {
			itemCount++
			log.Printf("get item %v %v", itemCount, item)
		}
		for _, request := range parseResult.Requests {
			e.Scheduler.Submit(request)
		}
	}
}
func createWorker(in chan Request, out chan ParseResult, r ReadyNotifier) {
	go func() {
		for {
			r.WorkerReady(in)
			request := <-in
			parseResult, err := work(request)
			if err != nil {
				fmt.Printf("错误%v",err)
			}
			out <- parseResult
		}
	}()
}
var requestCount = &sync.Map{}

func work(r Request) (ParseResult, error) {
	seelog.Tracef("fetch url :%s ",r.Url)
	body, err := fetcher.Fetcher(r.Url)
	if err != nil {
		cInt := 0
		count,ok := requestCount.Load(r.Url)
		if ok {
			cInt = count.(int)
		}
		if  cInt < config.REQUEST_ERROR_NUMBER {
			seelog.Warnf("请求%v失败,错误信息为%v,错误次数为%v,回到队列中",r.Url,err,cInt)
			requestCount.Store(r.Url,cInt+1)
			return ParseResult{
				Requests:[]Request{r},
			},err
		}
		seelog.Warnf("请求%v失败,错误信息为%v,错误次数为%v,放弃请求",r.Url,err,cInt)
		return ParseResult{}, err
	}
	seelog.Tracef("fetch url :%s finish ",r.Url)
	var parseResult ParseResult
	if	body != nil {
		parseResult = r.ParserFunc(body)
	}
	return parseResult, err
}

