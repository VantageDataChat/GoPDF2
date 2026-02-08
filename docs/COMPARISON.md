# GoPDF2 vs PyMuPDF Feature Comparison

**[English](COMPARISON.md) | [中文](COMPARISON_zh.md)**

This document provides a detailed comparison between GoPDF2 (pure Go) and PyMuPDF (Python bindings for the MuPDF C library).

> GoPDF2 is a pure Go implementation with no CGO dependencies. PyMuPDF wraps the MuPDF C library via CGO, giving it access to a full rendering engine.
> The two libraries have different focuses: GoPDF2 specializes in PDF generation and editing; PyMuPDF excels at PDF parsing, extraction, and rendering.

---

## Overview

| Category | GoPDF2 | PyMuPDF | Notes |
|----------|--------|---------|-------|
| **Pure Go / No CGO** | ✅ | ❌ | GoPDF2 needs no C compiler, cross-compiles easily |
| **PDF Generation** | ✅ | ✅ | Both can create PDFs from scratch |
| **Open/Modify Existing PDF** | ✅ | ✅ | |
| **HTML Rendering to PDF** | ✅ | ✅ | GoPDF2 has built-in HTML parser; PyMuPDF uses Story class |
| **Text Extraction** | ✅ | ✅ | GoPDF2 has pure-Go content stream parser |
| **Page Rendering (Rasterization)** | ❌ | ✅ | Requires MuPDF rendering engine |
| **OCR** | ❌ | ✅ | Requires Tesseract + MuPDF |
| **SVG/HTML/EPUB Conversion** | ❌ | ✅ | Requires MuPDF multi-format support |
| **Image Extraction** | ✅ | ✅ | GoPDF2 has pure-Go image extraction |
| **Form Fields (AcroForm)** | ✅ | ✅ | GoPDF2 supports text, checkbox, dropdown, radio, button, signature fields |
| **Digital Signatures** | ✅ | ✅ | GoPDF2 supports PKCS#7 signing and verification |

---

## Detailed Comparison

### 1. Document Lifecycle

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Create blank PDF | `Start(Config)` | `Document()` |
| Open from file | `OpenPDF(path)` | `Document(path)` |
| Open from memory | `OpenPDFFromBytes(data)` | `Document(stream=data)` |
| Open from stream | `OpenPDFFromStream(rs)` | `Document(stream=data)` |
| Save to file | `WritePdf(path)` | `save(path)` |
| Save to memory | `GetBytesPdf()` | `tobytes()` |
| Incremental save | ✅ `IncrementalSave()` | ✅ `saveIncr()` |
| Document cloning | ✅ `Clone()` | ❌ (manual serialize) |
| Close document | Auto GC | `close()` |
| Open encrypted PDF | ❌ | ✅ `authenticate()` |
| Multi-format support (XPS/EPUB/HTML) | ❌ | ✅ |

### 2. Page Management

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Add new page | ✅ `AddPage()` | ✅ `new_page()` |
| Delete page | ✅ `DeletePage(n)` | ✅ `delete_page(n)` |
| Batch delete pages | ✅ `DeletePages()` | ✅ `delete_pages()` |
| Copy page (reference) | ✅ `CopyPage(n)` | ✅ `copy_page()` |
| Full copy page | ✅ `CopyPage(n)` | ✅ `fullcopy_page()` |
| Move page | ✅ `MovePage()` | ✅ `move_page()` |
| Page reordering/selection | ✅ `SelectPages()` | ✅ `select()` |
| Extract pages to new doc | ✅ `ExtractPages()` | ✅ `select()` + `save()` |
| Merge multiple PDFs | ✅ `MergePages()` | ✅ `insert_pdf()` |
| Import PDF pages | ✅ `ImportPage()` | ✅ `insert_pdf()` |
| Insert arbitrary format | ❌ | ✅ `insert_file()` |
| Page rotation | ✅ `SetPageRotation()` | ✅ `set_rotation()` |
| Get page size | ✅ `GetPageSize()` | ✅ `page.rect` |
| Get all page sizes | ✅ `GetAllPageSizes()` | Manual iteration |
| Source PDF page count | ✅ `GetSourcePDFPageCount()` | ✅ `page_count` |
| Page cropbox | ✅ `SetPageCropBox()` | ✅ `page_cropbox()` |
| Chapter navigation (EPUB) | ❌ | ✅ `chapter_page_count()` etc. |

