package nyacrypt

import (
	"math/rand"
	"time"
)

//RandomNumber: 建立隨機數
//	`min`  int 最小值
//	`max`  int 最大值
//	return int 生成的隨機數
func RandomNumber(min int, max int) int {
	if max == min || max < min {
		return max
	}
	var timeStamp int64 = time.Now().UnixNano()
	if min > 0 {
		max += min
	}
	rand.Seed(timeStamp)
	var num int = rand.Intn(max)
	if min > 0 {
		num -= min
	}
	if num < min || num > max {
		return RandomNumber(min, max)
	}
	return num
}

//RandomString: 建立隨機字串
//	`length`  int    要生成的字串長度
//	`charDic` string 要包含的字串型別
//	0:所有數字 a:所有小寫字母 A:所有大寫字母 @:符號
//	例如: `0aA`: 包含所有數字、所有小寫字母、所有大寫字母
//	return    string 建立的隨機字串
func RandomString(length int, charDic string) string {
	var asciiLib []string = []string{}
	for _, c := range charDic {
		switch c {
		case '0':
			for i := 48; i <= 57; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
		case 'a':
			for i := 97; i <= 122; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
		case 'A':
			for i := 65; i <= 90; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
		case '@':
			for i := 33; i <= 47; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
			for i := 58; i <= 64; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
			for i := 91; i <= 96; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
			for i := 123; i <= 126; i++ {
				asciiLib = append(asciiLib, string(rune(i)))
			}
		}
	}
	var str string = ""
	for i := 0; i < length; i++ {
		var selectCharIArr [10]int = [10]int{0}
		for j := 0; j < 10; j++ {
			selectCharIArr[j] = RandomNumber(0, len(asciiLib)-1)
		}
		var selectCharI int = RandomNumber(0, 10)
		str += asciiLib[selectCharIArr[selectCharI]]
	}
	return str
}
