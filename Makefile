
coverage:
	go test --coverprofile=cover.out
	go tool cover -func=cover.out

test:
	go test

test-unit:
	go test -short

test-integrtion:
	go test -run=Integration
