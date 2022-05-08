// 時間相關轉換功能

package nyatime

import (
	"strconv"
	"time"
)

//TimeZone: 載入時區
//	`defaultZone` int 如果時區載入失敗，則使用此值，直接按小時差載入時區（例如 8）。
//	return *time.Location 時區
//	return error          可能的错误信息
func TimeZone(defaultZone int) (*time.Location, error) {
	timeZoneN, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		timeZoneN = time.FixedZone("CST", defaultZone*3600)
	}
	return timeZoneN, err
}

//timeStamp2timeString: 納秒時間戳轉換為時間字串
//	`timestamp` int64          納秒時間戳
//	`timeZone`  *time.Location 時區
//	return      string         時間字串
func TimeStamp2timeString(timestamp int64, timeZone *time.Location) string {
	var timeObj time.Time = time.Now()
	if timestamp > 0 {
		timeObj = time.Unix(0, timestamp)
	}
	timeObj = timeObj.In(timeZone)
	var timeStr string = timeObj.Format("2006-01-02 15:04:05")
	return timeStr
}

//TimeEapsedString: 獲取經過時間字串
//	`startTime` time.Time 起始時間
//	return      string    經過時間字串（格式： `00:00:00:00` ，對應 `天:時:分:秒` ）
func TimeEapsedString(startTime time.Time) string {
	// startTime = time.Now()
	var runTime time.Duration = time.Since(startTime)
	var m = ":"
	day, hour, minute, second := SecondsToUnit(int(runTime.Seconds()))
	return strconv.Itoa(day) + m + TimeIntAddZero(hour) + m + TimeIntAddZero(minute) + m + TimeIntAddZero(second)
}

//SecondsToUnit: 將 秒 轉換為 [天,時,分,秒]
//	`seconds` int 秒
//	return    int 天
//	return    int 時
//	return    int 分
//	return    int 秒
func SecondsToUnit(seconds int) (int, int, int, int) {
	var SecondsPerMinute int = 60
	var SecondsPerHour int = SecondsPerMinute * 60 // 3600
	var SecondsPerDay int = SecondsPerHour * 24    // 86400
	var day = seconds / SecondsPerDay
	var hour int = (seconds - day*SecondsPerDay) / SecondsPerHour
	var minute int = (seconds - day*SecondsPerDay - hour*SecondsPerHour) / SecondsPerMinute
	var second int = seconds - day*SecondsPerDay - hour*SecondsPerHour - minute*SecondsPerMinute
	return day, hour, minute, second
}

//TimeIntAddZero: 為時間數字不夠兩位的自動補0並轉為字串（輸入 int）
//	`t`    int    時間數字
//	return string 兩位時間字串
func TimeIntAddZero(t int) string {
	if t < 10 {
		return "0" + strconv.Itoa(t)
	} else {
		return strconv.Itoa(t)
	}
}

//TimeIntAddZero: 為時間數字不夠兩位的自動補0並轉為字串（輸入 float64）
//	`t`    float64 時間數字
//	return string  兩位時間字串
func TimeF64AddZero(t float64) string {
	return TimeIntAddZero(int(t))
}
