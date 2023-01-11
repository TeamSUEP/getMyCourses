package fetch

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/TeamSUEP/getMyCourses/config"
)

// 获取课程表所在页面源代码
func FetchCourses(cookieJar *cookiejar.Jar, semesterId string) (string, error) {
	fmt.Println("\n获取课表详情中...")

	// http 请求客户端
	client := &http.Client{
		Jar:       cookieJar,
		Transport: config.Tr,
	}

	// 第一次请求
	resp1, err := client.Get(config.SupwisdomUrl + "/eams/courseTableForStd.action")
	if err != nil {
		return "", err
	}
	defer resp1.Body.Close()

	// 读取
	content, err := io.ReadAll(resp1.Body)
	if err != nil {
		return "", err
	}

	// 检查
	temp := string(content)
	if !strings.Contains(temp, "bg.form.addInput(form,\"ids\",\"") {
		return "", errors.New("获取ids失败")
	}
	temp = temp[strings.Index(temp, "bg.form.addInput(form,\"ids\",\"")+29 : strings.Index(temp, "bg.form.addInput(form,\"ids\",\"")+50]
	ids := temp[:strings.Index(temp, "\");")]
	if semesterId == "null" {
		semesterId = resp1.Cookies()[0].Value
	}

	// 第二次请求
	formValues := url.Values{
		"ignoreHead":   {"1"},
		"setting.kind": {"std"},
		"startWeek":    {""},
		"semester.id":  {semesterId},
		"ids":          {ids},
	}

	req, err := http.NewRequest(http.MethodPost, config.SupwisdomUrl+"/eams/courseTableForStd!courseTable.action", strings.NewReader(formValues.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", config.UserAgent)

	// 发送
	resp2, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	// 读取
	content, err = io.ReadAll(resp2.Body)
	if err != nil {
		return "", err
	}

	// 检查
	temp = string(content)
	if !strings.Contains(temp, "课表格式说明") {
		return "", errors.New("获取课表失败")
	}

	fmt.Println("获取课表详情完成。")
	return temp, nil
}

// 获取当前教学周
func FetchLearnWeek(cookieJar *cookiejar.Jar) (int, error) {
	fmt.Println("\n获取当前教学周中...")

	// http 请求客户端
	client := &http.Client{
		Jar:       cookieJar,
		Transport: config.Tr,
	}

	resp, err := client.Get(config.SupwisdomUrl + "/eams/homeExt.action")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// 读取
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// 检查
	temp := string(content)
	if !strings.Contains(temp, "教学周") {
		return 0, errors.New("获取教学周失败")
	}
	temp = temp[strings.Index(temp, "id=\"teach-week\">") : strings.Index(temp, "教学周")+10]

	reg := regexp.MustCompile(`学期\s*<font size="\d+px">(\d+)<\/font>\s*教学周`)
	res := reg.FindStringSubmatch(temp)
	if len(res) < 2 {
		return 0, errors.New(temp + " 中未匹配到教学周")
	}

	fmt.Println("获取当前教学周完成。")
	return strconv.Atoi(res[1])
}
