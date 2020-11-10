package parser

import (
	"github.com/cihub/seelog"
	"osscheck/config"
	"osscheck/engine"
	"regexp"
	"sync"
)
const personRe = `<a href="([^"]*)" title="([^"]*)"><img class="lazyload" data-original="([^"]*)" />`

const linkRe = `<a href="([^\"]*)">\d*</a>`

var LinkMap = &sync.Map{}

func ParseCheck(contents []byte) engine.ParseResult {
	personReMust := regexp.MustCompile(personRe)
	actorMatches := personReMust.FindAllSubmatch(contents, -1)
	linkReMust := regexp.MustCompile(linkRe)
	linkMatches := linkReMust.FindAllSubmatch(contents,-1)
	result := engine.ParseResult{}
	//防止重复爬取
	for _,links := range  linkMatches {
		link := string(links[1])
		_,ok := LinkMap.Load(link)
		if !ok {
			LinkMap.Store(link,true)
			result.Requests = append(result.Requests, engine.Request{
				Url:       config.ROOTURL+link,
				ParserFunc: ParseCheck,
			})
		}
	}
	for _, m := range actorMatches {
		link := string(m[1])
			result.Requests = append(result.Requests, engine.Request{
				Url:       config.ROOTURL+link,
				ParserFunc:func(c []byte) engine.ParseResult {
					return ParseCheck(c)
				},
			})
			if string(m[2]) == "" {
				seelog.Errorf("相关信息为:%v,%v,%v,%v ",m[0],m[1],m[2],m[3])
			}
			result.Items = append(result.Items, "actor: "+string(m[2]))

	}
	return result
}
