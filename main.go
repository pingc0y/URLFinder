package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gookit/color"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	lock sync.Mutex
	wg   sync.WaitGroup
	mux  sync.Mutex
)

type config struct {
	Headers map[string]string `yaml:"headers"`
}

var ua = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

var conf config
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
)

var (
	risks = []string{"remove", "delete", "insert", "update", "logout"}
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
	flag.IntVar(&m, "m", 1, "set mode \n   1  normal \n   2  thorough \n   3  security \n")

	// 改变默认的 Usage
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `URLFinder 2022/9/23  by pingc
Usage: URLFinder [-h help] [-u url] [-d domainName] [-c cookie]  [-a user-agent]  [-m mode]  [-f urlFile]  [-o outFile] [-s status] [-i configFile]

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
	fmt.Println("         __   __   ___ _           _           \n /\\ /\\  /__\\ / /  / __(_)_ __   __| | ___ _ __ \n/ / \\ \\/ \\/// /  / _\\ | | '_ \\ / _` |/ _ \\ '__|\n\\ \\_/ / _  \\ /___ /   | | | | | (_| |  __/ |   \n \\___/\\/ \\_\\____\\/    |_|_| |_|\\__,_|\\___|_|   \n                                               ")
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

// 导出
func outFile() {
	//获取域名
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(u, 1)
	if len(hosts) == 0 {
		host = u
	} else {
		host = hosts[0]
	}

	//抓取的域名优先排序
	if s != "" {
		resultUrl = SelectSort(resultUrl)
		resultJs = SelectSort(resultJs)
	}
	resultJsHost, resultJsOther := urlDispose(resultJs, host, getHost(u))
	resultUrlHost, resultUrlOther := urlDispose(resultUrl, host, getHost(u))
	//输出到文件
	if strings.Contains(host, ":") {
		host = strings.Replace(host, ":", "：", -1)
	}
	file, err := os.OpenFile(o+"/"+host+".csv", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	file.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM，防止中文乱码
	// 写数据到文件
	if err != nil {
		fmt.Println("open file error:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if s == "" {
		writer.WriteString("url,source\n")
	} else {
		writer.WriteString("url,status,size,title,source\n")
	}

	if d == "" {
		writer.WriteString(strconv.Itoa(len(resultJsHost)) + " JS to " + getHost(u) + "\n")
	} else {
		writer.WriteString(strconv.Itoa(len(resultJsHost)+len(resultJsOther)) + " JS to " + d + "\n")
	}

	for _, j := range resultJsHost {
		var str = ""
		if len(j) == 3 {
			if strings.HasPrefix(j[1], "2") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", j[0], j[1], j[2], jstourl[j[0]])
			} else if strings.HasPrefix(j[1], "3") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", j[0], j[1], j[2], jstourl[j[0]])
			} else {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", j[0], j[1], j[2], jstourl[j[0]])
			}
		} else if len(j) == 2 {
			str = fmt.Sprintf("\"%s\",\"%s\",\"0\",,\"%s\"", j[0], j[1], jstourl[j[0]])
		} else if s == "" {
			str = fmt.Sprintf("\"%s\",\"%s\"", j[0], jstourl[j[0]])
		}
		writer.WriteString(str + "\n")
	}
	if d == "" {
		writer.WriteString("\n" + strconv.Itoa(len(resultJsOther)) + " JS to other\n")
	}
	for _, j := range resultJsOther {
		var str = ""
		if len(j) == 3 {
			if strings.HasPrefix(j[1], "2") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", j[0], j[1], j[2], jstourl[j[0]])
			} else if strings.HasPrefix(j[1], "3") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", j[0], j[1], j[2], jstourl[j[0]])
			} else {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", j[0], j[1], j[2], jstourl[j[0]])
			}
		} else if len(j) == 2 {
			str = fmt.Sprintf("\"%s\",\"%s\",\"0\",,\"%s\"", j[0], j[1], jstourl[j[0]])
		} else if s == "" {
			str = fmt.Sprintf("\"%s\",\"%s\"", j[0], jstourl[j[0]])
		}
		writer.WriteString(str + "\n")
	}

	writer.WriteString("\n\n")
	if d == "" {
		writer.WriteString(strconv.Itoa(len(resultUrlHost)) + " URL to " + getHost(u) + "\n")
	} else {
		writer.WriteString(strconv.Itoa(len(resultUrlHost)+len(resultUrlOther)) + " URL to " + d + "\n")
	}

	for _, u := range resultUrlHost {
		var str = ""
		if len(u) == 4 {
			if strings.HasPrefix(u[1], "2") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"", u[0], u[1], u[2], u[3], urltourl[u[0]])
			} else if strings.HasPrefix(u[1], "3") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"", u[0], u[1], u[2], u[3], urltourl[u[0]])
			} else {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"", u[0], u[1], u[2], u[3], urltourl[u[0]])
			}
		} else if len(u) == 3 {
			if strings.HasPrefix(u[1], "2") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", u[0], u[1], u[2], urltourl[u[0]])
			} else if strings.HasPrefix(u[1], "3") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", u[0], u[1], u[2], urltourl[u[0]])
			} else {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", u[0], u[1], u[2], urltourl[u[0]])
			}
		} else if len(u) == 2 {
			str = fmt.Sprintf("\"%s\",\"%s\",\"0\",,\"%s\"", u[0], u[1], urltourl[u[0]])
		} else if s == "" {
			str = fmt.Sprintf("\"%s\",\"%s\"", u[0], urltourl[u[0]])
		}
		writer.WriteString(str + "\n")
	}
	if d == "" {
		writer.WriteString("\n" + strconv.Itoa(len(resultUrlOther)) + " URL to other\n")
	}
	for _, u := range resultUrlOther {
		var str = ""
		if len(u) == 4 {
			if strings.HasPrefix(u[1], "2") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"", u[0], u[1], u[2], u[3], urltourl[u[0]])
			} else if strings.HasPrefix(u[1], "3") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"", u[0], u[1], u[2], u[3], urltourl[u[0]])
			} else {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"", u[0], u[1], u[2], u[3], urltourl[u[0]])
			}
		} else if len(u) == 3 {
			if strings.HasPrefix(u[1], "2") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", u[0], u[1], u[2], urltourl[u[0]])
			} else if strings.HasPrefix(u[1], "3") {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", u[0], u[1], u[2], urltourl[u[0]])
			} else {
				str = fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"", u[0], u[1], u[2], urltourl[u[0]])
			}
		} else if len(u) == 2 {
			str = fmt.Sprintf("\"%s\",\"%s\",\"0\",,\"%s\"", u[0], u[1], urltourl[u[0]])
		} else if s == "" {
			str = fmt.Sprintf("\"%s\",\"%s\"", u[0], urltourl[u[0]])
		}
		writer.WriteString(str + "\n")

	}

	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中

	fmt.Println(strconv.Itoa(len(resultJsHost)+len(resultJsOther))+"JS + "+strconv.Itoa(len(resultUrlHost)+len(resultUrlOther))+"URL --> ", file.Name())

	return
}

