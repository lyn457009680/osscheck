package main

import (
	"flag"
	"github.com/cihub/seelog"
	"osscheck/config"
	"osscheck/engine"
	"osscheck/parser"
	"osscheck/scheduler"
)

func init() {
	flag.StringVar(&config.ROOTURL, "h", "http://carservice.com/", "设置网站的域名")
	flag.StringVar(&config.CHECKURL, "t", "carservice.com", "设置检测OSS or CDN的域名")
	flag.IntVar(&config.FETCHER_INTERVAL, "d", 2000, "设置检测间隔毫秒")
}
func main() {
	flag.Parse()
	println(config.ROOTURL)
	defer func() {
		seelog.Flush()
	}()
	e := engine.ConcurrentEngine{
		Scheduler:   scheduler.QueuedScheduler{},
		WorkerCount: 5,
	}
	e.Run(engine.Request{
		Url: config.ROOTURL,
		ParserFunc: func(c []byte) engine.ParseResult {
			return parser.ParseCheck(c)
		},
	})
}
