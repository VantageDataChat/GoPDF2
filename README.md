# GoPDF2

[![Go Reference](https://pkg.go.dev/badge/github.com/VantageDataChat/GoPDF2.svg)](https://pkg.go.dev/github.com/VantageDataChat/GoPDF2)

**[English](README.md) | [中文](README_zh.md)**

GoPDF2 is a Go library for generating PDF documents. Forked from [gopdf](https://github.com/signintech/gopdf), it adds HTML rendering support, the ability to open and modify existing PDFs, and other enhancements.

Requires Go 1.13+.

## Features

- Unicode subfont embedding (Chinese, Japanese, Korean, etc.) with automatic subsetting to minimize file size
- **Open and modify existing PDFs** via `OpenPDF` — import all pages, overlay new content (text, images, HTML, drawings), and save
- **HTML-to-PDF rendering** via `InsertHTMLBox` — supports `<b>`, `<i>`, `<u>`, `<p>`, `<h1>`–`<h6>`, `<font>`, `<span style>`, `<img>`, `<ul>`/`<ol>`, `<hr>`, `<center>`, `<a>`, `<blockquote>`, and more
- **Watermark** — text and image watermarks with opacity, rotation, and tiling via `AddWatermarkText` / `AddWatermarkImage`
- **PDF annotations** — sticky notes, highlights, underlines, strikeouts, squares, circles, and free text via `AddAnnotation`
- **Page manipulation** — extract pages (`ExtractPages`), merge PDFs (`MergePages`), delete pages (`DeletePage`), copy pages (`CopyPage`)
- **Page inspection** — query page sizes (`GetPageSize`, `GetAllPageSizes`), source PDF page count (`GetSourcePDFPageCount`)
- Draw lines, ovals, rectangles (with rounded corners), curves, polygons
- Draw images (JPEG, PNG) with mask, crop, rotation, and transparency
- Password protection
- Font kerning
- Import existing PDF pages
- Table layout
- Header / footer callbacks
- Trim-box support
- Placeholder text (fill-in-later pattern)

## Installation

```bash
go get -u github.com/VantageDataChat/GoPDF2
```

## Quick Start

### Print Text

```go
package main

import (
    "log"
    "github.com/VantageDataChat/GoPDF2"
)

func main() {
    pdf := gopdf.GoPdf{}
    pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
    pdf.AddPage()

    if err := pdf.AddTTFFont("myfont", "path/to/font.ttf"); err != nil {
        log.Fatal(err)
    }
    if err := pdf.SetFont("myfont", "", 14); err != nil {
        log.Fatal(err)
    }

    pdf.Cell(nil, "Hello, World!")
    pdf.WritePdf("hello.pdf")
}
```

### InsertHTMLBox — Render HTML into PDF

The signature:

```go
func (gp *GoPdf) InsertHTMLBox(x, y, w, h float64, htmlStr string, opt HTMLBoxOption) (float64, error)
```

Example:

```go
package main

import (
    "log"
    "github.com/VantageDataChat/GoPDF2"
)

func main() {
    pdf := gopdf.GoPdf{}
    pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
    pdf.AddPage()

    // Load fonts — only glyphs actually used will be embedded (subset)
    if err := pdf.AddTTFFont("regular", "NotoSansSC-Regular.ttf"); err != nil {
        log.Fatal(err)
    }
    if err := pdf.AddTTFFontWithOption("bold", "NotoSansSC-Bold.ttf", gopdf.TtfOption{Style: gopdf.Bold}); err != nil {
        log.Fatal(err)
    }

    html := `
    <h2>GoPDF2 HTML Rendering</h2>
    <p>Supports <b>bold</b>, <i>italic</i>, <u>underline</u> and
       <font color="#e74c3c">colored text</font>.</p>
    <ul>
        <li>Auto line wrapping</li>
        <li>Ordered/unordered lists</li>
        <li>Image insertion</li>
    </ul>
    <hr/>
    <p style="font-size:10pt; color:gray">
        Font subsetting — only characters actually used are embedded.
    </p>`

    endY, err := pdf.InsertHTMLBox(40, 40, 515, 750, html, gopdf.HTMLBoxOption{
        DefaultFontFamily: "regular",
        DefaultFontSize:   12,
        BoldFontFamily:    "bold",
    })
    if err != nil {
        log.Fatal(err)
    }
    _ = endY // Y position after rendered content

    pdf.WritePdf("html_example.pdf")
}
```

#### HTMLBoxOption

| Field | Type | Description |
|---|---|---|
| `DefaultFontFamily` | `string` | Font family when HTML does not specify one (required) |
| `DefaultFontSize` | `float64` | Default font size in points (default 12) |
| `DefaultColor` | `[3]uint8` | Default text color `{R, G, B}` |
| `LineSpacing` | `float64` | Extra line spacing in document units |
| `BoldFontFamily` | `string` | Font family for `<b>` / `<strong>` |
| `ItalicFontFamily` | `string` | Font family for `<i>` / `<em>` |
| `BoldItalicFontFamily` | `string` | Font family for bold+italic |

#### Supported HTML Tags

| Tag | Effect |
|---|---|
| `<b>`, `<strong>` | Bold |
| `<i>`, `<em>` | Italic |
| `<u>` | Underline |
| `<br>` | Line break |
| `<p>`, `<div>` | Paragraph |
| `<h1>` – `<h6>` | Headings |
| `<font color="..." size="..." face="...">` | Font styling |
| `<span style="...">` | Inline CSS (color, font-size, font-family, font-weight, font-style, text-decoration, text-align) |
| `<img src="..." width="..." height="...">` | Image (local file path) |
| `<ul>`, `<ol>`, `<li>` | Lists |
| `<hr>` | Horizontal rule |
| `<center>` | Centered text |
| `<a href="...">` | Link (blue underlined text) |
| `<blockquote>` | Indented block |
| `<sub>`, `<sup>` | Subscript / superscript |

### Font Subsetting & File Size

GoPDF2 uses **font subsetting** by default. When you call `AddTTFFont`, the full TTF is parsed, but only the glyphs for characters actually used in the document are embedded in the output PDF. This is especially important for CJK fonts which can be 10-20 MB — the resulting PDF will only contain the few KB needed for the characters you used.

No extra configuration is needed; subsetting is automatic.

### Text Color

```go
// RGB
pdf.SetTextColor(156, 197, 140)
pdf.Cell(nil, "colored text")

// CMYK
pdf.SetTextColorCMYK(0, 6, 14, 0)
pdf.Cell(nil, "CMYK text")
```

### Image

```go
pdf.Image("gopher.jpg", 200, 50, nil)
```

### Links

```go
pdf.SetXY(30, 40)
pdf.Text("Link to example.com")
pdf.AddExternalLink("http://example.com/", 27.5, 28, 125, 15)
```

### Header & Footer

```go
pdf.AddHeader(func() {
    pdf.SetY(5)
    pdf.Cell(nil, "header")
})
pdf.AddFooter(func() {
    pdf.SetY(825)
    pdf.Cell(nil, "footer")
})
```

### Drawing

```go
// Line
pdf.SetLineWidth(2)
pdf.SetLineType("dashed")
pdf.Line(10, 30, 585, 30)

// Oval
pdf.Oval(100, 200, 500, 500)

// Polygon
pdf.SetStrokeColor(255, 0, 0)
pdf.SetFillColor(0, 255, 0)
pdf.Polygon([]gopdf.Point{{X: 10, Y: 30}, {X: 585, Y: 200}, {X: 585, Y: 250}}, "DF")

// Rounded rectangle
pdf.Rectangle(196.6, 336.8, 398.3, 379.3, "DF", 3, 10)
```

### Rotation

```go
pdf.Rotate(270.0, 100.0, 100.0)
pdf.Text("rotated")
pdf.RotateReset()
```

### Transparency

```go
pdf.SetTransparency(gopdf.Transparency{Alpha: 0.5, BlendModeType: ""})
```

### Password Protection

```go
pdf.Start(gopdf.Config{
    PageSize: *gopdf.PageSizeA4,
    Protection: gopdf.PDFProtectionConfig{
        UseProtection: true,
        Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy | gopdf.PermissionsModify,
        OwnerPass:     []byte("owner"),
        UserPass:      []byte("user"),
    },
})
```

### Import Existing PDF

```go
tpl := pdf.ImportPage("existing.pdf", 1, "/MediaBox")
pdf.UseImportedTemplate(tpl, 50, 100, 400, 0)
```

### Open and Modify Existing PDF

`OpenPDF` loads an existing PDF so you can draw new content on top of every page:

```go
pdf := gopdf.GoPdf{}
err := pdf.OpenPDF("input.pdf", nil)
if err != nil {
    log.Fatal(err)
}

pdf.AddTTFFont("myfont", "font.ttf")
pdf.SetFont("myfont", "", 14)

// Draw on page 1
pdf.SetPage(1)
pdf.SetXY(100, 100)
pdf.Cell(nil, "Watermark text")

// Draw on page 2
pdf.SetPage(2)
pdf.SetXY(200, 200)
pdf.Image("stamp.png", 200, 200, nil)

pdf.WritePdf("output.pdf")
```

Also available as `OpenPDFFromBytes(data, opt)` and `OpenPDFFromStream(rs, opt)`.

### Table

```go
table := pdf.NewTableLayout(10, 10, 25, 5)
table.AddColumn("CODE", 50, "left")
table.AddColumn("DESCRIPTION", 200, "left")
table.AddRow([]string{"001", "Product A"})
table.DrawTable()
```

### Placeholder Text

```go
pdf.PlaceHolderText("total", 30)
// ... after all pages created ...
pdf.FillInPlaceHoldText("total", "5", gopdf.Left)
```

### Watermark

Add text or image watermarks to PDF pages:

```go
// Single centered text watermark
pdf.SetPage(1)
pdf.AddWatermarkText(gopdf.WatermarkOption{
    Text:       "CONFIDENTIAL",
    FontFamily: "myfont",
    FontSize:   48,
    Opacity:    0.3,
    Angle:      45,
    Color:      [3]uint8{200, 200, 200},
})

// Tiled text watermark across the page
pdf.AddWatermarkText(gopdf.WatermarkOption{
    Text:       "DRAFT",
    FontFamily: "myfont",
    Repeat:     true,
})

// Apply text watermark to all pages
pdf.AddWatermarkTextAllPages(gopdf.WatermarkOption{
    Text:       "SAMPLE",
    FontFamily: "myfont",
})

// Image watermark (centered, 30% opacity)
pdf.AddWatermarkImage("logo.png", 0.3, 200, 200, 0)
```

### Annotations

Add PDF annotations (sticky notes, highlights, shapes, free text):

```go
// Sticky note
pdf.AddTextAnnotation(100, 100, "Reviewer", "Please check this section.")

// Highlight
pdf.AddHighlightAnnotation(50, 50, 200, 20, [3]uint8{255, 255, 0})

// Free text directly on the page
pdf.AddFreeTextAnnotation(100, 200, 250, 30, "Important note", 14)

// Full control via AddAnnotation
pdf.AddAnnotation(gopdf.AnnotationOption{
    Type:    gopdf.AnnotSquare,
    X:       50,
    Y:       300,
    W:       100,
    H:       50,
    Color:   [3]uint8{0, 0, 255},
    Content: "Review area",
})
```

### Page Manipulation

Extract, merge, delete, and copy pages:

```go
// Extract specific pages from a PDF
newPdf, _ := gopdf.ExtractPages("input.pdf", []int{1, 3, 5}, nil)
newPdf.WritePdf("pages_1_3_5.pdf")

// Merge multiple PDFs
merged, _ := gopdf.MergePages([]string{"doc1.pdf", "doc2.pdf"}, nil)
merged.WritePdf("merged.pdf")

// Delete a page (1-based)
pdf.DeletePage(2)

// Copy a page to the end
newPageNo, _ := pdf.CopyPage(1)
```

### Page Inspection

Query page sizes and metadata:

```go
// Get page size of a specific page
w, h, _ := pdf.GetPageSize(1)

// Get all page sizes
sizes := pdf.GetAllPageSizes()

// Get page count from a source PDF without importing
count, _ := gopdf.GetSourcePDFPageCount("input.pdf")

// Get page sizes from a source PDF
pageSizes, _ := gopdf.GetSourcePDFPageSizes("input.pdf")
```

## API Reference

See [docs/API.md](docs/API.md) (English) or [docs/API_zh.md](docs/API_zh.md) (中文).

## License

MIT
