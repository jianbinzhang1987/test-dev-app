package node

import (
	"deploymaster-pro-wails/internal"
	"os"
	"path/filepath"
	"testing"
)

func TestNodeService(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir := t.TempDir()

	// 创建存储
	storage, err := NewJSONStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// 创建服务
	service, err := NewService(storage)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// 测试添加节点
	t.Run("AddNode", func(t *testing.T) {
		node := &internal.Node{
			ID:       "test-001",
			Name:     "测试主节点",
			IP:       "192.168.1.100",
			Port:     22,
			Protocol: internal.ProtocolSFTP,
			IsMaster: true,
		}

		err := service.AddNode(node)
		if err != nil {
			t.Errorf("Failed to add node: %v", err)
		}

		// 验证节点已添加
		nodes := service.ListNodes()
		if len(nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(nodes))
		}
	})

	// 测试获取节点
	t.Run("GetNode", func(t *testing.T) {
		node, err := service.GetNode("test-001")
		if err != nil {
			t.Errorf("Failed to get node: %v", err)
		}

		if node.Name != "测试主节点" {
			t.Errorf("Expected name '测试主节点', got '%s'", node.Name)
		}
	})

	// 测试更新节点
	t.Run("UpdateNode", func(t *testing.T) {
		node, _ := service.GetNode("test-001")
		node.Name = "更新后的主节点"

		err := service.UpdateNode(node)
		if err != nil {
			t.Errorf("Failed to update node: %v", err)
		}

		// 验证更新
		updated, _ := service.GetNode("test-001")
		if updated.Name != "更新后的主节点" {
			t.Errorf("Node name was not updated")
		}
	})

	// 测试添加多个从节点
	t.Run("AddSlaveNodes", func(t *testing.T) {
		slave1 := &internal.Node{
			ID:       "slave-001",
			Name:     "从节点1",
			IP:       "192.168.1.101",
			Port:     22,
			Protocol: internal.ProtocolSFTP,
			IsMaster: false,
		}

		slave2 := &internal.Node{
			ID:       "slave-002",
			Name:     "从节点2",
			IP:       "192.168.1.102",
			Port:     22,
			Protocol: internal.ProtocolSCP,
			IsMaster: false,
		}

		service.AddNode(slave1)
		service.AddNode(slave2)

		slaves := service.GetSlaveNodes()
		if len(slaves) != 2 {
			t.Errorf("Expected 2 slave nodes, got %d", len(slaves))
		}
	})

	// 测试获取主节点
	t.Run("GetMasterNode", func(t *testing.T) {
		master, err := service.GetMasterNode()
		if err != nil {
			t.Errorf("Failed to get master node: %v", err)
		}

		if master.ID != "test-001" {
			t.Errorf("Expected master ID 'test-001', got '%s'", master.ID)
		}
	})

	// 测试删除节点
	t.Run("DeleteNode", func(t *testing.T) {
		err := service.DeleteNode("slave-001")
		if err != nil {
			t.Errorf("Failed to delete node: %v", err)
		}

		nodes := service.ListNodes()
		if len(nodes) != 2 {
			t.Errorf("Expected 2 nodes after deletion, got %d", len(nodes))
		}
	})

	// 测试持久化
	t.Run("Persistence", func(t *testing.T) {
		// 检查JSON文件是否存在
		jsonPath := filepath.Join(tmpDir, "nodes.json")
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			t.Error("JSON file was not created")
		}

		// 创建新的服务实例，验证数据已持久化
		newStorage, _ := NewJSONStorage(tmpDir)
		newService, _ := NewService(newStorage)

		nodes := newService.ListNodes()
		if len(nodes) != 2 {
			t.Errorf("Expected 2 nodes after reload, got %d", len(nodes))
		}
	})

	// 测试重复ID错误
	t.Run("DuplicateID", func(t *testing.T) {
		node := &internal.Node{
			ID:       "test-001",
			Name:     "重复节点",
			IP:       "192.168.1.200",
			Port:     22,
			Protocol: internal.ProtocolSFTP,
			IsMaster: false,
		}

		err := service.AddNode(node)
		if err != ErrNodeExists {
			t.Errorf("Expected ErrNodeExists, got %v", err)
		}
	})
}
