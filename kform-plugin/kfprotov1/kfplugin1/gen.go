//go:generate protoc -I . ./kfplugin.proto --go_out=./ --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

package kfplugin1
