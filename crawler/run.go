package crawler

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
)

var client *http.Client

// 全局变量 存储body
var ResBodyMap = make(map[string]string, 0)
var ResHeaderMap = make(map[string]proto.NetworkHeaders, 0)

func load() {

	if cmd.I {
		config.GetConfig("config.yaml")
	}
	if cmd.H {
		flag.Usage()
		os.Exit(0)
	}
	if cmd.U == "" && cmd.F == "" && cmd.FF == "" {
		fmt.Println("至少使用 -u -f -ff 指定一个url")
		os.Exit(0)
	}
	u, ok := url.Parse(cmd.U)
	if cmd.U != "" && ok != nil {
		fmt.Println("url格式错误,请填写正确url")
		os.Exit(0)
	}
	cmd.U = u.String()

	if cmd.T != 50 {
		config.Ch = make(chan int, cmd.T)
		config.Jsch = make(chan int, cmd.T/10*3)
		config.Urlch = make(chan int, cmd.T/10*7)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 30,
			KeepAlive: time.Second * 30,
		}).DialContext,
		MaxIdleConns:          cmd.T / 2,
		MaxIdleConnsPerHost:   cmd.T + 10,
		IdleConnTimeout:       time.Second * 90,
		TLSHandshakeTimeout:   time.Second * 90,
		ExpectContinueTimeout: time.Second * 10,
	}

	if cmd.X != "" {
		tr.DisableKeepAlives = true
		proxyUrl, parseErr := url.Parse(cmd.X)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	if cmd.I {
		util.SetProxyConfig(tr)
	}
	client = &http.Client{Timeout: time.Duration(cmd.TI) * time.Second,
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("Too many redirects")
			}
			if len(via) > 0 {
				if via[0] != nil && via[0].URL != nil {
					AddRedirect(via[0].URL.String())
				} else {
					AddRedirect(req.URL.String())
				}

			}
			return nil
		},
	}

}

func Run() {
	load()
	if cmd.F != "" {
		// 创建句柄
		fi, err := os.Open(cmd.F)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		r := bufio.NewReader(fi) // 创建 Reader
		for {
			lineBytes, err := r.ReadBytes('\n')
			//去掉字符串首尾空白字符,返回字符串
			if len(lineBytes) > 5 {
				line := util.GetProtocol(strings.TrimSpace(string(lineBytes)))
				cmd.U = line
				Initialization()
				start(cmd.U)
				Res()
				fmt.Println("----------------------------------------")
			}
			if err == io.EOF {
				break
			}

		}
		return
	}
	if cmd.FF != "" {
		// 创建句柄
		fi, err := os.Open(cmd.FF)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		r := bufio.NewReader(fi) // 创建 Reader
		Initialization()
		for {
			lineBytes, err := r.ReadBytes('\n')
			//去掉字符串首尾空白字符,返回字符串
			if len(lineBytes) > 5 {
				line := util.GetProtocol(strings.TrimSpace(string(lineBytes)))
				if cmd.U == "" {
					cmd.U = line
				}
				startFF(line)
				fmt.Println("----------------------------------------")
			}

			if err == io.EOF {
				break
			}
		}
		ValidateFF()
		Res()
		return
	}
	Initialization()
	cmd.U = util.GetProtocol(cmd.U)
	start(cmd.U)
	Res()
}
func startFF(u string) {
	fmt.Println("Target URL: " + u)
	config.Wg.Add(1)
	config.Ch <- 1
	go Spider(u, 1)
	config.Wg.Wait()
	config.Progress = 1
	fmt.Printf("\r\nSpider OK \n")
}

func ValidateFF() {
	result.ResultUrl = util.RemoveRepeatElement(result.ResultUrl)
	result.ResultJs = util.RemoveRepeatElement(result.ResultJs)
	if cmd.S != "" {
		fmt.Printf("Start %d Validate...\n", len(result.ResultUrl)+len(result.ResultJs))
		fmt.Printf("\r                    ")
		JsFuzz()
		//验证JS状态
		for i, s := range result.ResultJs {
			config.Wg.Add(1)
			config.Jsch <- 1
			// 判断响应数据是否已经在页面加载过程存储
			rod_flag := false
			if len(ResBodyMap[s.Url]) != 0 {
				rod_flag = true
			}
			go JsState(s.Url, i, result.ResultJs[i].Source, rod_flag)
		}
		//验证URL状态
		for i, s := range result.ResultUrl {
			config.Wg.Add(1)
			config.Urlch <- 1
			// 判断响应数据是否已经在页面加载过程存储
			rod_flag := false
			if len(ResBodyMap[s.Url]) != 0 {
				rod_flag = true
			}
			go UrlState(s.Url, i, rod_flag)
		}
		config.Wg.Wait()

		time.Sleep(1 * time.Second)
		fmt.Printf("\r                                           ")
		fmt.Printf("\rValidate OK \n\n")

		if cmd.Z != 0 {
			UrlFuzz()
			time.Sleep(1 * time.Second)
		}
	}
	AddSource()
}

