package diag

/*
import (
	"fmt"
	"sync"
)

type Recorder interface {
	Record(d Diagnostic)
	Get() Diagnostics
	Print()
}

type recorder struct {
	m        sync.RWMutex
	initOnce sync.Once
	diags    Diagnostics
}

func NewRecorder() Recorder {
	return &recorder{}
}

func (r *recorder) init() {
	r.initOnce.Do(func() {
		if r.diags == nil {
			r.diags = Diagnostics{}
		}
	})
}

func (r *recorder) Record(d Diagnostic) {
	r.init()
	r.m.Lock()
	defer r.m.Unlock()
	r.diags = append(r.diags, d.Diagnostic)
}

func (r *recorder) Get() Diagnostics {
	r.m.RLock()
	defer r.m.RUnlock()
	return r.diags
}

func (r *recorder) Print() {
	for _, d := range r.Get() {
		fmt.Println(d)
	}
}
*/
