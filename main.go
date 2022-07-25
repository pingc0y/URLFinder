package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/fatih/color"
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
)

var (
	resultJs  []string
	resultUrl []string
	endUrl    []string
)

var (
	h bool
	o bool
	s bool
	u string
	c string
	a string
	f string
	m int
)
var ua = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

func init() {
	flag.BoolVar(&h, "h", false, "this help")
	flag.BoolVar(&o, "o", false, "outFile")
	flag.BoolVar(&s, "s", false, "set status")

	flag.StringVar(&u, "u", "", "set url")
	flag.StringVar(&c, "c", "", "set cookie")
	flag.StringVar(&f, "f", "", "set urlFile")
	flag.StringVar(&a, "a", "", "set user-agent")
	flag.IntVar(&m, "m", 1, "set mode \n 1  normal \n 2  thorough \n")

	// 改变默认的 Usage
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `URLFinder 2022/7/25  by pingc
Usage: URLFinder [-h help] [-u url]  [-c cookie]  [-a user-agent]  [-m mode]  [-f urlFile]  [-o outFile] [-s status]

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
	if a != "" {
		ua = a
	}

	if f != "" {
		// 创建句柄
		fi, err := os.Open(f)
		if err != nil {
			panic(err)
		}
		//func NewReader(rd io.Reader) *Reader {}，返回的是bufio.Reader结构体
		r := bufio.NewReader(fi) // 创建 Reader
		for {
			resultJs = nil
			resultUrl = nil
			endUrl = nil
			//func (b *Reader) ReadBytes(delim byte) ([]byte, error) {}
			lineBytes, err := r.ReadBytes('\n')
			//去掉字符串首尾空白字符，返回字符串
			line := strings.TrimSpace(string(lineBytes))
			u = line
			wg.Add(1)
			go spider(u, true)
			fmt.Println("Start Spider URL: " + u)
			wg.Wait()
			fmt.Println("Spider OK")

			if s {
				fmt.Println("Start Validate...")
			}

			resultUrl = RemoveRepeatElement(resultUrl)
			resultJs = RemoveRepeatElement(resultJs)

			//验证JS状态
			for i, s := range resultJs {
				wg.Add(1)
				go jsState(s, i)
			}
			//验证URL状态
			for i, s := range resultUrl {
				wg.Add(1)
				go urlState(s, i)
			}
			wg.Wait()

			//打印还是输出
			if o {
				outFile()
			} else {
				print()
			}

			if err == io.EOF {
				break
			}

		}
		return
	}

	wg.Add(1)
	go spider(u, true)
	fmt.Println("         __   __   ___ _           _           \n /\\ /\\  /__\\ / /  / __(_)_ __   __| | ___ _ __ \n/ / \\ \\/ \\/// /  / _\\ | | '_ \\ / _` |/ _ \\ '__|\n\\ \\_/ / _  \\ /___ /   | | | | | (_| |  __/ |   \n \\___/\\/ \\_\\____\\/    |_|_| |_|\\__,_|\\___|_|   \n                                               ")

	fmt.Println("Start Spider URL: " + u)
	wg.Wait()
	fmt.Println("Spider OK")

	if s {
		fmt.Println("Start Validate...")
	}

	resultUrl = RemoveRepeatElement(resultUrl)
	resultJs = RemoveRepeatElement(resultJs)

	//验证JS状态
	for i, s := range resultJs {
		wg.Add(1)
		go jsState(s, i)
	}
	//验证URL状态
	for i, s := range resultUrl {
		wg.Add(1)
		go urlState(s, i)
	}
	wg.Wait()

	//打印还是输出
	if o {
		outFile()
	} else {
		print()
	}

}

