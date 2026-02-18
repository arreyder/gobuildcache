.PHONY: all build test clean test-integration test-s3-local

# Binary name
BINARY_NAME=gobuildcache

# Build directory
BUILD_DIR=./builds

all: build test

# Build the cache program
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run tests with the cache program
test-manual: build
	@echo "Running tests with cache program..."
	GOCACHEPROG="$(shell pwd)/$(BUILD_DIR)/$(BINARY_NAME)" DEBUG=true go test -v ./tests

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -rf $(BUILD_DIR)/cache

# Run the cache server directly
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

# Clear the cache
clear: build
	$(BUILD_DIR)/$(BINARY_NAME) clear

test:
	@echo "Running short tests..."
	go test -short -count 1 -v -race ./...

test-long:
	@echo "Running short and longer tests..."
	TEST_S3_BUCKET=test-go-build-cache AWS_ENDPOINT_URL_S3=https://t3.storage.dev AWS_ACCESS_KEY_ID=tid_GHaEn_WOoPpmoblCBaWQCWolCHEaZnRWKYYiGdpgWuuvEpUgaI AWS_SECRET_ACCESS_KEY=tsec_iT8vZcCZwmc17pmdovhN4YBx5k5KXUbdVfqbbMuYky+SBpK8mRlANt_0dR2O87+9I0HSfv AWS_REGION=auto go test -count 1 -v -race ./...

# Local S3 integration test using MinIO
MINIO_CONTAINER=gobuildcache-minio
MINIO_PORT=9000
MINIO_BUCKET=test-go-build-cache

test-s3-local:
	@echo "Starting MinIO container..."
	@docker rm -f $(MINIO_CONTAINER) 2>/dev/null || true
	@docker run -d --name $(MINIO_CONTAINER) \
		-p $(MINIO_PORT):9000 \
		minio/minio server /data
	@echo "Waiting for MinIO to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		docker exec $(MINIO_CONTAINER) mc alias set local http://localhost:9000 minioadmin minioadmin 2>/dev/null && break; \
		echo "  attempt $$i..."; \
		sleep 1; \
	done
	@echo "Creating test bucket..."
	@docker exec $(MINIO_CONTAINER) mc mb local/$(MINIO_BUCKET) 2>/dev/null || true
	@echo "Running S3 integration tests against local MinIO..."
	@TEST_S3_BUCKET=$(MINIO_BUCKET) \
		AWS_ENDPOINT_URL_S3=http://localhost:$(MINIO_PORT) \
		AWS_ACCESS_KEY_ID=minioadmin \
		AWS_SECRET_ACCESS_KEY=minioadmin \
		AWS_REGION=us-east-1 \
		GOBUILDCACHE_S3_PATH_STYLE=true \
		go test -count 1 -v -race -run TestCacheIntegrationS3 ./...; \
		EXIT_CODE=$$?; \
		echo "Cleaning up MinIO container..."; \
		docker rm -f $(MINIO_CONTAINER) 2>/dev/null || true; \
		exit $$EXIT_CODE

