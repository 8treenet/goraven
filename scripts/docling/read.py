#!/usr/bin/env python3
"""
Docling document reader CLI.
Usage: python3 read.py --input /path/to/file.pdf [--format markdown|text] [--max-chars 50000] [--pages 1-5]

Reads document content and outputs to stdout for direct consumption.
Unlike convert.py (which writes a .md file), this script is designed for
agent skills that need to read document content inline.

Supported formats:
- PDF, DOCX, PPTX, XLSX, HTML, MD, ASCIIDOC, CSV, LATEX
- IMAGE (PNG/JPEG/TIFF/BMP/WEBP)
- JSON_DOCLING, XML_JATS, XML_USPTO, XML_XBRL, METS_GBS
- Plain text files (.txt, .log, .json, .xml, .yaml, .yml, .toml, .ini, .cfg)
"""

import argparse
import logging
import os
import sys

from docling.document_converter import DocumentConverter, PdfFormatOption
from docling.datamodel.pipeline_options import (
    PdfPipelineOptions,
    TableStructureOptions,
    TableFormerMode,
)
from docling.datamodel.base_models import InputFormat


PLAIN_TEXT_EXTS = {".txt", ".log", ".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg"}


def build_converter() -> DocumentConverter:
    pipeline_opts = PdfPipelineOptions()
    pipeline_opts.do_ocr = False
    pipeline_opts.do_table_structure = True
    pipeline_opts.table_structure_options = TableStructureOptions(
        do_cell_matching=True,
        mode=TableFormerMode.ACCURATE,
    )

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
        InputFormat.IMAGE,
    ]

    return DocumentConverter(
        format_options={
            InputFormat.PDF: PdfFormatOption(pipeline_options=pipeline_opts),
        },
        allowed_formats=allowed,
    )


def parse_pages(pages_str: str):
    """Parse page range string like '1-5' or '3' or '1,3,5-8' into a set of page numbers."""
    pages = set()
    for part in pages_str.split(","):
        part = part.strip()
        if "-" in part:
            start, end = part.split("-", 1)
            pages.update(range(int(start), int(end) + 1))
        else:
            pages.add(int(part))
    return pages


def read_plain_text(input_path: str, max_chars: int) -> str:
    """Read plain text file directly."""
    with open(input_path, "r", encoding="utf-8", errors="replace") as f:
        content = f.read()
    if max_chars > 0 and len(content) > max_chars:
        content = content[:max_chars] + f"\n\n... [truncated, total {len(content)} chars]"
    return content


def read_document(input_path: str, fmt: str, max_chars: int, pages_str: str | None) -> str:
    """Convert document via docling and extract content."""
    converter = build_converter()
    result = converter.convert(input_path)
    doc = result.document

    if fmt == "markdown":
        content = doc.export_to_markdown()
    else:
        content = doc.export_to_text()

    # Page filtering (only meaningful for markdown with page markers)
    if pages_str:
        target_pages = parse_pages(pages_str)
        lines = content.split("\n")
        filtered = []
        current_page = 1
        for line in lines:
            # docling markdown may contain page separators like <!-- Page 2 -->
            if line.strip().startswith("<!-- Page"):
                try:
                    current_page = int(line.strip().split("Page")[1].split("-->")[0].strip())
                except (ValueError, IndexError):
                    pass
            if current_page in target_pages:
                filtered.append(line)
        content = "\n".join(filtered)

    if max_chars > 0 and len(content) > max_chars:
        content = content[:max_chars] + f"\n\n... [truncated, total {len(content)} chars]"

    return content


def main():
    parser = argparse.ArgumentParser(description="Docling document reader - output content to stdout")
    parser.add_argument("--input", required=True, help="Path to input document")
    parser.add_argument("--format", choices=["markdown", "text"], default="markdown",
                        help="Output format: markdown (default) or plain text")
    parser.add_argument("--max-chars", type=int, default=0,
                        help="Max characters to output (0 = unlimited)")
    parser.add_argument("--pages", default=None,
                        help="Page range to read, e.g. '1-5' or '1,3,5-8'")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    # Logging to stderr only (stdout is reserved for content output)
    log_level = logging.DEBUG if args.verbose else logging.WARNING
    logging.basicConfig(level=log_level, format="%(asctime)s - %(levelname)s - %(message)s",
                        stream=sys.stderr)
    logging.getLogger("docling").setLevel(log_level)

    input_path = args.input
    if not os.path.isfile(input_path):
        print(f"ERROR: input file not found: {input_path}", file=sys.stderr)
        sys.exit(1)

    try:
        ext = os.path.splitext(input_path)[1].lower()
        if ext in PLAIN_TEXT_EXTS:
            content = read_plain_text(input_path, args.max_chars)
        else:
            content = read_document(input_path, args.format, args.max_chars, args.pages)

        sys.stdout.write(content)
        if not content.endswith("\n"):
            sys.stdout.write("\n")
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
