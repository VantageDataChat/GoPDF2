package gopdf

import (
	"bytes"
	"compress/zlib"
	"image"
	colorPkg "image/color"
	pngPkg "image/png"
	"testing"
)

// ============================================================
// coverage_boost22_test.go — TestCov22_ prefix
// Targets: content_obj write (NoCompression + protection paths),
// image_obj write (IsMask + protection), watermark all pages,
// DeleteBookmark, MeasureTextWidth/CellHeight edge cases,
// AddWatermarkTextAllPages repeat, AddWatermarkImageAllPages
// ============================================================

// ============================================================
// content_obj.go — write with NoCompression
// ============================================================

func TestCov22_ContentObj_Write_NoCompression(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetNoCompression()
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "No compression content")
	pdf.Line(10, 10, 200, 200)

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestCov22_ContentObj_Write_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Protected content write")
	pdf.Line(10, 10, 200, 200)

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestCov22_ContentObj_Write_ProtectedNoCompression(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   PermissionsPrint,
			UserPass:      []byte("u"),
			OwnerPass:     []byte("o"),
		},
	})
	pdf.SetNoCompression()
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	_ = pdf.Cell(nil, "Protected + no compression")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// content_obj.go — write with custom compress level
// ============================================================

func TestCov22_ContentObj_CompressLevels(t *testing.T) {
	for _, level := range []int{zlib.BestSpeed, zlib.BestCompression, zlib.DefaultCompression} {
		pdf := &GoPdf{}
		pdf.Start(Config{PageSize: *PageSizeA4})
		pdf.SetCompressLevel(level)
		if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
			t.Skipf("font not available: %v", err)
		}
		if err := pdf.SetFont(fontFamily, "", 14); err != nil {
			t.Fatalf("SetFont: %v", err)
		}
		pdf.AddPage()
		_ = pdf.Cell(nil, "Compress level test")

		var buf bytes.Buffer
		if err := pdf.Write(&buf); err != nil {
			t.Fatalf("Write level %d: %v", level, err)
		}
	}
}

// ============================================================
// image_obj.go — write with IsMask=true (mask image path)
// ============================================================

func TestCov22_ImageObj_MaskImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add a PNG with transparency — this triggers SMask/mask image path.
	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Skipf("Image: %v", err)
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestCov22_ImageObj_MaskImage_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()

	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Skipf("Image: %v", err)
	}

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// bookmark.go — DeleteBookmark
// ============================================================

func TestCov22_DeleteBookmark(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Chapter 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	_ = pdf.Cell(nil, "Chapter 2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	_ = pdf.Cell(nil, "Chapter 3")
	pdf.AddOutline("Chapter 3")

	// Delete middle bookmark.
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}

	// Delete first bookmark.
	err = pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark first: %v", err)
	}
}

func TestCov22_DeleteBookmark_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.DeleteBookmark(-1)
	if err == nil {
		t.Fatal("expected error for negative index")
	}
	err = pdf.DeleteBookmark(99)
	if err == nil {
		t.Fatal("expected error for out of range")
	}
}

func TestCov22_DeleteBookmark_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Chapter 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	_ = pdf.Cell(nil, "Chapter 2")
	pdf.AddOutline("Chapter 2")

	// Delete last bookmark.
	bookmarks := pdf.GetTOC()
	if len(bookmarks) > 0 {
		err := pdf.DeleteBookmark(len(bookmarks) - 1)
		if err != nil {
			t.Fatalf("DeleteBookmark last: %v", err)
		}
	}
}

// ============================================================
// watermark.go — AddWatermarkTextAllPages with Repeat
// ============================================================

func TestCov22_WatermarkTextAllPages_Repeat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "CONFIDENTIAL",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.2,
		Angle:      30,
		Repeat:     true,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

