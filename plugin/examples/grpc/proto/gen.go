//go:generate protoc -I . ./kv.proto --go_out=./ --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

package proto
