package store

import (
	"context"
	"errors"
	"fmt"
)

type Store interface {
	// Get returns a plugin bytes
	Get(ctx context.Context, p Plugin) ([]byte, error)
	// Save stores a plugin in the store
	Save(ctx context.Context, p Plugin, v []byte) error
	// Delete removes a plugin from the store
	Delete(ctx context.Context, p Plugin) error
	// List returns the list of plugin matching certain criteria
	List(ctx context.Context, p Plugin) ([]Plugin, error)
	// Reset deletes all the plugins from the store
	Reset(ctx context.Context) error
	// Close closes the store client
	Close() error
}

func New(ctx context.Context, loc string) (Store, error) {
	return newBadgerDBStore(ctx, loc)
}

type Plugin struct {
	Project string // or namespace
	Name    string // name
	Os      string // os
	Arch    string // arch
	Version string // version
	Tags    map[string]string
}

const (
	maxTagsSize = 5000 // TODO:
)

func (p Plugin) Validate() error {
	if p.Project == "" {
		return errors.New("project name is mandatory")
	}
	if len(p.Project) > 0xFF {
		return fmt.Errorf("project name is too long: (%d> %d)", len(p.Project), 0xFF)
	}
	if p.Name == "" {
		return errors.New("plugin name is mandatory")
	}
	if p.Version == "" {
		return errors.New("plugin version is mandatory")
	}
	tagsLen := 0
	for k, v := range p.Tags {
		if k == "" || v == "" {
			return errors.New("invalid empty tag")
		}
		tag := k + "=" + v
		ltag := len([]byte(tag))
		if ltag > 0xFF {
			return fmt.Errorf("tag too long(<%d)", 0xFF)
		}
		tagsLen += len([]byte(tag)) + 1 // len + 1(len)
	}
	if tagsLen > maxTagsSize { // TODO:
		return fmt.Errorf("total tags size too big")
	}
	return nil
}