// 定义函数 url_parse，参数是一个字符串 u，返回值是三个字符串
func url_parse(u string) (string, string, string) {
	// 解析 u 为一个 URL 对象
	u_str, err := url.Parse(u)
	// 如果解析出错，就返回空字符串
	if err != nil {
		return "", "", ""
	}
	// 获取 URL 对象的 host、scheme、path 属性
	host := u_str.Host
	scheme := u_str.Scheme
	path := u_str.Path
	// 返回这三个属性的值
	return host, scheme, path
}

// 提取响应体中的 Base 标签信息
func extractBase(host, scheme, path, result string) (string, string, string, bool) {
	judge_base := false
	// 处理base标签
	re := regexp.MustCompile("base.{1,5}href.{1,5}(http.+?//[^\\s]+?)[\"'‘“]")
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
	} else { // 处理 "base 标签"
		re := regexp.MustCompile("(?i)base.{0,5}[:=]\\s*\"(.*?)\"")
		base := re.FindAllStringSubmatch(result, -1)
		if len(base) > 0 {
			pattern := "[^.\\/\\w]"
			re, _ := regexp.Compile(pattern)
			// 检查字符串是否包含匹配的字符
			result := re.MatchString(base[0][1])
			if !result { // 字符串中没有其他特殊字符
				if len(base[0][1]) > 1 && base[0][1][:2] == "./" { // base 路径从当前目录出发
					judge_base = true
					path = path[:strings.LastIndex(path, "/")] + base[0][1][1:]
				} else if len(base[0][1]) > 2 && base[0][1][:3] == "../" { // base 路径从上一级目录出发
					judge_base = true
					pattern := "^[./]+$"
					matched, _ := regexp.MatchString(pattern, base[0][1])
					if matched { // 处理的 base 路径中只有 ./的
						path = path[:strings.LastIndex(path, "/")+1] + base[0][1]
					} else {
						find_str := ""
						if strings.Contains(strings.TrimPrefix(base[0][1], "../"), "/") {
							find_str = base[0][1][3 : strings.Index(strings.TrimPrefix(base[0][1], "../"), "/")+3]
						} else {
							find_str = base[0][1][3:]
						}
						if strings.Contains(path, find_str) {
							path = path[:strings.Index(path, find_str)] + base[0][1][3:]
						} else {
							path = path[:strings.LastIndex(path, "/")+1] + base[0][1]
						}
					}
				} else if len(base[0][1]) > 4 && strings.HasPrefix(base[0][1], "http") { // base 标签包含协议
					judge_base = true
					path = base[0][1]
				} else if len(base[0][1]) > 0 {
					judge_base = true
					if base[0][1][0] == 47 { //base 路径从根目录出发
						path = base[0][1]
					} else { //base 路径未指明从哪路出发
						find_str := ""
						if strings.Contains(base[0][1], "/") {
							find_str = base[0][1][:strings.Index(base[0][1], "/")]
						} else {
							find_str = base[0][1]
						}
						if strings.Contains(path, find_str) {
							path = path[:strings.Index(path, find_str)] + base[0][1]
						} else {
							path = path[:strings.LastIndex(path, "/")+1] + base[0][1]
						}
					}
				}
				if !strings.HasSuffix(path, "/") {
					path += "/"
				}
			}
		}
	}
	return host, scheme, path, judge_base
}

