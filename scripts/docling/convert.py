#!/usr/bin/env python3
"""
Docling document converter CLI.
Usage: python3 convert.py --input /path/to/file.pdf [--output /path/to/output.md] [--ocr]
Converts various document formats to Markdown.

Supported formats:
- PDF, DOCX, PPTX, XLSX, HTML, MD, ASCIIDOC, CSV, LATEX
- AUDIO (WAV/MP3, requires ASR), VTT (WebVTT subtitles)
- JSON_DOCLING, XML_JATS, XML_USPTO, XML_XBRL, METS_GBS
- IMAGE (PNG/JPEG/TIFF/BMP/WEBP, requires --ocr)
- Plain text files (.txt, .log, .json, .xml, .yaml, .yml, .toml, .ini, .cfg)
"""

import argparse
import logging
import os
import platform
import sys

from docling.document_converter import DocumentConverter, PdfFormatOption
from docling.datamodel.pipeline_options import (
    PdfPipelineOptions,
    TableStructureOptions,
    TableFormerMode,
    EasyOcrOptions,
    OcrMacOptions,
)
from docling.datamodel.base_models import InputFormat


def build_converter(enable_ocr: bool) -> DocumentConverter:
    pipeline_opts = PdfPipelineOptions()
    pipeline_opts.do_ocr = enable_ocr
    pipeline_opts.do_table_structure = True
    pipeline_opts.table_structure_options = TableStructureOptions(
        do_cell_matching=True,
        mode=TableFormerMode.ACCURATE,
    )

    if enable_ocr:
        if platform.system() == "Darwin":
            pipeline_opts.ocr_options = OcrMacOptions()
        else:
            pipeline_opts.ocr_options = EasyOcrOptions()

    allowed = [
        InputFormat.PDF,
        InputFormat.DOCX,
        InputFormat.PPTX,
        InputFormat.XLSX,
        InputFormat.HTML,
        InputFormat.MD,
        InputFormat.ASCIIDOC,
        InputFormat.CSV,
        InputFormat.LATEX,
        InputFormat.JSON_DOCLING,
        InputFormat.XML_JATS,
        InputFormat.XML_USPTO,
        InputFormat.XML_XBRL,
        InputFormat.METS_GBS,
    ]
    if enable_ocr:
        allowed.append(InputFormat.IMAGE)

    return DocumentConverter(
        format_options={
            InputFormat.PDF: PdfFormatOption(pipeline_options=pipeline_opts),
        },
        allowed_formats=allowed,
    )


def main():
    parser = argparse.ArgumentParser(description="Docling document to Markdown converter")
    parser.add_argument("--input", required=True, help="Path to input document")
    parser.add_argument("--output", default=None, help="Path to output markdown file")
    parser.add_argument("--ocr", action="store_true", help="Enable OCR for image-based documents")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    # 配置日志
    log_level = logging.DEBUG if args.verbose else logging.INFO
    logging.basicConfig(level=log_level, format="%(asctime)s - %(levelname)s - %(message)s")
    # 启用 docling 内部日志
    logging.getLogger("docling").setLevel(log_level)

    input_path = args.input
    if not os.path.isfile(input_path):
        print(f"ERROR: input file not found: {input_path}", file=sys.stderr)
        sys.exit(1)

    if args.output:
        output_path = args.output
    else:
        base = os.path.splitext(input_path)[0]
        output_path = f"{base}.md"

    # 纯文本格式直接复制
    ext = os.path.splitext(input_path)[1].lower()
    plain_text_exts = {".txt", ".log", ".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg"}
    if ext in plain_text_exts:
        logging.info(f"Copying plain text file: {input_path}")
        os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
        with open(input_path, "r", encoding="utf-8", errors="replace") as fin:
            content = fin.read()
        with open(output_path, "w", encoding="utf-8") as fout:
            fout.write(content)
        print(output_path)
        return

    try:
        logging.info(f"Initializing converter (OCR={args.ocr})...")
        converter = build_converter(args.ocr)

        logging.info(f"Converting: {input_path}")
        result = converter.convert(input_path)

        logging.info("Exporting to markdown...")
        markdown = result.document.export_to_markdown()

        os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
        with open(output_path, "w", encoding="utf-8") as f:
            f.write(markdown)

        logging.info(f"Done: {output_path}")
        print(output_path)
    except Exception as e:
        logging.error(f"Conversion failed: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
