package nyalog

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var (
	timeZone             *time.Location
	timeZoneDefaultName  string = "Asia/Shanghai"
	timeZoneDefaultFixed int    = 8
)

// <颜色配置>
type ConsoleColor int8

const (
	ConsoleColorBlack  ConsoleColor = 0
	ConsoleColorRed    ConsoleColor = 1
	ConsoleColorGreen  ConsoleColor = 2
	ConsoleColorYellow ConsoleColor = 3
	ConsoleColorBlue   ConsoleColor = 4
	ConsoleColorPurple ConsoleColor = 5
	ConsoleColorCyan   ConsoleColor = 6
	ConsoleColorWhte   ConsoleColor = 7
)

type LogLevel int8

const (
	LogLevelDebug   LogLevel = 0
	LogLevelInfo    LogLevel = 1
	LogLevelWarning LogLevel = 2
	LogLevelError   LogLevel = 3
	LogLevelClash   LogLevel = 4
	LogLevelOK      LogLevel = 5
)

//將 ConsoleColor 物件轉換為顏色字串
func (p ConsoleColor) String() string {
	consoleColorString := [8]string{"black", "red", "green", "yellow", "blue", "purple", "cyan", "whte"}
	return consoleColorString[p]
}

//LogLevelData: 根據日誌記錄級別來確定輸出顏色
//	`lvl`  LogLevel     日誌記錄級別
//	return ConsoleColor 顏色
func LogLevelData(lvl LogLevel) ConsoleColor {
	switch lvl {
	case LogLevelDebug:
		return ConsoleColorWhte
	case LogLevelInfo:
		return ConsoleColorCyan
	case LogLevelWarning:
		return ConsoleColorYellow
	case LogLevelError:
		return ConsoleColorRed
	case LogLevelClash:
		return ConsoleColorRed
	case LogLevelOK:
		return ConsoleColorGreen
	default:
		return ConsoleColorCyan
	}
}

// </颜色配置>

// 向終端輸出日誌
//	`a` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func Log(a ...interface{}) {
	if timeZone == nil {
		timeZone = GetTimeZone("", -100)
	}
	var parameterLength = len(a)
	var strArr []string = make([]string, parameterLength)
	for i, b := range a { // interface{}
		strArr[i] = ToString(b)
	}
	fmt.Println(strArr)
}

//logF: 獲取當前函式名稱
//	return string 當前函式名稱
func FuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

//toString: 將各種資料型別轉換為字串
//	`v`    interface{} 字串、位元組、數字，以及包含上述值的陣列
//	return string      zhuanhuanh
func ToString(v interface{}) string {
	var s string = ""
	if v == nil {
		return s
	}
	switch v.(type) {
	case string:
		return v.(string)
	case float64:
		s = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case float32:
		s = strconv.FormatFloat(float64(v.(float32)), 'f', -1, 64)
	case int:
		s = strconv.Itoa(v.(int))
	case uint:
		s = strconv.Itoa(int(v.(uint)))
	case int8:
		s = strconv.Itoa(int(v.(int8)))
	case uint8:
		s = strconv.Itoa(int(v.(uint8)))
	case int16:
		s = strconv.Itoa(int(v.(int16)))
	case uint16:
		s = strconv.Itoa(int(v.(uint16)))
	case int32:
		s = strconv.Itoa(int(v.(int32)))
	case uint32:
		s = strconv.Itoa(int(v.(uint32)))
	case int64:
		s = strconv.FormatInt(v.(int64), 10)
	case uint64:
		s = strconv.FormatUint(v.(uint64), 10)
	case []byte:
		s = string(v.([]byte))
	default:
		newValue, _ := json.Marshal(v)
		s = string(newValue)
	}
	return s
}

//GetTimeZone: 設定和獲取當前時區。
//	`zone`     string 時區字串，如 "Asia/Shanghai" 。提供空字串則採用變數 timeZoneDefaultName 。
//	`fixedZone` int    根據 CST 加減小時數補償，範圍 -12 ~ 12 ，提供超範圍數值則採用變數 timeZoneDefaultFixed 。該項只在從系統獲取 zone 失敗後有效。
//	return *time.Location 時區物件
func GetTimeZone(zone string, fixedZone int) *time.Location {
	if zone == "" {
		zone = timeZoneDefaultName
	}
	if fixedZone < -12 || fixedZone > 12 {
		fixedZone = timeZoneDefaultFixed
	}
	timeZoneN, err := time.LoadLocation(zone)
	if err != nil {
		timeZoneN = time.FixedZone("CST", fixedZone*3600)
	}
	return timeZoneN
}

//timeStamp2timeString: 從時間戳獲取當前時間字串
//	`timestamp` int64  納秒時間戳，如果為 0 則提供當前時間字串
//	return      string 時間字串，格式 `yyyy-MM-dd HH:mm:ss`
func timeStamp2timeString(timestamp int64) string {
	var timeObj time.Time = time.Now()
	if timestamp > 0 {
		timeObj = time.Unix(0, timestamp)
	}
	timeObj = timeObj.In(timeZone)
	var timeStr string = timeObj.Format("2006-01-02 15:04:05")
	return timeStr
}
