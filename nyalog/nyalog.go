// 日誌輸出和記錄
package nyalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	timeZone *time.Location
)

type LogLevel int8

const (
	Debug   LogLevel = 0
	Info    LogLevel = 1
	OK      LogLevel = 2
	Warning LogLevel = 3
	Error   LogLevel = 4
	Clash   LogLevel = 5
	None    LogLevel = 6
)

// Log: 向終端輸出日誌
//
//	`setLevel` LogLevel 設定的日誌等級 (0-6)
//	`nowLevel` LogLevel 當前輸出的日誌等級 (0-5)
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
//	日誌输出示例: "[E 2022-03-11 15:04:05 main.main:18] ERROR"
func Log(setLevel LogLevel, nowLevel LogLevel, obj ...interface{}) {
	if nowLevel < setLevel {
		return
	}
	fmt.Fprintln(os.Stderr, logString(nowLevel, obj))
}

// LogD: 用於除錯時快速找到臨時性輸出，以紫色底色輸出
//
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func LogD(obj ...interface{}) {
	Colorful.WithFrontColor(White.String()).WithBackColor(Purple.String()).Println(strings.Join(interfaceArray2StringArray(obj), " "))
}

// LogF: 向終端輸出日誌，並將日誌內容寫入到檔案，路徑為 `當前執行檔案.log` 。
//
//	`setLevel` LogLevel 設定的日誌等級 (0-6)
//	`nowLevel` LogLevel 當前輸出的日誌等級 (0-5)
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func LogF(setLevel LogLevel, nowLevel LogLevel, obj ...interface{}) {
	if nowLevel < setLevel {
		return
	}
	LogFF(setLevel, nowLevel, "", obj...)
}

// LogFF: 向終端輸出日誌，並將日誌內容寫入到指定自定義檔案。
//
//	`setLevel` LogLevel 設定的日誌等級 (0-6)
//	`nowLevel` LogLevel 當前輸出的日誌等級 (0-5)
//	`path`  string   日誌檔案路徑
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func LogFF(setLevel LogLevel, nowLevel LogLevel, path string, obj ...interface{}) {
	if nowLevel < setLevel {
		return
	}
	var logStr string = logString(nowLevel, obj)
	var logPath string = path
	if len(path) == 0 {
		fmt.Fprintln(os.Stderr, logStr)
		file, err := exec.LookPath(os.Args[0])
		if err != nil {
			return
		}
		logPath, err = filepath.Abs(file)
		if err != nil {
			return
		}
		logPath += ".log"
	}
	fd, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	buf := []byte(logStr)
	fd.Write(buf)
	fd.Close()
}

// LogC: 向終端輸出日誌，根據日誌等級自動決定輸出顏色
//
//	`setLevel` LogLevel 設定的日誌等級 (0-6)
//	`nowLevel` LogLevel 當前輸出的日誌等級 (0-5)
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func LogC(setLevel LogLevel, nowLevel LogLevel, obj ...interface{}) {
	if nowLevel < setLevel {
		return
	}
	var colorStr string = LogLevelData(nowLevel).String()
	var logStr string = logString(nowLevel, obj)
	Colorful.WithFrontColor(colorStr).Fprintln(os.Stderr, logStr)
}

// LogCC: 向終端輸出日誌，並指定輸出顏色
//
//	`setLevel` LogLevel 設定的日誌等級 (0-6)
//	`nowLevel` LogLevel 當前輸出的日誌等級 (0-5)
//	`color` ConsoleColor 文字顏色
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func LogCC(setLevel LogLevel, nowLevel LogLevel, color ConsoleColor, obj ...interface{}) {
	if nowLevel < setLevel {
		return
	}
	var logStr string = logString(nowLevel, obj)
	Colorful.WithFrontColor(color.String()).Fprintln(os.Stderr, logStr)
}

// LogCCC: 向終端輸出日誌，並指定輸出前景顏色和背景顏色
//
//	`setLevel` LogLevel 設定的日誌等級 (0-6)
//	`nowLevel` LogLevel 當前輸出的日誌等級 (0-5)
//	`color`           ConsoleColor 文字顏色
//	`backgroundColor` ConsoleColor 文字底色
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
func LogCCC(setLevel LogLevel, nowLevel LogLevel, color ConsoleColor, backgroundColor ConsoleColor, obj ...interface{}) {
	if nowLevel < setLevel {
		return
	}
	var logStr string = logString(nowLevel, obj)
	Colorful.WithFrontColor(color.String()).WithBackColor(backgroundColor.String()).Fprintln(os.Stderr, logStr)
}

