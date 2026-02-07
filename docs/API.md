# GoPDF2 API Reference

**[English](API.md) | [中文](API_zh.md)**

Package `gopdf` — [github.com/VantageDataChat/GoPDF2](https://github.com/VantageDataChat/GoPDF2)

---

## Types

### GoPdf

The main struct for creating PDF documents.

```go
type GoPdf struct { /* unexported fields */ }
```

### Config

```go
type Config struct {
    Unit              int                 // UnitPT, UnitMM, UnitCM, UnitIN, UnitPX
    ConversionForUnit float64             // Custom conversion factor (overrides Unit)
    TrimBox           Box                 // Default trim box for all pages
    PageSize          Rect                // Default page size
    Protection        PDFProtectionConfig // Password protection settings
}
```

### Rect

```go
type Rect struct {
    W float64 // Width
    H float64 // Height
}
```

### Point

```go
type Point struct {
    X float64
    Y float64
}
```

### CellOption

```go
type CellOption struct {
    Align        int            // Left|Center|Right|Top|Bottom|Middle
    Border       int            // Left|Top|Right|Bottom|AllBorders
    Float        int            // Right|Bottom
    Transparency *Transparency
    BreakOption  *BreakOption
    CoefUnderlinePosition  float64
    CoefLineHeight         float64
    CoefUnderlineThickness float64
}
```

### HTMLBoxOption

```go
type HTMLBoxOption struct {
    DefaultFontFamily    string    // Required. Font family when HTML does not specify one.
    DefaultFontSize      float64   // Default font size in points (default 12).
    DefaultColor         [3]uint8  // Default text color {R, G, B}.
    LineSpacing          float64   // Extra line spacing in document units.
    BoldFontFamily       string    // Font family for <b>/<strong>.
    ItalicFontFamily     string    // Font family for <i>/<em>.
    BoldItalicFontFamily string    // Font family for bold+italic.
}
```

### TtfOption

```go
type TtfOption struct {
    UseKerning                bool
    Style                     int               // Regular|Bold|Italic
    OnGlyphNotFound           func(r rune)      // Debug callback when glyph is missing
    OnGlyphNotFoundSubstitute func(r rune) rune // Substitution callback
}
```

### Transparency

```go
type Transparency struct {
    Alpha         float64       // 0.0 (fully transparent) to 1.0 (opaque)
    BlendModeType BlendModeType // e.g. "/Normal", "/Multiply", etc.
}
```

### ImageOptions

```go
type ImageOptions struct {
    DegreeAngle    float64
    VerticalFlip   bool
    HorizontalFlip bool
    X, Y           float64
    Rect           *Rect
    Mask           *MaskOptions
    Crop           *CropOptions
    Transparency   *Transparency
}
```

### Predefined Page Sizes

`PageSizeA3`, `PageSizeA4`, `PageSizeA5`, `PageSizeLetter`, `PageSizeLegal`

### Unit Constants

`UnitPT`, `UnitMM`, `UnitCM`, `UnitIN`, `UnitPX`

### Font Style Constants

`Regular` (0), `Italic` (1), `Bold` (2), `Underline` (4)

### Alignment Constants

`Left`, `Right`, `Top`, `Bottom`, `Center`, `Middle`

---

## Document Lifecycle

### Start

```go
func (gp *GoPdf) Start(config Config)
```

Initialize the PDF document with the given configuration.

### AddPage / AddPageWithOption

```go
func (gp *GoPdf) AddPage()
func (gp *GoPdf) AddPageWithOption(opt PageOption)
```

### WritePdf

```go
func (gp *GoPdf) WritePdf(pdfPath string) error
```

Write the PDF to a file.

### Write / WriteTo / GetBytesPdf

```go
func (gp *GoPdf) Write(w io.Writer) error
func (gp *GoPdf) WriteTo(w io.Writer) (int64, error)
func (gp *GoPdf) GetBytesPdf() []byte
func (gp *GoPdf) GetBytesPdfReturnErr() ([]byte, error)
```

---

## Font Management

### AddTTFFont

```go
func (gp *GoPdf) AddTTFFont(family string, ttfpath string) error
func (gp *GoPdf) AddTTFFontWithOption(family string, ttfpath string, option TtfOption) error
func (gp *GoPdf) AddTTFFontByReader(family string, rd io.Reader) error
func (gp *GoPdf) AddTTFFontByReaderWithOption(family string, rd io.Reader, option TtfOption) error
func (gp *GoPdf) AddTTFFontData(family string, fontData []byte) error
func (gp *GoPdf) AddTTFFontDataWithOption(family string, fontData []byte, option TtfOption) error
```

Add a TTF font. The font is automatically subsetted — only glyphs used in the document are embedded in the output PDF.

### SetFont

```go
func (gp *GoPdf) SetFont(family string, style string, size interface{}) error
func (gp *GoPdf) SetFontWithStyle(family string, style int, size interface{}) error
func (gp *GoPdf) SetFontSize(fontSize float64) error
```

Set the active font. `style` string accepts `""`, `"B"`, `"I"`, `"U"`, or combinations like `"BI"`.

### KernOverride

```go
func (gp *GoPdf) KernOverride(family string, fn FuncKernOverride) error
```

---

## Text

### Cell / CellWithOption

```go
func (gp *GoPdf) Cell(rectangle *Rect, text string) error
func (gp *GoPdf) CellWithOption(rectangle *Rect, text string, opt CellOption) error
```

Render text in a cell at the current position.

### MultiCell / MultiCellWithOption

```go
func (gp *GoPdf) MultiCell(rectangle *Rect, text string) error
func (gp *GoPdf) MultiCellWithOption(rectangle *Rect, text string, opt CellOption) error
```

Render text with automatic line wrapping.

### Text

```go
func (gp *GoPdf) Text(text string) error
```

Print text at the current position without a cell box.

### MeasureTextWidth

```go
func (gp *GoPdf) MeasureTextWidth(text string) (float64, error)
```

### MeasureCellHeightByText

```go
func (gp *GoPdf) MeasureCellHeightByText(text string) (float64, error)
```

### SplitText / SplitTextWithWordWrap / SplitTextWithOption

```go
func (gp *GoPdf) SplitText(text string, width float64) ([]string, error)
func (gp *GoPdf) SplitTextWithWordWrap(text string, width float64) ([]string, error)
func (gp *GoPdf) SplitTextWithOption(text string, width float64, opt *BreakOption) ([]string, error)
```

### IsFitMultiCell / IsFitMultiCellWithNewline

```go
func (gp *GoPdf) IsFitMultiCell(rectangle *Rect, text string) (bool, float64, error)
func (gp *GoPdf) IsFitMultiCellWithNewline(rectangle *Rect, text string) (bool, float64, error)
```

---

## HTML Rendering

### InsertHTMLBox

```go
func (gp *GoPdf) InsertHTMLBox(x, y, w, h float64, htmlStr string, opt HTMLBoxOption) (float64, error)
```

Render HTML content into a rectangular area on the PDF page.

**Parameters:**
- `x, y` — top-left corner of the box (document units)
- `w, h` — width and height of the box (document units)
- `htmlStr` — the HTML string to render
- `opt` — rendering options (see `HTMLBoxOption`)

**Returns:** the Y position after the last rendered content, and any error.

**Supported tags:** `<b>`, `<strong>`, `<i>`, `<em>`, `<u>`, `<br>`, `<p>`, `<div>`, `<h1>`–`<h6>`, `<font>`, `<span>`, `<img>`, `<ul>`, `<ol>`, `<li>`, `<hr>`, `<center>`, `<a>`, `<blockquote>`, `<sub>`, `<sup>`

**Supported inline CSS properties** (via `style` attribute): `color`, `font-size`, `font-family`, `font-weight`, `font-style`, `text-decoration`, `text-align`

**Color formats:** `#RGB`, `#RRGGBB`, `rgb(r,g,b)`, CSS named colors (black, red, blue, etc.)

**Font size formats:** `12pt`, `16px`, `1.5em`, `150%`, named sizes (small, medium, large, etc.)

---

## Position

### SetX / GetX / SetY / GetY / SetXY

```go
func (gp *GoPdf) SetX(x float64)
func (gp *GoPdf) GetX() float64
func (gp *GoPdf) SetY(y float64)
func (gp *GoPdf) GetY() float64
func (gp *GoPdf) SetXY(x, y float64)
```

### Br

```go
func (gp *GoPdf) Br(h float64)
```

Line break — moves Y down by `h` and resets X to the left margin.

### SetNewY / SetNewYIfNoOffset / SetNewXY

```go
func (gp *GoPdf) SetNewY(y float64, h float64)
func (gp *GoPdf) SetNewYIfNoOffset(y float64, h float64)
func (gp *GoPdf) SetNewXY(y float64, x, h float64)
```

Set Y position with automatic page break if the content would exceed the page.

---

## Color

```go
func (gp *GoPdf) SetTextColor(r, g, b uint8)
func (gp *GoPdf) SetTextColorCMYK(c, m, y, k uint8)
func (gp *GoPdf) SetStrokeColor(r, g, b uint8)
func (gp *GoPdf) SetStrokeColorCMYK(c, m, y, k uint8)
func (gp *GoPdf) SetFillColor(r, g, b uint8)
func (gp *GoPdf) SetFillColorCMYK(c, m, y, k uint8)
func (gp *GoPdf) SetGrayFill(grayScale float64)
func (gp *GoPdf) SetGrayStroke(grayScale float64)
```

---

## Drawing

```go
func (gp *GoPdf) Line(x1, y1, x2, y2 float64)
func (gp *GoPdf) Oval(x1, y1, x2, y2 float64)
func (gp *GoPdf) Polygon(points []Point, style string)
func (gp *GoPdf) Rectangle(x0, y0, x1, y1 float64, style string, radius float64, radiusPointNum int) error
func (gp *GoPdf) RectFromUpperLeft(x, y, w, h float64)
func (gp *GoPdf) RectFromUpperLeftWithStyle(x, y, w, h float64, style string)
func (gp *GoPdf) RectFromUpperLeftWithOpts(opts DrawableRectOptions) error
func (gp *GoPdf) RectFromLowerLeft(x, y, w, h float64)
func (gp *GoPdf) RectFromLowerLeftWithStyle(x, y, w, h float64, style string)
func (gp *GoPdf) RectFromLowerLeftWithOpts(opts DrawableRectOptions) error
func (gp *GoPdf) Curve(x0, y0, x1, y1, x2, y2, x3, y3 float64, style string)
func (gp *GoPdf) SetLineWidth(width float64)
func (gp *GoPdf) SetLineType(linetype string)
func (gp *GoPdf) SetCustomLineType(dashArray []float64, dashPhase float64)
```

---

## Image

```go
func (gp *GoPdf) Image(picPath string, x, y float64, rect *Rect) error
func (gp *GoPdf) ImageFrom(img image.Image, x, y float64, rect *Rect) error
func (gp *GoPdf) ImageByHolder(img ImageHolder, x, y float64, rect *Rect) error
func (gp *GoPdf) ImageByHolderWithOptions(img ImageHolder, opts ImageOptions) error
```

### ImageHolder Constructors

```go
func ImageHolderByPath(path string) (ImageHolder, error)
func ImageHolderByBytes(b []byte) (ImageHolder, error)
func ImageHolderByReader(r io.Reader) (ImageHolder, error)
```

---

## Rotation & Transparency

```go
func (gp *GoPdf) Rotate(angle, x, y float64)
func (gp *GoPdf) RotateReset()
func (gp *GoPdf) SetTransparency(transparency Transparency) error
func (gp *GoPdf) ClearTransparency()
```

---

## Margins

```go
func (gp *GoPdf) SetMargins(left, top, right, bottom float64)
func (gp *GoPdf) SetLeftMargin(margin float64)
func (gp *GoPdf) SetTopMargin(margin float64)
func (gp *GoPdf) SetMarginRight(margin float64)
func (gp *GoPdf) SetMarginBottom(margin float64)
func (gp *GoPdf) Margins() (float64, float64, float64, float64)
```

---

## Links & Anchors

```go
func (gp *GoPdf) AddExternalLink(url string, x, y, w, h float64)
func (gp *GoPdf) AddInternalLink(anchor string, x, y, w, h float64)
func (gp *GoPdf) SetAnchor(name string)
```

---

## Header & Footer

```go
func (gp *GoPdf) AddHeader(f func())
func (gp *GoPdf) AddFooter(f func())
```

---

## Import Existing PDF

```go
func (gp *GoPdf) ImportPage(sourceFile string, pageno int, box string) int
func (gp *GoPdf) ImportPageStream(sourceStream *io.ReadSeeker, pageno int, box string) int
func (gp *GoPdf) UseImportedTemplate(tplid int, x, y, w, h float64)
func (gp *GoPdf) ImportPagesFromSource(source interface{}, box string) error
```

---

## Placeholder Text

```go
func (gp *GoPdf) PlaceHolderText(placeHolderName string, placeHolderWidth float64) error
func (gp *GoPdf) FillInPlaceHoldText(placeHolderName string, text string, align int) error
```

---

## Page Management

```go
func (gp *GoPdf) GetNumberOfPages() int
func (gp *GoPdf) SetPage(pageno int) error
```

---

## Unit Conversion

```go
func (gp *GoPdf) UnitsToPoints(u float64) float64
func (gp *GoPdf) PointsToUnits(u float64) float64
func (gp *GoPdf) UnitsToPointsVar(u ...*float64)
func (gp *GoPdf) PointsToUnitsVar(u ...*float64)
func UnitsToPoints(t int, u float64) float64
func PointsToUnits(t int, u float64) float64
```

---

## Miscellaneous

```go
func (gp *GoPdf) SetInfo(info PdfInfo)
func (gp *GoPdf) GetInfo() PdfInfo
func (gp *GoPdf) SetCompressLevel(level int)
func (gp *GoPdf) SetNoCompression()
func (gp *GoPdf) SetCharSpacing(charSpacing float64) error
func (gp *GoPdf) IsCurrFontContainGlyph(r rune) (bool, error)
func (gp *GoPdf) SaveGraphicsState()
func (gp *GoPdf) RestoreGraphicsState()
func (gp *GoPdf) ClipPolygon(points []Point)
func (gp *GoPdf) AddOutline(title string)
func (gp *GoPdf) AddOutlineWithPosition(title string) *OutlineObj
```

---

## FontContainer (Reusable Font Cache)

```go
type FontContainer struct { /* ... */ }

func (fc *FontContainer) AddTTFFont(family string, ttfpath string) error
func (fc *FontContainer) AddTTFFontWithOption(family string, ttfpath string, option TtfOption) error
func (fc *FontContainer) AddTTFFontByReader(family string, rd io.Reader) error
func (fc *FontContainer) AddTTFFontByReaderWithOption(family string, rd io.Reader, option TtfOption) error
func (fc *FontContainer) AddTTFFontData(family string, fontData []byte) error
func (fc *FontContainer) AddTTFFontDataWithOption(family string, fontData []byte, option TtfOption) error
func (gp *GoPdf) AddTTFFontFromFontContainer(family string, container *FontContainer) error
```

Pre-parse fonts once and reuse across multiple `GoPdf` instances for better performance.
