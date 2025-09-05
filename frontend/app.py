import streamlit as st
import requests
import json
from datetime import datetime
from utils.api_client import APIClient

# 頁面配置
st.set_page_config(
    page_title="genVideo&Sub - AI視頻生成平台",
    page_icon="🎬",
    layout="wide",
    initial_sidebar_state="expanded"
)

# 初始化API客戶端
api_client = APIClient()

def main():
    st.title("🎬 genVideo&Sub - AI視頻生成平台")
    st.markdown("---")
    
    # 側邊欄配置
    with st.sidebar:
        st.header("⚙️ 系統配置")
        
        # API服務器配置
        api_url = st.text_input(
            "API服務器地址", 
            value="http://localhost:8080",
            help="後端API服務器的地址"
        )
        api_client.set_base_url(api_url)
        
        # 檢查API連接狀態
        if st.button("🔍 檢查API連接"):
            status = api_client.check_health()
            if status:
                st.success("✅ API連接正常")
            else:
                st.error("❌ API連接失敗")
    
    # 主要功能區域
    tab1, tab2, tab3 = st.tabs(["📤 創建任務", "📊 任務狀態", "📋 任務列表"])
    
    with tab1:
        create_task_interface()
    
    with tab2:
        task_status_interface()
    
    with tab3:
        task_list_interface()

def create_task_interface():
    """創建任務界面"""
    st.header("創建新的視頻生成任務")
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader("基本信息")
        task_id = st.number_input("任務ID", min_value=1, value=1)
        image_path = st.text_input(
            "圖片URL", 
            placeholder="https://example.com/image.jpg",
            help="輸入圖片的URL地址"
        )
        prompt = st.text_area(
            "AI生成提示詞", 
            placeholder="描述您想要生成的視頻內容...",
            height=100
        )
    
    with col2:
        st.subheader("字幕設置")
        subtitle = st.text_area(
            "字幕內容", 
            placeholder="輸入字幕文字...",
            height=100
        )
        subtitle_color = st.selectbox(
            "字幕顏色", 
            ["white", "black", "red", "blue", "green", "yellow"]
        )
        subtitle_position = st.selectbox(
            "字幕位置", 
            ["bottom", "top", "center", "left", "right"]
        )
    
    # 預覽區域
    if image_path:
        st.subheader("圖片預覽")
        try:
            st.image(image_path, caption="輸入圖片", use_column_width=True)
        except:
            st.error("無法載入圖片，請檢查URL是否正確")
    
    # 提交按鈕
    if st.button("🚀 創建任務", type="primary"):
        if not all([task_id, image_path, subtitle, prompt]):
            st.error("請填寫所有必填欄位")
            return
        
        task_data = {
            "id": task_id,
            "image_path": image_path,
            "subtitle_color": subtitle_color,
            "subtitle_position": subtitle_position,
            "subtitle": subtitle,
            "prompt": prompt
        }
        
        with st.spinner("正在創建任務..."):
            result = api_client.create_task(task_data)
            if result:
                st.success(f"✅ 任務創建成功！外部ID: {result.get('external_id', 'N/A')}")
                st.json(result)
            else:
                st.error("❌ 任務創建失敗")

def task_status_interface():
    """任務狀態查詢界面"""
    st.header("查詢任務狀態")
    
    task_id = st.number_input("輸入任務ID", min_value=1, value=1)
    
    if st.button("🔍 查詢狀態"):
        with st.spinner("正在查詢任務狀態..."):
            status = api_client.get_task_status(task_id)
            if status:
                st.success("✅ 查詢成功")
                
                # 狀態顯示
                col1, col2, col3 = st.columns(3)
                with col1:
                    st.metric("任務ID", status.get('id', 'N/A'))
                with col2:
                    st.metric("狀態", status.get('status', 'N/A'))
                with col3:
                    st.metric("外部ID", status.get('external_id', 'N/A'))
                
                # 詳細信息
                st.json(status)
            else:
                st.error("❌ 查詢失敗或任務不存在")

def task_list_interface():
    """任務列表界面"""
    st.header("任務列表")
    
    if st.button("🔄 刷新列表"):
        with st.spinner("正在載入任務列表..."):
            tasks = api_client.get_task_list()
            if tasks:
                st.success(f"✅ 找到 {len(tasks)} 個任務")
                
                # 顯示任務表格
                if tasks:
                    import pandas as pd
                    df = pd.DataFrame(tasks)
                    st.dataframe(df, use_container_width=True)
                else:
                    st.info("📝 暫無任務")
            else:
                st.error("❌ 載入任務列表失敗")

if __name__ == "__main__":
    main()