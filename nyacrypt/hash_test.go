// 測試：雜湊相關函式的工作情況
package nyacrypt

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

//TestString: 字串雜湊測試
func TestString(t *testing.T) {
	println("===== String Test =====")
	var s string = "Hello, World!"
	var k string = "123456"
	println("String", s)
	println("MD5String", MD5String(s, ""))
	println("SHA1String", SHA1String(s, ""))
	println("SHA256String", SHA256String(s, ""))
	println("SHA512String", SHA512String(s, ""))
	println("String", s, "HMAC", k)
	println("MD5String-HMAC", MD5String(s, k))
	println("SHA1String-HMAC", SHA1String(s, k))
	println("SHA256String-HMAC", SHA256String(s, k))
	println("SHA512String-HMAC", SHA512String(s, k))
}

//TestFile: 檔案雜湊測試
func TestFile(t *testing.T) {
	println("===== File Test =====")
	var path string = "hash_test.go"
	println("path", path)
	file, err := os.Open(path)
	if err != nil {
		println(err.Error())
		return
	}

	println("MD5FilePath")
	str, err := MD5FilePath(path, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("SHA1FilePath")
	str, err = SHA1FilePath(path, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("SHA256FilePath")
	str, err = SHA256FilePath(path, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("SHA512FilePath")
	str, err = SHA512FilePath(path, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("MD5FileBig")
	str, err = MD5FileBig(file, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("SHA1FileBig")
	str, err = SHA1FileBig(file, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("SHA256FileBig")
	str, err = SHA256FileBig(file, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	println("SHA512FileBig")
	str, err = SHA512FileBig(file, "", -1)
	if err != nil {
		println(err.Error())
	} else {
		println(str)
	}

	file.Close()
}
