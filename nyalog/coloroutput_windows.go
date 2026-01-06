//go:build windows

package nyalog

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// kernel32 用來延遲載入 kernel32.dll
var kernel32 *syscall.LazyDLL = nil

// CmdPrint 函式用來在 Windows 平台上設定主控台的文字顏色，並輸出訊息。
// w 參數為 io.Writer 型態，代表輸出的目標，可以是標準輸出或標準錯誤輸出。
// s 參數為 interface{} 型態，代表要輸出的訊息。
// i 參數為 int 型態，代表要設定的顏色屬性（如文字顏色、背景顏色等）。
// 通過呼叫 Windows 操作系統 API SetConsoleTextAttribute 設定終端文字屬性。
// DOC: https://docs.microsoft.com/zh-cn/windows/console/setconsoletextattribute
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

	CloseHandle := kernel32.NewProc("CloseHandle")
	CloseHandle.Call(handle)
}
