package main

import (
	"flag"
	"github.com/cihub/seelog"
	"osscheck/config"
	"osscheck/engine"
	"osscheck/parser"
	"osscheck/request"
	"osscheck/scheduler"
)

func init() {
	flag.StringVar(&config.ROOTURL, "r", "http://carservice.com/", "设置网站的域名")
	flag.StringVar(&config.CHECKURL, "c", "carservice.com", "设置检测OSS or CDN的域名")
	flag.IntVar(&config.FETCHER_INTERVAL, "t", 2000, "设置检测间隔毫秒")
	flag.StringVar(&config.DEVICETYPE, "d", "MOBILE", "设置检测设备PC||MOBILE")
}
func main() {
	flag.Parse()
	logger, err := seelog.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		panic(err)
	}
	err = seelog.ReplaceLogger(logger)
	if err != nil {
		panic(err)
	}
	defer func() {
		seelog.Flush()
	}()
	logger.Info("检测URL地址" + config.ROOTURL)
	e := engine.ConcurrentEngine{
		Scheduler:   scheduler.QueuedScheduler{},
		WorkerCount: 5,
	}
	seeds := make([]request.Request, 0)
	DeviceType := []string{}
	if config.DEVICETYPE == "ALL" {
		DeviceType = []string{"PC", "MOBILE"}
	} else {
		DeviceType = []string{config.DEVICETYPE}
	}
	for _, v := range DeviceType {
		seeds = append(seeds, request.Request{
			Url:        config.ROOTURL,
			DeviceType: v,
			ParserFunc: func(c []byte, v string) request.ParseResult {
				return parser.ParseCheck(c, v)
			},
		})
	}
	e.Run(seeds)
}
