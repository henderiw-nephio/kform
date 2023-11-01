package recorder

import (
	"fmt"
	"sync"
)

type Recorder[T Record] interface {
	Record(r T)
	Get() Records[T]
	Print()
}

type recorder[T Record] struct {
	m        sync.RWMutex
	initOnce sync.Once
	records  Records[T]
}

func New[T Record]() Recorder[T] {
	return &recorder[T]{}
}

func (r *recorder[T]) init() {
	r.initOnce.Do(func() {
		if r.records == nil {
			r.records = []T{}
		}
	})
}

func (r *recorder[T]) Record(rec T) {
	r.init()
	r.m.Lock()
	defer r.m.Unlock()
	r.records = append(r.records, rec)
}

func (r *recorder[T]) Get() Records[T] {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.records
}

func (r *recorder[T]) Print() {
	for _, d := range r.Get() {
		fmt.Println(d)
	}
}
