test:
	go test ./... -v -cover

cover:
	go test ./... -coverprofile=cover.out
	go tool cover -html=cover.out

validate_version:
ifndef VERSION
	$(error VERSION is undefined)
endif

docker_build: validate_version
	docker build \
		-t codingconcepts/cdch:${VERSION} \
		--build-arg version=${VERSION} \
		.

docker_push: docker_build
	docker push codingconcepts/cdch:${VERSION}
	docker tag codingconcepts/cdch:${VERSION} codingconcepts/cdch:latest
	docker push codingconcepts/cdch:latest

release: validate_version
	- mkdir releases

	# linux (amd)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o cdch cmd/cdch/cdch.go ;\
	tar -zcvf ./releases/cdch_${VERSION}_linux_amd64.tar.gz ./cdch ;\

	# macos (arm)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=${VERSION}" -o cdch cmd/cdch/cdch.go ;\
	tar -zcvf ./releases/cdch_${VERSION}_macos_arm64.tar.gz ./cdch ;\

	# macos (amd)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o cdch cmd/cdch/cdch.go ;\
	tar -zcvf ./releases/cdch_${VERSION}_macos_amd64.tar.gz ./cdch ;\

	# windows (amd)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${VERSION}" -o cdch cmd/cdch/cdch.go ;\
	tar -zcvf ./releases/cdch_${VERSION}_windows_amd64.tar.gz ./cdch ;\

	rm ./cdch

teardown:
	docker ps -aq | xargs docker rm -f