# GoPDF2 API 参考手册

**[English](API.md) | [中文](API_zh.md)**

包 `gopdf` — [github.com/VantageDataChat/GoPDF2](https://github.com/VantageDataChat/GoPDF2)

---

## 类型定义

### GoPdf

PDF 文档的主结构体。

```go
type GoPdf struct { /* 内部字段 */ }
```

### Config

```go
type Config struct {
    Unit              int                 // UnitPT, UnitMM, UnitCM, UnitIN, UnitPX
    ConversionForUnit float64             // 自定义转换系数（覆盖 Unit）
    TrimBox           Box                 // 所有页面的默认裁切框
    PageSize          Rect                // 默认页面尺寸
    Protection        PDFProtectionConfig // 密码保护设置
}
```

### Rect

```go
type Rect struct {
    W float64 // 宽度
    H float64 // 高度
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
    DefaultFontFamily    string    // 必填。HTML 未指定字体时使用的默认字体族。
    DefaultFontSize      float64   // 默认字号（磅），默认 12。
    DefaultColor         [3]uint8  // 默认文字颜色 {R, G, B}。
    LineSpacing          float64   // 额外行间距（文档单位）。
    BoldFontFamily       string    // <b>/<strong> 使用的字体族。
    ItalicFontFamily     string    // <i>/<em> 使用的字体族。
    BoldItalicFontFamily string    // 粗斜体使用的字体族。
}
```

### TtfOption

```go
type TtfOption struct {
    UseKerning                bool              // 是否启用字距调整
    Style                     int               // Regular|Bold|Italic
    OnGlyphNotFound           func(r rune)      // 字形缺失时的调试回调
    OnGlyphNotFoundSubstitute func(r rune) rune // 字形缺失时的替换回调
}
```

### Transparency

```go
type Transparency struct {
    Alpha         float64       // 0.0（完全透明）到 1.0（不透明）
    BlendModeType BlendModeType // 如 "/Normal"、"/Multiply" 等
}
```

### 预定义页面尺寸

`PageSizeA0`–`PageSizeA10`、`PageSizeA3Landscape`–`PageSizeA10Landscape`、`PageSizeB0`–`PageSizeB10`、`PageSizeLetter`、`PageSizeLetterLandscape`、`PageSizeLegal`、`PageSizeLegalLandscape`、`PageSizeTabloid`、`PageSizeLedger`、`PageSizeStatement`、`PageSizeExecutive`、`PageSizeFolio`、`PageSizeQuarto`

### 单位常量

`UnitPT`（磅）、`UnitMM`（毫米）、`UnitCM`（厘米）、`UnitIN`（英寸）、`UnitPX`（像素）

### 字体样式常量

`Regular` (0)、`Italic` (1)、`Bold` (2)、`Underline` (4)

### 对齐常量

`Left`、`Right`、`Top`、`Bottom`、`Center`、`Middle`

---

## 文档生命周期

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

## 字体管理

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

字体会自动子集化 — 输出 PDF 中仅嵌入文档实际使用的字形，有效控制文件大小。

---

## 文本

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

## HTML 渲染

```go
func (gp *GoPdf) InsertHTMLBox(x, y, w, h float64, htmlStr string, opt HTMLBoxOption) (float64, error)
```

将 HTML 内容渲染到 PDF 页面的指定矩形区域内。

**参数：** `x, y`（左上角坐标）、`w, h`（区域尺寸）、`htmlStr`（HTML 字符串）、`opt`（渲染选项）。

**返回值：** 最后渲染内容之后的 Y 坐标。

**支持的标签：** `<b>`、`<strong>`、`<i>`、`<em>`、`<u>`、`<br>`、`<p>`、`<div>`、`<h1>`-`<h6>`、`<font>`、`<span>`、`<img>`、`<ul>`、`<ol>`、`<li>`、`<hr>`、`<center>`、`<a>`、`<blockquote>`、`<sub>`、`<sup>`

**支持的内联 CSS 属性：** `color`、`font-size`、`font-family`、`font-weight`、`font-style`、`text-decoration`、`text-align`

**颜色格式：** `#RGB`、`#RRGGBB`、`rgb(r,g,b)`、CSS 命名颜色

**字号格式：** `12pt`、`16px`、`1.5em`、`150%`、命名尺寸（small、medium、large 等）

---

## 位置控制

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

## 颜色

```go
func (gp *GoPdf) SetTextColor(r, g, b uint8)
func (gp *GoPdf) SetTextColorCMYK(c, m, y, k uint8)
func (gp *GoPdf) SetStrokeColor(r, g, b uint8)
func (gp *GoPdf) SetFillColor(r, g, b uint8)
func (gp *GoPdf) SetGrayFill(grayScale float64)
func (gp *GoPdf) SetGrayStroke(grayScale float64)
```

---

## 绘图

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

## 图片

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

## 旋转与透明度

```go
func (gp *GoPdf) Rotate(angle, x, y float64)
func (gp *GoPdf) RotateReset()
func (gp *GoPdf) SetTransparency(transparency Transparency) error
func (gp *GoPdf) ClearTransparency()
```

---

## 边距

```go
func (gp *GoPdf) SetMargins(left, top, right, bottom float64)
func (gp *GoPdf) Margins() (float64, float64, float64, float64)
```

---

## 链接、锚点、页眉、页脚

```go
func (gp *GoPdf) AddExternalLink(url string, x, y, w, h float64)
func (gp *GoPdf) AddInternalLink(anchor string, x, y, w, h float64)
func (gp *GoPdf) SetAnchor(name string)
func (gp *GoPdf) AddHeader(f func())
func (gp *GoPdf) AddFooter(f func())
```

---

## 导入已有 PDF

```go
func (gp *GoPdf) ImportPage(sourceFile string, pageno int, box string) int
func (gp *GoPdf) UseImportedTemplate(tplid int, x, y, w, h float64)
func (gp *GoPdf) ImportPagesFromSource(source interface{}, box string) error
```

---

## 打开并修改已有 PDF

```go
func (gp *GoPdf) OpenPDF(pdfPath string, opt *OpenPDFOption) error
func (gp *GoPdf) OpenPDFFromBytes(pdfData []byte, opt *OpenPDFOption) error
func (gp *GoPdf) OpenPDFFromStream(rs *io.ReadSeeker, opt *OpenPDFOption) error
```

打开已有 PDF 并导入所有页面，使新内容可以叠加在原有页面之上。原始页面内容作为背景保留，之后可使用任何绘图方法（Text、Cell、Image、InsertHTMLBox、Line 等）叠加新内容。

调用 `OpenPDF` 后，使用 `SetPage(n)`（从 1 开始）切换页面，绘制内容，最后调用 `WritePdf` 保存。

### OpenPDFOption

```go
type OpenPDFOption struct {
    Box        string               // PDF 页面框："/MediaBox"（默认）、"/CropBox" 等。
    Protection *PDFProtectionConfig // 可选，输出 PDF 的密码保护设置。
}
```

---

## 占位文本

```go
func (gp *GoPdf) PlaceHolderText(placeHolderName string, placeHolderWidth float64) error
func (gp *GoPdf) FillInPlaceHoldText(placeHolderName string, text string, align int) error
```

---

## 页面管理与单位转换

```go
func (gp *GoPdf) GetNumberOfPages() int
func (gp *GoPdf) SetPage(pageno int) error
func (gp *GoPdf) UnitsToPoints(u float64) float64
func (gp *GoPdf) PointsToUnits(u float64) float64
```

---

## 其他

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

## FontContainer（可复用字体缓存）

```go
type FontContainer struct { /* ... */ }
func (fc *FontContainer) AddTTFFont(family string, ttfpath string) error
func (fc *FontContainer) AddTTFFontByReader(family string, rd io.Reader) error
func (fc *FontContainer) AddTTFFontData(family string, fontData []byte) error
func (gp *GoPdf) AddTTFFontFromFontContainer(family string, container *FontContainer) error
```

预先解析字体并在多个 `GoPdf` 实例间复用，提升性能。

---

## 水印

```go
func (gp *GoPdf) AddWatermarkText(opt WatermarkOption) error
func (gp *GoPdf) AddWatermarkImage(imgPath string, opacity float64, imgW, imgH float64, angle float64) error
func (gp *GoPdf) AddWatermarkTextAllPages(opt WatermarkOption) error
func (gp *GoPdf) AddWatermarkImageAllPages(imgPath string, opacity float64, imgW, imgH float64, angle float64) error
```

### WatermarkOption

```go
type WatermarkOption struct {
    Text           string    // 水印文字（必填）
    FontFamily     string    // 字体族（必填，需预先加载）
    FontSize       float64   // 字号，单位磅（默认 48）
    Angle          float64   // 旋转角度，单位度（默认 45）
    Color          [3]uint8  // RGB 颜色（默认：浅灰色）
    Opacity        float64   // 0.0–1.0（默认 0.3）
    Repeat         bool      // 是否平铺
    RepeatSpacingX float64   // 平铺水平间距（默认 150）
    RepeatSpacingY float64   // 平铺垂直间距（默认 150）
}
```

---

## 注释

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
    X, Y, W, H  float64        // 注释矩形区域（文档单位）
    Title        string         // 作者名（便签注释）
    Content      string         // 注释文本内容
    Color        [3]uint8       // RGB 颜色（默认：黄色）
    Opacity      float64        // 0.0–1.0（默认 1.0）
    Open         bool           // 弹出窗口是否初始打开（文本注释）
    FontSize     float64        // 自由文本字号（默认 12）
}
```

---

## 页面操作

```go
func (gp *GoPdf) DeletePage(pageNo int) error
func (gp *GoPdf) CopyPage(pageNo int) (int, error)
func ExtractPages(pdfPath string, pages []int, opt *OpenPDFOption) (*GoPdf, error)
func ExtractPagesFromBytes(pdfData []byte, pages []int, opt *OpenPDFOption) (*GoPdf, error)
func MergePages(pdfPaths []string, opt *OpenPDFOption) (*GoPdf, error)
func MergePagesFromBytes(pdfDataSlices [][]byte, opt *OpenPDFOption) (*GoPdf, error)
```

---

## 页面信息查询

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
    Width      float64 // 页面宽度（磅）
    Height     float64 // 页面高度（磅）
    PageNumber int     // 页码（从 1 开始）
}
```