//输出
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
	//对IP做兼容
	re2 := regexp.MustCompile("(([01]?[0-9]{1,2}|2[0-4][0-9]|25[0-5])\\.){3}([01]?[0-9]{1,2}|2[0-4][0-9]|25[0-5])")
	hostIp := re2.FindAllString(u, 1)
	if len(hostIp) > 0 {
		host = hostIp[0]
	}
	CreateDir("out")
	//输出到文件

	file, err := os.OpenFile("out\\"+host+".txt", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("open file error:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	resultJs = SelectSort(resultJs)
	//抓取的域名优先排序
	resultJsHost, resultJsOther := urlDispose(resultJs, host, getHost(u))
	writer.WriteString(strconv.Itoa(len(resultJsHost)) + " JS to " + getHost(u) + "\n")
	for _, j := range resultJsHost {
		if strings.Contains(j, "  |  ") || !s {
			writer.WriteString(j + "\n")
		}
	}
	writer.WriteString("\n" + strconv.Itoa(len(resultJsOther)) + " JS to other\n")
	for _, j := range resultJsOther {
		if strings.Contains(j, "  |  ") || !s {
			writer.WriteString(j + "\n")
		}
	}

	writer.WriteString("\n\n")
	//抓取的域名优先排序
	resultUrl = SelectSort(resultUrl)
	resultUrlHost, resultUrlOther := urlDispose(resultUrl, host, getHost(u))
	writer.WriteString(strconv.Itoa(len(resultUrlHost)) + " URL to " + getHost(u) + "\n")
	for _, u := range resultUrlHost {
		if strings.Contains(u, "  |  ") || !s {
			writer.WriteString(u + "\n")
		}
	}
	writer.WriteString("\n" + strconv.Itoa(len(resultUrlOther)) + " URL to other\n")
	for _, u := range resultUrlOther {
		if strings.Contains(u, "  |  ") || !s {
			writer.WriteString(u + "\n")
		}
	}

	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中
	fmt.Println(strconv.Itoa(len(resultJs))+"JS + "+strconv.Itoa(len(resultUrl))+"URL --> ", "out\\"+host+".txt")

	return
}

//打印结果
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
	resultJs = SelectSort(resultJs)
	//抓取的域名优先排序
	resultJsHost, resultJsOther := urlDispose(resultJs, host, getHost(u))
	fmt.Println(strconv.Itoa(len(resultJsHost)) + " JS to " + getHost(u))
	for _, u := range resultJsHost {
		if strings.Contains(u, "  |  ") || !s {
			split := strings.Split(u, "  |  ")
			if len(split) == 3 {
				if strings.HasPrefix(split[2], "2") {
					color.Green(u)
				} else if strings.HasPrefix(split[2], "3") {
					color.Yellow(u)
				} else {
					color.Red(u)
				}
			} else if len(split) == 2 {
				color.Red(u)
			} else if !s {
				fmt.Println(u)
			}
		}
	}
	fmt.Println("\n" + strconv.Itoa(len(resultJsOther)) + " JS to other")
	for _, u := range resultJsOther {
		if strings.Contains(u, "  |  ") || !s {
			split := strings.Split(u, "  |  ")
			if len(split) == 3 {
				if strings.HasPrefix(split[2], "2") {
					color.Green(u)
				} else if strings.HasPrefix(split[2], "3") {
					color.Yellow(u)
				} else {
					color.Red(u)
				}
			} else if len(split) == 2 {
				color.Red(u)
			} else if !s {
				fmt.Println(u)
			}
		}
	}

	//打印URL
	fmt.Println("\n\n")
	resultUrl = SelectSort(resultUrl)
	//抓取的域名优先排序
	resultUrlHost, resultUrlOther := urlDispose(resultUrl, host, getHost(u))
	fmt.Println(strconv.Itoa(len(resultUrlHost)) + " URL to " + getHost(u))
	for _, u := range resultUrlHost {
		if strings.Contains(u, "  |  ") || !s {
			split := strings.Split(u, "  |  ")
			if len(split) == 3 {
				if strings.HasPrefix(split[2], "2") {
					color.Green(u)
				} else if strings.HasPrefix(split[2], "3") {
					color.Yellow(u)
				} else {
					color.Red(u)
				}
			} else if len(split) == 2 {
				color.Red(u)
			} else if !s {
				fmt.Println(u)
			}
		}
	}
	fmt.Println("\n" + strconv.Itoa(len(resultUrlOther)) + " URL to other")
	for _, u := range resultUrlOther {
		if strings.Contains(u, "  |  ") || !s {
			split := strings.Split(u, "  |  ")
			if len(split) == 3 {
				if strings.HasPrefix(split[2], "2") {
					color.Green(u)
				} else if strings.HasPrefix(split[2], "3") {
					color.Yellow(u)
				} else {
					color.Red(u)
				}
			} else if len(split) == 2 {
				color.Red(u)
			} else if !s {
				fmt.Println(u)
			}
		}
	}
}

