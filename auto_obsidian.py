import os
import subprocess
import datetime
import mlx_whisper
import ollama
import shutil
import re
import json

# ==================== 用户配置区域 ====================
# Obsidian 收件箱路径
OBSIDIAN_VAULT_PATH = "/Users/xnc/vault/Inbox"

# 音频附件存放文件夹
ASSETS_FOLDER_NAME = "assets"

# 🤖 模型配置
# 分析模型 (建议用逻辑强的，如 qwen2.5:14b 或 qwen3:8b)
MODEL_ANALYSIS = "qwen3:8b"

# 翻译模型 (建议用速度快的)
MODEL_TRANSLATION = "qwen3:8b"

# Whisper 模型
WHISPER_MODEL = "mlx-community/whisper-large-v3-turbo"
# ====================================================

def sanitize_filename(name):
    """
    【V6.1 文件名清洗】
    1. 移除 Markdown/Obsidian 敏感符 (#, ^, [, ])
    2. 移除系统非法字符
    3. 空格和标点转下划线 (Snake Case 风格)
    4. 长度截断
    """
    # 1. 移除 Obsidian 链接破坏者
    name = re.sub(r'[#\^\[\]]', "", name)

    # 2. 移除系统非法字符
    name = re.sub(r'[\\/*?:"<>|]', "", name)

    # 3. 替换空格和标点为下划线
    chars_to_replace = [" ", "　", "，", ",", "。", "：", ":", "“", "”", "‘", "’"]
    for char in chars_to_replace:
        name = name.replace(char, "_")

    # 4. 移除不可见字符
    name = "".join(ch for ch in name if ch.isprintable())

    # 5. 合并连续的下划线
    name = re.sub(r'_{2,}', '_', name)

    # 6. 去除首尾可能残留的下划线
    name = name.strip("_")

    # 7. 长度截断
    if len(name) > 80:
        name = name[:80]

    return name

def get_video_info(url):
    print("[信息] 正在获取视频标题...")
    try:
        cmd = [
            "yt-dlp", "--get-title",
            "--cookies-from-browser", "chrome",
            "--no-warnings", url
        ]
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        title = result.stdout.strip()
        if not title: raise ValueError("标题为空")

        safe_title = sanitize_filename(title)
        print(f"[成功] 标题获取成功: {safe_title}")
        return safe_title
    except Exception as e:
        print(f"[警告] 标题获取失败: {e}")
        return f"素材_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}"

def check_is_duplicate(target_filename):
    file_path = os.path.join(OBSIDIAN_VAULT_PATH, f"{target_filename}.md")
    if os.path.exists(file_path):
        print(f"[跳过] 笔记 [{target_filename}] 已存在。")
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
        print(f"\n[错误] 下载出错: {e.stderr.decode()}")
        return None

def clean_hallucinations(text):
    """清洗 Whisper 幻觉 (重复词)"""
    if not text: return text
    # 清洗单词重复
    text = re.sub(r'\b(\w+)(?:\s+\1\b){4,}', r'\1', text, flags=re.IGNORECASE)
    # 清洗短语重复
    text = re.sub(r'\b(.{2,20})(?:\s+\1\b){4,}', r'\1', text, flags=re.IGNORECASE)
    # 清洗 Thank you
    text = re.sub(r'(Thank you\.?(\s*)){2,}', 'Thank you.', text, flags=re.IGNORECASE)
    return text.strip()

def transcribe_audio(audio_file):
    print("\n[2/4] 正在转录 (MLX加速中)...")
    result = mlx_whisper.transcribe(audio_file, path_or_hf_repo=WHISPER_MODEL, verbose=True)

    # 立即清洗
    raw_text = result['text']
    cleaned_text = clean_hallucinations(raw_text)
    if len(raw_text) - len(cleaned_text) > 50:
        print(f"   [清理] 已自动清除幻觉文本 ({len(raw_text) - len(cleaned_text)} 字符)")

    result['text'] = cleaned_text
    return result

def format_original_text(whisper_result):
    segments = whisper_result.get('segments', [])
    if not segments: return whisper_result['text']
    return "\n".join([seg.get('text', '').strip() for seg in segments])

