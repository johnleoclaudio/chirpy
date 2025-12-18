build:
	go build -o ./bin/out

build_then_run:
	go build -o ./bin/out && ./bin/out

clean:
	rm -rf ./bin/out

run:
	go run main.go

## DB
sql_generate:
	sqlc generate

db_migrate_up:
	goose -dir ./sql/schema/ postgres "host=localhost port=5432 user=leo dbname=chirpy sslmode=disable" up

db_migrate_down:
	goose -dir ./sql/schema/ postgres "host=localhost port=5432 user=leo dbname=chirpy sslmode=disable" down
