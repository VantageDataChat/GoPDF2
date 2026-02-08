# GoPDF2 与 PyMuPDF 功能对比

**[English](COMPARISON.md) | [中文](COMPARISON_zh.md)**

本文档详细对比 GoPDF2（纯 Go 实现）与 PyMuPDF（基于 MuPDF C 库的 Python 绑定）的功能差异。

> GoPDF2 是纯 Go 实现，无 CGO 依赖；PyMuPDF 通过 CGO 绑定 MuPDF C 库，可调用底层渲染引擎。
> 两者定位不同：GoPDF2 专注于 PDF 生成与编辑，PyMuPDF 侧重于 PDF 解析、提取与渲染。

---

## 功能对比总览

| 功能类别 | GoPDF2 | PyMuPDF | 说明 |
|---------|--------|---------|------|
| **纯 Go / 无 CGO** | ✅ | ❌ | GoPDF2 无需 C 编译器，交叉编译友好 |
| **PDF 生成** | ✅ | ✅ | 两者均可从零创建 PDF |
| **打开/修改已有 PDF** | ✅ | ✅ | |
| **HTML 渲染到 PDF** | ✅ | ✅ | GoPDF2 内置 HTML 解析器；PyMuPDF 通过 Story 类 |
| **文本提取** | ✅ | ✅ | GoPDF2 内置纯 Go 内容流解析器 |
| **页面渲染（光栅化）** | ❌ | ✅ | 需要 MuPDF 渲染引擎 |
| **OCR 文字识别** | ❌ | ✅ | 需要 Tesseract + MuPDF |
| **SVG/HTML/EPUB 转换** | ❌ | ✅ | 需要 MuPDF 多格式支持 |
| **图片提取** | ✅ | ✅ | GoPDF2 内置纯 Go 图片提取 |
| **表单字段 (AcroForm)** | ✅ | ✅ | GoPDF2 支持文本、复选框、下拉框、单选、按钮、签名字段 |
| **数字签名** | ✅ | ✅ | GoPDF2 支持 PKCS#7 签名与验证 |

---

## 详细功能对比

### 1. 文档生命周期

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 创建空白 PDF | `Start(Config)` | `Document()` |
| 从文件打开 | `OpenPDF(path)` | `Document(path)` |
| 从内存打开 | `OpenPDFFromBytes(data)` | `Document(stream=data)` |
| 从流打开 | `OpenPDFFromStream(rs)` | `Document(stream=data)` |
| 保存到文件 | `WritePdf(path)` | `save(path)` |
| 保存到内存 | `GetBytesPdf()` | `tobytes()` |
| 增量保存 | ✅ `IncrementalSave()` | ✅ `saveIncr()` |
| 文档克隆 | ✅ `Clone()` | ❌（需手动序列化） |
| 关闭文档 | 自动 GC | `close()` |
| 打开加密 PDF | ❌ | ✅ `authenticate()` |
| 多格式支持 (XPS/EPUB/HTML) | ❌ | ✅ |

### 2. 页面管理

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 添加新页面 | ✅ `AddPage()` | ✅ `new_page()` |
| 删除页面 | ✅ `DeletePage(n)` | ✅ `delete_page(n)` |
| 批量删除页面 | ❌ | ✅ `delete_pages()` |
| 复制页面（引用） | ✅ `CopyPage(n)` | ✅ `copy_page()` |
| 完整复制页面 | ✅ `CopyPage(n)` | ✅ `fullcopy_page()` |
| 移动页面 | ❌ | ✅ `move_page()` |
| 页面重排/选择 | ✅ `SelectPages()` | ✅ `select()` |
| 提取页面到新文档 | ✅ `ExtractPages()` | ✅ `select()` + `save()` |
| 合并多个 PDF | ✅ `MergePages()` | ✅ `insert_pdf()` |
| 插入其他 PDF 页面 | ✅ `ImportPage()` | ✅ `insert_pdf()` |
| 插入任意格式文档 | ❌ | ✅ `insert_file()` |
| 页面旋转 | ✅ `SetPageRotation()` | ✅ `set_rotation()` |
| 获取页面尺寸 | ✅ `GetPageSize()` | ✅ `page.rect` |
| 获取所有页面尺寸 | ✅ `GetAllPageSizes()` | 需遍历 |
| 获取源 PDF 页数 | ✅ `GetSourcePDFPageCount()` | ✅ `page_count` |
| 页面裁切框 | ❌ | ✅ `page_cropbox()` |
| 章节导航 (EPUB) | ❌ | ✅ `chapter_page_count()` 等 |

