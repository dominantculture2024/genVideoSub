package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"genVideoSub/interfaces"
)

// CurlService Curl服務實現
type CurlService struct {
	client *http.Client
}

// NewCurlService 創建新的Curl服務
func NewCurlService() *CurlService {
	return &CurlService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExecuteCurl 執行curl請求
func (c *CurlService) ExecuteCurl(request *interfaces.CurlRequest) (*interfaces.CurlResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	
	// 設置超時
	if request.Timeout > 0 {
		c.client.Timeout = time.Duration(request.Timeout) * time.Second
	}
	
	// 創建HTTP請求
	var body io.Reader
	if request.Body != "" {
		body = strings.NewReader(request.Body)
	}
	
	httpReq, err := http.NewRequest(request.Method, request.URL, body)
	if err != nil {
		return &interfaces.CurlResponse{
			Error: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}
	
	// 設置請求頭
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}
	
	// 如果沒有設置Content-Type且有body，設置默認值
	if request.Body != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	
	// 執行請求
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return &interfaces.CurlResponse{
			Error: fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer resp.Body.Close()
	
	// 讀取響應體
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &interfaces.CurlResponse{
			StatusCode: resp.StatusCode,
			Error:      fmt.Sprintf("failed to read response: %v", err),
		}, nil
	}
	
	// 構建響應頭
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	return &interfaces.CurlResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(respBody),
	}, nil
}

// GenerateCurlCommand 生成curl命令
func (c *CurlService) GenerateCurlCommand(request *interfaces.CurlRequest) string {
	if request == nil {
		return ""
	}
	
	var cmd strings.Builder
	cmd.WriteString("curl")
	
	// 添加方法
	if request.Method != "" && request.Method != "GET" {
		cmd.WriteString(fmt.Sprintf(" -X %s", request.Method))
	}
	
	// 添加請求頭
	for key, value := range request.Headers {
		cmd.WriteString(fmt.Sprintf(" -H \"%s: %s\"", key, value))
	}
	
	// 添加請求體
	if request.Body != "" {
		// 轉義JSON中的雙引號
		body := strings.ReplaceAll(request.Body, "\"", "\\\"")
		cmd.WriteString(fmt.Sprintf(" -d \"%s\"", body))
	}
	
	// 添加超時
	if request.Timeout > 0 {
		cmd.WriteString(fmt.Sprintf(" --max-time %d", request.Timeout))
	}
	
	// 添加URL
	cmd.WriteString(fmt.Sprintf(" \"%s\"", request.URL))
	
	return cmd.String()
}

// ParseCurlCommand 解析curl命令
func (c *CurlService) ParseCurlCommand(command string) (*interfaces.CurlRequest, error) {
	if command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}
	
	request := &interfaces.CurlRequest{
		Method:  "GET",
		Headers: make(map[string]string),
		Timeout: 30,
	}
	
	// 移除curl命令開頭
	command = strings.TrimSpace(command)
	if strings.HasPrefix(command, "curl ") {
		command = strings.TrimPrefix(command, "curl ")
	}
	
	// 解析方法
	methodRegex := regexp.MustCompile(`-X\s+(\w+)`)
	if matches := methodRegex.FindStringSubmatch(command); len(matches) > 1 {
		request.Method = strings.ToUpper(matches[1])
	}
	
	// 解析請求頭
	headerRegex := regexp.MustCompile(`-H\s+["']([^:]+):\s*([^"']+)["']`)
	headerMatches := headerRegex.FindAllStringSubmatch(command, -1)
	for _, match := range headerMatches {
		if len(match) > 2 {
			request.Headers[strings.TrimSpace(match[1])] = strings.TrimSpace(match[2])
		}
	}
	
	// 解析請求體
	dataRegex := regexp.MustCompile(`-d\s+["']([^"']*)["']`)
	if matches := dataRegex.FindStringSubmatch(command); len(matches) > 1 {
		// 反轉義JSON
		body := strings.ReplaceAll(matches[1], "\\\"", "\"")
		request.Body = body
	}
	
	// 解析超時
	timeoutRegex := regexp.MustCompile(`--max-time\s+(\d+)`)
	if matches := timeoutRegex.FindStringSubmatch(command); len(matches) > 1 {
		if timeout, err := strconv.Atoi(matches[1]); err == nil {
			request.Timeout = timeout
		}
	}
	
	// 解析URL（通常在最後）
	urlRegex := regexp.MustCompile(`["']?(https?://[^\s"']+)["']?\s*$`)
	if matches := urlRegex.FindStringSubmatch(command); len(matches) > 1 {
		request.URL = matches[1]
	} else {
		return nil, fmt.Errorf("URL not found in curl command")
	}
	
	return request, nil
}

// ExecuteCurlFromCommand 從curl命令字符串執行請求
func (c *CurlService) ExecuteCurlFromCommand(command string) (*interfaces.CurlResponse, error) {
	request, err := c.ParseCurlCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse curl command: %w", err)
	}
	
	return c.ExecuteCurl(request)
}

// ConvertToAPIRequest 將curl請求轉換為API請求格式
func (c *CurlService) ConvertToAPIRequest(curlReq *interfaces.CurlRequest) (map[string]interface{}, error) {
	if curlReq == nil {
		return nil, fmt.Errorf("curl request cannot be nil")
	}
	
	apiReq := map[string]interface{}{
		"url":     curlReq.URL,
		"method":  curlReq.Method,
		"headers": curlReq.Headers,
		"timeout": curlReq.Timeout,
	}
	
	// 如果有請求體，嘗試解析為JSON
	if curlReq.Body != "" {
		var bodyData interface{}
		if err := json.Unmarshal([]byte(curlReq.Body), &bodyData); err == nil {
			apiReq["data"] = bodyData
		} else {
			apiReq["body"] = curlReq.Body
		}
	}
	
	return apiReq, nil
}

// TestConnection 測試連接
func (c *CurlService) TestConnection(url string) error {
	request := &interfaces.CurlRequest{
		URL:     url,
		Method:  "GET",
		Timeout: 10,
	}
	
	response, err := c.ExecuteCurl(request)
	if err != nil {
		return err
	}
	
	if response.Error != "" {
		return fmt.Errorf("connection test failed: %s", response.Error)
	}
	
	if response.StatusCode >= 400 {
		return fmt.Errorf("connection test failed with status: %d", response.StatusCode)
	}
	
	return nil
}