---

## 纸张尺寸查询

```go
func PaperSize(name string) *Rect
func PaperSizeNames() []string
```

按名称查询预定义纸张尺寸（不区分大小写）。支持的名称：`a0`–`a10`、`b0`–`b10`、`letter`、`legal`、`tabloid`、`ledger`、`statement`、`executive`、`folio`、`quarto`。加 `-l` 后缀为横向（如 `a4-l`、`letter-l`）。返回 Rect 的副本；名称无法识别时返回 nil。

---

## 页面旋转

```go
func (gp *GoPdf) SetPageRotation(pageNo int, angle int) error
func (gp *GoPdf) GetPageRotation(pageNo int) (int, error)
```

设置或获取页面的显示旋转角度。`angle` 必须是 90 的倍数（0、90、180、270）。此操作设置页面字典中的 `/Rotate` 条目，告诉 PDF 阅读器如何显示页面，但不修改页面内容。

---

## 页面重排

```go
func (gp *GoPdf) SelectPages(pages []int) (*GoPdf, error)
func SelectPagesFromFile(pdfPath string, pages []int, opt *OpenPDFOption) (*GoPdf, error)
func SelectPagesFromBytes(pdfData []byte, pages []int, opt *OpenPDFOption) (*GoPdf, error)
```

在新文档中重新排列页面。页码从 1 开始，可以重复。当前文档导出为字节后，仅按指定顺序重新导入选定的页面。

