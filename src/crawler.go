package main

import (
	"crypto/tls"
	"encoding/base64"
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
	Proxy   map[string]string `yaml:"proxy"`
}

var (
	risks   = []string{"remove", "delete", "insert", "update", "logout"}
	fuzzs   = [][]string{}
	fuzzNum int
)
var ua = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

var conf config

// 蜘蛛抓取页面内容
func spider(u string, is bool) {
	fmt.Printf("\rSpider %d", progress)
	mux.Lock()
	progress++
	mux.Unlock()

	//标记完成
	defer func() {
		wg.Done()
		//<- ch
	}()
	u, _ = url.QueryUnescape(u)
	if getEndUrl(u) {
		<-ch
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				<-ch
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
		split := strings.Split(x, "|")
		proxyUrl, parseErr := url.Parse(split[0])
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		if len(split) == 2 {
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(split[1]))
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)
		}
	}
	//加载yaml配置(proxy)
	if I {
		SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		ch <- 1
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
		<-ch
		return
	} else {
		defer response.Body.Close()

	}

	//提取url用于拼接其他url或js
	dataBytes, err := io.ReadAll(response.Body)
	if err != nil {
		<-ch
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
	jsFind(result, host, scheme, path, source, is)
	//提取url
	urlFind(result, host, scheme, path, source, is)

}

// 分析内容中的js
func jsFind(cont, host, scheme, path, source string, is bool) {
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
		".(https{0,1}:[^\\s,^',^’,^\",^”,^>,^<,^;,^(,^),^|,^*,^\\[]{2,250}?[^=,^*,^\\s,^',^’,^\",^”,^>,^<,^:,^;,^*,^|,^(,^),^\\[]{3}[.]js)",
		"[\",',‘,“]\\s{0,6}(/{0,1}[^\\s,^',^’,^\",^”,^|,^>,^<,^:,^;,^*,^(,^\\),^\\[]{2,250}?[^=,^*,^\\s,^',^’,^|,^\",^”,^>,^<,^:,^;,^*,^(,^),^\\[]{3}[.]js)",
		"=\\s{0,6}[\",',’,”]{0,1}\\s{0,6}(/{0,1}[^\\s,^',^’,^\",^”,^|,^>,^<,^;,^*,^(,^),^\\[]{2,250}?[^=,^*,^\\s,^',^’,^\",^”,^>,^|,^<,^:,^;,^*,^(,^),^\\[]{3}[.]js)",
	}
	host = scheme + "://" + host
	for _, re := range res {
		re := regexp.MustCompile(re)
		jss := re.FindAllStringSubmatch(cont, -1)
		jss = jsFilter(jss)
		//循环提取js放到结果中
		for _, js := range jss {
			if js[0] == "" {
				continue
			}
			if strings.HasPrefix(js[0], "https:") || strings.HasPrefix(js[0], "http:") {
				appendJs(js[0], source)
				if is || (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(js[0], false)
				}
			} else if strings.HasPrefix(js[0], "//") {
				appendJs(scheme+":"+js[0], source)
				if is || (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(scheme+":"+js[0], false)
				}

			} else if strings.HasPrefix(js[0], "/") {
				appendJs(host+js[0], source)
				if is || (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+js[0], false)
				}
			} else {
				appendJs(host+cata+js[0], source)
				if is || (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+cata+js[0], false)
				}
			}
		}

	}

}

