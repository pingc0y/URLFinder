package util

import (
	"encoding/json"
	"fmt"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/mode"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return s.IsDir()
}

// MergeArray 合并数组
func MergeArray(dest []mode.Link, src []mode.Link) (result []mode.Link) {
	result = make([]mode.Link, len(dest)+len(src))
	//将第一个数组传入result
	copy(result, dest)
	//将第二个数组接在尾部,也就是 len(dest):
	copy(result[len(dest):], src)
	return
}

// 对结果进行状态码排序
func SelectSort(arr []mode.Link) []mode.Link {
	length := len(arr)
	var sort []int
	for _, v := range arr {
		if v.Url == "" || len(v.Size) == 0 || v.Status == "timeout" {
			sort = append(sort, 999)
		} else {
			in, _ := strconv.Atoi(v.Status)
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
func UrlDispose(arr []mode.Link, url, host string) ([]mode.Link, []mode.Link) {
	var urls []mode.Link
	var urlts []mode.Link
	var other []mode.Link
	for _, v := range arr {
		if strings.Contains(v.Url, url) {
			urls = append(urls, v)
		} else {
			if host != "" && regexp.MustCompile(host).MatchString(v.Url) {
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

// 处理Headers配置
func SetHeadersConfig(header *http.Header) *http.Header {
	for k, v := range config.Conf.Headers {
		header.Add(k, v)
	}
	return header
}

// 设置proxy配置
func SetProxyConfig(tr *http.Transport) *http.Transport {
	if len(config.Conf.Proxy) > 0 {
		proxyUrl, parseErr := url.Parse(config.Conf.Proxy)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	return tr
}

// 提取顶级域名
func GetHost(u string) string {
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
func RemoveRepeatElement(list []mode.Link) []mode.Link {
	// 创建一个临时map用来存储数组元素
	temp := make(map[string]bool)
	var list2 []mode.Link
	index := 0
	for _, v := range list {

		//处理-d参数
		if cmd.D != "" {
			v.Url = domainNameFilter(v.Url)
		}
		if len(v.Url) > 10 {
			re := regexp.MustCompile("://([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
			hosts := re.FindAllString(v.Url, 1)
			if len(hosts) != 0 {
				// 遍历数组元素,判断此元素是否已经存在map中
				_, ok := temp[v.Url]
				if !ok {
					v.Url = strings.Replace(v.Url, "/./", "/", -1)
					list2 = append(list2, v)
					temp[v.Url] = true
				}
			}
		}
		index++

	}
	return list2
}

// 打印Fuzz进度
func PrintFuzz() {
	config.Mux.Lock()
	fmt.Printf("\rFuzz %.0f%%", float64(config.Progress+1)/float64(config.FuzzNum+1)*100)
	config.Progress++
	config.Mux.Unlock()
}

// 处理-d
func domainNameFilter(url string) string {
	re := regexp.MustCompile("://([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(url, 1)
	if len(hosts) != 0 {
		if !regexp.MustCompile(cmd.D).MatchString(hosts[0]) {
			url = ""
		}
	}
	return url
}

// 文件是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 数组去重
func UniqueArr(arr []string) []string {
	newArr := make([]string, 0)
	tempArr := make(map[string]bool, len(newArr))
	for _, v := range arr {
		if tempArr[v] == false {
			tempArr[v] = true
			newArr = append(newArr, v)
		}
	}
	return newArr
}

func GetDomains(lis []mode.Link) []string {
	var urls []string
	for i := range lis {
		re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
		hosts := re.FindAllString(lis[i].Url, 1)
		if len(hosts) > 0 {
			urls = append(urls, hosts[0])
		}
	}
	return UniqueArr(urls)
}

// 提取fuzz的目录结构
func PathExtract(urls []string) ([]string, []string) {
	var catalogues []string
	var targets []string
	if len(urls) == 0 {
		return nil, nil
	}
	par, _ := url.Parse(urls[0])
	host := par.Scheme + "://" + par.Host
	for _, v := range urls {
		parse, _ := url.Parse(v)
		catalogue := regexp.MustCompile("([^/]+?)/").FindAllStringSubmatch(parse.Path, -1)
		if !strings.HasSuffix(parse.Path, "/") {
			target := regexp.MustCompile(".*/([^/]+)").FindAllStringSubmatch(parse.Path, -1)
			if len(target) > 0 {
				targets = append(targets, target[0][1])
			}
		}
		for _, v := range catalogue {
			if !strings.Contains(v[1], "..") {
				catalogues = append(catalogues, v[1])
			}
		}

	}
	targets = append(targets, "upload")
	catalogues = UniqueArr(catalogues)
	targets = UniqueArr(targets)
	url1 := catalogues
	url2 := []string{}
	url3 := []string{}
	var path []string
	for _, v1 := range url1 {
		for _, v2 := range url1 {
			if !strings.Contains(v2, v1) {
				url2 = append(url2, "/"+v2+"/"+v1)
			}
		}
	}
	if cmd.Z == 3 {
		for _, v1 := range url1 {
			for _, v3 := range url2 {
				if !strings.Contains(v3, v1) {
					url3 = append(url3, v3+"/"+v1)
				}
			}
		}
	}
	for i := range url1 {
		url1[i] = "/" + url1[i]
	}
	if cmd.Z == 3 {
		path = make([]string, len(url1)+len(url2)+len(url3))
	} else {
		path = make([]string, len(url1)+len(url2))
	}
	copy(path, url1)
	copy(path[len(url1):], url2)
	if cmd.Z == 3 {
		copy(path[len(url1)+len(url2):], url3)
	}
	for i := range path {
		path[i] = host + path[i] + "/"
	}
	path = append(path, host+"/")
	return path, targets
}

// 去除状态码非404的404链接
func Del404(urls []mode.Link) []mode.Link {
	is := make(map[int]int)
	//根据长度分别存放
	for _, v := range urls {
		arr, ok := is[len(v.Size)]
		if ok {
			is[len(v.Size)] = arr + 1
		} else {
			is[len(v.Size)] = 1
		}
	}
	res := []mode.Link{}
	//如果某个长度的数量大于全部的3分之2,那么就判定它是404页面
	for i, v := range is {
		if v > len(urls)/2 {
			for _, vv := range urls {
				if len(vv.Size) != i {
					res = append(res, vv)
				}
			}
		}
	}
	return res

}

var (

	// for each request, a random UA will be selected from this list
	uas = [...]string{
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.68 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.61 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.62 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.107 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.121 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.88 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.72 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.94 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.98 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.98 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.63 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.95 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.106 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.87 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.82 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.74 Safari/537.36 Edg/99.0.1150.46",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.87 Safari/537.36 SE 2.X MetaSr 1.0",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.25 Safari/537.36 Core/1.70.3883.400 QQBrowser/10.8.4559.400",
	}

	nuas = len(uas)
)

func GetUserAgent() string {
	if cmd.A == "" {
		cmd.A = uas[rand.Intn(nuas)]
	}
	return cmd.A
}

func GetUpdate() {

	url := fmt.Sprintf("https://api.github.com/repos/pingc0y/URLFinder/releases/latest")
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	resp, err := client.Get(url)
	if err != nil {
		cmd.XUpdate = "更新检测失败"
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cmd.XUpdate = "更新检测失败"
		return
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	err = json.Unmarshal(body, &release)
	if err != nil {
		cmd.XUpdate = "更新检测失败"
		return
	}
	if release.TagName == "" {
		cmd.XUpdate = "更新检测失败"
		return
	}
	if cmd.Update != release.TagName {
		cmd.XUpdate = "有新版本可用: " + release.TagName
	} else {
		cmd.XUpdate = "已是最新版本"
	}

}
