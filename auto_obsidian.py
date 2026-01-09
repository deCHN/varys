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

# 音频附件存放的子文件夹名称 (会在 Inbox 下自动创建)
ASSETS_FOLDER_NAME = "assets"

# 模型选择
OLLAMA_MODEL = "qwen2.5:7b"
WHISPER_MODEL = "mlx-community/whisper-large-v3-turbo"

# 内部处理分块大小
INTERNAL_PROCESS_CHUNK = 2000
# ====================================================

def sanitize_filename(name):
    """清理文件名中的非法字符，防止报错"""
    # 去掉 / \ : * ? " < > | 以及换行符
    name = re.sub(r'[\\/*?:"<>|\n]', "", name)
    # 限制长度，防止文件名过长
    return name[:100].strip()

def get_video_info(url):
    """获取视频标题"""
    print("🔍 正在获取视频信息...")
    try:
        # 使用 yt-dlp 获取标题 (--get-title)
        result = subprocess.run(
            ["yt-dlp", "--get-title", url],
            capture_output=True, text=True, check=True
        )
        title = result.stdout.strip()
        safe_title = sanitize_filename(title)
        print(f"📄 标题获取成功: {safe_title}")
        return safe_title
    except Exception as e:
        print(f"⚠️ 无法获取标题，将使用时间戳代替。错误: {e}")
        return f"素材_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}"

def check_is_duplicate(target_filename):
    """检查文件是否已存在"""
    file_path = os.path.join(OBSIDIAN_VAULT_PATH, f"{target_filename}.md")
    if os.path.exists(file_path):
        print(f"⚠️ 跳过: 文件 [{target_filename}.md] 已存在。")
        return True
    return False

def download_audio(url, temp_filename):
    """下载音频到临时文件"""
    print(f"⬇️ [1/4] 正在下载...")

    # 临时文件模板
    output_template = f"{temp_filename}.%(ext)s"

    cmd = [
        "yt-dlp",
        "-x", "--audio-format", "m4a",
        "--cookies-from-browser", "chrome",
        "-o", output_template,
        "--no-playlist",
        "--progress",
        "--newline",
        url
    ]

    try:
        process = subprocess.Popen(cmd, stdout=sys.stdout, stderr=sys.stderr)
        process.wait()

        if process.returncode != 0:
            raise subprocess.CalledProcessError(process.returncode, cmd)

        # 找到下载的具体文件（yt-dlp 可能会自动修正扩展名）
        for file in os.listdir("."):
            if file.startswith(temp_filename) and file.endswith(".m4a"):
                return file
        return None
    except Exception as e:
        print(f"\n❌ 下载出错: {e}")
        return None

def transcribe_audio(audio_file):
    """Whisper 转录"""
    print("\n🎙️ [2/4] 正在转录 (MLX加速中)...")
    result = mlx_whisper.transcribe(
        audio_file,
        path_or_hf_repo=WHISPER_MODEL,
        verbose=True
    )
    return result

def generate_intelligence(full_text):
    """生成摘要、观点和标签"""
    print("\n🧠 [3/4] 正在生成中文摘要与标签...")

    # 强制中文 Prompt
    prompt = f"""
    你是一个专业的知识库整理助手。
    【重要指令】：
    1. 无论原文是什么语言（英语、德语等），**必须全程使用中文（简体）**回答。
    2. 不要输出“元数据”章节，直接输出以下两部分内容。

    【任务 1：提取标签】
    请根据内容提取 3-5 个核心标签，以哈希号开头，用空格分隔。
    格式示例：Tags: #经济 #AI #科技
    (请务必包含 "Tags:" 前缀，以便我后续提取)

    【任务 2：生成内容】
    ## 🧐 智能摘要
    (300字左右的中文摘要)

    ## 💡 核心观点
    - (观点1)
    - (观点2)
    - (观点3)

    【原文片段】:
    {full_text[:15000]}
    """

    response = ollama.chat(model=OLLAMA_MODEL, messages=[{'role': 'user', 'content': prompt}])
    return response['message']['content']

def translate_full_text_safely(full_text):
    print(f"\n🌍 [4/4] 正在全文翻译...")

    # 简单处理：如果文本太长，只翻译前 2000 字作为示例，或者分块翻译
    # 这里为了演示稳定性，先翻译第一块，避免时间过长
    chunk = full_text[:INTERNAL_PROCESS_CHUNK]

    prompt = f"""
    请将以下内容翻译成流畅的中文，保留专有名词（如 ETF, AI 等）。
    直接输出译文，不要解释。

    {chunk}
    """
    try:
        res = ollama.chat(model=OLLAMA_MODEL, messages=[{'role': 'user', 'content': prompt}])
        translation = res['message']['content']
        if len(full_text) > INTERNAL_PROCESS_CHUNK:
            translation += "\n\n(......文章较长，仅展示前2000字翻译......)"
        return translation
    except:
        return "(翻译服务暂时不可用)"

