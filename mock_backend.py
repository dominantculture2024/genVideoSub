#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
GenVideoSub Mock Backend Server
用於測試的簡單mock後端服務
"""

import json
import time
import uuid
from datetime import datetime
from flask import Flask, request, jsonify
from flask_cors import CORS
import os

app = Flask(__name__)
CORS(app)  # 允許跨域請求

# 模擬數據存儲
tasks = {}
files = {}

# Mock任務狀態
TASK_STATUS = {
    'PENDING': 'pending',
    'IN_PROGRESS': 'in_progress', 
    'COMPLETED': 'completed',
    'FAILED': 'failed'
}

# Mock任務數據
def create_mock_task(prompt, duration=5):
    task_id = str(uuid.uuid4())
    task = {
        'id': task_id,
        'prompt': prompt,
        'status': TASK_STATUS['PENDING'],
        'created_at': datetime.now().isoformat(),
        'updated_at': datetime.now().isoformat(),
        'duration': duration,
        'progress': 0,
        'result_url': None,
        'error_message': None
    }
    tasks[task_id] = task
    return task

# 模擬任務狀態更新
def update_task_status(task_id):
    if task_id not in tasks:
        return False
    
    task = tasks[task_id]
    current_time = datetime.now()
    created_time = datetime.fromisoformat(task['created_at'])
    elapsed = (current_time - created_time).total_seconds()
    
    if elapsed < 2:  # 前2秒為pending
        task['status'] = TASK_STATUS['PENDING']
        task['progress'] = 0
    elif elapsed < task['duration']:  # 處理中
        task['status'] = TASK_STATUS['IN_PROGRESS']
        task['progress'] = int((elapsed / task['duration']) * 100)
    else:  # 完成
        task['status'] = TASK_STATUS['COMPLETED']
        task['progress'] = 100
        task['result_url'] = f"https://example.com/videos/{task_id}.mp4"
    
    task['updated_at'] = current_time.isoformat()
    return True

# 健康檢查
@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({
        'status': 'ok',
        'message': 'GenVideoSub Mock Backend is running',
        'timestamp': int(time.time()),
        'version': '1.0.0-mock'
    })

# 創建任務
@app.route('/api/tasks', methods=['POST'])
def create_task():
    try:
        data = request.get_json()
        if not data or 'prompt' not in data:
            return jsonify({'error': 'Missing prompt parameter'}), 400
        
        prompt = data['prompt']
        duration = data.get('duration', 5)  # 默認5秒完成
        
        task = create_mock_task(prompt, duration)
        
        return jsonify({
            'success': True,
            'task_id': task['id'],
            'status': task['status'],
            'message': 'Task created successfully'
        }), 201
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 獲取任務列表
@app.route('/api/tasks', methods=['GET'])
def list_tasks():
    try:
        # 更新所有任務狀態
        for task_id in tasks:
            update_task_status(task_id)
        
        task_list = list(tasks.values())
        return jsonify({
            'success': True,
            'tasks': task_list,
            'total': len(task_list)
        })
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 獲取單個任務
@app.route('/api/tasks/<task_id>', methods=['GET'])
def get_task(task_id):
    try:
        if task_id not in tasks:
            return jsonify({'error': 'Task not found'}), 404
        
        update_task_status(task_id)
        task = tasks[task_id]
        
        return jsonify({
            'success': True,
            'task': task
        })
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 獲取任務狀態
@app.route('/api/tasks/<task_id>/status', methods=['GET'])
def get_task_status(task_id):
    try:
        if task_id not in tasks:
            return jsonify({'error': 'Task not found'}), 404
        
        update_task_status(task_id)
        task = tasks[task_id]
        
        return jsonify({
            'success': True,
            'task_id': task_id,
            'status': task['status'],
            'progress': task['progress'],
            'updated_at': task['updated_at']
        })
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 獲取任務結果
@app.route('/api/tasks/<task_id>/result', methods=['GET'])
def get_task_result(task_id):
    try:
        if task_id not in tasks:
            return jsonify({'error': 'Task not found'}), 404
        
        update_task_status(task_id)
        task = tasks[task_id]
        
        if task['status'] != TASK_STATUS['COMPLETED']:
            return jsonify({
                'success': False,
                'message': 'Task not completed yet',
                'status': task['status']
            }), 202
        
        return jsonify({
            'success': True,
            'task_id': task_id,
            'result_url': task['result_url'],
            'status': task['status']
        })
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 刪除任務
@app.route('/api/tasks/<task_id>', methods=['DELETE'])
def delete_task(task_id):
    try:
        if task_id not in tasks:
            return jsonify({'error': 'Task not found'}), 404
        
        del tasks[task_id]
        
        return jsonify({
            'success': True,
            'message': 'Task deleted successfully'
        })
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 文件上傳（模擬）
@app.route('/api/files/upload', methods=['POST'])
def upload_file():
    try:
        if 'file' not in request.files:
            return jsonify({'error': 'No file provided'}), 400
        
        file = request.files['file']
        if file.filename == '':
            return jsonify({'error': 'No file selected'}), 400
        
        file_id = str(uuid.uuid4())
        file_info = {
            'id': file_id,
            'filename': file.filename,
            'size': len(file.read()),
            'uploaded_at': datetime.now().isoformat(),
            'url': f'/api/files/{file_id}'
        }
        files[file_id] = file_info
        
        return jsonify({
            'success': True,
            'file_id': file_id,
            'file_info': file_info
        }), 201
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# 獲取文件信息
@app.route('/api/files/<file_id>/info', methods=['GET'])
def get_file_info(file_id):
    try:
        if file_id not in files:
            return jsonify({'error': 'File not found'}), 404
        
        return jsonify({
            'success': True,
            'file_info': files[file_id]
        })
        
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    print("🚀 Starting GenVideoSub Mock Backend Server...")
    print("📍 Server will run on: http://localhost:8080")
    print("🔧 Mock mode enabled - no real API calls will be made")
    print("📊 Health check: http://localhost:8080/health")
    print("📝 API docs: All endpoints support CORS for frontend integration")
    
    app.run(
        host='localhost',
        port=8080,
        debug=True,
        threaded=True
    )