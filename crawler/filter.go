package crawler

import (
	"github.com/pingc0y/URLFinder/config"
	"net/url"
	"regexp"
	"strings"
)

// 过滤JS
func jsFilter(str [][]string) [][]string {

	//对不需要的数据过滤
	for i := range str {
		str[i][0], _ = url.QueryUnescape(str[i][1])
		str[i][0] = strings.Replace(str[i][0], " ", "", -1)
		str[i][0] = strings.Replace(str[i][0], "\\/", "/", -1)
		str[i][0] = strings.Replace(str[i][0], "%3A", ":", -1)
		str[i][0] = strings.Replace(str[i][0], "%2F", "/", -1)

		//去除不是.js的链接
		if !strings.HasSuffix(str[i][0], ".js") && !strings.Contains(str[i][0], ".js?") {
			str[i][0] = ""
		}

		//过滤配置的黑名单
		for i2 := range config.JsFiler {
			re := regexp.MustCompile(config.JsFiler[i2])
			is := re.MatchString(str[i][0])
			if is {
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

		//去除不存在字符串和数字的url，判断为错误数据
		match, _ := regexp.MatchString("[a-zA-Z]+|[0-9]+", str[i][0])
		if !match {
			str[i][0] = ""
		}

		//对抓到的域名做处理
		re := regexp.MustCompile("([a-z0-9\\-]+\\.)+([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?").FindAllString(str[i][0], 1)
		if len(re) != 0 && !strings.HasPrefix(str[i][0], "http") && !strings.HasPrefix(str[i][0], "/") {
			str[i][0] = "http://" + str[i][0]
		}

		//过滤配置的黑名单
		for i2 := range config.UrlFiler {
			re := regexp.MustCompile(config.UrlFiler[i2])
			is := re.MatchString(str[i][0])
			if is {
				str[i][0] = ""
			}
		}

	}
	return str
}
