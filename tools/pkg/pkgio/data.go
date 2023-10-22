package pkgio

import (
	"fmt"
	"sync"
)

type Data struct {
	d map[string][]byte
	m sync.RWMutex
}

func NewData() *Data {
	return &Data{
		d: map[string][]byte{},
	}
}

func (r *Data) Add(name string, b []byte) {
	r.m.Lock()
	defer r.m.Unlock()
	r.d[name] = b
}

func (r *Data) Get() map[string]string {
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
