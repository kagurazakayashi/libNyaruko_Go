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
	"io/ioutil"
	"os"
)

//MD5 編碼字串（雜湊）
//	`data` string 要進行雜湊的源字串
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: MD5("Hello, World!") -> (32) 65A8E27D8879283831B664BD8B7F0AD4
func MD5(data string) string {
	return hashEncode(5, data)
}

//计算文件MD5 編碼字串（雜湊）
//	`filepath` string 要進行雜湊的文件地址
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	return error 错误信息
func MD5forFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	md5 := fmt.Sprintf("%x", md5.Sum(body))
	// runtime.GC()
	f.Close()
	return md5, nil
}

//SHA1 編碼字串（雜湊）
//	`data` string 要進行雜湊的源字串
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: SHA1("Hello, World!") -> (40) 0a0a9f2a6772942557ab5355d76af442f8f65e01
func SHA1(data string) string {
	return hashEncode(1, data)
}

//SHA256 編碼字串（雜湊）
//	`data` string 要進行雜湊的源字串
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: SHA256("Hello, World!") -> (64) dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f
func SHA256(data string) string {
	return hashEncode(256, data)
}

//SHA512 編碼字串（雜湊）
//	`data` string 要進行雜湊的源字串
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: SHA512("Hello, World!") -> (128) 374d794a95cdcfd8b35993185fef9ba368f160d8daf432d08ba9f1ed1e5abe6cc69291e0fa2fe0006a52570ef18c19def4e617c33ce52ef0a6e5fbe318cb0387
func SHA512(data string) string {
	return hashEncode(512, data)
}

//HMAC 編碼字串（雜湊）
//	`mode` int16  基礎雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`data` string 要進行雜湊的源字串
//	`key`  string 雜湊所需的金鑰
//	return string 雜湊之後的字串。如果失敗則返回空字串。
//	示例: HMAC(5, 123456, "Hello, World!") -> (32) 999a53e606cb7d95681e2004db63ef77
func HMAC(mode int16, key string, data string) string {
	var keyByte []byte = []byte(key)
	var dataByte []byte = []byte(data)
	var hmac hash.Hash = hmac.New(hashMode(mode), keyByte)
	_, err := hmac.Write(dataByte)
	if err != nil {
		return ""
	}
	dataByte = []byte(nil)
	dataByte = hmac.Sum(dataByte)
	return hex.EncodeToString(dataByte)
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

//hashEncode: 雜湊編碼
//	`mode` int16  雜湊模式 1:SHA1, 5:MD5, 256:SHA256, 512:SHA512
//	`data` string 要進行雜湊的源字串
//	return string 雜湊之後的字串，如果失敗則返回空字串。
func hashEncode(mode int16, data string) string {
	var h hash.Hash = hashMode(mode)()
	var dataByte []byte = []byte(data)
	_, err := h.Write(dataByte)
	if err != nil {
		return ""
	}
	dataByte = []byte(nil)
	dataByte = h.Sum(dataByte)
	return hex.EncodeToString(dataByte)
}
