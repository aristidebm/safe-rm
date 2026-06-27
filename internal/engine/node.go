package engine

import (
	"os"
	"path/filepath"

	"example.com/safe-rm/internal/config"
)

type DeleteTarget struct {
	Path   string
	Policy Policy
}

type Node struct {
	Path     string
	Name     string
	IsDir    bool
	Policy   Policy
	Children []*Node
	Parent   *Node
	Expanded bool
	Selected bool
	Depth    int
}

func (n *Node) HasChildren() bool {
	return n.IsDir && len(n.Children) > 0
}

func (n *Node) Expand() {
	n.Expanded = true
}

func (n *Node) Collapse() {
	n.Expanded = false
}

func (n *Node) Toggle() {
	n.Expanded = !n.Expanded
}

func (n *Node) Root() *Node {
	curr := n
	for curr.Parent != nil {
		curr = curr.Parent
	}
	return curr
}

func (n *Node) VisibleNodes() []*Node {
	var nodes []*Node
	nodes = append(nodes, n)
	if n.Expanded {
		for _, child := range n.Children {
			nodes = append(nodes, child.VisibleNodes()...)
		}
	}
	return nodes
}

func BuildTree(root string, cfg *config.Config) (*Node, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	policy, err := Route(root, cfg)
	if err != nil {
		return nil, err
	}

	return &Node{
		Path:   root,
		Name:   filepath.Base(root),
		IsDir:  fi.IsDir(),
		Policy: policy,
		Depth:  0,
	}, nil
}

func BuildTreeRecursive(root string, cfg *config.Config) (*Node, error) {
	return buildTreeRecursive(root, cfg, 0)
}

func buildTreeRecursive(path string, cfg *config.Config, depth int) (*Node, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	n := &Node{
		Path:  path,
		Name:  filepath.Base(path),
		IsDir: fi.IsDir(),
		Depth: depth,
	}

	isDanger, err := matchesAnyPattern(path, cfg.DangerList)
	if err != nil {
		return nil, err
	}
	isBypass, err := matchesAnyPattern(path, cfg.BypassList)
	if err != nil {
		return nil, err
	}

	n.Policy = policyFromFlags(isDanger, isBypass)

	if n.IsDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			child, err := buildTreeRecursive(childPath, cfg, depth+1)
			if err != nil {
				return nil, err
			}
			child.Parent = n
			n.Children = append(n.Children, child)
		}
	}

	return n, nil
}

func matchesAnyPattern(path string, patterns []string) (bool, error) {
	if len(patterns) == 0 {
		return false, nil
	}
	return MatchesAny(path, patterns)
}

func policyFromFlags(isDanger, isBypass bool) Policy {
	switch {
	case isDanger && isBypass:
		return PolicyDangerPermanent
	case isDanger && !isBypass:
		return PolicyDanger
	case !isDanger && isBypass:
		return PolicyPermanent
	default:
		return PolicySoftDelete
	}
}

func treeHasPolicy(root string, cfg *config.Config) (hasDanger bool, hasBypass bool, err error) {
	return treeHasPolicyRecursive(root, cfg)
}

func treeHasPolicyRecursive(path string, cfg *config.Config) (hasDanger bool, hasBypass bool, err error) {
	if len(cfg.DangerList) > 0 {
		match, mErr := MatchesAny(path, cfg.DangerList)
		if mErr != nil {
			return false, false, mErr
		}
		if match {
			hasDanger = true
		}
	}

	if len(cfg.BypassList) > 0 {
		match, mErr := MatchesAny(path, cfg.BypassList)
		if mErr != nil {
			return false, false, mErr
		}
		if match {
			hasBypass = true
		}
	}

	if hasDanger && hasBypass {
		return true, true, nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return hasDanger, hasBypass, nil
	}

	if fi.IsDir() {
		entries, rErr := os.ReadDir(path)
		if rErr != nil {
			return hasDanger, hasBypass, rErr
		}
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			cd, cb, cErr := treeHasPolicyRecursive(childPath, cfg)
			if cErr != nil {
				return false, false, cErr
			}
			if cd {
				hasDanger = true
			}
			if cb {
				hasBypass = true
			}
			if hasDanger && hasBypass {
				return true, true, nil
			}
		}
	}

	return hasDanger, hasBypass, nil
}
