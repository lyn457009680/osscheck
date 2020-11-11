package parser

import (
	"github.com/cihub/seelog"
	"osscheck/config"
	"osscheck/request"
	"regexp"
	"strings"
	"sync"
)

const IMGJSRe = `src=["|'](.*?)["|']` //图片 或者 js

const CSSRe = `href=["|'](.*?)["|']` //a 活着css

var LinkMap = &sync.Map{}

func ParseCheck(contents []byte, v string) request.ParseResult {
	IMGJSReMust := regexp.MustCompile(IMGJSRe)
	IMGJSMatches := IMGJSReMust.FindAllSubmatch(contents, -1)
	linkReMust := regexp.MustCompile(CSSRe)
	linkMatches := linkReMust.FindAllSubmatch(contents, -1)
	result := request.ParseResult{}
	//防止重复爬
	for _, links := range linkMatches {
		link := string(links[1])
		if link == "/" {
			continue //根目录不检索
		}
		if strings.Contains(link, ".css") {
			seelog.Info("检测到新资源地址" + link)
			result.Items = append(result.Items, link)
			continue
		}
		_, ok := LinkMap.Load(link)
		if !ok {
			LinkMap.Store(link, true)
			seelog.Info("检测到新链接地址" + link)
			result.Requests = append(result.Requests, request.Request{
				Url:        config.ROOTURL + link,
				DeviceType: v,
				ParserFunc: ParseCheck,
			})
		}
	}
	for _, m := range IMGJSMatches {
		link := string(m[1])
		if strings.Contains(link, ".html") {
			seelog.Info("检测到新链接地址" + link)
			_, ok := LinkMap.Load(link)
			if !ok {
				LinkMap.Store(link, true)
				result.Requests = append(result.Requests, request.Request{
					Url:        config.ROOTURL + link,
					ParserFunc: ParseCheck,
				})
			}
			continue
		}
		seelog.Info("检测到新资源地址" + link)
		result.Items = append(result.Items, link)
	}
	return result
}
