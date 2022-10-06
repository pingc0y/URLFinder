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
	lock sync.Mutex
	wg   sync.WaitGroup
	mux  sync.Mutex
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
)

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.BoolVar(&I, "i", false, "set configFile")
	flag.StringVar(&u, "u", "", "set url")
	flag.StringVar(&d, "d", "", "set domainName")
	flag.StringVar(&c, "c", "", "set cookie")
	flag.StringVar(&f, "f", "", "set urlFile")
	flag.StringVar(&o, "o", "", "set outFile")
	flag.StringVar(&a, "a", "", "set user-agent")
	flag.StringVar(&s, "s", "", "set status")
	flag.StringVar(&x, "x", "", "set httpProxy")
	flag.IntVar(&m, "m", 1, "set mode \n   1  normal \n   2  thorough \n   3  security \n")

	// 改变默认的 Usage
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `URLFinder 2022/10/6  by pingc
Usage: URLFinder [-h help] [-u url] [-d domainName] [-c cookie]  [-a user-agent]  [-m mode]  [-f urlFile]  [-o outFile] [-s status] [-i configFile] [-x httpProxy]

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
	wg.Add(1)
	jsinurl = make(map[string]string)
	jstourl = make(map[string]string)
	urltourl = make(map[string]string)
	fmt.Println("Start Spider URL: " + color.LightBlue.Sprintf(u))

	go spider(u, true)
	wg.Wait()
	progress = 1
	fmt.Println("\rSpider OK      ")

	resultUrl = RemoveRepeatElement(resultUrl)
	resultJs = RemoveRepeatElement(resultJs)

	if s != "" {
		fmt.Println("Start Validate...")
	}
	//验证JS状态
	for i, s := range resultJs {
		wg.Add(1)
		go jsState(s[0], i)
	}
	//验证URL状态
	for i, s := range resultUrl {
		wg.Add(1)
		go urlState(s[0], i)
	}
	wg.Wait()
	fmt.Println("\rValidate OK     ")

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