---

## 嵌入文件

```go
func (gp *GoPdf) AddEmbeddedFile(ef EmbeddedFile) error
```

将文件附加到 PDF 文档。文件会显示在 PDF 阅读器的附件面板中。

### EmbeddedFile

```go
type EmbeddedFile struct {
    Name        string    // 显示名称（必填）
    Content     []byte    // 文件内容（必填）
    MimeType    string    // MIME 类型（如 "text/plain"、"application/pdf"）
    Description string    // 可选描述
    ModDate     time.Time // 修改日期（默认：当前时间）
}
```

---

## 几何工具

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

带位置的矩形，支持几何运算。

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

2D 仿射变换矩阵。变换公式：`x' = a*x + c*y + e`，`y' = b*x + d*y + f`。

### Distance

```go
func Distance(p1, p2 Point) float64
```

计算两点之间的欧几里得距离。

---

## 内容元素 CRUD

### ContentElementType

```go
type ContentElementType int

const (
    ElementText             ContentElementType // 文本
    ElementImage            ContentElementType // 图片
    ElementLine             ContentElementType // 线条
    ElementRectangle        ContentElementType // 矩形
    ElementOval             ContentElementType // 椭圆
    ElementPolygon          ContentElementType // 多边形
    ElementCurve            ContentElementType // 贝塞尔曲线
    ElementImportedTemplate ContentElementType // 导入的模板
    ElementLineWidth        ContentElementType // 线宽设置
    ElementLineType         ContentElementType // 线型设置
    // ... 更多类型（ColorRGB、ColorCMYK、Rotate 等）
    ElementUnknown          ContentElementType // 未知类型
)

func (t ContentElementType) String() string
```

