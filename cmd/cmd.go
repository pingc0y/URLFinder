package cmd

import (
	"flag"
	"fmt"
	"github.com/gookit/color"
	"os"
)

var (
	h bool
	I bool
	M int
	S string
	U string
	D string
	C string
	A string
	b string
	F string
	O string
	X string
	T = 50
	Z int
)

func init() {
	flag.StringVar(&A, "a", "", "set user-agent\n设置user-agent请求头")
	flag.StringVar(&b, "b", "", "set baseurl\n设置baseurl路径")
	flag.StringVar(&C, "c", "", "set cookie\n设置cookie")
	flag.StringVar(&D, "d", "", "set domainName\n指定获取的域名")
	flag.StringVar(&F, "f", "", "set urlFile\n批量抓取url,指定文件路径")
	flag.BoolVar(&h, "h", false, "this help\n帮助信息")
	flag.BoolVar(&I, "i", false, "set configFile\n加载yaml配置文件（不存在时，会在当前目录创建一个默认yaml配置文件）")
	flag.IntVar(&M, "m", 1, "set mode\n抓取模式 \n   1 normal\n     正常抓取（默认） \n   2 thorough\n     深入抓取 （url深入一层,js深入三层，防止抓偏） \n   3 security\n     安全深入抓取（过滤delete，remove等敏感路由） \n   ")
	flag.StringVar(&O, "o", "", "set outFile\n结果导出到csv文件，需指定导出文件目录（.代表当前目录）")
	flag.StringVar(&S, "s", "", "set Status\n显示指定状态码，all为显示全部（多个状态码用,隔开）")
	flag.IntVar(&T, "t", 50, "set thread\n设置线程数（默认50）\n")
	flag.StringVar(&U, "u", "", "set Url\n目标URL")
	flag.StringVar(&X, "x", "", "set httpProxy\n设置代理,格式: http://username:password@127.0.0.1:8809")
	flag.IntVar(&Z, "z", 0, "set Fuzz\n对404链接进行fuzz(只对主域名下的链接生效,需要与-s一起使用） \n   1 decreasing\n     目录递减fuzz \n   2 2combination\n     2级目录组合fuzz（适合少量链接使用） \n   3 3combination\n     3级目录组合fuzz（适合少量链接使用） \n")

	// 改变默认的 Usage
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `Usage: URLFinder [-a user-agent] [-b baseurl] [-c cookie] [-d domainName] [-f urlFile]  [-h help]  [-i configFile]  [-m mode] [-o outFile]  [-s Status] [-t thread] [-u Url] [-x httpProxy] [-z fuzz]

Options:
`)
	flag.PrintDefaults()
}

func Parse() {
	color.LightCyan.Println("         __   __   ___ _           _           \n /\\ /\\  /__\\ / /  / __(_)_ __   __| | ___ _ __ \n/ / \\ \\/ \\/// /  / _\\ | | '_ \\ / _` |/ _ \\ '__|\n\\ \\_/ / _  \\ /___ /   | | | | | (_| |  __/ |   \n \\___/\\/ \\_\\____\\/    |_|_| |_|\\__,_|\\___|_|     \n\nBy: pingc0y\nUpdateTime: 2023/2/3\nGithub: https://github.com/pingc0y/URLFinder \n")
	flag.Parse()
	if h || (U == "" && F == "") {
		flag.Usage()
		os.Exit(0)
	}

}
