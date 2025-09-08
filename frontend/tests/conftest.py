import pytest
import os
import sys
from unittest.mock import patch, Mock

# Add the frontend directory to the Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

@pytest.fixture
def mock_streamlit():
    """Fixture to mock streamlit components for testing"""
    with patch('streamlit') as mock_st:
        # Configure common streamlit mock behaviors
        mock_st.text_input.return_value = "Test input"
        mock_st.text_area.return_value = "Test area"
        mock_st.number_input.return_value = 1
        mock_st.selectbox.return_value = "Option 1"
        mock_st.button.return_value = False
        mock_st.file_uploader.return_value = None
        mock_st.columns.return_value = [Mock(), Mock(), Mock()]
        mock_st.session_state = {}
        
        # Mock display functions
        mock_st.write = Mock()
        mock_st.success = Mock()
        mock_st.error = Mock()
        mock_st.warning = Mock()
        mock_st.info = Mock()
        mock_st.progress = Mock()
        mock_st.spinner = Mock()
        
        yield mock_st

@pytest.fixture
def mock_api_client():
    """Fixture to mock API client for testing"""
    with patch('utils.api_client.APIClient') as mock_client_class:
        mock_client = Mock()
        mock_client_class.return_value = mock_client
        
        # Configure default mock responses
        mock_client.create_task.return_value = {
            "id": "test-task-123",
            "status": "pending",
            "prompt": "Test prompt"
        }
        
        mock_client.get_task.return_value = {
            "id": "test-task-123",
            "status": "completed",
            "prompt": "Test prompt"
        }
        
        mock_client.get_task_status.return_value = {
            "status": "completed",
            "progress": 100
        }
        
        mock_client.get_task_result.return_value = {
            "status": "completed",
            "video_url": "http://example.com/video.mp4"
        }
        
        mock_client.list_tasks.return_value = {
            "tasks": [],
            "total": 0,
            "offset": 0,
            "limit": 10
        }
        
        mock_client.upload_file.return_value = {
            "id": "file-123",
            "url": "http://example.com/file.jpg",
            "filename": "test.jpg"
        }
        
        mock_client.delete_task.return_value = {
            "success": True,
            "message": "Task deleted"
        }
        
        yield mock_client

@pytest.fixture
def test_config():
    """Fixture to provide test configuration"""
    return {
        "base_url": "http://localhost:8080",
        "timeout": 30,
        "test_mode": True,
        "mock_enabled": True
    }

@pytest.fixture
def sample_task_data():
    """Fixture to provide sample task data for testing"""
    return {
        "id": "sample-task-123",
        "status": "pending",
        "prompt": "Generate a test video",
        "model": "test-model",
        "created_at": "2024-01-15T10:30:45Z",
        "updated_at": "2024-01-15T10:30:45Z"
    }

@pytest.fixture
def sample_task_list():
    """Fixture to provide sample task list for testing"""
    return {
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

@pytest.fixture
def mock_file_upload():
    """Fixture to provide mock file upload data"""
    mock_file = Mock()
    mock_file.name = "test_image.jpg"
    mock_file.read.return_value = b"fake image data"
    mock_file.type = "image/jpeg"
    mock_file.size = 1024
    return mock_file

@pytest.fixture(autouse=True)
def setup_test_environment():
    """Automatically set up test environment for all tests"""
    # Set test environment variables
    os.environ['TESTING'] = 'true'
    os.environ['MOCK_ENABLED'] = 'true'
    
    yield
    
    # Clean up after test
    if 'TESTING' in os.environ:
        del os.environ['TESTING']
    if 'MOCK_ENABLED' in os.environ:
        del os.environ['MOCK_ENABLED']

@pytest.fixture
def error_responses():
    """Fixture to provide various error response scenarios"""
    return {
        "network_error": {"error": "Network connection failed", "code": 500},
        "not_found": {"error": "Task not found", "code": 404},
        "rate_limit": {"error": "Rate limit exceeded", "code": 429},
        "bad_request": {"error": "Invalid request parameters", "code": 400},
        "unauthorized": {"error": "Unauthorized access", "code": 401},
        "server_error": {"error": "Internal server error", "code": 500}
    }

# Pytest configuration
def pytest_configure(config):
    """Configure pytest with custom markers"""
    config.addinivalue_line(
        "markers", "unit: mark test as a unit test"
    )
    config.addinivalue_line(
        "markers", "integration: mark test as an integration test"
    )
    config.addinivalue_line(
        "markers", "e2e: mark test as an end-to-end test"
    )
    config.addinivalue_line(
        "markers", "slow: mark test as slow running"
    )
    config.addinivalue_line(
        "markers", "requires_backend: mark test as requiring real backend"
    )

def pytest_collection_modifyitems(config, items):
    """Modify test collection to add markers based on test location"""
    for item in items:
        # Add markers based on test file location
        if "unit" in str(item.fspath):
            item.add_marker(pytest.mark.unit)
        elif "integration" in str(item.fspath):
            item.add_marker(pytest.mark.integration)
        elif "e2e" in str(item.fspath):
            item.add_marker(pytest.mark.e2e)
        
        # Add slow marker for tests that might take longer
        if "concurrent" in item.name or "performance" in item.name:
            item.add_marker(pytest.mark.slow)
        
        # Add backend requirement marker
        if "real_backend" in item.name:
            item.add_marker(pytest.mark.requires_backend)

# Custom pytest hooks
@pytest.hookimpl(tryfirst=True)
def pytest_runtest_setup(item):
    """Setup hook that runs before each test"""
    # Skip tests that require real backend if not available
    if item.get_closest_marker("requires_backend"):
        if os.getenv('TEST_WITH_REAL_BACKEND') != 'true':
            pytest.skip("Real backend tests disabled")

@pytest.hookimpl(tryfirst=True)
def pytest_runtest_teardown(item):
    """Teardown hook that runs after each test"""
    # Clean up any test artifacts
    pass