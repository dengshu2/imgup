// Package storage abstracts the object store behind imgup. Today only R2/S3 is
// implemented, but commands depend on the Uploader interface so additional
// backends can be added without touching the command layer.
package storage

import (
	"context"
	"time"
)

// ObjectInfo describes a stored object as returned by List.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// Uploader is the minimal contract every backend must satisfy.
type Uploader interface {
	// Exists reports whether an object with the given key is already present.
	Exists(ctx context.Context, key string) (bool, error)

	// Put stores content under key with the given content type. Implementations
	// should be safe to call after Exists returns false; callers use Exists to
	// make uploads idempotent (skip when the md5-derived key already exists).
	Put(ctx context.Context, key string, content []byte, contentType string) error

	// Check performs a lightweight authenticated round-trip used by `doctor` to
	// validate credentials and bucket access.
	Check(ctx context.Context) error

	// List returns objects whose key starts with prefix. When limit > 0 it caps
	// the number of returned objects.
	List(ctx context.Context, prefix string, limit int) ([]ObjectInfo, error)

	// Delete removes the object at key. Deleting a missing key is not an error
	// (standard S3 delete semantics).
	Delete(ctx context.Context, key string) error
}
