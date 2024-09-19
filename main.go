package main

import (
	"fmt"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"

	"github.com/TeamSUEP/getMyCourses/config"
	"github.com/TeamSUEP/getMyCourses/fetch"
	"github.com/TeamSUEP/getMyCourses/generate"
	"github.com/TeamSUEP/getMyCourses/login"
)

func main() {
	// 选择登录方式
	var choice int
	var err error
	fmt.Printf("1.统一身份认证登录: %s\n", config.IdsUrl)
	fmt.Printf("2.树维教务系统登录: %s/eams/localLogin.action\n", config.SupwisdomUrl)
	fmt.Print("\n请选择登录方式（1 或 2）：")
	fmt.Scanln(&choice)
	if choice != 1 && choice != 2 {
		fmt.Println("输入错误，请重新运行程序。")
		return
	}

	// 获取帐号和密码
	var username, password string
	fmt.Print("帐号: ")
	fmt.Scanln(&username)
	fmt.Print("密码: ")
	fmt.Scanln(&password)

	// 登录
	var cookieJar *cookiejar.Jar
	if choice == 1 {
		cookieJar, err = login.LoginViaIds(username, password, config.IdsService)
	} else if choice == 2 {
		cookieJar, err = login.LoginViaSupwisdom(username, password)
	} else {
		return
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 获取当前学期
	semesterId, projectId, err := fetch.GetCurrentSemester(cookieJar)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 获取学期列表
	err = fetch.PrintSemesterList(cookieJar)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 选择学期
	var temp string
	fmt.Printf("\n请输入学期ID（默认为当前学期 %s）：", semesterId)
	fmt.Scanln(&temp)
	if temp != "" {
		semesterId = temp
	}

	// 获取包含课程表的html源码
	html, err := fetch.FetchCourses(cookieJar, semesterId, projectId)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 获取当前教学周
	learnWeek, err := fetch.FetchLearnWeek(cookieJar)
	if err != nil {
		if err.Error() == "获取教学周失败" {
			fmt.Printf("获取教学周失败，请手动计算输入当前教学周：")
			fmt.Scanln(&learnWeek)
		} else {
			fmt.Println(err.Error())
			return
		}
	}

	// 计算校历第一周周日
	now := time.Now()
	location := time.FixedZone("UTC+8", 8*60*60)
	daySum := int(now.Weekday()) + learnWeek*7 - 7
	schoolStartDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -daySum)

	fmt.Println("\n当前为第", learnWeek, "教学周。")
	fmt.Println("计算得到本学期开始于", schoolStartDay.Format("2006-01-02"))

	// 从html源码生成ics文件内容
	ics, err := generate.GenerateIcs(html, schoolStartDay)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 保存到文件
	err = os.WriteFile("myCourses.ics", []byte(ics), 0644)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 提示文件路径
	path, err := filepath.Abs("myCourses.ics")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("\n已保存为：", path)
}