### ContentElement

```go
type ContentElement struct {
    Index    int                // 页面内容缓存中的 0 基索引
    Type     ContentElementType // 元素类型
    X, Y     float64           // 主要位置坐标
    X2, Y2   float64           // 次要位置坐标（线条、椭圆）
    Width    float64           // 宽度（矩形、图片）
    Height   float64           // 高度（矩形、图片）
    Text     string            // 文本内容（仅文本元素）
    FontSize float64           // 字号（仅文本元素）
}
```

### 查询方法

```go
func (gp *GoPdf) GetPageElements(pageNo int) ([]ContentElement, error)
func (gp *GoPdf) GetPageElementsByType(pageNo int, elemType ContentElementType) ([]ContentElement, error)
func (gp *GoPdf) GetPageElementCount(pageNo int) (int, error)
```

### 删除方法

```go
func (gp *GoPdf) DeleteElement(pageNo int, elementIndex int) error
func (gp *GoPdf) DeleteElementsByType(pageNo int, elemType ContentElementType) (int, error)
func (gp *GoPdf) DeleteElementsInRect(pageNo int, rx, ry, rw, rh float64) (int, error)
func (gp *GoPdf) ClearPage(pageNo int) error
```

### 修改方法

```go
func (gp *GoPdf) ModifyTextElement(pageNo int, elementIndex int, newText string) error
func (gp *GoPdf) ModifyElementPosition(pageNo int, elementIndex int, x, y float64) error
```

### 插入方法

```go
func (gp *GoPdf) InsertLineElement(pageNo int, x1, y1, x2, y2 float64) error
func (gp *GoPdf) InsertRectElement(pageNo int, x, y, w, h float64, style string) error
func (gp *GoPdf) InsertOvalElement(pageNo int, x1, y1, x2, y2 float64) error
func (gp *GoPdf) InsertElementAt(pageNo int, elementIndex int, newElement ICacheContent) error
func (gp *GoPdf) ReplaceElement(pageNo int, elementIndex int, newElement ICacheContent) error
```

---

## PDF 版本控制

