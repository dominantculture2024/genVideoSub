import streamlit as st
import pandas as pd
from datetime import datetime
from typing import Dict, List, Any
from utils.api_client import api_client
import time

def display_task_status(status: str) -> str:
    """Display task status with appropriate styling"""
    status_colors = {
        "pending": "🟡",
        "processing": "🔵", 
        "completed": "🟢",
        "failed": "🔴"
    }
    return f"{status_colors.get(status, '⚪')} {status.upper()}"

def format_datetime(timestamp: str) -> str:
    """Format timestamp for display"""
    try:
        dt = datetime.fromisoformat(timestamp.replace('Z', '+00:00'))
        return dt.strftime('%Y-%m-%d %H:%M:%S')
    except:
        return timestamp

def create_task_form():
    """Create task creation form"""
    st.subheader("🎬 創建視頻生成任務")
    
    with st.form("create_task_form"):
        col1, col2 = st.columns(2)
        
        with col1:
            prompt = st.text_area(
                "提示詞 *", 
                placeholder="描述您想要生成的視頻內容...",
                height=100
            )
            
            duration = st.slider(
                "視頻時長 (秒)", 
                min_value=1.0, 
                max_value=10.0, 
                value=5.0, 
                step=0.5
            )
            
            aspect_ratio = st.selectbox(
                "寬高比", 
                options=["16:9", "9:16", "1:1"],
                index=0
            )
        
        with col2:
            # File upload for reference image
            uploaded_file = st.file_uploader(
                "參考圖片 (可選)",
                type=["jpg", "jpeg", "png"],
                help="上傳參考圖片來指導視頻生成"
            )
            
            negative_prompt = st.text_area(
                "負面提示詞 (可選)",
                placeholder="描述您不希望在視頻中出現的內容...",
                height=60
            )
            
            cfg_scale = st.slider(
                "CFG Scale", 
                min_value=1.0, 
                max_value=20.0, 
                value=7.0, 
                step=0.5,
                help="控制生成內容與提示詞的相符程度"
            )
        
        submitted = st.form_submit_button("🚀 創建任務", use_container_width=True)
        
        if submitted:
            if not prompt.strip():
                st.error("請輸入提示詞")
                return
            
            with st.spinner("正在創建任務..."):
                # Handle file upload if provided
                image_url = None
                if uploaded_file is not None:
                    file_data = uploaded_file.read()
                    upload_result = api_client.upload_file(file_data, uploaded_file.name)
                    if "error" not in upload_result:
                        image_url = upload_result.get("url")
                    else:
                        st.error(f"文件上傳失敗: {upload_result['error']}")
                        return
                
                # Create task
                result = api_client.create_task(
                    prompt=prompt,
                    image_url=image_url,
                    duration=duration,
                    aspect_ratio=aspect_ratio,
                    negative_prompt=negative_prompt if negative_prompt.strip() else None,
                    cfg_scale=cfg_scale
                )
                
                if "error" not in result:
                    st.success(f"任務創建成功！任務ID: {result.get('id')}")
                    st.rerun()
                else:
                    st.error(f"任務創建失敗: {result['error']}")

