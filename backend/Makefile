.PHONY: test test-unit test-integration test-coverage test-benchmark mock-generate

# 運行所有測試
test:
	go test ./...

# 運行單元測試
test-unit:
	go test -short ./...

# 運行整合測試
test-integration:
	go test -run Integration ./...

# 生成測試覆蓋率報告
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 運行基準測試
test-benchmark:
	go test -bench=. ./...

# 生成Mock文件
mock-generate:
	mockgen -source=services/interfaces.go -destination=tests/mocks/mock_fal_service.go

# 清理測試文件
clean:
	rm -f coverage.out coverage.html
	rm -rf tests/tmp/*

# 運行測試並生成報告
test-report: test-coverage
	@echo "測試完成，覆蓋率報告已生成：coverage.html"