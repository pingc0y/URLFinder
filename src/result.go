package main

import (
	"bufio"
	"fmt"
	"github.com/gookit/color"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 导出csv
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

	//多相同url处理
	fileName := o + "/" + host + ".csv"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = o + "/" + host + "(" + strconv.Itoa(fileNum) + ").csv"
	}
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)

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
