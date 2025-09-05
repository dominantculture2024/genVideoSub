import requests
import json
from typing import Dict, List, Optional, Any
import streamlit as st

class APIClient:
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
    
    def _make_request(self, method: str, endpoint: str, **kwargs) -> Dict[str, Any]:
        """Make HTTP request with error handling"""
        url = f"{self.base_url}{endpoint}"
        try:
            response = self.session.request(method, url, **kwargs)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            st.error(f"API請求失敗: {str(e)}")
            return {"error": str(e)}
        except json.JSONDecodeError:
            st.error("API回應格式錯誤")
            return {"error": "Invalid JSON response"}
    
    def create_task(self, prompt: str, image_url: Optional[str] = None, 
                   duration: float = 5.0, aspect_ratio: str = "16:9",
                   negative_prompt: Optional[str] = None, cfg_scale: float = 7.0) -> Dict[str, Any]:
        """Create a new video generation task"""
        data = {
            "prompt": prompt,
            "duration": duration,
            "aspect_ratio": aspect_ratio,
            "cfg_scale": cfg_scale
        }
        if image_url:
            data["image_url"] = image_url
        if negative_prompt:
            data["negative_prompt"] = negative_prompt
            
        return self._make_request("POST", "/api/tasks", json=data)
    
    def get_task(self, task_id: str) -> Dict[str, Any]:
        """Get task information"""
        return self._make_request("GET", f"/api/tasks/{task_id}")
    
    def get_task_status(self, task_id: str) -> Dict[str, Any]:
        """Get task status"""
        return self._make_request("GET", f"/api/tasks/{task_id}/status")
    
    def get_task_result(self, task_id: str) -> Dict[str, Any]:
        """Get task result"""
        return self._make_request("GET", f"/api/tasks/{task_id}/result")
    
    def list_tasks(self, limit: int = 50, offset: int = 0) -> Dict[str, Any]:
        """List all tasks"""
        params = {"limit": limit, "offset": offset}
        return self._make_request("GET", "/api/tasks", params=params)
    
    def delete_task(self, task_id: str) -> Dict[str, Any]:
        """Delete a task"""
        return self._make_request("DELETE", f"/api/tasks/{task_id}")
    
    def upload_file(self, file_data: bytes, filename: str) -> Dict[str, Any]:
        """Upload a file"""
        files = {"file": (filename, file_data)}
        try:
            url = f"{self.base_url}/api/files/upload"
            response = self.session.post(url, files=files)
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            st.error(f"文件上傳失敗: {str(e)}")
            return {"error": str(e)}
    
    def get_file_info(self, file_id: str) -> Dict[str, Any]:
        """Get file information"""
        return self._make_request("GET", f"/api/files/{file_id}/info")
    
    def delete_file(self, file_id: str) -> Dict[str, Any]:
        """Delete a file"""
        return self._make_request("DELETE", f"/api/files/{file_id}")

# Global API client instance
api_client = APIClient()