### 3. 文本与字体

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 写入文本 | ✅ `Cell()` / `Text()` | ✅ `insert_text()` |
| 多行文本 | ✅ `MultiCell()` | ✅ `insert_textbox()` |
| 文本测量 | ✅ `MeasureTextWidth()` | ✅ `get_text_length()` |
| 自动换行 | ✅ `SplitTextWithWordWrap()` | ✅ 内置 |
| TTF 字体嵌入 | ✅ 自动子集化 | ✅ |
| 字体子集化 | ✅ 自动 | ✅ `subset_fonts()` |
| 字距调整 (Kerning) | ✅ | ✅ |
| 文本提取 | ✅ `ExtractPageText()` | ✅ `get_text()` |
| 文本搜索 | ❌ | ✅ `search_for()` |
| 文本块/字典提取 | ✅ `ExtractTextFromPage()` | ✅ `get_text("dict")` |
| 字体信息查询 | ✅ `GetFonts()` | ✅ `get_page_fonts()` |
| 字体提取 | ❌ | ✅ `extract_font()` |
| CJK 字体支持 | ✅ | ✅ |

### 4. 绘图

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 线条 | ✅ `Line()` | ✅ `draw_line()` |
| 矩形 | ✅ `Rectangle()` | ✅ `draw_rect()` |
| 圆角矩形 | ✅ | ❌（需手动绘制） |
| 椭圆/圆 | ✅ `Oval()` | ✅ `draw_circle()` / `draw_oval()` |
| 多边形 | ✅ `Polygon()` | ✅ `draw_polygon()` |
| 贝塞尔曲线 | ✅ `Curve()` | ✅ `draw_bezier()` |
| 折线 | ❌ | ✅ `draw_polyline()` |
| 扇形 | ❌ | ✅ `draw_sector()` |
| 虚线/线型 | ✅ `SetLineType()` | ✅ |
| 填充色/描边色 | ✅ | ✅ |
| 旋转 | ✅ `Rotate()` | ✅ |
| 透明度 | ✅ `SetTransparency()` | ✅ |
| 图形状态保存/恢复 | ✅ `SaveGraphicsState()` | ✅ |

### 5. 图片

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 插入图片 (JPEG/PNG) | ✅ `Image()` | ✅ `insert_image()` |
| 从内存插入图片 | ✅ `ImageFrom()` | ✅ |
| 图片遮罩/裁剪 | ✅ | ✅ |
| 图片旋转 | ✅ | ✅ |
| 图片透明度 | ✅ | ✅ |
| 图片提取 | ✅ `ExtractImagesFromPage()` | ✅ `extract_image()` |
| 图片信息查询 | ✅ `ExtractImagesFromPage()` | ✅ `get_page_images()` |
| 图片删除 | ❌ | ✅ `delete_image()` |
| 图片重压缩 | ❌ | ✅ `rewrite_images()` |
| SVG 插入 | ❌ | ✅ |
| Pixmap 渲染 | ❌ | ✅ `get_pixmap()` |

### 6. 注释 (Annotations)

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 文本注释（便签） | ✅ | ✅ `add_text_annot()` |
| 高亮 | ✅ | ✅ `add_highlight_annot()` |
| 下划线 | ✅ | ✅ `add_underline_annot()` |
| 删除线 | ✅ | ✅ `add_strikeout_annot()` |
| 矩形注释 | ✅ | ✅ `add_rect_annot()` |
| 圆形注释 | ✅ | ✅ `add_circle_annot()` |
| 自由文本 | ✅ | ✅ `add_freetext_annot()` |
| 墨迹注释 | ❌ | ✅ `add_ink_annot()` |
| 折线注释 | ❌ | ✅ `add_polyline_annot()` |
| 多边形注释 | ❌ | ✅ `add_polygon_annot()` |
| 线条注释 | ❌ | ✅ `add_line_annot()` |
| 图章注释 | ❌ | ✅ `add_stamp_annot()` |
| 波浪线注释 | ❌ | ✅ `add_squiggly_annot()` |
| 插入符注释 | ❌ | ✅ `add_caret_annot()` |
| 文件附件注释 | ❌ | ✅ `add_file_annot()` |
| 涂改注释 (Redaction) | ❌ | ✅ `add_redact_annot()` |
| 删除注释 | ❌ | ✅ `delete_annot()` |
| 修改注释属性 | ❌ | ✅ Annot 类方法 |
| 注释遍历 | ❌ | ✅ `annots()` |
| 应用涂改 | ❌ | ✅ `apply_redactions()` |

