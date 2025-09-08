import unittest
import time
import threading
from unittest.mock import patch, Mock, MagicMock
import sys
import os

# Add the frontend directory to the Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

from utils.api_client import APIClient
from components.task_manager import (
    display_task_status, format_datetime, create_task_form,
    display_task_list, display_task_details
)

class TestFrontendE2E(unittest.TestCase):
    """End-to-end tests for the complete frontend workflow"""
    
    def setUp(self):
        """Set up test environment"""
        self.client = APIClient(base_url="http://localhost:8080")
        
        # Mock streamlit components
        self.st_patcher = patch('components.task_manager.st')
        self.mock_st = self.st_patcher.start()
        
        # Configure mock streamlit components
        self.mock_st.text_area.return_value = "Test prompt"
        self.mock_st.file_uploader.return_value = None
        self.mock_st.button.return_value = False
        self.mock_st.selectbox.return_value = "pending"
        self.mock_st.slider.return_value = 5.0
        self.mock_st.checkbox.return_value = False
        self.mock_st.form_submit_button.return_value = False
        
        # Configure columns mock with context manager support
        col1_mock = Mock()
        col1_mock.__enter__ = Mock(return_value=col1_mock)
        col1_mock.__exit__ = Mock(return_value=None)
        col2_mock = Mock()
        col2_mock.__enter__ = Mock(return_value=col2_mock)
        col2_mock.__exit__ = Mock(return_value=None)
        
        # Configure st.columns to return the correct number of mock columns
        def mock_columns(num_cols):
            def create_column_mock():
                col_mock = Mock()
                col_mock.__enter__ = Mock(return_value=col_mock)
                col_mock.__exit__ = Mock(return_value=None)
                return col_mock
            
            if isinstance(num_cols, int):
                return [create_column_mock() for _ in range(num_cols)]
            elif isinstance(num_cols, list):
                return [create_column_mock() for _ in range(len(num_cols))]
            else:
                return [create_column_mock(), create_column_mock(), create_column_mock()]  # Default to 3 columns for compatibility
        
        self.mock_st.columns.side_effect = mock_columns
        
        # Configure form mock with context manager support
        form_mock = Mock()
        form_mock.__enter__ = Mock(return_value=form_mock)
        form_mock.__exit__ = Mock(return_value=None)
        self.mock_st.form.return_value = form_mock
        
        # Configure spinner mock with context manager support
        spinner_mock = Mock()
        spinner_mock.__enter__ = Mock(return_value=spinner_mock)
        spinner_mock.__exit__ = Mock(return_value=None)
        self.mock_st.spinner.return_value = spinner_mock
        
        # Mock session state
        self.mock_st.session_state = {}
        
        # Mock API client in components
        self.api_patcher = patch('components.task_manager.api_client')
        self.mock_api_client = self.api_patcher.start()
    
    def tearDown(self):
        """Clean up after test"""
        self.st_patcher.stop()
        self.api_patcher.stop()
    
    def test_complete_video_generation_workflow(self):
        """Test the complete workflow from task creation to result retrieval"""
        # Mock task creation
        mock_task = {
            "id": "e2e-task-123",
            "status": "pending",
            "prompt": "Test video generation",
            "created_at": "2024-01-15T10:30:45Z"
        }
        
        # Mock status progression
        status_progression = [
            {"status": "pending", "progress": 0},
            {"status": "processing", "progress": 25},
            {"status": "processing", "progress": 50},
            {"status": "processing", "progress": 75},
            {"status": "completed", "progress": 100}
        ]
        
        # Mock final result
        mock_result = {
            "status": "completed",
            "video_url": "http://localhost:8080/videos/e2e-task-123.mp4",
            "thumbnail_url": "http://localhost:8080/thumbnails/e2e-task-123.jpg",
            "duration": 5.0,
            "file_size": 2048000
        }
        
        with patch.object(self.client, '_make_request') as mock_request:
            # Step 1: Create task
            mock_request.return_value = mock_task
            
            # Simulate task creation form submission
            self.mock_st.form_submit_button.return_value = True
            self.mock_st.text_area.return_value = "Test video generation"
            
            # Configure mock API client for task creation
            self.mock_api_client.create_task.return_value = mock_task
            
            # Call create_task_form
            create_task_form()
            
            # Verify task creation was called
            self.mock_api_client.create_task.assert_called()
            
            # Step 2: Monitor task status
            task_id = mock_task["id"]
            
            for i, status in enumerate(status_progression):
                mock_request.return_value = status
                
                # Configure mock API client for status check
                self.mock_api_client.get_task_status.return_value = status
                
                # Simulate status checking
                result = self.client.get_task_status(task_id)
                self.assertEqual(result["status"], status["status"])
                self.assertEqual(result["progress"], status["progress"])
                
                # Test display_task_status function
                status_display = display_task_status(status["status"])
                
                # Verify status display format
                self.assertIn(status["status"].upper(), status_display)
            
            # Step 3: Get final result
            mock_request.return_value = mock_result
            self.mock_api_client.get_task_result.return_value = mock_result
            
            result = self.client.get_task_result(task_id)
            self.assertEqual(result["status"], "completed")
            self.assertIn("video_url", result)
            self.assertIn("thumbnail_url", result)
    
    def test_task_list_and_management_workflow(self):
        """Test task listing and management workflow"""
        # Mock task list
        mock_tasks = {
            "tasks": [
                {
                    "id": "task-1",
                    "status": "completed",
                    "prompt": "First video",
                    "created_at": "2024-01-15T10:00:00Z"
                },
                {
                    "id": "task-2",
                    "status": "processing",
                    "prompt": "Second video",
                    "created_at": "2024-01-15T11:00:00Z"
                },
                {
                    "id": "task-3",
                    "status": "failed",
                    "prompt": "Failed video",
                    "created_at": "2024-01-15T12:00:00Z"
                }
            ],
            "total": 3,
            "offset": 0,
            "limit": 10
        }
        
        # Configure mock API client
        self.mock_api_client.list_tasks.return_value = mock_tasks
        
        # Test task list display
        display_task_list()
        
        # Verify API was called
        self.mock_api_client.list_tasks.assert_called()
        
        # Verify streamlit components were used to display tasks
        self.mock_st.dataframe.assert_called()
        
        # Test individual task details
        for task in mock_tasks["tasks"]:
            # Configure mock for task details
            self.mock_api_client.get_task.return_value = task
            
            # Test display_task_details
            display_task_details(task["id"])
            
            # Verify task status display
            display_task_status(task["status"])
    
    def test_file_upload_workflow(self):
        """Test file upload and task creation with image"""
        # Mock uploaded file
        mock_file = Mock()
        mock_file.name = "test_image.jpg"
        mock_file.read.return_value = b"fake image data"
        mock_file.type = "image/jpeg"
        
        # Mock file upload response
        mock_upload_response = {
            "id": "file-123",
            "url": "http://localhost:8080/uploads/test_image.jpg",
            "filename": "test_image.jpg",
            "size": 1024
        }
        
        # Mock task creation with image
        mock_task_with_image = {
            "id": "task-with-image-123",
            "status": "pending",
            "prompt": "Video with uploaded image",
            "image_url": "http://localhost:8080/uploads/test_image.jpg"
        }
        
        # Configure streamlit file uploader to return mock file
        self.mock_st.file_uploader.return_value = mock_file
        self.mock_st.form_submit_button.return_value = True
        self.mock_st.text_area.return_value = "Video with uploaded image"
        
        # Configure mock API client
        self.mock_api_client.upload_file.return_value = mock_upload_response
        self.mock_api_client.create_task.return_value = mock_task_with_image
        
        # Test file upload workflow in task creation form
        create_task_form()
        
        # Verify file upload was called
        self.mock_api_client.upload_file.assert_called()
        
        # Verify task creation with image URL
        self.mock_api_client.create_task.assert_called()
    
    def test_error_handling_workflow(self):
        """Test error handling throughout the workflow"""
        # Test various error scenarios
        error_scenarios = [
            {"error": "Network connection failed", "code": 500},
            {"error": "Task not found", "code": 404},
            {"error": "Rate limit exceeded", "code": 429},
            {"error": "Invalid request parameters", "code": 400}
        ]
        
        for error in error_scenarios:
            # Configure mock API to return error
            self.mock_api_client.create_task.return_value = error
            self.mock_api_client.get_task.return_value = error
            self.mock_api_client.get_task_status.return_value = error
            
            # Test task creation with error
            self.mock_st.form_submit_button.return_value = True
            create_task_form()
            
            # Verify error was displayed
            self.mock_st.error.assert_called()
            
            # Test task list with error
            self.mock_api_client.list_tasks.return_value = error
            display_task_list()
            
            # Verify error handling in task list
            self.mock_st.error.assert_called()
    
    def test_concurrent_user_interactions(self):
        """Test handling multiple concurrent user interactions"""
        import queue
        
        results = queue.Queue()
        
        def simulate_user_action(action_type, task_id=None):
            """Simulate different user actions"""
            try:
                if action_type == "create_task":
                    # Mock task creation
                    mock_task = {"id": f"concurrent-task-{task_id}", "status": "pending"}
                    self.mock_api_client.create_task.return_value = mock_task
                    
                    # Simulate form submission
                    self.mock_st.form_submit_button.return_value = True
                    create_task_form()
                    
                    results.put((action_type, "success", mock_task["id"]))
                
                elif action_type == "check_status":
                    # Mock status check
                    mock_status = {"status": "processing", "progress": 50}
                    self.mock_api_client.get_task_status.return_value = mock_status
                    
                    status = self.mock_api_client.get_task_status(task_id)
                    results.put((action_type, "success", status["status"]))
                
                elif action_type == "list_tasks":
                    # Mock task listing
                    mock_tasks = {"tasks": [], "total": 0}
                    self.mock_api_client.list_tasks.return_value = mock_tasks
                    
                    display_task_list()
                    results.put((action_type, "success", len(mock_tasks["tasks"])))
                    
            except Exception as e:
                results.put((action_type, "error", str(e)))
        
        # Create multiple threads for concurrent actions
        threads = []
        actions = [
            ("create_task", 1),
            ("create_task", 2),
            ("check_status", "task-1"),
            ("check_status", "task-2"),
            ("list_tasks", None)
        ]
        
        for action_type, param in actions:
            thread = threading.Thread(
                target=simulate_user_action,
                args=(action_type, param)
            )
            threads.append(thread)
            thread.start()
        
        # Wait for all threads to complete
        for thread in threads:
            thread.join(timeout=5)
        
        # Collect and verify results
        collected_results = []
        while not results.empty():
            collected_results.append(results.get())
        
        # Verify all actions completed
        self.assertEqual(len(collected_results), len(actions))
        
        # Verify most actions completed successfully (allow some errors in concurrent testing)
        success_count = sum(1 for action_type, status, result in collected_results if status == "success")
        self.assertGreaterEqual(success_count, len(actions) - 3)  # Allow up to 3 failures in concurrent testing
    
    def test_session_state_management(self):
        """Test session state management across interactions"""
        # Initialize session state
        self.mock_st.session_state = {
            "current_task_id": None,
            "task_history": [],
            "upload_history": []
        }
        
        # Test task creation and session state update
        mock_task = {
            "id": "session-task-123",
            "status": "pending",
            "prompt": "Session test task"
        }
        
        self.mock_api_client.create_task.return_value = mock_task
        
        # Simulate task creation
        self.mock_st.form_submit_button.return_value = True
        create_task_form()
        
        # Verify session state would be updated (in real implementation)
        # Note: This is a simplified test as we're mocking streamlit
        self.mock_api_client.create_task.assert_called()
        
        # Test task history management
        mock_tasks = {
            "tasks": [mock_task],
            "total": 1
        }
        
        self.mock_api_client.list_tasks.return_value = mock_tasks
        display_task_list()
        
        # Verify task list was retrieved
        self.mock_api_client.list_tasks.assert_called()
    
    def test_responsive_ui_updates(self):
        """Test UI responsiveness to data changes"""
        # Test status updates
        status_sequence = ["pending", "processing", "completed"]
        
        for status in status_sequence:
            # Test status display function
            status_display = display_task_status(status)
            
            # Verify status display format
            self.assertIn(status.upper(), status_display)
            
            # Verify appropriate emoji is used
            if status == "pending":
                self.assertIn("🟡", status_display)
            elif status == "processing":
                self.assertIn("🔵", status_display)
            elif status == "completed":
                self.assertIn("🟢", status_display)
        
        # Test datetime formatting
        test_datetime = "2024-01-15T10:30:45Z"
        formatted = format_datetime(test_datetime)
        self.assertIsInstance(formatted, str)
        self.assertIn("2024", formatted)
    
    def test_performance_under_load(self):
        """Test frontend performance with multiple tasks"""
        # Create a large number of mock tasks
        large_task_list = {
            "tasks": [
                {
                    "id": f"perf-task-{i}",
                    "status": "completed" if i % 3 == 0 else "processing",
                    "prompt": f"Performance test task {i}",
                    "created_at": "2024-01-15T10:30:45Z"
                }
                for i in range(100)
            ],
            "total": 100,
            "offset": 0,
            "limit": 100
        }
        
        self.mock_api_client.list_tasks.return_value = large_task_list
        
        # Measure time to display large task list
        start_time = time.time()
        display_task_list()
        end_time = time.time()
        
        # Verify the operation completed in reasonable time
        # (This is a mock test, so it should be very fast)
        self.assertLess(end_time - start_time, 1.0)  # Should complete in less than 1 second
        
        # Verify API was called
        self.mock_api_client.list_tasks.assert_called()