// 打印
func print() {

	//获取域名
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(u, 1)
	if len(hosts) == 0 {
		host = u
	} else {
		host = hosts[0]
	}
	//打印JS
	if s != "" {
		resultJs = SelectSort(resultJs)
	}
	//抓取的域名优先排序

	resultJsHost, resultJsOther := urlDispose(resultJs, host, getHost(u))

	ulen := ""
	if len(resultUrl) != 0 {
		uleni := 0
		for _, s := range resultUrl {
			uleni += len(s[0])
		}
		ulen = strconv.Itoa(uleni/len(resultUrl) + 10)
	}
	jlen := ""
	if len(resultJs) != 0 {
		jleni := 0
		for _, s := range resultJs {
			jleni += len(s[0])
		}
		jlen = strconv.Itoa(jleni/len(resultJs) + 10)
	}
	if d == "" {
		fmt.Println(strconv.Itoa(len(resultJsHost)) + " JS to " + getHost(u))
	} else {
		fmt.Println(strconv.Itoa(len(resultJsHost)+len(resultJsOther)) + " JS to " + d)
	}
	for _, j := range resultJsHost {
		if len(j) == 3 {
			if strings.HasPrefix(j[1], "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightGreen.Sprintf(" [ status: %s, size: %s ]\n", j[1], j[2]))
			} else if strings.HasPrefix(j[1], "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightYellow.Sprintf(" [ status: %s, size: %s ]\n", j[1], j[2]))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightRed.Sprintf(" [ status: %s, size: %s ]\n", j[1], j[2]))
			}
		} else if len(j) == 2 {
			fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightRed.Sprintf(" [ status: %s, size: 0 ]\n", j[1]))
		} else if s == "" {
			fmt.Printf(color.LightBlue.Sprintf(j[0]) + "\n")
		}
	}
	if d == "" {
		fmt.Println("\n" + strconv.Itoa(len(resultJsOther)) + " JS to other")
	}
	for _, j := range resultJsOther {
		if len(j) == 3 {
			if strings.HasPrefix(j[1], "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightGreen.Sprintf(" [ status: %s, size: %s ]\n", j[1], j[2]))
			} else if strings.HasPrefix(j[1], "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightYellow.Sprintf(" [ status: %s, size: %s ]\n", j[1], j[2]))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightRed.Sprintf(" [ status: %s, size: %s ]\n", j[1], j[2]))
			}
		} else if len(j) == 2 {
			fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j[0]) + color.LightRed.Sprintf(" [ status: %s, size: 0 ]\n", j[1]))
		} else if s == "" {
			fmt.Printf(color.LightBlue.Sprintf(j[0]) + "\n")
		}
	}

	//打印URL
	fmt.Println("\n\n")
	if s != "" {
		resultUrl = SelectSort(resultUrl)
	}
	//抓取的域名优先排序
	resultUrlHost, resultUrlOther := urlDispose(resultUrl, host, getHost(u))

	if d == "" {
		fmt.Println(strconv.Itoa(len(resultUrlHost)) + " URL to " + getHost(u))
	} else {
		fmt.Println(strconv.Itoa(len(resultUrlHost)+len(resultUrlOther)) + " URL to " + d)
	}

	for _, u := range resultUrlHost {
		if len(u) == 4 {
			if strings.HasPrefix(u[1], "0") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightGreen.Sprintf(" [ %s ]\n", u[3]))
			} else if strings.HasPrefix(u[1], "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightGreen.Sprintf(" [ status: %s, size: %s, title: %s ]\n", u[1], u[2], u[3]))
			} else if strings.HasPrefix(u[1], "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightYellow.Sprintf(" [ status: %s, size: %s, title: %s ]\n", u[1], u[2], u[3]))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightRed.Sprintf(" [ status: %s, size: %s, title: %s ]\n", u[1], u[2], u[3]))
			}
		} else if len(u) == 3 {
			if strings.HasPrefix(u[1], "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightGreen.Sprintf(" [ status: %s, size: %s ]\n", u[1], u[2]))
			} else if strings.HasPrefix(u[1], "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightYellow.Sprintf(" [ status: %s, size: %s ]\n", u[1], u[2]))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightRed.Sprintf(" [ status: %s, size: %s ]\n", u[1], u[2]))
			}
		} else if len(u) == 2 {
			fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightRed.Sprintf(" [ status: %s, size: 0 ]\n", u[1]))
		} else if s == "" {
			fmt.Printf(color.LightBlue.Sprintf(u[0]) + "\n")
		}
	}
	if d == "" {
		fmt.Println("\n" + strconv.Itoa(len(resultUrlOther)) + " URL to other")
	}
	for _, u := range resultUrlOther {
		if len(u) == 4 {
			if strings.HasPrefix(u[1], "0") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightGreen.Sprintf(" [ %s ]\n", u[3]))
			} else if strings.HasPrefix(u[1], "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightGreen.Sprintf(" [ status: %s, size: %s, title: %s ]\n", u[1], u[2], u[3]))
			} else if strings.HasPrefix(u[1], "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightYellow.Sprintf(" [ status: %s, size: %s, title: %s ]\n", u[1], u[2], u[3]))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightRed.Sprintf(" [ status: %s, size: %s, title: %s ]\n", u[1], u[2], u[3]))
			}
		} else if len(u) == 3 {
			if strings.HasPrefix(u[1], "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightGreen.Sprintf(" [ status: %s, size: %s ]\n", u[1], u[2]))
			} else if strings.HasPrefix(u[1], "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightYellow.Sprintf(" [ status: %s, size: %s ]\n", u[1], u[2]))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightRed.Sprintf(" [ status: %s, size: %s ]\n", u[1], u[2]))
			}
		} else if len(u) == 2 {
			fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u[0]) + color.LightRed.Sprintf(" [ status: %s, size: 0 ]\n", u[1]))
		} else if s == "" {
			fmt.Printf(color.LightBlue.Sprintf(u[0]) + "\n")
		}
	}
}

