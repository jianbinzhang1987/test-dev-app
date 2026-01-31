package credential

import (
	"testing"
)

func TestCredentialStore(t *testing.T) {
	store := NewStore("", nil)

	nodeID := "test-node-123"
	username := "testuser"
	password := "test-password-123"
	passphrase := "test-passphrase-456"

	t.Run("SetAndGetPassword", func(t *testing.T) {
		// 存储密码
		err := store.SetPassword(nodeID, username, password)
		if err != nil {
			t.Fatalf("Failed to set password: %v", err)
		}

		// 获取密码
		retrievedPassword, err := store.GetPassword(nodeID, username)
		if err != nil {
			t.Fatalf("Failed to get password: %v", err)
		}

		if retrievedPassword != password {
			t.Errorf("Expected password %s, got %s", password, retrievedPassword)
		}

		// 检查密码是否存在
		if !store.HasPassword(nodeID, username) {
			t.Error("HasPassword should return true")
		}
	})

	t.Run("SetAndGetKeyPassphrase", func(t *testing.T) {
		// 存储密钥密码短语
		err := store.SetKeyPassphrase(nodeID, passphrase)
		if err != nil {
			t.Fatalf("Failed to set key passphrase: %v", err)
		}

		// 获取密钥密码短语
		retrievedPassphrase, err := store.GetKeyPassphrase(nodeID)
		if err != nil {
			t.Fatalf("Failed to get key passphrase: %v", err)
		}

		if retrievedPassphrase != passphrase {
			t.Errorf("Expected passphrase %s, got %s", passphrase, retrievedPassphrase)
		}

		// 检查密钥密码短语是否存在
		if !store.HasKeyPassphrase(nodeID) {
			t.Error("HasKeyPassphrase should return true")
		}
	})

	t.Run("DeletePassword", func(t *testing.T) {
		// 删除密码
		err := store.DeletePassword(nodeID, username)
		if err != nil {
			t.Fatalf("Failed to delete password: %v", err)
		}

		// 验证已删除
		_, err = store.GetPassword(nodeID, username)
		if err == nil {
			t.Error("Password should have been deleted")
		}

		if store.HasPassword(nodeID, username) {
			t.Error("HasPassword should return false after deletion")
		}
	})

	t.Run("DeleteAll", func(t *testing.T) {
		// 先设置密码和密钥密码短语
		_ = store.SetPassword(nodeID, username, password)
		_ = store.SetKeyPassphrase(nodeID, passphrase)

		// 删除所有凭据
		err := store.DeleteAll(nodeID, username)
		if err != nil {
			t.Fatalf("Failed to delete all credentials: %v", err)
		}

		// 验证都已删除
		if store.HasPassword(nodeID, username) {
			t.Error("Password should have been deleted")
		}
		if store.HasKeyPassphrase(nodeID) {
			t.Error("Key passphrase should have been deleted")
		}
	})

	t.Run("GetNonExistentPassword", func(t *testing.T) {
		_, err := store.GetPassword("non-existent-node", "user")
		if err == nil {
			t.Error("Expected error for non-existent password")
		}
	})
}
