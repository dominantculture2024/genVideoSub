import unittest
import requests
import json
import time
import sys
import os
from unittest.mock import patch, Mock

# Add the frontend directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

from utils.api_client import APIClient

class TestAPIIntegration(unittest.TestCase):
    """Integration tests for API client with mock backend"""
    
    @classmethod
    def setUpClass(cls):
        """Set up test class with mock backend URL"""
        cls.backend_url = "http://localhost:8080"
        cls.client = APIClient(base_url=cls.backend_url)
        cls.mock_enabled = True  # Flag to indicate we're using mock service
    
    def setUp(self):
        """Set up each test"""
        # Mock streamlit to avoid import issues in tests
        self.st_patcher = patch('utils.api_client.st')
        self.mock_st = self.st_patcher.start()
    
    def tearDown(self):
        """Clean up after each test"""
        self.st_patcher.stop()
    
    def test_backend_connection(self):
        """Test basic connection to backend"""
        try:
            # Try to make a simple request
            response = requests.get(f"{self.backend_url}/health", timeout=5)
            self.assertTrue(response.status_code in [200, 404])  # 404 is OK if endpoint doesn't exist
        except requests.exceptions.ConnectionError:
            self.skipTest("Backend not available for integration testing")
    
    def test_create_and_retrieve_task(self):
        """Test creating a task and retrieving its information"""
        # Mock successful responses
        mock_create_response = {
            "id": "test-task-123",
            "status": "pending",
            "prompt": "test prompt",
            "duration": 5.0,
            "aspect_ratio": "16:9"
        }
        
        mock_get_response = {
            "id": "test-task-123",
            "status": "pending",
            "prompt": "test prompt",
            "duration": 5.0,
            "aspect_ratio": "16:9",
            "created_at": "2024-01-15T10:30:45Z"
        }
        
        with patch.object(self.client, '_make_request') as mock_request:
            mock_request.side_effect = [mock_create_response, mock_get_response]
            
            # Create task
            create_result = self.client.create_task(
                prompt="test prompt",
                duration=5.0,
                aspect_ratio="16:9"
            )
            
            self.assertNotIn("error", create_result)
            self.assertEqual(create_result["id"], "test-task-123")
            self.assertEqual(create_result["status"], "pending")
            
            # Retrieve task
            task_id = create_result["id"]
            get_result = self.client.get_task(task_id)
            
            self.assertNotIn("error", get_result)
            self.assertEqual(get_result["id"], task_id)
            self.assertEqual(get_result["prompt"], "test prompt")
    
    def test_task_status_progression(self):
        """Test task status progression from pending to completed"""
        task_id = "test-task-456"
        
        # Mock status progression
        status_responses = [
            {"status": "pending", "progress": 0},
            {"status": "processing", "progress": 50},
            {"status": "completed", "progress": 100}
        ]
        
        with patch.object(self.client, '_make_request') as mock_request:
            mock_request.side_effect = status_responses
            
            # Check initial status
            status1 = self.client.get_task_status(task_id)
            self.assertEqual(status1["status"], "pending")
            self.assertEqual(status1["progress"], 0)
            
            # Check processing status
            status2 = self.client.get_task_status(task_id)
            self.assertEqual(status2["status"], "processing")
            self.assertEqual(status2["progress"], 50)
            
            # Check completed status
            status3 = self.client.get_task_status(task_id)
            self.assertEqual(status3["status"], "completed")
            self.assertEqual(status3["progress"], 100)
    
    def test_task_result_retrieval(self):
        """Test retrieving task results"""
        task_id = "test-task-789"
        
        # Mock result responses
        mock_result_pending = {
            "status": "processing",
            "message": "Task is still processing"
        }
        
        mock_result_completed = {
            "status": "completed",
            "video_url": "http://example.com/videos/test-task-789.mp4",
            "thumbnail_url": "http://example.com/thumbnails/test-task-789.jpg",
            "duration": 5.0,
            "file_size": 1024000
        }
        
        with patch.object(self.client, '_make_request') as mock_request:
            mock_request.side_effect = [mock_result_pending, mock_result_completed]
            
            # Try to get result while processing
            result1 = self.client.get_task_result(task_id)
            self.assertEqual(result1["status"], "processing")
            self.assertNotIn("video_url", result1)
            
            # Get result when completed
            result2 = self.client.get_task_result(task_id)
            self.assertEqual(result2["status"], "completed")
            self.assertIn("video_url", result2)
            self.assertEqual(result2["video_url"], "http://example.com/videos/test-task-789.mp4")
    
    def test_list_tasks_pagination(self):
        """Test listing tasks with pagination"""
        # Mock task list responses
        mock_tasks_page1 = {
            "tasks": [
                {"id": "task-1", "status": "completed", "prompt": "prompt 1"},
                {"id": "task-2", "status": "processing", "prompt": "prompt 2"}
            ],
            "total": 5,
            "offset": 0,
            "limit": 2
        }
        
        mock_tasks_page2 = {
            "tasks": [
                {"id": "task-3", "status": "pending", "prompt": "prompt 3"},
                {"id": "task-4", "status": "failed", "prompt": "prompt 4"}
            ],
            "total": 5,
            "offset": 2,
            "limit": 2
        }
        
        with patch.object(self.client, '_make_request') as mock_request:
            mock_request.side_effect = [mock_tasks_page1, mock_tasks_page2]
            
            # Get first page
            page1 = self.client.list_tasks(limit=2, offset=0)
            self.assertNotIn("error", page1)
            self.assertEqual(len(page1["tasks"]), 2)
            self.assertEqual(page1["tasks"][0]["id"], "task-1")
            self.assertEqual(page1["total"], 5)
            
            # Get second page
            page2 = self.client.list_tasks(limit=2, offset=2)
            self.assertNotIn("error", page2)
            self.assertEqual(len(page2["tasks"]), 2)
            self.assertEqual(page2["tasks"][0]["id"], "task-3")
    
    def test_file_upload_workflow(self):
        """Test file upload and task creation workflow"""
        # Mock file upload response
        mock_upload_response = {
            "id": "file-123",
            "url": "http://example.com/uploads/image.jpg",
            "filename": "test_image.jpg",
            "size": 1024
        }
        
        # Mock task creation with image
        mock_task_response = {
            "id": "task-with-image-123",
            "status": "pending",
            "prompt": "test with image",
            "image_url": "http://example.com/uploads/image.jpg"
        }
        
        with patch.object(self.client.session, 'post') as mock_post:
            # Mock successful upload
            mock_response = Mock()
            mock_response.json.return_value = mock_upload_response
            mock_response.raise_for_status.return_value = None
            mock_post.return_value = mock_response
            
            with patch.object(self.client, '_make_request') as mock_request:
                mock_request.return_value = mock_task_response
                
                # Upload file
                file_data = b"fake image data"
                upload_result = self.client.upload_file(file_data, "test_image.jpg")
                
                self.assertNotIn("error", upload_result)
                self.assertEqual(upload_result["filename"], "test_image.jpg")
                
                # Create task with uploaded image
                task_result = self.client.create_task(
                    prompt="test with image",
                    image_url=upload_result["url"]
                )
                
                self.assertNotIn("error", task_result)
                self.assertEqual(task_result["image_url"], upload_result["url"])
    
    def test_error_handling(self):
        """Test various error scenarios"""
        # Test network error
        with patch.object(self.client.session, 'request', side_effect=requests.exceptions.ConnectionError("Connection failed")):
            result = self.client.get_task("test-id")
            self.assertIn("error", result)
            self.mock_st.error.assert_called()
        
        # Test HTTP error
        mock_response = Mock()
        mock_response.raise_for_status.side_effect = requests.exceptions.HTTPError("404 Not Found")
        
        with patch.object(self.client.session, 'request', return_value=mock_response):
            result = self.client.get_task("nonexistent-id")
            self.assertIn("error", result)
        
        # Test JSON decode error
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.side_effect = json.JSONDecodeError("Invalid JSON", "", 0)
        
        with patch.object(self.client.session, 'request', return_value=mock_response):
            result = self.client.get_task("test-id")
            self.assertEqual(result, {"error": "Invalid JSON response"})
    
    def test_concurrent_requests(self):
        """Test handling multiple concurrent requests"""
        import threading
        import queue
        
        results = queue.Queue()
        
        def make_request(task_id):
            """Make a request in a separate thread"""
            try:
                with patch.object(self.client, '_make_request') as mock_request:
                    mock_request.return_value = {"id": task_id, "status": "completed"}
                    result = self.client.get_task(task_id)
                    results.put((task_id, result))
            except Exception as e:
                results.put((task_id, {"error": str(e)}))
        
        # Create multiple threads
        threads = []
        task_ids = [f"task-{i}" for i in range(5)]
        
        for task_id in task_ids:
            thread = threading.Thread(target=make_request, args=(task_id,))
            threads.append(thread)
            thread.start()
        
        # Wait for all threads to complete
        for thread in threads:
            thread.join(timeout=5)
        
        # Collect results
        collected_results = {}
        while not results.empty():
            task_id, result = results.get()
            collected_results[task_id] = result
        
        # Verify all requests completed successfully
        self.assertEqual(len(collected_results), len(task_ids))
        for task_id in task_ids:
            self.assertIn(task_id, collected_results)
            self.assertNotIn("error", collected_results[task_id])
    
    def test_task_deletion(self):
        """Test task deletion"""
        task_id = "task-to-delete"
        
        # Mock successful deletion
        mock_delete_response = {"success": True, "message": "Task deleted successfully"}
        
        with patch.object(self.client, '_make_request') as mock_request:
            mock_request.return_value = mock_delete_response
            
            result = self.client.delete_task(task_id)
            
            self.assertNotIn("error", result)
            self.assertTrue(result["success"])
            
            # Verify the correct endpoint was called
            mock_request.assert_called_once_with("DELETE", f"/api/tasks/{task_id}")
    
    def test_file_management(self):
        """Test file information and deletion"""
        file_id = "file-123"
        
        # Mock file info response
        mock_file_info = {
            "id": file_id,
            "filename": "test.jpg",
            "size": 1024,
            "content_type": "image/jpeg",
            "created_at": "2024-01-15T10:30:45Z"
        }
        
        # Mock file deletion response
        mock_delete_response = {"success": True, "message": "File deleted successfully"}
        
        with patch.object(self.client, '_make_request') as mock_request:
            mock_request.side_effect = [mock_file_info, mock_delete_response]
            
            # Get file info
            info_result = self.client.get_file_info(file_id)
            self.assertNotIn("error", info_result)
            self.assertEqual(info_result["filename"], "test.jpg")
            
            # Delete file
            delete_result = self.client.delete_file(file_id)
            self.assertNotIn("error", delete_result)
            self.assertTrue(delete_result["success"])

