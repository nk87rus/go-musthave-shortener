package model

import (
	"sync"
)

type Resetter interface {
	Reset()
}

// Pool - generic пул объектов с типом T.
// T должен реализовать интерфейс Resetter.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New - создаёт новый пул
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get - возвращает объект из пула.
func (p *Pool[T]) Get() T {
	obj := p.pool.Get().(T)
	obj.Reset()
	return obj
}

// Put - возвращает ранее созданный объект в пул.
func (p *Pool[T]) Put(obj T) {
	p.pool.Put(obj)
}
