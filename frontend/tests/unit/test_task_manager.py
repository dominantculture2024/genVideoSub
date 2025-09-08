import unittest
from unittest.mock import Mock, patch, MagicMock
import sys
import os
import pandas as pd
from datetime import datetime

# Add the frontend directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))

from components.task_manager import (
    display_task_status,
    format_datetime,
    create_task_form,
    display_task_list,
    display_task_details
)

class TestTaskManagerUtils(unittest.TestCase):
    """Test utility functions in task_manager"""
    
    def test_display_task_status_known_statuses(self):
        """Test display_task_status with known statuses"""
        test_cases = [
            ("pending", "🟡 PENDING"),
            ("processing", "🔵 PROCESSING"),
            ("completed", "🟢 COMPLETED"),
            ("failed", "🔴 FAILED")
        ]
        
        for status, expected in test_cases:
            with self.subTest(status=status):
                result = display_task_status(status)
                self.assertEqual(result, expected)
    
    def test_display_task_status_unknown_status(self):
        """Test display_task_status with unknown status"""
        result = display_task_status("unknown")
        self.assertEqual(result, "⚪ UNKNOWN")
    
    def test_format_datetime_valid_iso(self):
        """Test format_datetime with valid ISO timestamp"""
        timestamp = "2024-01-15T10:30:45Z"
        result = format_datetime(timestamp)
        # Should return formatted datetime string
        self.assertRegex(result, r"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}")
    
    def test_format_datetime_invalid(self):
        """Test format_datetime with invalid timestamp"""
        timestamp = "invalid-timestamp"
        result = format_datetime(timestamp)
        self.assertEqual(result, timestamp)
    
    def test_format_datetime_empty(self):
        """Test format_datetime with empty timestamp"""
        timestamp = ""
        result = format_datetime(timestamp)
        self.assertEqual(result, timestamp)

