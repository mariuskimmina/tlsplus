package acme

import "context"

type Storage interface {
	// Locker provides atomic synchronization
	// operations, making Storage safe to share.
	Locker

	// Store puts value at key.
	Store(ctx context.Context, key string, value []byte) error

	// Load retrieves the value at key.
	Load(ctx context.Context, key string) ([]byte, error)

	// Delete deletes key. An error should be
	// returned only if the key still exists
	// when the method returns.
	Delete(ctx context.Context, key string) error

	// Exists returns true if the key exists
	// and there was no error checking.
	Exists(ctx context.Context, key string) bool

	// List returns all keys that match prefix.
	// If recursive is true, non-terminal keys
	// will be enumerated (i.e. "directories"
	// should be walked); otherwise, only keys
	// prefixed exactly by prefix will be listed.
	List(ctx context.Context, prefix string, recursive bool) ([]string, error)
}

// Locker facilitates synchronization of certificate tasks across
// machines and networks.
type Locker interface {
	// Lock acquires the lock for key, blocking until the lock
	// can be obtained or an error is returned. Note that, even
	// after acquiring a lock, an idempotent operation may have
	// already been performed by another process that acquired
	// the lock before - so always check to make sure idempotent
	// operations still need to be performed after acquiring the
	// lock.
	//
	// The actual implementation of obtaining of a lock must be
	// an atomic operation so that multiple Lock calls at the
	// same time always results in only one caller receiving the
	// lock at any given time.
	//
	// To prevent deadlocks, all implementations (where this concern
	// is relevant) should put a reasonable expiration on the lock in
	// case Unlock is unable to be called due to some sort of network
	// failure or system crash. Additionally, implementations should
	// honor context cancellation as much as possible (in case the
	// caller wishes to give up and free resources before the lock
	// can be obtained).
	Lock(ctx context.Context, key string) error

	// Unlock releases the lock for key. This method must ONLY be
	// called after a successful call to Lock, and only after the
	// critical section is finished, even if it errored or timed
	// out. Unlock cleans up any resources allocated during Lock.
	Unlock(ctx context.Context, key string) error
}
