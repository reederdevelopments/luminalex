package webx

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Rockup-Consulting/std/x/logx"
)

// ====================================================================
// ETAG CACHING

type memCache struct {
	mu sync.Mutex
	// map[assetpath]md5hash
	c map[string]string
}

func newMemCache() memCache {
	return memCache{
		c: map[string]string{},
	}
}

func (c *memCache) get(route string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fileHash, ok := c.c[route]
	if ok {
		return fileHash, true
	}

	return "", false
}

func (c *memCache) set(route string, r io.Reader) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fileHash, err := md5Util(r)
	if err != nil {
		return "", err
	}

	c.c[route] = fileHash
	return fileHash, nil
}

func md5Util(r io.Reader) (string, error) {
	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func StaticHandler(fs fs.FS, l *log.Logger, isDev bool) http.HandlerFunc {
	c := newMemCache()
	fileServer := http.FileServer(http.FS(fs))

	if !isDev {
		l = logx.NewDiscard()
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// In development, we don't cache to allow for hot-reloading.
		if isDev {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			fileServer.ServeHTTP(w, r)
			return
		}

		etag, ok := c.get(path)

		if !ok {
			l.Printf("etag not set, hashing file: %s", path)

			file, err := fs.Open(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					http.NotFound(w, r)
					return
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			etag, err = c.set(path, file)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if etag == r.Header.Get("If-None-Match") {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("ETag", etag)
		fileServer.ServeHTTP(w, r)
	}
}
