package nyaapiserver

import (
	"encoding/json"
	"net/http"
)

// BindJSON 會將 HTTP 請求本文中的 JSON 資料反序列化至 v 指向的目標結構中。
// 呼叫端必須傳入可寫入的指標型別，否則 json.Unmarshal 將無法正確填入解析結果。
// 當請求本文不是合法 JSON，或目標型別與 JSON 結構不相容時，會回傳對應錯誤。
func (req *HTTPRequest) BindJSON(v interface{}) error {
	return json.Unmarshal(req.Body, v)
}

// JSONResponse 會建立一個內容為 JSON 的 HTTPResponse，並自動設定 Content-Type 為 application/json。
// data 會先被序列化為 JSON 後再寫入回應本文；若序列化失敗，則改為回傳 500 Internal Server Error。
// 此函式適合用於快速建構標準 JSON API 回應內容。
func JSONResponse(statusCode int, data interface{}) *HTTPResponse {
	body, err := json.Marshal(data)
	if err != nil {
		// 當 JSON 序列化失敗時，回傳標準化的 500 錯誤回應，避免將無效內容寫入回應本文。
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
