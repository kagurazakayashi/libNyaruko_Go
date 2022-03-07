package nyatools

import (
	"fmt"
	"io/ioutil"
	"os"
)

func readFile(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("[读取文件失败]:" + err.Error())
	}
	return string(bytes), nil
}

//创建文件夹
func CreateDir(dir string) (bool, error) {
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return false, fmt.Errorf("[创建文件夹失败]:" + err.Error())
	}
	return true, nil
}
