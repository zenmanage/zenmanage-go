package zenmanage

import (
	"testing"
	"time"
)

func TestInMemoryCacheSetGetDelete(t *testing.T) {
	c := NewInMemoryCache()
	if err := c.Set("k", "v", time.Minute); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	v, ok, err := c.Get("k")
	if err != nil || !ok || v != "v" {
		t.Fatalf("get failed")
	}
	_ = c.Delete("k")
	_, ok, _ = c.Get("k")
	if ok {
		t.Fatalf("expected delete")
	}
}

func TestInMemoryCacheExpiry(t *testing.T) {
	c := NewInMemoryCache()
	_ = c.Set("k", "v", 5*time.Millisecond)
	time.Sleep(15 * time.Millisecond)
	_, ok, _ := c.Get("k")
	if ok {
		t.Fatalf("expected key to expire")
	}
}
