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

# 音频附件存放文件夹
ASSETS_FOLDER_NAME = "assets"

# 模型配置 (双模型架构)
# 1. 分析模型：负责摘要、观点、深度评估 (建议用更聪明的模型，如 qwen2.5:14b, gemini-3-flash)
MODEL_ANALYSIS = "qwen3:8b"

# 2. 翻译模型：负责全文翻译 (建议用速度快的模型，如 qwen2.5:7b, llama3)
MODEL_TRANSLATION = "qwen3:8b"

# Whisper 模型
WHISPER_MODEL = "mlx-community/whisper-large-v3-turbo"

# 翻译分块大小
INTERNAL_PROCESS_CHUNK = 1500
# ====================================================

def sanitize_filename(name):
    name = re.sub(r'[\\/*?:"<>|]', "", name)
    name = name.replace("\n", "").replace("\r", "").strip()
    return name[:80]

def get_video_info(url):
    print("正在获取视频标题...")
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
        print(f"标题获取失败: {e}")
        return f"素材_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}"

def check_is_duplicate(target_filename):
    file_path = os.path.join(OBSIDIAN_VAULT_PATH, f"{target_filename}.md")
    if os.path.exists(file_path):
        print(f"跳过: 笔记已存在。")
        return True
    return False

def download_audio(url, temp_filename):
    print(f"[1/4] 正在下载音频...")
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
        print(f"\n下载出错: {e.stderr.decode()}")
        return None

def transcribe_audio(audio_file):
    print("\n[2/4] 正在转录 (MLX加速中)...")
    return mlx_whisper.transcribe(audio_file, path_or_hf_repo=WHISPER_MODEL, verbose=True)

def format_original_text(whisper_result):
    segments = whisper_result.get('segments', [])
    if not segments: return whisper_result['text']
    return "\n".join([seg.get('text', '').strip() for seg in segments])

def generate_intelligence_analysis(full_text):
    """
    【修复版】开启流式输出，实时显示进度
    """
    print(f"\n[3/4] 正在进行深度分析 (模型: {MODEL_ANALYSIS})...")

    prompt = f"""
    你是一个专业的战略分析师。请阅读以下文本，并输出严格的 JSON 数据。

    【任务要求】
    1. 必须输出 JSON 格式。
    2. 必须使用**简体中文**。
    3. JSON 需包含以下字段：
       - "tags": [标签列表]
       - "summary": "300字左右的摘要"
       - "key_points": ["核心观点1", "核心观点2"...]
       - "assessment": {{
            "authenticity": "真实性评估",
            "effectiveness": "有效性评估",
            "timeliness": "实时性评估",
            "alternatives": "替代方案"
         }}

    【待分析文本】:
    {full_text}
    """

    try:
        print("   -> 正在思考中 (请稍候，即将开始生成)...")

        # 核心修改 1: 添加 stream=True
        stream = ollama.chat(
            model=MODEL_ANALYSIS,
            messages=[{'role': 'user', 'content': prompt}],
            stream=True
        )

        full_response_content = ""

        # 核心修改 2: 循环打印每个片段
        for chunk in stream:
            part = chunk['message']['content']
            # end="" 防止自动换行，flush=True 强制立即刷新缓冲区显示
            print(part, end="", flush=True)
            full_response_content += part

        print("\n\n   -> 生成完毕，正在解析 JSON...")

        # 后续逻辑不变：解析 JSON
        match = re.search(r"\{.*\}", full_response_content, re.DOTALL)
        if match:
            return json.loads(match.group(0))
        else:
            raise ValueError("未找到 JSON 格式的大括号")

    except Exception as e:
        print(f"\n分析失败，回退模式: {e}")
        return {
            "tags": ["待整理"],
            "summary": "分析中断或失败，请检查日志。",
            "key_points": [],
            "assessment": {}
        }

