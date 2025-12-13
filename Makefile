build:
	go build -o ./bin/out

build_then_run:
	go build -o ./bin/out && ./bin/out

clean:
	rm -rf ./bin/out

