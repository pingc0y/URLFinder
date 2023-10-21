package crawler

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
)

// 蜘蛛抓取页面内容
func Spider(u string, num int) {
	is := true
	defer func() {
		config.Wg.Done()
		if is {
			<-config.Ch
		}

	}()
	config.Mux.Lock()
	fmt.Printf("\rStart %d Spider...", config.Progress)
	config.Progress++
	config.Mux.Unlock()
	//标记完成

	u, _ = url.QueryUnescape(u)
	if num > 1 && cmd.D != "" && !regexp.MustCompile(cmd.D).MatchString(u) {
		return
	}
	if GetEndUrl(u) {
		return
	}
	if cmd.M == 3 {
		for _, v := range config.Risks {
			if strings.Contains(u, v) {
				return
			}
		}
	}
	AppendEndUrl(u)
	request, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}

	request.Header.Set("Accept-Encoding", "gzip") //使用gzip压缩传输数据让访问更快
	request.Header.Set("User-Agent", util.GetUserAgent())
	request.Header.Set("Accept", "*/*")
	u_str, err := url.Parse(u)
	if err != nil {
		return
	}
	request.Header.Set("Referer", u_str.Scheme+"://"+u_str.Host) //####

	//增加header选项
	if cmd.C != "" {
		request.Header.Set("Cookie", cmd.C)
	}
	//加载yaml配置(headers)
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}

	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	result := ""
	//解压
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body) // gzip解压缩
		if err != nil {
			return
		}
		defer reader.Close()
		con, err := io.ReadAll(reader)
		if err != nil {
			return
		}
		result = string(con)
	} else {
		//提取url用于拼接其他url或js
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return
		}
		//字节数组 转换成 字符串
		result = string(dataBytes)
	}
	path := response.Request.URL.Path
	host := response.Request.URL.Host
	scheme := response.Request.URL.Scheme
	source := scheme + "://" + host + path
	judge_base := false //####
	//处理base标签
	re := regexp.MustCompile("base.{1,5}href.{1,5}(http.+?//[^\\s]+?)[\"'‘“]")
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
	} else { // 处理 "base 标签"
		re := regexp.MustCompile("(?i)base.{0,5}[:=]\\s*\"(.*?)\"")
		base := re.FindAllStringSubmatch(result, -1)
		if len(base) > 0 {
			pattern := "[^.\\/\\w]"
			re, _ := regexp.Compile(pattern)
			// 检查字符串是否包含匹配的字符
			result := re.MatchString(base[0][1])
			if !result { // 字符串中没有其他特殊字符
				if len(base[0][1]) > 1 && base[0][1][:2] == "./" { //base 路径从当前目录出发
					judge_base = true
					path = path[:strings.LastIndex(path, "/")] + base[0][1][1:]
				} else if len(base[0][1]) > 2 && base[0][1][:3] == "../" { //base 路径从上一级目录出发
					judge_base = true
					pattern := "^[./]+$"
					matched, _ := regexp.MatchString(pattern, base[0][1])
					if matched { // 处理的 base 路径中只有 ./的
						path = path[:strings.LastIndex(path, "/")+1] + base[0][1]
					} else {
						find_str := ""
						if strings.Contains(strings.TrimPrefix(base[0][1], "../"), "/") {
							find_str = base[0][1][3 : strings.Index(strings.TrimPrefix(base[0][1], "../"), "/")+3]
						} else {
							find_str = base[0][1][3:]
						}
						if strings.Contains(path, find_str) {
							path = path[:strings.Index(path, find_str)] + base[0][1][3:]
						} else {
							path = path[:strings.LastIndex(path, "/")+1] + base[0][1]
						}
					}
				} else if len(base[0][1]) > 4 && strings.HasPrefix(base[0][1], "http") { // base标签包含协议
					judge_base = true
					path = base[0][1]
				} else if len(base[0][1]) > 0 {
					judge_base = true
					if base[0][1][0] == 47 { //base 路径从根目录出发
						path = base[0][1]
					} else { //base 路径未指明从哪出发
						find_str := ""
						if strings.Contains(base[0][1], "/") {
							find_str = base[0][1][:strings.Index(base[0][1], "/")]
						} else {
							find_str = base[0][1]
						}
						if strings.Contains(path, find_str) {
							path = path[:strings.Index(path, find_str)] + base[0][1]
						} else {
							path = path[:strings.LastIndex(path, "/")+1] + base[0][1]
						}
					}
				}
				if !strings.HasSuffix(path, "/") {
					path += "/"
				}
			}
		}
	}

	is = false
	<-config.Ch
	//提取js
	jsFind(result, host, scheme, path, u, num, judge_base)
	//提取url
	urlFind(result, host, scheme, path, u, num, judge_base)
	// 防止base判断错误
	if judge_base {
		jsFind(result, host, scheme, path, u, num, false)
		urlFind(result, host, scheme, path, u, num, false)
	}
	//提取信息
	infoFind(result, source)

}

// 打印Validate进度
func PrintProgress() {
	config.Mux.Lock()
	num := len(result.ResultJs) + len(result.ResultUrl)
	fmt.Printf("\rValidate %.0f%%", float64(config.Progress+1)/float64(num+1)*100)
	config.Progress++
	config.Mux.Unlock()
}
