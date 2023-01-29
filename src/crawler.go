package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type config struct {
	Headers map[string]string `yaml:"headers"`
	Proxy   string            `yaml:"proxy"`
}

var (
	risks   = []string{"remove", "delete", "insert", "update", "logout"}
	fuzzs   = []Link{}
	fuzzNum int
)
var ua = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

var conf config

// 蜘蛛抓取页面内容
func spider(u string, num int) {
	var is bool
	fmt.Printf("\rSpider %d", progress)
	mux.Lock()
	progress++
	mux.Unlock()
	//标记完成
	defer func() {
		wg.Done()
		if !is {
			<-ch
		}

	}()
	u, _ = url.QueryUnescape(u)
	if num > 1 && d != "" && !strings.Contains(u, d) {
		return
	}
	if getEndUrl(u) {
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				return
			}
		}
	}
	appendEndUrl(u)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	//配置代理
	if x != "" {
		proxyUrl, parseErr := url.Parse(conf.Proxy)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	//加载yaml配置(proxy)
	if I {
		SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	//增加header选项
	request.Header.Add("Cookie", c)
	request.Header.Add("User-Agent", ua)
	request.Header.Add("Accept", "*/*")
	//加载yaml配置（headers）
	if I {
		SetHeadersConfig(&request.Header)
	}
	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		return
	} else {
		defer response.Body.Close()

	}

	//提取url用于拼接其他url或js
	dataBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	path := response.Request.URL.Path
	host := response.Request.URL.Host
	scheme := response.Request.URL.Scheme
	source := scheme + "://" + host + path

	//字节数组 转换成 字符串
	result := string(dataBytes)
	//处理base标签
	re := regexp.MustCompile("base.{1,5}href.{1,5}(http.+?//[^\\s]+?)[\",',‘,“]")
	base := re.FindAllStringSubmatch(result, -1)
	if len(base) > 0 {
		host = regexp.MustCompile("http.*?//([^/]+)").FindAllStringSubmatch(base[0][1], -1)[0][1]
		scheme = regexp.MustCompile("(http.*?)://").FindAllStringSubmatch(base[0][1], -1)[0][1]
		paths := regexp.MustCompile("http.*?//.*?(/.*)").FindAllStringSubmatch(base[0][1], -1)
		if len(paths) > 0 {
			path = paths[0][1]
		} else {
			path = "/"
		}
	}
	<-ch
	is = true

	//提取js
	jsFind(result, host, scheme, path, source, num)
	//提取url
	urlFind(result, host, scheme, path, source, num)
	//提取信息
	infoFind(result, source)

}

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
	res := []string{
		".(https{0,1}:[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}?[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{3}[.]js)",
		"[\",',‘,“]\\s{0,6}(/{0,1}[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}?[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{3}[.]js)",
		"=\\s{0,6}[\",',’,”]{0,1}\\s{0,6}(/{0,1}[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{2,250}?[-a-zA-Z0-9（）@:%_\\+.~#?&//=]{3}[.]js)",
	}
	host = scheme + "://" + host
	for _, re := range res {
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
				appendJs(js[0], source)
				if num < 5 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(js[0], num+1)
				}
			} else if strings.HasPrefix(js[0], "//") {
				appendJs(scheme+":"+js[0], source)
				if num < 5 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(scheme+":"+js[0], num+1)
				}

			} else if strings.HasPrefix(js[0], "/") {
				appendJs(host+js[0], source)
				if num < 5 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+js[0], num+1)
				}
			} else if strings.HasPrefix(js[0], "./") {
				appendJs(host+"/"+js[0], source)
				if num < 5 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+"/"+js[0], num+1)
				}
			} else {
				appendJs(host+cata+js[0], source)
				if num < 5 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+cata+js[0], num+1)
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
	res := []string{
		"[\",',‘,“]\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//=]{2,250}?)\\s{0,6}[\",',‘,“]",
		"=\\s{0,6}(https{0,1}:[-a-zA-Z0-9()@:%_\\+.~#?&//=]{2,250})",
		"[\",',‘,“]\\s{0,6}([#,.]{0,2}/[-a-zA-Z0-9()@:%_\\+.~#?&//=]{2,250}?)\\s{0,6}[\",',‘,“]",
		"\"([-a-zA-Z0-9()@:%_\\+.~#?&//=]+?[/]{1}[-a-zA-Z0-9()@:%_\\+.~#?&//=]+?)\"",
		"href\\s{0,6}=\\s{0,6}[\",',‘,“]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//=]{2,250})|action\\s{0,6}=\\s{0,6}[\",',‘,“]{0,1}\\s{0,6}([-a-zA-Z0-9()@:%_\\+.~#?&//=]{2,250})",
	}
	for _, re := range res {
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
				appendUrl(url[0], source)
				if num < 2 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(url[0], num+1)
				}
			} else if strings.HasPrefix(url[0], "//") {
				appendUrl(scheme+":"+url[0], source)
				if num < 2 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(scheme+":"+url[0], num+1)
				}
			} else if strings.HasPrefix(url[0], "/") {
				urlz := ""
				if b != "" {
					urlz = b + url[0]
				} else {
					urlz = host + url[0]
				}
				appendUrl(urlz, source)
				if num < 2 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(urlz, num+1)
				}
			} else if !strings.HasSuffix(source, ".js") {
				urlz := ""
				if b != "" {
					if strings.HasSuffix(b, "/") {
						urlz = b + url[0]
					} else {
						urlz = b + "/" + url[0]
					}
				} else {
					urlz = host + cata + url[0]
				}
				appendUrl(urlz, source)
				if num < 2 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(urlz, num+1)
				}
			} else if strings.HasSuffix(source, ".js") {
				appendUrl(jsinurl[host+path]+url[0], source)
				if num < 2 && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(jsinurl[host+path]+url[0], num+1)
				}
			}
		}
	}
}

