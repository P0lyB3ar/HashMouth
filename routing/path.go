package routing

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// Path represents a route through multiple nodes
type Path struct {
	Nodes []string // Ordered list of node IDs
}

// NewPath creates a new path with the given nodes
func NewPath(nodes []string) (*Path, error) {
	if len(nodes) == 0 {
		return nil, errors.New("path must contain at least one node")
	}
	return &Path{Nodes: nodes}, nil
}

// Length returns the number of nodes in the path
func (p *Path) Length() int {
	return len(p.Nodes)
}

// GetNode returns the node ID at the given index
func (p *Path) GetNode(index int) (string, error) {
	if index < 0 || index >= len(p.Nodes) {
		return "", errors.New("index out of bounds")
	}
	return p.Nodes[index], nil
}

// Contains checks if a node is in the path
func (p *Path) Contains(nodeID string) bool {
	for _, node := range p.Nodes {
		if node == nodeID {
			return true
		}
	}
	return false
}

// PathBuilder helps construct paths through the network
type PathBuilder struct {
	availableNodes []string
	minPathLength  int
	maxPathLength  int
}

// NewPathBuilder creates a new path builder
func NewPathBuilder(nodes []string, minLength, maxLength int) (*PathBuilder, error) {
	if minLength < 1 {
		return nil, errors.New("minimum path length must be at least 1")
	}
	if maxLength < minLength {
		return nil, errors.New("maximum path length must be >= minimum path length")
	}
	if len(nodes) < minLength {
		return nil, errors.New("not enough nodes for minimum path length")
	}

	return &PathBuilder{
		availableNodes: nodes,
		minPathLength:  minLength,
		maxPathLength:  maxLength,
	}, nil
}

// BuildRandomPath creates a random path through available nodes
func (pb *PathBuilder) BuildRandomPath() (*Path, error) {
	if len(pb.availableNodes) == 0 {
		return nil, errors.New("no nodes available")
	}

	// Determine path length
	lengthRange := pb.maxPathLength - pb.minPathLength + 1
	lengthOffset, err := rand.Int(rand.Reader, big.NewInt(int64(lengthRange)))
	if err != nil {
		return nil, err
	}
	pathLength := pb.minPathLength + int(lengthOffset.Int64())

	// Ensure we don't exceed available nodes
	if pathLength > len(pb.availableNodes) {
		pathLength = len(pb.availableNodes)
	}

	// Select random nodes without replacement
	selectedNodes := make([]string, 0, pathLength)
	usedIndices := make(map[int]bool)

	for len(selectedNodes) < pathLength {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(pb.availableNodes))))
		if err != nil {
			return nil, err
		}
		index := int(idx.Int64())

		if !usedIndices[index] {
			usedIndices[index] = true
			selectedNodes = append(selectedNodes, pb.availableNodes[index])
		}
	}

	return NewPath(selectedNodes)
}

// BuildPathExcluding creates a path that excludes certain nodes
func (pb *PathBuilder) BuildPathExcluding(excludeNodes []string) (*Path, error) {
	// Filter available nodes
	filtered := make([]string, 0)
	excludeMap := make(map[string]bool)
	for _, node := range excludeNodes {
		excludeMap[node] = true
	}

	for _, node := range pb.availableNodes {
		if !excludeMap[node] {
			filtered = append(filtered, node)
		}
	}

	if len(filtered) < pb.minPathLength {
		return nil, errors.New("not enough nodes after exclusion")
	}

	// Create temporary builder with filtered nodes
	tempBuilder, err := NewPathBuilder(filtered, pb.minPathLength, pb.maxPathLength)
	if err != nil {
		return nil, err
	}

	return tempBuilder.BuildRandomPath()
}

// BuildMultiplePaths creates multiple diverse paths
func (pb *PathBuilder) BuildMultiplePaths(count int) ([]*Path, error) {
	if count <= 0 {
		return nil, errors.New("count must be positive")
	}

	paths := make([]*Path, 0, count)
	for i := 0; i < count; i++ {
		path, err := pb.BuildRandomPath()
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// Validate checks if a path is valid
func (p *Path) Validate() error {
	if len(p.Nodes) == 0 {
		return errors.New("path is empty")
	}

	// Check for duplicate nodes
	seen := make(map[string]bool)
	for _, node := range p.Nodes {
		if node == "" {
			return errors.New("path contains empty node ID")
		}
		if seen[node] {
			return errors.New("path contains duplicate nodes")
		}
		seen[node] = true
	}

	return nil
}

// Reverse returns a new path with nodes in reverse order
func (p *Path) Reverse() *Path {
	reversed := make([]string, len(p.Nodes))
	for i, node := range p.Nodes {
		reversed[len(p.Nodes)-1-i] = node
	}
	return &Path{Nodes: reversed}
}

// Clone creates a deep copy of the path
func (p *Path) Clone() *Path {
	nodes := make([]string, len(p.Nodes))
	copy(nodes, p.Nodes)
	return &Path{Nodes: nodes}
}
