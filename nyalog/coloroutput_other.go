//go:build !windows

package nyalog

import "io"

// CmdPrint 在非 Windows 平台上為空實作，不會被實際呼叫（因為 isWindows() 會返回 false）。
func CmdPrint(w io.Writer, s interface{}, i int) {
}
