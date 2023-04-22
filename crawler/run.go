package crawler

import (
	"bufio"
	"fmt"
	"github.com/gookit/color"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/crawler/fuzz"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

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
		fmt.Printf("\r                                           ")
		fuzz.JsFuzz()
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
			fuzz.UrlFuzz()
			time.Sleep(1 * time.Second)
		}
	}
	AddSource()

}

func Res() {
	//打印还是输出
	if len(cmd.O) > 0 {
		result.OutFileJson()
		result.OutFileCsv()
		result.OutFileHtml()
	} else {
		result.Print()
	}
}

func Run() {

	if cmd.O != "" {
		if !util.IsDir(cmd.O) {
			return
		}
	}
	if cmd.I {
		config.GetConfig("config.yaml")
	}
	if cmd.H || (cmd.U == "" && cmd.F == "" && cmd.FF == "") {
		fmt.Println("请至少使用 -u 或 -f 指定一个url")
		os.Exit(0)
	}
	if cmd.U != "" && !regexp.MustCompile("https{0,1}://").MatchString(cmd.U) {
		fmt.Println("url格式错误,请填写正确url")
		os.Exit(0)
	}

	if cmd.T != 50 {
		config.Ch = make(chan int, cmd.T+1)
		config.Jsch = make(chan int, cmd.T/2+1)
		config.Urlch = make(chan int, cmd.T/2+1)
	}
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

func AppendJs(url string, urltjs string) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	url = strings.Replace(url, "/./", "/", -1)
	for _, eachItem := range result.ResultJs {
		if eachItem.Url == url {
			return
		}
	}
	result.ResultJs = append(result.ResultJs, mode.Link{Url: url})
	if strings.HasSuffix(urltjs, ".js") {
		result.Jsinurl[url] = result.Jsinurl[urltjs]
	} else {
		re := regexp.MustCompile("[a-zA-z]+://[^\\s]*/|[a-zA-z]+://[^\\s]*")
		u := re.FindAllStringSubmatch(urltjs, -1)
		result.Jsinurl[url] = u[0][0]
	}
	result.Jstourl[url] = urltjs

}

func AppendUrl(url string, urlturl string) {
	config.Lock.Lock()
	defer config.Lock.Unlock()
	url = strings.Replace(url, "/./", "/", -1)
	for _, eachItem := range result.ResultUrl {
		if eachItem.Url == url {
			return
		}
	}
	result.ResultUrl = append(result.ResultUrl, mode.Link{Url: url})
	result.Urltourl[url] = urlturl
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

func AddSource() {
	for i := range result.ResultJs {
		result.ResultJs[i].Source = result.Jstourl[result.ResultJs[i].Url]
	}
	for i := range result.ResultUrl {
		result.ResultUrl[i].Source = result.Urltourl[result.ResultUrl[i].Url]
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
}