```go
type PDFVersion int

const (
    PDFVersion14 PDFVersion = 14 // PDF 1.4
    PDFVersion15 PDFVersion = 15 // PDF 1.5
    PDFVersion16 PDFVersion = 16 // PDF 1.6
    PDFVersion17 PDFVersion = 17 // PDF 1.7（默认）
    PDFVersion20 PDFVersion = 20 // PDF 2.0
)

func (v PDFVersion) String() string  // "1.7"
func (v PDFVersion) Header() string  // "%PDF-1.7"

func (gp *GoPdf) SetPDFVersion(v PDFVersion)
func (gp *GoPdf) GetPDFVersion() PDFVersion
```

---

## 垃圾回收

```go
type GarbageCollectLevel int

const (
    GCNone    GarbageCollectLevel = 0 // 不执行
    GCCompact GarbageCollectLevel = 1 // 移除空对象，重新编号
    GCDedup   GarbageCollectLevel = 2 // 额外去重相同对象
)

func (gp *GoPdf) GarbageCollect(level GarbageCollectLevel) int
func (gp *GoPdf) GetObjectCount() int
func (gp *GoPdf) GetLiveObjectCount() int
```

`GarbageCollect` 移除 `DeletePage` 等操作留下的空/已删除对象。返回移除的对象数量。

---

## 页面标签

```go
type PageLabelStyle string

const (
    PageLabelDecimal    PageLabelStyle = "D" // 1, 2, 3, ...
    PageLabelRomanUpper PageLabelStyle = "R" // I, II, III, ...
    PageLabelRomanLower PageLabelStyle = "r" // i, ii, iii, ...
    PageLabelAlphaUpper PageLabelStyle = "A" // A, B, C, ...
    PageLabelAlphaLower PageLabelStyle = "a" // a, b, c, ...
    PageLabelNone       PageLabelStyle = ""  // 无编号
)

type PageLabel struct {
    PageIndex int            // 此标签范围起始的 0 基页面索引
    Style     PageLabelStyle // 编号样式
    Prefix    string         // 可选前缀（如 "附录 "）
    Start     int            // 起始编号（默认 1）
}

func (gp *GoPdf) SetPageLabels(labels []PageLabel)
func (gp *GoPdf) GetPageLabels() []PageLabel
```

---

## 对象 ID

```go
type ObjID int

func (id ObjID) Index() int     // 0 基数组索引
func (id ObjID) Ref() int       // 1 基 PDF 对象引用号
func (id ObjID) RefStr() string // "5 0 R"
func (id ObjID) IsValid() bool
```

PDF 对象标识符的类型化包装器，提供比原始 int 索引更安全的类型检查。

---

## XMP 元数据

```go
type XMPMetadata struct {
    // Dublin Core
    Title       string
    Creator     []string
    Description string
    Subject     []string
    Rights      string
    Language    string

    // XMP 基本属性
    CreatorTool string
    CreateDate  time.Time
    ModifyDate  time.Time

    // PDF 特定属性
    Producer string
    Keywords string
    Trapped  string // "True"、"False"、"Unknown"

    // PDF/A 合规
    PDFAConformance string // "A"、"B"、"U"
    PDFAPart        int    // 1、2、3

    // 自定义属性
    Custom map[string]string
}

func (gp *GoPdf) SetXMPMetadata(meta XMPMetadata)
func (gp *GoPdf) GetXMPMetadata() *XMPMetadata
```

在 PDF 目录中嵌入 XMP 元数据流。支持 Dublin Core、XMP 基本属性、PDF 特定属性和 PDF/A 合规属性。

---

## 增量保存

```go
func (gp *GoPdf) IncrementalSave(originalData []byte, modifiedIndices []int) ([]byte, error)
func (gp *GoPdf) WriteIncrementalPdf(pdfPath string, originalData []byte, modifiedIndices []int) error
```

将修改过的对象作为增量更新追加到原始 PDF 数据之后。如果 `modifiedIndices` 为 nil，则写入所有对象。对于大文档，这比完全重写快得多。

---

## 文档克隆

```go
func (gp *GoPdf) Clone() (*GoPdf, error)
```

通过序列化和重新导入创建 GoPdf 实例的深拷贝。克隆完全独立——对一个的修改不会影响另一个。页眉/页脚回调不会被克隆。
