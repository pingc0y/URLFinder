package result

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/mode"
	"github.com/pingc0y/URLFinder/util"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//go:embed report.html
var html string

var (
	ResultJs  []mode.Link
	ResultUrl []mode.Link
	Fuzzs     []mode.Link
	Infos     []mode.Info

	EndUrl   []string
	Jsinurl  map[string]string
	Jstourl  map[string]string
	Urltourl map[string]string
	Domains  []string
)

func outHtmlString(link mode.Link) string {
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
					<a href="` + link.Source + `" target="_blank" style="display:inline-bconfig.Lock">
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

func outHtmlDomainString(domain string) string {
	ht := `<tr class="ant-table-row ant-table-row-level-0" data-row-key="0">
				<td class="ant-table-column-has-actions ant-table-column-has-sorters">
					` + domain + `
				</td>
			</tr>`
	return ht
}

// 导出csv
func OutFileCsv() {
	//获取域名
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(cmd.U, 1)
	if len(hosts) == 0 {
		host = cmd.U
	} else {
		host = hosts[0]
	}

	//抓取的域名优先排序
	if cmd.S != "" {
		ResultUrl = util.SelectSort(ResultUrl)
		ResultJs = util.SelectSort(ResultJs)
	}
	ResultJsHost, ResultJsOther := util.UrlDispose(ResultJs, host, util.GetHost(cmd.U))
	ResultUrlHost, ResultUrlOther := util.UrlDispose(ResultUrl, host, util.GetHost(cmd.U))
	Domains = util.GetDomains(util.MergeArray(ResultJs, ResultUrl))
	//输出到文件
	if strings.Contains(host, ":") {
		host = strings.Replace(host, ":", "：", -1)
	}
	//在当前文件夹创建文件夹
	err := os.MkdirAll(cmd.O+"/"+host, 0644)
	if err != nil {
		fmt.Printf(cmd.O+"/"+host+" 目录创建失败 ：%s", err)
		return
	}
	//多相同url处理
	fileName := cmd.O + "/" + host + "/" + host + ".csv"
	for fileNum := 1; util.Exists(fileName); fileNum++ {
		fileName = cmd.O + "/" + host + "/" + host + "(" + strconv.Itoa(fileNum) + ").csv"
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
	if cmd.S == "" {
		writer.WriteString("url,Source\n")
	} else {
		writer.WriteString("url,Status,Size,Title,Source\n")
	}

	if cmd.D == "" {
		writer.WriteString(strconv.Itoa(len(ResultJsHost)) + " JS to " + util.GetHost(cmd.U) + "\n")
	} else {
		writer.WriteString(strconv.Itoa(len(ResultJsHost)+len(ResultJsOther)) + " JS to " + cmd.D + "\n")
	}

	for _, j := range ResultJsHost {
		if cmd.S != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"\n", j.Url, j.Status, j.Size, j.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", j.Url, j.Source))
		}
	}
	if cmd.D == "" {
		writer.WriteString("\n" + strconv.Itoa(len(ResultJsOther)) + " JS to Other\n")
	}
	for _, j := range ResultJsOther {
		if cmd.S != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",,\"%s\"\n", j.Url, j.Status, j.Size, j.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", j.Url, j.Source))
		}
	}
	writer.WriteString("\n\n")
	if cmd.D == "" {
		writer.WriteString(strconv.Itoa(len(ResultUrlHost)) + " URL to " + util.GetHost(cmd.U) + "\n")
	} else {
		writer.WriteString(strconv.Itoa(len(ResultUrlHost)+len(ResultUrlOther)) + " URL to " + cmd.D + "\n")
	}

	for _, u := range ResultUrlHost {
		if cmd.S != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", u.Url, u.Status, u.Size, u.Title, u.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\n", u.Url, u.Source))
		}
	}
	if cmd.D == "" {
		writer.WriteString("\n" + strconv.Itoa(len(ResultUrlOther)) + " URL to Other\n")
	}
	for _, u := range ResultUrlOther {
		if cmd.S != "" {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", u.Url, u.Status, u.Size, u.Title, u.Source))
		} else {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", u.Url, u.Source))
		}
	}
	if cmd.S != "" && cmd.Z != 0 {
		writer.WriteString("\n" + strconv.Itoa(len(Fuzzs)) + " URL to Fuzz\n")
		Fuzzs = util.SelectSort(Fuzzs)
		for _, u := range Fuzzs {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", u.Url, u.Status, u.Size, u.Title, "Fuzz"))
		}
	}

	writer.WriteString("\n" + strconv.Itoa(len(Domains)) + " Domain\n")
	for _, u := range Domains {
		writer.WriteString(fmt.Sprintf("\"%s\"\n", u))
	}

	writer.WriteString("\n Phone \n")
	for i := range Infos {
		for i2 := range Infos[i].Phone {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", Infos[i].Phone[i2], Infos[i].Source))
		}
	}
	writer.WriteString("\n Email \n")
	for i := range Infos {
		for i2 := range Infos[i].Email {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", Infos[i].Email[i2], Infos[i].Source))
		}
	}
	writer.WriteString("\n IDcard \n")
	for i := range Infos {
		for i2 := range Infos[i].IDcard {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", Infos[i].IDcard[i2], Infos[i].Source))
		}
	}
	writer.WriteString("\n JWT \n")
	for i := range Infos {
		for i2 := range Infos[i].JWT {
			writer.WriteString(fmt.Sprintf("\"%s\",\"%s\"\n", Infos[i].JWT[i2], Infos[i].Source))
		}
	}

	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中

	fmt.Println(strconv.Itoa(len(ResultJsHost)+len(ResultJsOther))+"JS + "+strconv.Itoa(len(ResultUrlHost)+len(ResultUrlOther))+"URL --> ", file.Name())

	return
}