class TestAPIIntegrationWithMockBackend(unittest.TestCase):
    """Integration tests specifically designed for mock backend"""
    
    def setUp(self):
        """Set up test with mock backend configuration"""
        self.client = APIClient(base_url="http://localhost:8080")
        
        # Mock streamlit
        self.st_patcher = patch('utils.api_client.st')
        self.mock_st = self.st_patcher.start()
    
    def tearDown(self):
        """Clean up after test"""
        self.st_patcher.stop()
    
    def test_mock_service_behavior(self):
        """Test specific mock service behaviors"""
        # Mock responses that simulate the mock service behavior
        mock_responses = {
            "create": {"id": "mock-task-123", "status": "pending"},
            "status_pending": {"status": "pending", "progress": 0},
            "status_processing": {"status": "processing", "progress": 50},
            "status_completed": {"status": "completed", "progress": 100},
            "result": {
                "status": "completed",
                "video_url": "http://localhost:8080/mock/videos/mock-task-123.mp4",
                "thumbnail_url": "http://localhost:8080/mock/thumbnails/mock-task-123.jpg"
            }
        }
        
        with patch.object(self.client, '_make_request') as mock_request:
            # Test task creation
            mock_request.return_value = mock_responses["create"]
            create_result = self.client.create_task("mock test prompt")
            self.assertEqual(create_result["id"], "mock-task-123")
            
            # Test status progression
            task_id = create_result["id"]
            
            # Pending status
            mock_request.return_value = mock_responses["status_pending"]
            status = self.client.get_task_status(task_id)
            self.assertEqual(status["status"], "pending")
            
            # Processing status
            mock_request.return_value = mock_responses["status_processing"]
            status = self.client.get_task_status(task_id)
            self.assertEqual(status["status"], "processing")
            
            # Completed status
            mock_request.return_value = mock_responses["status_completed"]
            status = self.client.get_task_status(task_id)
            self.assertEqual(status["status"], "completed")
            
            # Get result
            mock_request.return_value = mock_responses["result"]
            result = self.client.get_task_result(task_id)
            self.assertEqual(result["status"], "completed")
            self.assertIn("video_url", result)
            self.assertIn("mock", result["video_url"])  # Verify it's a mock URL
    
    def test_mock_error_simulation(self):
        """Test mock service error simulation"""
        # Mock error responses
        error_responses = [
            {"error": "Task not found", "code": 404},
            {"error": "Internal server error", "code": 500},
            {"error": "Rate limit exceeded", "code": 429}
        ]
        
        with patch.object(self.client, '_make_request') as mock_request:
            for error_response in error_responses:
                mock_request.return_value = error_response
                result = self.client.get_task("error-test-id")
                self.assertIn("error", result)
                self.assertIn("code", result)

if __name__ == '__main__':
    unittest.main()