// 分析内容中的敏感信息
func infoFind(cont, source string) {
	//手机号码
	phone := "['\"](1(3([0-35-9]\\d|4[1-8])|4[14-9]\\d|5([\\d]\\d|7[1-79])|66\\d|7[2-35-8]\\d|8\\d{2}|9[89]\\d)\\d{7})['\"]"
	email := "['\"]([\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[\\w](?:[\\w-]*[\\w])?)['\"]"
	IDcard := "['\"]((\\d{8}(0\\d|10|11|12)([0-2]\\d|30|31)\\d{3}$)|(\\d{6}(18|19|20)\\d{2}(0[1-9]|10|11|12)([0-2]\\d|30|31)\\d{3}(\\d|X|x)))['\"]"
	jwt := "['\"](ey[A-Za-z0-9_-]{10,}\\.[A-Za-z0-9._-]{10,}|ey[A-Za-z0-9_\\/+-]{10,}\\.[A-Za-z0-9._\\/+-]{10,})['\"]"
	phones := regexp.MustCompile(phone).FindAllStringSubmatch(cont, -1)
	emails := regexp.MustCompile(email).FindAllStringSubmatch(cont, -1)
	IDcards := regexp.MustCompile(IDcard).FindAllStringSubmatch(cont, -1)
	jwts := regexp.MustCompile(jwt).FindAllStringSubmatch(cont, -1)
	info := Info{}
	for i := range phones {
		info.Phone = append(info.Phone, phones[i][1])
	}
	for i := range emails {
		info.Email = append(info.Email, emails[i][1])
	}
	for i := range IDcards {
		info.IDcard = append(info.IDcard, IDcards[i][1])
	}
	for i := range jwts {
		info.JWT = append(info.JWT, jwts[i][1])
	}
	info.Source = source
	if len(info.Phone) != 0 || len(info.IDcard) != 0 || len(info.JWT) != 0 || len(info.Email) != 0 {
		appendInfo(info)
	}

}

// 过滤JS
func jsFilter(str [][]string) [][]string {

	//对不需要的数据过滤
	for i := range str {
		str[i][0], _ = url.QueryUnescape(str[i][1])
		str[i][0] = strings.Replace(str[i][0], " ", "", -1)
		str[i][0] = strings.Replace(str[i][0], "\\/", "/", -1)
		str[i][0] = strings.Replace(str[i][0], "%3A", ":", -1)
		str[i][0] = strings.Replace(str[i][0], "%2F", "/", -1)

		match, _ := regexp.MatchString("[.]js", str[i][0])
		if !match {
			str[i][0] = ""
		}
		//过滤指定字段
		fstr := []string{"www.w3.org", "example.com", "github.com"}
		for _, v := range fstr {
			if strings.Contains(str[i][0], v) {
				str[i][0] = ""
			}
		}

	}
	return str

}

// jsfuzz
func jsFuzz() {
	jsFuzzFile := []string{
		"login.js",
		"app.js",
		"main.js",
		"config.js",
		"admin.js",
		"info.js",
		"open.js",
		"user.js",
		"input.js",
		"list.js",
		"upload.js"}
	paths := []string{}
	for i := range resultJs {
		re := regexp.MustCompile("(.+/)[^/]+.js").FindAllStringSubmatch(resultJs[i].Url, -1)
		if len(re) != 0 {
			paths = append(paths, re[0][1])
		}
	}
	paths = uniqueArr(paths)
	for i := range paths {
		for i2 := range jsFuzzFile {
			resultJs = append(resultJs, Link{
				Url:    paths[i] + jsFuzzFile[i2],
				Source: "Fuzz",
			})
		}
	}
}