### 7. 水印

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 文字水印 | ✅ `AddWatermarkText()` | 需手动实现 |
| 图片水印 | ✅ `AddWatermarkImage()` | 需手动实现 |
| 全页面水印 | ✅ `AddWatermarkTextAllPages()` | 需手动实现 |
| 平铺水印 | ✅ `Repeat: true` | 需手动实现 |
| 水印透明度/旋转 | ✅ | 需手动实现 |

> GoPDF2 提供开箱即用的水印 API，PyMuPDF 需要通过绘图 API 手动组合实现。

### 8. 书签/目录 (TOC/Outlines)

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 获取目录 | ✅ `GetTOC()` | ✅ `get_toc()` |
| 设置目录 | ✅ `SetTOC()` | ✅ `set_toc()` |
| 层级化目录 | ✅ | ✅ |
| 添加单个书签 | ✅ `AddOutline()` | ✅ |
| 修改单个书签 | ❌ | ✅ `set_toc_item()` |
| 删除单个书签 | ❌ | ✅ `del_toc_item()` |
| 书签颜色/粗体/斜体 | ❌ | ✅ |
| 书签折叠控制 | ❌ | ✅ `collapse` 参数 |

### 9. 元数据

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 标准元数据 (Info) | ✅ `SetInfo()` | ✅ `set_metadata()` |
| XMP 元数据 | ✅ `SetXMPMetadata()` | ✅ `set_xml_metadata()` |
| 获取 XMP 元数据 | ✅ `GetXMPMetadata()` | ✅ `get_xml_metadata()` |
| PDF/A 合规 | ✅ XMP 中设置 | ✅ |
| Dublin Core | ✅ | ✅ |

### 10. 嵌入文件

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 添加嵌入文件 | ✅ `AddEmbeddedFile()` | ✅ `embfile_add()` |
| 获取嵌入文件 | ❌ | ✅ `embfile_get()` |
| 删除嵌入文件 | ❌ | ✅ `embfile_del()` |
| 更新嵌入文件 | ❌ | ✅ `embfile_upd()` |
| 嵌入文件信息 | ❌ | ✅ `embfile_info()` |
| 嵌入文件列表 | ❌ | ✅ `embfile_names()` |
| 嵌入文件数量 | ❌ | ✅ `embfile_count()` |

### 11. 可选内容组 (OCG/图层)

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 添加 OCG | ✅ `AddOCG()` | ✅ `add_ocg()` |
| 获取所有 OCG | ✅ `GetOCGs()` | ✅ `get_ocgs()` |
| 图层配置 | ❌ | ✅ `add_layer()` / `get_layers()` |
| 切换图层 | ❌ | ✅ `switch_layer()` |
| OCMD（成员字典） | ❌ | ✅ `set_ocmd()` / `get_ocmd()` |
| 图层 UI 配置 | ❌ | ✅ `layer_ui_configs()` |
| 批量设置 OCG 状态 | ❌ | ✅ `set_layer()` |

### 12. 页面布局与显示

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 页面布局 | ✅ `SetPageLayout()` | ✅ `set_pagelayout()` |
| 页面模式 | ✅ `SetPageMode()` | ✅ `set_pagemode()` |
| MarkInfo | ❌ | ✅ `set_markinfo()` |
| 页面标签 | ✅ `SetPageLabels()` | ✅ `set_page_labels()` |
| 获取页面标签 | ✅ `GetPageLabels()` | ✅ `get_page_labels()` |
| 按标签查找页面 | ❌ | ✅ `get_page_numbers()` |

### 13. 安全与加密

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 密码保护 | ✅ `PDFProtectionConfig` | ✅ `save(encryption=...)` |
| 权限控制 | ✅ | ✅ |
| 打开加密 PDF | ❌ | ✅ `authenticate()` |
| 加密方法选择 | 有限 | ✅ 多种加密标准 |
| 数字签名 | ✅ `SignPDF()` / `VerifySignature()` | ✅ `get_sigflags()` |

