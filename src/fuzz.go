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
	disposes, _ := urlDispose(append(resultUrl, Link{Url: u, Status: "200", Size: "0"}), host, "")
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
	if strings.Contains(s, strconv.Itoa(code)) || s == "all" {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}
		body := string(dataBytes)
		re := regexp.MustCompile("<Title>(.*?)</Title>")
		title := re.FindAllStringSubmatch(body, -1)
		lock.Lock()
		if len(title) != 0 {
			fuzzs = append(fuzzs, Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: title[0][1], Source: "Fuzz"})
		} else {
			fuzzs = append(fuzzs, Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: "", Source: "Fuzz"})
		}
		lock.Unlock()
	}

}
func fuzz1(disposes []Link) {
	dispose404 := []string{}
	for _, v := range disposes {
		if v.Status == "404" {
			dispose404 = append(dispose404, v.Url)
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
		if !strings.HasSuffix(v, "/") {
			_submatch := regexp.MustCompile("/.+(/[^/]+)").FindAllStringSubmatch(v, -1)
			if len(_submatch) != 0 {
				reh2 = _submatch[0][1]
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

func fuzz2(disposes []Link) {
	disposex := []string{}
	dispose404 := []string{}
	for _, v := range disposes {
		if v.Status == "404" {
			dispose404 = append(dispose404, v.Url)
		}
		//防止太多跑不完
		if len(dispose404) > 20 {
			if v.Status != "timeout" && v.Status != "404" {
				disposex = append(disposex, v.Url)
			}
		} else {
			if v.Status != "timeout" {
				disposex = append(disposex, v.Url)
			}
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
