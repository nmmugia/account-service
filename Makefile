include .env
export $(shell sed 's/=.*//' .env)

start:
	@go run src/main.go
lint:
	@golangci-lint run
tests:
	@go test -v ./test/...
tests-%:
	@go test -v ./test/... -run=$(shell echo $* | sed 's/_/./g')
testsum:
	@cd test && gotestsum --format testname
swagger:
	@cd src && swag init
migration-%:
	@migrate create -ext sql -dir src/database/migrations create-table-$(subst :,_,$*)
migrate-up:
	@migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -path src/database/migrations up
migrate-down:
	@migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -path src/database/migrations down
migrate-docker-up:
	docker run -v ./src/database/migrations:/migrations --network account-service_account-network migrate/migrate -path=/migrations/ -database postgres://$(DB_USER):$(DB_PASSWORD)@postgresdb:$(DB_PORT)/$(DB_NAME)?sslmode=disable up
migrate-docker-down:
	@docker run -v ./src/database/migrations:/migrations --network account-service_account-network migrate/migrate -path=/migrations/ -database postgres://$(DB_USER):$(DB_PASSWORD)@postgresdb:$(DB_PORT)/$(DB_NAME)?sslmode=disable down -all
docker:
	@chmod -R 755 ./src/database/init
	@docker-compose up --build
docker-test:
	docker compose -f docker-compose.yml -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from account-service
docker-down:
	@docker-compose down --rmi all --volumes --remove-orphans
docker-cache:
	@docker builder prune -f
test-unit-coverage:
	@go test -v -coverprofile=coverage.out -coverpkg=account-service/src/... ./test/unit/...
	@go tool cover -html=coverage.out -o coverage.html
test-integration-coverage:
	@go test -v -tags integration -coverprofile=coverage.out -coverpkg=account-service/src/... ./test/integration
	@go tool cover -html=coverage.out -o coverage.html
test-all-coverage:
	@go test -v -tags integration -coverprofile=coverage.out -coverpkg=account-service/src/... ./test/... -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html