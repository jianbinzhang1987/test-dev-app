package topology

import (
	"deploymaster-pro-wails/internal"
)

// Service 拓扑服务
type Service struct{}

// NewService 创建拓扑服务
func NewService() *Service {
	return &Service{}
}

// GetTopologyData 获取拓扑结构数据
func (s *Service) GetTopologyData(nodes []*internal.Node) *internal.TopologyData {
	topology := &internal.TopologyData{
		Total:  len(nodes),
		Slaves: make([]*internal.Node, 0),
	}

	// 分离主节点和从节点
	for _, node := range nodes {
		if node.IsMaster {
			topology.Master = node
		} else {
			topology.Slaves = append(topology.Slaves, node)
		}
	}

	return topology
}
