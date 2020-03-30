default:
	docker-compose up --force-recreate --build

test:
	docker build . -t ascheret/easerver:test -f Dockerfile.test && docker run -t --rm ascheret/easerver:test go test -v -cover ./...

tidy:
	go fmt ./... && go vet ./...