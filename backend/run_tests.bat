@echo off
echo Running genVideoSub Backend Tests...
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

echo Go version:
go version
echo.

REM Run unit tests
echo Running unit tests...
go test ./tests/unit/... -v
if %errorlevel% neq 0 (
    echo Unit tests failed!
    pause
    exit /b 1
)

echo.
echo Running integration tests...
go test ./tests/integration_test.go ./tests/test_setup.go -v
if %errorlevel% neq 0 (
    echo Integration tests failed!
    pause
    exit /b 1
)

echo.
echo Running production mode tests...
go test ./tests/production_mode_test.go ./tests/test_setup.go -v
if %errorlevel% neq 0 (
    echo Production mode tests failed!
    pause
    exit /b 1
)

echo.
echo Running e2e tests...
go test ./tests/e2e/... -v
if %errorlevel% neq 0 (
    echo E2E tests failed!
    pause
    exit /b 1
)

echo.
echo All tests passed successfully!
pause