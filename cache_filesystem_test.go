package zenmanage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileSystemCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := NewFileSystemCache(dir)
	if err := c.Set("k", "v", time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	v, ok, err := c.Get("k")
	if err != nil || !ok || v != "v" {
		t.Fatalf("get failed")
	}
	if err := c.Delete("k"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_, ok, _ = c.Get("k")
	if ok {
		t.Fatalf("expected deleted key")
	}
}

func TestFileSystemCacheExpiryAndClear(t *testing.T) {
	dir := t.TempDir()
	c := NewFileSystemCache(dir)
	_ = c.Set("a", "1", 1*time.Millisecond)
	_ = c.Set("b", "2", time.Minute)
	time.Sleep(5 * time.Millisecond)
	_, ok, _ := c.Get("a")
	if ok {
		t.Fatalf("expected expired key")
	}
	if err := c.Clear(); err != nil {
		t.Fatalf("clear failed: %v", err)
	}
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			t.Fatalf("expected no cache files after clear")
		}
	}
}
