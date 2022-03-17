// 雜湊計算
package nyacrypt

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math"
	"os"
)

//MD5String: 計算 MD5 字串雜湊
//	`data` string 要進行雜湊的源字串
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	示例: MD5String("Hello, World!") -> (32) 65A8E27D8879283831B664BD8B7F0AD4
func MD5String(data string, key string) string {
	return hashEncodeString(5, data, key)
}

//MD5FilePath: 計算檔案 MD5 雜湊（基於檔案路徑）
//	`filePath` string 要進行雜湊檔案路徑
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾；大檔案讀取模式下，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func MD5FilePath(filePath string, key string, chunk float64) (string, error) {
	return hashEncodeFilePath(5, filePath, chunk, key)
}

//MD5File: 計算小型檔案 MD5 雜湊（基於 io.Reader 檔案物件）
//	`file` io.Reader 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾
func MD5File(file io.Reader, key string) (string, error) {
	return hashEncodeFile(5, file, nil, 0, key)
}

//MD5FileBig: 計算大型檔案 MD5 雜湊（基於 *os.File 檔案物件）
//	`file` *os.File 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：大檔案讀取模式下(chunk!=0)，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func MD5FileBig(file *os.File, key string, chunk float64) (string, error) {
	return hashEncodeFile(5, nil, file, chunk, key)
}

//SHA1String: 計算 SHA1 字串雜湊
//	`data` string 要進行雜湊的源字串
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: SHA1String("Hello, World!") -> (40) 0a0a9f2a6772942557ab5355d76af442f8f65e01
func SHA1String(data string, key string) string {
	return hashEncodeString(1, data, key)
}

//SHA1FilePath: 計算檔案 SHA1 雜湊（基於檔案路徑）
//	`filePath` string 要進行雜湊檔案路徑
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾；大檔案讀取模式下，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func SHA1FilePath(filePath string, key string, chunk float64) (string, error) {
	return hashEncodeFilePath(1, filePath, chunk, key)
}

//SHA1File: 計算小型檔案 SHA1 雜湊（基於 io.Reader 檔案物件）
//	`file` io.Reader 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾
func SHA1File(file io.Reader, key string) (string, error) {
	return hashEncodeFile(1, file, nil, 0, key)
}

//SHA1FileBig: 計算大型檔案 SHA1 雜湊（基於 *os.File 檔案物件）
//	`file` *os.File 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：大檔案讀取模式下(chunk!=0)，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func SHA1FileBig(file *os.File, key string, chunk float64) (string, error) {
	return hashEncodeFile(1, nil, file, chunk, key)
}

//SHA256String: 計算 SHA256 字串雜湊
//	`data` string 要進行雜湊的源字串
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: SHA256String("Hello, World!") -> (64) dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f
func SHA256String(data string, key string) string {
	return hashEncodeString(256, data, key)
}

//SHA256FilePath: 計算檔案 SHA256 雜湊（基於檔案路徑）
//	`filePath` string 要進行雜湊檔案路徑
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾；大檔案讀取模式下，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func SHA256FilePath(filePath string, key string, chunk float64) (string, error) {
	return hashEncodeFilePath(256, filePath, chunk, key)
}

//SHA256File: 計算小型檔案 SHA256 雜湊（基於 io.Reader 檔案物件）
//	`file` io.Reader 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾
func SHA256File(file io.Reader, key string) (string, error) {
	return hashEncodeFile(256, file, nil, 0, key)
}

//SHA256FileBig: 計算大型檔案 SHA256 雜湊（基於 *os.File 檔案物件）
//	`file` *os.File 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：大檔案讀取模式下(chunk!=0)，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func SHA256FileBig(file *os.File, key string, chunk float64) (string, error) {
	return hashEncodeFile(256, nil, file, chunk, key)
}

//SHA512String: 計算 SHA512 字串雜湊
//	`data` string 要進行雜湊的源字串
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: SHA512String("Hello, World!") -> (128) 374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6cc69291e0fa2fe0006a52570ef18c19def4e617c33ce52ef0a6e5fbe318cb0387
func SHA512String(data string, key string) string {
	return hashEncodeString(512, data, key)
}

//SHA512FilePath: 計算檔案 SHA512 雜湊（基於檔案路徑）
//	`data` string 要進行雜湊的源字串
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾；大檔案讀取模式下，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func SHA512FilePath(filePath string, key string, chunk float64) (string, error) {
	return hashEncodeFilePath(512, filePath, chunk, key)
}

//SHA512File: 計算小型檔案 SHA512 雜湊（基於 io.Reader 檔案物件）
//	`file` io.Reader 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾
func SHA512File(file io.Reader, key string) (string, error) {
	return hashEncodeFile(512, file, nil, 0, key)
}

//SHA512FileBig: 計算大型檔案 SHA512 雜湊（基於 *os.File 檔案物件）
//	`file` *os.File 檔案物件
//	`key`  string 使用 HMAC 演算法帶 Key 運算，空字串為不使用
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：大檔案讀取模式下(chunk!=0)，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func SHA512FileBig(file *os.File, key string, chunk float64) (string, error) {
	return hashEncodeFile(512, nil, file, chunk, key)
}

