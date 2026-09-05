package zenmanage

import (
	"fmt"
	"testing"
)

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

// crossSDKVector is a fixed (salt, identifier) -> bucket vector shared across
// every SDK's test suite (see sdks/zenmanage-php's RolloutBucketTest.php and
// zenmanage-javascript's rollout.test.ts). Every SDK's CRC32B/bucketing
// implementation must produce identical results for these inputs — that's
// what makes percentage rollouts deterministic across languages.
type crossSDKVector struct {
	salt       string
	identifier string
	unsigned   uint32
	bucket     int
	at50       bool
	at25       bool
	at10       bool
}

var crossSDKVectors = []crossSDKVector{
	{salt: "test-salt", identifier: "user-0", unsigned: 2211483234, bucket: 34, at50: true, at25: false, at10: false},
	{salt: "test-salt", identifier: "user-2", unsigned: 1843326798, bucket: 98, at50: false, at25: false, at10: false},
	{salt: "abc123", identifier: "ctx-alpha", unsigned: 2997997254, bucket: 54, at50: false, at25: false, at10: false},
	{salt: "abc123", identifier: "ctx-beta", unsigned: 58423103, bucket: 3, at50: true, at25: true, at10: true},
	{salt: "rollout-salt", identifier: "user-100", unsigned: 2395497973, bucket: 73, at50: false, at25: false, at10: false},
	{salt: "rollout-salt", identifier: "user-200", unsigned: 2358172588, bucket: 88, at50: false, at25: false, at10: false},
	{salt: "fixed-salt", identifier: "user-42", unsigned: 1886245039, bucket: 39, at50: true, at25: false, at10: false},
}

func TestCRC32BMatchesCrossSDKVectors(t *testing.T) {
	for _, v := range crossSDKVectors {
		t.Run(fmt.Sprintf("%s/%s", v.salt, v.identifier), func(t *testing.T) {
			got := CRC32B(v.salt + ":" + v.identifier)
			if got != v.unsigned {
				t.Fatalf("CRC32B(%q) = %d, want %d", v.salt+":"+v.identifier, got, v.unsigned)
			}
			if bucket := int(got % 100); bucket != v.bucket {
				t.Fatalf("bucket = %d, want %d", bucket, v.bucket)
			}
		})
	}
}

func TestIsInBucketMatchesCrossSDKVectors(t *testing.T) {
	for _, v := range crossSDKVectors {
		t.Run(fmt.Sprintf("%s/%s", v.salt, v.identifier), func(t *testing.T) {
			if got := IsInBucket(v.salt, v.identifier, 50); got != v.at50 {
				t.Fatalf("IsInBucket(..., 50) = %v, want %v", got, v.at50)
			}
			if got := IsInBucket(v.salt, v.identifier, 25); got != v.at25 {
				t.Fatalf("IsInBucket(..., 25) = %v, want %v", got, v.at25)
			}
			if got := IsInBucket(v.salt, v.identifier, 10); got != v.at10 {
				t.Fatalf("IsInBucket(..., 10) = %v, want %v", got, v.at10)
			}
		})
	}
}