def generate_intelligence_analysis(full_text):
    """32k上下文 + JSON强制模式"""
    print(f"\n[3/4] 正在进行深度分析 (模型: {MODEL_ANALYSIS})...")

    # 变量初始化
    full_response_content = ""
    current_context_size = 32000 # 32k 上下文

    prompt = f"""
    [Role]
    Professional Strategic Analyst.

    [Task]
    Analyze the provided text.
    Output the result in strict JSON format.
    Language: Simplified Chinese (简体中文).

    [JSON Structure]
    {{
        "tags": ["tag1", "tag2"],
        "summary": "Full summary (300 words+)",
        "key_points": ["point1", "point2"],
        "assessment": {{
            "authenticity": "Evaluation",
            "effectiveness": "Evaluation",
            "timeliness": "Evaluation",
            "alternatives": "Evaluation"
        }}
    }}

    [Input Text]
    {full_text[:30000]}
    """

    try:
        print(f"   -> 正在思考中 (Context: {current_context_size})...")
        stream = ollama.chat(
            model=MODEL_ANALYSIS,
            messages=[{'role': 'user', 'content': prompt}],
            format='json',
            options={
                "temperature": 0.1,
                "num_ctx": current_context_size,
                "num_predict": 2500,
                "repeat_penalty": 1.1
            },
            stream=True
        )

        for chunk in stream:
            part = chunk['message']['content']
            print(part, end="", flush=True)
            full_response_content += part

        print("\n\n   -> 生成完毕，正在解析...")
        return json.loads(full_response_content)

    except Exception as e:
        print(f"\n[错误] 分析错误: {e}")
        # 尝试暴力提取
        try:
            match = re.search(r"(\{.*\})", full_response_content, re.DOTALL)
            if match: return json.loads(match.group(1))
        except: pass

        return {
            "tags": ["分析失败"],
            "summary": f"分析失败: {str(e)}",
            "key_points": [],
            "assessment": {}
        }

def translate_full_text_loop(full_text):
    """分块翻译 + 防卡死"""
    CHUNK_SIZE = 1500
    print(f"\n[4/4] 正在全文翻译 (模型: {MODEL_TRANSLATION})...")

    chunks = [full_text[i:i+CHUNK_SIZE] for i in range(0, len(full_text), CHUNK_SIZE)]
    translated_parts = []

    for i, chunk in enumerate(chunks):
        print(f"   -> 翻译进度: {i+1}/{len(chunks)} ... ", end="", flush=True)
        prompt = f"Translate to Simplified Chinese. Keep format.\n\n{chunk}"
        try:
            res = ollama.chat(
                model=MODEL_TRANSLATION,
                messages=[{'role': 'user', 'content': prompt}],
                options={"temperature": 0.3, "num_ctx": 4096, "num_predict": 2048}
            )
            translated_parts.append(res['message']['content'])
            print("[OK]")
        except:
            print("[Err]")
            translated_parts.append("\n[翻译失败]\n")

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
    print("\n[保存] 正在写入 Obsidian...")
    md_filename = f"{OBSIDIAN_VAULT_PATH}/{title}.md"
    os.makedirs(os.path.dirname(md_filename), exist_ok=True)

    tags = data.get("tags", [])
    tags_clean = [t.replace("#", "").strip() for t in tags]
    tags_yaml = "\n".join([f"  - {t}" for t in tags_clean])

    points_md = "\n".join([f"- {p}" for p in data.get("key_points", [])])
    assess = data.get("assessment", {})

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

### 智能评估
| 维度 | 评估内容 |
| :--- | :--- |
| **真实性** | {assess.get('authenticity', 'N/A')} |
| **有效性** | {assess.get('effectiveness', 'N/A')} |
| **实时性** | {assess.get('timeliness', 'N/A')} |
| **替代策略** | {assess.get('alternatives', 'N/A')} |

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
    print(f"[成功] 完成！笔记已创建: {md_filename}")

def main():
    print("=== Auto-Clipper V6.1 (纯净版) ===")
    url = input("\n请输入链接: ").strip()
    if not url: return

    # 1. 获取标题
    title = get_video_info(url)
    if check_is_duplicate(title): return

    # 2. 下载音频
    temp_name = f"temp_{datetime.datetime.now().strftime('%H%M%S')}"
    dl_file = download_audio(url, temp_name)
    if not dl_file: return

    try:
        # 3. 转录 + 清洗幻觉
        whisper_res = transcribe_audio(dl_file)
        full_text = whisper_res['text']
        lang = whisper_res.get('language', 'en')

        formatted_orig = format_original_text(whisper_res)

        # 4. 智能分析
        analysis_data = generate_intelligence_analysis(full_text)

        # 5. 全文翻译
        translated = ""
        if lang != 'zh':
            translated = translate_full_text_loop(full_text)

        # 6. 归档音频
        final_audio = move_audio_to_vault(dl_file, title)

        # 7. 保存
        save_to_obsidian(url, title, analysis_data, formatted_orig, translated, lang, final_audio)

    except Exception as e:
        print(f"[错误] 运行出错: {e}")
    finally:
        if os.path.exists(dl_file): os.remove(dl_file)

if __name__ == "__main__":
    main()
