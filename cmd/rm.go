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

type deleteTarget struct {
	Path   string
	Policy engine.Policy
}

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

	var targets []deleteTarget

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

		policy, err := engine.Route(abs, cfg)
		if err != nil {
			return err
		}

		if interactive {
			policy = engine.PolicyDanger
		}

		targets = append(targets, deleteTarget{Path: abs, Policy: policy})
	}

	var directTargets []deleteTarget
	var dangerTargets []deleteTarget

	for _, t := range targets {
		switch t.Policy {
		case engine.PolicyDanger, engine.PolicyDangerPermanent:
			dangerTargets = append(dangerTargets, t)
		default:
			directTargets = append(directTargets, t)
		}
	}

	if len(dangerTargets) > 0 {
		log.Infof("%d targets require TUI confirmation", len(dangerTargets))

		items := make([]tui.ConfirmItem, len(dangerTargets))
		for i, t := range dangerTargets {
			fi, err := os.Stat(t.Path)
			isDir := err == nil && fi.IsDir()
			items[i] = tui.ConfirmItem{
				Path:     t.Path,
				IsDir:    isDir,
				Policy:   t.Policy,
				Selected: true,
			}
		}

		confirmedItems, confirmed := tui.RunConfirm(items)
		if !confirmed {
			log.Infof("TUI confirmation aborted by user")
			return nil
		}

		for _, item := range confirmedItems {
			if !item.Selected {
				continue
			}

			var err error
			switch item.Policy {
			case engine.PolicyDangerPermanent:
				err = engine.PermanentDelete(item.Path)
			default:
				err = engine.SoftDelete(item.Path, cfg, false)
			}

			if err != nil {
				log.Errorf("failed to delete %s: %v", item.Path, err)
				return err
			}

			if verbose {
				switch item.Policy {
				case engine.PolicyDangerPermanent:
					fmt.Fprintf(os.Stderr, "safe-rm: permanently deleted %s\n", item.Path)
				default:
					fmt.Fprintf(os.Stderr, "safe-rm: trashed %s\n", item.Path)
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
