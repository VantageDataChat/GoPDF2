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

### Predefined Page Sizes

`PageSizeA0`–`PageSizeA10`, `PageSizeA3Landscape`–`PageSizeA10Landscape`, `PageSizeB0`–`PageSizeB10`, `PageSizeLetter`, `PageSizeLetterLandscape`, `PageSizeLegal`, `PageSizeLegalLandscape`, `PageSizeTabloid`, `PageSizeLedger`, `PageSizeStatement`, `PageSizeExecutive`, `PageSizeFolio`, `PageSizeQuarto`

### Unit Constants

`UnitPT`, `UnitMM`, `UnitCM`, `UnitIN`, `UnitPX`

### Font Style Constants

`Regular` (0), `Italic` (1), `Bold` (2), `Underline` (4)

### Alignment Constants

`Left`, `Right`, `Top`, `Bottom`, `Center`, `Middle`

---

## Document Lifecycle

```go
func (gp *GoPdf) Start(config Config)
func (gp *GoPdf) AddPage()
func (gp *GoPdf) AddPageWithOption(opt PageOption)
func (gp *GoPdf) WritePdf(pdfPath string) error
func (gp *GoPdf) Write(w io.Writer) error
func (gp *GoPdf) WriteTo(w io.Writer) (int64, error)
func (gp *GoPdf) GetBytesPdf() []byte
func (gp *GoPdf) GetBytesPdfReturnErr() ([]byte, error)
```

---

## Font Management

```go
func (gp *GoPdf) AddTTFFont(family string, ttfpath string) error
func (gp *GoPdf) AddTTFFontWithOption(family string, ttfpath string, option TtfOption) error
func (gp *GoPdf) AddTTFFontByReader(family string, rd io.Reader) error
func (gp *GoPdf) AddTTFFontData(family string, fontData []byte) error
func (gp *GoPdf) SetFont(family string, style string, size interface{}) error
func (gp *GoPdf) SetFontWithStyle(family string, style int, size interface{}) error
func (gp *GoPdf) SetFontSize(fontSize float64) error
func (gp *GoPdf) KernOverride(family string, fn FuncKernOverride) error
```

Fonts are automatically subsetted — only glyphs used in the document are embedded.

---

## Text

```go
func (gp *GoPdf) Cell(rectangle *Rect, text string) error
func (gp *GoPdf) CellWithOption(rectangle *Rect, text string, opt CellOption) error
func (gp *GoPdf) MultiCell(rectangle *Rect, text string) error
func (gp *GoPdf) MultiCellWithOption(rectangle *Rect, text string, opt CellOption) error
func (gp *GoPdf) Text(text string) error
func (gp *GoPdf) MeasureTextWidth(text string) (float64, error)
func (gp *GoPdf) MeasureCellHeightByText(text string) (float64, error)
func (gp *GoPdf) SplitText(text string, width float64) ([]string, error)
func (gp *GoPdf) SplitTextWithWordWrap(text string, width float64) ([]string, error)
func (gp *GoPdf) IsFitMultiCell(rectangle *Rect, text string) (bool, float64, error)
```

---

## HTML Rendering

```go
func (gp *GoPdf) InsertHTMLBox(x, y, w, h float64, htmlStr string, opt HTMLBoxOption) (float64, error)
```

Render HTML content into a rectangular area on the PDF page.

**Parameters:** `x, y` (top-left corner), `w, h` (box size), `htmlStr` (HTML string), `opt` (rendering options).

**Returns:** the Y position after the last rendered content.

**Supported tags:** `<b>`, `<strong>`, `<i>`, `<em>`, `<u>`, `<br>`, `<p>`, `<div>`, `<h1>`-`<h6>`, `<font>`, `<span>`, `<img>`, `<ul>`, `<ol>`, `<li>`, `<hr>`, `<center>`, `<a>`, `<blockquote>`, `<sub>`, `<sup>`

**Supported inline CSS:** `color`, `font-size`, `font-family`, `font-weight`, `font-style`, `text-decoration`, `text-align`

**Color formats:** `#RGB`, `#RRGGBB`, `rgb(r,g,b)`, CSS named colors

**Font size formats:** `12pt`, `16px`, `1.5em`, `150%`, named sizes (small, medium, large, etc.)

---

## Position

