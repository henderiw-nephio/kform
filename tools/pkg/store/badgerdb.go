package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

type badgerDBStore struct {
	db *badger.DB
}

func newBadgerDBStore(ctx context.Context, p string) (Store, error) {
	d := &badgerDBStore{}
	var err error
	d.db, err = d.openDB(ctx, p)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *badgerDBStore) Get(ctx context.Context, p Plugin) ([]byte, error) {
	err := p.Validate()
	if err != nil {
		return nil, err
	}
	p.Tags = nil // Get ignores tags (for now)
	k := buildKey(p)
	var v []byte

	err = d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.PrefetchSize = 1
		var err error

		it := txn.NewIterator(opts)
		defer it.Close()

		it.Seek(k)
		v, err = it.Item().ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (d *badgerDBStore) Save(ctx context.Context, p Plugin, v []byte) error {
	err := p.Validate()
	if err != nil {
		return err
	}
	k := buildKey(p)
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
}

func (d *badgerDBStore) Delete(ctx context.Context, p Plugin) error {
	err := p.Validate()
	if err != nil {
		return err
	}
	k := buildKey(p)
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(k)
	})
}

func (d *badgerDBStore) List(ctx context.Context, p Plugin) ([]Plugin, error) {
	rs := make([]Plugin, 0)
	if p.Project == "" {
		projects, err := d.listProjects()
		if err != nil {
			return nil, err
		}
		for _, pj := range projects {
			prs, err := d.List(ctx, Plugin{
				Project: pj,
				Name:    p.Name,
				Version: p.Version,
				Tags:    p.Tags,
			})
			if err != nil {
				return nil, err
			}
			rs = append(rs, prs...)
		}
		return rs, nil
	}

	var prefix []byte
	switch {
	case p.Name == "" && p.Version == "":
		prefix = buildKeyItem(p.Project)
	case p.Name != "" && p.Version == "":
		prefix = buildKeyItem(p.Project)
		prefix = append(prefix, buildKeyItem(p.Name)...)
	case p.Name != "" && p.Version != "":
		prefix = buildKeyItem(p.Project)
		prefix = append(prefix, buildKeyItem(p.Name)...)
		prefix = append(prefix, buildKeyItem(p.Version)...)
	case p.Name == "" && p.Version != "":
		return nil, errors.New("cannot list plugins by version without name")
	}

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			k := it.Item().KeyCopy(nil)
			cp, err := keyToPlugin(k)
			if err != nil {
				return err
			}
			for k, v := range p.Tags {
				if vv, ok := cp.Tags[k]; ok {
					if v == vv {
						rs = append(rs, cp)
					}
				}
			}

		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (d *badgerDBStore) Reset(ctx context.Context) error {
	return d.db.DropAll()
}

func (d *badgerDBStore) Close() error {
	return d.db.Close()
}

func (d *badgerDBStore) openDB(ctx context.Context, name string) (*badger.DB, error) {
	opts := badger.DefaultOptions(name).
		WithLoggingLevel(badger.WARNING)
		// WithCompression(options.None)
		// WithBlockCacheSize(0)

	bdb, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return bdb, nil
}

func (d *badgerDBStore) listProjects() ([]string, error) {
	projectsMap := make(map[string]struct{})
	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			proj := getProjectName(k)
			if _, ok := projectsMap[proj]; !ok {
				projectsMap[proj] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	projects := make([]string, 0, len(projectsMap))
	for p := range projectsMap {
		projects = append(projects, p)
	}
	sort.Strings(projects)
	return projects, nil
}

//		list of length(1 byte) + item
//	    +----------------------------------------------------------------------------------------
//		| +-------------+----------+-------------+--------+----------+----------+----------------
//		| | l | project | l | name | l | version | l | os | l | arch | l | tag1 | l | tag1 | ....
//		| +-------------+----------+-------------+--------+----------+----------+----------+-----
//		+----------------------------------------------------------------------------------------
func buildKey(p Plugin) []byte {
	b := new(bytes.Buffer)
	// project
	b.Write(buildKeyItem(p.Project))
	// name
	b.Write(buildKeyItem(p.Name))
	// os
	b.Write(buildKeyItem(p.Os))
	// arch
	b.Write(buildKeyItem(p.Arch))
	// version
	b.Write(buildKeyItem(p.Version))
	// tags
	tags := make([]string, 0, len(p.Tags))
	for k, v := range p.Tags {
		tags = append(tags, k+"="+v)
	}
	sort.Strings(tags)
	for _, tag := range tags {
		b.Write(buildKeyItem(tag))
	}
	return b.Bytes()
}

func keyToPlugin(k []byte) (Plugin, error) {
	items := readKeyItems(k)
	if len(items) < 5 {
		return Plugin{}, fmt.Errorf("failed to read plugin key: key too short")
	}

	p := Plugin{
		Project: items[0],
		Name:    items[1],
		Os:      items[2],
		Arch:    items[3],
		Version: items[4],
		Tags:    map[string]string{},
	}

	for _, t := range items[5:] {
		if i := strings.Index(t, "="); i > 0 {
			p.Tags[t[:i]] = t[i+1:]
			continue
		}
		return Plugin{}, fmt.Errorf("invalid tag %s in key", t)
	}
	return p, nil
}

func getProjectName(k []byte) string {
	lk := len(k)
	if lk < 2 {
		return ""
	}
	projLen := k[0]
	if lk < 1+int(projLen) {
		return ""
	}
	proj := string(k[1 : projLen+1])
	return proj
}

func buildKeyItem(i string) []byte {
	l := len(i)
	if l > 0xFF {
		return nil
	}

	b := make([]byte, l+1)
	b[0] = uint8(l)

	copy(b[1:], []byte(i))
	return b
}

func readKeyItems(k []byte) []string {
	rs := make([]string, 0, 5)
	lk := uint(len(k))
	offset := uint(0)
	for {
		l := uint(k[offset])
		offset++
		v := k[offset : offset+l]
		rs = append(rs, string(v))
		offset += l
		if offset+1 >= lk {
			break
		}
	}
	return rs
}
