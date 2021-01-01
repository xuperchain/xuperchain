for dir in `ls`;do
  if [ -f ${dir}/main.go ] ;then
    GOOS=js GOARCH=wasm go build -o wasm/${dir}.wasm ${dir}/main.go
  fi
done