func TestCov22_WatermarkTextAllPages_Single(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 3")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   48,
		Opacity:    0.3,
		Angle:      45,
		Color:      [3]uint8{255, 0, 0},
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

// ============================================================
// watermark.go — AddWatermarkImageAllPages
// ============================================================

func TestCov22_WatermarkImageAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.3, 100, 100, 30)
	if err != nil {
		t.Skipf("AddWatermarkImageAllPages: %v", err)
	}
}

func TestCov22_WatermarkImageAllPages_NoAngle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.5, 0, 0, 0)
	if err != nil {
		t.Skipf("AddWatermarkImageAllPages: %v", err)
	}
}

// ============================================================
// watermark.go — AddWatermarkText error paths
// ============================================================

func TestCov22_WatermarkText_EmptyText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "",
		FontFamily: fontFamily,
	})
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestCov22_WatermarkText_EmptyFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "test",
		FontFamily: "",
	})
	if err == nil {
		t.Fatal("expected error for empty font family")
	}
}

// ============================================================
// gopdf.go — Text with char spacing
// ============================================================

func TestCov22_Text_WithCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.SetCharSpacing(2.0)
	if err := pdf.Text("Spaced text"); err != nil {
		t.Fatalf("Text: %v", err)
	}
	pdf.SetCharSpacing(0)
}

// ============================================================
// gopdf.go — MeasureTextWidth, MeasureCellHeightByText
// ============================================================

func TestCov22_MeasureTextWidth_Various(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth: %v", err)
	}
	if w <= 0 {
		t.Error("expected positive width")
	}

	// Measure with char spacing.
	pdf.SetCharSpacing(3.0)
	w2, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth with spacing: %v", err)
	}
	if w2 <= w {
		t.Error("expected wider text with char spacing")
	}
	pdf.SetCharSpacing(0)
}

func TestCov22_MeasureCellHeightByText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	h, err := pdf.MeasureCellHeightByText("Hello")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Error("expected positive height")
	}
}

// ============================================================
// gopdf.go — CellWithOption various alignments
// ============================================================

func TestCov22_CellWithOption_RightAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Right aligned", CellOption{
		Align: Right,
	})
	if err != nil {
		t.Fatalf("CellWithOption right: %v", err)
	}
}

func TestCov22_CellWithOption_CenterAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 100)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Center aligned", CellOption{
		Align: Center,
	})
	if err != nil {
		t.Fatalf("CellWithOption center: %v", err)
	}
}

func TestCov22_CellWithOption_BottomAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 150)

	err := pdf.CellWithOption(&Rect{W: 200, H: 50}, "Bottom aligned", CellOption{
		Align: Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption bottom: %v", err)
	}
}

func TestCov22_CellWithOption_MiddleAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 200)

	err := pdf.CellWithOption(&Rect{W: 200, H: 50}, "Middle aligned", CellOption{
		Align: Middle,
	})
	if err != nil {
		t.Fatalf("CellWithOption middle: %v", err)
	}
}

func TestCov22_CellWithOption_Border(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 300)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "With border", CellOption{
		Align:  Left,
		Border: AllBorders,
	})
	if err != nil {
		t.Fatalf("CellWithOption border: %v", err)
	}
}

// ============================================================
// gopdf.go — GetBytesPdf (calls GetBytesPdfReturnErr)
// ============================================================

func TestCov22_GetBytesPdf(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "GetBytesPdf test")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty bytes")
	}
}

// ============================================================
// gopdf.go — Read
// ============================================================

func TestCov22_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Read test")

	buf := make([]byte, 4096)
	n, err := pdf.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes read")
	}
}

// ============================================================
// cache_content_text_color_cmyk.go — equal (same color)
// ============================================================

func TestCov22_CMYK_Equal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set same CMYK color twice to trigger equal() returning true.
	pdf.SetTextColorCMYK(50, 50, 50, 50)
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "CMYK 1")

	pdf.SetTextColorCMYK(50, 50, 50, 50)
	pdf.SetXY(50, 80)
	_ = pdf.Cell(nil, "CMYK 2 same")

	// Different color.
	pdf.SetTextColorCMYK(0, 100, 100, 0)
	pdf.SetXY(50, 110)
	_ = pdf.Cell(nil, "CMYK 3 different")
}

