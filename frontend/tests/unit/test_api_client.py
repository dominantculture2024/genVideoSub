import unittest
import json
from unittest.mock import Mock, patch, MagicMock
import requests
import sys
import os

# Add the frontend directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

from utils.api_client import APIClient

class TestAPIClient(unittest.TestCase):
    """Test cases for APIClient"""
    
    def setUp(self):
        """Set up test fixtures"""
        self.client = APIClient(base_url="http://localhost:8080")
        self.mock_response = Mock()
        self.mock_response.json.return_value = {"success": True}
        self.mock_response.raise_for_status.return_value = None
    
    @patch('utils.api_client.requests.Session')
    def test_init(self, mock_session):
        """Test APIClient initialization"""
        client = APIClient("http://test.com/")
        self.assertEqual(client.base_url, "http://test.com")
        mock_session.assert_called_once()
    
    @patch('utils.api_client.st')
    def test_make_request_success(self, mock_st):
        """Test successful API request"""
        with patch.object(self.client.session, 'request', return_value=self.mock_response):
            result = self.client._make_request("GET", "/test")
            self.assertEqual(result, {"success": True})
    
    @patch('utils.api_client.st')
    def test_make_request_http_error(self, mock_st):
        """Test HTTP error handling"""
        mock_response = Mock()
        mock_response.raise_for_status.side_effect = requests.exceptions.HTTPError("404 Not Found")
        
        with patch.object(self.client.session, 'request', return_value=mock_response):
            result = self.client._make_request("GET", "/test")
            self.assertIn("error", result)
            mock_st.error.assert_called_once()
    
    @patch('utils.api_client.st')
    def test_make_request_json_error(self, mock_st):
        """Test JSON decode error handling"""
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.side_effect = json.JSONDecodeError("Invalid JSON", "", 0)
        
        with patch.object(self.client.session, 'request', return_value=mock_response):
            result = self.client._make_request("GET", "/test")
            self.assertEqual(result, {"error": "Invalid JSON response"})
            mock_st.error.assert_called_with("API回應格式錯誤")
    
    def test_create_task_minimal(self):
        """Test creating task with minimal parameters"""
        expected_data = {
            "prompt": "test prompt",
            "duration": 5.0,
            "aspect_ratio": "16:9",
            "cfg_scale": 7.0
        }
        
        with patch.object(self.client, '_make_request', return_value={"id": "test-id"}) as mock_request:
            result = self.client.create_task("test prompt")
            mock_request.assert_called_once_with("POST", "/api/tasks", json=expected_data)
            self.assertEqual(result, {"id": "test-id"})
    
    def test_create_task_full_parameters(self):
        """Test creating task with all parameters"""
        expected_data = {
            "prompt": "test prompt",
            "duration": 10.0,
            "aspect_ratio": "9:16",
            "cfg_scale": 8.0,
            "image_url": "http://example.com/image.jpg",
            "negative_prompt": "bad quality"
        }
        
        with patch.object(self.client, '_make_request', return_value={"id": "test-id"}) as mock_request:
            result = self.client.create_task(
                prompt="test prompt",
                image_url="http://example.com/image.jpg",
                duration=10.0,
                aspect_ratio="9:16",
                negative_prompt="bad quality",
                cfg_scale=8.0
            )
            mock_request.assert_called_once_with("POST", "/api/tasks", json=expected_data)
            self.assertEqual(result, {"id": "test-id"})
    
    def test_get_task(self):
        """Test getting task information"""
        with patch.object(self.client, '_make_request', return_value={"id": "test-id"}) as mock_request:
            result = self.client.get_task("test-id")
            mock_request.assert_called_once_with("GET", "/api/tasks/test-id")
            self.assertEqual(result, {"id": "test-id"})
    
    def test_get_task_status(self):
        """Test getting task status"""
        with patch.object(self.client, '_make_request', return_value={"status": "processing"}) as mock_request:
            result = self.client.get_task_status("test-id")
            mock_request.assert_called_once_with("GET", "/api/tasks/test-id/status")
            self.assertEqual(result, {"status": "processing"})
    
    def test_get_task_result(self):
        """Test getting task result"""
        with patch.object(self.client, '_make_request', return_value={"video_url": "http://example.com/video.mp4"}) as mock_request:
            result = self.client.get_task_result("test-id")
            mock_request.assert_called_once_with("GET", "/api/tasks/test-id/result")
            self.assertEqual(result, {"video_url": "http://example.com/video.mp4"})
    
    def test_list_tasks_default(self):
        """Test listing tasks with default parameters"""
        with patch.object(self.client, '_make_request', return_value={"tasks": []}) as mock_request:
            result = self.client.list_tasks()
            mock_request.assert_called_once_with("GET", "/api/tasks", params={"limit": 50, "offset": 0})
            self.assertEqual(result, {"tasks": []})
    
    def test_list_tasks_custom_params(self):
        """Test listing tasks with custom parameters"""
        with patch.object(self.client, '_make_request', return_value={"tasks": []}) as mock_request:
            result = self.client.list_tasks(limit=10, offset=20)
            mock_request.assert_called_once_with("GET", "/api/tasks", params={"limit": 10, "offset": 20})
            self.assertEqual(result, {"tasks": []})
    
    def test_delete_task(self):
        """Test deleting a task"""
        with patch.object(self.client, '_make_request', return_value={"success": True}) as mock_request:
            result = self.client.delete_task("test-id")
            mock_request.assert_called_once_with("DELETE", "/api/tasks/test-id")
            self.assertEqual(result, {"success": True})
    
    @patch('utils.api_client.st')
    def test_upload_file_success(self, mock_st):
        """Test successful file upload"""
        mock_response = Mock()
        mock_response.json.return_value = {"url": "http://example.com/file.jpg"}
        mock_response.raise_for_status.return_value = None
        
        with patch.object(self.client.session, 'post', return_value=mock_response):
            result = self.client.upload_file(b"file data", "test.jpg")
            self.assertEqual(result, {"url": "http://example.com/file.jpg"})
    
    @patch('utils.api_client.st')
    def test_upload_file_error(self, mock_st):
        """Test file upload error handling"""
        with patch.object(self.client.session, 'post', side_effect=requests.exceptions.RequestException("Upload failed")):
            result = self.client.upload_file(b"file data", "test.jpg")
            self.assertIn("error", result)
            mock_st.error.assert_called_once()
    
    def test_get_file_info(self):
        """Test getting file information"""
        with patch.object(self.client, '_make_request', return_value={"filename": "test.jpg"}) as mock_request:
            result = self.client.get_file_info("file-id")
            mock_request.assert_called_once_with("GET", "/api/files/file-id/info")
            self.assertEqual(result, {"filename": "test.jpg"})
    
    def test_delete_file(self):
        """Test deleting a file"""
        with patch.object(self.client, '_make_request', return_value={"success": True}) as mock_request:
            result = self.client.delete_file("file-id")
            mock_request.assert_called_once_with("DELETE", "/api/files/file-id")
            self.assertEqual(result, {"success": True})

