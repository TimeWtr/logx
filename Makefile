.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: ut
ut:
	@go test -race ./...

.PHONY: clean
clean:
	@rm -f logx.test
	@cd logs && rm -rf *


.PHONY: check
check:
	@$(MAKE) --no-print-directory tidy
	@#$(MAKE) --no-print-directory ut