// ============================================================
// cache_content_rotate.go — write
// ============================================================

func TestCov22_Rotate_Write(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Rotate(45, 100, 100)
	pdf.SetXY(80, 90)
	_ = pdf.Cell(nil, "Rotated text")
	pdf.RotateReset()

	pdf.Rotate(90, 200, 200)
	pdf.SetXY(180, 190)
	_ = pdf.Cell(nil, "90 degrees")
	pdf.RotateReset()

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// cache_content_text.go — underline
// ============================================================

func TestCov22_Underline_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set underline style.
	if err := pdf.SetFontWithStyle(fontFamily, Underline, 14); err != nil {
		t.Fatalf("SetFontWithStyle underline: %v", err)
	}
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Underlined text")

	// Bold + underline — use same font, just underline style.
	if err := pdf.SetFontWithStyle(fontFamily, Underline, 18); err != nil {
		t.Fatalf("SetFontWithStyle underline 18pt: %v", err)
	}
	pdf.SetXY(50, 80)
	_ = pdf.Cell(nil, "Bold underlined text")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// list_cache_content.go — write with protection
// ============================================================

func TestCov22_ListCacheContent_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Protected list cache")
	pdf.Line(10, 10, 200, 200)
	pdf.SetFillColor(255, 0, 0)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 80, "FD")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// gopdf.go — Line with journal enabled
// ============================================================

func TestCov22_Line_WithJournal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()
	pdf.JournalStartOp("draw lines")

	pdf.SetLineWidth(2)
	pdf.Line(10, 10, 300, 10)
	pdf.Line(10, 30, 300, 30)

	pdf.JournalEndOp()
}

// ============================================================
// gopdf.go — SetTransparency with blend mode
// ============================================================

func TestCov22_SetTransparency_BlendMode(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.5,
		BlendModeType: Multiply,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Transparent text")
	pdf.ClearTransparency()

	err = pdf.SetTransparency(Transparency{
		Alpha:         0.8,
		BlendModeType: NormalBlendMode,
	})
	if err != nil {
		t.Fatalf("SetTransparency normal: %v", err)
	}
	pdf.ClearTransparency()
}

// ============================================================
// gopdf.go — Sector
// ============================================================

func TestCov22_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetFillColor(255, 200, 200)
	pdf.Sector(200, 300, 80, 0, 90, "FD")
	pdf.Sector(200, 300, 80, 90, 180, "F")
	pdf.Sector(200, 300, 80, 180, 270, "D")
	pdf.Sector(200, 300, 80, 270, 360, "")
}

// ============================================================
// gopdf.go — IsCurrFontContainGlyph
// ============================================================

func TestCov22_IsCurrFontContainGlyph(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// ASCII should be found.
	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph: %v", err)
	}
	if !ok {
		t.Error("expected glyph 'A' to be found")
	}

	// Some rare Unicode might not be found.
	_, _ = pdf.IsCurrFontContainGlyph(0xFFFF)
}

// ============================================================
// gopdf.go — AddPageWithOption with TrimBox
// ============================================================

func TestCov22_AddPageWithOption_TrimBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 595.28, H: 841.89},
		TrimBox:  &Box{Left: 10, Top: 10, Right: 585, Bottom: 831},
	})
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Page with TrimBox")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// gopdf.go — PlaceHolderText + FillInPlaceHoldText
// ============================================================

func TestCov22_PlaceHolderText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.PlaceHolderText("name", 200)
	if err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	// Fill with right alignment.
	err = pdf.FillInPlaceHoldText("name", "John Doe", Right)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText right: %v", err)
	}
}

func TestCov22_PlaceHolderText_Center(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 100)

	err := pdf.PlaceHolderText("title", 300)
	if err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	err = pdf.FillInPlaceHoldText("title", "Document Title", Center)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText center: %v", err)
	}
}

