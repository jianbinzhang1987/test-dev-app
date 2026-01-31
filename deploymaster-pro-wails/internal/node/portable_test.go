package node

import (
	"deploymaster-pro-wails/internal/credential"
	"os"
	"path/filepath"
	"testing"
)

func TestPortableModeAndEncryption(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "portable-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. 初始化存储（模拟便携模式数据目录）
	storage, err := NewJSONStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create JSON storage: %v", err)
	}

	// 2. 初始化凭据存储（传入数据目录和加密器）
	credStore := credential.NewStore(tmpDir, storage.GetCrypto())

	nodeID := "node-1"
	username := "admin"
	password := "secret123"

	// 3. 存储凭据
	if err := credStore.SetPassword(nodeID, username, password); err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// 4. 验证 credentials.json 是否生成
	credFile := filepath.Join(tmpDir, "credentials.json")
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		t.Error("credentials.json was not created")
	}

	// 5. 验证文件内容是否加密（不应包含明文密码）
	content, _ := os.ReadFile(credFile)
	if string(content) == "" || (len(content) > 0 && (contains(string(content), password))) {
		t.Errorf("Password should be encrypted in file, but found plaintext or empty. Content: %s", string(content))
	}

	// 6. 验证读取是否正确解密
	retrieved, err := credStore.GetPassword(nodeID, username)
	if err != nil {
		t.Fatalf("Failed to get password: %v", err)
	}
	if retrieved != password {
		t.Errorf("Expected decrypted password %s, got %s", password, retrieved)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