### 14. 文档清洗与优化

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 文档清洗 | ✅ `Scrub()` | ✅ `scrub()` |
| 垃圾回收 | ✅ `GarbageCollect()` | ✅ `save(garbage=N)` |
| 对象去重 | ✅ `GCDedup` | ✅ `save(garbage=3)` |
| 文档统计 | ✅ `GetDocumentStats()` | 需手动统计 |
| 线性化 (Web 优化) | ❌ | ✅ `save(linear=True)` |
| 内容流清理 | ❌ | ✅ `save(clean=True)` |
| 图片重压缩 | ❌ | ✅ `rewrite_images()` |
| 颜色空间转换 | ❌ | ✅ `recolor()` |

### 15. 内容元素操作

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 列出页面元素 | ✅ `GetPageElements()` | ❌（需解析内容流） |
| 按类型筛选元素 | ✅ `GetPageElementsByType()` | ❌ |
| 删除单个元素 | ✅ `DeleteElement()` | ❌ |
| 按类型删除元素 | ✅ `DeleteElementsByType()` | ❌ |
| 区域删除元素 | ✅ `DeleteElementsInRect()` | ❌ |
| 清空页面 | ✅ `ClearPage()` | ❌（需重写内容流） |
| 修改文本元素 | ✅ `ModifyTextElement()` | ❌ |
| 修改元素位置 | ✅ `ModifyElementPosition()` | ❌ |
| 插入新元素 | ✅ `InsertLineElement()` 等 | ❌ |

> GoPDF2 的内容元素 CRUD 是独有功能，因为 GoPDF2 在内存中维护结构化的内容缓存，可以直接操作单个元素。PyMuPDF 的内容流是扁平的二进制数据，无法直接进行元素级操作。

### 16. PDF 底层操作

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| PDF 版本控制 | ✅ `SetPDFVersion()` | 通过 header 控制 |
| 类型化对象 ID | ✅ `ObjID` | ✅ xref 系统 |
| 对象数量统计 | ✅ `GetObjectCount()` | ✅ `xref_length()` |
| 读取对象定义 | ❌ | ✅ `xref_object()` |
| 修改对象定义 | ❌ | ✅ `update_object()` |
| 读取/写入字典键 | ❌ | ✅ `xref_get_key()` / `xref_set_key()` |
| 读取/写入流数据 | ❌ | ✅ `xref_stream()` / `update_stream()` |
| 复制对象 | ❌ | ✅ `xref_copy()` |
| PDF Catalog 访问 | ❌ | ✅ `pdf_catalog()` |
| PDF Trailer 访问 | ❌ | ✅ `pdf_trailer()` |

### 17. 日志/撤销 (Journalling)

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 启用日志 | ❌ | ✅ `journal_enable()` |
| 撤销/重做 | ❌ | ✅ `journal_undo()` / `journal_redo()` |
| 保存/加载日志 | ❌ | ✅ `journal_save()` / `journal_load()` |
| 操作命名 | ❌ | ✅ `journal_start_op()` |

### 18. 表单字段 (Widgets/AcroForm)

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| 添加表单字段 | ✅ `AddFormField()` | ✅ `add_widget()` |
| 添加文本字段 | ✅ `AddTextField()` | ✅ `add_widget()` |
| 添加复选框 | ✅ `AddCheckbox()` | ✅ `add_widget()` |
| 添加下拉框 | ✅ `AddDropdown()` | ✅ `add_widget()` |
| 添加签名字段 | ✅ `AddSignatureField()` | ✅ `add_widget()` |
| 获取表单字段 | ✅ `GetFormFields()` | ✅ Widget 类 |
| 删除表单字段 | ❌ | ✅ `delete_widget()` |
| 修改表单值 | ❌ | ✅ Widget 类 |
| 表单 PDF 检测 | ❌ | ✅ `is_form_pdf` |
| 固化注释/表单 | ❌ | ✅ `bake()` |

### 19. 其他功能

