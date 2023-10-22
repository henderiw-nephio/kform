//go:generate protoc -I . ./grpc_stdio.proto ./grpc_controller.proto ./grpc_broker.proto --go_out=./ --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative

package plugin
