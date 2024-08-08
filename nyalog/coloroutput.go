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

var (
	modeArr                   = []int{0, 1, 4, 5, 6, 7}
	kernel32 *syscall.LazyDLL = nil
)

type ColorOutput struct {
	frontColor int
	backColor  int
	mode       int
}

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

func isWindows() bool {
	return runtime.GOOS == "windows"
}

var Colorful ColorOutput = ColorOutput{frontColor: CmdGreen, backColor: CmdBlack, mode: ModeDefault}

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

// 其中0x1B是标记，[开始定义颜色，依次为：模式，背景色，前景色，0代表恢复默认颜色。
func (c ColorOutput) Println(str interface{}) {
	if isWindows() {
		// 背景色 | 前景色
		// 注意，简单的或操作是错误的，比如 4 | 2，实际是 6 即 黄色，和预期的红底绿字不一致。
		// 应该构成1个8位的二进制，前四位是背景色，后四位是前景色，因此背景色需要左移4位。
		CmdPrint(os.Stdout, str, (c.backColor<<4)|c.frontColor)
	} else {
		fmt.Printf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, c.mode, c.backColor, c.frontColor, str, 0x1B)
	}
}

func (c ColorOutput) Fprintln(w io.Writer, str interface{}) {
	if isWindows() {
		CmdPrint(w, str, (c.backColor<<4)|c.frontColor)
	} else {
		fmt.Fprintf(w, "%c[%d;%d;%dm%s%c[0m\n", 0x1B, c.mode, c.backColor, c.frontColor, str, 0x1B)
	}
}

func (c ColorOutput) WithFrontColor(color string) ColorOutput {
	color = strings.ToLower(color)
	c.frontColor = useColor(color, true)
	return c
}

func (c ColorOutput) WithBackColor(color string) ColorOutput {
	color = strings.ToLower(color)
	c.backColor = useColor(color, false)
	return c
}

func (c ColorOutput) WithMode(mode int) ColorOutput {
	a := garray.NewIntArrayFrom(modeArr, true)
	bo := a.Contains(mode)
	if bo {
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

func CmdPrint(w io.Writer, s interface{}, i int) {
	if kernel32 == nil {
		kernel32 = syscall.NewLazyDLL("kernel32.dll")
	}
	proc := kernel32.NewProc("SetConsoleTextAttribute")
	var handle uintptr
	if w == os.Stdout {
		handle, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(i))
	} else {
		handle, _, _ = proc.Call(uintptr(syscall.Stderr), uintptr(i))
	}
	fmt.Fprintf(w, "%v\n", s)
	// fmt.Println(s)
	// handle, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(7))
	CloseHandle := kernel32.NewProc("CloseHandle")
	CloseHandle.Call(handle)
}

func ResetColor(w io.Writer) {
	if isWindows() {
		CmdPrint(w, "", CmdWhite)
	} else {
		fmt.Fprintf(w, "%c[0m", 0x1B)
	}
}