### 3. Text & Fonts

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Write text | ✅ `Cell()` / `Text()` | ✅ `insert_text()` |
| Multi-line text | ✅ `MultiCell()` | ✅ `insert_textbox()` |
| Text measurement | ✅ `MeasureTextWidth()` | ✅ `get_text_length()` |
| Word wrap | ✅ `SplitTextWithWordWrap()` | ✅ Built-in |
| TTF font embedding | ✅ Auto-subsetting | ✅ |
| Font subsetting | ✅ Automatic | ✅ `subset_fonts()` |
| Kerning | ✅ | ✅ |
| Text extraction | ✅ `ExtractPageText()` | ✅ `get_text()` |
| Text search | ❌ | ✅ `search_for()` |
| Text block/dict extraction | ✅ `ExtractTextFromPage()` | ✅ `get_text("dict")` |
| Font info query | ✅ `GetFonts()` | ✅ `get_page_fonts()` |
| Font extraction | ❌ | ✅ `extract_font()` |
| CJK font support | ✅ | ✅ |

### 4. Drawing

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Lines | ✅ `Line()` | ✅ `draw_line()` |
| Rectangles | ✅ `Rectangle()` | ✅ `draw_rect()` |
| Rounded rectangles | ✅ | ❌ (manual) |
| Ovals/Circles | ✅ `Oval()` | ✅ `draw_circle()` / `draw_oval()` |
| Polygons | ✅ `Polygon()` | ✅ `draw_polygon()` |
| Bezier curves | ✅ `Curve()` | ✅ `draw_bezier()` |
| Polylines | ✅ `Polyline()` | ✅ `draw_polyline()` |
| Sectors | ✅ `Sector()` | ✅ `draw_sector()` |
| Dash/line types | ✅ `SetLineType()` | ✅ |
| Fill/stroke colors | ✅ | ✅ |
| Rotation | ✅ `Rotate()` | ✅ |
| Transparency | ✅ `SetTransparency()` | ✅ |
| Graphics state save/restore | ✅ `SaveGraphicsState()` | ✅ |

### 5. Images

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Insert image (JPEG/PNG) | ✅ `Image()` | ✅ `insert_image()` |
| Insert from memory | ✅ `ImageFrom()` | ✅ |
| Image mask/crop | ✅ | ✅ |
| Image rotation | ✅ | ✅ |
| Image transparency | ✅ | ✅ |
| Image extraction | ✅ `ExtractImagesFromPage()` | ✅ `extract_image()` |
| Image info query | ✅ `ExtractImagesFromPage()` | ✅ `get_page_images()` |
| Image deletion | ✅ `DeleteImages()` | ✅ `delete_image()` |
| Image recompression | ✅ `RecompressImages()` | ✅ `rewrite_images()` |
| SVG insertion | ✅ `ImageSVG()` | ✅ |
| Pixmap rendering | ✅ `RenderPageToImage()` | ✅ `get_pixmap()` |

### 6. Annotations

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Text (sticky note) | ✅ | ✅ `add_text_annot()` |
| Highlight | ✅ | ✅ `add_highlight_annot()` |
| Underline | ✅ | ✅ `add_underline_annot()` |
| Strikeout | ✅ | ✅ `add_strikeout_annot()` |
| Rectangle | ✅ | ✅ `add_rect_annot()` |
| Circle | ✅ | ✅ `add_circle_annot()` |
| Free text | ✅ | ✅ `add_freetext_annot()` |
| Ink | ❌ | ✅ `add_ink_annot()` |
| Polyline | ❌ | ✅ `add_polyline_annot()` |
| Polygon | ❌ | ✅ `add_polygon_annot()` |
| Line | ❌ | ✅ `add_line_annot()` |
| Stamp | ❌ | ✅ `add_stamp_annot()` |
| Squiggly | ❌ | ✅ `add_squiggly_annot()` |
| Caret | ❌ | ✅ `add_caret_annot()` |
| File attachment | ❌ | ✅ `add_file_annot()` |
| Redaction | ❌ | ✅ `add_redact_annot()` |
| Delete annotation | ❌ | ✅ `delete_annot()` |
| Modify annotation | ❌ | ✅ Annot class methods |
| Iterate annotations | ❌ | ✅ `annots()` |
| Apply redactions | ❌ | ✅ `apply_redactions()` |