def translate_full_text_loop(full_text):
    """
    使用【翻译模型】进行全文翻译
    """
    print(f"\n[4/4] 正在全文翻译 (使用模型: {MODEL_TRANSLATION})...")

    chunks = [full_text[i:i+INTERNAL_PROCESS_CHUNK] for i in range(0, len(full_text), INTERNAL_PROCESS_CHUNK)]
    translated_parts = []

    for i, chunk in enumerate(chunks):
        print(f"   -> 翻译进度: {i+1}/{len(chunks)}")
        prompt = f"将以下文本翻译成流畅的简体中文，保留段落结构，不要解释：\n\n{chunk}"
        try:
            # 使用翻译模型
            res = ollama.chat(model=MODEL_TRANSLATION, messages=[{'role': 'user', 'content': prompt}])
            translated_parts.append(res['message']['content'])
        except:
            translated_parts.append("\n(翻译失败)\n")

    return "\n\n".join(translated_parts)

def move_audio_to_vault(local_audio_file, target_name):
    assets_dir = os.path.join(OBSIDIAN_VAULT_PATH, ASSETS_FOLDER_NAME)
    os.makedirs(assets_dir, exist_ok=True)
    final_name = f"{target_name}.m4a"
    dest_path = os.path.join(assets_dir, final_name)
    if os.path.exists(dest_path): os.remove(dest_path)
    shutil.move(local_audio_file, dest_path)
    return final_name

def save_to_obsidian(url, title, data, original, translated, lang, audio_name):
    print("\n正在写入 Obsidian...")
    md_filename = f"{OBSIDIAN_VAULT_PATH}/{title}.md"
    os.makedirs(os.path.dirname(md_filename), exist_ok=True)

    # 1. 标签
    tags_yaml = "\n".join([f"  - {t}" for t in data.get("tags", [])])

    # 2. 观点列表
    points_md = "\n".join([f"- {p}" for p in data.get("key_points", [])])

    # 3. 智能评估板块
    assess = data.get("assessment", {})
    assessment_md = f"""
### 智能评估
| 维度 | 评估内容 |
| :--- | :--- |
| **真实性** | {assess.get('authenticity', 'N/A')} |
| **有效性** | {assess.get('effectiveness', 'N/A')} |
| **实时性** | {assess.get('timeliness', 'N/A')} |
| **替代策略** | {assess.get('alternatives', 'N/A')} |
"""

    # 4. 翻译板块
    trans_section = f"## 全文翻译\n\n{translated}\n\n---\n" if lang != 'zh' else ""

    content = f"""---
created: {datetime.datetime.now().strftime("%Y-%m-%d %H:%M")}
source: "{url}"
type: auto_clipper
language: {lang}
tags:
{tags_yaml}
---

# {title}

## 智能摘要

{data.get("summary", "")}

### 核心观点

{points_md}

{assessment_md}

---

## 音频回放

![[{ASSETS_FOLDER_NAME}/{audio_name}]]

---
{trans_section}
## 原始内容

{original}
"""
    with open(md_filename, "w", encoding="utf-8") as f:
        f.write(content)
    print(f"完成！笔记已创建: {md_filename}")

def main():
    print("=== Auto-Clipper V5.0 (双模型智能版) ===")
    url = input("\n请输入链接: ").strip()
    if not url: return

    title = get_video_info(url)
    if check_is_duplicate(title): return

    temp_name = f"temp_{datetime.datetime.now().strftime('%H%M%S')}"
    dl_file = download_audio(url, temp_name)
    if not dl_file: return

    try:
        whisper_res = transcribe_audio(dl_file)
        full_text = whisper_res['text']
        lang = whisper_res.get('language', 'en')

        # 核心逻辑
        formatted_orig = format_original_text(whisper_res)

        # 步骤 1: 使用【分析模型】做深度思考
        analysis_data = generate_intelligence_analysis(full_text)

        # 步骤 2: 使用【翻译模型】做长文本翻译 (如果需要)
        translated = ""
        if lang != 'zh':
            translated = translate_full_text_loop(full_text)

        audio_final = move_audio_to_vault(dl_file, title)
        save_to_obsidian(url, title, analysis_data, formatted_orig, translated, lang, audio_final)

    except Exception as e:
        print(f"错误: {e}")
    finally:
        if os.path.exists(dl_file): os.remove(dl_file)

if __name__ == "__main__":
    main()