```go
func (gp *GoPdf) SetX(x float64)
func (gp *GoPdf) GetX() float64
func (gp *GoPdf) SetY(y float64)
func (gp *GoPdf) GetY() float64
func (gp *GoPdf) SetXY(x, y float64)
func (gp *GoPdf) Br(h float64)
func (gp *GoPdf) SetNewY(y float64, h float64)
func (gp *GoPdf) SetNewYIfNoOffset(y float64, h float64)
func (gp *GoPdf) SetNewXY(y float64, x, h float64)
```

---

## Color

```go
func (gp *GoPdf) SetTextColor(r, g, b uint8)
func (gp *GoPdf) SetTextColorCMYK(c, m, y, k uint8)
func (gp *GoPdf) SetStrokeColor(r, g, b uint8)
func (gp *GoPdf) SetFillColor(r, g, b uint8)
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
func (gp *GoPdf) Curve(x0, y0, x1, y1, x2, y2, x3, y3 float64, style string)
func (gp *GoPdf) SetLineWidth(width float64)
func (gp *GoPdf) SetLineType(linetype string)
```

---

## Image

```go
func (gp *GoPdf) Image(picPath string, x, y float64, rect *Rect) error
func (gp *GoPdf) ImageFrom(img image.Image, x, y float64, rect *Rect) error
func (gp *GoPdf) ImageByHolder(img ImageHolder, x, y float64, rect *Rect) error
func (gp *GoPdf) ImageByHolderWithOptions(img ImageHolder, opts ImageOptions) error
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
func (gp *GoPdf) Margins() (float64, float64, float64, float64)
```

---

## Links, Anchors, Header, Footer

```go
func (gp *GoPdf) AddExternalLink(url string, x, y, w, h float64)
func (gp *GoPdf) AddInternalLink(anchor string, x, y, w, h float64)
func (gp *GoPdf) SetAnchor(name string)
func (gp *GoPdf) AddHeader(f func())
func (gp *GoPdf) AddFooter(f func())
```

---

## Import Existing PDF

```go
func (gp *GoPdf) ImportPage(sourceFile string, pageno int, box string) int
func (gp *GoPdf) UseImportedTemplate(tplid int, x, y, w, h float64)
func (gp *GoPdf) ImportPagesFromSource(source interface{}, box string) error
```

---

## Open and Modify Existing PDF

```go
func (gp *GoPdf) OpenPDF(pdfPath string, opt *OpenPDFOption) error
func (gp *GoPdf) OpenPDFFromBytes(pdfData []byte, opt *OpenPDFOption) error
func (gp *GoPdf) OpenPDFFromStream(rs *io.ReadSeeker, opt *OpenPDFOption) error
```

Open an existing PDF and import all pages so that new content can be drawn on top of them. The original page content is preserved as a background; any drawing method (Text, Cell, Image, InsertHTMLBox, Line, etc.) can then overlay new content.

After calling `OpenPDF`, use `SetPage(n)` (1-based) to switch pages, draw content, then call `WritePdf` to save.

### OpenPDFOption

```go
type OpenPDFOption struct {
    Box        string               // PDF box: "/MediaBox" (default), "/CropBox", etc.
    Protection *PDFProtectionConfig // Optional password protection for output.
}
```

---

## Placeholder Text

```go
func (gp *GoPdf) PlaceHolderText(placeHolderName string, placeHolderWidth float64) error
func (gp *GoPdf) FillInPlaceHoldText(placeHolderName string, text string, align int) error
```

---

## Page Management & Unit Conversion

```go
func (gp *GoPdf) GetNumberOfPages() int
func (gp *GoPdf) SetPage(pageno int) error
func (gp *GoPdf) UnitsToPoints(u float64) float64
func (gp *GoPdf) PointsToUnits(u float64) float64
```

---

## Miscellaneous

```go
func (gp *GoPdf) SetInfo(info PdfInfo)
func (gp *GoPdf) SetCompressLevel(level int)
func (gp *GoPdf) SetCharSpacing(charSpacing float64) error
func (gp *GoPdf) IsCurrFontContainGlyph(r rune) (bool, error)
func (gp *GoPdf) SaveGraphicsState()
func (gp *GoPdf) RestoreGraphicsState()
func (gp *GoPdf) AddOutline(title string)
```

---

## FontContainer (Reusable Font Cache)

```go
type FontContainer struct { /* ... */ }
func (fc *FontContainer) AddTTFFont(family string, ttfpath string) error
func (fc *FontContainer) AddTTFFontByReader(family string, rd io.Reader) error
func (fc *FontContainer) AddTTFFontData(family string, fontData []byte) error
func (gp *GoPdf) AddTTFFontFromFontContainer(family string, container *FontContainer) error
```

Pre-parse fonts once and reuse across multiple `GoPdf` instances for better performance.

