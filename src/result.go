package main

import (
	"bufio"
	_ "embed"

	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//go:embed report.html
var html string

func outHtmlString(link Link) string {
	ht := `<tr class="ant-table-row ant-table-row-level-0" data-row-key="0">
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					<a href="` + link.Url + `" target="_blank" >
						` + link.Url + ` </a>
				</td>
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					` + link.Status + `
				</td>
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					` + link.Size + `
				</td>
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					` + link.Title + `
				</td>
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					<a href="` + link.Source + `" target="_blank" style="display:inline-block">
						` + link.Source + ` </a>
				</td>
			</tr>`
	return ht
}

func outHtmlInfoString(ty, val, sou string) string {
	ht := `<tr class="ant-table-row ant-table-row-level-0" data-row-key="0">
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					` + ty + `
				</td>
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					` + val + `
				</td>
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					<a href="` + sou + `" target="_blank" >
						` + sou + ` </a>
				</td>
			</tr>`
	return ht
}

// 导出csv
func outFileCsv() {
	addSource()
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
	//在当前文件夹创建文件夹
	err := os.MkdirAll(o+"/"+host, 0644)
	if err != nil {
		fmt.Printf(o+"/"+host+" 目录创建失败 ：%s", err)
		return
	}
	//多相同url处理
	fileName := o + "/" + host + "/" + host + ".csv"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = o + "/" + host + "/" + host + "(" + strconv.Itoa(fileNum) + ").csv"
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
		writer.WriteString("url,Source\n")
	} else {
		writer.WriteString("url,Status,Size,Title,Source\n")
	}

	if d == "" {
		writer.WriteString(strconv.Itoa(len(resultJsHost)) + " JS to " + getHost(u) + "\n")
	} else {
		writer.WriteString(strconv.Itoa(len(resultJsHost)+len(resultJsOther)) + " JS to " + d + "\n")
	}

	for _, j := range resultJsHost {
		if s != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"\n", j.Url, j.Status, j.Size, j.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", j.Url, j.Source))
		}
	}
	if d == "" {
		writer.WriteString("\n" + strconv.Itoa(len(resultJsOther)) + " JS to Other\n")
	}
	for _, j := range resultJsOther {
		if s != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"\n", j.Url, j.Status, j.Size, j.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", j.Url, j.Source))
		}
	}
	writer.WriteString("\n\n")
	if d == "" {
		writer.WriteString(strconv.Itoa(len(resultUrlHost)) + " URL to " + getHost(u) + "\n")
	} else {
		writer.WriteString(strconv.Itoa(len(resultUrlHost)+len(resultUrlOther)) + " URL to " + d + "\n")
	}

	for _, u := range resultUrlHost {
		if s != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", u.Url, u.Status, u.Size, u.Title, u.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\n", u.Url, u.Source))
		}
	}
	if d == "" {
		writer.WriteString("\n" + strconv.Itoa(len(resultUrlOther)) + " URL to Other\n")
	}
	for _, u := range resultUrlOther {
		if s != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", u.Url, u.Status, u.Size, u.Title, u.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", u.Url, u.Source))
		}
	}
	if s != "" && z != 0 {
		writer.WriteString("\n" + strconv.Itoa(len(fuzzs)) + " URL to Fuzz\n")
		fuzzs = SelectSort(fuzzs)
		for _, u := range fuzzs {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", u.Url, u.Status, u.Size, u.Title, "Fuzz"))
		}
	}

	writer.WriteString("\n Phone \n")
	for i := range infos {
		for i2 := range infos[i].Phone {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", infos[i].Phone[i2], infos[i].Source))
		}
	}
	writer.WriteString("\n Email \n")
	for i := range infos {
		for i2 := range infos[i].Email {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", infos[i].Email[i2], infos[i].Source))
		}
	}
	writer.WriteString("\n IDcard \n")
	for i := range infos {
		for i2 := range infos[i].IDcard {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", infos[i].IDcard[i2], infos[i].Source))
		}
	}
	writer.WriteString("\n JWT \n")
	for i := range infos {
		for i2 := range infos[i].JWT {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", infos[i].JWT[i2], infos[i].Source))
		}
	}

	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中

	fmt.Println(strconv.Itoa(len(resultJsHost)+len(resultJsOther))+"JS + "+strconv.Itoa(len(resultUrlHost)+len(resultUrlOther))+"URL --> ", file.Name())

	return
}