// 获取网页加载的事件的响应体
func rod_spider(u string, num int) {
	// 初始化浏览器
	launch := launcher.New().Headless(true).Set("test-type").Set("ignore-certificate-errors").
		NoSandbox(true).Set("disable-gpu").Set("disable-plugins").Set("incognito").
		Set("no-default-browser-check").Set("disable-dev-shm-usage").
		Set("disable-plugins").MustLaunch()
	browser := rod.New().ControlURL(launch).MustConnect()

	// 添加关闭
	defer browser.Close()

	// 设置浏览器的证书错误处理，忽略所有证书错误
	browser.MustIgnoreCertErrors(true)

	// 设置浏览器打开的页面
	pageTarget := proto.TargetCreateTarget{URL: u}
	page, err := browser.Page(pageTarget)
	if err != nil {
		fmt.Println(err)
	}

	// 在最后关闭页面
	defer func() {
		err := page.Close()
		if err != nil {
			// 处理错误
			fmt.Println(err)
		}
	}()

	// 设置页面的超时时间为 40 秒
	page = page.Timeout(40 * time.Second)

	// 创建一个空的 map，键是 proto.NetworkRequestID 类型，值是 string 类型
	requestMap := make(map[string]string, 0)

	// 使用 go 语句开启一个协程，在协程中处理页面的一些事件
	go page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		// 处理 JavaScript 对话框
		_ = proto.PageHandleJavaScriptDialog{Accept: true, PromptText: ""}.Call(page)
	}, func(e *proto.NetworkResponseReceived) {
		// 获取请求的 ID 和 URL
		ResponseURL := e.Response.URL
		// fmt.Println(e.Response.URL, e.RequestID)
		ResHeaderMap[ResponseURL] = e.Response.Headers

		// 在 requestMap 中填充数据
		requestMap[ResponseURL] = ""

	})()

	// 等待页面加载完成，并处理可能出现的错误
	pageLoadErr := page.WaitLoad()
	if pageLoadErr != nil {
		fmt.Println(pageLoadErr)
	}

	// 等待页面的 DOM 结构稳定
	page.WaitStable(2 * time.Second)

	// 打印页面源码
	htmlStr, err := page.HTML()
	if err != nil {
		fmt.Println(err)
	}

	for url, _ := range requestMap {
		// 调用 page.GetResource 方法来获取响应体
		ResponseBody, _ := page.GetResource(url)
		requestMap[url] = string(ResponseBody)
	}

	// 存储页面源码
	requestMap[u] = string(htmlStr)
	// fmt.Println(requestMap[u])

	// 遍历响应体，提取 Base 标签、提取 js 、提取 url 、
	for url, body := range requestMap {
		// 判断响应体是否为空
		if len(body) == 0 {
			continue
		}

		// 遍历 BodyFiler 切片中的每个元素
		re := regexp.MustCompile("\\.jpeg\\?|\\.jpg\\?|\\.png\\?|.gif\\?|www\\.w3\\.org|example\\.com|.*,$|.*\\.jpeg$|.*\\.jpg$|.*\\.png$|.*\\.gif$|.*\\.ico$|.*\\.svg$|.*\\.vue$|.*\\.ts$")
		if re.MatchString(url) {
			continue
		}

		// fmt.Println("目标url及响应体信息:  ", url, len(body))

		// 添加body数据
		ResBodyMap[url] = body

		// 将响应头数据转换成map存储
		Res_header := make(map[string]string, 0)
		if len(ResHeaderMap[url]) != 0 {
			data, err := json.Marshal(ResHeaderMap[url])
			if err != nil {
				fmt.Println(err)
			}
			err = json.Unmarshal(data, &Res_header)
			if err != nil {
				fmt.Println(err)
			}
		}

		// 添加首页动态加载的数据
		if strings.HasSuffix(url, ".js") || strings.Contains(url, ".js?") {
			result.ResultJs = append(result.ResultJs, mode.Link{Url: url, Status: strconv.Itoa(200), Size: strconv.Itoa(len(body)), ResponseHeaders: Res_header, ResponseBody: body})
			// AppendJs(url, u)
		} else {
			result.ResultUrl = append(result.ResultUrl, mode.Link{Url: url, Status: strconv.Itoa(200), Size: strconv.Itoa(len(body)), ResponseHeaders: Res_header, ResponseBody: body})
			// AppendUrl(url, u)
		}

		host, scheme, path := url_parse(url)

		judge_base := false
		host, scheme, path, judge_base = extractBase(host, scheme, path, body)

		//提取js
		jsFind(body, host, scheme, path, u, num, judge_base)
		//提取url
		urlFind(body, host, scheme, path, u, num, judge_base)
		// 防止base判断错误
		if judge_base {
			jsFind(body, host, scheme, path, u, num, false)
			urlFind(body, host, scheme, path, u, num, false)
		}

	}

}

func start(u string) {
	fmt.Println("Target URL: " + u)

	// config.Wg.Add(1)
	// config.Ch <- 1
	// go Spider(u, 1) // ###
	rod_spider(u, 1)
	// config.Wg.Wait()
	// config.Progress = 1

	fmt.Printf("\r\nRod_Spider OK \n")
	result.ResultUrl = util.RemoveRepeatElement(result.ResultUrl)
	result.ResultJs = util.RemoveRepeatElement(result.ResultJs)
	if cmd.S != "" {
		fmt.Printf("Start %d Validate...\n", len(result.ResultUrl)+len(result.ResultJs))
		fmt.Printf("\r                    ")
		JsFuzz()
		//验证JS状态
		for i, s := range result.ResultJs {
			config.Wg.Add(1)
			config.Jsch <- 1
			// 判断响应数据是否已经在页面加载过程存储
			rod_flag := false
			if len(ResBodyMap[s.Url]) != 0 {
				rod_flag = true
			}
			go JsState(s.Url, i, result.ResultJs[i].Source, rod_flag)
		}
		//验证URL状态
		for i, s := range result.ResultUrl {
			config.Wg.Add(1)
			config.Urlch <- 1
			// 判断响应数据是否已经在页面加载过程存储
			rod_flag := false
			if len(ResBodyMap[s.Url]) != 0 {
				rod_flag = true
			}
			go UrlState(s.Url, i, rod_flag)
		}
		config.Wg.Wait()

		time.Sleep(1 * time.Second)
		fmt.Printf("\r                                           ")
		fmt.Printf("\rValidate OK \n\n")

		if cmd.Z != 0 {
			UrlFuzz()
			time.Sleep(1 * time.Second)
		}
	}
	AddSource()
}