class TestTaskManagerComponents(unittest.TestCase):
    """Test Streamlit components in task_manager"""
    
    def setUp(self):
        """Set up test fixtures"""
        # Mock streamlit
        self.mock_st = Mock()
        self.mock_api_client = Mock()
        
        # Patch streamlit and api_client
        self.st_patcher = patch('components.task_manager.st', self.mock_st)
        self.api_patcher = patch('components.task_manager.api_client', self.mock_api_client)
        
        self.st_patcher.start()
        self.api_patcher.start()
    
    def tearDown(self):
        """Clean up patches"""
        self.st_patcher.stop()
        self.api_patcher.stop()
    
    def test_create_task_form_setup(self):
        """Test create_task_form UI setup"""
        # Mock form context manager
        form_mock = MagicMock()
        form_context = MagicMock()
        form_context.__enter__ = MagicMock(return_value=form_mock)
        form_context.__exit__ = MagicMock(return_value=None)
        self.mock_st.form.return_value = form_context
        
        # Mock columns with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        
        # Mock form inputs
        self.mock_st.text_area.return_value = "test prompt"
        self.mock_st.slider.return_value = 5.0
        self.mock_st.selectbox.return_value = "16:9"
        self.mock_st.file_uploader.return_value = None
        self.mock_st.form_submit_button.return_value = False
        
        create_task_form()
        
        # Verify form components were created
        self.mock_st.form.assert_called_once_with("create_task_form")
        self.mock_st.columns.assert_called_once_with(2)
        self.assertEqual(self.mock_st.text_area.call_count, 2)  # 提示詞 and 負面提示詞
        self.mock_st.slider.assert_called()
        self.mock_st.selectbox.assert_called_once()
        self.mock_st.file_uploader.assert_called_once()
        self.mock_st.form_submit_button.assert_called_once()
    
    def test_create_task_form_submit_success(self):
        """Test create_task_form successful submission"""
        # Mock form context manager
        form_mock = MagicMock()
        form_context = MagicMock()
        form_context.__enter__ = MagicMock(return_value=form_mock)
        form_context.__exit__ = MagicMock(return_value=None)
        self.mock_st.form.return_value = form_context
        
        # Mock columns with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        
        # Mock form inputs
        self.mock_st.text_area.return_value = "test prompt"
        self.mock_st.slider.return_value = 5.0
        self.mock_st.selectbox.return_value = "16:9"
        self.mock_st.file_uploader.return_value = None
        self.mock_st.form_submit_button.return_value = True  # Submit clicked
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock API response
        self.mock_api_client.create_task.return_value = {"id": "task-123"}
        
        create_task_form()
        
        # Verify API was called and success message shown
        self.mock_api_client.create_task.assert_called_once()
        self.mock_st.success.assert_called_with("任務創建成功！任務ID: task-123")
    
    def test_create_task_form_submit_empty_prompt(self):
        """Test create_task_form with empty prompt"""
        # Mock form context manager
        form_mock = MagicMock()
        form_context = MagicMock()
        form_context.__enter__ = MagicMock(return_value=form_mock)
        form_context.__exit__ = MagicMock(return_value=None)
        self.mock_st.form.return_value = form_context
        
        # Mock columns with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        
        # Mock form inputs with empty prompt
        self.mock_st.text_area.side_effect = ["", ""]  # Empty prompt and negative prompt
        self.mock_st.slider.side_effect = [5.0, 7.0]  # duration and cfg_scale
        self.mock_st.selectbox.return_value = "16:9"
        self.mock_st.file_uploader.return_value = None
        self.mock_st.form_submit_button.return_value = True  # Submit clicked
        
        create_task_form()
        
        # Verify error message was shown and API not called
        self.mock_st.error.assert_called_with("請輸入提示詞")
        self.mock_api_client.create_task.assert_not_called()
    
    def test_create_task_form_with_file_upload(self):
        """Test task creation with file upload"""
        # Mock form context manager
        form_mock = MagicMock()
        form_context = MagicMock()
        form_context.__enter__ = MagicMock(return_value=form_mock)
        form_context.__exit__ = MagicMock(return_value=None)
        self.mock_st.form.return_value = form_context
        
        # Mock columns with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        
        # Mock uploaded file
        mock_file = Mock()
        mock_file.name = "test.jpg"
        mock_file.read.return_value = b"file data"
        
        # Mock form inputs
        self.mock_st.text_area.return_value = "test prompt"
        self.mock_st.slider.side_effect = [5.0, 7.0]
        self.mock_st.selectbox.return_value = "16:9"
        self.mock_st.file_uploader.return_value = mock_file
        self.mock_st.form_submit_button.return_value = True
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock API responses
        self.mock_api_client.upload_file.return_value = {"url": "http://example.com/image.jpg"}
        self.mock_api_client.create_task.return_value = {"id": "task-123"}
        
        create_task_form()
        
        # Verify file upload and task creation
        self.mock_api_client.upload_file.assert_called_once_with(b"file data", "test.jpg")
        self.mock_api_client.create_task.assert_called_once()
        
        # Check that image_url was passed to create_task
        call_args = self.mock_api_client.create_task.call_args
        self.assertIn('image_url', call_args.kwargs)
        self.assertEqual(call_args.kwargs['image_url'], "http://example.com/image.jpg")
    
    def test_display_task_list_no_tasks(self):
        """Test display_task_list with no tasks"""
        # Mock API response with no tasks
        self.mock_api_client.list_tasks.return_value = {"tasks": []}
        
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        col3_mock = MagicMock()
        col3_mock.__enter__ = MagicMock(return_value=col3_mock)
        col3_mock.__exit__ = MagicMock(return_value=None)
        
        # Mock columns to return different numbers based on call
        def mock_columns_side_effect(num_cols):
            if num_cols == 2:
                return [col1_mock, col2_mock]
            elif num_cols == 3 or isinstance(num_cols, list):
                return [col1_mock, col2_mock, col3_mock]
            else:
                return [col1_mock, col2_mock, col3_mock]
        
        self.mock_st.columns.side_effect = mock_columns_side_effect
        self.mock_st.button.return_value = False
        self.mock_st.checkbox.return_value = False
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        display_task_list()
        
        # Verify info message was shown
        self.mock_st.info.assert_called_with("暫無任務")
    
    def test_display_task_list_with_tasks(self):
        """Test display_task_list with tasks"""
        # Mock API response with tasks
        mock_tasks = [
            {"id": "task-1", "status": "completed", "prompt": "Test prompt 1", "duration": 5.0, "aspect_ratio": "16:9", "created_at": "2024-01-01T00:00:00Z"},
            {"id": "task-2", "status": "processing", "prompt": "Test prompt 2", "duration": 3.0, "aspect_ratio": "9:16", "created_at": "2024-01-01T01:00:00Z"}
        ]
        self.mock_api_client.list_tasks.return_value = {"tasks": mock_tasks}
        
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        col3_mock = MagicMock()
        col3_mock.__enter__ = MagicMock(return_value=col3_mock)
        col3_mock.__exit__ = MagicMock(return_value=None)
        
        # Mock columns to return different numbers based on call
        def mock_columns_side_effect(num_cols):
            if num_cols == 2:
                return [col1_mock, col2_mock]
            elif num_cols == 3 or isinstance(num_cols, list):
                return [col1_mock, col2_mock, col3_mock]
            else:
                return [col1_mock, col2_mock, col3_mock]
        
        self.mock_st.columns.side_effect = mock_columns_side_effect
        self.mock_st.button.return_value = False
        self.mock_st.checkbox.return_value = False
        self.mock_st.selectbox.return_value = None  # No task selected
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock pandas DataFrame
        with patch('components.task_manager.pd.DataFrame') as mock_df:
            mock_df_instance = Mock()
            mock_df.return_value = mock_df_instance
            mock_df_instance.drop.return_value = mock_df_instance
            
            display_task_list()
            
            # Verify DataFrame was created and displayed
            mock_df.assert_called_once()
            self.mock_st.dataframe.assert_called_once()
    
    def test_display_task_list_api_error(self):
        """Test display_task_list with API error"""
        # Mock API error response instead of exception
        self.mock_api_client.list_tasks.return_value = {"error": "API Error"}
        
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        col3_mock = MagicMock()
        col3_mock.__enter__ = MagicMock(return_value=col3_mock)
        col3_mock.__exit__ = MagicMock(return_value=None)
        
        # Mock columns to return different numbers based on call
        def mock_columns_side_effect(num_cols):
            if num_cols == 2:
                return [col1_mock, col2_mock]
            elif num_cols == 3 or isinstance(num_cols, list):
                return [col1_mock, col2_mock, col3_mock]
            else:
                return [col1_mock, col2_mock, col3_mock]
        
        self.mock_st.columns.side_effect = mock_columns_side_effect
        self.mock_st.button.return_value = False
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        display_task_list()
        
        # Verify error message was shown
        self.mock_st.error.assert_called_with("載入任務失敗: API Error")
    
    def test_display_task_details_status_check(self):
        """Test display_task_details status check"""
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        self.mock_st.button.side_effect = [True, False, False]  # First button (status) clicked, others not, others not
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock API response
        self.mock_api_client.get_task_status.return_value = {"status": "processing", "progress": 50}
        
        display_task_details("task-123")
        
        # Verify API was called and result displayed
        self.mock_api_client.get_task_status.assert_called_once_with("task-123")
        self.mock_st.json.assert_called_once_with({"status": "processing", "progress": 50})
    
    def test_display_task_details_result_check_completed(self):
        """Test display_task_details result check for completed task"""
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        self.mock_st.button.side_effect = [False, True, False]  # Second button (result) clicked, others not
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock API response for completed task
        self.mock_api_client.get_task_result.return_value = {
            "status": "completed",
            "video_url": "http://example.com/video.mp4"
        }
        
        display_task_details("task-123")
        
        # Verify API was called and video displayed
        self.mock_api_client.get_task_result.assert_called_once_with("task-123")
        self.mock_st.success.assert_called_once_with("視頻生成完成！")
        self.mock_st.video.assert_called_once_with("http://example.com/video.mp4")
        self.mock_st.markdown.assert_called()
    
    def test_display_task_details_result_check_not_completed(self):
        """Test display_task_details result check for non-completed task"""
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        self.mock_st.button.side_effect = [False, True, False]  # Second button (result) clicked, others not
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock API response for non-completed task
        self.mock_api_client.get_task_result.return_value = {
            "status": "processing",
            "message": "Task is still processing"
        }
        
        display_task_details("task-123")
        
        # Verify API was called but no video displayed
        self.mock_api_client.get_task_result.assert_called_once_with("task-123")
        self.mock_st.video.assert_not_called()
    
    def test_display_task_details_api_error(self):
        """Test display_task_details with API error"""
        # Mock UI components with context manager support
        col1_mock = MagicMock()
        col1_mock.__enter__ = MagicMock(return_value=col1_mock)
        col1_mock.__exit__ = MagicMock(return_value=None)
        col2_mock = MagicMock()
        col2_mock.__enter__ = MagicMock(return_value=col2_mock)
        col2_mock.__exit__ = MagicMock(return_value=None)
        
        self.mock_st.columns.return_value = [col1_mock, col2_mock]
        self.mock_st.button.side_effect = [True, False, False]  # First button (status) clicked, others not
        
        # Mock spinner context manager
        mock_spinner = Mock()
        mock_spinner.__enter__ = Mock(return_value=mock_spinner)
        mock_spinner.__exit__ = Mock(return_value=False)
        self.mock_st.spinner.return_value = mock_spinner
        
        # Mock API error response
        self.mock_api_client.get_task_status.return_value = {"error": "Task not found"}
        
        display_task_details("task-123")
        
        # Verify error was shown
        self.mock_st.error.assert_called_with("載入失敗: Task not found")

