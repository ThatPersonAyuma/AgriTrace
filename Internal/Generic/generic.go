package generic

import (
	"sync"
)

type UserLogin struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

type Result[T any, return_effect any] struct{
	Value T
	Effect func() return_effect
}

// Lazy memoizes result of f() and is safe for concurrent use.
type Lazy[T any] struct {
	once sync.Once
	f    func() (T, error)
	mu  sync.Mutex // protects v and err reads while f may still be running
	v   T
	err error
	done bool
}

func NewLazy[T any](f func() (T, error)) Lazy[T] {
	return Lazy[T]{f: f}
}

func (l *Lazy[T]) Get() (T, error) { // Impure
	// Ensure f runs only once
	l.once.Do(func() {
		v, err := l.f()
		l.mu.Lock()
		l.v = v
		l.err = err
		l.done = true
		l.mu.Unlock()
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.v, l.err
}