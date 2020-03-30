default:
	docker-compose up --force-recreate --build

test:
	docker build . -t ascheret/easerver:test --target build && docker run -a STDOUT -a STDERR --rm ascheret/easerver:test go test -v -cover ./...