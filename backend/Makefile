MODULES = ./common ./services/api

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

lint:
	$(call run_on_modules, golangci-lint run --fix)

format:
	$(call run_on_modules,  golangci-lint fmt)