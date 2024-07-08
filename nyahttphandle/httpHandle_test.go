package nyahttphandle

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {

	// 创建一个模拟的 ResponseWriter
	recorder := httptest.NewRecorder()

	// 使用 recorder 作为 ResponseWriter
	// 例如，传递给一个处理函数
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backMessageByte := AlertInfoJsonKV(w, 2, 1.1, 9001, "", "test")
		fmt.Println(string(backMessageByte))
	})
	// 创建一个请求实例
	req, _ := http.NewRequest("GET", "/", nil)

	// 调用处理函数，传入 recorder 和请求实例
	handler.ServeHTTP(recorder, req)

	// 输出 recorder 中记录的响应
	fmt.Println(recorder.Body.String())
}