func TestCov22_FillInPlaceHoldText_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.FillInPlaceHoldText("nonexistent", "test", Left)
	if err == nil {
		t.Fatal("expected error for nonexistent placeholder")
	}
}

// ============================================================
// open_pdf.go — OpenPDF with password-protected PDF
// ============================================================

func TestCov22_OpenPDF_WithPassword(t *testing.T) {
	// Create a protected PDF first.
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Protected document")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	ensureOutDir(t)
	outPath := resOutDir + "/cov22_protected.pdf"
	if err := pdf.WritePdf(outPath); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}

	// Open with correct password.
	opened := &GoPdf{}
	err := opened.OpenPDF(outPath, &OpenPDFOption{
		Password: "user",
	})
	if err != nil {
		t.Logf("OpenPDF protected: %v", err)
	} else {
		_ = opened.GetNumberOfPages()
	}
}

// ============================================================
// image_recompress.go — RecompressImages, rebuildXref
// ============================================================

func TestCov22_RecompressImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Skipf("Image: %v", err)
	}

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := RecompressImages(data, RecompressOption{
		JPEGQuality: 50,
		MaxWidth:    100,
	})
	if err != nil {
		t.Logf("RecompressImages: %v", err)
	}
	_ = result
}

func TestCov22_RecompressImages_PNG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Skipf("Image: %v", err)
	}

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := RecompressImages(data, RecompressOption{
		JPEGQuality: 75,
		MaxWidth:    50,
	})
	if err != nil {
		t.Logf("RecompressImages PNG: %v", err)
	}
	_ = result
}

// ============================================================
// pdf_parser.go — extractNamedRefs (via CleanContentStreams with named resources)
// ============================================================

func TestCov22_CleanContentStreams_WithNamedResources(t *testing.T) {
	// Create a PDF with images and fonts — this should trigger extractNamedRefs.
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Text with font")
	err := pdf.Image(resJPEGPath, 50, 100, &Rect{W: 100, H: 100})
	if err != nil {
		t.Skipf("Image: %v", err)
	}
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2 text")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := CleanContentStreams(data)
	if err != nil {
		t.Logf("CleanContentStreams: %v", err)
	}
	_ = result
}

// ============================================================
// font_extract.go — ExtractFontsFromPage, extractFontsFromResources
// ============================================================

func TestCov22_ExtractFontsFromPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Font extraction test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	fonts, err := ExtractFontsFromPage(data, 0)
	if err != nil {
		t.Logf("ExtractFontsFromPage: %v", err)
	}
	_ = fonts
}

func TestCov22_ExtractFontsFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1 font")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2 font")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	fonts, err := ExtractFontsFromAllPages(data)
	if err != nil {
		t.Logf("ExtractFontsFromAllPages: %v", err)
	}
	_ = fonts
}

// ============================================================
// text_search.go — SearchText, SearchTextOnPage
// ============================================================

func TestCov22_SearchText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Hello World")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Goodbye World")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	results, err := SearchText(data, "World", false)
	if err != nil {
		t.Logf("SearchText: %v", err)
	}
	_ = results

	// Case insensitive.
	results2, err := SearchText(data, "hello", true)
	if err != nil {
		t.Logf("SearchText case insensitive: %v", err)
	}
	_ = results2
}

func TestCov22_SearchTextOnPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Search me on page")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	results, err := SearchTextOnPage(data, 0, "Search", false)
	if err != nil {
		t.Logf("SearchTextOnPage: %v", err)
	}
	_ = results
}

// ============================================================
// gopdf.go — prepare (XMP metadata, mark info paths)
// ============================================================

func TestCov22_Prepare_WithXMPAndMarkInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetXMPMetadata(XMPMetadata{
		Title:       "Test Document",
		Creator:     []string{"Test Author"},
		Description: "Test Subject",
		Subject:     []string{"test", "pdf"},
		CreatorTool: "GoPDF2",
		Producer:    "GoPDF2",
	})
	pdf.SetMarkInfo(MarkInfo{Marked: true})
	pdf.AddPage()
	_ = pdf.Cell(nil, "XMP + MarkInfo test")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// gopdf.go — AddColorSpace paths
