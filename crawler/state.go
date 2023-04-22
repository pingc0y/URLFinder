package crawler

import (
	"crypto/tls"
	"fmt"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 检测js访问状态码
func JsState(u string, i int, sou string) {
	defer func() {
		config.Wg.Done()
		<-config.Jsch
		PrintProgress()
	}()
	if cmd.S == "" {
		result.ResultJs[i].Url = u
		return
	}
	if cmd.M == 3 {
		for _, v := range config.Risks {
			if strings.Contains(u, v) {
				result.ResultJs[i] = mode.Link{Url: u, Status: "疑似危险路由"}
				return
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	if cmd.X != "" {
		proxyUrl, parseErr := url.Parse(cmd.X)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	//加载yaml配置(proxy)
	if cmd.I {
		util.SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		result.ResultJs[i].Url = ""
		return
	}
	if cmd.C != "" {
		request.Header.Add("Cookie", cmd.C)
	}
	//增加header选项
	request.Header.Add("User-Agent", util.GetUserAgent())
	request.Header.Add("Accept", "*/*")
	//加载yaml配置
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}

	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") && cmd.S == "" {
			result.ResultJs[i] = mode.Link{Url: u, Status: "timeout", Size: "0"}

		} else {
			result.ResultJs[i].Url = ""
		}
		return
	} else {
		defer response.Body.Close()
	}

	code := response.StatusCode
	if strings.Contains(cmd.S, strconv.Itoa(code)) || cmd.S == "all" && (sou != "Fuzz" && code == 200) {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}
		result.ResultJs[i] = mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length)}
	} else {
		result.ResultJs[i].Url = ""
	}
}

// 检测url访问状态码
func UrlState(u string, i int) {

	defer func() {
		config.Wg.Done()
		<-config.Urlch
		PrintProgress()
	}()
	if cmd.S == "" {
		result.ResultUrl[i].Url = u
		return
	}
	if cmd.M == 3 {
		for _, v := range config.Risks {
			if strings.Contains(u, v) {
				result.ResultUrl[i] = mode.Link{Url: u, Status: "0", Size: "0", Title: "疑似危险路由,已跳过验证"}
				return
			}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//配置代理
	if cmd.X != "" {
		proxyUrl, parseErr := url.Parse(cmd.X)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	//加载yaml配置(proxy)
	if cmd.I {
		util.SetProxyConfig(tr)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		result.ResultUrl[i].Url = ""
		return
	}

	if cmd.C != "" {
		request.Header.Add("Cookie", cmd.C)
	}
	//增加header选项
	request.Header.Add("User-Agent", util.GetUserAgent())
	request.Header.Add("Accept", "*/*")

	//加载yaml配置
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}
	//处理返回结果
	response, err := client.Do(request)

	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") && cmd.S == "all" {
			result.ResultUrl[i] = mode.Link{Url: u, Status: "timeout", Size: "0"}
		} else {
			result.ResultUrl[i].Url = ""
		}
		return
	} else {
		defer response.Body.Close()
	}

	code := response.StatusCode
	if strings.Contains(cmd.S, strconv.Itoa(code)) || cmd.S == "all" {
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
			result.ResultUrl[i] = mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: title[0][1]}
		} else {
			result.ResultUrl[i] = mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length)}
		}
	} else {
		result.ResultUrl[i].Url = ""
	}
}
