package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"genVideoSub/config"
	"genVideoSub/interfaces"
	"genVideoSub/services"
	"github.com/gin-gonic/gin"
)

// APIHandler API相關處理器
type APIHandler struct {
	config          *config.Config
	providerManager *services.APIProviderManager
	curlService     interfaces.CurlInterface
}

// NewAPIHandler 創建新的API處理器
func NewAPIHandler(cfg *config.Config, providerManager *services.APIProviderManager, curlService interfaces.CurlInterface) *APIHandler {
	return &APIHandler{
		config:          cfg,
		providerManager: providerManager,
		curlService:     curlService,
	}
}

// SwitchProvider 切換API提供商
func (h *APIHandler) SwitchProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Provider string `json:"provider"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 驗證提供商
	provider := interfaces.APIProvider(request.Provider)
	if err := h.providerManager.SwitchProvider(provider); err != nil {
		http.Error(w, fmt.Sprintf("Failed to switch provider: %v", err), http.StatusBadRequest)
		return
	}

	// 更新配置
	if err := h.config.SetAPIProvider(request.Provider); err != nil {
		log.Printf("Warning: Failed to update config: %v", err)
	}

	response := map[string]interface{}{
		"success":         true,
		"current_provider": request.Provider,
		"message":         fmt.Sprintf("Switched to provider: %s", request.Provider),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Gin 兼容的處理器方法

// GetProvidersGin Gin版本的獲取提供商
func (h *APIHandler) GetProvidersGin(c *gin.Context) {
	h.GetProviders(c.Writer, c.Request)
}

// SwitchProviderGin Gin版本的切換提供商
func (h *APIHandler) SwitchProviderGin(c *gin.Context) {
	h.SwitchProvider(c.Writer, c.Request)
}

// GetAPIConfigGin Gin版本的獲取API配置
func (h *APIHandler) GetAPIConfigGin(c *gin.Context) {
	h.GetAPIConfig(c.Writer, c.Request)
}

// UpdateAPIConfigGin Gin版本的更新API配置
func (h *APIHandler) UpdateAPIConfigGin(c *gin.Context) {
	h.UpdateAPIConfig(c.Writer, c.Request)
}

// TestConnectionGin Gin版本的測試連接
func (h *APIHandler) TestConnectionGin(c *gin.Context) {
	h.TestConnection(c.Writer, c.Request)
}

// ExecuteCurlGin Gin版本的執行curl
func (h *APIHandler) ExecuteCurlGin(c *gin.Context) {
	h.ExecuteCurl(c.Writer, c.Request)
}

// ExecuteCurlCommandGin Gin版本的執行curl命令
func (h *APIHandler) ExecuteCurlCommandGin(c *gin.Context) {
	h.ExecuteCurlCommand(c.Writer, c.Request)
}

// GenerateCurlCommandGin Gin版本的生成curl命令
func (h *APIHandler) GenerateCurlCommandGin(c *gin.Context) {
	h.GenerateCurlCommand(c.Writer, c.Request)
}

// ParseCurlCommandGin Gin版本的解析curl命令
func (h *APIHandler) ParseCurlCommandGin(c *gin.Context) {
	h.ParseCurlCommand(c.Writer, c.Request)
}

// GetProviders 獲取所有可用的API提供商
func (h *APIHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providers := h.providerManager.ListProviders()
	current := h.providerManager.GetCurrentProvider()
	status := h.providerManager.GetProviderStatus()

	response := map[string]interface{}{
		"providers":        providers,
		"current_provider": current,
		"provider_status":  status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExecuteCurl 執行curl請求
func (h *APIHandler) ExecuteCurl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.config.CurlEnabled {
		http.Error(w, "Curl functionality is disabled", http.StatusForbidden)
		return
	}

	var request interfaces.CurlRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 設置默認超時
	if request.Timeout == 0 {
		request.Timeout = h.config.CurlTimeout
	}

	response, err := h.curlService.ExecuteCurl(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Curl execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExecuteCurlCommand 從curl命令執行請求
func (h *APIHandler) ExecuteCurlCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.config.CurlEnabled {
		http.Error(w, "Curl functionality is disabled", http.StatusForbidden)
		return
	}

	var request struct {
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	curlService, ok := h.curlService.(*services.CurlService)
	if !ok {
		http.Error(w, "Curl service not available", http.StatusInternalServerError)
		return
	}

	response, err := curlService.ExecuteCurlFromCommand(request.Command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Curl command execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateCurlCommand 生成curl命令
func (h *APIHandler) GenerateCurlCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request interfaces.CurlRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	command := h.curlService.GenerateCurlCommand(&request)

	response := map[string]interface{}{
		"command": command,
		"request": request,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ParseCurlCommand 解析curl命令
func (h *APIHandler) ParseCurlCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	parsedRequest, err := h.curlService.ParseCurlCommand(request.Command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse curl command: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"parsed_request": parsedRequest,
		"original_command": request.Command,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAPIConfig 獲取API配置信息
func (h *APIHandler) GetAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 隱藏敏感信息
	providerConfigs := make(map[string]interface{})
	for name, config := range h.config.APIProviderConfigs {
		providerConfigs[name] = map[string]interface{}{
			"base_url": config.BaseURL,
			"timeout":  config.Timeout,
			"enabled":  config.Enabled,
			"has_api_key": config.APIKey != "",
		}
	}

	response := map[string]interface{}{
		"mode":                h.config.Mode,
		"environment":         h.config.Environment,
		"current_provider":    h.config.APIProvider,
		"provider_configs":    providerConfigs,
		"curl_enabled":        h.config.CurlEnabled,
		"curl_timeout":        h.config.CurlTimeout,
		"curl_retries":        h.config.CurlRetries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateAPIConfig 更新API配置
func (h *APIHandler) UpdateAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Mode        string `json:"mode,omitempty"`
		Environment string `json:"environment,omitempty"`
		CurlEnabled *bool  `json:"curl_enabled,omitempty"`
		CurlTimeout *int   `json:"curl_timeout,omitempty"`
		CurlRetries *int   `json:"curl_retries,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 更新配置
	if request.Mode != "" {
		h.config.Mode = request.Mode
	}
	if request.Environment != "" {
		h.config.Environment = request.Environment
	}
	if request.CurlEnabled != nil {
		h.config.CurlEnabled = *request.CurlEnabled
	}
	if request.CurlTimeout != nil {
		h.config.CurlTimeout = *request.CurlTimeout
	}
	if request.CurlRetries != nil {
		h.config.CurlRetries = *request.CurlRetries
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
		"config": map[string]interface{}{
			"mode":         h.config.Mode,
			"environment": h.config.Environment,
			"curl_enabled": h.config.CurlEnabled,
			"curl_timeout": h.config.CurlTimeout,
			"curl_retries": h.config.CurlRetries,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestConnection 測試API連接
func (h *APIHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Provider string `json:"provider,omitempty"`
		URL      string `json:"url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var testURL string
	if request.URL != "" {
		testURL = request.URL
	} else if request.Provider != "" {
		if config, exists := h.config.APIProviderConfigs[request.Provider]; exists {
			testURL = config.BaseURL
		} else {
			http.Error(w, "Provider not found", http.StatusBadRequest)
			return
		}
	} else {
		// 使用當前提供商
		if config := h.config.GetCurrentProviderConfig(); config != nil {
			testURL = config.BaseURL
		} else {
			http.Error(w, "No provider configured", http.StatusBadRequest)
			return
		}
	}

	curlService, ok := h.curlService.(*services.CurlService)
	if !ok {
		http.Error(w, "Curl service not available", http.StatusInternalServerError)
		return
	}

	err := curlService.TestConnection(testURL)
	response := map[string]interface{}{
		"url":     testURL,
		"success": err == nil,
	}

	if err != nil {
		response["error"] = err.Error()
	} else {
		response["message"] = "Connection successful"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}