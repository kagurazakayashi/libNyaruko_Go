package nyahttphandle

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/* 語言配置檔案示例 test.csv :
id,en,chs,cht  //第一行確定語言列表, id標記用可以是任意字串，隨後跟自定義語言程式碼
200,OK,成功执行请求,成功執行請求  //200 為 massageID，隨後跟每個語言的文字，逗號分隔。
100,"Hello, World!","你好，世界！","你好，世界！"  //文字中有 , 時用 "..." 包含當前文字。
101,"Hello, ""Miyabi"" !","你好，“Miyabi”！","你好，「Miyabi」！"  //文字中有 " 時用 "" 代替，並用 "..." 包含當前文字。
引號和逗號的處理遵循 CSV 檔案格式的語法。部分電子表格軟體可能儲存為非 UTF-8 編碼和非 LF 換行符，注意這些軟體儲存後可能需要轉換。
*/

var alertinfo [][]string = [][]string{}      // 提示資訊文字庫
var alertinfoLanguages []string = []string{} // 語言碼列表
var alertinfoLanguageLen int = 0             // 支援的語言數量
var alertinfoMaxID int = 0                   // 最大碼值

//alertInfoTemplateLoad 載入語言配置檔案（请先执行该函数再继续使用 AlertInfoJson(KV)(M) ）
//	`filePath` string 語言配置檔案(csv)路徑
//	資料儲存到 alertinfo ，無需重複載入
func AlertInfoTemplateLoad(filePath string) {
	FileHandle, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	var firstLine bool = true
	var configTxts [][]string = [][]string{}
	var infoLength int = 0
	for {
		lineB, _, err := lineReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				// OK
				break
			} else {
				return
			}
		}
		var line string = string(lineB)
		var lineArr []string = alertInfoSub(line)
		if firstLine {
			alertinfoLanguageLen = len(lineArr) - 1
			alertinfoLanguages = make([]string, alertinfoLanguageLen)
			for i := 0; i < alertinfoLanguageLen; i++ {
				alertinfoLanguages[i] = lineArr[i]
			}
			firstLine = false
		} else if len(lineArr)-1 == alertinfoLanguageLen {
			// TODO: 去引号
			configTxts = append(configTxts, lineArr)
			infoID, err := strconv.Atoi(lineArr[0])
			if err != nil {
				continue
			}
			if infoID > alertinfoMaxID {
				alertinfoMaxID = infoID
			}
		}
	}
	infoLength = len(configTxts)                       // 單行條目長度
	alertinfo = make([][]string, alertinfoLanguageLen) // 支援的語言陣列（建立），序號對應 alertinfoLanguages
	for i := 0; i < alertinfoLanguageLen; i++ {        // 遍歷支援的語言陣列（空白陣列寫入）
		var languageArr []string = make([]string, alertinfoMaxID+1) // 當前語言陣列（建立，長陣列）
		for j := 0; j < infoLength; j++ {                           // 遍歷單行條目陣列
			var nowline []string = configTxts[j]    // 當前單行條目陣列
			textID, err := strconv.Atoi(nowline[0]) // 資訊 ID
			if err != nil {
				continue
			}
			var thisLanguageText = nowline[i+1]    // 取當前語言的文字
			languageArr[textID] = thisLanguageText // 寫入當前語言陣列
		}
		alertinfo[i] = languageArr // 將當前語言陣列寫入總陣列
	}
	fmt.Println(len(alertinfo))
}

//AlertInfoJson: 獲取可以直接用於返回客戶端的訊息 JSON 模板
//	`languageID` int    語言 ID
//	`massageID`  int    資訊 ID
//	return       string 取出的文字
//	示例：配置檔案第一行為 `id,en,chs`, 第二行為 `200,OK,成功` 時：
//	AlertInfoJson(1, 200) -> {"code":"200","msg":"OK"}
func AlertInfoJson(languageID int, massageID int) []byte {
	return AlertInfoJsonKV(languageID, massageID, "", "")
}

