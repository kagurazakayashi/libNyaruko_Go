// 終端輸出顏色控制
package nyalog

type ConsoleColor int8

const (
	ConsoleColorBlack  ConsoleColor = 0
	ConsoleColorRed    ConsoleColor = 1
	ConsoleColorGreen  ConsoleColor = 2
	ConsoleColorYellow ConsoleColor = 3
	ConsoleColorBlue   ConsoleColor = 4
	ConsoleColorPurple ConsoleColor = 5
	ConsoleColorCyan   ConsoleColor = 6
	ConsoleColorWhite  ConsoleColor = 7
)

//將 ConsoleColor 物件轉換為顏色字串
func (p ConsoleColor) String() string {
	var consoleColorString [8]string = [8]string{"black", "red", "green", "yellow", "blue", "purple", "cyan", "white"}
	return consoleColorString[p]
}

//將 LogLevel 物件轉換為單字母
func (p LogLevel) String() string {
	var cogLevelChar [6]string = [6]string{"D", "I", "W", "E", "X", "O"}
	return cogLevelChar[p]
}

//LogLevelData: 根據日誌記錄級別來確定輸出顏色
//	`lvl`  LogLevel     日誌記錄級別
//	return  顏色
func LogLevelData(lvl LogLevel) ConsoleColor {
	switch lvl {
	case LogLevelDebug:
		return ConsoleColorWhite
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
