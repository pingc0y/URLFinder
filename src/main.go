package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	lock  sync.Mutex
	wg    sync.WaitGroup
	mux   sync.Mutex
	ch    = make(chan int, t)
	jsch  = make(chan int, t/2)
	urlch = make(chan int, t/2)
)

type Link struct {
	Url    string
	Status string
	Size   string
	Title  string
	Source string
}

type Info struct {
	Phone  []string
	Email  []string
	IDcard []string
	JWT    []string
	Source string
}

var progress = 1
var (
	resultJs  []Link
	resultUrl []Link
	infos     []Info
	endUrl    []string
	jsinurl   map[string]string
	jstourl   map[string]string
	urltourl  map[string]string
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
	b string
	f string
	o string
	x string
	t = 50
	z int
)

func init() {
	flag.StringVar(&a, "a", "", "set user-agent\n设置user-agent请求头")
	flag.StringVar(&b, "b", "", "set baseurl\n设置baseurl路径")
	flag.StringVar(&c, "c", "", "set cookie\n设置cookie")
	flag.StringVar(&d, "d", "", "set domainName\n指定获取的域名")
	flag.StringVar(&f, "f", "", "set urlFile\n批量抓取url,指定文件路径")
	flag.BoolVar(&h, "h", false, "this help\n帮助信息")
	flag.BoolVar(&I, "i", false, "set configFile\n加载yaml配置文件（不存在时，会在当前目录创建一个默认yaml配置文件）")
	flag.IntVar(&m, "m", 1, "set mode\n抓取模式 \n   1 normal\n     正常抓取（默认） \n   2 thorough\n     深入抓取 （url只深入一层，防止抓偏） \n   3 security\n     安全深入抓取（过滤delete，remove等敏感路由） \n   ")
	flag.StringVar(&o, "o", "", "set outFile\n结果导出到csv文件，需指定导出文件目录（.代表当前目录）")
	flag.StringVar(&s, "s", "", "set Status\n显示指定状态码，all为显示全部（多个状态码用,隔开）")
	flag.IntVar(&t, "t", 50, "set thread\n设置线程数（默认50）\n")
	flag.StringVar(&u, "u", "", "set Url\n目标URL")
	flag.StringVar(&x, "x", "", "set httpProxy\n设置代理,格式: http://username:password@127.0.0.1:8809")
	flag.IntVar(&z, "z", 0, "set Fuzz\n对404链接进行fuzz(只对主域名下的链接生效,需要与-s一起使用） \n   1 decreasing\n     目录递减fuzz \n   2 2combination\n     2级目录组合fuzz（适合少量链接使用） \n   3 3combination\n     3级目录组合fuzz（适合少量链接使用） \n")

	// 改变默认的 Usage
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `URLFinder 2023/1/29  by pingc0y
Usage: URLFinder [-a user-agent] [-b baseurl] [-c cookie] [-d domainName] [-f urlFile]  [-h help]  [-i configFile]  [-m mode] [-o outFile]  [-s Status] [-t thread] [-u Url] [-x httpProxy] [-z fuzz]

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
	color.LightCyan.Println("         __   __   ___ _           _           \n /\\ /\\  /__\\ / /  / __(_)_ __   __| | ___ _ __ \n/ / \\ \\/ \\/// /  / _\\ | | '_ \\ / _` |/ _ \\ '__|\n\\ \\_/ / _  \\ /___ /   | | | | | (_| |  __/ |   \n \\___/\\/ \\_\\____\\/    |_|_| |_|\\__,_|\\___|_|     \n\nBy: pingc0y\nUpdateTime: 2023/1/29\nGithub: https://github.com/pingc0y/URLFinder \n")
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
	infos = []Info{}
	fmt.Println("Start Spider URL: " + color.LightBlue.Sprintf(u))
	wg.Add(1)
	ch <- 1
	go spider(u, 1)
	wg.Wait()
	progress = 1
	fmt.Printf("\rSpider OK      \n")
	resultUrl = RemoveRepeatElement(resultUrl)
	resultJs = RemoveRepeatElement(resultJs)
	if s != "" {
		fmt.Printf("Start %d Validate...\n", len(resultUrl)+len(resultJs))
		fmt.Printf("\r                                           ")
		jsFuzz()
		//验证JS状态
		for i, s := range resultJs {
			wg.Add(1)
			jsch <- 1
			go jsState(s.Url, i, resultJs[i].Source)
		}
		//验证URL状态
		for i, s := range resultUrl {
			wg.Add(1)
			urlch <- 1
			go urlState(s.Url, i)
		}
		wg.Wait()

		time.Sleep(1 * time.Second)
		fmt.Printf("\r                                           ")
		fmt.Printf("\rValidate OK  \n")

		if z != 0 {
			fuzz()
			time.Sleep(1 * time.Second)
		}
	}

	//打印还是输出
	if len(o) > 0 {
		outFileJson()
		outFileCsv()
		outFileHtml()

	} else {
		print()
	}
}