//AlertInfoJsonKV: 獲取可以直接用於返回客戶端的訊息 JSON 模板，並可以附帶一個自定義鍵值
//	`languageID` int    語言 ID
//	`massageID`  int    資訊 ID
//	`key`        string 自定義鍵
//	`value`      string 自定義值
//	return       string 取出的文字
//	示例：配置檔案第一行為 `id,en,chs`, 第二行為 `200,OK,成功` 時：
//	AlertInfoJsonKV(1, 200, "token", "1145141919810") -> {"code":"1001","msg":"OK","token":"1145141919810"}
func AlertInfoJsonKV(languageID int, massageID int, key string, value string) []byte {
	var jsonMap map[string]string = alertInfoJsonGenMap(languageID, massageID)
	if value != "" {
		if key == "" {
			jsonMap["data"] = value
		} else {
			jsonMap[key] = value
		}
	}
	return alertInfoJsonGenJson(jsonMap)
}

//AlertInfoJsonKVM: 獲取可以直接用於返回客戶端的訊息 JSON 模板，並可以附帶多個自定義鍵值
//	`languageID` int 語言 ID
//	`massageID`  int 資訊 ID
//	`keyValue` map[string]string 自定義鍵值
//	return       string 取出的文字
//	示例：配置檔案第一行為 `id,en,chs`, 第二行為 `200,OK,成功` 時：
//	AlertInfoJsonKVM(1, 200, map[string]string{"token": "1145141919810", "next": "a.html"}) -> {"code":"1001","msg":"OK","token":"1145141919810","next":"a.html"}
func AlertInfoJsonKVM(languageID int, massageID int, keyValue map[string]string) []byte {
	var jsonMap map[string]string = alertInfoJsonGenMap(languageID, massageID)
	if len(keyValue) > 0 {
		for k, v := range keyValue {
			jsonMap[k] = v
		}
	}
	return alertInfoJsonGenJson(jsonMap)
}

//alertInfoJsonGenMap: 建立 JSON 模板的基本資訊
//	`languageID` int 語言 ID
//	`massageID`  int 資訊 ID
//	return map[string]string 待生成 JSON 的字典
func alertInfoJsonGenMap(languageID int, massageID int) map[string]string {
	var massageText string = alertInfoGet(languageID, massageID)
	return map[string]string{
		"code": strconv.Itoa(massageID),
		"msg":  massageText,
	}
}

//alertInfoJsonGenJson: 將待生成 JSON 的字典生成為 JSON 位元組
//	`jsonMap` map[string]string 待生成 JSON 的字典
//	return    []byte            JSON 位元組
func alertInfoJsonGenJson(jsonMap map[string]string) []byte {
	bytes, err := json.Marshal(jsonMap)
	if err != nil {
		return []byte{}
	}
	return bytes
}

//alertInfoGet: 取出資訊文本
//	`languageID` int    語言 ID
//	`massageID`  int    資訊 ID
//	return       string 取出的文字
//	示例：配置檔案第一行為 `id,en,chs`, 第二行為 `200,OK,成功` 時：
//	alertInfoGet(1, 200) -> "OK"
func alertInfoGet(languageID int, massageID int) string {
	if len(alertinfo) == 0 || languageID >= alertinfoLanguageLen || massageID > alertinfoMaxID {
		return ""
	}
	return alertinfo[languageID][massageID]
}

//alertInfoSub: 識別雙引號轉義
//	`line` string   CSV 單行字串
//	return []string 本行中每列的文字資料
//	示例: 100,"Hello, ""World""!","你好，“世界”！"
//  -> ['100', 'Hello, "World"!', '你好，“世界”！']
func alertInfoSub(line string) []string {
	var nowLineArr []string = []string{}
	line = strings.ReplaceAll(line, "\"\"", "\"")   // 轉義引號
	var lineArr []string = strings.Split(line, ",") // 逗號分隔
	var tempStr string = ""
	var writeing bool = false
	for _, unit := range lineArr {
		if writeing {
			tempStr = tempStr + "," + unit
			if unit[len(unit)-1] == '"' { // 是否以引號結尾
				writeing = false
				nowLineArr = append(nowLineArr, tempStr[0:len(tempStr)-1])
			}
		} else {
			if unit[0] == '"' { // 是否以引號開頭
				writeing = true
				tempStr = unit[1:]
			} else {
				nowLineArr = append(nowLineArr, unit)
			}
		}
	}
	return nowLineArr
}

// func test() {
// 	AlertInfoTemplateLoad("alertinfo.csv")
// 	fmt.Println(string(AlertInfoJsonKV(0, 1001, "token", "1145141919810")))
// }
