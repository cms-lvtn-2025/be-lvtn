PROTOC = protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative
GEN = go run scripts/gen_skeleton.go

# map service -> port

proto-common:
	$(PROTOC) proto/common/common.proto

proto-academic:
	$(PROTOC) proto/academic/academic.proto
	$(GEN) academic AcademicService 50051

proto-council:
	$(PROTOC) proto/council/council.proto
	$(GEN) council CouncilService 50052

proto-file:
	$(PROTOC) proto/file/file.proto
	$(GEN) file FileService 50053

proto-role:
	$(PROTOC) proto/role/role.proto
	$(GEN) role RoleService 50054

proto-thesis:
	$(PROTOC) proto/thesis/thesis.proto
	$(GEN) thesis ThesisService 50055

proto-user:
	$(PROTOC) proto/user/user.proto
	$(GEN) user UserService 50056

# Generate all services
all: proto-common proto-academic proto-council proto-file proto-role proto-thesis proto-user

# Build targets
build-academic:
	cd src/service/academic && go build -o academic .

build-council:
	cd src/service/council && go build -o council .

build-file:
	cd src/service/file && go build -o file .

build-role:
	cd src/service/role && go build -o role .

build-thesis:
	cd src/service/thesis && go build -o thesis .

build-user:
	cd src/service/user && go build -o user .

build: build-academic build-council build-file build-role build-thesis build-user

# Docker targets
docker-build-academic:
	docker build -f docker/academic.Dockerfile -t academic:latest .

docker-build-council:
	docker build -f docker/council.Dockerfile -t council:latest .

docker-build-file:
	docker build -f docker/file.Dockerfile -t file:latest .

docker-build-role:
	docker build -f docker/role.Dockerfile -t role:latest .

docker-build-thesis:
	docker build -f docker/thesis.Dockerfile -t thesis:latest .

docker-build-user:
	docker build -f docker/user.Dockerfile -t user:latest .

docker-build: docker-build-academic docker-build-council docker-build-file docker-build-role docker-build-thesis docker-build-user

# Docker Compose
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Clean targets
clean-build:
	rm -f src/service/*/academic src/service/*/council src/service/*/file
	rm -f src/service/*/role src/service/*/thesis src/service/*/user

clean-proto:
	rm -f proto/**/*.pb.go

clean-gen:
	rm -rf src/service/*/handler
	rm -rf env/*.env
	rm -rf docker/*.Dockerfile

clean: clean-build clean-proto clean-gen

# Run individual services
run-academic:
	cd src/service/academic && go run main.go

run-council:
	cd src/service/council && go run main.go

run-file:
	cd src/service/file && go run main.go

run-role:
	cd src/service/role && go run main.go

run-thesis:
	cd src/service/thesis && go run main.go

run-user:
	cd src/service/user && go run main.go

# Help
help:
	@echo "Available targets:"
	@echo "  all                - Generate all proto and services"
	@echo "  proto-<service>    - Generate specific service (academic, council, file, role, thesis, user)"
	@echo ""
	@echo "  build              - Build all services"
	@echo "  build-<service>    - Build specific service"
	@echo ""
	@echo "  docker-build       - Build all Docker images"
	@echo "  docker-build-<service> - Build specific Docker image"
	@echo ""
	@echo "  up                 - Start all services with docker-compose"
	@echo "  down               - Stop all services"
	@echo "  logs               - View logs from all services"
	@echo ""
	@echo "  run-<service>      - Run specific service locally"
	@echo ""
	@echo "  clean              - Clean all generated files"
	@echo "  clean-build        - Clean built binaries"
	@echo "  clean-proto        - Clean generated proto files"
	@echo "  clean-gen          - Clean generated handlers/env/docker files"
	@echo ""
	@echo "  help               - Show this help message"

.PHONY: all build docker-build up down logs clean clean-build clean-proto clean-gen help \
	proto-academic proto-council proto-file proto-role proto-thesis proto-user \
	build-academic build-council build-file build-role build-thesis build-user \
	docker-build-academic docker-build-council docker-build-file docker-build-role docker-build-thesis docker-build-user \
	run-academic run-council run-file run-role run-thesis run-user