// 导出json
func OutFileJson() {
	jsons := make(map[string]interface{})
	var info map[string][]map[string]string
	//获取域名
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(cmd.U, 1)
	if len(hosts) == 0 {
		host = cmd.U
	} else {
		host = hosts[0]
	}
	//抓取的域名优先排序
	if cmd.S != "" {
		ResultUrl = util.SelectSort(ResultUrl)
		ResultJs = util.SelectSort(ResultJs)
	}
	ResultJsHost, ResultJsOther := util.UrlDispose(ResultJs, host, util.GetHost(cmd.U))
	ResultUrlHost, ResultUrlOther := util.UrlDispose(ResultUrl, host, util.GetHost(cmd.U))
	Domains = util.GetDomains(util.MergeArray(ResultJs, ResultUrl))

	if len(Infos) > 0 {
		info = make(map[string][]map[string]string)
		info["IDcard"] = nil
		info["JWT"] = nil
		info["Email"] = nil
		info["Phone"] = nil
	}

	for i := range Infos {
		for i2 := range Infos[i].IDcard {
			info["IDcard"] = append(info["IDcard"], map[string]string{"IDcard": Infos[i].JWT[i2], "Source": Infos[i].Source})
		}
		for i2 := range Infos[i].JWT {
			info["JWT"] = append(info["JWT"], map[string]string{"JWT": Infos[i].JWT[i2], "Source": Infos[i].Source})
		}
		for i2 := range Infos[i].Email {
			info["Email"] = append(info["Email"], map[string]string{"Email": Infos[i].Email[i2], "Source": Infos[i].Source})
		}
		for i2 := range Infos[i].Phone {
			info["Phone"] = append(info["Phone"], map[string]string{"Phone": Infos[i].Phone[i2], "Source": Infos[i].Source})
		}
	}

	//输出到文件
	if strings.Contains(host, ":") {
		host = strings.Replace(host, ":", "：", -1)
	}
	//在当前文件夹创建文件夹
	err := os.MkdirAll(cmd.O+"/"+host, 0644)
	if err != nil {
		fmt.Printf(cmd.O+"/"+host+" 目录创建失败 ：%s", err)
		return
	}
	//多相同url处理
	fileName := cmd.O + "/" + host + "/" + host + ".json"
	for fileNum := 1; util.Exists(fileName); fileNum++ {
		fileName = cmd.O + "/" + host + "/" + host + "(" + strconv.Itoa(fileNum) + ").json"
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("创建失败：%s", err)
		return
	}
	if cmd.D == "" {
		jsons["jsOther"] = ResultJsOther
		jsons["urlOther"] = ResultUrlOther
	}
	jsons["js"] = ResultJsHost
	jsons["url"] = ResultUrlHost
	jsons["info"] = info
	jsons["fuzz"] = Fuzzs
	jsons["domain"] = Domains
	if cmd.S != "" && cmd.Z != 0 {
		Fuzzs = util.SelectSort(Fuzzs)
		jsons["fuzz"] = Fuzzs
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
	fmt.Println(strconv.Itoa(len(ResultJsHost)+len(ResultJsOther))+"JS + "+strconv.Itoa(len(ResultUrlHost)+len(ResultUrlOther))+"URL --> ", file.Name())
	return
}

// 导出html
func OutFileHtml() {
	//获取域名
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(cmd.U, 1)
	if len(hosts) == 0 {
		host = cmd.U
	} else {
		host = hosts[0]
	}

	//抓取的域名优先排序
	if cmd.S != "" {
		ResultUrl = util.SelectSort(ResultUrl)
		ResultJs = util.SelectSort(ResultJs)
	}
	ResultJsHost, ResultJsOther := util.UrlDispose(ResultJs, host, util.GetHost(cmd.U))
	ResultUrlHost, ResultUrlOther := util.UrlDispose(ResultUrl, host, util.GetHost(cmd.U))
	Domains = util.GetDomains(util.MergeArray(ResultJs, ResultUrl))
	//输出到文件
	if strings.Contains(host, ":") {
		host = strings.Replace(host, ":", "：", -1)
	}
	//在当前文件夹创建文件夹
	err := os.MkdirAll(cmd.O+"/"+host, 0644)
	if err != nil {
		fmt.Printf(cmd.O+"/"+host+" 目录创建失败 ：%s", err)
		return
	}
	//多相同url处理
	fileName := cmd.O + "/" + host + "/" + host + ".html"
	for fileNum := 1; util.Exists(fileName); fileNum++ {
		fileName = cmd.O + "/" + host + "/" + host + "(" + strconv.Itoa(fileNum) + ").html"
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

	if cmd.D == "" {
		html = strings.Replace(html, "{urlHost}", util.GetHost(cmd.U), -1)
	} else {
		html = strings.Replace(html, "{urlHost}", cmd.D, -1)
	}
	var ResultJsHostStr string
	for _, j := range ResultJsHost {
		ResultJsHostStr += outHtmlString(j)
	}
	html = strings.Replace(html, "{JS}", ResultJsHostStr, -1)

	var ResultJsOtherStr string
	for _, j := range ResultJsOther {
		ResultJsOtherStr += outHtmlString(j)
	}
	html = strings.Replace(html, "{JSOther}", ResultJsOtherStr, -1)

	var ResultUrlHostStr string
	for _, u := range ResultUrlHost {
		ResultUrlHostStr += outHtmlString(u)
	}
	html = strings.Replace(html, "{URL}", ResultUrlHostStr, -1)

	var ResultUrlOtherStr string
	for _, u := range ResultUrlOther {
		ResultUrlOtherStr += outHtmlString(u)
	}
	html = strings.Replace(html, "{URLOther}", ResultUrlOtherStr, -1)

	var FuzzsStr string
	if cmd.S != "" && cmd.Z != 0 {
		Fuzzs = util.SelectSort(Fuzzs)
		for _, u := range Fuzzs {
			FuzzsStr += outHtmlString(u)
		}
	}
	html = strings.Replace(html, "{Fuzz}", FuzzsStr, -1)

	var DomainsStr string
	for _, u := range Domains {
		DomainsStr += outHtmlDomainString(u)
	}
	html = strings.Replace(html, "{Domains}", DomainsStr, -1)

	var Infostr string
	for i := range Infos {
		for i2 := range Infos[i].Phone {
			Infostr += outHtmlInfoString("Phone", Infos[i].Phone[i2], Infos[i].Source)
		}
	}
	for i := range Infos {
		for i2 := range Infos[i].Email {
			Infostr += outHtmlInfoString("Email", Infos[i].Email[i2], Infos[i].Source)
		}
	}
	for i := range Infos {
		for i2 := range Infos[i].IDcard {
			Infostr += outHtmlInfoString("IDcard", Infos[i].IDcard[i2], Infos[i].Source)
		}
	}
	for i := range Infos {
		for i2 := range Infos[i].JWT {
			Infostr += outHtmlInfoString("JWT", Infos[i].JWT[i2], Infos[i].Source)
		}
	}
	html = strings.Replace(html, "{Info}", Infostr, -1)
	writer.WriteString(html)
	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中
	fmt.Println(strconv.Itoa(len(ResultJsHost)+len(ResultJsOther))+"JS + "+strconv.Itoa(len(ResultUrlHost)+len(ResultUrlOther))+"URL --> ", file.Name())
	return
}

// 打印
func Print() {
	//获取域名
	var host string
	re := regexp.MustCompile("([a-z0-9\\-]+\\.)*([a-z0-9\\-]+\\.[a-z0-9\\-]+)(:[0-9]+)?")
	hosts := re.FindAllString(cmd.U, 1)
	if len(hosts) == 0 {
		host = cmd.U
	} else {
		host = hosts[0]
	}
	//打印JS
	if cmd.S != "" {
		ResultJs = util.SelectSort(ResultJs)
		ResultUrl = util.SelectSort(ResultUrl)

	}
	//抓取的域名优先排序
	ResultJsHost, ResultJsOther := util.UrlDispose(ResultJs, host, util.GetHost(cmd.U))
	ResultUrlHost, ResultUrlOther := util.UrlDispose(ResultUrl, host, util.GetHost(cmd.U))
	Domains = util.GetDomains(util.MergeArray(ResultJs, ResultUrl))
	var ulen string
	if len(ResultUrl) != 0 {
		uleni := 0
		for _, u := range ResultUrl {
			uleni += len(u.Url)
		}
		ulen = strconv.Itoa(uleni/len(ResultUrl) + 10)
	}
	var jlen string
	if len(ResultJs) != 0 {
		jleni := 0
		for _, j := range ResultJs {
			jleni += len(j.Url)
		}
		jlen = strconv.Itoa(jleni/len(ResultJs) + 10)
	}
	if cmd.D == "" {
		fmt.Println(strconv.Itoa(len(ResultJsHost)) + " JS to " + util.GetHost(cmd.U))
	} else {
		fmt.Println(strconv.Itoa(len(ResultJsHost)+len(ResultJsOther)) + " JS to " + cmd.D)
	}
	for _, j := range ResultJsHost {
		if cmd.S != "" {
			if strings.HasPrefix(j.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			} else if strings.HasPrefix(j.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+jlen+"s", j.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s ]\n", j.Status, j.Size))
			}
		} else if cmd.S == "" {
			fmt.Printf(color.LightBlue.Sprintf(j.Url) + "\n")
		}
	}
	if cmd.D == "" {
		fmt.Println("\n" + strconv.Itoa(len(ResultJsOther)) + " JS to Other")
	}
	for _, j := range ResultJsOther {
		if cmd.S != "" {
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

	fmt.Println("\n  ")

	if cmd.D == "" {
		fmt.Println(strconv.Itoa(len(ResultUrlHost)) + " URL to " + util.GetHost(cmd.U))
	} else {
		fmt.Println(strconv.Itoa(len(ResultUrlHost)+len(ResultUrlOther)) + " URL to " + cmd.D)
	}

	for _, u := range ResultUrlHost {
		if cmd.S != "" && len(u.Title) != 0 {
			if u.Status == "疑似危险路由" {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ %s ]\n", u.Status))
			} else if strings.HasPrefix(u.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else if strings.HasPrefix(u.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			}
		} else if cmd.S != "" {
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
	if cmd.D == "" {
		fmt.Println("\n" + strconv.Itoa(len(ResultUrlOther)) + " URL to Other")
	}
	for _, u := range ResultUrlOther {
		if cmd.S != "" && len(u.Title) != 0 {
			if u.Status == "疑似危险路由" {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ %s ]\n", u.Status))
			} else if strings.HasPrefix(u.Status, "2") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightGreen.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else if strings.HasPrefix(u.Status, "3") {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightYellow.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			} else {
				fmt.Printf(color.LightBlue.Sprintf("%-"+ulen+"s", u.Url) + color.LightRed.Sprintf(" [ Status: %s, Size: %s, Title: %s ]\n", u.Status, u.Size, u.Title))
			}
		} else if cmd.S != "" {
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

	if cmd.S != "" && cmd.Z != 0 {
		fmt.Println("\n" + strconv.Itoa(len(Fuzzs)) + " URL to Fuzz")
		Fuzzs = util.SelectSort(Fuzzs)
		for _, u := range Fuzzs {
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
	fmt.Println("\n" + strconv.Itoa(len(Domains)) + " Domain")
	for _, u := range Domains {
		fmt.Printf(color.LightBlue.Sprintf("%s \n", u))

	}

	if len(Infos) > 0 {
		fmt.Println("\n Phone ")
		for i := range Infos {
			for i2 := range Infos[i].Phone {
				fmt.Printf(color.LightBlue.Sprintf("%-10s", Infos[i].Phone[i2]) + color.LightGreen.Sprintf(" [ %s ]\n", Infos[i].Source))
			}
		}
		fmt.Println("\n Email ")
		for i := range Infos {
			for i2 := range Infos[i].Email {
				fmt.Printf(color.LightBlue.Sprintf("%-10s", Infos[i].Email[i2]) + color.LightGreen.Sprintf(" [ %s ]\n", Infos[i].Source))
			}
		}
		fmt.Println("\n IDcard ")
		for i := range Infos {
			for i2 := range Infos[i].IDcard {
				fmt.Printf(color.LightBlue.Sprintf("%-10s", Infos[i].IDcard[i2]) + color.LightGreen.Sprintf(" [ %s ]\n", Infos[i].Source))
			}
		}
		fmt.Println("\n JWT ")
		for i := range Infos {
			for i2 := range Infos[i].JWT {
				fmt.Printf(color.LightBlue.Sprintf("%-10s", Infos[i].JWT[i2]) + color.LightGreen.Sprintf(" [ %s ]\n", Infos[i].Source))
			}
		}

	}

}
