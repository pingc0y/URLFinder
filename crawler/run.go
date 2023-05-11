package crawler

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var client *http.Client

func load() {
	if cmd.O != "" {
		if !util.IsDir(cmd.O) {
			return
		}
	}
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
				line := strings.TrimSpace(string(lineBytes))
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
				line := strings.TrimSpace(string(lineBytes))
				if cmd.U == "" {
					cmd.U = line
				}
				start(line)
				fmt.Println("----------------------------------------")
			}

			if err == io.EOF {
				break
			}
		}
		Res()
		return
	}
	Initialization()
	start(cmd.U)
	Res()
}

func start(u string) {
	fmt.Println("Target URL: " + color.LightBlue.Sprintf(u))
	config.Wg.Add(1)
	config.Ch <- 1
	go Spider(u, 1)
	config.Wg.Wait()
	config.Progress = 1
	fmt.Printf("\r\nSpider OK \n")
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
			go JsState(s.Url, i, result.ResultJs[i].Source)
		}
		//验证URL状态
		for i, s := range result.ResultUrl {
			config.Wg.Add(1)
			config.Urlch <- 1
			go UrlState(s.Url, i)
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
		fmt.Println("未获取到数据")
		return
	}
	//打印还是输出
	if len(cmd.O) > 0 {
		result.OutFileJson()
		result.OutFileCsv()
		result.OutFileHtml()
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
