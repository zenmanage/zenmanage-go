package zenmanage

import "testing"

func TestCRC32BDeterministic(t *testing.T) {
	a := CRC32B("salt:user-123")
	b := CRC32B("salt:user-123")
	if a != b {
		t.Fatalf("expected deterministic checksum")
	}
}

func TestIsInBucketBounds(t *testing.T) {
	if IsInBucket("salt", "user-1", 0) {
		t.Fatalf("expected 0 percent to exclude user")
	}
	if !IsInBucket("salt", "user-1", 100) {
		t.Fatalf("expected 100 percent to include user")
	}
	if IsInBucket("salt", "", 100) {
		t.Fatalf("expected empty identifier to be excluded")
	}
}

func TestIsInBucketDeterministic(t *testing.T) {
	left := IsInBucket("rollout-salt", "user-abc", 30)
	right := IsInBucket("rollout-salt", "user-abc", 30)
	if left != right {
		t.Fatalf("expected deterministic result")
	}
}

func TestIsInBucketMonotonicity(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := "user-" + string(rune('a'+(i%26)))
		at10 := IsInBucket("salt", id, 10)
		at50 := IsInBucket("salt", id, 50)
		if at10 && !at50 {
			t.Fatalf("expected monotonic rollout for id %s", id)
		}
	}
}