// logString 将输入的参数组装成字符串
//
//	`level` LogLevel 日誌等級 (0-5)
//	`obj` ...interface{} 要輸出的變數（會自動嘗試轉換成字串）
//	return string    準備輸出的字串
func logString(level LogLevel, obj []interface{}) string {
	var prefix string = logPrefix(level)
	var infoArr []string = interfaceArray2StringArray(obj)
	return prefix + strings.Join(infoArr, " ")
}

// prefix: 日誌輸出字首
//
//	`level` LogLevel 日誌等級 (0-5)
//	return  []string 字首字串單元陣列
//	[日誌等級字元, 日期時間]
func logPrefix(level LogLevel) string {
	if timeZone == nil {
		timeZoneN, err := GetTimeZone("", -100)
		if err != nil {
			timeZone = time.UTC
		} else {
			timeZone = timeZoneN
		}
	}
	var s [2]string = [2]string{"[", "]"}
	var ExArr []string = []string{
		s[0] + level.String() + s[1],
		s[0] + timeStamp2timeString(0) + s[1],
		// s[0] + logFuncInfo() + s[1],
	}
	return strings.Join(ExArr, "") + " "
}

// interfaceArray2StringArray: 將泛型別陣列轉換為字串陣列
//
//	`objs` []interface{} 泛型別陣列
//	return []string      字串陣列
func interfaceArray2StringArray(objs []interface{}) []string {
	var parameterLength = len(objs)
	if parameterLength == 0 {
		return []string{}
	}
	var strArr []string = make([]string, parameterLength)
	for i, o := range objs { // interface{}
		strArr[i] = ToString(o)
	}
	return strArr
}

// logFuncInfo: 獲取當前程式碼行號、函式名稱
//
//	return string 函式名稱:行號
func FuncInfo() string {
	pc, _, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	f := runtime.FuncForPC(pc)
	return f.Name() + ":" + strconv.Itoa(line)
}

// toString: 將各種資料型別轉換為字串
//
//	`v`    interface{} 字串、位元組、數字，以及包含上述值的陣列
//	return string      zhuanhuanh
func ToString(v interface{}) string {
	var s string = ""
	if v == nil {
		return s
	}
	switch t := v.(type) {
	case string:
		s = t
	case float64:
		s = strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		s = strconv.FormatFloat(float64(t), 'f', -1, 64)
	case int:
		s = strconv.Itoa(t)
	case uint:
		s = strconv.Itoa(int(t))
	case int8:
		s = strconv.Itoa(int(t))
	case uint8:
		s = strconv.Itoa(int(t))
	case int16:
		s = strconv.Itoa(int(t))
	case uint16:
		s = strconv.Itoa(int(t))
	case int32:
		s = strconv.Itoa(int(t))
	case uint32:
		s = strconv.Itoa(int(t))
	case int64:
		s = strconv.FormatInt(t, 10)
	case uint64:
		s = strconv.FormatUint(t, 10)
	case []byte:
		s = string(t)
	default:
		newValue, _ := json.Marshal(t)
		s = string(newValue)
		s = formatJSON(s)
	}
	return s
}

// JSON: 格式化 JSON ，最佳化輸出的易讀性
//
//	`data` string 源 JSON 字符串
//	return string 美化后的 JSON 字符串
func formatJSON(data string) string {
	var str bytes.Buffer
	var err error = json.Indent(&str, []byte(data), "", "    ")
	if err != nil {
		return ""
	}
	return str.String()
}

// GetTimeZone: 設定和獲取當前時區。
//
//	优先判断`zone`是否为""
//	`zone`     string 時區字串，如 "Asia/Shanghai" 。提供空字串則採用變數 timeZoneDefaultName 。
//	`fixedZone` int    根據 CST 加減小時數補償，範圍 -12 ~ 12 ，提供超範圍數值則採用變數 timeZoneDefaultFixed 。該項只在從系統獲取 zone 失敗後有效。
//	return *time.Location 時區物件
func GetTimeZone(zone string, fixedZone int) (timeZoneN *time.Location, err error) {
	if zone == "" {
		if fixedZone < -12 || fixedZone > 12 {
			return nil, fmt.Errorf("E")
		} else {
			timeZoneN = time.FixedZone("CST", fixedZone*3600)
		}
	} else {
		timeZoneN, err = time.LoadLocation(zone)
		if err != nil {
			return nil, err
		}
	}
	return timeZoneN, nil
}

// timeStamp2timeString: 從時間戳獲取當前時間字串
//
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