---

## Watermark

```go
func (gp *GoPdf) AddWatermarkText(opt WatermarkOption) error
func (gp *GoPdf) AddWatermarkImage(imgPath string, opacity float64, imgW, imgH float64, angle float64) error
func (gp *GoPdf) AddWatermarkTextAllPages(opt WatermarkOption) error
func (gp *GoPdf) AddWatermarkImageAllPages(imgPath string, opacity float64, imgW, imgH float64, angle float64) error
```

### WatermarkOption

```go
type WatermarkOption struct {
    Text           string    // Watermark text (required)
    FontFamily     string    // Font family (required, must be pre-loaded)
    FontSize       float64   // Font size in points (default 48)
    Angle          float64   // Rotation angle in degrees (default 45)
    Color          [3]uint8  // RGB color (default: light gray)
    Opacity        float64   // 0.0–1.0 (default 0.3)
    Repeat         bool      // Tile across the page
    RepeatSpacingX float64   // Horizontal spacing for tiling (default 150)
    RepeatSpacingY float64   // Vertical spacing for tiling (default 150)
}
```

---

## Annotations

```go
func (gp *GoPdf) AddAnnotation(opt AnnotationOption)
func (gp *GoPdf) AddTextAnnotation(x, y float64, title, content string)
func (gp *GoPdf) AddHighlightAnnotation(x, y, w, h float64, color [3]uint8)
func (gp *GoPdf) AddFreeTextAnnotation(x, y, w, h float64, text string, fontSize float64)
```

### AnnotationOption

```go
type AnnotationOption struct {
    Type         AnnotationType // AnnotText, AnnotHighlight, AnnotUnderline, AnnotStrikeOut, AnnotSquare, AnnotCircle, AnnotFreeText
    X, Y, W, H  float64        // Annotation rectangle in document units
    Title        string         // Author name (for sticky notes)
    Content      string         // Annotation text content
    Color        [3]uint8       // RGB color (default: yellow)
    Opacity      float64        // 0.0–1.0 (default 1.0)
    Open         bool           // Initially open popup (text annotations)
    FontSize     float64        // Font size for FreeText (default 12)
}
```

---

## Page Manipulation

```go
func (gp *GoPdf) DeletePage(pageNo int) error
func (gp *GoPdf) CopyPage(pageNo int) (int, error)
func ExtractPages(pdfPath string, pages []int, opt *OpenPDFOption) (*GoPdf, error)
func ExtractPagesFromBytes(pdfData []byte, pages []int, opt *OpenPDFOption) (*GoPdf, error)
func MergePages(pdfPaths []string, opt *OpenPDFOption) (*GoPdf, error)
func MergePagesFromBytes(pdfDataSlices [][]byte, opt *OpenPDFOption) (*GoPdf, error)
```

---

## Page Inspection

```go
func (gp *GoPdf) GetPageSize(pageNo int) (w, h float64, err error)
func (gp *GoPdf) GetAllPageSizes() []PageInfo
func GetSourcePDFPageCount(pdfPath string) (int, error)
func GetSourcePDFPageCountFromBytes(pdfData []byte) (int, error)
func GetSourcePDFPageSizes(pdfPath string) (map[int]PageInfo, error)
func GetSourcePDFPageSizesFromBytes(pdfData []byte) (map[int]PageInfo, error)
```

### PageInfo

```go
type PageInfo struct {
    Width      float64 // Page width in points
    Height     float64 // Page height in points
    PageNumber int     // 1-based page number
}
```

---

## Paper Size Lookup

```go
func PaperSize(name string) *Rect
func PaperSizeNames() []string
```

Look up predefined paper sizes by name (case-insensitive). Supported names: `a0`–`a10`, `b0`–`b10`, `letter`, `legal`, `tabloid`, `ledger`, `statement`, `executive`, `folio`, `quarto`. Append `-l` for landscape (e.g. `a4-l`, `letter-l`). Returns a copy of the Rect; returns nil if name is not recognized.

---

## Page Rotation

```go
func (gp *GoPdf) SetPageRotation(pageNo int, angle int) error
func (gp *GoPdf) GetPageRotation(pageNo int) (int, error)
```

Set or get the display rotation for a page. `angle` must be a multiple of 90 (0, 90, 180, 270). This sets the `/Rotate` entry in the page dictionary — it tells PDF viewers how to display the page but does not modify the page content.

---

## Page Reordering

