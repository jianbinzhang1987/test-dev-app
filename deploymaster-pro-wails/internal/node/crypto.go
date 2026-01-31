package node

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Crypto 处理器，处理配置加密解密
type Crypto struct {
	key []byte
}

// NewCrypto 从指定目录加载或创建密钥
func NewCrypto(dataDir string) (*Crypto, error) {
	keyPath := filepath.Join(dataDir, "key.txt")
	var key []byte

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// 生成新密钥
		key = make([]byte, 32) // AES-256
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, err
		}
		// 存储为十六进制字符串以便于查看，但权限限制为 600
		if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
			return nil, err
		}
	} else {
		// 读取已有密钥
		hexKey, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		key, err = hex.DecodeString(string(hexKey))
		if err != nil {
			return nil, fmt.Errorf("invalid key format: %v", err)
		}
	}

	return &Crypto{key: key}, nil
}

// Encrypt 加密字符串
func (c *Crypto) Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// Decrypt 解密字符串
func (c *Crypto) Decrypt(cipherTextHex string) (string, error) {
	cipherText, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("cipherText too short")
	}

	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
