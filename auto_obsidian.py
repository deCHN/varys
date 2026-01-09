import os
import subprocess
import datetime
import mlx_whisper
import ollama
import sys
import shutil
import re

# ==================== 用户配置区域 ====================
# Obsidian 收件箱路径
OBSIDIAN_VAULT_PATH = "/Users/xnc/vault/Inbox"

# 音频附件存放的子文件夹名称
ASSETS_FOLDER_NAME = "assets"

# 模型选择 (建议使用 qwen2.5:7b 或 llama3:8b)
OLLAMA_MODEL = "qwen2.5:7b"
WHISPER_MODEL = "mlx-community/whisper-large-v3-turbo"

# 内部处理分块大小
INTERNAL_PROCESS_CHUNK = 2000
# ====================================================

def sanitize_filename(name):
    """清理文件名中的非法字符"""
    # 替换掉 / \ : * ? " < > | 为下划线或空
    name = re.sub(r'[\\/*?:"<>|]', "", name)
    # 去除换行符和多余空格
    name = name.replace("\n", "").replace("\r", "").strip()
    # 限制长度
    return name[:80]

def get_video_info(url):
    """获取视频标题 (带 Cookie，防止因未登录导致获取失败)"""
    print("🔍 正在获取视频标题...")
    try:
        # 修复：添加 --cookies-from-browser 参数，与下载保持一致
        cmd = [
            "yt-dlp",
            "--get-title",
            "--cookies-from-browser", "chrome",
            "--no-warnings",
            url
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        title = result.stdout.strip()

        if not title:
            raise ValueError("获取到的标题为空")

        safe_title = sanitize_filename(title)
        print(f"📄 标题获取成功: {safe_title}")
        return safe_title
    except Exception as e:
        print(f"⚠️ 标题获取失败 (将使用时间戳代替)。错误信息: {e}")
        return f"素材_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}"

def check_is_duplicate(target_filename):
    """检查文件是否已存在"""
    file_path = os.path.join(OBSIDIAN_VAULT_PATH, f"{target_filename}.md")
    if os.path.exists(file_path):
        print(f"⚠️ 跳过: 笔记 [{target_filename}.md] 已存在。")
        return True
    return False

def download_audio(url, temp_filename):
    """下载音频"""
    print(f"⬇️ [1/4] 正在下载音频...")
    output_template = f"{temp_filename}.%(ext)s"

    cmd = [
        "yt-dlp",
        "-x", "--audio-format", "m4a",
        "--cookies-from-browser", "chrome",
        "-o", output_template,
        "--no-playlist",
        "--newline", # 简化输出
        url
    ]

    try:
        # 这一步不需要实时显示详细进度条，只要不报错就行
        subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)

        # 查找下载的文件
        for file in os.listdir("."):
            if file.startswith(temp_filename) and file.endswith(".m4a"):
                return file
        return None
    except subprocess.CalledProcessError as e:
        print(f"\n❌ 下载出错: {e.stderr.decode()}")
        return None

def transcribe_audio(audio_file):
    """Whisper 转录"""
    print("\n🎙️ [2/4] 正在转录 (MLX加速中)...")
    # verbose=False 减少刷屏，只看结果
    result = mlx_whisper.transcribe(
        audio_file,
        path_or_hf_repo=WHISPER_MODEL,
        verbose=True
    )
    return result

def generate_intelligence(full_text):
    """生成中文摘要与标签"""
    print("\n🧠 [3/4] 正在生成中文摘要与标签...")

    # 修复：将【强制中文】指令放在最后，并强化语气
    prompt = f"""
    你是一个专业的中文知识库整理助手。

    【待处理文本片段】:
    {full_text[:12000]} ... (下略)

    【任务要求】
    1. **提取标签**: 请分析文本，提取 3-5 个核心关键词作为标签。
       格式必须严格为: "Tags: #标签1 #标签2 #标签3"

    2. **生成摘要**: 无论原文是德语、英语还是其他语言，**必须使用简体中文**进行总结。
       不要使用英文！不要使用德文！

    3. **输出格式**: 请直接输出以下 Markdown 内容：

    Tags: (这里填你提取的标签)

    ## 🧐 智能摘要
    (这里写 300 字左右的**中文**摘要)

    ## 💡 核心观点
    - (**中文**观点1)
    - (**中文**观点2)
    - (**中文**观点3)
    - (**中文**观点4)

    【再次强调】：所有输出内容必须是**中文**！
    """

    response = ollama.chat(model=OLLAMA_MODEL, messages=[{'role': 'user', 'content': prompt}])
    return response['message']['content']

