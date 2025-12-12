SUBDIRS := frontend backend/services/api # backend/common backend/services/scheduler backend/services/worker backend/services/incident-manager
RECURSIVE_COMMANDS := format build lint clean

.PHONY: $(RECURSIVE_COMMANDS)

GREEN   := \033[32m
YELLOW  := \033[33m
RESET   := \033[0m

$(RECURSIVE_COMMANDS):
	@for dir in $(SUBDIRS); do \
		printf "$(YELLOW)Entering: $$dir$(RESET)\n"; \
		$(MAKE) --no-print-directory -C $$dir $@ || exit 1; \
	done