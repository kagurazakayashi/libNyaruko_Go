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

// Bash 前景色
const (
	FrontBlack  = iota + 30 // 30黑色
	FrontRed                // 31紅色
	FrontGreen              // 32綠色
	FrontYellow             // 33黃色
	FrontBlue               // 34藍色
	FrontPurple             // 35紫紅色
	FrontCyan               // 36青藍色
	FrontWhite              // 37白色
)

// Bash 背景色
const (
	BackBlack  = iota + 40 // 40黑色
	BackRed                // 41紅色
	BackGreen              // 42綠色
	BackYellow             // 43黃色
	BackBlue               // 44藍色
	BackPurple             // 45紫紅色
	BackCyan               // 46青藍色
	BackWhite              // 47白色
)

// Bash 顯示模式
const (
	ModeDefault   = 0 // 終端預設設定
	ModeHighLight = 1 // 高亮顯示
	ModeLine      = 4 // 使用下劃線
	ModeFlash     = 5 // 閃爍
	ModeReWhite   = 6 // 反白顯示
	ModeHidden    = 7 // 不可見
)

// CMD
const (
	CmdBlack       = iota // 0黑色
	CmdBlue               // 1藍色
	CmdGreen              // 2綠色
	CmdAqua               // 3淺綠色
	CmdRed                // 4红色
	CmdPurple             // 5紫色
	CmdYellow             // 6黃色
	CmdWhite              // 7白色
	CmdGray               // 8灰色
	CmdLightBlue          // 9淡藍色
	CmdLightGreen         // A淡綠色
	CmdLightAqua          // B淡淺綠色
	CmdLightRed           // C淡紅色
	CmdLightPurple        // D淡紫色
	CmdLightYellow        // E淡黃色
	CmdBrightWhite        // F亮白色
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
