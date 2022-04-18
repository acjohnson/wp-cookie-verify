FULLTAG=wp-cookie-verify:latest
DOCKERFILE=Dockerfile
all: build

build:
	go build
	docker build -t $(FULLTAG) -f $(DOCKERFILE) .
push: build
	docker push $(FULLTAG)
