// MySQL 資料庫連線池
package nyamysql

import (
	"encoding/json"
	"fmt"
	"time"
)

type MySQLPool MySQLPoolT
type MySQLPoolT struct {
	mySQLConfig   MySQLDBConfig
	mySQLLink     int
	mySQLLinkMax  int
	mySQLPoolList []*NyaMySQL
	waitCount     int
	waitTime      int
	err           error
}

// NewPool: 建立新的 NyaMySQL 池，代替 New 。
func NewPool(configJsonString string, linkMax int) *MySQLPool {
	var mySQLConfig MySQLDBConfig
	err := json.Unmarshal([]byte(configJsonString), &mySQLConfig)
	if err != nil {
		return &MySQLPool{err: err}
	}
	return NewPoolC(mySQLConfig, linkMax)
}

// NewPoolC: 同上, `configJsonString` 改為 `mySQLConfig` 以支援直接配置輸入
func NewPoolC(mySQLConfig MySQLDBConfig, linkMax int) *MySQLPool {
	var o = &MySQLPool{
		mySQLConfig:   mySQLConfig,
		mySQLLink:     0,
		mySQLLinkMax:  linkMax,
		mySQLPoolList: []*NyaMySQL{},
		waitCount:     10,
		waitTime:      500,
		err:           nil,
	}
	for i := 0; i < o.mySQLLinkMax; i++ {
		o.mySQLPoolList = append(o.mySQLPoolList, nil)
	}
	return o
}

// mysqlIsRun: 檢查並建立新的 MySQL 連線。
//
// 若連線數已達上限，則等待可用連線。
// 引數 isShowPrint 控制是否列印錯誤資訊。
// 返回值：新建的 MySQL 連線索引（或 -1 表示失敗），以及可能的錯誤資訊。
func (p *MySQLPool) MysqlIsRun(isShowPrint bool) (int, error) {
	// 如果連線數已滿，則進入等待
	if p.mySQLLink >= p.mySQLLinkMax {
		wc := 0
		for {
			if p.mySQLLink < p.mySQLLinkMax {
				break // 退出等待
			}
			if wc > p.waitCount {
				return -1, fmt.Errorf("MySQL connections are full") // 連線數超限
			}
			wc += 1
			time.Sleep(time.Duration(p.waitTime) * time.Millisecond) // 休眠等待
		}
	}

	// 建立新的 MySQL 連線
	nyaMS := NewC(p.mySQLConfig, nil)
	if nyaMS.Error() != nil {
		if isShowPrint {
			println("MySQL DB Link error:", nyaMS.Error().Error())
		}
		return -1, nyaMS.Error()
	}

	// 查詢空位存放新連線
	ii := 0
	for i := 0; i < len(p.mySQLPoolList); i++ {
		if p.mySQLPoolList[i] == nil {
			ii = i
			break
		}
	}

	// 記錄新的連線
	p.mySQLLink += 1
	p.mySQLPoolList[ii] = nyaMS
	fmt.Println("MySQL connection successful!")
	fmt.Println("MySQL Link Number:", p.mySQLLink, ii, len(p.mySQLPoolList))
	return ii, nil
}

// mysqlClose: 關閉指定索引的 MySQL 連線。
//
// 引數 i 表示要關閉的連線索引，isShowPrint 控制是否列印關閉資訊。
func (p *MySQLPool) MysqlClose(i int, isShowPrint bool) {
	// 檢查索引是否有效
	if i < 0 || i >= len(p.mySQLPoolList) {
		fmt.Println("MySQL is Close!", p.mySQLLink, i, len(p.mySQLPoolList))
		return
	}

	// 關閉連線並釋放資源
	if p.mySQLPoolList[i] != nil {
		fmt.Println("Close MySQL Link!", p.mySQLLink, i, len(p.mySQLPoolList))
		p.mySQLPoolList[i].Close()
		p.mySQLPoolList[i] = nil
		p.mySQLLink -= 1

		// 防止連線數低於 0
		if p.mySQLLink < 0 {
			p.mySQLLink = 0
		}

		if isShowPrint {
			println("MySQL Close Connection! Current number of connections:", p.mySQLLink)
		}
	}
}
