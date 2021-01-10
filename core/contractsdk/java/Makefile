build: build-sdk build-example
test: test-sdk test-example
clean: clean-sdk clean-example
build-sdk:
	mvn package
	mvn install 

build-example:
	make -C example  build 

test-sdk:
	echo "TBD"
test-example:
	make -C example test

clean-sdk:
	mvn clean
clean-example:
	make -C example clean