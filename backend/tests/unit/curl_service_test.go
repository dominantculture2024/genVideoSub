package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"genVideoSub/interfaces"
	"genVideoSub/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCurlService 測試cURL服務
func TestCurlService(t *testing.T) {
	service := services.NewCurlService()
	require.NotNil(t, service)
}

// TestGenerateCurlCommand 測試生成cURL命令
func TestGenerateCurlCommand(t *testing.T) {
	service := services.NewCurlService()
	
	tests := []struct {
		name     string
		request  *interfaces.CurlRequest
		expected []string // 預期包含的字符串
	}{
		{
			name: "simple_get_request",
			request: &interfaces.CurlRequest{
				URL:    "https://api.example.com/test",
				Method: "GET",
			},
			expected: []string{"curl", "https://api.example.com/test"},
		},
		{
			name: "post_request_with_headers",
			request: &interfaces.CurlRequest{
				URL:    "https://api.example.com/test",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": "Bearer token123",
				},
				Body:    `{"test": "data"}`,
				Timeout: 30,
			},
			expected: []string{
				"curl",
				"-X POST",
				"Content-Type: application/json",
				"Authorization: Bearer token123",
				`-d "{\"test\": \"data\"}"`,
				"--max-time 30",
				"https://api.example.com/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := service.GenerateCurlCommand(tt.request)
			for _, expected := range tt.expected {
				assert.Contains(t, cmd, expected)
			}
		})
	}
}

// TestParseCurlCommand 測試解析cURL命令
func TestParseCurlCommand(t *testing.T) {
	service := services.NewCurlService()

	tests := []struct {
		name     string
		curlCmd  string
		expected *interfaces.CurlRequest
		wantErr  bool
	}{
		{
			name:    "simple_get",
			curlCmd: "curl https://api.example.com/test",
			expected: &interfaces.CurlRequest{
				URL:    "https://api.example.com/test",
				Method: "GET",
			},
		},
		{
			name:    "post_with_data",
			curlCmd: `curl -X POST -H "Content-Type: application/json" -d '{"test": "data"}' https://api.example.com/test`,
			expected: &interfaces.CurlRequest{
				URL:    "https://api.example.com/test",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"test": "data"}`,
			},
		},
		{
			name:    "invalid_command",
			curlCmd: "invalid command",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := service.ParseCurlCommand(tt.curlCmd)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected.URL, req.URL)
			assert.Equal(t, tt.expected.Method, req.Method)
			assert.Equal(t, tt.expected.Body, req.Body)
			for k, v := range tt.expected.Headers {
				assert.Equal(t, v, req.Headers[k])
			}
		})
	}
}

// TestExecuteCurl 測試執行cURL請求
func TestExecuteCurl(t *testing.T) {
	// 創建測試服務器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	service := services.NewCurlService()

	req := &interfaces.CurlRequest{
		URL:    server.URL,
		Method: "GET",
		Timeout: 10,
	}

	resp, err := service.ExecuteCurl(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, resp.Body, "success")
}

// TestExecuteCurlTimeout 測試cURL超時
func TestExecuteCurlTimeout(t *testing.T) {
	// 創建慢響應的測試服務器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := services.NewCurlService()

	req := &interfaces.CurlRequest{
		URL:     server.URL,
		Method:  "GET",
		Timeout: 1, // 1秒超時
	}

	_, err := service.ExecuteCurl(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestConvertToAPIRequest 測試轉換為API請求格式
func TestConvertToAPIRequest(t *testing.T) {
	service := services.NewCurlService()

	curlReq := &interfaces.CurlRequest{
		URL:    "https://api.example.com/test",
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Body:    `{"test": "data"}`,
		Timeout: 30,
	}

	apiReq := service.ConvertToAPIRequest(curlReq)

	assert.Equal(t, "https://api.example.com/test", apiReq.URL)
	assert.Equal(t, "POST", apiReq.Method)
	assert.Equal(t, `{"test": "data"}`, apiReq.Body)
	assert.Equal(t, "application/json", apiReq.Headers["Content-Type"])
	assert.Equal(t, "Bearer token123", apiReq.Headers["Authorization"])
	assert.Equal(t, 30*time.Second, apiReq.Timeout)
}

// TestTestConnection 測試連接測試
func TestTestConnection(t *testing.T) {
	// 創建測試服務器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := services.NewCurlService()

	// 測試成功連接
	err := service.TestConnection(server.URL)
	assert.NoError(t, err)

	// 測試失敗連接
	err = service.TestConnection("http://invalid-url-that-does-not-exist.com")
	assert.Error(t, err)
}