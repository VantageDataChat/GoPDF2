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

`PageSizeA3`、`PageSizeA4`、`PageSizeA5`、`PageSizeLetter`、`PageSizeLegal`

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
