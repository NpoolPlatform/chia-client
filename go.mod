module github.com/NpoolPlatform/libent-cruder

go 1.17

require (
	entgo.io/ent v0.10.1
	github.com/google/uuid v1.3.0
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.41.0

replace entgo.io/ent => entgo.io/ent v0.11.2
