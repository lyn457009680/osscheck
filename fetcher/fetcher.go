package fetcher

import (
	"fmt"
	"github.com/cihub/seelog"
	"io/ioutil"
	"net/http"
	"osscheck/config"
	"sync/atomic"
	"time"
)

var timeLimiter = time.Tick(time.Duration(config.FETCHER_INTERVAL) * time.Millisecond)

var requestCount int64

func Fetcher(url string, device_type string) ([]byte, error) {
	atomic.AddInt64(&requestCount, 1)
	seelog.Tracef("fetcher调用统计: %v", requestCount)
	<-timeLimiter
	var err error
	timeout := time.Duration(60 * time.Second)
	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	if err != nil {
		seelog.Errorf("请求出错:%v", err)
	}
	if device_type == "PC" {
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	} else {
		req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.75 Mobile Safari/537.36")
	}
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	content, err := ioutil.ReadAll(resp.Body)
	return content, err
}
