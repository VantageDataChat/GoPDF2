# GoPDF2

[![Go Reference](https://pkg.go.dev/badge/github.com/VantageDataChat/gopdf2.svg)](https://pkg.go.dev/github.com/VantageDataChat/gopdf2)

**[English](README.md) | [中文](README_zh.md)**

GoPDF2 is a Go library for generating PDF documents. Forked from [gopdf](https://github.com/signintech/gopdf), it adds HTML rendering support and other enhancements.

Requires Go 1.13+.

## Features

- Unicode subfont embedding (Chinese, Japanese, Korean, etc.) with automatic subsetting to minimize file size
- **HTML-to-PDF rendering** via `InsertHTMLBox` — supports `<b>`, `<i>`, `<u>`, `<p>`, `<h1>`–`<h6>`, `<font>`, `<span style>`, `<img>`, `<ul>`/`<ol>`, `<hr>`, `<center>`, `<a>`, `<blockquote>`, and more
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
go get -u github.com/VantageDataChat/gopdf2
```

## Quick Start

### Print Text

```go
package main

import (
    "log"
    "github.com/VantageDataChat/gopdf2"
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
    "github.com/VantageDataChat/gopdf2"
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

## API Reference

See [docs/API.md](docs/API.md) (English) or [docs/API_zh.md](docs/API_zh.md) (中文).

## License

MIT
