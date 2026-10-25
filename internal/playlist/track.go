package playlist

import (
	"path/filepath"
	"strings"
)

// Track is one audio file in a playlist.
type Track struct {
	Path string
}

// Title is the file name without its extension.
func (t Track) Title() string {
	name := t.FileName()
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func (t Track) FileName() string { return filepath.Base(t.Path) }

func (t Track) Dir() string { return filepath.Dir(t.Path) }
