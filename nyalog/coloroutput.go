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

// Linux -------------------------

// 前景 背景 颜色
// -------------------------
// 30  40  黑色
// 31  41  红色
// 32  42  绿色
// 33  43  黄色
// 34  44  蓝色
// 35  45  紫红色
// 36  46  青蓝色
// 37  47  白色
//
// 代码 意义
// -------------------------
//  0  终端默认设置
//  1  高亮显示
//  4  使用下划线
//  5  闪烁
//  7  反白显示
//  8  不可见

// Windows -------------------------
// cmd下查看颜色编号： color /?
// 0 = Black       8 = Gray
// 1 = Blue        9 = Light Blue
// 2 = Green       A = Light Green
// 3 = Aqua        B = Light Aqua
// 4 = Red         C = Light Red
// 5 = Purple      D = Light Purple
// 6 = Yellow      E = Light Yellow
// 7 = White       F = Bright White

const (
	FrontBlack = iota + 30
	FrontRed
	FrontGreen
	FrontYellow
	FrontBlue
	FrontPurple
	FrontCyan
	FrontWhite
)

const (
	BackBlack = iota + 40
	BackRed
	BackGreen
	BackYellow
	BackBlue
	BackPurple
	BackCyan
	BackWhite
)

const (
	ModeDefault   = 0
	ModeHighLight = 1
	ModeLine      = 4
	ModeFlash     = 5
	ModeReWhite   = 6
	ModeHidden    = 7
)

/*

windows cmd下查看颜色编号： color /?

Sets the default console foreground and background colors.

COLOR [attr]

  attr        Specifies color attribute of console output

Color attributes are specified by TWO hex digits -- the first
corresponds to the background; the second the foreground.  Each digit
can be any of the following values:

    0 = Black       8 = Gray
    1 = Blue        9 = Light Blue
    2 = Green       A = Light Green
    3 = Aqua        B = Light Aqua
    4 = Red         C = Light Red
    5 = Purple      D = Light Purple
    6 = Yellow      E = Light Yellow
    7 = White       F = Bright White

If no argument is given, this command restores the color to what it was
when CMD.EXE started.  This value either comes from the current console
window, the /T command line switch or from the DefaultColor registry
value.

The COLOR command sets ERRORLEVEL to 1 if an attempt is made to execute
the COLOR command with a foreground and background color that are the
same.

Example: "COLOR fc" produces light red on bright white

*/

const (
	CmdBlack  = 0
	CmdRed    = 4
	CmdGreen  = 2
	CmdYellow = 6
	CmdBlue   = 1
	CmdPurple = 5
	CmdCyan   = 3
	CmdWhite  = 7
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

var Colorful ColorOutput
var frontMap map[string]int
var backMap map[string]int

func init() {
	Colorful = ColorOutput{frontColor: CmdGreen, backColor: CmdBlack, mode: ModeDefault}

	if runtime.GOOS != "windows" {
		frontMap = make(map[string]int)
		frontMap["black"] = FrontBlack
		frontMap["red"] = FrontRed
		frontMap["green"] = FrontGreen
		frontMap["yellow"] = FrontYellow
		frontMap["blue"] = FrontBlue
		frontMap["purple"] = FrontPurple
		frontMap["cyan"] = FrontCyan
		frontMap["white"] = FrontWhite

		backMap = make(map[string]int)
		backMap["black"] = BackBlack
		backMap["red"] = BackRed
		backMap["green"] = BackGreen
		backMap["yellow"] = BackYellow
		backMap["blue"] = BackBlue
		backMap["purple"] = BackPurple
		backMap["cyan"] = BackCyan
		backMap["white"] = BackWhite
	} else {
		frontMap = make(map[string]int)
		frontMap["black"] = CmdBlack
		frontMap["red"] = CmdRed
		frontMap["green"] = CmdGreen
		frontMap["yellow"] = CmdYellow
		frontMap["blue"] = CmdBlue
		frontMap["purple"] = CmdPurple
		frontMap["cyan"] = CmdCyan
		frontMap["white"] = CmdWhite

		backMap = make(map[string]int)
		backMap["black"] = CmdBlack
		backMap["red"] = CmdRed
		backMap["green"] = CmdGreen
		backMap["yellow"] = CmdYellow
		backMap["blue"] = CmdBlue
		backMap["purple"] = CmdPurple
		backMap["cyan"] = CmdCyan
		backMap["white"] = CmdWhite
	}
}

// 其中0x1B是标记，[开始定义颜色，依次为：模式，背景色，前景色，0代表恢复默认颜色。
func (c ColorOutput) Println(str interface{}) {
	if runtime.GOOS != "windows" {
		fmt.Printf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, c.mode, c.backColor, c.frontColor, str, 0x1B)
	} else {
		// 背景色 | 前景色
		// 注意，简单的或操作是错误的，比如 4 | 2，实际是 6 即 黄色，和预期的红底绿字不一致。
		// 应该构成1个8位的二进制，前四位是背景色，后四位是前景色，因此背景色需要左移4位。
		CmdPrint(os.Stdout, str, (c.backColor<<4)|c.frontColor)
	}
}

func (c ColorOutput) Fprintln(w io.Writer, str interface{}) {
	if runtime.GOOS != "windows" {
		fmt.Fprintf(w, "%c[%d;%d;%dm%s%c[0m\n", 0x1B, c.mode, c.backColor, c.frontColor, str, 0x1B)
	} else {
		CmdPrint(w, str, (c.backColor<<4)|c.frontColor)
	}
}

func (c ColorOutput) WithFrontColor(color string) ColorOutput {
	color = strings.ToLower(color)
	co, ok := frontMap[color]
	if ok {
		c.frontColor = co
	}
	return c
}

func (c ColorOutput) WithBackColor(color string) ColorOutput {
	color = strings.ToLower(color)
	co, ok := backMap[color]
	if ok {
		c.backColor = co
	}

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
//
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