// 导出json
func outFileJson() {
	addSource()
	jsons := make(map[string]interface{})
	var info map[string][]map[string]string
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
	if len(infos) > 0 {
		info = make(map[string][]map[string]string)
		info["IDcard"] = nil
		info["JWT"] = nil
		info["Email"] = nil
		info["Phone"] = nil
	}

	for i := range infos {
		for i2 := range infos[i].IDcard {
			info["IDcard"] = append(info["IDcard"], map[string]string{"IDcard": infos[i].JWT[i2], "Source": infos[i].Source})
		}
		for i2 := range infos[i].JWT {
			info["JWT"] = append(info["JWT"], map[string]string{"JWT": infos[i].JWT[i2], "Source": infos[i].Source})
		}
		for i2 := range infos[i].Email {
			info["Email"] = append(info["Email"], map[string]string{"Email": infos[i].Email[i2], "Source": infos[i].Source})
		}
		for i2 := range infos[i].Phone {
			info["Phone"] = append(info["Phone"], map[string]string{"Phone": infos[i].Phone[i2], "Source": infos[i].Source})
		}
	}

	//输出到文件
	if strings.Contains(host, ":") {
		host = strings.Replace(host, ":", "：", -1)
	}
	//在当前文件夹创建文件夹
	err := os.MkdirAll(o+"/"+host, 0644)
	if err != nil {
		fmt.Printf(o+"/"+host+" 目录创建失败 ：%s", err)
		return
	}
	//多相同url处理
	fileName := o + "/" + host + "/" + host + ".json"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = o + "/" + host + "/" + host + "(" + strconv.Itoa(fileNum) + ").json"
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("创建失败：%s", err)
		return
	}
	if d == "" {
		jsons["jsOther"] = resultJsOther
		jsons["urlOther"] = resultUrlOther
	}
	jsons["js"] = resultJsHost
	jsons["url"] = resultUrlHost
	jsons["info"] = info
	jsons["fuzz"] = fuzzs
	if s != "" && z != 0 {
		fuzzs = SelectSort(fuzzs)
		jsons["fuzz"] = fuzzs
	}

	defer file.Close()

	data, err := json.Marshal(jsons)
	if err != nil {
		log.Printf("json化失败：%s", err)
		return
	}
	buf := bufio.NewWriter(file)
	// 字节写入
	buf.Write(data)
	// 将缓冲中的数据写入
	err = buf.Flush()
	if err != nil {
		log.Println("json保存失败:", err)
	}
	fmt.Println(strconv.Itoa(len(resultJsHost)+len(resultJsOther))+"JS + "+strconv.Itoa(len(resultUrlHost)+len(resultUrlOther))+"URL --> ", file.Name())
	return
}

