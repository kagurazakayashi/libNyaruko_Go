package nyatools

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type NyaFileInfo struct {
	name    string    // 檔案的名字（不含副檔名）
	size    int64     // 普通檔案返回值表示其大小，特殊檔案的返值含義各系統不同
	mode    string    // 檔案的屬性
	modTime time.Time // 檔案的修改時間
	isDir   bool      // 是否為資料夾
}

//NyaReadFileLineHandler: FileReadLine 的代理方法
//	`lineText` string 當前行的文字內容
//	`lineNum`  uint   當前行號
//	`isEnd`    bool   本條是否為檔案的末尾
//	return     error  可能遇到的錯誤
type NyaReadFileLineHandler func(lineText string, lineNum uint, isEnd bool, err error)

func SetNyaReadFileLineHandler(handler NyaReadFileLineHandler) {
	readFileLineHandler = handler
}

var (
	readFileLineHandler NyaReadFileLineHandler
)

//ReadFile: 讀取文字檔案
//	`filePath` string 檔案路徑
//	return     string 讀取到的內容
//	return     error  可能遇到的錯誤
func FileRead(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

//ReadFile: 逐行讀取文字檔案（適用於大量行大檔案）
//	需要实现 NyaReadFileLineHandler 代理方法接收
//	`filePath` string 檔案路徑
//	return     error  可能遇到的錯誤
func FileReadLine(filePath string) error {
	FileHandle, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)
	var lineNum uint = 0
	for {
		lineB, _, err := lineReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				if readFileLineHandler != nil {
					readFileLineHandler(string(lineB), lineNum, true, nil)
				}
				break
			} else {
				if readFileLineHandler != nil {
					readFileLineHandler("", lineNum, false, err)
				}
				return err
			}
		}
		if readFileLineHandler != nil {
			readFileLineHandler(string(lineB), lineNum, false, nil)
		}
		lineNum++
	}
	return nil
}

//FileExist: 檔案或資料夾是否存在
//	`filePath` string 檔案路徑
//	return     int8   判斷結果( >0 即存在):
//	-1 發生錯誤
//	 0 不存在
//	 1 存在，是檔案
//	 2 存在，是資料夾
//	 3 存在，不知道是檔案還是資料夾
//	return     error  可能遇到的錯誤
func FileExist(filePath string) (int8, error) {
	fileInfo, err := FileInfo(filePath)
	if err != nil {
		if os.IsExist(err) {
			return 3, nil
		} else if os.IsNotExist(err) {
			return 0, nil
		} else {
			return -1, err
		}
	}
	if fileInfo.isDir {
		return 2, nil
	} else {
		return 1, nil
	}
}

//FileInfo: 獲取檔案資訊
//	`filePath` string      檔案路徑
//	return     NyaFileInfo 檔案資訊
//	return     error       可能遇到的錯誤
func FileInfo(filePath string) (NyaFileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	var fInfo NyaFileInfo = NyaFileInfo{}
	if fileInfo != nil {
		fInfo = NyaFileInfo{
			name:    fileInfo.Name(),
			size:    fileInfo.Size(),
			mode:    fileInfo.Mode().String(),
			modTime: fileInfo.ModTime(),
			isDir:   fileInfo.IsDir(),
		}
	}
	if err != nil {
		return fInfo, err
	}
	return fInfo, nil
}

//FileCopy: 複製檔案
//	`srcPath` string 原始檔案路徑
//	`dstPath` string 目標檔案路徑
//	return    int64  寫入的資料量
//	return    error  可能遇到的錯誤
func FileCopy(srcPath string, dstPath string) (written int64, err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

//FolderCreate: 新建資料夾
//	`dirPath` string 新建文件夹路径
//	return    bool   是否实际创建（若为 false 且无 error ，则文件夹已经存在）
//	return    error  可能遇到的错误
func FolderCreate(dirPath string) (bool, error) {
	_, err := os.Stat(dirPath)
	if err == nil {
		return false, nil
	}
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return false, err
	}
	return true, nil
}
