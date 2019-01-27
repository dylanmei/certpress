default: build

test:
	go test -v cover

build: clean
	go build -v -o bin/certpress .

clean:
	rm -rf bin/
