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

def clean_hallucinations(text):
    """
    🧹 清洗 Whisper 的复读机幻觉 (例如: feel feel feel...)
    原理：使用正则匹配重复出现 5 次以上的单词或短语
    """
    if not text: return text

    # 1. 清洗单词重复 (例如: feel feel feel)
    # \b(\w+)(?:\s+\1\b){4,} -> 匹配一个单词，后面跟着 4 次以上相同的单词
    text = re.sub(r'\b(\w+)(?:\s+\1\b){4,}', r'\1', text, flags=re.IGNORECASE)

    # 2. 清洗短语重复 (例如: and easiest and easiest)
    # 匹配 2-10 个字符长度的短语，重复 4 次以上
    text = re.sub(r'\b(.{2,20})(?:\s+\1\b){4,}', r'\1', text, flags=re.IGNORECASE)

    # 3. 清洗常见的 Whisper 幻觉词 (如果你发现还有其他的，可以加在这里)
    # 有些版本的 Whisper 会疯狂输出 "Thank you." 或 "Bye."
    # 这里是一个保守的清洗，只去掉末尾连续的 Thank you
    text = re.sub(r'(Thank you\.?(\s*)){2,}', 'Thank you.', text, flags=re.IGNORECASE)

    return text.strip()

def transcribe_audio(audio_file):
    """Whisper 转录 + 自动清洗"""
    print("\n🎙️ [2/4] 正在转录 (MLX加速中)...")
    result = mlx_whisper.transcribe(
        audio_file,
        path_or_hf_repo=WHISPER_MODEL,
        verbose=True
    )

    # === 新增：立即清洗幻觉 ===
    raw_text = result['text']
    cleaned_text = clean_hallucinations(raw_text)

    # 如果清洗掉了大量字符，打印提示
    if len(raw_text) - len(cleaned_text) > 50:
        print(f"   🧹 已自动清除 Whisper 幻觉文本 ({len(raw_text) - len(cleaned_text)} 字符)")

    # 更新 result 中的 text
    result['text'] = cleaned_text
    return result

def format_original_text(whisper_result):
    segments = whisper_result.get('segments', [])
    if not segments: return whisper_result['text']
    return "\n".join([seg.get('text', '').strip() for seg in segments])

def generate_intelligence_analysis(full_text):
    """
    【V5.4 稳健调试版】修复变量作用域 + 32k上下文 + 暴力JSON清洗
    """
    print(f"\n🧠 [3/4] 正在进行深度分析 (模型: {MODEL_ANALYSIS})...")

    # === 变量初始化 (修复 Pyright 报错) ===
    # 必须在 try 之前定义，否则如果 try 第一行就挂了，except 里打印会再次报错
    full_response_content = ""

    # === 1. 动态计算需要的上下文 ===
    # 英文单词数 * 1.5 ≈ Token数。你的文本约 3800 词 ≈ 5700 Tokens。
    # 我们设置 32000 (32k) 绰绰有余，能容纳 2 小时的视频字幕。
    current_context_size = 32000

    # === 2. 简化的 Prompt ===
    # 对于 8B 模型，Prompt 越像代码越好。不要用太复杂的自然语言。
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
        "summary": "Full summary of the content (300 words+)",
        "key_points": ["point1", "point2", "point3"],
        "assessment": {{
            "authenticity": "Evaluation of factuality",
            "effectiveness": "Evaluation of logic/method",
            "timeliness": "Is the info up-to-date?",
            "alternatives": "Alternative viewpoints or solutions"
        }}
    }}

    [Input Text]
    {full_text[:30000]}
    """

    try:
        print(f"   -> 正在思考中 (Context: {current_context_size} tokens)...")

        stream = ollama.chat(
            model=MODEL_ANALYSIS,
            messages=[{'role': 'user', 'content': prompt}],
            # 强制 JSON 模式
            format='json',
            options={
                "temperature": 0.1,      # 极度理性
                "num_ctx": current_context_size, # 【关键】扩大内存，防止截断
                "num_predict": 2500,     # 允许输出较长的回答
                "repeat_penalty": 1.1    # 防止复读机
            },
            stream=True
        )

        for chunk in stream:
            part = chunk['message']['content']
            print(part, end="", flush=True)
            full_response_content += part

        print("\n\n   -> 生成完毕，正在解析...")

        # === 3. 暴力清洗与解析 ===
        # 有时候模型还是会输出 ```json ... ``` 哪怕我们开了 format='json'

        # 步骤 A: 尝试直接解析
        try:
            return json.loads(full_response_content)
        except json.JSONDecodeError:
            # 步骤 B: 如果失败，使用正则提取最外层大括号
            print("   -> 标准解析失败，尝试暴力提取...")
            match = re.search(r"(\{.*\})", full_response_content, re.DOTALL)
            if match:
                json_str = match.group(1)
                # 再次清洗：有时 JSON 里的换行符会导致错误
                # 这里做一个简单的清理（视情况而定）
                return json.loads(json_str)
            else:
                raise ValueError("未找到任何 {} 结构")

    except Exception as e:
        print(f"\n❌ 分析发生错误: {e}")

        # === 调试信息的关键修复 ===
        print("\n🔎 === [调试] 模型原始输出 (Raw Output) ===")
        print("↓" * 30)
        # 这里现在绝对安全了，因为 full_response_content 在最上面定义了
        print(full_response_content if full_response_content else "(无内容/连接超时)")
        print("↑" * 30)
        print("💡 请截图以上信息以便排查问题。\n")

        # 返回兜底数据，确保后续流程不中断
        return {
            "tags": ["分析失败"],
            "summary": f"智能分析未能完成。错误信息: {str(e)}",
            "key_points": [],
            "assessment": {
                "authenticity": "N/A", "effectiveness": "N/A", "timeliness": "N/A", "alternatives": "N/A"
            }
        }

def translate_full_text_loop(full_text):
    """
    【V5.5 修复版】增加防卡死机制 (Time-out protection)
    """
    # 1. 动态调整分块大小 (建议稍微小一点，1500字符一段比较稳)
    CHUNK_SIZE = 1500

    print(f"\n🌍 [4/4] 正在全文翻译 (使用模型: {MODEL_TRANSLATION})...")
    print(f"   -> 总字符数: {len(full_text)} | 分块大小: {CHUNK_SIZE}")

    chunks = [full_text[i:i+CHUNK_SIZE] for i in range(0, len(full_text), CHUNK_SIZE)]
    total_chunks = len(chunks)
    translated_parts = []

    for i, chunk in enumerate(chunks):
        # 打印当前进度，flush=True 确保立即显示
        print(f"   -> 翻译进度: {i+1}/{total_chunks} ... ", end="", flush=True)

        prompt = f"""
        Translate the following text into Simplified Chinese (简体中文).
        Keep the format. Do not add explanations.

        Text:
        {chunk}
        """

        try:
            # === 核心修复: 添加 options 限制 ===
            # 这能防止模型陷入无限循环
            response = ollama.chat(
                model=MODEL_TRANSLATION,
                messages=[{'role': 'user', 'content': prompt}],
                options={
                    "temperature": 0.3,    # 低温，保证翻译准确不胡编
                    "num_ctx": 4096,       # 翻译不需要太大上下文，4k足够
                    "num_predict": 2048,   # 【关键】强制止损！防止无限生成
                }
            )

            content = response['message']['content']
            translated_parts.append(content)
            print("✅") # 打印对勾表示这一块完成了

        except Exception as e:
            print(f"❌ (跳过: {str(e)})")
            translated_parts.append(f"\n[该片段翻译失败]\n")

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