// 过滤URL
func urlFilter(str [][]string) [][]string {

	//对不需要的数据过滤
	for i := range str {
		str[i][0], _ = url.QueryUnescape(str[i][1])
		str[i][0] = strings.Replace(str[i][0], " ", "", -1)
		str[i][0] = strings.Replace(str[i][0], "\\/", "/", -1)
		str[i][0] = strings.Replace(str[i][0], "%3A", ":", -1)
		str[i][0] = strings.Replace(str[i][0], "%2F", "/", -1)

		//过滤包含指定内容
		fstr := []string{".js?", ".css?", ".jpeg?", ".jpg?", ".png?", ".gif?", "github.com", "www.w3.org", "example.com", "<", ">", "{", "}", "[", "]", "|", "^", ";", "/js/", ".src", ".replace", ".url", ".att", ".href", "location.href", "javascript:", "location:", ".createObject", ":location", ".path", "*#__PURE__*", "*$0*", "\\n"}
		for _, v := range fstr {
			if strings.Contains(str[i][0], v) {
				str[i][0] = ""
			}
		}

		match, _ := regexp.MatchString("[a-zA-Z]+|[0-9]+", str[i][0])
		if !match {
			str[i][0] = ""
		}
		//过滤指定后缀
		zstr := []string{".js", ".css", ".scss", ",", ".jpeg", ".jpg", ".png", ".gif", ".ico", ".svg", ".vue", ".ts"}

		for _, v := range zstr {
			if strings.HasSuffix(str[i][0], v) {
				str[i][0] = ""
			}
		}
		//对抓到的域名做处理
		re := regexp.MustCompile("([a-z0-9\\-]+\\.)+([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?").FindAllString(str[i][0], 1)
		if len(re) != 0 && !strings.HasPrefix(str[i][0], "http") && !strings.HasPrefix(str[i][0], "/") {
			str[i][0] = "http://" + str[i][0]
		}

	}
	return str
}

// 检测js访问状态码
func jsState(u string, i int, sou string) {
	defer func() {
		wg.Done()
		<-jsch
		printProgress()
	}()
	if s == "" {
		resultJs[i].Url = u
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				resultJs[i] = Link{Url: u, Status: "疑似危险路由"}
				return
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	//配置代理
	if x != "" {
		proxyUrl, parseErr := url.Parse(conf.Proxy)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	//加载yaml配置(proxy)
	if I {
		SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		resultJs[i].Url = ""
		return
	}

	//增加header选项
	request.Header.Add("Cookie", c)
	request.Header.Add("User-Agent", ua)
	request.Header.Add("Accept", "*/*")
	//加载yaml配置
	if I {
		SetHeadersConfig(&request.Header)
	}

	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") && s == "" {
			resultJs[i] = Link{Url: u, Status: "timeout", Size: "0"}

		} else {
			resultJs[i].Url = ""
		}
		return
	} else {
		defer response.Body.Close()
	}

	code := response.StatusCode
	if strings.Contains(s, strconv.Itoa(code)) || s == "all" && (sou != "Fuzz" && code == 200) {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}
		resultJs[i] = Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length)}
	} else {
		resultJs[i].Url = ""
	}
}

// 检测url访问状态码
func urlState(u string, i int) {

	defer func() {
		wg.Done()
		<-urlch
		printProgress()
	}()
	if s == "" {
		resultUrl[i].Url = u
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				resultUrl[i] = Link{Url: u, Status: "0", Size: "0", Title: "疑似危险路由，已跳过验证"}
				return
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	if x != "" {
		proxyUrl, parseErr := url.Parse(conf.Proxy)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	//加载yaml配置(proxy)
	if I {
		SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		resultUrl[i].Url = ""
		return
	}
	//增加header选项
	request.Header.Add("Cookie", c)
	request.Header.Add("User-Agent", ua)
	request.Header.Add("Accept", "*/*")
	//加载yaml配置
	if I {
		SetHeadersConfig(&request.Header)
	}
	//处理返回结果
	response, err := client.Do(request)

	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") && s == "all" {
			resultUrl[i] = Link{Url: u, Status: "timeout", Size: "0"}
		} else {
			resultUrl[i].Url = ""
		}
		return
	} else {
		defer response.Body.Close()
	}

	code := response.StatusCode
	if strings.Contains(s, strconv.Itoa(code)) || s == "all" {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}
		body := string(dataBytes)
		re := regexp.MustCompile("<[tT]itle>(.*?)</[tT]itle>")
		title := re.FindAllStringSubmatch(body, -1)
		if len(title) != 0 {
			resultUrl[i] = Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: title[0][1]}
		} else {
			resultUrl[i] = Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length)}
		}
	} else {
		resultUrl[i].Url = ""
	}
}