//蜘蛛抓取页面内容
func spider(ur string, is bool) {

	//标记完成
	defer wg.Done()
	url, _ := url.QueryUnescape(ur)
	if getEndUrl(url) {
		return
	}
	appendEndUrl(url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	//增加header选项
	reqest.Header.Add("Cookie", c)
	reqest.Header.Add("User-Agent", ua)

	//处理返回结果
	response, err := client.Do(reqest)
	if err != nil {
		return
	} else {
		defer response.Body.Close()

	}

	//提取url用于拼接其他url或js
	// 去读数据内容为 bytes
	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	path := response.Request.URL.Path
	host := response.Request.URL.Host
	scheme := response.Request.URL.Scheme

	//字节数组 转换成 字符串
	result := string(dataBytes)

	//提取js
	jsFind(result, host, scheme, path, is)
	//提取url
	urlFind(result, host, scheme, path, is)

}

//分析内容中的js
func jsFind(cont, host, scheme, path string, is bool) {
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
		"http[^\\s,^',^’,^\",^>,^<,^;,^(,^),^\\[]{2,250}?[.]js",
		"[\",']/[^\\s,^',^’,^\",^>,^<,^:,^;,\\\\*,^(,^),^\\[]2,250}?[.]js",
		"=[^\\s,^',^’,^\",^>,^<,^;,\\\\*,^(,^),^\\[]{2,250}?[.]js",
		"=[\",'][^\\s,^',^’,^\",^>,^<,^;,\\\\*,^(,^),^\\[]{2,250}?[.]js",
	}
	host = scheme + "://" + host
	for _, re := range res {
		re := regexp.MustCompile(re)
		jss := re.FindAllStringSubmatch(cont, -1)
		jss = jsFilter(jss)
		//循环提取js放到结果中
		for _, js := range jss {
			if strings.HasPrefix(js[0], "https:") || strings.HasPrefix(js[0], "http:") {
				appendJs(js[0])
				if is || m == 2 {
					wg.Add(1)
					go spider(js[0], false)
				}
			} else if strings.HasPrefix(js[0], "//") {
				appendJs(scheme + ":" + js[0])
				if is || m == 2 {
					wg.Add(1)
					go spider(scheme+":"+js[0], false)
				}

			} else if strings.HasPrefix(js[0], "/") {
				appendJs(host + js[0])
				if is || m == 2 {
					wg.Add(1)
					go spider(host+js[0], false)
				}
			} else {
				appendJs(host + cata + js[0])
				if is || m == 2 {
					wg.Add(1)
					go spider(host+cata+js[0], false)
				}
			}
		}
	}

}

//分析内容中的url
func urlFind(cont, host, scheme, path string, is bool) {
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
		"[\",']http[^\\s,^',^’,^\",^>,^<,^),^(]{2,250}?[\",']",
		"=http[^\\s,^',^’,^\",^>,^<,^),^(]{2,250}",
		"[\",']/[^\\s,^',^’,^\",^>,^<,^\\:,^),^(]{2,250}?[\",']",
		"(href|action).{0,3}=.{0,3}[\",'][^\\s,^',^’,^\",^>,^<,^),^(]{2,250}",
		"(href|action).{0,3}=.{0,3}[^\\s,^',^’,^\",^>,^<,^),^(]{2,250}",
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
				appendUrl(url[0])
				if is && m == 2 {
					wg.Add(1)
					go spider(url[0], false)
				}
			} else if strings.HasPrefix(url[0], "//") {
				appendUrl(scheme + ":" + url[0])
				if is && m == 2 {
					wg.Add(1)
					go spider(scheme+":"+url[0], false)
				}
			} else if strings.HasPrefix(url[0], "/") {
				appendUrl(host + url[0])
				if is && m == 2 {
					wg.Add(1)
					go spider(host+url[0], false)
				}
			} else if !strings.HasSuffix(path, ".js") {
				appendUrl(host + cata + url[0])
				if is && m == 2 {
					wg.Add(1)
					go spider(host+cata+url[0], false)
				}
			}
		}

	}

}

//过滤JS
func jsFilter(str [][]string) [][]string {
	//对不需要的数据过滤
	for i := range str {
		str[i][0] = strings.Replace(str[i][0], " ", "", -1)
		str[i][0] = strings.Replace(str[i][0], "\\/", "/", -1)
		str[i][0] = strings.Replace(str[i][0], "\"", "", -1)
		str[i][0] = strings.Replace(str[i][0], "%3A", ":", -1)
		str[i][0] = strings.Replace(str[i][0], "%2F", "/", -1)

		if strings.HasPrefix(str[i][0], "=") {
			str[i][0] = strings.Replace(str[i][0], "=", "", 1)
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

//过滤URL
func urlFilter(str [][]string) [][]string {
	//对不需要的数据过滤
	for i := range str {
		str[i][0] = strings.Replace(str[i][0], " ", "", -1)
		str[i][0] = strings.Replace(str[i][0], "\\/", "/", -1)
		str[i][0] = strings.Replace(str[i][0], "\"", "", -1)
		str[i][0] = strings.Replace(str[i][0], "'", "", -1)
		str[i][0] = strings.Replace(str[i][0], "href=\"", "", 1)
		str[i][0] = strings.Replace(str[i][0], "href='", "", 1)
		str[i][0] = strings.Replace(str[i][0], "%3A", ":", -1)
		str[i][0] = strings.Replace(str[i][0], "%2F", "/", -1)

		if strings.HasPrefix(str[i][0], "=") {
			str[i][0] = strings.Replace(str[i][0], "=", "", 1)
		}
		if strings.HasPrefix(str[i][0], "href=") {
			str[i][0] = strings.Replace(str[i][0], "href=", "", 1)
		}
		fstr := []string{".js?", ".css?", ".jpeg?", ".jpg?", ".png?", ".gif?", "github.com", "www.w3.org", "example.com", "<", ">", "{", "}", "[", "]", "|", "^", ";", "/js/", "location.href", "javascript:void"}
		for _, v := range fstr {
			if strings.Contains(str[i][0], v) {
				str[i][0] = ""

			}
		}
		if strings.HasSuffix(str[i][0], ".js") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ",") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ".css") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ".jpeg") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ".jpg") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ".png") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ".gif") {
			str[i][0] = ""
		} else if strings.HasSuffix(str[i][0], ".ico") {
			str[i][0] = ""
		}
	}
	return str
}