```go
func (gp *GoPdf) SelectPages(pages []int) (*GoPdf, error)
func SelectPagesFromFile(pdfPath string, pages []int, opt *OpenPDFOption) (*GoPdf, error)
func SelectPagesFromBytes(pdfData []byte, pages []int, opt *OpenPDFOption) (*GoPdf, error)
```

Rearrange pages in a new document. Pages are 1-based and may be repeated. The current document is exported to bytes, then only the selected pages are re-imported in the specified order.

---

## Embedded Files

```go
func (gp *GoPdf) AddEmbeddedFile(ef EmbeddedFile) error
```

Attach a file to the PDF document. The file appears in the PDF viewer's attachment panel.

### EmbeddedFile

```go
type EmbeddedFile struct {
    Name        string    // Display name (required)
    Content     []byte    // File content (required)
    MimeType    string    // MIME type (e.g. "text/plain", "application/pdf")
    Description string    // Optional description
    ModDate     time.Time // Modification date (default: now)
}
```

---

## Geometry

### RectFrom

```go
type RectFrom struct {
    X, Y, W, H float64
}

func (r RectFrom) Contains(px, py float64) bool
func (r RectFrom) ContainsRect(other RectFrom) bool
func (r RectFrom) Intersects(other RectFrom) bool
func (r RectFrom) Intersection(other RectFrom) RectFrom
func (r RectFrom) Union(other RectFrom) RectFrom
func (r RectFrom) IsEmpty() bool
func (r RectFrom) Area() float64
func (r RectFrom) Center() Point
func (r RectFrom) Normalize() RectFrom
```

A positioned rectangle with geometric operations.

### Matrix

```go
type Matrix struct {
    A, B, C, D, E, F float64
}

func IdentityMatrix() Matrix
func TranslateMatrix(tx, ty float64) Matrix
func ScaleMatrix(sx, sy float64) Matrix
func RotateMatrix(degrees float64) Matrix
func (m Matrix) Multiply(other Matrix) Matrix
func (m Matrix) TransformPoint(x, y float64) (float64, float64)
func (m Matrix) IsIdentity() bool
```

2D affine transformation matrix. Transformation: `x' = a*x + c*y + e`, `y' = b*x + d*y + f`.

### Distance

```go
func Distance(p1, p2 Point) float64
```

Euclidean distance between two points.

---

## Content Element CRUD

### ContentElementType

```go
type ContentElementType int

const (
    ElementText             ContentElementType
    ElementImage            ContentElementType
    ElementLine             ContentElementType
    ElementRectangle        ContentElementType
    ElementOval             ContentElementType
    ElementPolygon          ContentElementType
    ElementCurve            ContentElementType
    ElementImportedTemplate ContentElementType
    ElementLineWidth        ContentElementType
    ElementLineType         ContentElementType
    // ... and more (ColorRGB, ColorCMYK, Rotate, etc.)
    ElementUnknown          ContentElementType
)

func (t ContentElementType) String() string
```

### ContentElement

```go
type ContentElement struct {
    Index    int                // 0-based position in the page's content cache
    Type     ContentElementType // Element type
    X, Y     float64           // Primary position
    X2, Y2   float64           // Secondary position (lines, ovals)
    Width    float64           // Width (rectangles, images)
    Height   float64           // Height (rectangles, images)
    Text     string            // Text content (text elements only)
    FontSize float64           // Font size (text elements only)
}
```

### Query Methods

```go
func (gp *GoPdf) GetPageElements(pageNo int) ([]ContentElement, error)
func (gp *GoPdf) GetPageElementsByType(pageNo int, elemType ContentElementType) ([]ContentElement, error)
func (gp *GoPdf) GetPageElementCount(pageNo int) (int, error)
```

### Delete Methods

```go
func (gp *GoPdf) DeleteElement(pageNo int, elementIndex int) error
func (gp *GoPdf) DeleteElementsByType(pageNo int, elemType ContentElementType) (int, error)
func (gp *GoPdf) DeleteElementsInRect(pageNo int, rx, ry, rw, rh float64) (int, error)
func (gp *GoPdf) ClearPage(pageNo int) error
```

### Modify Methods

```go
func (gp *GoPdf) ModifyTextElement(pageNo int, elementIndex int, newText string) error
func (gp *GoPdf) ModifyElementPosition(pageNo int, elementIndex int, x, y float64) error
```

### Insert Methods

