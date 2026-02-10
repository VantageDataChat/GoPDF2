# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-02-11

### Features

- PDF generation with Unicode subfont embedding and automatic subsetting
- Open and modify existing PDFs via `OpenPDF` / `OpenPDFFromBytes` / `OpenPDFFromStream`
- HTML-to-PDF rendering via `InsertHTMLBox`
- Text and image watermarks with opacity, rotation, and tiling
- 20+ annotation types (sticky notes, highlights, shapes, redaction, etc.)
- Page manipulation: extract, merge, delete, copy, move, reorder
- Text extraction with positions (`ExtractTextFromPage` / `ExtractPageText`)
- Multi-format text extraction: plain text, blocks, words, HTML, JSON (`ExtractTextFormatted`)
- Text search and replace in content streams (`SearchText` / `ReplaceText`)
- Image extraction with metadata (`ExtractImagesFromPage`)
- Image recompression (`RecompressImages`)
- SVG insertion (`ImageSVG`)
- Basic page rendering (`RenderPageToImage`)
- Form fields / AcroForm: text, checkbox, dropdown, radio, button, signature
- Digital signatures: PKCS#7 signing and verification (`SignPDF` / `VerifySignature`)
- AES-128 and AES-256 encryption (`SetEncryption`)
- RC4 password protection
- Content element CRUD: list, query, delete, modify, reposition individual elements
- TOC / bookmarks: read and write hierarchical outline trees
- XMP metadata with Dublin Core and PDF/A support
- Embedded file attachments (full CRUD)
- Optional Content Groups (OCG / layers)
- Page labels, page layout, page mode
- Incremental save for fast updates on large documents
- Document cloning, scrubbing, garbage collection
- Linearization (web optimization)
- Content stream cleaning
- Colorspace conversion
- PDF/A validation (1a, 1b, 2a, 2b)
- PDF document comparison (`ComparePDF`)
- Page transformations: scale, rotate, skew (`ScalePage` / `TransformPage` / `ScalePageInPDF`)
- Enhanced redaction with custom colors and overlay text (`ApplyRedactionsEnhanced` / `RedactText`)
- Link CRUD: get, delete, extract links from existing PDFs
- Journalling (undo/redo)
- Geometry utilities: `RectFrom`, `Matrix`, `Distance`
- Drawing: lines, rectangles, rounded rectangles, ovals, polygons, polylines, curves, sectors
- Table layout, header/footer callbacks, placeholder text
- PDF version control (1.4–2.0)
- Pure Go, no CGO dependencies