### 7. Watermarks

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Text watermark | ✅ `AddWatermarkText()` | Manual implementation |
| Image watermark | ✅ `AddWatermarkImage()` | Manual implementation |
| All-pages watermark | ✅ `AddWatermarkTextAllPages()` | Manual implementation |
| Tiled watermark | ✅ `Repeat: true` | Manual implementation |
| Watermark opacity/rotation | ✅ | Manual implementation |

> GoPDF2 provides ready-to-use watermark APIs. PyMuPDF requires manual composition via drawing APIs.

### 8. TOC / Bookmarks

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Get TOC | ✅ `GetTOC()` | ✅ `get_toc()` |
| Set TOC | ✅ `SetTOC()` | ✅ `set_toc()` |
| Hierarchical TOC | ✅ | ✅ |
| Add single bookmark | ✅ `AddOutline()` | ✅ |
| Modify single bookmark | ❌ | ✅ `set_toc_item()` |
| Delete single bookmark | ❌ | ✅ `del_toc_item()` |
| Bookmark color/bold/italic | ❌ | ✅ |
| Bookmark collapse control | ❌ | ✅ `collapse` parameter |

### 9. Metadata

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Standard metadata (Info) | ✅ `SetInfo()` | ✅ `set_metadata()` |
| XMP metadata | ✅ `SetXMPMetadata()` | ✅ `set_xml_metadata()` |
| Get XMP metadata | ✅ `GetXMPMetadata()` | ✅ `get_xml_metadata()` |
| PDF/A compliance | ✅ Via XMP | ✅ |
| Dublin Core | ✅ | ✅ |

### 10. Embedded Files

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Add embedded file | ✅ `AddEmbeddedFile()` | ✅ `embfile_add()` |
| Get embedded file | ❌ | ✅ `embfile_get()` |
| Delete embedded file | ❌ | ✅ `embfile_del()` |
| Update embedded file | ❌ | ✅ `embfile_upd()` |
| Embedded file info | ❌ | ✅ `embfile_info()` |
| List embedded files | ❌ | ✅ `embfile_names()` |
| Embedded file count | ❌ | ✅ `embfile_count()` |

### 11. Optional Content Groups (OCG/Layers)

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Add OCG | ✅ `AddOCG()` | ✅ `add_ocg()` |
| Get all OCGs | ✅ `GetOCGs()` | ✅ `get_ocgs()` |
| Layer configurations | ❌ | ✅ `add_layer()` / `get_layers()` |
| Switch layers | ❌ | ✅ `switch_layer()` |
| OCMD (membership dict) | ❌ | ✅ `set_ocmd()` / `get_ocmd()` |
| Layer UI config | ❌ | ✅ `layer_ui_configs()` |
| Batch set OCG states | ❌ | ✅ `set_layer()` |

### 12. Page Layout & Display

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Page layout | ✅ `SetPageLayout()` | ✅ `set_pagelayout()` |
| Page mode | ✅ `SetPageMode()` | ✅ `set_pagemode()` |
| MarkInfo | ❌ | ✅ `set_markinfo()` |
| Page labels | ✅ `SetPageLabels()` | ✅ `set_page_labels()` |
| Get page labels | ✅ `GetPageLabels()` | ✅ `get_page_labels()` |
| Find pages by label | ❌ | ✅ `get_page_numbers()` |

### 13. Security & Encryption

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Password protection | ✅ `PDFProtectionConfig` | ✅ `save(encryption=...)` |
| Permission control | ✅ | ✅ |
| Open encrypted PDF | ❌ | ✅ `authenticate()` |
| Encryption method selection | Limited | ✅ Multiple standards |
| Digital signatures | ✅ `SignPDF()` / `VerifySignature()` | ✅ `get_sigflags()` |

