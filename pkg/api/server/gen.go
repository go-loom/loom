package server

//go:generate rm ./service.pb.go
//go:generate rm ./service.twirp.go
//go:generate retool do  protoc --twirp_out=. --go_out=. ./service.proto