// 分析内容中的url
func urlFind(cont, host, scheme, path, source string, is bool) {
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
		"[\",',‘,“]\\s{0,6}(https{0,1}:[^\\s,^',^’,^\",^”,^>,^<,^),^(]{2,250}?)\\s{0,6}[\",',‘,“]",
		"=\\s{0,6}(https{0,1}:[^\\s,^',^’,^\",^”,^>,^<,^),^(]{2,250})",
		"[\",',‘,“]\\s{0,6}([#,.]{0,2}/[^\\s,^',^’,^\",^”,^>,^<,^:,^),^(]{2,250}?)\\s{0,6}[\",',‘,“]",
		"href\\s{0,6}=\\s{0,6}[\",',‘,“]{0,1}\\s{0,6}([^\\s,^',^’,^\",^“,^>,^<,^,^+),^(]{2,250})|action\\s{0,6}=\\s{0,6}[\",',‘,“]{0,1}\\s{0,6}([^\\s,^',^’,^\",^“,^>,^<,^,^+),^(]{2,250})",
	}
	for _, re := range res {
		re := regexp.MustCompile(re)
		urls := re.FindAllStringSubmatch(cont, -1)
		urls = urlFilter(urls)
		//循环提取url放到结果中
		for _, url := range urls {
			if url[0] == "" {
				continue
			}
			if strings.HasPrefix(url[0], "https:") || strings.HasPrefix(url[0], "http:") {
				appendUrl(url[0], source)
				if is && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(url[0], false)
				}
			} else if strings.HasPrefix(url[0], "//") {
				appendUrl(scheme+":"+url[0], source)
				if is && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(scheme+":"+url[0], false)
				}
			} else if strings.HasPrefix(url[0], "/") {
				appendUrl(host+url[0], source)
				if is && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+url[0], false)
				}
			} else if !strings.HasSuffix(source, ".js") {
				appendUrl(host+cata+url[0], source)
				if is && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(host+cata+url[0], false)
				}
			} else if strings.HasSuffix(source, ".js") {
				appendUrl(jsinurl[host+path]+url[0], source)
				if is && (m == 2 || m == 3) {
					wg.Add(1)
					ch <- 1
					go spider(jsinurl[host+path]+url[0], false)
				}
			}
		}

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
		fstr := []string{".js?", ".css?", ".jpeg?", ".jpg?", ".png?", ".gif?", "github.com", "www.w3.org", "example.com", "<", ">", "{", "}", "[", "]", "|", "^", ";", "/js/", ".src", ".url", ".att", ".href", "location.href", "javascript:", "location:", ".createObject", ":location", ".path", "*#__PURE__*", "\\n"}
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

	}
	return str
}

// 检测js访问状态码
func jsState(u string, i int) {
	defer func() {
		wg.Done()
		<-jsch
		printProgress()
	}()
	if s == "" {
		resultJs[i][0] = u
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				resultJs[i] = []string{u, "", "", "", "疑似危险路由"}
				return
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	if x != "" {
		split := strings.Split(x, "|")
		proxyUrl, parseErr := url.Parse(split[0])
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		if len(split) == 2 {
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(split[1]))
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)
		}
	}
	//加载yaml配置(proxy)
	if I {
		SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		resultJs[i][0] = ""
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
			resultJs[i] = []string{u, "timeout"}

		} else {
			resultJs[i][0] = ""
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
		resultJs[i] = []string{u, strconv.Itoa(code), strconv.Itoa(length)}
	} else {
		resultJs[i][0] = ""
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
		resultUrl[i][0] = u
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				resultUrl[i] = []string{u, "0", "0", "疑似危险路由，已跳过验证"}
				return
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	if x != "" {
		split := strings.Split(x, "|")
		proxyUrl, parseErr := url.Parse(split[0])
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		if len(split) == 2 {
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(split[1]))
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)
		}
	}
	//加载yaml配置(proxy)
	if I {
		SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		resultUrl[i][0] = ""
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
			resultUrl[i] = []string{u, "timeout"}
		} else {
			resultUrl[i][0] = ""
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
		re := regexp.MustCompile("<title>(.*?)</title>")
		title := re.FindAllStringSubmatch(body, -1)
		if len(title) != 0 {
			resultUrl[i] = []string{u, strconv.Itoa(code), strconv.Itoa(length), title[0][1]}
		} else {
			resultUrl[i] = []string{u, strconv.Itoa(code), strconv.Itoa(length)}
		}
	} else {
		resultUrl[i][0] = ""
	}
}

// fuzz
func fuzz() {
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(u, 1)
	if len(hosts) == 0 {
		host = u
	} else {
		host = hosts[0]
	}
	if d != "" {
		host = d
	}
	disposes, _ := urlDispose(resultUrl, host, "")
	if z == 2 || z == 3 {
		fuzz2(disposes)
	} else if z != 0 {
		fuzz1(disposes)
	}

	fmt.Println("\rFuzz OK      ")
}

