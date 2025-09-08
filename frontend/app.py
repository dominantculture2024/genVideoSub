import streamlit as st
from streamlit_option_menu import option_menu
from components.task_manager import (
    create_task_form,
    display_task_list,
    display_task_status_checker
)
from utils.api_client import api_client

# 頁面配置
st.set_page_config(
    page_title="GenVideoSub - AI視頻生成服務",
    page_icon="🎬",
    layout="wide",
    initial_sidebar_state="expanded"
)

# 自定義CSS樣式
st.markdown("""
<style>
.main-header {
    text-align: center;
    padding: 1rem 0;
    background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
    color: white;
    border-radius: 10px;
    margin-bottom: 2rem;
}

.status-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 20px;
    font-size: 0.8rem;
    font-weight: bold;
}

.status-pending { background-color: #ffeaa7; color: #2d3436; }
.status-processing { background-color: #74b9ff; color: white; }
.status-completed { background-color: #00b894; color: white; }
.status-failed { background-color: #e17055; color: white; }
</style>
""", unsafe_allow_html=True)

def main():
    """主函數"""
    # 主標題
    st.markdown("""
    <div class="main-header">
        <h1>🎬 GenVideoSub - AI視頻生成服務</h1>
        <p>基於 fal.ai Kling Video API 的智能視頻生成平台</p>
    </div>
    """, unsafe_allow_html=True)
    
    # 側邊欄導航
    with st.sidebar:
        st.markdown("### 🚀 導航菜單")
        selected = option_menu(
            None,
            ["創建任務", "狀態查詢", "任務管理"],
            icons=["plus-circle-fill", "search", "list-task"],
            menu_icon="cast",
            default_index=0,
            styles={
                "container": {"padding": "0!important", "background-color": "#fafafa"},
                "icon": {"color": "#667eea", "font-size": "18px"}, 
                "nav-link": {
                    "font-size": "16px", 
                    "text-align": "left", 
                    "margin":"0px", 
                    "--hover-color": "#eee"
                },
                "nav-link-selected": {"background-color": "#667eea"},
            }
        )
        
        # 側邊欄信息
        st.markdown("---")
        st.markdown("### 📊 系統信息")
        
        # 檢查後端連接狀態
        try:
            result = api_client.list_tasks(limit=1)
            if "error" not in result:
                st.success("🟢 後端服務正常")
            else:
                st.error("🔴 後端服務異常")
        except:
            st.error("🔴 無法連接後端")
        
        st.markdown("""
        ### 📝 使用說明
        1. **創建任務**: 輸入提示詞生成視頻
        2. **狀態查詢**: 查看任務進度和結果
        3. **任務管理**: 管理所有視頻生成任務
        
        ### 🔧 技術支持
        - 基於 fal.ai Kling Video API
        - 支持多種視頻格式和比例
        - 異步任務處理
        """)
    
    # 主內容區域
    if selected == "創建任務":
        create_task_page()
    elif selected == "狀態查詢":
        status_query_page()
    elif selected == "任務管理":
        task_management_page()

def create_task_page():
    """創建任務頁面"""
    create_task_form()
    
    # 顯示最近任務
    st.markdown("---")
    st.subheader("📈 最近任務")
    
    try:
        result = api_client.list_tasks(limit=5)
        if "error" not in result and result.get("tasks"):
            tasks = result["tasks"][:5]
            for i, task in enumerate(tasks):
                with st.expander(f"任務 {i+1}: {task.get('id', '')[:8]}..."):
                    col1, col2, col3 = st.columns(3)
                    with col1:
                        st.write(f"**狀態**: {task.get('status', 'unknown')}")
                    with col2:
                        st.write(f"**時長**: {task.get('duration', 0)}s")
                    with col3:
                        st.write(f"**比例**: {task.get('aspect_ratio', '')}")
                    st.write(f"**提示詞**: {task.get('prompt', '')[:100]}...")
        else:
            st.info("暫無最近任務")
    except Exception as e:
        st.error(f"載入最近任務失敗: {str(e)}")

def status_query_page():
    """狀態查詢頁面"""
    display_task_status_checker()

def task_management_page():
    """任務管理頁面"""
    display_task_list()

if __name__ == "__main__":
    main()