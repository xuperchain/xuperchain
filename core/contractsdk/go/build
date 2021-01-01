for contract in `ls example|grep -v call`;do
   GOOS=js GOARCH=wasm go build -o build/${contract}.wasm  github.com/xuperchain/xuperchain/core/contractsdk/go/example/$contract
done

