// Package pool предоставляет типизированный пул объектов со сбросом состояния.
package pool

import "sync"

// Resetter сбрасывает состояние объекта перед его повторным использованием.
type Resetter interface {
	Reset()
}

// Pool хранит временные объекты одного типа и допускает параллельные Get и Put.
// Создаётся через New; после начала использования Pool нельзя копировать.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New создаёт пул с фабрикой новых объектов для Get при пустом пуле.
// Фабрика должна быть ненулевой, допускать параллельные вызовы и каждый раз
// возвращать отдельный готовый к использованию объект, отличный от nil.
func New[T Resetter](factory func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return factory() },
		},
	}
}

// Get возвращает объект из пула или создаёт новый через фабрику.
// Полученный объект принадлежит вызывающему коду до передачи в Put.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put вызывает Reset и затем возвращает объект в пул.
// value должен быть отличен от nil, включая типизированный nil.
// После Put объект передан во владение пула.
func (p *Pool[T]) Put(value T) {
	value.Reset()
	p.pool.Put(value)
}
