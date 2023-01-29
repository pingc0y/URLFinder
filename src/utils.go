package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
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

// 对结果进行状态码排序
func SelectSort(arr []Link) []Link {
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
func urlDispose(arr []Link, url, host string) ([]Link, []Link) {
	var urls []Link
	var urlts []Link
	var other []Link
	for _, v := range arr {
		if strings.Contains(v.Url, url) {
			urls = append(urls, v)
		} else {
			if host != "" && strings.Contains(v.Url, host) {
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
func RemoveRepeatElement(list []Link) []Link {
	// 创建一个临时map用来存储数组元素
	temp := make(map[string]bool)
	var list2 []Link
	index := 0
	for _, v := range list {

		//处理-d参数
		if d != "" {
			v.Url = domainNameFilter(v.Url)
		}
		if len(v.Url) > 10 {
			re := regexp.MustCompile("://([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
			hosts := re.FindAllString(v.Url, 1)
			if len(hosts) != 0 {
				// 遍历数组元素，判断此元素是否已经存在map中
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

// 读取配置文件
func GetConfig(path string) {
	con := &config{}
	if f, err := os.Open(path); err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {

			con.Headers = map[string]string{"Cookie": c, "User-Agent": ua, "Accept": "*/*"}
			con.Proxy = ""
			data, err2 := yaml.Marshal(con)
			err2 = os.WriteFile(path, data, 0644)
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

// 处理Headers配置
func SetHeadersConfig(header *http.Header) *http.Header {
	for k, v := range conf.Headers {
		header.Add(k, v)
	}
	return header
}

// proxy
func SetProxyConfig(tr *http.Transport) *http.Transport {
	if len(conf.Proxy) > 0 {
		proxyUrl, parseErr := url.Parse(conf.Proxy)
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}
	return tr
}

// 打印Validate进度
func printProgress() {
	num := len(resultJs) + len(resultUrl)
	fmt.Printf("\rValidate %.0f%%", float64(progress+1)/float64(num+1)*100)
	mux.Lock()
	progress++
	mux.Unlock()
}

// 打印Fuzz进度
func printFuzz() {
	fmt.Printf("\rFuzz %.0f%%", float64(progress+1)/float64(fuzzNum+1)*100)
	mux.Lock()
	progress++
	mux.Unlock()
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

// 文件是否存在
func exists(path string) bool {
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
func uniqueArr(arr []string) []string {
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

// 提取fuzz的目录结构
func pathExtract(urls []string) ([]string, []string) {
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
	catalogues = uniqueArr(catalogues)
	targets = uniqueArr(targets)
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
	if z == 3 {
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
	if z == 3 {
		path = make([]string, len(url1)+len(url2)+len(url3))
	} else {
		path = make([]string, len(url1)+len(url2))
	}
	copy(path, url1)
	copy(path[len(url1):], url2)
	if z == 3 {
		copy(path[len(url1)+len(url2):], url3)
	}
	for i := range path {
		path[i] = host + path[i] + "/"
	}
	path = append(path, host+"/")
	return path, targets
}

// 去除状态码非404的404链接
func del404(urls []Link) []Link {
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
	res := []Link{}
	//如果某个长度的数量大于全部的3分之2，那么就判定它是404页面
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

func appendJs(url string, urltjs string) {
	lock.Lock()
	defer lock.Unlock()
	url = strings.Replace(url, "/./", "/", -1)
	for _, eachItem := range resultJs {
		if eachItem.Url == url {
			return
		}
	}
	resultJs = append(resultJs, Link{Url: url})
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
		if eachItem.Url == url {
			return
		}
	}
	resultUrl = append(resultUrl, Link{Url: url})
	if o != "" {
		urltourl[url] = urlturl
	}
}

func appendInfo(info Info) {
	lock.Lock()
	defer lock.Unlock()
	infos = append(infos, info)
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

func addSource() {
	for i := range resultJs {
		resultJs[i].Source = jstourl[resultJs[i].Url]
	}
	for i := range resultUrl {
		resultUrl[i].Source = urltourl[resultUrl[i].Url]
	}

}
