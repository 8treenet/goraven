package seed

const SystemSkillGoRavenDocParse = `---
name: goraven-doc-parse
description: "读取或转换文档（PDF/DOCX/PPTX/XLSX/HTML/CSV/LaTeX）为文本。不支持图片和扫描件。"
---

# 文档读取与转换

通过 shell 调用 Python 脚本处理文档，提取其中的文字和结构信息。

### 支持格式

- PDF（仅含文字层的，扫描件/图片 PDF 无法处理）
- DOCX（Word）
- PPTX（PowerPoint）
- XLSX（Excel）
- HTML
- CSV
- LaTeX
- ASCIIDOC

### 不支持

- 图片（PNG/JPEG/TIFF 等）——无 OCR 能力
- 扫描件 PDF（无文字层）

纯文本文件（.txt、.json、.yaml、.xml、.md 等）无需使用本技能，直接用文件读取工具即可。

## 脚本位置

- 读取脚本：/goraven/scripts/docling/read.py
- 转换脚本：/goraven/scripts/docling/convert.py

Python 环境已就绪（venv 在 PATH 中），直接用 python 命令调用。

## 模式一：读取（read）

将文档内容提取到终端输出，适合即时阅读和理解文档内容。

### 命令格式

python /goraven/scripts/docling/read.py --input <文件路径> [--format markdown|text] [--max-chars <上限>] [--pages <页码>]

### 参数说明

- --input（必填）：文档文件的绝对路径
- --format：输出格式，markdown（默认，保留标题/表格结构）或 text（纯文本）
- --max-chars：最大输出字符数，超出部分截断。大文档建议设置（如 50000）避免输出过长
- --pages：页码过滤，如 1-5 或 1,3,5-8，仅读取指定页

### 使用示例

# 读取 PDF 全文
python /goraven/scripts/docling/read.py --input /path/to/report.pdf

# 只读前 3 页，限制输出长度
python /goraven/scripts/docling/read.py --input /path/to/report.pdf --pages 1-3 --max-chars 30000

# 以纯文本格式读取 Word 文档
python /goraven/scripts/docling/read.py --input /path/to/doc.docx --format text

### 适用场景

- 用户要求阅读、总结、分析某个文档
- 需要从文档中提取信息回答问题
- 预览文档内容决定后续操作

## 模式二：转换（convert）

将文档转换为 Markdown 文件保存到磁盘，适合持久化存储或后续加工。

### 命令格式

python /goraven/scripts/docling/convert.py --input <源文件路径> --output <输出.md路径>

### 参数说明

- --input（必填）：源文档的绝对路径
- --output（必填）：输出 Markdown 文件的绝对路径（含 .md 后缀）

### 使用示例

# 将 PDF 转为 Markdown 保存
python /goraven/scripts/docling/convert.py --input /path/to/report.pdf --output /path/to/output/report.md

### 适用场景

- 用户要求将文档转为 Markdown 格式保存
- 需要持久化文档的结构化文本版本
- 为后续编辑、引用或知识库入库做准备

## 规则

- 文件路径必须使用绝对路径
- 输出目录不存在时脚本会自动创建
- 首次调用可能较慢（模型加载），后续调用会快很多
- 如果脚本报错 ModuleNotFoundError，说明 docling 未安装，执行：uv pip install --no-cache -r /goraven/scripts/docling/requirements.txt
- 读取大文档时优先使用 --max-chars 或 --pages 控制输出量，避免占满上下文
- 转换后的 .md 文件存放位置由用户指定或放在源文件同目录下
`

const SystemSkillGoRavenDocParseEn = `---
name: goraven-doc-parse
description: "Read or convert documents (PDF/DOCX/PPTX/XLSX/HTML/CSV/LaTeX) to text. Images and scanned files not supported."
---

# Document Reading & Conversion

Process documents via Python scripts through shell, extracting text and structural information.

### Supported Formats

- PDF (text layer only — scanned/image PDFs cannot be processed)
- DOCX (Word)
- PPTX (PowerPoint)
- XLSX (Excel)
- HTML
- CSV
- LaTeX
- ASCIIDOC

### Not Supported

- Images (PNG/JPEG/TIFF, etc.) — no OCR capability
- Scanned PDFs (no text layer)

Plain text files (.txt, .json, .yaml, .xml, .md, etc.) do not need this skill — use the file read tool directly.

## Script Locations

- Read script: /goraven/scripts/docling/read.py
- Convert script: /goraven/scripts/docling/convert.py

The Python environment is ready (venv is in PATH). Call directly with the python command.

## Mode 1: Read

Extract document content to terminal output. Best for immediate reading and comprehension.

### Command Format

python /goraven/scripts/docling/read.py --input <file_path> [--format markdown|text] [--max-chars <limit>] [--pages <range>]

### Parameters

- --input (required): absolute path to the document
- --format: output format — markdown (default, preserves headings/tables) or text (plain text)
- --max-chars: maximum output characters; content beyond this is truncated. Recommended for large documents (e.g. 50000)
- --pages: page filter, e.g. 1-5 or 1,3,5-8 — only read specified pages

### Examples

# Read entire PDF
python /goraven/scripts/docling/read.py --input /path/to/report.pdf

# Read first 3 pages with output limit
python /goraven/scripts/docling/read.py --input /path/to/report.pdf --pages 1-3 --max-chars 30000

# Read Word document as plain text
python /goraven/scripts/docling/read.py --input /path/to/doc.docx --format text

### Use Cases

- User asks to read, summarize, or analyze a document
- Need to extract information from a document to answer questions
- Preview document content before deciding next steps

## Mode 2: Convert

Convert a document to a Markdown file saved on disk. Best for persistent storage or further processing.

### Command Format

python /goraven/scripts/docling/convert.py --input <source_path> --output <output.md_path>

### Parameters

- --input (required): absolute path to the source document
- --output (required): absolute path for the output Markdown file (with .md extension)

### Examples

# Convert PDF to Markdown
python /goraven/scripts/docling/convert.py --input /path/to/report.pdf --output /path/to/output/report.md

### Use Cases

- User asks to convert a document to Markdown format
- Need a persistent structured text version of a document
- Prepare content for editing, referencing, or knowledge base ingestion

## Rules

- Always use absolute file paths
- Output directories are created automatically if missing
- First invocation may be slow (model loading); subsequent calls are much faster
- If the script reports ModuleNotFoundError, docling is not installed. Run: uv pip install --no-cache -r /goraven/scripts/docling/requirements.txt
- For large documents, prefer --max-chars or --pages to control output volume and avoid filling the context
- Place converted .md files where the user specifies, or in the same directory as the source file
`
