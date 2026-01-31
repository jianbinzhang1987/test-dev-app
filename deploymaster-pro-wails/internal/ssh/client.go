package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client SSH客户端封装
type Client struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

// NewClient 创建SSH客户端（密码认证）
func NewClient(username, password string) *Client {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 注意：生产环境应该验证主机密钥
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
	}
}

// NewClientWithKey 使用私钥创建SSH客户端（从字节数组）
func NewClientWithKey(username string, privateKey []byte) (*Client, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key failed: %w", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
	}, nil
}

// NewClientWithKeyFile 从文件路径加载私钥创建SSH客户端
// keyPath: 私钥文件的绝对路径
// passphrase: 私钥密码短语（如果私钥未加密则为空字符串）
func NewClientWithKeyFile(username, keyPath, passphrase string) (*Client, error) {
	// 读取私钥文件
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// 解析私钥
	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
	}, nil
}

// NewClientWithAgent 使用SSH Agent创建客户端
// 需要系统中运行ssh-agent并设置SSH_AUTH_SOCK环境变量
func NewClientWithAgent(username string) (*Client, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
	}

	agentClient := agent.NewClient(conn)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
	}, nil
}

// Connect 连接到远程服务器
func (c *Client) Connect(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	client, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return fmt.Errorf("ssh dial failed: %w", err)
	}

	c.client = client
	return nil
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ExecuteCommand 执行远程命令
func (c *Client) ExecuteCommand(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session failed: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("execute command failed: %w", err)
	}

	return string(output), nil
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	return c.client != nil
}
