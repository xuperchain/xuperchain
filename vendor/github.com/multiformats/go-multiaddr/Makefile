gx:
	go get github.com/whyrusleeping/gx
	go get github.com/whyrusleeping/gx-go

covertools:
	go get golang.org/x/tools/cmd/cover

deps: gx covertools
	gx --verbose install --global
	gx-go rewrite

publish:
	gx-go rewrite --undo

conformance: tmp/multiaddr
	go build -o tmp/multiaddr/test/go-multiaddr ./multiaddr
	cd tmp/multiaddr/test && MULTIADDR_BIN="./go-multiaddr" go test -v

tmp/multiaddr:
	mkdir -p tmp/
	git clone https://github.com/multiformats/multiaddr tmp/multiaddr/

clean:
	rm -rf tmp/

.PHONY: gx covertools deps publish conformance clean
