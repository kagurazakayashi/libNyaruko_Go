package nyaapiserver

import (
	"encoding/json"
	"net/http"
)

// BindJSON 是一個輔助方法，用於將 HTTP 請求的 Body (JSON 格式) 反序列化至指定的 Go 結構體中。
// 參數 v 必須是一個指標 (pointer)。
func (req *HTTPRequest) BindJSON(v interface{}) error {
	return json.Unmarshal(req.Body, v)
}

// JSONResponse 是一個輔助函式，用於快速建立包含 JSON 資料的 HTTPResponse。
// 它會自動將傳入的 data 序列化為 JSON，並設定對應的 Content-Type 標頭。
func JSONResponse(statusCode int, data interface{}) *HTTPResponse {
	body, err := json.Marshal(data)
	if err != nil {
		// 若序列化失敗，退回 500 內部伺服器錯誤
		return &HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       []byte(`{"error": "Internal Server Error: JSON marshal failed"}`),
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}

	return &HTTPResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}
