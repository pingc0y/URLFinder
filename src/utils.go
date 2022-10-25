package main

import (
	"encoding/base64"
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
			if host != "" && strings.Contains(v[0], host) {
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
					v[0] = strings.Replace(v[0], "/./", "/", -1)
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
			con.Proxy = map[string]string{"host": "", "username": "", "password": ""}
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
	if len(conf.Proxy["host"]) > 0 {
		proxyUrl, parseErr := url.Parse(conf.Proxy["host"])
		if parseErr != nil {
			fmt.Println("代理地址错误: \n" + parseErr.Error())
			os.Exit(1)
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		if len(conf.Proxy["username"]) > 0 && len(conf.Proxy["password"]) > 0 {
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(conf.Proxy["username"]+":"+conf.Proxy["password"]))
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)
		}
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
		target := regexp.MustCompile(".*/([^/]+)").FindAllStringSubmatch(parse.Path, -1)
		for _, v := range catalogue {
			catalogues = append(catalogues, v[1])
		}
		if len(target) > 0 {
			targets = append(targets, target[0][1])
		}
	}
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
func del404(urls [][]string) [][]string {
	is := make(map[string]int)
	//根据长度分别存放
	for _, v := range urls {
		arr, ok := is[v[2]]
		if ok {
			is[v[2]] = arr + 1
		} else {
			is[v[2]] = 1
		}

	}
	res := [][]string{}
	//如果某个长度的数量大于全部的3分之2，那么就判定它是404页面
	for i, v := range is {
		if v > len(urls)/2 {
			for _, vv := range urls {
				if vv[2] != i {
					res = append(res, vv)
				}
			}
		}
	}
	return res

}