func Res() {
	if len(result.ResultJs) == 0 && len(result.ResultUrl) == 0 {
		fmt.Println(os.Stdout, cmd.U, "Data not captured")
		return
	}
	//打印还是输出
	if len(cmd.O) > 0 {
		if strings.HasSuffix(cmd.O, ".json") {
			result.OutFileJson(cmd.O)
		} else if strings.HasSuffix(cmd.O, ".html") {
			result.OutFileHtml(cmd.O)
		} else if strings.HasSuffix(cmd.O, ".csv") {
			result.OutFileCsv(cmd.O)
		} else {
			result.OutFileJson("")
			result.OutFileCsv("")
			result.OutFileHtml("")
		}
	} else {
		UrlToRedirect()
		result.Print()
	}
}

func AppendJs(ur string, urltjs string) int {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	if len(result.ResultUrl)+len(result.ResultJs) >= cmd.MA {
		return 1
	}
	_, err := url.Parse(ur)
	if err != nil {
		return 2
	}

	// 过滤其他ip ####
	host1, _, _ := url_parse(ur)
	host2, _, _ := url_parse(urltjs)
	if host1 != host2 {
		return 2
	}

	for _, eachItem := range result.ResultJs {
		if eachItem.Url == ur {
			return 0
		}
	}

	result.ResultJs = append(result.ResultJs, mode.Link{Url: ur})
	if strings.HasSuffix(urltjs, ".js") {
		result.Jsinurl[ur] = result.Jsinurl[urltjs]
	} else {
		re := regexp.MustCompile("[a-zA-z]+://[^\\s]*/|[a-zA-z]+://[^\\s]*")
		u := re.FindAllStringSubmatch(urltjs, -1)
		result.Jsinurl[ur] = u[0][0]
	}
	result.Jstourl[ur] = urltjs
	return 0

}

func AppendUrl(ur string, urlturl string) int {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	if len(result.ResultUrl)+len(result.ResultJs) >= cmd.MA {
		return 1
	}
	_, err := url.Parse(ur)
	if err != nil {
		return 2
	}

	// 过滤其他ip ####
	host1, _, _ := url_parse(ur)
	host2, _, _ := url_parse(urlturl)
	if host1 != host2 {
		return 2
	}

	for _, eachItem := range result.ResultUrl {
		if eachItem.Url == ur {
			return 0
		}
	}
	url.Parse(ur)
	result.ResultUrl = append(result.ResultUrl, mode.Link{Url: ur})
	result.Urltourl[ur] = urlturl
	return 0
}

func AppendInfo(info mode.Info) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	result.Infos = append(result.Infos, info)
}

func AppendEndUrl(url string) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	for _, eachItem := range result.EndUrl {
		if eachItem == url {
			return
		}
	}
	result.EndUrl = append(result.EndUrl, url)

}

func GetEndUrl(url string) bool {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	for _, eachItem := range result.EndUrl {
		if eachItem == url {
			return true
		}
	}
	return false

}

func AddRedirect(url string) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	result.Redirect[url] = true
}

func AddSource() {
	for i := range result.ResultJs {
		result.ResultJs[i].Source = result.Jstourl[result.ResultJs[i].Url]
	}
	for i := range result.ResultUrl {
		result.ResultUrl[i].Source = result.Urltourl[result.ResultUrl[i].Url]
	}

}

func UrlToRedirect() {
	for i := range result.ResultJs {
		if result.ResultJs[i].Status == "302" {
			result.ResultJs[i].Url = result.ResultJs[i].Url + " -> " + result.ResultJs[i].Redirect
		}
	}
	for i := range result.ResultUrl {
		if result.ResultUrl[i].Status == "302" {
			result.ResultUrl[i].Url = result.ResultUrl[i].Url + " -> " + result.ResultUrl[i].Redirect
		}
	}

}

func Initialization() {
	result.ResultJs = []mode.Link{}
	result.ResultUrl = []mode.Link{}
	result.Fuzzs = []mode.Link{}
	result.Infos = []mode.Info{}
	result.EndUrl = []string{}
	result.Domains = []string{}
	result.Jsinurl = make(map[string]string)
	result.Jstourl = make(map[string]string)
	result.Urltourl = make(map[string]string)
	result.Redirect = make(map[string]bool)
}
