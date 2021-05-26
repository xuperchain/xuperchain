module github.com/xuperchain/xuperchain

go 1.14

require (
	github.com/ChainSafe/go-schnorrkel v0.0.0-20200626160457-b38283118816 // indirect
	github.com/antihax/optional v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.32.4 // indirect
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/dgraph-io/badger/v2 v2.0.0-rc.2 // indirect
	github.com/docker/go-connections v0.4.1-0.20180821093606-97c2040d34df // indirect
	github.com/emirpasic/gods v1.12.1-0.20201118132343-79df803e554c // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fsouza/go-dockerclient v1.6.0 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.2-0.20200707131729-196ae77b8a26 // indirect
	github.com/google/gofuzz v1.1.1-0.20200604201612-c04b05f3adfa // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.3
	github.com/hyperledger/burrow v0.30.5
	github.com/ipfs/go-ipfs-addr v0.0.1 // indirect
	github.com/libp2p/go-libp2p-kad-dht v0.8.2 // indirect
	github.com/libp2p/go-libp2p-noise v0.1.1 // indirect
	github.com/libp2p/go-tcp-transport v0.2.1 // indirect
	github.com/manifoldco/promptui v0.7.0
	github.com/miekg/dns v1.1.31 // indirect
	github.com/multiformats/go-multiaddr v0.3.1 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca
	github.com/xuperchain/crypto v0.0.0-20201028025054-4d560674bcd6
	github.com/xuperchain/log15 v0.0.0-20190620081506-bc88a9198230
	github.com/xuperchain/xupercore v0.0.0-20210525054057-4162b6943567
	github.com/xuperchain/xvm v0.0.0-20210126142521-68fd016c56d7 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/mod v0.1.1-0.20191209134235-331c550502dd // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200824131525-c12d262b63d8 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.32.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace github.com/hyperledger/burrow => github.com/xuperchain/burrow v0.30.6-0.20210115120720-3da1be35a1e2
