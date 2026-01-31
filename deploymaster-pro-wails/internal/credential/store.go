package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "deploymaster-pro"

	// 凭据类型前缀
	passwordPrefix   = "ssh-password"
	passphrasePrefix = "ssh-key-passphrase"
)

// Credential 项
type CredEntry struct {
	Account string `json:"account"`
	Secret  string `json:"secret"` // 加密后的密文
}

// Store 凭据存储服务
type Store struct {
	serviceName  string
	dataDir      string
	cryptoDirect interface {
		Encrypt(plainText string) (string, error)
		Decrypt(cipherTextHex string) (string, error)
	}
	mu sync.RWMutex
}

// NewStore 创建凭据存储实例
// 如果指定了 dataDir，则使用其中的 credentials.json 存储加密数据（便携模式）
func NewStore(dataDir string, crypto interface {
	Encrypt(plainText string) (string, error)
	Decrypt(cipherTextHex string) (string, error)
}) *Store {
	return &Store{
		serviceName:  ServiceName,
		dataDir:      dataDir,
		cryptoDirect: crypto,
	}
}

// SetPassword 存储SSH密码
// nodeID: 节点ID
// username: SSH用户名
// password: SSH密码
func (s *Store) SetPassword(nodeID, username, password string) error {
	account := fmt.Sprintf("%s-%s-%s", passwordPrefix, nodeID, username)
	return s.set(account, password)
}

// GetPassword 获取SSH密码
func (s *Store) GetPassword(nodeID, username string) (string, error) {
	account := fmt.Sprintf("%s-%s-%s", passwordPrefix, nodeID, username)
	return s.get(account)
}

// DeletePassword 删除SSH密码
func (s *Store) DeletePassword(nodeID, username string) error {
	account := fmt.Sprintf("%s-%s-%s", passwordPrefix, nodeID, username)
	return s.delete(account)
}

// SetKeyPassphrase 存储SSH私钥的密码短语
// nodeID: 节点ID
// passphrase: 私钥密码短语
func (s *Store) SetKeyPassphrase(nodeID, passphrase string) error {
	account := fmt.Sprintf("%s-%s", passphrasePrefix, nodeID)
	return s.set(account, passphrase)
}

// GetKeyPassphrase 获取SSH私钥的密码短语
func (s *Store) GetKeyPassphrase(nodeID string) (string, error) {
	account := fmt.Sprintf("%s-%s", passphrasePrefix, nodeID)
	return s.get(account)
}

// DeleteKeyPassphrase 删除SSH私钥的密码短语
func (s *Store) DeleteKeyPassphrase(nodeID string) error {
	account := fmt.Sprintf("%s-%s", passphrasePrefix, nodeID)
	return s.delete(account)
}

// ===== 内部辅助函数 =====

func (s *Store) getCredFilePath() string {
	if s.dataDir == "" {
		return ""
	}
	return filepath.Join(s.dataDir, "credentials.json")
}

func (s *Store) loadFile() (map[string]string, error) {
	path := s.getCredFilePath()
	if path == "" {
		return nil, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []CredEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	res := make(map[string]string)
	for _, e := range entries {
		res[e.Account] = e.Secret
	}
	return res, nil
}

func (s *Store) saveFile(creds map[string]string) error {
	path := s.getCredFilePath()
	if path == "" {
		return nil
	}

	var entries []CredEntry
	for k, v := range creds {
		entries = append(entries, CredEntry{Account: k, Secret: v})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (s *Store) set(account, secret string) error {
	// 如果有加密器，则先加密
	encrypted := secret
	if s.cryptoDirect != nil {
		var err error
		encrypted, err = s.cryptoDirect.Encrypt(secret)
		if err != nil {
			return err
		}
	}

	// 如果是便携模式，存入文件
	if s.dataDir != "" {
		s.mu.Lock()
		defer s.mu.Unlock()
		creds, err := s.loadFile()
		if err != nil {
			return err
		}
		creds[account] = encrypted
		return s.saveFile(creds)
	}

	// 否则回退到系统密钥链
	return keyring.Set(s.serviceName, account, encrypted)
}

func (s *Store) get(account string) (string, error) {
	var encrypted string
	var err error

	// 优先从文件读取
	if s.dataDir != "" {
		s.mu.RLock()
		creds, loadErr := s.loadFile()
		s.mu.RUnlock()
		if loadErr == nil {
			if val, ok := creds[account]; ok {
				encrypted = val
			}
		}
	}

	// 如果文件里没有，尝试从密钥链读取（或者 dataDir 为空时）
	if encrypted == "" {
		encrypted, err = keyring.Get(s.serviceName, account)
		if err != nil {
			return "", err
		}
	}

	// 如果有加密器，则解密
	if s.cryptoDirect != nil {
		return s.cryptoDirect.Decrypt(encrypted)
	}

	return encrypted, nil
}

func (s *Store) delete(account string) error {
	if s.dataDir != "" {
		s.mu.Lock()
		defer s.mu.Unlock()
		creds, err := s.loadFile()
		if err != nil {
			return err
		}
		delete(creds, account)
		return s.saveFile(creds)
	}

	return keyring.Delete(s.serviceName, account)
}

// DeleteAll 删除节点的所有凭据
// 用于删除节点时清理相关凭据
func (s *Store) DeleteAll(nodeID, username string) error {
	// 删除密码（可能不存在，忽略错误）
	_ = s.DeletePassword(nodeID, username)
	// 删除密钥密码短语（可能不存在，忽略错误）
	_ = s.DeleteKeyPassphrase(nodeID)
	return nil
}

// HasPassword 检查是否存储了密码
func (s *Store) HasPassword(nodeID, username string) bool {
	_, err := s.GetPassword(nodeID, username)
	return err == nil
}

// HasKeyPassphrase 检查是否存储了密钥密码短语
func (s *Store) HasKeyPassphrase(nodeID string) bool {
	_, err := s.GetKeyPassphrase(nodeID)
	return err == nil
}
