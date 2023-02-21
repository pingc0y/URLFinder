package crawler

import (
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"regexp"
	"strings"
)

// 分析内容中的js
func jsFind(cont, host, scheme, path, source string, num int) {
	var cata string
	care := regexp.MustCompile("/.*/{1}|/")
	catae := care.FindAllString(path, -1)
	if len(catae) == 0 {
		cata = "/"
	} else {
		cata = catae[0]
	}
	//js匹配正则
	host = scheme + "://" + host
	for _, re := range config.JsFind {
		reg := regexp.MustCompile(re)
		jss := reg.FindAllStringSubmatch(cont, -1)
		//return
		jss = jsFilter(jss)
		//循环提取js放到结果中
		for _, js := range jss {
			if js[0] == "" {
				continue
			}
			if strings.HasPrefix(js[0], "https:") || strings.HasPrefix(js[0], "http:") {
				AppendJs(js[0], source)
				if num < 5 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(js[0], num+1)
				}
			} else if strings.HasPrefix(js[0], "//") {
				AppendJs(scheme+":"+js[0], source)
				if num < 5 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(scheme+":"+js[0], num+1)
				}

			} else if strings.HasPrefix(js[0], "/") {
				AppendJs(host+js[0], source)
				if num < 5 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(host+js[0], num+1)
				}
			} else if strings.HasPrefix(js[0], "./") {
				AppendJs(host+"/"+js[0], source)
				if num < 5 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(host+"/"+js[0], num+1)
				}
			} else {
				AppendJs(host+cata+js[0], source)
				if num < 5 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(host+cata+js[0], num+1)
				}
			}
		}

	}

}

// 分析内容中的url
func urlFind(cont, host, scheme, path, source string, num int) {
	var cata string
	care := regexp.MustCompile("/.*/{1}|/")
	catae := care.FindAllString(path, -1)
	if len(catae) == 0 {
		cata = "/"
	} else {
		cata = catae[0]
	}
	host = scheme + "://" + host

	//url匹配正则

	for _, re := range config.UrlFind {
		reg := regexp.MustCompile(re)
		urls := reg.FindAllStringSubmatch(cont, -1)
		//fmt.Println(urls)
		urls = urlFilter(urls)

		//循环提取url放到结果中
		for _, url := range urls {
			if url[0] == "" {
				continue
			}
			if strings.HasPrefix(url[0], "https:") || strings.HasPrefix(url[0], "http:") {
				AppendUrl(url[0], source)
				if num < 2 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(url[0], num+1)
				}
			} else if strings.HasPrefix(url[0], "//") {
				AppendUrl(scheme+":"+url[0], source)
				if num < 2 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(scheme+":"+url[0], num+1)
				}
			} else if strings.HasPrefix(url[0], "/") {
				urlz := ""
				if cmd.B != "" {
					urlz = cmd.B + url[0]
				} else {
					urlz = host + url[0]
				}
				AppendUrl(urlz, source)
				if num < 2 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(urlz, num+1)
				}
			} else if !strings.HasSuffix(source, ".js") {
				urlz := ""
				if cmd.B != "" {
					if strings.HasSuffix(cmd.B, "/") {
						urlz = cmd.B + url[0]
					} else {
						urlz = cmd.B + "/" + url[0]
					}
				} else {
					urlz = host + cata + url[0]
				}
				AppendUrl(urlz, source)
				if num < 2 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(urlz, num+1)
				}
			} else if strings.HasSuffix(source, ".js") {
				AppendUrl(result.Jsinurl[host+path]+url[0], source)
				if num < 2 && (cmd.M == 2 || cmd.M == 3) {
					config.Wg.Add(1)
					config.Ch <- 1
					go Spider(result.Jsinurl[host+path]+url[0], num+1)
				}
			}
		}
	}
}

// 分析内容中的敏感信息
func infoFind(cont, source string) {
	info := mode.Info{}
	//手机号码
	for i := range config.Phone {
		phones := regexp.MustCompile(config.Phone[i]).FindAllStringSubmatch(cont, -1)
		for i := range phones {
			info.Phone = append(info.Phone, phones[i][1])
		}
	}

	for i := range config.Email {
		emails := regexp.MustCompile(config.Email[i]).FindAllStringSubmatch(cont, -1)
		for i := range emails {
			info.Email = append(info.Email, emails[i][1])
		}
	}

	for i := range config.IDcard {
		IDcards := regexp.MustCompile(config.IDcard[i]).FindAllStringSubmatch(cont, -1)
		for i := range IDcards {
			info.IDcard = append(info.IDcard, IDcards[i][1])
		}
	}

	for i := range config.Jwt {
		Jwts := regexp.MustCompile(config.Jwt[i]).FindAllStringSubmatch(cont, -1)
		for i := range Jwts {
			info.JWT = append(info.JWT, Jwts[i][1])
		}
	}

	info.Source = source
	if len(info.Phone) != 0 || len(info.IDcard) != 0 || len(info.JWT) != 0 || len(info.Email) != 0 {
		AppendInfo(info)
	}

}
