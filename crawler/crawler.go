package crawler

import (
	"compress/gzip"
	"fmt"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/result"
	"github.com/pingc0y/URLFinder/util"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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
	//增加header选项
	if cmd.C != "" {
		request.Header.Set("Cookie", cmd.C)
	}
	//加载yaml配置(headers)
	if cmd.I {
		util.SetHeadersConfig(&request.Header)
	}

	//处理返回结果
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//client = &http.Client{Timeout: time.Duration(cmd.TI) * time.Second,
	//	Transport: tr,
	//	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//		if len(via) >= 10 {
	//			return fmt.Errorf("Too many redirects")
	//		}
	//		if len(via) > 0 {
	//			if via[0] != nil && via[0].URL != nil {
	//				result.Redirect[via[0].URL.String()] = true
	//			} else {
	//				result.Redirect[req.URL.String()] = true
	//			}
	//
	//		}
	//		return nil
	//	},
	//}
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
	}
	is = false
	<-config.Ch
	//提取js
	jsFind(result, host, scheme, path, u, num)
	//提取url
	urlFind(result, host, scheme, path, u, num)
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
