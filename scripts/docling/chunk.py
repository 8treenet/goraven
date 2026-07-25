#!/usr/bin/env python3
"""
Docling document chunker CLI.
Usage: python3 chunk.py --input /path/to/file.pdf [--output /path/to/output.jsonl] [--ocr]
Converts documents to Markdown internally, then chunks using HybridChunker.
Output: JSON lines file, one JSON object per line per chunk.
Each chunk record: {"text", "heading", "page", "block_type", "chunk_index"}
"""

import argparse
import json
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
from docling.chunking import HybridChunker


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


def extract_chunks(result, output_path: str):
    logging.info("Chunking document with HybridChunker...")
    chunker = HybridChunker()
    chunks = list(chunker.chunk(result.document))

    os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
    with open(output_path, "w", encoding="utf-8") as fout:
        for idx, chunk in enumerate(chunks):
            heading = ""
            if chunk.meta.headings:
                heading = " > ".join(chunk.meta.headings)

            block_types = set()
            page_numbers = set()
            for item in chunk.meta.doc_items:
                if hasattr(item, "label") and item.label:
                    block_types.add(str(item.label))
                if hasattr(item, "prov") and item.prov:
                    for p in item.prov:
                        if hasattr(p, "page_no") and p.page_no is not None:
                            page_numbers.add(p.page_no)

            block_type = ", ".join(sorted(block_types)) if block_types else "text"
            page = min(page_numbers) if page_numbers else 1

            record = {
                "text": chunk.text,
                "heading": heading,
                "page": page,
                "block_type": block_type,
                "chunk_index": idx + 1,
            }
            fout.write(json.dumps(record, ensure_ascii=False) + "\n")

    logging.info(f"Done: {output_path} ({len(chunks)} chunks)")
    print(output_path)


def main():
    parser = argparse.ArgumentParser(description="Docling document chunker")
    parser.add_argument("--input", required=True, help="Path to input document")
    parser.add_argument("--output", default=None, help="Path to output JSONL file")
    parser.add_argument("--ocr", action="store_true", help="Enable OCR for image-based documents")
    parser.add_argument("--verbose", "-v", action="store_true", help="Enable verbose logging")
    args = parser.parse_args()

    log_level = logging.DEBUG if args.verbose else logging.INFO
    logging.basicConfig(level=log_level, format="%(asctime)s - %(levelname)s - %(message)s")
    logging.getLogger("docling").setLevel(log_level)

    if not os.path.isfile(args.input):
        print(f"ERROR: input file not found: {args.input}", file=sys.stderr)
        sys.exit(1)

    if args.output:
        output_path = args.output
    else:
        base = os.path.splitext(args.input)[0]
        output_path = f"{base}_chunks.jsonl"

    try:
        logging.info(f"Converting (OCR={args.ocr}): {args.input}")
        converter = build_converter(args.ocr)
        result = converter.convert(args.input)
        extract_chunks(result, output_path)
    except Exception as e:
        logging.error(f"Chunking failed: {e}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    main()
