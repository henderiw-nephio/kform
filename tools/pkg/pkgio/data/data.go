package data

import (
	"fmt"
	"sync"
)

type Data struct {
	d map[string][]byte
	m sync.RWMutex
}

func New() *Data {
	return &Data{
		d: map[string][]byte{},
	}
}

func (r *Data) Add(name string, b []byte) {
	r.m.Lock()
	defer r.m.Unlock()
	r.d[name] = b
}

func (r *Data) Update(name string, b []byte) {
	r.m.Lock()
	defer r.m.Unlock()
	r.d[name] = b
}

func (r *Data) Delete(name string) {
	r.m.Lock()
	defer r.m.Unlock()
	delete(r.d, name)
}

func (r *Data) Exists(name string) bool {
	r.m.RLock()
	defer r.m.RUnlock()
	_, exists := r.d[name]
	return exists
}

func (r *Data) Get(name string) ([]byte, error) {
	r.m.RLock()
	defer r.m.RUnlock()
	b, ok := r.d[name]
	if !ok {
		return nil, fmt.Errorf("not found, key: %s", name)
	}
	return b, nil
}

func (r *Data) List() map[string]string {
	r.m.RLock()
	defer r.m.RUnlock()
	f := make(map[string]string, len(r.d))
	for n, b := range r.d {
		f[n] = string(b)
	}
	return f
}

func (r *Data) Print() {
	r.m.RLock()
	defer r.m.RUnlock()
	for path, data := range r.d {
		fmt.Printf("## file: %s ##\n", path)
		fmt.Println(string(data))
	}
}
