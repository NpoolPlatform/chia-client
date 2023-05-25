module github.com/NpoolPlatform/libent-cruder

go 1.17

require (
	entgo.io/ent v0.10.1
	github.com/NpoolPlatform/message v0.0.0-20220501030927-34f296682c0c
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.7.1-0.20210427113832-6241f9ab9942
	google.golang.org/protobuf v1.28.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.41.0

replace entgo.io/ent => entgo.io/ent v0.11.2