//检测js访问状态码
func jsState(u string, i int) {
	defer wg.Done()
	if !s {
		resultJs[i] = u
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: tr}
	reqest, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	//增加header选项
	reqest.Header.Add("Cookie", c)
	reqest.Header.Add("User-Agent", ua)
	//处理返回结果
	response, err := client.Do(reqest)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") {
			resultJs[i] = u + "  |  超时"
		} else {
			resultJs[i] = ""
		}
		return
	}

	code := response.StatusCode
	var length int
	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		length = 0
	} else {
		length = len(dataBytes)
	}
	resultJs[i] = u + "  |  " + strconv.Itoa(length) + "  |  " + strconv.Itoa(code)
}

//检测url访问状态码
func urlState(u string, i int) {
	defer wg.Done()
	if !s {
		resultUrl[i] = u
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 8 * time.Second, Transport: tr}
	reqest, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	//增加header选项
	reqest.Header.Add("Cookie", c)
	reqest.Header.Add("User-Agent", ua)
	//处理返回结果
	response, err := client.Do(reqest)
	if err != nil {
		if strings.Contains(err.Error(), "Client.Timeout") {
			resultUrl[i] = u + "  |  超时"
		} else {
			resultUrl[i] = ""
		}
		return
	}

	code := response.StatusCode
	var length int
	dataBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		length = 0
	} else {
		length = len(dataBytes)
	}
	resultUrl[i] = u + "  |  " + strconv.Itoa(length) + "  |  " + strconv.Itoa(code)
}

func appendJs(url string) {
	lock.Lock()
	defer lock.Unlock()
	for _, eachItem := range resultJs {
		if eachItem == url {
			return
		}
	}
	resultJs = append(resultJs, url)

}

func appendUrl(url string) {
	lock.Lock()
	defer lock.Unlock()
	for _, eachItem := range resultUrl {
		if eachItem == url {
			return
		}
	}
	resultUrl = append(resultUrl, url)

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

//对结果进行状态码排序
func SelectSort(arr []string) []string {
	length := len(arr)
	var sort []int
	for _, v := range arr {
		if strings.Contains(v, "  |  ") {
			if strings.Contains(v, "|  超时") {
				sort = append(sort, 999)
			} else {
				s := strings.Split(v, "  |  ")
				in, _ := strconv.Atoi(s[2])
				sort = append(sort, in)
			}
		} else {
			sort = append(sort, 1000)
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

//对结果进行状态码与URL排序排序
func urlDispose(arr []string, url, host string) ([]string, []string) {
	var urls []string
	var urlts []string
	var other []string
	for _, v := range arr {
		if strings.Contains(v, url) {
			urls = append(urls, v)
		} else {
			if strings.Contains(v, host) {
				urlts = append(urls, v)
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

//判断文件夹是否存在
func HasDir(path string) (bool, error) {
	_, _err := os.Stat(path)
	if _err == nil {
		return true, nil
	}
	if os.IsNotExist(_err) {
		return false, nil
	}
	return false, _err
}

//创建文件夹
func CreateDir(path string) {
	_exist, _err := HasDir(path)
	if _err != nil {
		fmt.Printf("获取文件夹异常 -> %v\n", _err)
		return
	}
	if _exist {
	} else {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			fmt.Printf("创建目录异常 -> %v\n", err)
		} else {
		}
	}
}

//提取顶级域名
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

//去重+去除错误url
func RemoveRepeatElement(list []string) []string {
	// 创建一个临时map用来存储数组元素
	temp := make(map[string]bool)
	var list2 []string
	index := 0
	for _, v := range list {
		if len(v) > 10 {
			re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
			hosts := re.FindAllString(v, 1)
			if len(hosts) != 0 {
				// 遍历数组元素，判断此元素是否已经存在map中
				_, ok := temp[v]
				if !ok {
					list2 = append(list2, v)
					temp[v] = true
				}
			}
		}
		index++

	}
	return list2
}