### 14. Document Scrubbing & Optimization

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Document scrubbing | ✅ `Scrub()` | ✅ `scrub()` |
| Garbage collection | ✅ `GarbageCollect()` | ✅ `save(garbage=N)` |
| Object deduplication | ✅ `GCDedup` | ✅ `save(garbage=3)` |
| Document statistics | ✅ `GetDocumentStats()` | Manual counting |
| Linearization (Web optimize) | ❌ | ✅ `save(linear=True)` |
| Content stream cleaning | ❌ | ✅ `save(clean=True)` |
| Image recompression | ✅ `RecompressImages()` | ✅ `rewrite_images()` |
| Colorspace conversion | ❌ | ✅ `recolor()` |

### 15. Content Element Operations

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| List page elements | ✅ `GetPageElements()` | ❌ (requires stream parsing) |
| Filter elements by type | ✅ `GetPageElementsByType()` | ❌ |
| Delete single element | ✅ `DeleteElement()` | ❌ |
| Delete elements by type | ✅ `DeleteElementsByType()` | ❌ |
| Delete elements in rect | ✅ `DeleteElementsInRect()` | ❌ |
| Clear page | ✅ `ClearPage()` | ❌ (requires stream rewrite) |
| Modify text element | ✅ `ModifyTextElement()` | ❌ |
| Modify element position | ✅ `ModifyElementPosition()` | ❌ |
| Insert new elements | ✅ `InsertLineElement()` etc. | ❌ |

> Content Element CRUD is unique to GoPDF2. GoPDF2 maintains a structured content cache in memory, enabling direct manipulation of individual elements. PyMuPDF's content streams are flat binary data that cannot be manipulated at the element level.

### 16. Low-level PDF Operations

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| PDF version control | ✅ `SetPDFVersion()` | Via header |
| Typed object IDs | ✅ `ObjID` | ✅ xref system |
| Object count | ✅ `GetObjectCount()` | ✅ `xref_length()` |
| Read object definition | ❌ | ✅ `xref_object()` |
| Modify object definition | ❌ | ✅ `update_object()` |
| Read/write dict keys | ❌ | ✅ `xref_get_key()` / `xref_set_key()` |
| Read/write stream data | ❌ | ✅ `xref_stream()` / `update_stream()` |
| Copy objects | ❌ | ✅ `xref_copy()` |
| PDF Catalog access | ❌ | ✅ `pdf_catalog()` |
| PDF Trailer access | ❌ | ✅ `pdf_trailer()` |

### 17. Journalling (Undo/Redo)

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Enable journalling | ❌ | ✅ `journal_enable()` |
| Undo/Redo | ❌ | ✅ `journal_undo()` / `journal_redo()` |
| Save/Load journal | ❌ | ✅ `journal_save()` / `journal_load()` |
| Named operations | ❌ | ✅ `journal_start_op()` |

### 18. Form Fields (Widgets/AcroForm)

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| Add form field | ✅ `AddFormField()` | ✅ `add_widget()` |
| Add text field | ✅ `AddTextField()` | ✅ `add_widget()` |
| Add checkbox | ✅ `AddCheckbox()` | ✅ `add_widget()` |
| Add dropdown | ✅ `AddDropdown()` | ✅ `add_widget()` |
| Add signature field | ✅ `AddSignatureField()` | ✅ `add_widget()` |
| Get form fields | ✅ `GetFormFields()` | ✅ Widget class |
| Delete form field | ❌ | ✅ `delete_widget()` |
| Modify form values | ❌ | ✅ Widget class |
| Form PDF detection | ❌ | ✅ `is_form_pdf` |
| Bake annotations/fields | ❌ | ✅ `bake()` |

### 19. Other Features

