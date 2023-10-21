package crawler

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
)

// 检测js访问状态码
func JsState(u string, i int, sou string, rod_flag bool) {

	defer func() {
		config.Wg.Done()
		<-config.Jsch
		PrintProgress()
	}()

	// 首页动态加载的数据 已经存储
	if rod_flag {
		return
	}

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

	//加载yaml配置(proxy)
	//配置代理
	var redirect string
	ur, err2 := url.Parse(u)
	if err2 != nil {
		return
	}
	request, err := http.NewRequest("GET", ur.String(), nil)
	if err != nil {
		result.ResultJs[i].Url = ""
		return
	}
	if cmd.C != "" {
		request.Header.Set("Cookie", cmd.C)
	}
	//增加header选项
	request.Header.Set("User-Agent", util.GetUserAgent())
	request.Header.Set("Accept", "*/*")
	u_str, err := url.Parse(u)
	if err != nil {
		return
	}
	request.Header.Set("Referer", u_str.Scheme+"://"+u_str.Host) //####

	//加载yaml配置
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//client = &http.Client{Timeout: time.Duration(cmd.TI) * time.Second,
	//	Transport: tr,
	//	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//		if len(via) >= 10 {
	//			return fmt.Errorf("Too many redirects")
	//		}
	//		if len(via) > 0 {
	//			if via[0] != nil && via[0].URL != nil {
	//				result.Redirect[via[0].URL.String()] = true
	//			} else {
	//				result.Redirect[req.URL.String()] = true
	//			}
	//
	//		}
	//		return nil
	//	},
	//}
	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") && cmd.S == "" {
			result.ResultJs[i] = mode.Link{Url: u, Status: "timeout", Size: "0"}

		} else {
			result.ResultJs[i].Url = ""
		}
		return
	}
	defer response.Body.Close()

	code := response.StatusCode
	if strings.Contains(cmd.S, strconv.Itoa(code)) || cmd.S == "all" && (sou != "Fuzz" && code == 200) {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}

		res_body := string(dataBytes)
		res_headers := make(map[string]string, 0)
		// 遍历响应头中的所有键值对
		for k, v := range response.Header {
			// 如果值是一个切片，取第一个元素作为值，否则忽略该键值对
			if len(v) > 0 {
				res_headers[k] = v[0]
			}
		}

		config.Lock.Lock()
		if result.Redirect[ur.String()] {
			code = 302
			redirect = response.Request.URL.String()
		}
		config.Lock.Unlock()
		result.ResultJs[i] = mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Redirect: redirect, ResponseHeaders: res_headers, ResponseBody: res_body}
	} else {
		result.ResultJs[i].Url = ""
	}
}

// 检测url访问状态码
func UrlState(u string, i int, rod_flag bool) {
	defer func() {
		config.Wg.Done()
		<-config.Urlch
		PrintProgress()
	}()

	// 首页动态加载的数据 已经存储
	if rod_flag {
		return
	}
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

	var redirect string
	ur, err2 := url.Parse(u)
	if err2 != nil {
		return
	}
	request, err := http.NewRequest("GET", ur.String(), nil)
	if err != nil {
		result.ResultUrl[i].Url = ""
		return
	}

	if cmd.C != "" {
		request.Header.Set("Cookie", cmd.C)
	}
	//增加header选项
	request.Header.Set("User-Agent", util.GetUserAgent())
	request.Header.Set("Accept", "*/*")
	u_str, err := url.Parse(u)
	if err != nil {
		return
	}
	request.Header.Set("Referer", u_str.Scheme+"://"+u_str.Host) //####

	//加载yaml配置
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//client = &http.Client{Timeout: time.Duration(cmd.TI) * time.Second,
	//	Transport: tr,
	//	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//		if len(via) >= 10 {
	//			return fmt.Errorf("Too many redirects")
	//		}
	//		if len(via) > 0 {
	//			if via[0] != nil && via[0].URL != nil {
	//				result.Redirect[via[0].URL.String()] = true
	//			} else {
	//				result.Redirect[req.URL.String()] = true
	//			}
	//
	//		}
	//		return nil
	//	},
	//}
	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") && cmd.S == "all" {
			result.ResultUrl[i] = mode.Link{Url: u, Status: "timeout", Size: "0"}
		} else {
			result.ResultUrl[i].Url = ""
		}
		return
	}
	defer response.Body.Close()

	code := response.StatusCode
	if strings.Contains(cmd.S, strconv.Itoa(code)) || cmd.S == "all" {
		var length int
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			length = 0
		} else {
			length = len(dataBytes)
		}

		res_body := string(dataBytes)
		res_headers := make(map[string]string, 0)
		// 遍历响应头中的所有键值对
		for k, v := range response.Header {
			// 如果值是一个切片，取第一个元素作为值，否则忽略该键值对
			if len(v) > 0 {
				res_headers[k] = v[0]
			}
		}

		re := regexp.MustCompile("<[tT]itle>(.*?)</[tT]itle>")
		title := re.FindAllStringSubmatch(res_body, -1)
		config.Lock.Lock()
		if result.Redirect[ur.String()] {
			code = 302
			redirect = response.Request.URL.String()
		}
		config.Lock.Unlock()

		if len(title) != 0 {
			result.ResultUrl[i] = mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Title: title[0][1], Redirect: redirect, ResponseHeaders: res_headers, ResponseBody: res_body}
		} else {
			result.ResultUrl[i] = mode.Link{Url: u, Status: strconv.Itoa(code), Size: strconv.Itoa(length), Redirect: redirect, ResponseHeaders: res_headers, ResponseBody: res_body}
		}
	} else {
		result.ResultUrl[i].Url = ""
	}
}
