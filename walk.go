package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

type FolderStat struct {
	Key    string
	Files  int64
	Bytes  int64
	Errors int64
}

type Summary struct {
	Files    int64
	Bytes    int64
	Dirs     int64
	Errors   int64
	ByFolder map[string]*FolderStat
}

func (s *Summary) sortedFolders() []*FolderStat {
	out := make([]*FolderStat, 0, len(s.ByFolder))
	for _, v := range s.ByFolder {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Bytes == out[j].Bytes {
			return out[i].Key < out[j].Key
		}
		return out[i].Bytes > out[j].Bytes
	})
	return out
}

func (s *Summary) addFile(top string, size int64) {
	s.Files++
	s.Bytes += size
	st := s.folder(top)
	st.Files++
	st.Bytes += size
}

func (s *Summary) addDir(top string) {
	s.Dirs++
	s.folder(top)
}

func (s *Summary) addErr(top string) {
	s.Errors++
	s.folder(top).Errors++
}

func (s *Summary) folder(top string) *FolderStat {
	if s.ByFolder == nil {
		s.ByFolder = map[string]*FolderStat{}
	}
	st, ok := s.ByFolder[top]
	if !ok {
		st = &FolderStat{Key: top}
		s.ByFolder[top] = st
	}
	return st
}

type itemHandler func(src string, info os.FileInfo) error

func expandRoot(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("源路径为空")
	}
	if len(root) == 2 && root[1] == ':' {
		root = strings.ToUpper(root[:1]) + `:\`
	} else if len(root) >= 3 && root[1] == ':' && (root[2] == '\\' || root[2] == '/') && len(strings.Trim(root[3:], `\/`)) == 0 {
		root = strings.ToUpper(root[:1]) + `:\`
	} else {
		abs, err := filepath.Abs(root)
		if err != nil {
			return "", err
		}
		root = abs
	}
	if _, err := os.Lstat(root); err != nil {
		return "", err
	}
	return filepath.Clean(root), nil
}

func isReparsePoint(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	st, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return false
	}
	return st.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}

func canonicalDir(path string) string {
	path = normalizeAbs(path)
	eval, err := filepath.EvalSymlinks(path)
	if err != nil {
		return strings.ToLower(path)
	}
	return strings.ToLower(filepath.Clean(eval))
}

func walkSources(ctx context.Context, roots []string, opt Options, onErr func(src string, err error), fn itemHandler) error {
	visited := map[string]bool{}
	for _, root := range roots {
		abs, err := expandRoot(root)
		if err != nil {
			onErr(root, err)
			continue
		}
		err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if walkErr != nil {
				onErr(path, walkErr)
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			info, err := d.Info()
			if err != nil {
				onErr(path, err)
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			skip, _, _ := skipReason(path, info.IsDir(), opt)
			if skip {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				canon := canonicalDir(path)
				if visited[canon] {
					return filepath.SkipDir
				}
				visited[canon] = true
				if isReparsePoint(info) && path != abs {
					if target, err := os.Readlink(path); err == nil {
						targetAbs := target
						if !filepath.IsAbs(target) {
							targetAbs = filepath.Join(filepath.Dir(path), target)
						}
						targetAbs = normalizeAbs(targetAbs)
						if opt.DestAbs != "" && isBaseOf(opt.DestAbs, targetAbs) {
							return filepath.SkipDir
						}
					}
				}
			}

			return fn(path, info)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func Scan(ctx context.Context, roots []string, opt Options, onErr func(src string, err error)) (*Summary, error) {
	sum := &Summary{ByFolder: map[string]*FolderStat{}}
	err := walkSources(ctx, roots, opt, func(src string, e error) {
		vol, rel := volumeAndRel(src)
		sum.addErr(topLevelKey(vol, rel))
		if onErr != nil {
			onErr(src, e)
		}
	}, func(src string, info os.FileInfo) error {
		vol, rel := volumeAndRel(src)
		top := topLevelKey(vol, rel)
		if info.IsDir() {
			sum.addDir(top)
			return nil
		}
		sum.addFile(top, info.Size())
		return nil
	})
	return sum, err
}