| Feature | GoPDF2 | PyMuPDF |
|---------|--------|---------|
| HTML rendering | ✅ `InsertHTMLBox()` | ✅ Story class |
| Table layout | ✅ `NewTableLayout()` | ❌ (manual) |
| Header/footer callbacks | ✅ `AddHeader()` / `AddFooter()` | ❌ (manual) |
| Placeholder text | ✅ `PlaceHolderText()` | ❌ |
| Paper size lookup | ✅ `PaperSize()` | ✅ `paper_size()` |
| Geometry utilities | ✅ `RectFrom` / `Matrix` | ✅ `Rect` / `Matrix` |
| Links | ✅ `AddExternalLink()` | ✅ `insert_link()` |
| Anchors | ✅ `SetAnchor()` | ✅ |
| Font container reuse | ✅ `FontContainer` | ❌ |
| Compression level control | ✅ `SetCompressLevel()` | ✅ `save(deflate=True)` |
| Trim-box | ✅ | ✅ |

---

## GoPDF2 Unique Strengths

1. **Pure Go** — No CGO, simple compilation, easy cross-compilation, container-friendly
2. **Content Element CRUD** — Direct manipulation of individual page elements (text, images, lines, etc.) — not possible in PyMuPDF
3. **Ready-to-use Watermark API** — Text/image watermarks, tiling, all-pages — one line of code
4. **Table Layout** — Built-in table generator
5. **Header/Footer Callbacks** — Auto-execute on every page
6. **Placeholder Text** — Reserve-then-fill pattern
7. **Font Container Reuse** — Share parsed fonts across documents
8. **Document Cloning** — One-line deep copy
9. **Document Statistics** — One-line document structure overview

## PyMuPDF Unique Strengths

1. **Full-fidelity Page Rendering** — High-quality PDF rendering via MuPDF engine
2. **Text Search** — Search text on pages with position results
3. **OCR** — Optical character recognition via Tesseract
4. **Multi-format Support** — Open XPS, EPUB, HTML, SVG, images, etc.
5. **Journalling** — Undo/redo operations
6. **Redaction Annotations** — Securely and permanently remove sensitive content
7. **Low-level PDF Operations** — Direct read/write of PDF objects, dict keys, streams
8. **Linearization** — Generate web-optimized PDFs
9. **More Annotation Types** — Ink, polyline, polygon, stamp, squiggly, caret, file attachment, redaction, etc.

---

## Recommended Use Cases

| Scenario | Recommendation |
|----------|---------------|
| Generate PDF reports in Go projects | GoPDF2 |
| Overlay content on existing PDFs | GoPDF2 |
| Fine-grained page element manipulation | GoPDF2 |
| Batch watermarking | GoPDF2 |
| Containerized / no C compiler environments | GoPDF2 |
| Extract text/images from PDFs | GoPDF2 (basic) / PyMuPDF (advanced) |
| PDF to image conversion | PyMuPDF |
| OCR text recognition | PyMuPDF |
| Form filling and processing | GoPDF2 (create) / PyMuPDF (full read/write) |
| Multi-format document conversion | PyMuPDF |
| Low-level PDF object manipulation | PyMuPDF |
| Redaction / secure content removal | PyMuPDF |

---

## Future Expansion Possibilities

Features theoretically implementable in pure Go, with varying complexity:

| Feature | Difficulty | Notes |
|---------|-----------|-------|
| More annotation types | Medium | Ink, polyline, polygon, stamp, etc. |
| MarkInfo | Low | Simple dictionary write |
| Embedded file read/delete | Medium | Requires Names dictionary parsing |
| ~~Form fields (AcroForm)~~ | ~~High~~ | ✅ Implemented — `AddFormField()`, `AddTextField()`, `AddCheckbox()`, `AddDropdown()` |
| ~~Text extraction~~ | ~~Very High~~ | ✅ Implemented — `ExtractTextFromPage()`, `ExtractPageText()` |
| ~~Image extraction~~ | ~~Very High~~ | ✅ Implemented — `ExtractImagesFromPage()`, `ExtractImagesFromAllPages()` |
| ~~Digital signatures~~ | ~~High~~ | ✅ Implemented — `SignPDF()`, `VerifySignature()` |
| Page rendering | Not feasible | ✅ Basic rendering implemented — `RenderPageToImage()` (lightweight; full-fidelity requires MuPDF) |
| OCR | Not feasible | Requires Tesseract or similar |
| Journalling (undo/redo) | High | Requires operation recording and replay |
| Linearization | High | Requires PDF file structure reorganization |