def translate_full_text_safely(full_text):
    print(f"\n🌍 [4/4] 正在全文翻译...")
    chunk = full_text[:INTERNAL_PROCESS_CHUNK]
    prompt = f"请将以下文本翻译成流畅的简体中文，直接输出译文：\n\n{chunk}"
    try:
        res = ollama.chat(model=OLLAMA_MODEL, messages=[{'role': 'user', 'content': prompt}])
        return res['message']['content'] + "\n\n(......篇幅较长，仅展示开头部分翻译......)"
    except:
        return "(翻译服务不可用)"

def move_audio_to_vault(local_audio_file, target_name):
    """移动音频文件"""
    assets_dir = os.path.join(OBSIDIAN_VAULT_PATH, ASSETS_FOLDER_NAME)
    os.makedirs(assets_dir, exist_ok=True)

    # 重命名为：视频标题.m4a
    final_name = f"{target_name}.m4a"
    dest_path = os.path.join(assets_dir, final_name)

    # 如果目标文件已存在，先删除旧的，防止报错
    if os.path.exists(dest_path):
        os.remove(dest_path)

    shutil.move(local_audio_file, dest_path)
    return final_name

def extract_tags(llm_output):
    """提取 Tags 并清理正文"""
    tags = ["待整理"]

    # 匹配 Tags: #tag1 #tag2...
    match = re.search(r"Tags:\s*(.*)", llm_output, re.IGNORECASE)
    if match:
        tag_line = match.group(1)
        # 提取所有带 # 的词，或者被空格分隔的词
        extracted = re.findall(r"#?(\w[\w\d\-_]+)", tag_line)
        if extracted:
            # 过滤掉 "Tags" 本身如果被误吸入
            tags = [t for t in extracted if t.lower() != "tags"]

    # 从正文中删除 Tags 这一行
    cleaned_body = re.sub(r"Tags:.*(\n|$)", "", llm_output, flags=re.IGNORECASE).strip()
    return tags, cleaned_body

def save_to_obsidian(url, title, llm_output, original_text, translated_text, lang_code, audio_filename):
    print("\n💾 正在写入 Obsidian...")

    md_filename = f"{OBSIDIAN_VAULT_PATH}/{title}.md"
    os.makedirs(os.path.dirname(md_filename), exist_ok=True)

    tags_list, cleaned_body = extract_tags(llm_output)
    # 组装 YAML 格式的 tags
    tags_yaml = "\n".join([f"  - {t}" for t in tags_list])

    translation_section = ""
    if lang_code != 'zh':
        translation_section = f"## 🌍 全文翻译\n> 💡 AI 翻译预览\n\n{translated_text}\n\n---\n"

    content = f"""---
created: {datetime.datetime.now().strftime("%Y-%m-%d %H:%M")}
source: "{url}"
type: auto_clipper
language: {lang_code}
tags:
{tags_yaml}
---

# {title}

{cleaned_body}

---
## 🎧 音频回放
![[{ASSETS_FOLDER_NAME}/{audio_filename}]]

---
{translation_section}
## 📝 原始内容 (Original)

{original_text}

---
*Generated by Auto-Clipper V3.1*
"""
    with open(md_filename, "w", encoding="utf-8") as f:
        f.write(content)
    print(f"✅ 完成！笔记已创建: {md_filename}")

def main():
    print("=== Auto-Clipper V3.1 (修复版) ===")
    url = input("\n请输入链接: ").strip()
    if not url: return

    # 1. 获取标题 (修复了 Cookie 问题)
    title = get_video_info(url)

    # 2. 查重 (根据标题查重)
    if check_is_duplicate(title):
        return

    # 3. 下载音频 (使用临时名)
    temp_id = datetime.datetime.now().strftime("%H%M%S")
    temp_name = f"temp_{temp_id}"
    downloaded_file = download_audio(url, temp_name)

    if not downloaded_file:
        print("❌ 音频下载失败，流程终止。")
        return

    try:
        # 4. 转录
        whisper_result = transcribe_audio(downloaded_file)
        full_text = whisper_result['text']
        lang = whisper_result.get('language', 'en')
        print(f"   -> 语言: {lang}")

        # 5. 生成 (强化中文 Prompt)
        llm_output = generate_intelligence(full_text)

        # 6. 翻译
        translated = ""
        if lang != 'zh':
            translated = translate_full_text_safely(full_text)

        # 7. 归档音频
        final_audio_name = move_audio_to_vault(downloaded_file, title)

        # 8. 保存
        save_to_obsidian(url, title, llm_output, full_text, translated, lang, final_audio_name)

    except Exception as e:
        print(f"❌ 运行出错: {e}")
    finally:
        # 清理可能残留的临时文件
        if os.path.exists(f"{temp_name}.m4a"):
            os.remove(f"{temp_name}.m4a")

if __name__ == "__main__":
    main()
