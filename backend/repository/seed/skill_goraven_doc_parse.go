package seed

const SystemSkillGoRavenDocParse = `---
name: goraven-doc-parse
description: "读取或转换文档（PDF/DOCX/PPTX/XLSX/HTML/CSV/LaTeX/ASCIIDOC）为文本。不支持图片和扫描件。聊天附件中的文档是原始文件，需通过本技能读取。"
---

# 文档读取与转换

使用 goraven_doc_parse 工具处理文档，支持两种模式：

- **read**：提取文档文本到工具返回值，适合即时阅读和理解内容
- **convert**：将文档转换为 Markdown 文件保存到磁盘，适合持久化或后续加工

### 支持格式

- PDF（仅含文字层的，扫描件/图片 PDF 无法处理）
- DOCX（Word）、PPTX（PowerPoint）、XLSX（Excel）
- HTML、CSV、LaTeX、ASCIIDOC

### 不支持

- 图片（PNG/JPEG/TIFF 等）
- 扫描件 PDF（无文字层）

纯文本文件（.txt、.json、.yaml、.xml、.md 等）无需本技能，直接用文件读取工具即可。

## read 模式：读取内容

参数：

- mode（必填）：固定为 "read"
- file_path（必填）：文档在沙盒内的绝对路径，聊天附件直接使用 <goraven-upload> 标签中给出的路径
- format：markdown（默认，保留标题/表格结构）或 text（纯文本）
- max_chars：最大输出字符数，超出截断，缺省 50000
- pages：页码过滤，如 "1-5" 或 "1,3,5-8"，仅 PDF 有效

参数示例：

{"mode": "read", "file_path": "/goraven/data/users/admin/temp/report.pdf"}

{"mode": "read", "file_path": "/goraven/data/users/admin/temp/report.pdf", "pages": "1-3", "max_chars": 30000}

{"mode": "read", "file_path": "/goraven/data/users/admin/temp/report.docx", "format": "text"}

返回 content 为文档文本；truncated 为 true 表示内容被截断，可用 pages 分页或增大 max_chars 后续读取。

## convert 模式：转换为 Markdown

参数：

- mode（必填）：固定为 "convert"
- file_path（必填）：源文档在沙盒内的绝对路径，聊天附件直接使用 <goraven-upload> 标签中给出的路径
- output_path（必填）：输出 Markdown 文件的绝对路径（.md 后缀，且不能与 file_path 相同）

参数示例：

{"mode": "convert", "file_path": "/goraven/data/users/admin/temp/report.pdf", "output_path": "/goraven/data/users/admin/documents/report.md"}

返回 output_path 为保存后的文件路径。

## 规则

- 文件路径必须使用沙盒内的绝对路径（附件取 <goraven-upload> 标签中的路径）；工具对工作空间下的相对形式路径也做了兼容处理
- 读取大文档优先用 max_chars 或 pages 控制输出量，避免占满上下文
- 首次调用可能较慢（模型加载），属正常现象
- 若工具报错缺少 docling 依赖，把错误信息中的安装命令转告用户执行
- 转换后的 .md 存放位置由用户指定或放在源文件同目录
`

const SystemSkillGoRavenDocParseEn = `---
name: goraven-doc-parse
description: "Read or convert documents (PDF/DOCX/PPTX/XLSX/HTML/CSV/LaTeX/ASCIIDOC) to text. Images and scanned files not supported. Chat attachments are stored as original files — use this skill to read them."
---

# Document Reading & Conversion

Use the goraven_doc_parse tool to process documents. Two modes:

- **read**: extract document text into the tool response. Best for immediate reading and comprehension.
- **convert**: convert the document to a Markdown file on disk. Best for persistent storage or further processing.

### Supported Formats

- PDF (text layer only — scanned/image PDFs cannot be processed)
- DOCX (Word), PPTX (PowerPoint), XLSX (Excel)
- HTML, CSV, LaTeX, ASCIIDOC

### Not Supported

- Images (PNG/JPEG/TIFF, etc.)
- Scanned PDFs (no text layer)

Plain text files (.txt, .json, .yaml, .xml, .md, etc.) do not need this skill — use the file read tool directly.

## Mode 1: read

Parameters:

- mode (required): always "read"
- file_path (required): absolute path of the document in the sandbox; for chat attachments use the path given in the <goraven-upload> tag
- format: markdown (default, preserves headings/tables) or text (plain text)
- max_chars: max output characters, truncated beyond. Default 50000
- pages: page filter, e.g. "1-5" or "1,3,5-8", PDF only

Parameter examples:

{"mode": "read", "file_path": "/goraven/data/users/admin/temp/report.pdf"}

{"mode": "read", "file_path": "/goraven/data/users/admin/temp/report.pdf", "pages": "1-3", "max_chars": 30000}

{"mode": "read", "file_path": "/goraven/data/users/admin/temp/report.docx", "format": "text"}

Returns content as the document text; truncated = true means content was cut off — use pages or a larger max_chars to continue reading.

## Mode 2: convert

Parameters:

- mode (required): always "convert"
- file_path (required): absolute path of the source document in the sandbox; for chat attachments use the path given in the <goraven-upload> tag
- output_path (required): absolute path for the output Markdown file (.md extension, must differ from file_path)

Parameter example:

{"mode": "convert", "file_path": "/goraven/data/users/admin/temp/report.pdf", "output_path": "/goraven/data/users/admin/documents/report.md"}

Returns output_path as the saved file path.

## Rules

- Always use absolute paths inside the sandbox (for attachments use the path from the <goraven-upload> tag); workspace-relative path forms are tolerated as a fallback
- For large documents, prefer max_chars or pages to control output volume and avoid filling the context
- The first call may be slow (model loading) — this is normal
- If the tool reports a missing docling dependency, relay the install command from the error message to the user
- Place converted .md files where the user specifies, or in the same directory as the source file
`