//hashEncode: 計算雜湊編碼
//	`mode` int16  雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`dataByte` []byte 要進行雜湊的資料
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
func hashEncode(mode int16, dataByte []byte, key string) (string, error) {
	var dataByteN []byte = dataByte
	if len(key) == 0 {
		var h hash.Hash = hashMode(mode)()
		_, err := h.Write(dataByteN)
		if err != nil {
			return "", err
		}
		dataByteN = []byte(nil)
		dataByteN = h.Sum(dataByteN)
		var hashStr string = fmt.Sprintf("%x", dataByteN)
		return hashStr, nil
	} else { // HMAC
		var keyByte []byte = []byte(key)
		var hmac hash.Hash = hmac.New(hashMode(mode), keyByte)
		_, err := hmac.Write(dataByteN)
		if err != nil {
			return "", err
		}
		dataByteN = []byte(nil)
		dataByteN = hmac.Sum(dataByteN)
		return hex.EncodeToString(dataByteN), nil
	}
}

//hashEncodeString: 計算字串雜湊
//	`mode` int16  雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`data` string 要進行雜湊的源字串
//	return string 雜湊之後的字串，如果失敗則返回空字串。
func hashEncodeString(mode int16, data string, key string) string {
	var dataByte []byte = []byte(data)
	hashStr, err := hashEncode(mode, dataByte, key)
	if err != nil {
		return ""
	}
	return hashStr
}

//hashEncodeFilePath: 計算檔案雜湊（基於檔案路徑）
//	`mode` int16 雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`filePath` string 檔案路徑
func hashEncodeFilePath(mode int16, filePath string, chunk float64, key string) (string, error) {
	if chunk == 0 {
		return hashEncodeSmallFilePath(mode, filePath, key)
	} else {
		return hashEncodeBigFilePath(mode, filePath, chunk)
	}
}

//hashEncodeFile: 計算檔案雜湊（基於檔案物件）
//	`mode` int16 雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`smallFile` io.Reader 小型檔案物件
//	`bigFile`   *os.File  大型檔案物件
//	`chunk` float64 分塊讀取的每塊資料量(B)。其中 0 為不分塊(整體讀入), -1 為使用預設分塊大小 4KB 。
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意: smallFile 和 bigFile 選擇其一輸入，另一個傳 nil 。
//	注意: 如果 chunk 為 0 , 必須使用 smallFile , 否則必須使用 bigFile 。
//	注意：大檔案讀取模式下(chunk!=0)，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func hashEncodeFile(mode int16, smallFile io.Reader, bigFile *os.File, chunk float64, key string) (string, error) {
	if chunk == 0 && smallFile != nil {
		return hashEncodeSmallFile(mode, smallFile, key)
	} else if bigFile != nil {
		return hashEncodeBigFile(mode, bigFile, chunk)
	}
	return "", fmt.Errorf("no file obj")
}

//hashEncodeSmallFilePath: 計算小型檔案雜湊（基於檔案路徑）
//	`mode` int16 雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`filePath` string 檔案路徑
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
func hashEncodeSmallFilePath(mode int16, filePath string, key string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	hashStr, err := hashEncodeSmallFile(mode, file, key)
	err2 := file.Close()
	if err != nil {
		return "", err
	}
	if err2 != nil {
		return "", err2
	}
	return hashStr, nil
}

//hashEncodeSmallFile: 計算小型檔案雜湊（基於檔案物件）
//	`mode` int16 雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`file` io.Reader 檔案物件
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：小檔案讀取模式下，會從檔案當前指標位置開始讀取，計算後文件指標在末尾
func hashEncodeSmallFile(mode int16, file io.Reader, key string) (string, error) {
	dataByte, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return hashEncode(mode, dataByte, key)
}

//hashEncodeBigFilePath: 計算大型檔案雜湊（基於檔案路徑）
//	`mode` int16 雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`filePath` string 檔案路徑
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
func hashEncodeBigFilePath(mode int16, filePath string, chunk float64) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	hashStr, err := hashEncodeBigFile(mode, file, chunk)
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}
	return hashStr, nil
}

//hashEncodeBigFile: 計算大型檔案雜湊（基於檔案物件）
//	`mode` int16 雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`file` io.Reader 檔案物件
//	return string 雜湊之後的字串，如果失敗則返回空字串
//	return error  可能發生的錯誤
//	注意：大檔案讀取模式下，會將檔案指標移到開頭從頭讀取，計算後文件指標在開頭。
func hashEncodeBigFile(mode int16, file *os.File, chunk float64) (string, error) {
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	file.Seek(0, 0)
	var filesize int64 = info.Size()
	if chunk <= 0 {
		chunk = 4096
	}
	var blocks uint64 = uint64(math.Ceil(float64(filesize) / float64(chunk)))
	var h hash.Hash = hashMode(mode)()
	for i := uint64(0); i < blocks; i++ {
		var startByte int64 = int64(float64(i) * chunk)
		var needByte float64 = float64(filesize - startByte)
		if needByte > chunk {
			needByte = chunk
		}
		var blocksize int = int(needByte)
		buf := make([]byte, blocksize)
		_, err = file.Read(buf)
		if err != nil {
			return "", err
		}
		_, err = io.WriteString(h, string(buf))
		if err != nil {
			return "", err
		}
	}
	file.Seek(0, 0)
	var dataByte []byte = []byte(nil)
	dataByte = h.Sum(dataByte)
	return hex.EncodeToString(dataByte), nil
}

//hashMode: 指定雜湊演算法
//	`mode` int16  雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	return func() hash.Hash 雜湊函式
func hashMode(mode int16) func() hash.Hash {
	switch mode {
	case 1:
		return sha1.New
	case 5:
		return md5.New
	case 256:
		return sha256.New
	case 512:
		return sha512.New
	}
	return nil
}