def display_task_list():
    """Display list of all tasks"""
    st.subheader("📋 任務列表")
    
    # Refresh button
    col1, col2, col3 = st.columns([1, 1, 4])
    with col1:
        if st.button("🔄 刷新", use_container_width=True):
            st.rerun()
    
    with col2:
        auto_refresh = st.checkbox("自動刷新", value=False)
    
    if auto_refresh:
        time.sleep(2)
        st.rerun()
    
    # Get tasks
    with st.spinner("載入任務列表..."):
        result = api_client.list_tasks(limit=100)
        
        if "error" in result:
            st.error(f"載入任務失敗: {result['error']}")
            return
        
        tasks = result.get("tasks", [])
        
        if not tasks:
            st.info("暫無任務")
            return
        
        # Create DataFrame for display
        task_data = []
        for task in tasks:
            task_data.append({
                "ID": task.get("id", "")[:8] + "...",
                "狀態": display_task_status(task.get("status", "unknown")),
                "提示詞": task.get("prompt", "")[:50] + "..." if len(task.get("prompt", "")) > 50 else task.get("prompt", ""),
                "時長": f"{task.get('duration', 0)}s",
                "寬高比": task.get("aspect_ratio", ""),
                "創建時間": format_datetime(task.get("created_at", "")),
                "完整ID": task.get("id", "")
            })
        
        df = pd.DataFrame(task_data)
        
        # Display table
        st.dataframe(
            df.drop(columns=["完整ID"]),
            use_container_width=True,
            hide_index=True
        )
        
        # Task details section
        st.subheader("📊 任務詳情")
        
        # Task selection
        selected_task_id = st.selectbox(
            "選擇任務查看詳情",
            options=[task["完整ID"] for task in task_data],
            format_func=lambda x: f"{x[:8]}... - {next(task['提示詞'] for task in task_data if task['完整ID'] == x)}"
        )
        
        if selected_task_id:
            display_task_details(selected_task_id)

def display_task_details(task_id: str):
    """Display detailed information for a specific task"""
    col1, col2 = st.columns(2)
    
    with col1:
        if st.button("📊 查看狀態", use_container_width=True):
            with st.spinner("載入任務狀態..."):
                result = api_client.get_task_status(task_id)
                if "error" not in result:
                    st.json(result)
                else:
                    st.error(f"載入失敗: {result['error']}")
    
    with col2:
        if st.button("🎬 查看結果", use_container_width=True):
            with st.spinner("載入任務結果..."):
                result = api_client.get_task_result(task_id)
                if "error" not in result:
                    if result.get("status") == "completed" and result.get("video_url"):
                        st.success("視頻生成完成！")
                        st.video(result["video_url"])
                        st.markdown(f"**下載鏈接:** [點擊下載]({result['video_url']})")
                    else:
                        st.info("任務尚未完成或無結果")
                        st.json(result)
                else:
                    st.error(f"載入失敗: {result['error']}")
    
    # Delete task button
    if st.button("🗑️ 刪除任務", type="secondary"):
        if st.confirm("確定要刪除此任務嗎？"):
            result = api_client.delete_task(task_id)
            if "error" not in result:
                st.success("任務已刪除")
                st.rerun()
            else:
                st.error(f"刪除失敗: {result['error']}")

def display_task_status_checker():
    """Display task status checker interface"""
    st.subheader("🔍 任務狀態查詢")
    
    task_id = st.text_input(
        "任務ID",
        placeholder="輸入完整的任務ID...",
        help="輸入您要查詢的任務ID"
    )
    
    col1, col2 = st.columns(2)
    
    with col1:
        if st.button("查詢狀態", use_container_width=True) and task_id:
            with st.spinner("查詢中..."):
                result = api_client.get_task_status(task_id)
                if "error" not in result:
                    st.success("查詢成功")
                    
                    # Display status info
                    status = result.get("status", "unknown")
                    st.markdown(f"**狀態:** {display_task_status(status)}")
                    
                    if "progress" in result:
                        st.progress(result["progress"] / 100.0)
                        st.markdown(f"**進度:** {result['progress']}%")
                    
                    if "message" in result:
                        st.markdown(f"**消息:** {result['message']}")
                    
                    # Show full response
                    with st.expander("完整回應"):
                        st.json(result)
                else:
                    st.error(f"查詢失敗: {result['error']}")
    
    with col2:
        if st.button("查詢結果", use_container_width=True) and task_id:
            with st.spinner("查詢中..."):
                result = api_client.get_task_result(task_id)
                if "error" not in result:
                    if result.get("status") == "completed":
                        st.success("任務已完成！")
                        
                        if result.get("video_url"):
                            st.video(result["video_url"])
                            st.markdown(f"**下載鏈接:** [點擊下載]({result['video_url']})")
                        
                        # Show metadata
                        with st.expander("視頻信息"):
                            st.json(result)
                    else:
                        st.info(f"任務狀態: {result.get('status', 'unknown')}")
                        st.json(result)
                else:
                    st.error(f"查詢失敗: {result['error']}")