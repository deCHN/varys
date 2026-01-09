import os
import subprocess
import datetime
import mlx_whisper
import ollama
import sys
import shutil
import re
import json

# ==================== 用户配置区域 ====================
# Obsidian 收件箱路径
OBSIDIAN_VAULT_PATH = "/Users/xnc/vault/Inbox"

# 音频附件存放的子文件夹名称
ASSETS_FOLDER_NAME = "assets"

# 模型选择
OLLAMA_MODEL = "qwen2.5:7b"
WHISPER_MODEL = "mlx-community/whisper-large-v3-turbo"

# 翻译分块大小 (字符数)
INTERNAL_PROCESS_CHUNK = 1500
# ====================================================

def sanitize_filename(name):
    """清理文件名"""
    name = re.sub(r'[\\/*?:"<>|]', "", name)
    name = name.replace("\n", "").replace("\r", "").strip()
    return name[:80]

def get_video_info(url):
    """获取视频标题"""
    print("🔍 正在获取视频标题...")
    try:
        cmd = [
            "yt-dlp", "--get-title",
            "--cookies-from-browser", "chrome",
            "--no-warnings", url
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        title = result.stdout.strip()
        if not title: raise ValueError("标题为空")
        return sanitize_filename(title)
    except Exception as e:
        print(f"⚠️ 标题获取失败: {e}")
        return f"素材_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}"

def check_is_duplicate(target_filename):
    file_path = os.path.join(OBSIDIAN_VAULT_PATH, f"{target_filename}.md")
    if os.path.exists(file_path):
        print(f"⚠️ 跳过: 笔记已存在。")
        return True
    return False

def download_audio(url, temp_filename):
    """下载音频"""
    print(f"⬇️ [1/4] 正在下载音频...")
    output_template = f"{temp_filename}.%(ext)s"
    cmd = [
        "yt-dlp", "-x", "--audio-format", "m4a",
        "--cookies-from-browser", "chrome",
        "-o", output_template, "--no-playlist", "--newline", url
    ]
    try:
        subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
        for file in os.listdir("."):
            if file.startswith(temp_filename) and file.endswith(".m4a"):
                return file
        return None
    except subprocess.CalledProcessError as e:
        print(f"\n❌ 下载出错: {e.stderr.decode()}")
        return None

def transcribe_audio(audio_file):
    """Whisper 转录 (返回完整对象以获取 segments)"""
    print("\n🎙️ [2/4] 正在转录 (MLX加速中)...")
    result = mlx_whisper.transcribe(
        audio_file,
        path_or_hf_repo=WHISPER_MODEL,
        verbose=True
    )
    return result

def format_original_text(whisper_result):
    """
    将 Whisper 的原始文本进行分段处理。
    如果只有一个大段，尝试按 segments 加换行。
    """
    segments = whisper_result.get('segments', [])
    if not segments:
        return whisper_result['text']

    formatted_text = ""
    for seg in segments:
        text = seg.get('text', '').strip()
        # 每段话后面加换行，形成自然的阅读流
        formatted_text += f"{text}\n"

    return formatted_text

def generate_intelligence_json(full_text):
    """
    使用 JSON 模式生成结构化数据，彻底避免正文出现多余的元数据文本。
    """
    print("\n🧠 [3/4] 正在生成智能摘要 (JSON模式)...")

    prompt = f"""
    你是一个专业的中文知识库助手。请阅读以下文本（可能是外语），并提取信息。

    【任务要求】
    1. **必须输出标准的 JSON 格式**。
    2. **必须使用简体中文** 回答所有内容。
    3. JSON 需包含三个字段: "tags" (标签列表), "summary" (摘要文本), "key_points" (核心观点列表)。

    【JSON 格式示例】
    {{
        "tags": ["经济", "投资", "AI"],
        "summary": "这段视频主要讲述了...",
        "key_points": [
            "观点一...",
            "观点二..."
        ]
    }}

    【待处理文本】:
    {full_text[:12000]}
    """

    try:
        response = ollama.chat(model=OLLAMA_MODEL, messages=[{'role': 'user', 'content': prompt}])
        content = response['message']['content']

        # 提取 JSON 部分 (防止 LLM 在 JSON 外面说废话)
        match = re.search(r"\{.*\}", content, re.DOTALL)
        if match:
            json_str = match.group(0)
            data = json.loads(json_str)
            return data
        else:
            raise ValueError("未找到 JSON")

    except Exception as e:
        print(f"⚠️ JSON 解析失败，回退到普通文本模式: {e}")
        # 兜底返回
        return {
            "tags": ["待整理"],
            "summary": "自动摘要生成失败，请手动检查。",
            "key_points": []
        }

def translate_full_text_loop(full_text):
    """
    循环分块翻译全文，确保完整性。
    """
    print(f"\n🌍 [4/4] 正在全文翻译 ({len(full_text)} 字符)...")

    # 按长度切分
    chunks = [full_text[i:i+INTERNAL_PROCESS_CHUNK] for i in range(0, len(full_text), INTERNAL_PROCESS_CHUNK)]
    total_chunks = len(chunks)
    translated_parts = []

    for i, chunk in enumerate(chunks):
        print(f"   -> 翻译进度: {i+1}/{total_chunks}")
        prompt = f"""
        请将以下文本翻译成流畅的**简体中文**。
        要求：
        1. **保留段落结构**，不要合并成一大块。
        2. 遇到专业术语保留原文或括号标注。
        3. 直接输出译文，不要解释。

        【原文片段】：
        {chunk}
        """
        try:
            res = ollama.chat(model=OLLAMA_MODEL, messages=[{'role': 'user', 'content': prompt}])
            translated_parts.append(res['message']['content'])
        except Exception:
            translated_parts.append("\n(该片段翻译失败)\n")

    return "\n\n".join(translated_parts)

def move_audio_to_vault(local_audio_file, target_name):
    assets_dir = os.path.join(OBSIDIAN_VAULT_PATH, ASSETS_FOLDER_NAME)
    os.makedirs(assets_dir, exist_ok=True)
    final_name = f"{target_name}.m4a"
    dest_path = os.path.join(assets_dir, final_name)

    if os.path.exists(dest_path):
        os.remove(dest_path)
    shutil.move(local_audio_file, dest_path)
    return final_name

def save_to_obsidian(url, title, data_json, original_text, translated_text, lang_code, audio_filename):
    print("\n💾 正在写入 Obsidian...")
    md_filename = f"{OBSIDIAN_VAULT_PATH}/{title}.md"
    os.makedirs(os.path.dirname(md_filename), exist_ok=True)

    # 1. 构建 YAML
    tags = data_json.get("tags", ["待整理"])
    tags_yaml = "\n".join([f"  - {t}" for t in tags])

    # 2. 构建摘要和观点 (使用极简 Lucide 风格)
    summary = data_json.get("summary", "")
    key_points = data_json.get("key_points", [])

    points_md = ""
    for p in key_points:
        points_md += f"- {p}\n"

    # 3. 组装翻译部分
    translation_section = ""
    if lang_code != 'zh':
        translation_section = f"""
## 全文翻译

{translated_text}

---
"""

    # 4. 最终内容组装
    content = f"""---
created: {datetime.datetime.now().strftime("%Y-%m-%d %H:%M")}
source: "{url}"
type: auto_clipper
language: {lang_code}
tags:
{tags_yaml}
---

# {title}

## 智能摘要

{summary}

### 核心观点

{points_md}

---

## 音频回放

![[{ASSETS_FOLDER_NAME}/{audio_filename}]]

---
{translation_section}
## 原始内容

{original_text}
"""

    with open(md_filename, "w", encoding="utf-8") as f:
        f.write(content)
    print(f"✅ 完成！笔记已创建: {md_filename}")

def main():
    print("=== Auto-Clipper V4 (Ultimate) ===")
    url = input("\n请输入链接: ").strip()
    if not url: return

    title = get_video_info(url)
    if check_is_duplicate(title): return

    temp_name = f"temp_{datetime.datetime.now().strftime('%H%M%S')}"
    downloaded_file = download_audio(url, temp_name)
    if not downloaded_file: return

    try:
        # 转录
        whisper_result = transcribe_audio(downloaded_file)
        full_text = whisper_result['text']
        lang = whisper_result.get('language', 'en')
        print(f"   -> 语言: {lang}")

        # 格式化原始内容 (分段)
        formatted_original = format_original_text(whisper_result)

        # 生成智能信息 (JSON)
        intelligence_data = generate_intelligence_json(full_text)

        # 全文翻译 (循环分块)
        translated = ""
        if lang != 'zh':
            translated = translate_full_text_loop(full_text)

        # 归档音频
        final_audio_name = move_audio_to_vault(downloaded_file, title)

        # 保存
        save_to_obsidian(url, title, intelligence_data, formatted_original, translated, lang, final_audio_name)

    except Exception as e:
        print(f"❌ 运行出错: {e}")
    finally:
        if os.path.exists(downloaded_file): os.remove(downloaded_file)

if __name__ == "__main__":
    main()