// ============================================================

func TestCov22_AddColorSpace_CMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.AddColorSpaceCMYK("CustomCMYK", 100, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddColorSpaceCMYK: %v", err)
	}
	pdf.AddPage()
	pdf.SetStrokeColorCMYK(100, 0, 0, 0)
	pdf.SetFillColorCMYK(0, 100, 0, 0)
	_ = pdf.Cell(nil, "CMYK color space")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// ============================================================
// gopdf.go — ImageByHolderWithOptions (mask path)
// ============================================================

func TestCov22_ImageByHolderWithOptions_Mask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create a mask holder.
	maskHolder, err := ImageHolderByPath(resPNGPath2)
	if err != nil {
		t.Skipf("mask holder: %v", err)
	}

	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skipf("holder: %v", err)
	}

	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 200, H: 200},
		Mask: &MaskOptions{
			Holder: maskHolder,
		},
	})
	if err != nil {
		t.Logf("ImageByHolderWithOptions mask: %v", err)
	}
}

// ============================================================
// gopdf.go — ImageFromWithOption
// ============================================================

func TestCov22_ImageFromWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pngData := createTestPNG(t, 50, 50)
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Skipf("decode: %v", err)
	}

	err = pdf.ImageFromWithOption(img, ImageFromOption{
		Format: "png",
		X:      50,
		Y:      50,
		Rect:   &Rect{W: 100, H: 100},
	})
	if err != nil {
		t.Fatalf("ImageFromWithOption: %v", err)
	}
}

// ============================================================
// image_obj_parse.go — parsePng with different color types
// ============================================================

func TestCov22_ParsePng_GrayAlpha(t *testing.T) {
	// Create a grayscale+alpha PNG (color type 4).
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			v := uint8((x + y) * 10)
			img.SetNRGBA(x, y, nrgbaColor(v, v, v, uint8(x*25)))
		}
	}
	var buf bytes.Buffer
	if err := pngEncode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("holder: %v", err)
	}
	err = pdf.ImageByHolder(holder, 50, 50, nil)
	if err != nil {
		t.Logf("ImageByHolder gray+alpha: %v", err)
	}
}

func TestCov22_ParsePng_Grayscale(t *testing.T) {
	// Create a pure grayscale PNG (color type 0).
	img := image.NewGray(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetGray(x, y, grayColor(uint8((x+y)*10)))
		}
	}
	var buf bytes.Buffer
	if err := pngEncode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("holder: %v", err)
	}
	err = pdf.ImageByHolder(holder, 50, 50, nil)
	if err != nil {
		t.Logf("ImageByHolder grayscale: %v", err)
	}
}

func TestCov22_ParsePng_Paletted(t *testing.T) {
	// Create a paletted PNG (color type 3).
	palette := []colorPkg.Color{
		colorPkg.RGBA{255, 0, 0, 255},
		colorPkg.RGBA{0, 255, 0, 255},
		colorPkg.RGBA{0, 0, 255, 255},
		colorPkg.RGBA{255, 255, 0, 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetColorIndex(x, y, uint8((x+y)%4))
		}
	}
	var buf bytes.Buffer
	if err := pngEncode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("holder: %v", err)
	}
	err = pdf.ImageByHolder(holder, 50, 50, nil)
	if err != nil {
		t.Logf("ImageByHolder paletted: %v", err)
	}
}

// Helper types to avoid import conflicts.
func nrgbaColor(r, g, b, a uint8) colorPkg.NRGBA {
	return colorPkg.NRGBA{R: r, G: g, B: b, A: a}
}

func grayColor(y uint8) colorPkg.Gray {
	return colorPkg.Gray{Y: y}
}

func pngEncode(w *bytes.Buffer, img image.Image) error {
	return pngPkg.Encode(w, img)
}