// 蜘蛛抓取页面内容
func spider(ur string, is bool) {
	fmt.Printf("\rSpider %d", progress)
	mux.Lock()
	progress++
	mux.Unlock()

	//标记完成
	defer wg.Done()
	url, _ := url.QueryUnescape(ur)
	if getEndUrl(url) {
		return
	}
	if m == 3 {
		for _, v := range risks {
			if strings.Contains(url, v) {
				return
			}
		}
	}
	appendEndUrl(url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	//增加header选项
	request.Header.Add("Cookie", c)
	request.Header.Add("User-Agent", ua)
	request.Header.Add("Accept", "*/*")
	//加载yaml配置
	if I {
		request.Header = SetHeadersConfig(request.Header)
	}
	//处理返回结果
	response, err := client.Do(request)
	if err != nil {
		return
	} else {
		defer response.Body.Close()

	}

	//提取url用于拼接其他url或js
	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
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

	//提取js
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
		".(http[^\\s,^',^’,^\",^”,^>,^<,^;,^\\(,^),^\\|,^\\*,^\\[]{2,250}?[^=,^\\*,^\\s,^',^’,^\",^”,^>,^<,^:,^;,^\\*,^\\|,^\\(,^),^\\[]{3}[.]js)",
		"[\",',‘,“]\\s{0,6}(/{0,1}[^\\s,^',^’,^\",^”,^\\|,^>,^<,^:,^;,^\\*,^\\(,^\\),^\\[]{2,250}?[^=,^\\*,^\\s,^',^’,^\\|,^\",^”,^>,^<,^:,^;,^\\*,^\\(,^),^\\[]{3}[.]js)",
		"=\\s{0,6}[\",',’,”]{0,1}\\s{0,6}(/{0,1}[^\\s,^',^’,^\",^”,^\\|,^>,^<,^;,^\\*,^\\(,^),^\\[]{2,250}?[^=,^\\*,^\\s,^',^’,^\",^”,^>,^\\|,^<,^:,^;,^\\*,^\\(,^),^\\[]{3}[.]js)",
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
				if is || m == 2 || m == 3 {
					wg.Add(1)
					go spider(js[0], false)
				}
			} else if strings.HasPrefix(js[0], "//") {
				appendJs(scheme+":"+js[0], source)
				if is || m == 2 || m == 3 {
					wg.Add(1)
					go spider(scheme+":"+js[0], false)
				}

			} else if strings.HasPrefix(js[0], "/") {
				appendJs(host+js[0], source)
				if is || m == 2 || m == 3 {
					wg.Add(1)
					go spider(host+js[0], false)
				}
			} else {
				appendJs(host+cata+js[0], source)
				if is || m == 2 || m == 3 {
					wg.Add(1)
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
		"[\",',‘,“]\\s{0,6}(http[^\\s,^',^’,^\",^”,^>,^<,^),^\\(]{2,250}?)\\s{0,6}[\",',‘,“]",
		"=\\s{0,6}(http[^\\s,^',^’,^\",^”,^>,^<,^),^\\(]{2,250})",
		"[\",',‘,“]\\s{0,6}(#{0,1}/[^\\s,^',^’,^\",^”,^>,^<,^\\:,^),^\\(]{2,250}?)\\s{0,6}[\",',‘,“]",
		"href\\s{0,6}=\\s{0,6}[\",',‘,“]{0,1}\\s{0,6}([^\\s,^',^’,^\",^“,^>,^<,^),^\\(]{2,250})|action\\s{0,6}=\\s{0,6}[\",',‘,“]{0,1}\\s{0,6}([^\\s,^',^’,^\",^“,^>,^<,^),^\\(]{2,250})",
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
				if is && m == 2 || m == 3 {
					wg.Add(1)
					go spider(url[0], false)
				}
			} else if strings.HasPrefix(url[0], "//") {
				appendUrl(scheme+":"+url[0], source)
				if is && m == 2 || m == 3 {
					wg.Add(1)
					go spider(scheme+":"+url[0], false)
				}
			} else if strings.HasPrefix(url[0], "/") {
				appendUrl(host+url[0], source)
				if is && m == 2 || m == 3 {
					wg.Add(1)
					go spider(host+url[0], false)
				}
			} else if !strings.HasSuffix(source, ".js") {
				appendUrl(host+cata+url[0], source)
				if is && m == 2 || m == 3 {
					wg.Add(1)
					go spider(host+cata+url[0], false)
				}
			} else if strings.HasSuffix(source, ".js") {
				appendUrl(jsinurl[host+path]+url[0], source)
				if is && m == 2 || m == 3 {
					wg.Add(1)
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
		fstr := []string{".js?", ".css?", ".jpeg?", ".jpg?", ".png?", ".gif?", "github.com", "www.w3.org", "example.com", "<", ">", "{", "}", "[", "]", "|", "^", ";", "/js/", "location.href", "javascript:void", "\\n"}
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
		zstr := []string{".js", ".css", ",", ".jpeg", ".jpg", ".png", ".gif", ".ico", ".svg"}

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
	defer wg.Done()
	defer printProgress()
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
		request.Header = SetHeadersConfig(request.Header)
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
	}

	code := response.StatusCode
	if strings.Contains(s, strconv.Itoa(code)) || s == "all" {
		var length int
		dataBytes, err := ioutil.ReadAll(response.Body)
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
	defer wg.Done()
	defer printProgress()
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
		request.Header = SetHeadersConfig(request.Header)
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
	}

	code := response.StatusCode
	if strings.Contains(s, strconv.Itoa(code)) || s == "all" {
		var length int
		dataBytes, err := ioutil.ReadAll(response.Body)
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

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return s.IsDir()
}

// 对结果进行状态码排序
func SelectSort(arr [][]string) [][]string {
	length := len(arr)
	var sort []int
	for _, v := range arr {
		if v[0] == "" || len(v) == 1 || v[1] == "timeout" {
			sort = append(sort, 999)
		} else {
			in, _ := strconv.Atoi(v[1])
			sort = append(sort, in)
		}
	}
	if length <= 1 {
		return arr
	} else {
		for i := 0; i < length-1; i++ { //只剩一个元素不需要索引
			min := i                          //标记索引
			for j := i + 1; j < length; j++ { //每次选出一个极小值
				if sort[min] > sort[j] {
					min = j //保存极小值的索引
				}
			}
			if i != min {
				sort[i], sort[min] = sort[min], sort[i] //数据交换
				arr[i], arr[min] = arr[min], arr[i]     //数据交换
			}
		}
		return arr
	}
}

// 对结果进行URL排序
func urlDispose(arr [][]string, url, host string) ([][]string, [][]string) {
	var urls [][]string
	var urlts [][]string
	var other [][]string
	for _, v := range arr {
		if strings.Contains(v[0], url) {
			urls = append(urls, v)
		} else {
			if strings.Contains(v[0], host) {
				urlts = append(urlts, v)
			} else {
				other = append(other, v)
			}
		}
	}

	for _, v := range urlts {
		urls = append(urls, v)
	}

	return RemoveRepeatElement(urls), RemoveRepeatElement(other)
}

// 提取顶级域名
func getHost(u string) string {

	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	var host string
	hosts := re.FindAllString(u, 1)
	if len(hosts) == 0 {
		host = u
	} else {
		host = hosts[0]
	}
	re2 := regexp.MustCompile("[^.]*?\\.[^.,^:]*")
	host2 := re2.FindAllString(host, -1)
	re3 := regexp.MustCompile("(([01]?[0-9]{1,3}|2[0-4][0-9]|25[0-5])\\.){3}([01]?[0-9]{1,3}|2[0-4][0-9]|25[0-5])")
	hostIp := re3.FindAllString(u, -1)
	if len(hostIp) == 0 {
		if len(host2) == 1 {
			host = host2[0]
		} else {
			re3 := regexp.MustCompile("\\.[^.]*?\\.[^.,^:]*")
			var ho string
			hos := re3.FindAllString(host, -1)

			if len(hos) == 0 {
				ho = u
			} else {
				ho = hos[len(hos)-1]
			}
			host = strings.Replace(ho, ".", "", 1)
		}
	} else {
		return hostIp[0]
	}
	return host
}

// 去重+去除错误url
func RemoveRepeatElement(list [][]string) [][]string {
	// 创建一个临时map用来存储数组元素
	temp := make(map[string]bool)
	var list2 [][]string
	index := 0
	for _, v := range list {

		//处理-d参数
		if d != "" {
			v[0] = domainNameFilter(v[0])
		}

		if len(v[0]) > 10 {
			re := regexp.MustCompile("://([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
			hosts := re.FindAllString(v[0], 1)
			if len(hosts) != 0 {
				// 遍历数组元素，判断此元素是否已经存在map中
				_, ok := temp[v[0]]
				if !ok {
					list2 = append(list2, v)
					temp[v[0]] = true
				}
			}
		}
		index++

	}
	return list2
}

// 读取配置文件
func GetConfig(path string) {
	con := &config{}
	if f, err := os.Open(path); err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {

			con.Headers = map[string]string{"Cookie": c, "User-Agent": ua, "Accept": "*/*"}
			data, err2 := yaml.Marshal(con)
			err2 = ioutil.WriteFile(path, data, 0644)
			if err2 != nil {
				fmt.Println(err)
			} else {
				fmt.Println("未找到配置文件,已在当面目录下创建配置文件: config.yaml")
			}
		} else {
			fmt.Println("配置文件错误,请尝试重新生成配置文件")
			fmt.Println(err)
		}
		os.Exit(1)
	} else {
		yaml.NewDecoder(f).Decode(con)
		conf = *con
	}
}

// 处理配置
func SetHeadersConfig(header http.Header) http.Header {
	for k, v := range conf.Headers {
		header.Add(k, v)
	}
	return header
}

// 处理-d
func domainNameFilter(url string) string {
	re := regexp.MustCompile("://([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(url, 1)
	if len(hosts) != 0 {
		if !strings.Contains(hosts[0], d) {
			url = ""
		}
	}
	return url
}

// 打印进度
func printProgress() {
	num := len(resultJs) + len(resultUrl)
	fmt.Printf("\rValidate %.0f%%", float64(progress+1)/float64(num+1)*100)
	mux.Lock()
	progress++
	mux.Unlock()

}
