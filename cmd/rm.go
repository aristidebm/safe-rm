package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"example.com/safe-rm/internal/engine"
	"example.com/safe-rm/internal/log"
	"example.com/safe-rm/internal/tui"

	"github.com/spf13/cobra"
)

func rmRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Usage()
		return nil
	}

	for _, arg := range args {
		if arg == "." || arg == "/" {
			err := fmt.Errorf("refusing to delete %q", arg)
			log.Errorf("%v", err)
			return err
		}
	}

	var directTargets []engine.DeleteTarget
	var dangerTargets []*engine.Node

	for _, arg := range args {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return fmt.Errorf("cannot resolve path %q: %w", arg, err)
		}

		if _, err := os.Stat(abs); os.IsNotExist(err) {
			if force {
				log.Infof("skipping non-existent %s (force)", abs)
				continue
			}
			return fmt.Errorf("cannot remove %q: no such file or directory", abs)
		}

		fi, err := os.Stat(abs)
		if err != nil {
			return err
		}

		if fi.IsDir() && !recursive {
			return fmt.Errorf("cannot remove %q: is a directory", abs)
		}

		var policy engine.Policy
		if fi.IsDir() && recursive {
			policy, err = engine.RouteRecursive(abs, cfg)
		} else {
			policy, err = engine.Route(abs, cfg)
		}
		if err != nil {
			return err
		}

		if interactive {
			policy = engine.PolicyDanger
		}

		switch policy {
		case engine.PolicyDanger, engine.PolicyDangerPermanent:
			var root *engine.Node
			if fi.IsDir() && recursive {
				root, err = engine.BuildTreeRecursive(abs, cfg)
			} else {
				root, err = engine.BuildTree(abs, cfg)
			}
			if err != nil {
				return err
			}
			if root == nil {
				continue
			}
			root.Selected = true
			dangerTargets = append(dangerTargets, root)
		default:
			directTargets = append(directTargets, engine.DeleteTarget{Path: abs, Policy: policy})
		}
	}

	if len(dangerTargets) > 0 {
		log.Infof("%d targets require TUI confirmation", len(dangerTargets))

		for _, root := range dangerTargets {
			confirmedRoot, confirmed := tui.RunConfirm(root)
			if !confirmed {
				log.Infof("TUI confirmation aborted by user")
				return nil
			}

			processed := make(map[string]bool)

			for _, node := range confirmedRoot.VisibleNodes() {
				if !node.Selected || node == confirmedRoot {
					continue
				}

				if processed[node.Path] {
					continue
				}

				if node.IsDir && len(node.Children) > 0 {
					continue
				}

				var err error
				switch node.Policy {
				case engine.PolicyDangerPermanent:
					err = engine.PermanentDelete(node.Path)
				default:
					err = engine.SoftDelete(node.Path, cfg, false)
				}

				if err != nil {
					log.Errorf("failed to delete %s: %v", node.Path, err)
					return err
				}

				processed[node.Path] = true

				if verbose {
					switch node.Policy {
					case engine.PolicyDangerPermanent:
						fmt.Fprintf(os.Stderr, "safe-rm: permanently deleted %s\n", node.Path)
					default:
						fmt.Fprintf(os.Stderr, "safe-rm: trashed %s\n", node.Path)
					}
				}
			}

			for _, node := range confirmedRoot.VisibleNodes() {
				if !node.Selected || node == confirmedRoot {
					continue
				}
				if !node.IsDir || len(node.Children) == 0 {
					continue
				}
				allChildrenProcessed := true
				for _, child := range node.Children {
					if !processed[child.Path] {
						allChildrenProcessed = false
						break
					}
				}
				if allChildrenProcessed {
					if err := os.RemoveAll(node.Path); err != nil {
						log.Warnf("failed to remove empty directory %s: %v", node.Path, err)
					}
				}
			}
		}
	}

	for _, t := range directTargets {
		var err error

		switch t.Policy {
		case engine.PolicyPermanent:
			err = engine.PermanentDelete(t.Path)
		case engine.PolicySoftDelete:
			err = engine.SoftDelete(t.Path, cfg, false)
		}

		if err != nil {
			log.Errorf("failed to delete %s: %v", t.Path, err)
			return err
		}

		if verbose {
			switch t.Policy {
			case engine.PolicyPermanent:
				fmt.Fprintf(os.Stderr, "safe-rm: permanently deleted %s\n", t.Path)
			case engine.PolicySoftDelete:
				fmt.Fprintf(os.Stderr, "safe-rm: trashed %s\n", t.Path)
			}
		}
	}

	return nil
}
