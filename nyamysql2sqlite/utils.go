package mysql2sqlite

import "errors"

// 自定義錯誤型別
var (
	ErrUnsupportedStatement = errors.New("unsupported SQL statement type")
)
