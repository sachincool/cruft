package fsutil

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// SizeResult is the output of Size: total bytes plus optional
// top-N largest immediate-or-recursive entries for the TUI preview.
type SizeResult struct {
	Bytes        int64
	LastModified time.Time
	SamplePaths  []SamplePath
}

type SamplePath struct {
	Path         string
	Bytes        int64
	LastModified time.Time
}

// Size walks path and sums regular file sizes. If sampleN > 0, it also
// returns the top-N largest immediate children of path (with their own
// recursive sizes) for the TUI preview.
//
// If path does not exist, returns a zero SizeResult and nil error.
func Size(ctx context.Context, path string, sampleN int) (SizeResult, error) {
	path = Expand(path)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SizeResult{}, nil
		}
		return SizeResult{}, err
	}
	if !info.IsDir() {
		return SizeResult{
			Bytes:        info.Size(),
			LastModified: info.ModTime(),
		}, nil
	}

	// Tree total.
	var total int64
	var newestMtime time.Time
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// Skip unreadable entries rather than failing the whole walk.
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.Type().IsRegular() {
			if fi, err := d.Info(); err == nil {
				atomic.AddInt64(&total, fi.Size())
				if fi.ModTime().After(newestMtime) {
					newestMtime = fi.ModTime()
				}
			}
		}
		return nil
	})
	if err != nil && err != ctx.Err() {
		return SizeResult{}, err
	}

	res := SizeResult{Bytes: total, LastModified: newestMtime}

	if sampleN <= 0 {
		return res, nil
	}

	// Top-N largest immediate children.
	entries, err := os.ReadDir(path)
	if err != nil {
		return res, nil
	}

	type sized struct{ s SamplePath }
	sizes := make([]sized, 0, len(entries))
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 8)
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		e := e
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			child := filepath.Join(path, e.Name())
			var b int64
			var mt time.Time
			if e.IsDir() {
				_ = filepath.WalkDir(child, func(_ string, d fs.DirEntry, werr error) error {
					if werr != nil {
						return nil
					}
					if d.Type().IsRegular() {
						if fi, err := d.Info(); err == nil {
							b += fi.Size()
							if fi.ModTime().After(mt) {
								mt = fi.ModTime()
							}
						}
					}
					return nil
				})
			} else {
				if fi, err := e.Info(); err == nil {
					b = fi.Size()
					mt = fi.ModTime()
				}
			}
			mu.Lock()
			sizes = append(sizes, sized{SamplePath{Path: child, Bytes: b, LastModified: mt}})
			mu.Unlock()
		}()
	}
	wg.Wait()

	sort.Slice(sizes, func(i, j int) bool { return sizes[i].s.Bytes > sizes[j].s.Bytes })
	n := sampleN
	if n > len(sizes) {
		n = len(sizes)
	}
	for i := 0; i < n; i++ {
		res.SamplePaths = append(res.SamplePaths, sizes[i].s)
	}
	return res, nil
}