// 导出html
func outFileHtml() {
	addSource()
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
	//在当前文件夹创建文件夹
	err := os.MkdirAll(o+"/"+host, 0644)
	if err != nil {
		fmt.Printf(o+"/"+host+" 目录创建失败 ：%s", err)
		return
	}
	//多相同url处理
	fileName := o + "/" + host + "/" + host + ".html"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = o + "/" + host + "/" + host + "(" + strconv.Itoa(fileNum) + ").html"
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

	if d == "" {
		html = strings.Replace(html, "{urlHost}", getHost(u), -1)
	} else {
		html = strings.Replace(html, "{urlHost}", d, -1)
	}
	var resultJsHostStr string
	for _, j := range resultJsHost {
		resultJsHostStr += outHtmlString(j)
	}
	html = strings.Replace(html, "{JS}", resultJsHostStr, -1)

	var resultJsOtherStr string
	for _, j := range resultJsOther {
		resultJsOtherStr += outHtmlString(j)
	}
	html = strings.Replace(html, "{JSOther}", resultJsOtherStr, -1)

	var resultUrlHostStr string
	for _, u := range resultUrlHost {
		resultUrlHostStr += outHtmlString(u)
	}
	html = strings.Replace(html, "{URL}", resultUrlHostStr, -1)

	var resultUrlOtherStr string
	for _, u := range resultUrlOther {
		resultUrlOtherStr += outHtmlString(u)
	}
	html = strings.Replace(html, "{URLOther}", resultUrlOtherStr, -1)

	var fuzzsStr string
	if s != "" && z != 0 {
		fuzzs = SelectSort(fuzzs)
		for _, u := range fuzzs {
			fuzzsStr += outHtmlString(u)
		}
	}
	html = strings.Replace(html, "{Fuzz}", fuzzsStr, -1)

	var infoStr string
	for i := range infos {
		for i2 := range infos[i].Phone {
			infoStr += outHtmlInfoString("Phone", infos[i].Phone[i2], infos[i].Source)
		}
	}
	for i := range infos {
		for i2 := range infos[i].Email {
			infoStr += outHtmlInfoString("Email", infos[i].Email[i2], infos[i].Source)
		}
	}
	for i := range infos {
		for i2 := range infos[i].IDcard {
			infoStr += outHtmlInfoString("IDcard", infos[i].IDcard[i2], infos[i].Source)
		}
	}
	for i := range infos {
		for i2 := range infos[i].JWT {
			infoStr += outHtmlInfoString("JWT", infos[i].JWT[i2], infos[i].Source)
		}
	}
	html = strings.Replace(html, "{Info}", infoStr, -1)
	writer.WriteString(html)
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
		resultUrl = SelectSort(resultUrl)

	}
	//抓取的域名优先排序
	resultJsHost, resultJsOther := urlDispose(resultJs, host, getHost(u))
	resultUrlHost, resultUrlOther := urlDispose(resultUrl, host, getHost(u))

	var ulen string
	if len(resultUrl) != 0 {
		uleni := 0
		for _, u := range resultUrl {
			uleni += len(u.Url)
		}
		ulen = strconv.Itoa(uleni/len(resultUrl) + 10)
	}
	var jlen string
	if len(resultJs) != 0 {
		jleni := 0
		for _, j := range resultJs {
			jleni += len(j.Url)
		}
		jlen = strconv.Itoa(jleni/len(resultJs) + 10)
	}
	if d == "" {
		fmt.Println(strconv.Itoa(len(resultJsHost)) + " JS to " + getHost(u))
	} else {
		fmt.Println(strconv.Itoa(len(resultJsHost)+len(resultJsOther)) + " JS to " + d)
	}
	for _, j := range resultJsHost {
		if s != "" {
			if strings.HasPrefix(j.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			} else if strings.HasPrefix(j.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			}
		} else if s == "" {
			fmt.Printf(color.LightBlue.Sprintf(j.Url) + "\n")
		}
	}
	if d == "" {
		fmt.Println("\n" + strconv.Itoa(len(resultJsOther)) + " JS to Other")
	}
	for _, j := range resultJsOther {
		if s != "" {
			if strings.HasPrefix(j.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			} else if strings.HasPrefix(j.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			}
		} else {
			fmt.Printf(color.LightBlue.Sprintf(j.Url) + "\n")
		}
	}

	//打印URL
	fmt.Println("\n\n")

	if d == "" {
		fmt.Println(strconv.Itoa(len(resultUrlHost)) + " URL to " + getHost(u))
	} else {
		fmt.Println(strconv.Itoa(len(resultUrlHost)+len(resultUrlOther)) + " URL to " + d)
	}

	for _, u := range resultUrlHost {
		if s != "" && len(u.Title) != 0 {
			if u.Status == "疑似危险路由" {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ %s ]\n", u.Status))
			} else if strings.HasPrefix(u.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else if strings.HasPrefix(u.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			}
		} else if s != "" {
			if strings.HasPrefix(u.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
			} else if strings.HasPrefix(u.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
			}
		} else {
			fmt.Printf(color.LightBlue.Sprintf(u.Url) + "\n")
		}
	}
	if d == "" {
		fmt.Println("\n" + strconv.Itoa(len(resultUrlOther)) + " URL to Other")
	}
	for _, u := range resultUrlOther {
		if s != "" && len(u.Title) != 0 {
			if u.Status == "疑似危险路由" {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ %s ]\n", u.Status))
			} else if strings.HasPrefix(u.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else if strings.HasPrefix(u.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			}
		} else if s != "" {
			if strings.HasPrefix(u.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
			} else if strings.HasPrefix(u.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
			}
		} else {
			fmt.Printf(color.LightBlue.Sprintf(u.Url) + "\n")
		}
	}

	if s != "" && z != 0 {
		fmt.Println("\n" + strconv.Itoa(len(fuzzs)) + " URL to Fuzz")
		fuzzs = SelectSort(fuzzs)
		for _, u := range fuzzs {
			if len(u.Title) != 0 {
				if u.Status == "疑似危险路由" {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ %s ]\n", u.Status))
				} else if strings.HasPrefix(u.Status, "2") {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
				} else if strings.HasPrefix(u.Status, "3") {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
				} else {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
				}
			} else {
				if strings.HasPrefix(u.Status, "2") {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
				} else if strings.HasPrefix(u.Status, "3") {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
				} else {
					fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s ]\n", u.Status, u.Size))
				}
			}
		}
	}
	fmt.Println("\n Phone ")
	for i := range infos {
		for i2 := range infos[i].Phone {
			fmt.Printf(color.LightBlue.Sprintf("%-10s", infos[i].Phone[i2]) + color.LightGreen.Sprintf(" [ %s ]\n", infos[i].Source))
		}
	}
	fmt.Println("\n Email ")
	for i := range infos {
		for i2 := range infos[i].Email {
			fmt.Printf(color.LightBlue.Sprintf("%-10s", infos[i].Email[i2]))
		}
	}
	fmt.Println("\n IDcard ")
	for i := range infos {
		for i2 := range infos[i].IDcard {
			fmt.Printf(color.LightBlue.Sprintf("%-10s", infos[i].IDcard[i2]))
		}
	}
	fmt.Println("\n JWT ")
	for i := range infos {
		for i2 := range infos[i].JWT {
			fmt.Printf(color.LightBlue.Sprintf("%-10s", infos[i].JWT[i2]))
		}
	}

}