// fuzz请求
func fuzzGet(u string) {
	defer func() {
		wg.Done()
		<-ch
		printFuzz()
	}()
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(u, v) {
				return
			}
		}
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	if x != "" {
		split := strings.Split(x, "|")
		proxyUrl, parseErr := url.Parse(split[0])
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		if len(split) == 2 {
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(split[1]))
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)
		}
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
	//加载yaml配置
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
	code := response.StatusCode
	if code != 404 && (strings.Contains(s, strconv.Itoa(code)) || s == "all") {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}
		body := string(dataBytes)
		re := regexp.MustCompile("<title>(.*?)</title>")
		title := re.FindAllStringSubmatch(body, -1)
		lock.Lock()
		if len(title) != 0 {
			fuzzs = append(fuzzs, []string{u, strconv.Itoa(code), strconv.Itoa(length), title[0][1]})
		} else {
			fuzzs = append(fuzzs, []string{u, strconv.Itoa(code), strconv.Itoa(length)})
		}
		lock.Unlock()
	}

}
func fuzz1(disposes [][]string) {
	dispose404 := []string{}
	for _, v := range disposes {
		if v[1] == "404" {
			dispose404 = append(dispose404, v[0])
		}
	}
	fuzz1s := []string{}
	host := ""
	if len(dispose404) != 0 {
		host = regexp.MustCompile("(http.{0,1}://.+?)/").FindAllStringSubmatch(dispose404[0]+"/", -1)[0][1]
	}

	for _, v := range dispose404 {
		submatch := regexp.MustCompile("http.{0,1}://.+?(/.*)").FindAllStringSubmatch(v, -1)
		if len(submatch) != 0 {
			v = submatch[0][1]
		} else {
			continue
		}
		v1 := v
		v2 := v
		reh2 := ""
		if strings.HasSuffix(v, "/") {
			submatch := regexp.MustCompile("/.+(/[^/]+)").FindAllStringSubmatch(v, -1)
			if len(submatch) != 0 {
				reh2 = submatch[0][1]
			} else {
				continue
			}
		}
		for {
			re1 := regexp.MustCompile("/.+?(/.+)").FindAllStringSubmatch(v1, -1)
			re2 := regexp.MustCompile("(/.+)/[^/]+").FindAllStringSubmatch(v2, -1)
			if len(re1) == 0 && len(re2) == 0 {
				break
			}
			if len(re1) > 0 {
				v1 = re1[0][1]
				fuzz1s = append(fuzz1s, host+v1)
			}
			if len(re2) > 0 {
				v2 = re2[0][1]
				fuzz1s = append(fuzz1s, host+v2+reh2)
			}
		}
	}
	fuzz1s = uniqueArr(fuzz1s)
	fuzzNum = len(fuzz1s)
	progress = 1
	fmt.Printf("Start %d Fuzz...\n", fuzzNum)
	fmt.Printf("\r                                           ")
	for _, v := range fuzz1s {
		wg.Add(1)
		ch <- 1
		go fuzzGet(v)
	}
	wg.Wait()
	fuzzs = del404(fuzzs)
}

func fuzz2(disposes [][]string) {
	disposex := []string{}
	dispose404 := []string{}
	for _, v := range disposes {
		if v[1] != "timeout" {
			disposex = append(disposex, v[0])
		}
		if v[1] == "404" {
			dispose404 = append(dispose404, v[0])
		}
	}
	dispose, _ := pathExtract(disposex)
	_, targets := pathExtract(dispose404)
	fuzzNum = len(dispose) * len(targets)
	progress = 1
	fmt.Printf("Start %d Fuzz...\n", len(dispose)*len(targets))
	fmt.Printf("\r                                           ")
	for _, v := range dispose {
		for _, vv := range targets {
			wg.Add(1)
			ch <- 1
			go fuzzGet(v + vv)
		}
	}
	wg.Wait()
	fuzzs = del404(fuzzs)
}
