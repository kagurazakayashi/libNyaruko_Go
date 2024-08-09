/*
修改自 github.com/phprao/ColorOutput
*/
package nyalog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/gogf/gf/container/garray"
)

// 宣告一個整數陣列 modeArr，包含六個元素，代表不同的模式
var (
	modeArr = []int{0, 1, 4, 5, 6, 7} // 模式陣列

	// 宣告一個 LazyDLL 型態的變數 kernel32，並初始化為 nil，這是用來延遲載入 kernel32.dll
	kernel32 *syscall.LazyDLL = nil
)

// 定義一個結構體 ColorOutput，用來表示輸出的顏色和模式
type ColorOutput struct {
	frontColor int // 前景顏色
	backColor  int // 背景顏色
	mode       int // 模式
}

// colorfulWindows 函式根據輸入的顏色字串，返回一個包含前景色、背景色和模式的 ColorOutput 結構。
// 此函式專門用於 Windows 系統中的顏色配置。
// 參數:
//   - color: 字串類型，代表所需的顏色（如 "black", "red", "green" 等）。
//
// 返回值:
//   - ColorOutput 結構，包含前景色、背景色和模式的設定。
func colorfulWindows(color string) ColorOutput {
	switch color {
	case "black":
		return ColorOutput{frontColor: CmdBlack, backColor: CmdBlack, mode: ModeDefault}
	case "red":
		return ColorOutput{frontColor: CmdRed, backColor: CmdBlack, mode: ModeDefault}
	case "green":
		return ColorOutput{frontColor: CmdGreen, backColor: CmdBlack, mode: ModeDefault}
	case "yellow":
		return ColorOutput{frontColor: CmdYellow, backColor: CmdBlack, mode: ModeDefault}
	case "blue":
		return ColorOutput{frontColor: CmdBlue, backColor: CmdBlack, mode: ModeDefault}
	case "purple":
		return ColorOutput{frontColor: CmdPurple, backColor: CmdBlack, mode: ModeDefault}
	case "cyan":
		return ColorOutput{frontColor: CmdAqua, backColor: CmdBlack, mode: ModeDefault}
	case "white":
		return ColorOutput{frontColor: CmdWhite, backColor: CmdBlack, mode: ModeDefault}
	default:
		return ColorOutput{frontColor: CmdGreen, backColor: CmdBlack, mode: ModeDefault}
	}
}

// colorfulLinux 函式根據輸入的顏色字串，返回一個包含前景色、背景色和模式的 ColorOutput 結構。
// 此函式專門用於 Linux 系統中的顏色配置。
// 參數:
//   - color: 字串類型，代表所需的顏色（如 "black", "red", "green" 等）。
//
// 返回值:
//   - ColorOutput 結構，包含前景色、背景色和模式的設定。
func colorfulLinux(color string) ColorOutput {
	switch color {
	case "black":
		return ColorOutput{frontColor: FrontBlack, backColor: BackBlack, mode: ModeDefault}
	case "red":
		return ColorOutput{frontColor: FrontRed, backColor: BackBlack, mode: ModeDefault}
	case "green":
		return ColorOutput{frontColor: FrontGreen, backColor: BackBlack, mode: ModeDefault}
	case "yellow":
		return ColorOutput{frontColor: FrontYellow, backColor: BackBlack, mode: ModeDefault}
	case "blue":
		return ColorOutput{frontColor: FrontBlue, backColor: BackBlack, mode: ModeDefault}
	case "purple":
		return ColorOutput{frontColor: FrontPurple, backColor: BackBlack, mode: ModeDefault}
	case "cyan":
		return ColorOutput{frontColor: FrontCyan, backColor: BackBlack, mode: ModeDefault}
	case "white":
		return ColorOutput{frontColor: FrontWhite, backColor: BackBlack, mode: ModeDefault}
	default:
		return ColorOutput{frontColor: FrontGreen, backColor: BackBlack, mode: ModeDefault}
	}
}

// isWindows 函式用於檢查當前的操作系統是否為 Windows。
// 返回值:
//   - bool 類型，若操作系統為 Windows 則返回 true，否則返回 false。
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// Colorful 變數是一個 ColorOutput 結構的實例，其預設前景色為綠色 (CmdGreen)，背景色為黑色 (CmdBlack)，模式為默認模式 (ModeDefault)。
// 此變數可用於設定終端的顏色輸出。
var Colorful ColorOutput = ColorOutput{frontColor: CmdGreen, backColor: CmdBlack, mode: ModeDefault}

// useColor 函式根據輸入的顏色字串與前景/背景標誌，返回相應的顏色值。
// 該函式會根據當前的操作系統選擇適當的顏色配置函式 (colorfulWindows 或 colorfulLinux)。
// 參數:
//   - color: 字串類型，代表所需的顏色（如 "black", "red", "green" 等）。
//   - isFrontColor: 布林值，若為 true 則返回前景色，否則返回背景色。
//
// 返回值:
//   - int 類型，對應的前景或背景顏色的整數值。
func useColor(color string, isFrontColor bool) int {
	if isWindows() {
		var co ColorOutput = colorfulWindows(color)
		if isFrontColor {
			return co.frontColor
		} else {
			return co.backColor
		}
	} else {
		var co ColorOutput = colorfulLinux(color)
		if isFrontColor {
			return co.frontColor
		} else {
			return co.backColor
		}
	}
}