def move_audio_to_vault(local_audio_file, target_filename):
    """将音频移动到 Obsidian 的 assets 文件夹"""
    # 1. 确保 assets 文件夹存在
    assets_dir = os.path.join(OBSIDIAN_VAULT_PATH, ASSETS_FOLDER_NAME)
    os.makedirs(assets_dir, exist_ok=True)

    # 2. 目标路径
    final_audio_name = f"{target_filename}.m4a"
    dest_path = os.path.join(assets_dir, final_audio_name)

    # 3. 移动文件
    shutil.move(local_audio_file, dest_path)
    print(f"📦 音频已归档至: {ASSETS_FOLDER_NAME}/{final_audio_name}")

    return final_audio_name

def extract_tags(llm_output):
    """从 LLM 输出中提取 Tags 行"""
    tags = ["#待整理"] # 默认标签

    # 寻找以 "Tags:" 开头的行
    match = re.search(r"Tags:\s*(.*)", llm_output, re.IGNORECASE)
    if match:
        tag_str = match.group(1)
        # 提取所有 #xxx
        found = re.findall(r"(#\w+)", tag_str)
        if found:
            tags = found

    # 从正文中移除 Tags 这一行，避免正文重复显示
    cleaned_output = re.sub(r"Tags:.*", "", llm_output, flags=re.IGNORECASE).strip()
    return tags, cleaned_output

def save_to_obsidian(url, title, llm_output, original_text, translated_text, lang_code, audio_filename):
    """保存 Markdown"""
    print("\n💾 正在写入 Obsidian...")

    md_filename = f"{OBSIDIAN_VAULT_PATH}/{title}.md"
    os.makedirs(os.path.dirname(md_filename), exist_ok=True)

    # 处理标签
    tags_list, cleaned_llm_body = extract_tags(llm_output)
    tags_yaml = "\n".join([f"  - {t.replace('#', '')}" for t in tags_list]) # YAML 格式不用 #

    # 组装翻译
    translation_section = ""
    if lang_code != 'zh':
        translation_section = f"""
## 🌍 全文翻译 (Translated)
> 💡 以下内容由 AI 自动翻译。

{translated_text}

---
"""

    # 组装播放器链接 (Obsidian 格式)
    # 格式: ![[filename.m4a]]
    audio_player = f"## 🎧 音频回放\n![[{ASSETS_FOLDER_NAME}/{audio_filename}]]"

    content = f"""---
created: {datetime.datetime.now().strftime("%Y-%m-%d %H:%M")}
source: "{url}"
type: auto_clipper
language: {lang_code}
tags:
{tags_yaml}
---

# {title}

{cleaned_llm_body}

---
{audio_player}

---
{translation_section}

## 📝 原始内容 (Original)

{original_text}

---
*Generated by Auto-Clipper V3*
"""

    with open(md_filename, "w", encoding="utf-8") as f:
        f.write(content)

    print(f"✅ 笔记已创建: {md_filename}")

def main():
    print("=== 个人知识库自动抓取工具 (V3 完美版) ===")

    url = input("\n请输入链接: ").strip()
    if not url: return

    # 1. 获取标题 (用于文件名)
    title = get_video_info(url)

    # 2. 查重
    if check_is_duplicate(title):
        return

    # 3. 下载音频 (使用临时文件名，避免特殊字符问题)
    temp_id = datetime.datetime.now().strftime("%H%M%S")
    temp_audio_name = f"temp_audio_{temp_id}"

    downloaded_file = download_audio(url, temp_audio_name)
    if not downloaded_file: return

    try:
        # 4. 转录
        whisper_result = transcribe_audio(downloaded_file)
        full_text = whisper_result['text']
        detected_lang = whisper_result.get('language', 'en')
        print(f"   -> 检测到语言: {detected_lang}")

        # 5. LLM 生成 (强制中文)
        llm_output = generate_intelligence(full_text)

        # 6. 翻译 (非中文时)
        translated_text = ""
        if detected_lang != 'zh':
            translated_text = translate_full_text_safely(full_text)

        # 7. 移动音频文件
        final_audio_name = move_audio_to_vault(downloaded_file, title)

        # 8. 保存笔记
        save_to_obsidian(url, title, llm_output, full_text, translated_text, detected_lang, final_audio_name)

    except Exception as e:
        print(f"❌ 发生未知错误: {e}")
        # 如果出错，清理临时文件
        if os.path.exists(downloaded_file):
            os.remove(downloaded_file)

if __name__ == "__main__":
    main()