class TestTaskManagerDataProcessing(unittest.TestCase):
    """Test data processing functions in task_manager"""
    
    def test_task_data_transformation(self):
        """Test task data transformation for display"""
        # Mock task data
        mock_tasks = [
            {
                "id": "task-123-456-789-abc-def",
                "status": "completed",
                "prompt": "A very long prompt that should be truncated because it exceeds fifty characters",
                "duration": 5.0,
                "aspect_ratio": "16:9",
                "created_at": "2024-01-15T10:30:45Z"
            }
        ]
        
        # Calculate expected truncated prompt (50 chars + "...")
        original_prompt = "A very long prompt that should be truncated because it exceeds fifty characters"
        expected_prompt = original_prompt[:50] + "..."
        
        # Expected transformation
        expected_data = {
            "ID": "task-123...",
            "狀態": "🟢 COMPLETED",
            "提示詞": expected_prompt,
            "時長": "5.0s",
            "寬高比": "16:9",
            "創建時間": format_datetime("2024-01-15T10:30:45Z"),
            "完整ID": "task-123-456-789-abc-def"
        }
        
        # Simulate the data transformation logic from display_task_list
        task_data = []
        for task in mock_tasks:
            prompt = task.get("prompt", "")
            truncated_prompt = prompt[:50] + "..." if len(prompt) > 50 else prompt
            
            task_data.append({
                "ID": task.get("id", "")[:8] + "...",
                "狀態": display_task_status(task.get("status", "unknown")),
                "提示詞": truncated_prompt,
                "時長": f"{task.get('duration', 0)}s",
                "寬高比": task.get("aspect_ratio", ""),
                "創建時間": format_datetime(task.get("created_at", "")),
                "完整ID": task.get("id", "")
            })
        
        # Verify transformation
        self.assertEqual(len(task_data), 1)
        self.assertEqual(task_data[0]["ID"], expected_data["ID"])
        self.assertEqual(task_data[0]["狀態"], expected_data["狀態"])
        self.assertEqual(task_data[0]["提示詞"], expected_data["提示詞"])
        self.assertEqual(task_data[0]["時長"], expected_data["時長"])
        self.assertEqual(task_data[0]["寬高比"], expected_data["寬高比"])
        self.assertEqual(task_data[0]["完整ID"], expected_data["完整ID"])

if __name__ == '__main__':
    unittest.main()