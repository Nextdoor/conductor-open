SHELL=/bin/bash
SHA1 := $(shell git rev-parse --short HEAD)

DOCKER_IMAGE ?= conductor
DOCKER_REGISTRY ?= hub.docker.com
DOCKER_NAMESPACE ?= nextdoor

TARGET_DOCKER_NAME := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE):$(DOCKER_TAG)

GO_DIRS=core cmd services shared

.PHONY: all imports test glide

all: imports docker-build docker-run

imports:
	@goimports -local github.com -w $(GO_DIRS)

test:
	@./test.sh

glide:
	@echo "Installing Go Dependencies"
	@which glide || curl https://glide.sh/get | sh
	glide install

DOCKER_BUILD_ARGS := -t $(DOCKER_IMAGE)

ifdef CACHE_FROM
DOCKER_BUILD_ARGS := --cache-from $(CACHE_FROM) $(DOCKER_BUILD_ARGS)
endif

define DOCKER_RUN_ARGS
--name $(DOCKER_IMAGE) \
--env LOGLEVEL=DEBUG \
--env-file envfile \
--volume $(shell pwd)/resources/frontend:/app/frontend \
--publish 80:80 \
--publish 443:443 \
--link conductor-postgres \
--hostname conductor-dev
endef

# Check if interactive shell.
INTERACTIVE = $(shell [ "`tty`" != "not a tty" ] && echo true || echo false)
ifeq ($(INTERACTIVE),true)
DOCKER_RUN_INTERACTIVE_ARGS = -it
endif

.PHONY: docker-build docker-run docker-stop docker-logs docker-tag docker-push docker-login docker-populate-cache

docker-build:
	@echo "Building Conductor Docker image"
	docker build $(DOCKER_BUILD_ARGS) .

docker-run: docker-stop
	@echo "Running $(DOCKER_IMAGE)"
	[ -e envfile ] || touch envfile
	docker run $(DOCKER_RUN_ARGS) $(DOCKER_RUN_INTERACTIVE_ARGS) $(DOCKER_IMAGE)

docker-stop:
	@echo "Stopping $(DOCKER_IMAGE)"
	@docker rm -f $(DOCKER_IMAGE) 2>/dev/null \
		|| echo "No existing container running"

docker-logs:
	@echo "Running $(DOCKER_IMAGE)"
	docker logs -f $(DOCKER_IMAGE)

docker-tag:
	@echo "Tagging $(DOCKER_IMAGE) as $(TARGET_DOCKER_NAME)"
	docker tag -f $(DOCKER_IMAGE) $(TARGET_DOCKER_NAME)

docker-push: docker-tag
	@echo "Pushing $(DOCKER_IMAGE) to $(TARGET_DOCKER_NAME)"
	docker push $(TARGET_DOCKER_NAME)

docker-login:
	@echo "Logging into $(DOCKER_REGISTRY)"
	@docker login \
		-u $(DOCKER_USER) \
		-p "$(value DOCKER_PASS)" $(DOCKER_REGISTRY)

docker-populate-cache:
	@echo "Attempting to download $(DOCKER_IMAGE)"
	@docker pull "$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(DOCKER_IMAGE)" && \
		docker images -a || exit 0

.PHONY: frontend

frontend:
	$(MAKE) -C frontend

PGDB=conductor
PGHOST=localhost
PGPORT=5432
PGUSER=conductor
PGPASS=conductor
PGDATA=/var/lib/postgresql/data/conductor

define PG_ARGS
--name conductor-postgres \
--publish 5432:5432 \
--env POSTGRES_USER=$(PGUSER) \
--env POSTGRES_PASSWORD=$(PGPASS) \
--env POSTGRES_DB=$(PGDB) \
--env PGDATA=$(PGDATA) \
--detach
endef

.PHONY: postgres postgres-perm psql postgres-wipe test-data

postgres:
	docker rm -f conductor-postgres || true
	docker run $(PG_ARGS) postgres

postgres-perm:
	docker rm -f conductor-postgres || true
	docker run $(PG_ARGS) -v $$HOME/data/conductor:$(PGDATA) postgres

postgres-wipe:
	PGPASSWORD=conductor dropdb -h localhost -U conductor conductor || true
	PGPASSWORD=conductor createdb -h localhost -U conductor conductor || true

psql:
	PGPASSWORD=$(PGPASS) \
		psql \
		-h $(PGHOST) \
		-p $(PGPORT) \
		-d $(PGDB) \
		-U $(PGUSER)

test-data: postgres-wipe
	export POSTGRES_HOST=localhost; \
	set -a; \
	if [[ -e testenv ]]; then \
		source testenv; \
	fi; \
	go run cmd/test_data.go

# README.md manipulation

.PHONY: gravizool readme edit-readme

gravizool:
	which gravizool || go get github.com/swaggy/gravizool && go get github.com/swaggy/gravizool

readme: gravizool
	gravizool -b=false -e README.md

edit-readme: gravizool
	gravizool -b=false -d README.md
