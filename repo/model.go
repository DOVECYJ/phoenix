package repo

import (
	"time"
	"unsafe"

	"github.com/go-rel/rel"
	"github.com/go-rel/rel/where"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

// Database model that provide base persistent functions.
//
// Usage:
//
//	type User struct {
//		Model[User]
//		Name string
//		Age int
//	}
//
// Then you can use `user.Create()` to create a user.
type Model[T any] struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Model[T]) Create(repo rel.Repository) error {
	return repo.Insert(context.Background(), (*T)(unsafe.Pointer(m)))
}

func (m *Model[T]) Update(repo rel.Repository) error {
	return repo.Update(context.Background(), (*T)(unsafe.Pointer(m)))
}

func (m *Model[T]) Read(repo rel.Repository) error {
	return repo.Find(context.Background(), (*T)(unsafe.Pointer(m)), where.Eq("id", m.ID))
}

func (m *Model[T]) Delete(repo rel.Repository) error {
	return repo.Delete(context.Background(), (*T)(unsafe.Pointer(m)))
}

func (m *Model[T]) Remove(repo rel.Repository) error {
	return repo.Delete(context.Background(), (*T)(unsafe.Pointer(m)), rel.Unscoped(true))
}

// Cacheable database model which added cache function.
type CachedModel[T any] struct {
	Model[T]
}

func (m *CachedModel[T]) Create(repo rel.Repository, redis *redis.Client) error {
	// if err := m.Model.Create(db); err != nil {
	// 	return err
	// }

	panic("not implemant")
}

func (m *CachedModel[T]) Update(repo rel.Repository, redis *redis.Client) error {
	panic("not implemant")
}

func (m *CachedModel[T]) Read(repo rel.Repository, redis *redis.Client) error {
	panic("not implemant")
}

func (m *CachedModel[T]) Delete(repo rel.Repository, redis *redis.Client) error {
	panic("not implemant")
}

func (m *CachedModel[T]) Remove(repo rel.Repository, redis *redis.Client) error {
	panic("not implemant")
}
