package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/safe-rm/internal/config"
	"example.com/safe-rm/internal/log"
	"golang.org/x/sys/unix"
)

type Policy int

const (
	PolicySoftDelete Policy = iota
	PolicyPermanent
	PolicyDanger
	PolicyDangerPermanent
)

func (p Policy) String() string {
	switch p {
	case PolicySoftDelete:
		return "soft-delete"
	case PolicyPermanent:
		return "permanent"
	case PolicyDanger:
		return "danger (soft)"
	case PolicyDangerPermanent:
		return "danger (permanent)"
	default:
		return "unknown"
	}
}

func Route(path string, cfg *config.Config) (Policy, error) {
	return routePath(path, cfg, false)
}

func RouteRecursive(path string, cfg *config.Config) (Policy, error) {
	return routePath(path, cfg, true)
}

func routePath(path string, cfg *config.Config, recursive bool) (Policy, error) {
	isDanger := false
	isBypass := false

	if len(cfg.DangerList) > 0 {
		match, err := MatchesAny(path, cfg.DangerList)
		if err != nil {
			return PolicySoftDelete, err
		}
		isDanger = match
	}

	if len(cfg.BypassList) > 0 {
		match, err := MatchesAny(path, cfg.BypassList)
		if err != nil {
			return PolicySoftDelete, err
		}
		isBypass = match
	}

	if !isDanger && !isBypass && recursive {
		fi, err := os.Stat(path)
		if err == nil && fi.IsDir() {
			childDanger, childBypass, err := scanDir(path, cfg)
			if err != nil {
				return PolicySoftDelete, err
			}
			if childDanger {
				isDanger = true
			}
			if childBypass {
				isBypass = true
			}
		}
	}

	switch {
	case isDanger && isBypass:
		return PolicyDangerPermanent, nil
	case isDanger && !isBypass:
		return PolicyDanger, nil
	case !isDanger && isBypass:
		return PolicyPermanent, nil
	default:
		return PolicySoftDelete, nil
	}
}

func scanDir(dir string, cfg *config.Config) (hasDanger bool, hasBypass bool, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, false, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(dir, entry.Name())

		if len(cfg.DangerList) > 0 {
			match, err := MatchesAny(childPath, cfg.DangerList)
			if err != nil {
				return false, false, err
			}
			if match {
				hasDanger = true
			}
		}

		if len(cfg.BypassList) > 0 {
			match, err := MatchesAny(childPath, cfg.BypassList)
			if err != nil {
				return false, false, err
			}
			if match {
				hasBypass = true
			}
		}

		if entry.IsDir() {
			cd, cb, err := scanDir(childPath, cfg)
			if err != nil {
				return false, false, err
			}
			if cd {
				hasDanger = true
			}
			if cb {
				hasBypass = true
			}
		}

		if hasDanger && hasBypass {
			return true, true, nil
		}
	}

	return hasDanger, hasBypass, nil
}

func SoftDelete(path string, cfg *config.Config, freeDesktop bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fi, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warnf("source disappeared before soft-delete: %s", abs)
			return nil
		}
		return err
	}

	trashDir, err := cfg.ResolvedTrashDir()
	if err != nil {
		return err
	}

	indexDir, err := config.IndexDir()
	if err != nil {
		return err
	}

	entry, err := NewEntry(abs, fi.IsDir())
	if err != nil {
		return err
	}
	entry.FreeDesktop = freeDesktop

	dstDir := filepath.Join(trashDir, "files")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	dst := filepath.Join(dstDir, entry.ID)

	if err := moveOrCopy(abs, dst); err != nil {
		return fmt.Errorf("move to trash: %w", err)
	}

	if freeDesktop {
		infoDir := filepath.Join(trashDir, "info")
		if err := os.MkdirAll(infoDir, 0755); err != nil {
			os.Remove(dst)
			return err
		}
		infoPath := filepath.Join(infoDir, entry.ID+".trashinfo")
		info := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
			abs, entry.TrashedAt.Format("2006-01-02T15:04:05"))
		if err := os.WriteFile(infoPath, []byte(info), 0644); err != nil {
			os.Remove(dst)
			return err
		}
	}

	indexPath := filepath.Join(indexDir, "trash.jsonl")
	if err := AppendEntry(indexPath, entry); err != nil {
		log.Errorf("failed to append trash entry: %v", err)
	}

	log.Infof("soft-deleted %s -> %s (policy=%s)", abs, dst, "soft-delete")
	return nil
}

func PermanentDelete(path string) error {
	log.Infof("permanently deleting %s", path)
	return os.RemoveAll(path)
}

func moveOrCopy(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	if isCrossDevice(err) {
		return copyThenRemove(src, dst)
	}

	return err
}

func isCrossDevice(err error) bool {
	le, ok := err.(*os.LinkError)
	if !ok {
		return false
	}
	return le.Err == unix.EXDEV
}

func copyThenRemove(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		if err := copyDir(src, dst); err != nil {
			return err
		}
	} else {
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}

	return os.RemoveAll(src)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func IsRootOrCWD(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	if abs == "/" {
		return true
	}

	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	return abs == cwd
}
