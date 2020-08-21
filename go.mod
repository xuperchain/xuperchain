module github.com/xuperchain/xuperchain

go 1.12

require github.com/xuperchain/xupercore v0.0.0

replace github.com/xuperchain/xupercore => ../xupercore

require (
	github.com/aws/aws-sdk-go v1.34.5
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/dgraph-io/badger/v2 v2.0.0-rc.2
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway v1.9.2
	github.com/hashicorp/golang-lru v0.5.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/syndtr/goleveldb v1.0.0
	github.com/xuperchain/log15 v0.0.0-20190620081506-bc88a9198230
	golang.org/x/crypto v0.0.0-20190927123631-a832865fa7ad
	google.golang.org/grpc v1.24.0
)
