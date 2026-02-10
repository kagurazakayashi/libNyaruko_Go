package nyanats

import (
	"time"

	"github.com/nats-io/nats.go"
)

// Subscribe 訂閱指定的主題（Theme），並在收到訊息時執行回呼函式（Callback）。
// 該方法會自動處理訊息解密，並在回呼函式執行後，將回傳值加密並回覆至 Reply 主題（若存在）。
func (p *NyaNATS) Subscribe(theme string, callback func(m string) string) error {
	// 若初始化過程已有錯誤，直接回傳
	if p.err != nil {
		return p.err
	}

	// 呼叫 NATS 底層訂閱
	_, err := p.natsConn.Subscribe(theme, func(m *nats.Msg) {

		// 根據訊息主題嘗試解密資料
		data, err := p.decrypt(m.Subject, m.Data)
		if err != nil {
			p.logf("<- [%s](ERR) %v", m.Subject, err)
			return
		}
		p.logf("<- [%s](%d) %s", theme, len(m.Data), data)

		// 執行使用者定義的回呼邏輯，取得回覆內容
		replyContent := callback(string(data))

		// 若訊息包含 Reply 地址，則進行回覆處理
		if m.Reply != "" {

			// 加密回覆內容
			encryptedReply, err := p.encrypt(m.Subject, []byte(replyContent))
			if err != nil {
				p.logf("-> [%s](ERR) %v", m.Reply, err)
				return
			}

			// 將加密後的資料發送回 NATS
			if err := m.Respond(encryptedReply); err != nil {
				p.logf("-> [%s](ERR) %v", m.Reply, err)
			} else {
				p.logf("-> [%s](%d) %s", m.Reply, len(encryptedReply), replyContent)
			}
		}
	})

	// 紀錄訂閱成功或失敗的日誌
	if err != nil {
		p.logf("S+ [%s](ERR) %v", theme, err)
	} else {
		p.logf("S+ [%s]", theme)
	}
	return err
}

// Publish 將訊息加密後發布至指定的主題。
func (p *NyaNATS) Publish(theme string, message string) error {
	// 檢查連線狀態
	if p.err != nil {
		return p.err
	}

	// 發布前先進行資料加密
	data, err := p.encrypt(theme, []byte(message))
	if err != nil {
		return err
	}

	// 執行 NATS 發布操作
	err = p.natsConn.Publish(theme, data)
	if err != nil {
		p.logf("-> [%s](ERR) %v", theme, err)
	} else {
		p.logf("-> [%s](%d) %s", theme, len(data), message)
	}
	return err
}

// Request 發送請求訊息並等待對方的回覆，支援設定逾時時間。
// 此方法會自動對請求內容進行加密，並對收到的回覆內容進行解密。
func (p *NyaNATS) Request(theme string, message string, timeout time.Duration) (string, error) {
	// 檢查連線狀態
	if p.err != nil {
		return "", p.err
	}

	// 加密請求訊息
	data, err := p.encrypt(theme, []byte(message))
	if err != nil {
		return "", err
	}

	p.logf("-> [%s](%d) %v", theme, len(data), message)

	// 發送請求並等待回傳
	msg, err := p.natsConn.Request(theme, data, timeout)
	if err != nil {
		p.logf("<- [%s](ERR) %v", theme, err)
		return "", err
	}

	// 解密回傳的訊息內容
	decryptedData, err := p.decrypt(theme, msg.Data)
	if err != nil {
		p.logf("<- [%s](ERR) %v", theme, err)
		return "", err
	}

	p.logf("<- [%s](%d) %s", theme, len(msg.Data), decryptedData)
	return string(decryptedData), nil
}
