package crawler

import (
	"regexp"
	"strings"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
)

// 分析内容中的js
func jsFind(cont, host, scheme, path, source string, num int, judge_base bool) {
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

			// base标签的处理 ####
			if judge_base {
				js[0] = path + js[0]
			}

			if strings.HasPrefix(js[0], "https:") || strings.HasPrefix(js[0], "http:") {
				switch AppendJs(js[0], source) {
				case 0:
					if num <= config.JsSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(js[0], num+1)
					}
				case 1:
					return
				case 2:
					continue
				}

			} else if strings.HasPrefix(js[0], "//") {
				switch AppendJs(scheme+":"+js[0], source) {
				case 0:
					if num <= config.JsSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(scheme+":"+js[0], num+1)
					}
				case 1:
					return
				case 2:
					continue
				}

			} else if strings.HasPrefix(js[0], "/") {
				switch AppendJs(host+js[0], source) {
				case 0:
					if num <= config.JsSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(host+js[0], num+1)
					}
				case 1:
					return
				case 2:
					continue
				}

			} else {
				switch AppendJs(host+cata+js[0], source) {
				case 0:
					if num <= config.JsSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(host+cata+js[0], num+1)
					}
				case 1:
					return
				case 2:
					continue
				}

			}
		}

	}

}

// 分析内容中的url
func urlFind(cont, host, scheme, path, source string, num int, judge_base bool) {
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
		urls = urlFilter(urls)

		//循环提取url放到结果中
		for _, url := range urls {
			if url[0] == "" {
				continue
			}

			// base标签的处理 ####
			if judge_base {
				url[0] = path + url[0]
			}

			if strings.HasPrefix(url[0], "https:") || strings.HasPrefix(url[0], "http:") {
				switch AppendUrl(url[0], source) {
				case 0:
					if num <= config.UrlSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(url[0], num+1)
					}
				case 1:
					return
				case 2:
					continue
				}
			} else if strings.HasPrefix(url[0], "//") {
				switch AppendUrl(scheme+":"+url[0], source) {
				case 0:
					if num <= config.UrlSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(scheme+":"+url[0], num+1)
					}
				case 1:
					return
				case 2:
					continue
				}

			} else if strings.HasPrefix(url[0], "/") {
				urlz := ""
				if cmd.B != "" {
					urlz = cmd.B + url[0]
				} else {
					urlz = host + url[0]
				}
				switch AppendUrl(urlz, source) {
				case 0:
					if num <= config.UrlSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(urlz, num+1)
					}
				case 1:
					return
				case 2:
					continue
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
				switch AppendUrl(urlz, source) {
				case 0:
					if num <= config.UrlSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(urlz, num+1)
					}
				case 1:
					return
				case 2:
					continue
				}

			} else if strings.HasSuffix(source, ".js") {
				urlz := ""
				if cmd.B != "" {
					if strings.HasSuffix(cmd.B, "/") {
						urlz = cmd.B + url[0]
					} else {
						urlz = cmd.B + "/" + url[0]
					}
				} else {
					config.Lock.Lock()
					su := result.Jsinurl[source]
					config.Lock.Unlock()
					if strings.HasSuffix(su, "/") {
						urlz = su + url[0]
					} else {
						urlz = su + "/" + url[0]
					}
				}
				switch AppendUrl(urlz, source) {
				case 0:
					if num <= config.UrlSteps && (cmd.M == 2 || cmd.M == 3) {
						config.Wg.Add(1)
						config.Ch <- 1
						go Spider(urlz, num+1)
					}
				case 1:
					return
				case 2:
					continue
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
	for i := range config.Other {
		Others := regexp.MustCompile(config.Other[i]).FindAllStringSubmatch(cont, -1)
		for i := range Others {
			info.Other = append(info.Other, Others[i][1])
		}
	}

	info.Source = source
	if len(info.Phone) != 0 || len(info.IDcard) != 0 || len(info.JWT) != 0 || len(info.Email) != 0 || len(info.Other) != 0 {
		AppendInfo(info)
	}

}