class TestAPIClientIntegration(unittest.TestCase):
    """Integration tests for APIClient with mock backend"""
    
    def setUp(self):
        """Set up test fixtures"""
        self.client = APIClient(base_url="http://localhost:8080")
    
    @patch('utils.api_client.st')
    def test_full_workflow_mock(self, mock_st):
        """Test complete workflow with mocked responses"""
        # Mock responses for different endpoints
        mock_responses = {
            "/api/tasks": {"id": "task-123", "status": "pending"},
            "/api/tasks/task-123/status": {"status": "processing"},
            "/api/tasks/task-123/result": {"status": "completed", "video_url": "http://example.com/video.mp4"}
        }
        
        def mock_request(method, endpoint, **kwargs):
            return mock_responses.get(endpoint, {"error": "Not found"})
        
        with patch.object(self.client, '_make_request', side_effect=mock_request):
            # Create task
            create_result = self.client.create_task("test prompt")
            self.assertEqual(create_result["id"], "task-123")
            
            # Check status
            status_result = self.client.get_task_status("task-123")
            self.assertEqual(status_result["status"], "processing")
            
            # Get result
            result = self.client.get_task_result("task-123")
            self.assertEqual(result["status"], "completed")
            self.assertIn("video_url", result)

if __name__ == '__main__':
    unittest.main()