```go
func (gp *GoPdf) InsertLineElement(pageNo int, x1, y1, x2, y2 float64) error
func (gp *GoPdf) InsertRectElement(pageNo int, x, y, w, h float64, style string) error
func (gp *GoPdf) InsertOvalElement(pageNo int, x1, y1, x2, y2 float64) error
func (gp *GoPdf) InsertElementAt(pageNo int, elementIndex int, newElement ICacheContent) error
func (gp *GoPdf) ReplaceElement(pageNo int, elementIndex int, newElement ICacheContent) error
```

---

## PDF Version Control

```go
type PDFVersion int

const (
    PDFVersion14 PDFVersion = 14 // PDF 1.4
    PDFVersion15 PDFVersion = 15 // PDF 1.5
    PDFVersion16 PDFVersion = 16 // PDF 1.6
    PDFVersion17 PDFVersion = 17 // PDF 1.7 (default)
    PDFVersion20 PDFVersion = 20 // PDF 2.0
)

func (v PDFVersion) String() string  // "1.7"
func (v PDFVersion) Header() string  // "%PDF-1.7"

func (gp *GoPdf) SetPDFVersion(v PDFVersion)
func (gp *GoPdf) GetPDFVersion() PDFVersion
```

---

## Garbage Collection

```go
type GarbageCollectLevel int

const (
    GCNone    GarbageCollectLevel = 0 // No-op
    GCCompact GarbageCollectLevel = 1 // Remove null objects, renumber
    GCDedup   GarbageCollectLevel = 2 // Also deduplicate identical objects
)

func (gp *GoPdf) GarbageCollect(level GarbageCollectLevel) int
func (gp *GoPdf) GetObjectCount() int
func (gp *GoPdf) GetLiveObjectCount() int
```

`GarbageCollect` removes null/deleted objects left by `DeletePage` and other operations. Returns the number of objects removed.

---

## Page Labels

```go
type PageLabelStyle string

const (
    PageLabelDecimal    PageLabelStyle = "D" // 1, 2, 3, ...
    PageLabelRomanUpper PageLabelStyle = "R" // I, II, III, ...
    PageLabelRomanLower PageLabelStyle = "r" // i, ii, iii, ...
    PageLabelAlphaUpper PageLabelStyle = "A" // A, B, C, ...
    PageLabelAlphaLower PageLabelStyle = "a" // a, b, c, ...
    PageLabelNone       PageLabelStyle = ""  // No numbering
)

type PageLabel struct {
    PageIndex int            // 0-based page index where this range starts
    Style     PageLabelStyle // Numbering style
    Prefix    string         // Optional prefix (e.g. "Appendix ")
    Start     int            // Starting number (default 1)
}

func (gp *GoPdf) SetPageLabels(labels []PageLabel)
func (gp *GoPdf) GetPageLabels() []PageLabel
```

---

## Object ID

```go
type ObjID int

func (id ObjID) Index() int     // 0-based array index
func (id ObjID) Ref() int       // 1-based PDF object reference
func (id ObjID) RefStr() string // "5 0 R"
func (id ObjID) IsValid() bool
```

Typed wrapper for PDF object identifiers, providing type safety over raw int indices.

---

## XMP Metadata

```go
type XMPMetadata struct {
    // Dublin Core
    Title       string
    Creator     []string
    Description string
    Subject     []string
    Rights      string
    Language    string

    // XMP Basic
    CreatorTool string
    CreateDate  time.Time
    ModifyDate  time.Time

    // PDF-specific
    Producer string
    Keywords string
    Trapped  string // "True", "False", "Unknown"

    // PDF/A conformance
    PDFAConformance string // "A", "B", "U"
    PDFAPart        int    // 1, 2, 3

    // Custom properties
    Custom map[string]string
}

func (gp *GoPdf) SetXMPMetadata(meta XMPMetadata)
func (gp *GoPdf) GetXMPMetadata() *XMPMetadata
```

Embeds an XMP metadata stream in the PDF catalog. Supports Dublin Core, XMP Basic, PDF-specific, and PDF/A conformance properties.

---

## Incremental Save

```go
func (gp *GoPdf) IncrementalSave(originalData []byte, modifiedIndices []int) ([]byte, error)
func (gp *GoPdf) WriteIncrementalPdf(pdfPath string, originalData []byte, modifiedIndices []int) error
```

Appends only modified objects to the original PDF data as an incremental update. If `modifiedIndices` is nil, all objects are written. This is significantly faster than a full rewrite for large documents.

---

## Document Cloning

```go
func (gp *GoPdf) Clone() (*GoPdf, error)
```

Creates a deep copy of the GoPdf instance by serializing and re-importing. The clone is fully independent — changes to one do not affect the other. Header/footer callbacks are not cloned.
