package nyanats

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// getKey 根據傳入的主題（theme）取得對應的加密金鑰。
// 如果該主題有專屬金鑰則優先使用，否則回傳全域的預設金鑰。
func (p *NyaNATS) getKey(theme string) []byte {
	// 檢查 themeKeys map 中是否存在該主題的金鑰
	if key, ok := p.themeKeys[theme]; ok {
		return key
	}
	// 若找不到則回傳 struct 初始化時設定的 defaultKey
	return p.defaultKey
}

// encrypt 使用 AES-GCM 演算法對明文進行加密。
// 流程包含：取得金鑰、建立加密器、產生隨機 Nonce、執行密封（Seal）。
func (p *NyaNATS) encrypt(theme string, plaintext []byte) ([]byte, error) {
	key := p.getKey(theme)
	// 如果金鑰長度為 0（未設定加密），則直接回傳原始明文，不進行加密動作
	if len(key) == 0 {
		return plaintext, nil
	}

	// 根據金鑰建立 AES 區塊加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 建立 GCM (Galois/Counter Mode) 實例，提供加密與完整性校驗
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 產生一個長度符合 GCM 標準的隨機 nonce（Number used once）
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 執行加密：gcm.Seal 的第一個參數是 dst，會將 nonce 附加在密文前方。
	// 最終回傳格式為：[Nonce][Ciphertext][Tag]
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt 使用 AES-GCM 演算法對密文進行解密。
// 它會從密文中拆分出 Nonce，再利用對應的金鑰還原回明文。
func (p *NyaNATS) decrypt(theme string, ciphertext []byte) ([]byte, error) {
	key := p.getKey(theme)
	// 如果金鑰長度為 0，視為未加密，直接回傳傳入的資料
	if len(key) == 0 {
		return ciphertext, nil
	}

	// 建立 AES 區塊解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 建立 GCM 實例
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 取得 GCM 標準的 Nonce 長度
	nonceSize := gcm.NonceSize()
	l := len(ciphertext)
	// 檢查密文長度是否至少包含 Nonce，避免切片溢位
	if l < nonceSize {
		return nil, fmt.Errorf("E: KEY %d < %d", l, nonceSize)
	}

	// 從 ciphertext 拆分出 Nonce（前半部分）與實際的加密資料（後半部分）
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 執行解密並驗證訊息完整性（Open 函式會自動檢查 Tag）
	return gcm.Open(nil, nonce, actualCiphertext, nil)
}
