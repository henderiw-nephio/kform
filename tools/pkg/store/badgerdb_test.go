package store

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"testing"
)

var testStore = "teststore"

func Test_readKeyItems(t *testing.T) {
	type args struct {
		k []byte
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no_tags",
			args: args{
				k: []byte{
					0x2, 0x61, 0x62,
					0x3, 0x62, 0x62, 0x62,
					0x1, 0x63,
				},
			},
			want: []string{"ab", "bbb", "c"},
		},
		{
			name: "with_tags",
			args: args{
				k: []byte{
					0x2, 0x61, 0x62,
					0x3, 0x62, 0x62, 0x62,
					0x1, 0x63,
					0x2, 0x65, 0x66,
					0x1, 0x65,
					0x3, 0x65, 0x66, 0x67,
				},
			},
			want: []string{"ab", "bbb", "c", "ef", "e", "efg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readKeyItems(tt.args.k); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readKeyItems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_keyToPlugin(t *testing.T) {
	type args struct {
		k []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Plugin
		wantErr bool
	}{
		{
			name: "no_tags",
			args: args{
				k: []byte{
					0x2, 0x61, 0x62,
					0x3, 0x62, 0x62, 0x62,
					0x1, 0x63,
					0x1, 0x64,
					0x1, 0x65,
				},
			},
			want: Plugin{
				Project: "ab",
				Name:    "bbb",
				Os:      "c",
				Arch:    "d",
				Version: "e",
				Tags:    map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "with_tags",
			args: args{
				k: []byte{
					0x2, 0x61, 0x62,
					0x3, 0x62, 0x62, 0x62,
					0x1, 0x63,
					0x1, 0x63,
					0x1, 0x63,
					0x3, 0x65, 0x3d, 0x66,
					0x3, 0x66, 0x3d, 0x65,
					0x4, 0x65, 0x66, 0x3d, 0x67,
				},
			},
			want: Plugin{
				Project: "ab",
				Name:    "bbb",
				Os:      "c",
				Arch:    "c",
				Version: "c",
				Tags:    map[string]string{"e": "f", "f": "e", "ef": "g"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keyToPlugin(tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("keyToPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keyToPlugin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_badgerDBStore_Save(t *testing.T) {
	type args struct {
		ctx context.Context
		p   Plugin
		v   []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				ctx: context.TODO(),
				p: Plugin{
					Project: "proj",
					Name:    "plugin1",
					Os:      "Linux",
					Arch:    "x86",
					Version: "v0.1.0",
					Tags:    map[string]string{},
				},
				v: bytes.Repeat([]byte{1}, 1024*1024*100), // 100MB
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := newBadgerDBStore(tt.args.ctx, testStore)
			if err != nil {
				t.Errorf("Failed to create store: %v", err)
				t.Fail()
			}
			defer func() {
				d.Close()
				os.RemoveAll(testStore)
			}()
			if err := d.Save(tt.args.ctx, tt.args.p, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("badgerDBStore.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_badgerDBStore_Get(t *testing.T) {
	v := bytes.Repeat([]byte{1}, 1024*1024*100)
	type args struct {
		ctx context.Context
		p   Plugin
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				ctx: context.TODO(),
				p: Plugin{
					Project: "proj",
					Name:    "plugin1",
					Os:      "Linux",
					Arch:    "x86",
					Version: "v0.1.0",
					Tags:    map[string]string{},
				},
				// v: bytes.Repeat([]byte{1}, 1024*1024*100), // 100MB
			},
			want:    v,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := newBadgerDBStore(tt.args.ctx, testStore)
			if err != nil {
				t.Errorf("Failed to create store: %v", err)
				t.Fail()
			}
			defer func() {
				d.Close()
				os.RemoveAll(testStore)
			}()

			if err := d.Save(tt.args.ctx, tt.args.p, v); (err != nil) != tt.wantErr {
				t.Errorf("badgerDBStore.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := d.Get(tt.args.ctx, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("badgerDBStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("badgerDBStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
