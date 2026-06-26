package engine

import "time"

type TrashEntry struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	TrashedAt    time.Time `json:"trashed_at"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"is_dir"`
	Checksum     string    `json:"checksum"`
	Permanent    bool      `json:"permanent"`
	FreeDesktop  bool      `json:"free_desktop"`
}

func NewEntry(originalPath string, isDir bool) (*TrashEntry, error) {
	return nil, nil
}

func AppendEntry(indexPath string, e *TrashEntry) error {
	return nil
}

func ReadAllEntries(indexPath string) ([]*TrashEntry, error) {
	return nil, nil
}

func WriteAllEntries(indexPath string, entries []*TrashEntry) error {
	return nil
}
