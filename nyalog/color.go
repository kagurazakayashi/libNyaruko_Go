// 終端輸出顏色控制
package nyalog

type ConsoleColor int8

const (
	Black  ConsoleColor = 0
	Red    ConsoleColor = 1
	Green  ConsoleColor = 2
	Yellow ConsoleColor = 3
	Blue   ConsoleColor = 4
	Purple ConsoleColor = 5
	Cyan   ConsoleColor = 6
	White  ConsoleColor = 7
)

// 將 ConsoleColor 物件轉換為顏色字串
func (p ConsoleColor) String() string {
	var consoleColorString [8]string = [8]string{"black", "red", "green", "yellow", "blue", "purple", "cyan", "white"}
	return consoleColorString[p]
}

// 將 LogLevel 物件轉換為單字母
func (p LogLevel) String() string {
	var cogLevelChar [7]string = [7]string{"D", "I", "O", "W", "E", "X", ""}
	return cogLevelChar[p]
}

// LogLevelData: 根據日誌記錄級別來確定輸出顏色
//
//	`lvl`  LogLevel     日誌記錄級別
//	return  顏色
func LogLevelData(lvl LogLevel) ConsoleColor {
	switch lvl {
	case Debug: //0
		return White
	case Info: //1
		return Cyan
	case OK: //2
		return Green
	case Warning: //3
		return Yellow
	case Error: //4
		return Red
	case Clash: //5
		return Red
	case None: //6
		return Black
	default:
		return Cyan
	}
}
