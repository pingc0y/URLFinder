package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	lock  sync.Mutex
	wg    sync.WaitGroup
	mux   sync.Mutex
	ch    = make(chan int, t)
	jsch  = make(chan int, t/2)
	urlch = make(chan int, t/2)
)

var progress int = 1
var (
	resultJs  [][]string
	resultUrl [][]string
	endUrl    []string
	jsinurl   map[string]string
	jstourl   map[string]string
	urltourl  map[string]string
)

var (
	Green  = color.Style{color.LightYellow}.Render
	Blue   = color.Style{color.LightBlue}.Render
	Red    = color.Style{color.LightRed}.Render
	Yellow = color.Style{color.LightYellow}.Render
)
var (
	h bool
	I bool
	m int
	s string
	u string
	d string
	c string
	a string
	f string
	o string
	x string
	t int = 50
	z int
)

func init() {
	flag.StringVar(&a, "a", "", "set user-agent\n设置user-agent请求头")
	flag.StringVar(&c, "c", "", "set cookie\n设置cookie")
	flag.StringVar(&d, "d", "", "set domainName\n指定获取的域名")
	flag.StringVar(&f, "f", "", "set urlFile\n批量抓取url,指定文件路径")
	flag.BoolVar(&h, "h", false, "this help\n帮助信息（可以看到当前版本更新日期）")
	flag.BoolVar(&I, "i", false, "set configFile\n加载yaml配置文件（不存在时，会在当前目录创建一个默认yaml配置文件）")
	flag.IntVar(&m, "m", 1, "set mode\n抓取模式 \n   1 normal\n     正常抓取（默认） \n   2 thorough\n     深入抓取 （url只深入一层，防止抓偏） \n   3 security\n     安全深入抓取（过滤delete，remove等敏感路由） \n   ")
	flag.StringVar(&o, "o", "", "set outFile\n结果导出到csv文件，需指定导出文件目录（.代表当前目录）")
	flag.StringVar(&s, "s", "", "set status\n显示指定状态码，all为显示全部（多个状态码用,隔开）")
	flag.IntVar(&t, "t", 50, "set thread\n设置线程数（默认50）\n")
	flag.StringVar(&u, "u", "", "set url\n目标URL")
	flag.StringVar(&x, "x", "", "set httpProxy\n设置http代理,格式: http://127.0.0.1:8877|username:password （无需身份验证就不写后半部分）")
	flag.IntVar(&z, "z", 0, "set Fuzz\n对404链接进行fuzz(只对主域名下的链接生效,需要与-s一起使用） \n   1 decreasing\n     目录递减fuzz \n   2 2combination\n     2级目录组合fuzz \n   3 3combination\n     3级目录组合fuzz（适合少量链接使用） \n")

	// 改变默认的 Usage
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `URLFinder 2022/10/25  by pingc
Usage: URLFinder [-a user-agent] [-c cookie] [-d domainName] [-f urlFile]  [-h help]  [-i configFile]  [-m mode] [-o outFile]  [-s status] [-t thread] [-u url] [-x httpProxy] [-z fuzz]

Options:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
		return
	}
	if u == "" && f == "" {
		flag.Usage()
		return
	}
	color.LightCyan.Println("         __   __   ___ _           _           \n /\\ /\\  /__\\ / /  / __(_)_ __   __| | ___ _ __ \n/ / \\ \\/ \\/// /  / _\\ | | '_ \\ / _` |/ _ \\ '__|\n\\ \\_/ / _  \\ /___ /   | | | | | (_| |  __/ |   \n \\___/\\/ \\_\\____\\/    |_|_| |_|\\__,_|\\___|_|   \n                                               ")
	if a != "" {
		ua = a
	}
	if o != "" {
		if !IsDir(o) {
			return
		}
	}
	if I {
		GetConfig("config.yaml")
	}
	if t != 50 {
		ch = make(chan int, t+1)
		jsch = make(chan int, t/2+1)
		urlch = make(chan int, t/2+1)
	}
	if f != "" {
		// 创建句柄
		fi, err := os.Open(f)
		if err != nil {
			panic(err)
		}
		r := bufio.NewReader(fi) // 创建 Reader
		for {
			resultJs = nil
			resultUrl = nil
			endUrl = nil

			lineBytes, err := r.ReadBytes('\n')
			//去掉字符串首尾空白字符，返回字符串
			line := strings.TrimSpace(string(lineBytes))
			u = line
			start(u)

			if err == io.EOF {
				break
			}
			fmt.Println("----------------------------------------")

		}
		return
	}
	start(u)
}

func start(u string) {

	jsinurl = make(map[string]string)
	jstourl = make(map[string]string)
	urltourl = make(map[string]string)
	fmt.Println("Start Spider URL: " + color.LightBlue.Sprintf(u))
	wg.Add(1)
	ch <- 1
	go spider(u, true)
	wg.Wait()
	progress = 1
	fmt.Printf("\rSpider OK      \n")

	resultUrl = RemoveRepeatElement(resultUrl)
	resultJs = RemoveRepeatElement(resultJs)
	if s != "" {
		fmt.Printf("Start %d Validate...\n", len(resultUrl)+len(resultJs))
		fmt.Printf("\r                                           ")
		//验证JS状态
		for i, s := range resultJs {
			wg.Add(1)
			jsch <- 1
			go jsState(s[0], i)
		}
		//验证URL状态
		for i, s := range resultUrl {
			wg.Add(1)
			urlch <- 1
			go urlState(s[0], i)
		}
		wg.Wait()
		fmt.Printf("\r                                           ")
		fmt.Printf("\rValidate OK  \n")

		if z != 0 {
			fuzz()
		}
	}

	//打印还是输出
	if len(o) > 0 {
		outFile()
	} else {
		print()
	}
}

func appendJs(url string, urltjs string) {
	lock.Lock()
	defer lock.Unlock()
	url = strings.Replace(url, "/./", "/", -1)
	for _, eachItem := range resultJs {
		if eachItem[0] == url {
			return
		}
	}
	resultJs = append(resultJs, []string{url})
	if strings.HasSuffix(urltjs, ".js") {
		jsinurl[url] = jsinurl[urltjs]
	} else {
		re := regexp.MustCompile("[a-zA-z]+://[^\\s]*/|[a-zA-z]+://[^\\s]*")
		u := re.FindAllStringSubmatch(urltjs, -1)
		jsinurl[url] = u[0][0]
	}
	if o != "" {
		jstourl[url] = urltjs
	}

}

func appendUrl(url string, urlturl string) {
	lock.Lock()
	defer lock.Unlock()
	url = strings.Replace(url, "/./", "/", -1)
	for _, eachItem := range resultUrl {
		if eachItem[0] == url {
			return
		}
	}
	resultUrl = append(resultUrl, []string{url})
	if o != "" {
		urltourl[url] = urlturl
	}

}

func appendEndUrl(url string) {
	lock.Lock()
	defer lock.Unlock()
	for _, eachItem := range endUrl {
		if eachItem == url {
			return
		}
	}
	endUrl = append(endUrl, url)

}

func getEndUrl(url string) bool {
	lock.Lock()
	defer lock.Unlock()
	for _, eachItem := range endUrl {
		if eachItem == url {
			return true
		}
	}
	return false

}
