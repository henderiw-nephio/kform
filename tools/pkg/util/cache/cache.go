/*
Copyright 2022 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

const (
	// errors
	NotFound = "not found"
)

type Cache[T1 any] interface {
	Get(NSN) (T1, error)
	Add(context.Context, NSN, T1) error
	Upsert(context.Context, NSN, T1)
	Delete(context.Context, NSN)
	List() map[NSN]T1
	AddWatch(fn ResourceCallbackFn)
}

func New[T1 any]() Cache[T1] {
	return &cache[T1]{
		db:         map[NSN]T1{},
		callbackFn: []ResourceCallbackFn{},
	}
}

type cache[T1 any] struct {
	m          sync.RWMutex
	db         map[NSN]T1
	callbackFn []ResourceCallbackFn
}

// Get return the type
func (r *cache[T1]) Get(nsn NSN) (T1, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	x, ok := r.db[nsn]
	if !ok {
		return *new(T1), fmt.Errorf("%s, nsn: %s", NotFound, nsn.String())
	}
	return x, nil
}

func (r *cache[T1]) List() map[NSN]T1 {
	r.m.RLock()
	defer r.m.RUnlock()

	l := map[NSN]T1{}
	for nsn, x := range r.db {
		l[nsn] = x
	}
	return l
}

func (r *cache[T1]) Add(ctx context.Context, nsn NSN, newd T1) error {
	// if an error is returned the entry already exists
	if _, err := r.Get(nsn); err == nil {
		return fmt.Errorf("duplicate entry %v", nsn)
	}
	// update the cache before calling the callback since the cb fn will use this data
	r.update(nsn, newd)

	for _, cb := range r.callbackFn {
		cb(ctx, AddAction, nsn, newd)
	}
	return nil
}

// Upsert creates or updates the entry in the cache
func (r *cache[T1]) Upsert(ctx context.Context, nsn NSN, newd T1) {
	exists := true
	oldd, err := r.Get(nsn)
	if err != nil {
		exists = false
	}
	// update the cache before calling the callback since the cb fn will use this data
	r.update(nsn, newd)

	// call callback if data got changed or if no data exists
	if exists {
		if !reflect.DeepEqual(oldd, newd) {
			for _, cb := range r.callbackFn {
				cb(ctx, UpsertAction, nsn, newd)
			}
		}
	} else {
		for _, cb := range r.callbackFn {
			cb(ctx, AddAction, nsn, newd)
		}
	}

}

func (r *cache[T1]) update(nsn NSN, newd T1) {
	r.m.Lock()
	defer r.m.Unlock()
	r.db[nsn] = newd
}

func (r *cache[T1]) delete(nsn NSN) {
	r.m.Lock()
	defer r.m.Unlock()
	delete(r.db, nsn)
}

// Delete deletes the entry in the cache
func (r *cache[T1]) Delete(ctx context.Context, nsn NSN) {
	// only if an exisitng object gets deleted we
	// call the registered callbacks
	exists := true
	d, err := r.Get(nsn)
	if err != nil {
		exists = false
	}
	// if exists call the callback
	if exists {
		for _, cb := range r.callbackFn {
			cb(ctx, DeleteAction, nsn, d)
		}
	}
	// delete the entry to ensure the cb uses the proper data
	r.delete(nsn)

}

func (r *cache[T1]) AddWatch(fn ResourceCallbackFn) {
	r.m.Lock()
	defer r.m.Unlock()
	found := false
	for _, cb := range r.callbackFn {
		if reflect.ValueOf(cb).Pointer() == reflect.ValueOf(fn).Pointer() {
			found = true
		}
	}
	if !found {
		r.callbackFn = append(r.callbackFn, fn)
	}
}
