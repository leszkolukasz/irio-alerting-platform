MODULES = ./common ./services/api ./services/incident-manager ./services/logger ./services/notifier ./services/scheduler ./services/worker

.PHONY: lint format

GREEN   := \033[32m
YELLOW  := \033[33m
RESET   := \033[0m

define run_on_modules
	@for module in $(MODULES); do \
		printf "$(YELLOW)↳  Entering module: $$module$(RESET)\n"; \
		(cd $$module && $1) || exit 1; \
	done
endef

define get_registry
	cd ../infra/01-bootstrap && terraform output -raw registry_url
endef


define build_image
    printf "$(YELLOW) Building image for service: $1 with tag: $2$(RESET)\n"; \
	REGISTRY=$$($(call get_registry)); \
	if [ -z "$$REGISTRY" ]; then \
		echo "Error: Docker registry not found"; \
		exit 1; \
	fi; \
    docker build -f ./services/$1/Dockerfile . -t $${REGISTRY}/alerting-platform-$1:$2
endef

define push_image
    printf "$(YELLOW) Pushing image for service: $1 with tag: $2$(RESET)\n"; \
	REGISTRY=$$($(call get_registry)); \
	if [ -z "$$REGISTRY" ]; then \
		echo "Error: Docker registry not found"; \
		exit 1; \
	fi; \
    docker push $${REGISTRY}/alerting-platform-$1:$2
endef

lint:
	$(call run_on_modules, golangci-lint run --fix)

format:
	$(call run_on_modules,  golangci-lint fmt)

test:
	$(call run_on_modules, go test ./... -v)

bump:
	@perl -i -pe 's/app_version = "v(\d+)\.(\d+)\.(\d+)"/"app_version = \"v$$1.".($$2+1).".0\""/e' ../infra/03-helm/terraform.tfvars
	@sed -i "s/build_time = \".*\"/build_time = \"$$(date '+%Y-%m-%d %H:%M')\"/" ../infra/03-helm/terraform.tfvars

build:
	@VERSION=$$(grep 'app_version' ../infra/03-helm/terraform.tfvars | awk -F'"' '{print $$2}'); \
	printf "$(GREEN) Building Docker images with version: $$VERSION$(RESET)\n"; \
	$(call build_image,api,$$VERSION); \
	$(call build_image,incident-manager,$$VERSION); \
	$(call build_image,logger,$$VERSION); \
	$(call build_image,notifier,$$VERSION); \
	$(call build_image,scheduler,$$VERSION); \
	$(call build_image,worker,$$VERSION)

push:
	@VERSION=$$(grep 'app_version' ../infra/03-helm/terraform.tfvars | awk -F'"' '{print $$2}'); \
	printf "$(GREEN) Pushing Docker images with version: $$VERSION$(RESET)\n"; \
	$(call push_image,api,$$VERSION); \
	$(call push_image,incident-manager,$$VERSION); \
	$(call push_image,logger,$$VERSION); \
	$(call push_image,notifier,$$VERSION); \
	$(call push_image,scheduler,$$VERSION); \
	$(call push_image,worker,$$VERSION)