// ColorOutput 是一個結構，用於控制文字輸出的顏色。
// Println 方法將文本輸出到標準輸出，並根據操作系統選擇適當的顏色格式。
// 若在 Windows 系統上，則使用 CmdPrint 函數來處理顏色輸出。
// 若在其他操作系統上，則使用 ANSI 色碼進行顏色控制。
func (c ColorOutput) Println(str interface{}) {
	if isWindows() {
		// 在 Windows 系統上，組合背景色和前景色的顏色碼。
		// 背景色需要左移4位來構成8位的二進制，前四位表示背景色，後四位表示前景色。
		CmdPrint(os.Stdout, str, (c.backColor<<4)|c.frontColor)
	} else {
		// 在其他操作系統上，使用 ANSI 色碼來設定顏色。
		// 0x1B 是 ESC 字元，後面的數字分別是模式、背景色、前景色。
		// %c[0m 用於重置顏色設置。
		fmt.Printf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, c.mode, c.backColor, c.frontColor, str, 0x1B)
	}
}

// Fprintln 方法與 Println 類似，但輸出至指定的 io.Writer。
// 根據操作系統，選擇適當的顏色格式輸出。
func (c ColorOutput) Fprintln(w io.Writer, str interface{}) {
	if isWindows() {
		// 在 Windows 系統上，使用 CmdPrint 進行顏色輸出。
		CmdPrint(w, str, (c.backColor<<4)|c.frontColor)
	} else {
		// 在其他操作系統上，使用 ANSI 色碼來設定顏色並輸出到指定的 io.Writer。
		fmt.Fprintf(w, "%c[%d;%d;%dm%s%c[0m\n", 0x1B, c.mode, c.backColor, c.frontColor, str, 0x1B)
	}
}

// WithFrontColor 設定前景色並返回 ColorOutput 結構。
// 接受一個表示顏色的字串，並將其轉換為小寫。
func (c ColorOutput) WithFrontColor(color string) ColorOutput {
	color = strings.ToLower(color)
	// 將顏色字串轉換為顏色代碼並賦值給 frontColor 屬性。
	c.frontColor = useColor(color, true)
	return c
}

// WithBackColor 設定背景色並返回 ColorOutput 結構。
// 接受一個表示顏色的字串，並將其轉換為小寫。
func (c ColorOutput) WithBackColor(color string) ColorOutput {
	color = strings.ToLower(color)
	// 將顏色字串轉換為顏色代碼並賦值給 backColor 屬性。
	c.backColor = useColor(color, false)
	return c
}

// WithMode 設定顯示模式並返回 ColorOutput 結構。
// 接受一個整數值表示模式，並檢查該模式是否有效。
func (c ColorOutput) WithMode(mode int) ColorOutput {
	// 創建一個新的整數數組來存放模式，並檢查模式是否存在於數組中。
	a := garray.NewIntArrayFrom(modeArr, true)
	bo := a.Contains(mode)
	if bo {
		// 若模式有效，則設置 mode 屬性。
		c.mode = mode
	}
	return c
}

// Windows

// 通过调用windows操作系统API设置终端文本属性，包括：前景色，背景色，高亮。可同时设置多个属性，使用竖线 | 隔开。
// DOC: https://docs.microsoft.com/zh-cn/windows/console/setconsoletextattribute
// Usage: https://docs.microsoft.com/zh-cn/windows/console/using-the-high-level-input-and-output-functions
// 属性值: https://docs.microsoft.com/zh-cn/windows/console/console-screen-buffers#character-attributes
// SetConsoleTextAttribute函数用于设置显示后续写入文本的颜色。在退出之前，程序会还原原始控制台输入模式和颜色属性。但是微软官方
// 建议使用“虚拟终端”来实现终端控制，而且是跨平台的。
// 建议使用基于windows提供的“虚拟终端序列”来实现兼容多平台的终端控制，比如：https://github.com/gookit/color
// https://docs.microsoft.com/zh-cn/windows/console/console-virtual-terminal-sequences
// https://docs.microsoft.com/zh-cn/windows/console/console-virtual-terminal-sequences#samples

// CmdPrint 函式用來在 Windows 平台上設定主控台的文字顏色，並輸出訊息。
// w 參數為 io.Writer 型態，代表輸出的目標，可以是標準輸出或標準錯誤輸出。
// s 參數為 interface{} 型態，代表要輸出的訊息。
// i 參數為 int 型態，代表要設定的顏色屬性（如文字顏色、背景顏色等）。
func CmdPrint(w io.Writer, s interface{}, i int) {
	// 檢查 kernel32 是否已經被初始化，如果還沒有就初始化它。
	if kernel32 == nil {
		kernel32 = syscall.NewLazyDLL("kernel32.dll") // 載入 kernel32.dll 動態連結庫
	}

	// 取得 SetConsoleTextAttribute 函式的位址，該函式用於設定主控台文字屬性。
	proc := kernel32.NewProc("SetConsoleTextAttribute")

	var handle uintptr
	// 根據輸出目標是標準輸出還是標準錯誤輸出來設定文字顏色。
	if w == os.Stdout {
		handle, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(i))
	} else {
		handle, _, _ = proc.Call(uintptr(syscall.Stderr), uintptr(i))
	}

	// 輸出訊息到指定的 Writer（如標準輸出或標準錯誤輸出）。
	fmt.Fprintf(w, "%v\n", s)

	// 初始化 CloseHandle 函式，用於關閉打開的句柄（handle）。
	CloseHandle := kernel32.NewProc("CloseHandle")
	// 關閉先前開啟的句柄。
	CloseHandle.Call(handle)
}

// ResetColor 函式用來重置主控台的文字顏色。
// w 參數為 io.Writer 型態，代表輸出的目標，可以是標準輸出或標準錯誤輸出。
func ResetColor(w io.Writer) {
	if isWindows() { // 檢查是否為 Windows 平台
		CmdPrint(w, "", CmdWhite) // 在 Windows 上重置顏色
	} else {
		// 在其他平台上使用 ANSI 轉義序列重置顏色
		fmt.Fprintf(w, "%c[0m", 0x1B)
	}
}
