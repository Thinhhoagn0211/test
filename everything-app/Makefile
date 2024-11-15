DB_URL = postgresql://root:secret@0.0.0.0:5432/everything_pg?sslmode=disable
postgres:
	docker run --name postgres12 --restart=always -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16.3-alpine3.19
createdb:
	docker exec -it postgres12 createdb --username=root --owner=root everything_pg

dropdb:
	docker exec -it postgres12 dropdb everything_pg
gen:
	protoc --proto_path=file-index/proto file-index/proto/*.proto --go_out=file-index/pb --go-grpc_out=file-index/pb
clean:
	rm pb/*.go

migrateup:
	migrate --path db/migration/ --database "$(DB_URL)" --verbose up
migratedown:
	migrate --path db/migration/ --database "$(DB_URL)" --verbose down
sqlc:
	sqlc generate
client:
	go run file-index/cmd/client/main.go -address 192.168.137.130:8080
server:
	go run file-index/cmd/server/main.go -port 8080
api:
	go run file-search/main.go
cert:
	cd cert; ./gen.sh; cd ..
build:
	docker compose up --build -d
swagger:
	swag init -g file-search/api/server.go -o file-search/docs/swagger
up:
	docker compose up -d
down:
	docker compose down --remove-orphans
delete:
	docker compose down --remove-orphans -v
evans:
	evans --host 192.168.137.130 --port 8080 -r repl
	
.PHONY: gen clean client server cert up down delete migrateup migratedown