package generate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 课程具体时间，周几第几节
type CourseTime struct {
	dayOfTheWeek int
	timeOfTheDay int
}

// 课程信息
type Course struct {
	teacherId   string
	teacherName string
	courseId    string
	courseName  string
	roomId      string
	roomName    string
	weeks       string
	courseTimes []CourseTime
}

// 作息时间表，临港上课时间
var ClassStartTime = []string{
	"082000",
	"091000",
	"101000",
	"110000",
	"115000",
	"132000",
	"141000",
	"151000",
	"160000",
	"165000",
	"181500",
	"190500",
	"195500",
}

// 作息时间表，临港下课时间
var classEndTime = []string{
	"090500",
	"095500",
	"105500",
	"114500",
	"123000",
	"140500",
	"145500",
	"155500",
	"164500",
	"173000",
	"190000",
	"195000",
	"203500",
}

// ics文件用到的星期几简称
var dayOfWeek = []string{
	"MO",
	"TU",
	"WE",
	"TH",
	"FR",
	"SA",
	"SU",
}

// 从html源码生成ics文件内容
func GenerateIcs(html string, schoolStartDay time.Time) (string, error) {
	fmt.Println("\n生成ics文件中...")
	// 利用正则匹配有效信息
	// https://regex101.com
	var myCourses []Course
	reg1 := regexp.MustCompile(`TaskActivity\("(.*)","(.*)","(.*)","(.*)\(.*\)","(.*)","(.*)","(.*)"\);((?:\s*index =\d+\*unitCount\+\d+;\s*.*\s)+)`)
	reg2 := regexp.MustCompile(`\s*index =(\d+)\*unitCount\+(\d+);\s*`)
	coursesStr := reg1.FindAllStringSubmatch(html, -1)
	for _, courseStr := range coursesStr {
		// // debug
		// for i, str := range courseStr {
		// 	fmt.Println(i, str)
		// }
		var course Course
		course.teacherId = courseStr[1]
		course.teacherName = courseStr[2]
		course.courseId = courseStr[3]
		course.courseName = courseStr[4]
		course.roomId = courseStr[5]
		course.roomName = courseStr[6]
		course.weeks = courseStr[7]
		for _, indexStr := range strings.Split(courseStr[8], "table0.activities[index][table0.activities[index].length]=activity;") {
			if !strings.Contains(indexStr, "unitCount") {
				continue
			}
			var courseTime CourseTime
			courseTime.dayOfTheWeek, _ = strconv.Atoi(reg2.FindStringSubmatch(indexStr)[1])
			courseTime.timeOfTheDay, _ = strconv.Atoi(reg2.FindStringSubmatch(indexStr)[2])
			course.courseTimes = append(course.courseTimes, courseTime)
		}
		myCourses = append(myCourses, course)
	}

	// 生成ics文件头
	var icsData string
	icsData = `BEGIN:VCALENDAR
PRODID:-//TeamSUEP//getMyCourses 20230111//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:myCourses
X-WR-TIMEZONE:Asia/Shanghai
BEGIN:VTIMEZONE
TZID:Asia/Shanghai
X-LIC-LOCATION:Asia/Shanghai
BEGIN:STANDARD
TZOFFSETFROM:+0800
TZOFFSETTO:+0800
TZNAME:CST
DTSTART:19700101T000000
END:STANDARD
END:VTIMEZONE` + "\n"

	num := 0
	for _, course := range myCourses {
		var weekDay, st, en int
		weekDay = course.courseTimes[0].dayOfTheWeek
		st = 12
		en = -1
		// 课程上下课时间
		for _, courseTime := range course.courseTimes {
			if st > courseTime.timeOfTheDay {
				st = courseTime.timeOfTheDay
			}
			if en < courseTime.timeOfTheDay {
				en = courseTime.timeOfTheDay
			}
		}

		// debug信息
		num++
		fmt.Println("")
		fmt.Println(num)
		fmt.Println(course.courseName)
		fmt.Println("周" + strconv.Itoa(weekDay+1) + " 第" + strconv.Itoa(st+1) + "-" + strconv.Itoa(en+1) + "节")

		// 统计要上课的周
		var periods []string
		var startWeek []int
		byday := dayOfWeek[weekDay]
		for i := 0; i < 53; i++ {
			if course.weeks[i] != '1' {
				continue
			}
			if i+1 >= 53 {
				startWeek = append(startWeek, i)
				periods = append(periods, "RRULE:FREQ=WEEKLY;WKST=SU;COUNT=1;INTERVAL=1;BYDAY="+byday)
				// debug信息
				fmt.Println("第" + strconv.Itoa(i) + "周")
				continue
			}
			if course.weeks[i+1] == '1' {
				// 连续周合并
				var j int
				for j = i + 1; j < 53; j++ {
					if course.weeks[j] != '1' {
						break
					}
				}
				startWeek = append(startWeek, i)
				periods = append(periods, "RRULE:FREQ=WEEKLY;WKST=SU;COUNT="+strconv.Itoa(j-i)+";INTERVAL=1;BYDAY="+byday)
				// debug信息
				fmt.Println("第" + strconv.Itoa(i) + "-" + strconv.Itoa(j-1) + "周")
				i = j - 1
			} else {
				// 单双周合并
				var j int
				for j = i + 1; j+1 < 53; j += 2 {
					if course.weeks[j] == '1' || course.weeks[j+1] == '0' {
						break
					}
				}
				startWeek = append(startWeek, i)
				periods = append(periods, "RRULE:FREQ=WEEKLY;WKST=SU;COUNT="+strconv.Itoa((j+1-i)/2)+";INTERVAL=2;BYDAY="+byday)
				// debug信息
				if i%2 == 0 {
					fmt.Printf("双")
				} else {
					fmt.Printf("单")
				}
				fmt.Println(strconv.Itoa(i) + "-" + strconv.Itoa(j-1) + "周")
				i = j - 1
			}
		}

		// 生成ics文件中的EVENT
		for i := 0; i < len(periods); i++ {
			var eventData string
			eventData = `BEGIN:VEVENT` + "\n"
			startDate := schoolStartDay.AddDate(0, 0, (startWeek[i]-1)*7+weekDay+1)

			eventData = eventData + `DTSTART;TZID=Asia/Shanghai:` + startDate.Format("20060102T") + ClassStartTime[st] + "\n"
			eventData = eventData + `DTEND;TZID=Asia/Shanghai:` + startDate.Format("20060102T") + classEndTime[en] + "\n"

			eventData = eventData + periods[i] + "\n"
			eventData = eventData + `DTSTAMP:` + time.Now().Format("20060102T150405Z") + "\n"
			// eventData = eventData + `UID:` + uuid.New().String() + "\n"
			eventData = eventData + `CREATED:` + time.Now().Format("20060102T150405Z") + "\n"
			// eventData = eventData + `DESCRIPTION:` + "\n"
			eventData = eventData + `LAST-MODIFIED:` + time.Now().Format("20060102T150405Z") + "\n"
			eventData = eventData + `LOCATION:` + course.roomName + "\n"
			eventData = eventData + `SEQUENCE:0
STATUS:CONFIRMED` + "\n"
			eventData = eventData + `SUMMARY:` + course.courseName + " " + course.teacherName + "\n"

			eventData = eventData + `TRANSP:OPAQUE
END:VEVENT` + "\n"
			icsData = icsData + eventData
		}
	}
	icsData = icsData + `END:VCALENDAR`

	fmt.Println("\n生成ics文件完成。")
	return icsData, nil
}