| 功能 | GoPDF2 | PyMuPDF |
|------|--------|---------|
| HTML 渲染 | ✅ `InsertHTMLBox()` | ✅ Story 类 |
| 表格布局 | ✅ `NewTableLayout()` | ❌（需手动实现） |
| 页眉/页脚回调 | ✅ `AddHeader()` / `AddFooter()` | ❌（需手动实现） |
| 占位文本 | ✅ `PlaceHolderText()` | ❌ |
| 纸张尺寸查询 | ✅ `PaperSize()` | ✅ `paper_size()` |
| 几何工具 | ✅ `RectFrom` / `Matrix` | ✅ `Rect` / `Matrix` |
| 链接 | ✅ `AddExternalLink()` | ✅ `insert_link()` |
| 锚点 | ✅ `SetAnchor()` | ✅ |
| 字体容器复用 | ✅ `FontContainer` | ❌ |
| 压缩级别控制 | ✅ `SetCompressLevel()` | ✅ `save(deflate=True)` |
| Trim-box | ✅ | ✅ |

---

## GoPDF2 独有优势

1. **纯 Go 实现** — 无 CGO 依赖，编译简单，交叉编译友好，适合容器化部署
2. **内容元素 CRUD** — 可直接操作页面上的单个内容元素（文本、图片、线条等），这是 PyMuPDF 无法做到的
3. **开箱即用的水印 API** — 文字/图片水印、平铺、全页面，一行代码搞定
4. **表格布局** — 内置表格生成器
5. **页眉/页脚回调** — 自动在每页执行
6. **占位文本** — 先占位后填充的模式
7. **字体容器复用** — 多文档共享字体解析结果
8. **文档克隆** — 一行代码深拷贝
9. **文档统计** — 一行代码获取文档结构概览

## PyMuPDF 独有优势

1. **页面渲染** — 可将 PDF 页面渲染为图片 (Pixmap)
2. **文本搜索** — 在页面中搜索文本并返回位置
3. **OCR** — 结合 Tesseract 进行文字识别
4. **多格式支持** — 可打开 XPS、EPUB、HTML、SVG、图片等格式
5. **日志/撤销** — 操作可撤销/重做
6. **涂改注释** — 安全地永久删除敏感内容
7. **PDF 底层操作** — 直接读写 PDF 对象、字典键、流数据
8. **图片重压缩** — 批量重压缩嵌入图片
9. **线性化** — 生成 Web 优化的 PDF
10. **更多注释类型** — 墨迹、折线、多边形、图章、波浪线、插入符、文件附件、涂改等

---

## 适用场景建议

| 场景 | 推荐 |
|------|------|
| Go 项目中生成 PDF 报告 | GoPDF2 |
| 在已有 PDF 上叠加内容 | GoPDF2 |
| 需要精细操作页面元素 | GoPDF2 |
| 批量添加水印 | GoPDF2 |
| 容器化/无 C 编译器环境 | GoPDF2 |
| 从 PDF 提取文本/图片 | GoPDF2（基础）/ PyMuPDF（高级） |
| PDF 转图片 | PyMuPDF |
| OCR 文字识别 | PyMuPDF |
| 表单填写与处理 | GoPDF2（创建）/ PyMuPDF（完整读写） |
| 多格式文档转换 | PyMuPDF |
| 需要底层 PDF 对象操作 | PyMuPDF |
| 涂改/安全删除敏感内容 | PyMuPDF |

---

## 未来可扩展方向

以下功能在纯 Go 中理论上可实现，但复杂度较高：

| 功能 | 难度 | 说明 |
|------|------|------|
| 更多注释类型 | 中 | 墨迹、折线、多边形、图章等 |
| MarkInfo | 低 | 简单的字典写入 |
| 嵌入文件读取/删除 | 中 | 需要解析 Names 字典 |
| ~~表单字段 (AcroForm)~~ | ~~高~~ | ✅ 已实现 — `AddFormField()`、`AddTextField()`、`AddCheckbox()`、`AddDropdown()` |
| ~~文本提取~~ | ~~极高~~ | ✅ 已实现 — `ExtractTextFromPage()`、`ExtractPageText()` |
| ~~图片提取~~ | ~~极高~~ | ✅ 已实现 — `ExtractImagesFromPage()`、`ExtractImagesFromAllPages()` |
| ~~数字签名~~ | ~~高~~ | ✅ 已实现 — `SignPDF()`、`VerifySignature()` |
| 页面渲染 | 不可行 | 需要完整的渲染引擎 |
| OCR | 不可行 | 需要 Tesseract 或类似引擎 |
| 日志/撤销 | 高 | 需要实现操作记录与回放 |
| 线性化 | 高 | 需要重新组织 PDF 文件结构 |
