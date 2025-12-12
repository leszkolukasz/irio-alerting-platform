SUBDIRS := frontend backend
RECURSIVE_COMMANDS := format lint

.PHONY: $(RECURSIVE_COMMANDS)

GREEN   := \033[32m
YELLOW  := \033[33m
RESET   := \033[0m

check: format lint

$(RECURSIVE_COMMANDS):
	@printf "$(GREEN)Running '$@'$(RESET)\n\n"
	@for dir in $(SUBDIRS); do \
		printf "$(YELLOW)Entering project: $$dir$(RESET)\n"; \
		$(MAKE) --no-print-directory -C $$dir $@ || exit 1; \
		printf "\n"; \
	done