.PHONY: lint format build push

lint:
	@pnpm lint

format:
	@pnpm format

build:
	@pnpm build

define get_bucket_name
    cd ../infra/01-bootstrap && terraform output -raw bucket_name
endef

push:
	@TARGET_BUCKET=$$($(call get_bucket_name)); \
	if [ -z "$$TARGET_BUCKET" ]; then \
		echo "Error: Bucket name not found"; \
		exit 1; \
	fi; \
	echo "Uploading static files to: gs://$$TARGET_BUCKET"; \
	gcloud storage cp -r dist/* gs://$$TARGET_BUCKET; \
	echo "Deployment complete!"