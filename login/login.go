package login

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/TeamSUEP/getMyCourses/config"
	"github.com/antchfx/htmlquery"
)

// 树维教务系统登录：/eams/localLogin.action
func LoginViaSupwisdom(username string, password string) (*cookiejar.Jar, error) {
	fmt.Println("\n树维教务系统登录中...")

	// Cookie 自动维护
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// http 请求客户端
	client := &http.Client{
		Jar:       cookieJar,
		Transport: config.Tr,
	}

	// 获取表单
	resp1, err := client.Get(config.SupwisdomUrl + "/eams/localLogin.action")
	if err != nil {
		return nil, err
	}
	defer resp1.Body.Close()

	// 读取
	content, err := io.ReadAll(resp1.Body)
	if err != nil {
		return nil, err
	}

	// 检查
	temp := string(content)
	if !strings.Contains(temp, "CryptoJS.SHA1(") {
		return nil, errors.New("登录页面打开失败，请检查。")
	}

	// 对密码进行SHA1哈希
	temp = temp[strings.Index(temp, "CryptoJS.SHA1(")+15 : strings.Index(temp, "CryptoJS.SHA1(")+52]
	password = temp + password
	bytes := sha1.Sum([]byte(password))
	password = hex.EncodeToString(bytes[:])

	// 获取验证码
	resp2, err := client.Get(config.SupwisdomUrl + "/eams/captcha/image.action")
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	// 读取
	content, err = io.ReadAll(resp2.Body)
	if err != nil {
		return nil, err
	}

	// 写入文件
	err = os.WriteFile("captcha.jpg", content, 0644)
	if err != nil {
		return nil, err
	}

	// 人工输入验证码
	path, err := filepath.Abs("captcha.jpg")
	if err != nil {
		return nil, err
	}
	fmt.Print("请打开 " + path + " 查看并输入验证码：")
	var captcha string
	fmt.Scanln(&captcha)

	// 登录
	formValues := make(url.Values)
	formValues.Set("username", username)
	formValues.Set("password", password)
	formValues.Set("session_locale", "zh_CN")
	formValues.Set("captcha_response", captcha)
	formValues.Set("encodedPassword", "")
	formValues.Set("hashCookie", strings.TrimSuffix(temp, "-"))

	req, err := http.NewRequest(http.MethodPost, config.SupwisdomUrl+"/eams/localLogin.action", strings.NewReader(formValues.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", config.UserAgent)

	// 发送
	resp3, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp3.Body.Close()

	// 检查
	succeed, err := checkLogin(cookieJar)
	if err != nil {
		return nil, err
	}
	if !succeed {
		// 读取
		content, err := io.ReadAll(resp3.Body)
		if err != nil {
			return nil, err
		}
		// 提取错误信息
		if strings.Contains(string(content), "WrongCaptcha") {
			return nil, errors.New("验证码错误。")
		}
		s, err := htmlquery.Parse(strings.NewReader(string(content)))
		if err != nil {
			return nil, err
		}
		msg, err := htmlquery.QueryAll(s, "//title")
		if err != nil {
			return nil, err
		}
		if len(msg) > 0 {
			return nil, errors.New(htmlquery.InnerText(msg[0]))
		}
		return nil, errors.New("登录失败。")
	}

	fmt.Println("树维教务系统登录完成。")
	return cookieJar, nil
}

// 统一身份认证
// 人生苦短，为什么不去看看 Python 呢？
// https://github.com/TeamSUEP/SUEP-course-elect/blob/main/ids.py
func LoginViaIds(username string, password string, service string) (*cookiejar.Jar, error) {
	fmt.Println("\n统一身份认证登录中...")

	// Cookie 自动维护
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// http 请求客户端
	client := &http.Client{
		Jar:       cookieJar,
		Transport: config.Tr,
	}

	// 获取表单
	req, err := http.NewRequest(http.MethodGet, config.IdsUrl+"/authserver/login", nil)
	if err != nil {
		return nil, err
	}

	// 添加参数
	q := req.URL.Query()
	q.Add("service", service)
	req.URL.RawQuery = q.Encode()

	// 发送
	resp1, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp1.Body.Close()

	// 读取
	content, err := io.ReadAll(resp1.Body)
	if err != nil {
		return nil, err
	}

	// 检查
	temp := string(content)
	if !strings.Contains(temp, "<form") {
		return nil, errors.New("登录页面打开失败，请检查 " + config.IdsUrl)
	}

	// 提取表单信息
	s, err := htmlquery.Parse(strings.NewReader(temp))
	if err != nil {
		return nil, err
	}
	form, err := htmlquery.QueryAll(s, "//form//input")
	if err != nil {
		return nil, err
	}
	formValues := make(url.Values)
	for _, v := range form {
		if htmlquery.SelectAttr(v, "name") != "" && htmlquery.SelectAttr(v, "value") != "" {
			formValues.Set(htmlquery.SelectAttr(v, "name"), htmlquery.SelectAttr(v, "value"))
		}
	}
	formValues.Set("username", username)
	formValues.Set("password", password)

	// 登录
	req, err = http.NewRequest(http.MethodPost, config.IdsUrl+"/authserver/login", strings.NewReader(formValues.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", config.UserAgent)

	// 添加参数
	q = req.URL.Query()
	q.Add("service", service)
	req.URL.RawQuery = q.Encode()

	// 发送
	resp2, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	// 检查
	succeed, err := checkLogin(cookieJar)
	if err != nil {
		return nil, err
	}
	if !succeed {
		// 读取
		content, err := io.ReadAll(resp2.Body)
		if err != nil {
			return nil, err
		}
		// 提取错误信息
		s, err := htmlquery.Parse(strings.NewReader(string(content)))
		if err != nil {
			return nil, err
		}
		msg, err := htmlquery.QueryAll(s, "//*[@id='msg']")
		if err != nil {
			return nil, err
		}
		if len(msg) > 0 {
			return nil, errors.New(htmlquery.InnerText(msg[0]))
		}
		return nil, errors.New("登录失败。")
	}

	fmt.Println("统一身份登录认证登录完成。")
	return cookieJar, nil
}

func checkLogin(cookieJar *cookiejar.Jar) (bool, error) {
	// http 请求客户端
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: config.Tr,
		Jar:       cookieJar,
	}

	// 请求
	resp, err := client.Get(config.SupwisdomUrl + "/eams/home.action")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// 检查
	if resp.StatusCode == 200 {
		return true, nil
	}

	return false, nil
}