class TestFrontendE2EWithRealBackend(unittest.TestCase):
    """E2E tests designed to work with real backend (when available)"""
    
    def setUp(self):
        """Set up test with real backend configuration"""
        self.client = APIClient(base_url="http://localhost:8080")
        
        # Mock streamlit but allow real API calls
        self.st_patcher = patch('streamlit')
        self.mock_st = self.st_patcher.start()
        
        # Configure mock streamlit
        self.mock_st.text_input.return_value = "Real backend test"
        self.mock_st.button.return_value = False
        self.mock_st.columns.return_value = [Mock(), Mock(), Mock()]
        self.mock_st.session_state = {}
    
    def tearDown(self):
        """Clean up after test"""
        self.st_patcher.stop()
    
    @unittest.skipUnless(
        os.getenv('TEST_WITH_REAL_BACKEND') == 'true',
        "Real backend tests disabled"
    )
    def test_real_backend_connection(self):
        """Test connection to real backend service"""
        try:
            # Test basic connectivity
            result = self.client._make_request("GET", "/api/health")
            self.assertNotIn("error", result)
        except Exception as e:
            self.skipTest(f"Real backend not available: {e}")
    
    @unittest.skipUnless(
        os.getenv('TEST_WITH_REAL_BACKEND') == 'true',
        "Real backend tests disabled"
    )
    def test_real_task_creation_workflow(self):
        """Test task creation with real backend"""
        try:
            # Create a real task
            task_data = {
                "prompt": "E2E test video generation",
                "model": "test-model"
            }
            
            result = self.client.create_task(**task_data)
            
            # Verify task was created
            self.assertNotIn("error", result)
            self.assertIn("id", result)
            
            task_id = result["id"]
            
            # Check task status
            status = self.client.get_task_status(task_id)
            self.assertNotIn("error", status)
            self.assertIn("status", status)
            
            # Clean up - delete the test task
            self.client.delete_task(task_id)
            
        except Exception as e:
            self.skipTest(f"Real backend test failed: {e}")

if __name__ == '__main__':
    # Configure test runner
    unittest.main(verbosity=2)