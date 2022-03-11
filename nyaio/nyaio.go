package nyatools

import (
	"bufio"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type NyaFileInfo struct {
	name    string    // 檔案的名字（不含副檔名）
	size    int64     // 普通檔案返回值表示其大小，特殊檔案的返值含義各系統不同
	mode    string    // 檔案的屬性
	modTime time.Time // 檔案的修改時間
	isDir   bool      // 是否為資料夾
}

// <代理方法>

//NyaReadFileLineHandler: FileReadLine 的代理方法
//	`lineText` string 當前行的文字內容
//	`lineNum`  uint   當前行號
//	`isEnd`    bool   本條是否為檔案的末尾
//	return     error  可能遇到的錯誤
type NyaReadFileLineHandler func(lineText string, lineNum uint, isEnd bool, err error)

func SetNyaReadFileLineHandler(handler NyaReadFileLineHandler) {
	readFileLineHandler = handler
}

// </代理方法>

// <可選配置>
type Option struct {
	permission uint32 // 建立檔案/資料夾的許可權。預設按 UMASK 022，預設目錄許可權為755，預設檔案許可權為644。
}
type OptionConfig func(*Option)

func Option_permission(v uint32) OptionConfig {
	return func(p *Option) {
		p.permission = v
	}
}

// </可選配置>

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
//  `options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`permission` uint32 新建檔案的許可權，預設 644
//	return    int64  寫入的資料量
//	return    error  可能遇到的錯誤
func FileCopy(srcPath string, dstPath string, options ...OptionConfig) (written int64, err error) {
	option := &Option{permission: 644}
	for _, o := range options {
		o(option)
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, fs.FileMode(option.permission))
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

//FolderCreate: 新建資料夾
//	`dirPath` string 新建文件夹路径
//  `options` ...OptionConfig 可選配置，執行 `Option_*` 函式輸入
//		`permission` uint32 新建資料夾的許可權，預設 755
//	return    bool   是否实际创建（若为 false 且无 error ，则文件夹已经存在）
//	return    error  可能遇到的错误
func FolderCreate(dirPath string, options ...OptionConfig) (bool, error) {
	option := &Option{permission: 755}
	for _, o := range options {
		o(option)
	}
	_, err := os.Stat(dirPath)
	if err == nil {
		return false, nil
	}
	err = os.MkdirAll(dirPath, fs.FileMode(option.permission))
	if err != nil {
		return false, err
	}
	return true, nil
}

//FileList: 獲取指定資料夾下的所有檔案
//	`dirPth`    string 要搜尋的資料夾路徑
//	`recursive` bool   是否需要遍歷子資料夾
//	`suffix`    string 關鍵詞，只列出包含指定關鍵詞的檔名（英文不區分大小寫）
func FileList(dirPth string, recursive bool, suffix string) ([]string, error) {
	var files []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	// PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) // 忽略大小寫
	for _, fi := range dir {
		if fi.IsDir() {
			if recursive {
				subFiles, err := FileList(dirPth, recursive, suffix)
				if err != nil {
					return nil, err
				} else {
					files = append(files, subFiles...)
				}
			} else {
				continue
			}
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			// files = append(files, dirPth+PthSep+fi.Name())
			files = append(files, dirPth+"/"+fi.Name())
		}
	}
	return files, nil
}

//ListDir: 獲取指定目錄下的所有資料夾
func ListDir(dirPth string) ([]string, error) {
	var files []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	// PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() {
			// files = append(files, dirPth+PthSep+fi.Name())
			files = append(files, dirPth+"/"+fi.Name())
		}
	}
	return files, nil
}
