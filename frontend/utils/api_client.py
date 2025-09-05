import requests
import json
from typing import Dict, List, Optional

class APIClient:
    """API客戶端類，用於與後端API通信"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })
    
    def set_base_url(self, url: str):
        """設置API基礎URL"""
        self.base_url = url.rstrip('/')
    
    def check_health(self) -> bool:
        """檢查API健康狀態"""
        try:
            response = self.session.get(f"{self.base_url}/api/health", timeout=5)
            return response.status_code == 200
        except Exception as e:
            print(f"Health check failed: {e}")
            return False
    
    def create_task(self, task_data: Dict) -> Optional[Dict]:
        """創建新任務"""
        try:
            response = self.session.post(
                f"{self.base_url}/api/tasks",
                json=task_data,
                timeout=30
            )
            if response.status_code == 200:
                return response.json()
            else:
                print(f"Create task failed: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Create task error: {e}")
            return None
    
    def get_task_status(self, task_id: int) -> Optional[Dict]:
        """獲取任務狀態"""
        try:
            response = self.session.get(
                f"{self.base_url}/api/tasks/{task_id}/status",
                timeout=10
            )
            if response.status_code == 200:
                return response.json()
            else:
                print(f"Get task status failed: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Get task status error: {e}")
            return None
    
    def get_task_list(self) -> Optional[List[Dict]]:
        """獲取任務列表"""
        try:
            response = self.session.get(
                f"{self.base_url}/api/tasks",
                timeout=10
            )
            if response.status_code == 200:
                return response.json()
            else:
                print(f"Get task list failed: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Get task list error: {e}")
            return None
    
    def upload_completed_video(self, task_id: int, video_file, external_task_id: str = None) -> Optional[Dict]:
        """上傳完成的視頻文件"""
        try:
            files = {'video_file': video_file}
            data = {}
            if external_task_id:
                data['external_task_id'] = external_task_id
            
            # 移除JSON headers for file upload
            headers = {'Accept': 'application/json'}
            
            response = requests.post(
                f"{self.base_url}/api/tasks/{task_id}/completed",
                files=files,
                data=data,
                headers=headers,
                timeout=300  # 5 minutes for large file upload
            )
            
            if response.status_code == 200:
                return response.json()
            else:
                print(f"Upload video failed: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Upload video error: {e}")
            return None
    
    def delete_task(self, task_id: int) -> bool:
        """刪除任務"""
        try:
            response = self.session.delete(
                f"{self.base_url}/api/tasks/{task_id}",
                timeout=10
            )
            return response.status_code == 200
        except Exception as e:
            print(f"Delete task error: {e}")
            return False
    
    def get_metrics(self) -> Optional[Dict]:
        """獲取系統指標"""
        try:
            response = self.session.get(
                f"{self.base_url}/api/metrics",
                timeout=10
            )
            if response.status_code == 200:
                return response.json()
            else:
                print(f"Get metrics failed: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            print(f"Get metrics error: {e}")
            return None