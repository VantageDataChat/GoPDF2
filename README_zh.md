# GoPDF2

[![Go Reference](https://pkg.go.dev/badge/github.com/VantageDataChat/GoPDF2.svg)](https://pkg.go.dev/github.com/VantageDataChat/GoPDF2)

**[English](README.md) | [中文](README_zh.md)**

GoPDF2 是一个用 Go 编写的 PDF 生成库。基于 [gopdf](https://github.com/signintech/gopdf) 开发，新增了 HTML 渲染等功能。

需要 Go 1.13+。

## 功能特性

- Unicode 子集字体嵌入（中文、日文、韩文等），自动子集化以最小化文件体积
- **HTML 转 PDF 渲染** — 通过 `InsertHTMLBox` 支持 `<b>`、`<i>`、`<u>`、`<p>`、`<h1>`–`<h6>`、`<font>`、`<span style>`、`<img>`、`<ul>`/`<ol>`、`<hr>`、`<center>`、`<a>`、`<blockquote>` 等标签
- 绘制线条、椭圆、矩形（支持圆角）、曲线、多边形
- 插入图片（JPEG、PNG），支持遮罩、裁剪、旋转、透明度
- PDF 密码保护
- 字体字距调整（Kerning）
- 导入已有 PDF 页面
- 表格布局
- 页眉/页脚回调
- Trim-box 支持
- 占位文本（先占位后填充）

## 安装

```bash
go get -u github.com/VantageDataChat/GoPDF2
```

## 快速开始

### 输出文本

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

    pdf.Cell(nil, "你好，世界！")
    pdf.WritePdf("hello.pdf")
}
```

### InsertHTMLBox — 将 HTML 渲染到 PDF

函数签名：

```go
func (gp *GoPdf) InsertHTMLBox(x, y, w, h float64, htmlStr string, opt HTMLBoxOption) (float64, error)
```

示例：

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

    // 加载字体 — 仅实际使用的字形会被嵌入（子集化）
    if err := pdf.AddTTFFont("regular", "NotoSansSC-Regular.ttf"); err != nil {
        log.Fatal(err)
    }
    if err := pdf.AddTTFFontWithOption("bold", "NotoSansSC-Bold.ttf", gopdf.TtfOption{Style: gopdf.Bold}); err != nil {
        log.Fatal(err)
    }

    html := `
    <h2>GoPDF2 HTML 渲染示例</h2>
    <p>支持<b>加粗</b>、<i>斜体</i>、<u>下划线</u>和
       <font color="#e74c3c">彩色文字</font>。</p>
    <ul>
        <li>自动换行与分段</li>
        <li>有序/无序列表</li>
        <li>图片插入</li>
    </ul>
    <hr/>
    <p style="font-size:10pt; color:gray">
        字体子集嵌入 — 仅包含文档中实际使用的字符，有效控制文件大小。
    </p>`

    endY, err := pdf.InsertHTMLBox(40, 40, 515, 750, html, gopdf.HTMLBoxOption{
        DefaultFontFamily: "regular",
        DefaultFontSize:   12,
        BoldFontFamily:    "bold",
    })
    if err != nil {
        log.Fatal(err)
    }
    _ = endY // 渲染结束后的 Y 坐标

    pdf.WritePdf("html_example.pdf")
}
```

#### HTMLBoxOption 配置项

| 字段 | 类型 | 说明 |
|---|---|---|
| `DefaultFontFamily` | `string` | HTML 未指定字体时使用的默认字体族（必填） |
| `DefaultFontSize` | `float64` | 默认字号，单位为磅（默认 12） |
| `DefaultColor` | `[3]uint8` | 默认文字颜色 `{R, G, B}` |
| `LineSpacing` | `float64` | 额外行间距，单位为文档单位 |
| `BoldFontFamily` | `string` | `<b>` / `<strong>` 使用的字体族 |
| `ItalicFontFamily` | `string` | `<i>` / `<em>` 使用的字体族 |
| `BoldItalicFontFamily` | `string` | 粗斜体使用的字体族 |

#### 支持的 HTML 标签

| 标签 | 效果 |
|---|---|
| `<b>`、`<strong>` | 加粗 |
| `<i>`、`<em>` | 斜体 |
| `<u>` | 下划线 |
| `<br>` | 换行 |
| `<p>`、`<div>` | 段落 |
| `<h1>` – `<h6>` | 标题 |
| `<font color="..." size="..." face="...">` | 字体样式 |
| `<span style="...">` | 内联 CSS（color、font-size、font-family、font-weight、font-style、text-decoration、text-align） |
| `<img src="..." width="..." height="...">` | 图片（本地文件路径） |
| `<ul>`、`<ol>`、`<li>` | 列表 |
| `<hr>` | 水平线 |
| `<center>` | 居中文本 |
| `<a href="...">` | 链接（蓝色下划线文字） |
| `<blockquote>` | 缩进引用块 |
| `<sub>`、`<sup>` | 下标/上标 |

### 字体子集化与文件大小控制

GoPDF2 默认使用**字体子集化**。调用 `AddTTFFont` 时会解析完整的 TTF 文件，但输出 PDF 时仅嵌入文档中实际使用的字形。这对中日韩字体尤为重要 — 一个 CJK 字体文件可能有 10-20 MB，但生成的 PDF 只会包含所用字符对应的几 KB 数据。

无需额外配置，子集化是自动完成的。

### 文字颜色

```go
// RGB 模式
pdf.SetTextColor(156, 197, 140)
pdf.Cell(nil, "彩色文字")

// CMYK 模式
pdf.SetTextColorCMYK(0, 6, 14, 0)
pdf.Cell(nil, "CMYK 文字")
```

### 插入图片

```go
pdf.Image("gopher.jpg", 200, 50, nil)
```

### 链接

```go
pdf.SetXY(30, 40)
pdf.Text("链接到 example.com")
pdf.AddExternalLink("http://example.com/", 27.5, 28, 125, 15)
```

### 页眉与页脚

```go
pdf.AddHeader(func() {
    pdf.SetY(5)
    pdf.Cell(nil, "页眉")
})
pdf.AddFooter(func() {
    pdf.SetY(825)
    pdf.Cell(nil, "页脚")
})
```

### 绘图

```go
// 线条
pdf.SetLineWidth(2)
pdf.SetLineType("dashed")
pdf.Line(10, 30, 585, 30)

// 椭圆
pdf.Oval(100, 200, 500, 500)

// 多边形
pdf.SetStrokeColor(255, 0, 0)
pdf.SetFillColor(0, 255, 0)
pdf.Polygon([]gopdf.Point{{X: 10, Y: 30}, {X: 585, Y: 200}, {X: 585, Y: 250}}, "DF")

// 圆角矩形
pdf.Rectangle(196.6, 336.8, 398.3, 379.3, "DF", 3, 10)
```

### 旋转

```go
pdf.Rotate(270.0, 100.0, 100.0)
pdf.Text("旋转文字")
pdf.RotateReset()
```

### 透明度

```go
pdf.SetTransparency(gopdf.Transparency{Alpha: 0.5, BlendModeType: ""})
```

### 密码保护

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

### 导入已有 PDF

```go
tpl := pdf.ImportPage("existing.pdf", 1, "/MediaBox")
pdf.UseImportedTemplate(tpl, 50, 100, 400, 0)
```

### 表格

```go
table := pdf.NewTableLayout(10, 10, 25, 5)
table.AddColumn("编号", 50, "left")
table.AddColumn("描述", 200, "left")
table.AddRow([]string{"001", "产品 A"})
table.DrawTable()
```

### 占位文本

```go
pdf.PlaceHolderText("total", 30)
// ... 创建完所有页面后 ...
pdf.FillInPlaceHoldText("total", "5", gopdf.Left)
```

## API 参考

参见 [docs/API.md](docs/API.md)（English）或 [docs/API_zh.md](docs/API_zh.md)（中文）。

## 许可证

MIT
