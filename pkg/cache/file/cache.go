package filecache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/justinbarrick/farm/pkg/cache"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileCache struct {
	CacheDir string
}

func NewFileCache(cacheDir string) (*FileCache, error) {
	cache := FileCache{
		CacheDir: cacheDir,
	}

	err := os.Mkdir(cacheDir, 0777)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	err = os.Mkdir(filepath.Join(cacheDir, "in"), 0777)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	err = os.Mkdir(filepath.Join(cacheDir, "out"), 0777)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	return &cache, nil
}

func (c *FileCache) Copy(src, dst string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}

	return nil
}

func (c *FileCache) Get(entry cache.CacheEntry) error {
	cacheKey := filepath.Join(c.CacheDir, "out", entry.Hash)
	return c.Copy(cacheKey, entry.Filename)
}

func (c *FileCache) Set(filePath string) (cache.CacheEntry, error) {
	fileSum := sha256.New()

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return cache.CacheEntry{}, err
	}

	fileSum.Write(data)

	cacheKey := fmt.Sprintf("%x", fileSum.Sum(nil))
	cacheOut := filepath.Join(c.CacheDir, "out", cacheKey)

	c.Copy(filePath, cacheOut)

	return cache.CacheEntry{
		Filename: filePath,
		Hash:     cacheKey,
	}, nil
}

func (c *FileCache) LoadCacheManifest(cacheKey string) ([]cache.CacheEntry, error) {
	cachePath := filepath.Join(c.CacheDir, "in", cacheKey)

	cacheFile, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer cacheFile.Close()

	entries := []cache.CacheEntry{}
	err = json.NewDecoder(cacheFile).Decode(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *FileCache) DumpCacheManifest(cacheKey string, entries []cache.CacheEntry) error {
	cachePath := filepath.Join(c.CacheDir, "in", cacheKey)

	cacheFile, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer cacheFile.Close()

	return json.NewEncoder(cacheFile).Encode(entries)
}