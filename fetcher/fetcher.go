package fetcher

import (
	"bufio"
	"fmt"
	"github.com/cihub/seelog"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"osscheck/config"
	"strings"
	"sync/atomic"
	"time"
)

var timeLimiter = time.Tick(time.Duration(config.FETCHER_INTERVAL) * time.Millisecond)

var requestCount int64
var successCount int64
var  judgeRe = `is not available because it is categorized as`
func Fetcher(url string) ([]byte, error) {
	atomic.AddInt64(&requestCount, 1)
	seelog.Tracef("fetcher调用统计: %v",requestCount)
	<-timeLimiter
	var err error
	timeout := time.Duration(60 * time.Second)
	client := &http.Client{
		Transport: &http.Transport{
		},
		Timeout:timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	if err  != nil {
		seelog.Errorf("代理出错:%v",err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(url,".jpg") && resp.StatusCode == http.StatusNotFound  {
			seelog.Infof("获取的图片返回404,url为:%v",url)
			return nil,nil
		}
		return nil, fmt.Errorf("wrong status code: %d", resp.StatusCode)
	}
	bufioReader := bufio.NewReader(resp.Body)
	e := determineEncoding(bufioReader)
	utf8Reader := transform.NewReader(bufioReader, e.NewDecoder())
	content ,err :=  ioutil.ReadAll(utf8Reader)
	judgeResponse(content)
	return content,err
}

func judgeResponse (contents []byte)  bool {
	s := string(contents)
	if strings.Contains(s,judgeRe) {
		return false;
	}
	return true
}
func determineEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	if err != nil {
		log.Printf("fetch error : %v", err)
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}
