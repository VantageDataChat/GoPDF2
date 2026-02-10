package gopdf

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost19_test.go — TestCov19_ prefix
// Targets: watermark AllPages, bookmark delete/collect, open_pdf
// branches, image_recompress rebuildXref, content_obj write,
// image_obj write, gopdf.go Text/Line/Read/GetBytesPdf/Cell/
// MeasureTextWidth/MeasureCellHeightByText, list_cache_content,
// WriteIncrementalPdf, form_field findCurrentPageObjID,
// cache_content_text underline, cache_content_rotate,
// pdf_parser extractNamedRefs, subset_font_obj charCodeToGlyphIndex
// ============================================================

// ============================================================
// watermark.go — AddWatermarkTextAllPages, AddWatermarkImageAllPages
// ============================================================

func TestCov19_AddWatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   48,
		Opacity:    0.3,
		Angle:      45,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

func TestCov19_AddWatermarkTextAllPages_Repeat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "CONFIDENTIAL",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.2,
		Angle:      30,
		Repeat:     true,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages repeat: %v", err)
	}
}

func TestCov19_AddWatermarkImageAllPages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.3, 200, 200, 45)
	if err != nil {
		t.Fatalf("AddWatermarkImageAllPages: %v", err)
	}
}

func TestCov19_AddWatermarkImageAllPages_NoAngle(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.5, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImageAllPages no angle: %v", err)
	}
}

// ============================================================
// bookmark.go — DeleteBookmark with linked-list edge cases
// ============================================================

func TestCov19_DeleteBookmark_First(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")
	pdf.AddPage()
	pdf.AddOutline("Chapter 3")

	// Delete the first bookmark (index 0).
	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark(0): %v", err)
	}
}

func TestCov19_DeleteBookmark_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")
	pdf.AddPage()
	pdf.AddOutline("Chapter 3")

	toc := pdf.GetTOC()
	lastIdx := len(toc) - 1
	err := pdf.DeleteBookmark(lastIdx)
	if err != nil {
		t.Fatalf("DeleteBookmark(last): %v", err)
	}
}

func TestCov19_DeleteBookmark_Middle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")
	pdf.AddPage()
	pdf.AddOutline("Chapter 3")

	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark(1): %v", err)
	}
}

func TestCov19_DeleteBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")

	err := pdf.DeleteBookmark(-1)
	if err == nil {
		t.Fatal("expected error for negative index")
	}
	err = pdf.DeleteBookmark(99)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestCov19_DeleteBookmark_NoOutlines(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.DeleteBookmark(0)
	if err == nil {
		t.Fatal("expected error when no outlines")
	}
}

// ============================================================
// bookmark.go — collectOutlineObjs with hierarchical outlines
// ============================================================

func TestCov19_CollectOutlineObjs_Hierarchical(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")
	pdf.AddPage()
	pdf.AddOutline("Chapter 3")

	// SetTOC with hierarchy to exercise collectOutlineObjs with children.
	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Part A", PageNo: 1},
		{Level: 2, Title: "Section A.1", PageNo: 1},
		{Level: 2, Title: "Section A.2", PageNo: 2},
		{Level: 1, Title: "Part B", PageNo: 3},
		{Level: 2, Title: "Section B.1", PageNo: 3},
	})
	if err != nil {
		t.Fatalf("SetTOC: %v", err)
	}

	toc := pdf.GetTOC()
	if len(toc) < 5 {
		t.Errorf("expected at least 5 TOC items, got %d", len(toc))
	}
}

// ============================================================
// open_pdf.go — OpenPDFFromStream, OpenPDF with protection
// ============================================================

func TestCov19_OpenPDFFromStream(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	rs := io.ReadSeeker(bytes.NewReader(data))
	pdf := GoPdf{}
	if err := pdf.OpenPDFFromStream(&rs, nil); err != nil {
		t.Fatalf("OpenPDFFromStream: %v", err)
	}
	if pdf.GetNumberOfPages() == 0 {
		t.Fatal("expected pages")
	}
}

func TestCov19_OpenPDFFromStream_WithBox(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	rs := io.ReadSeeker(bytes.NewReader(data))
	pdf := GoPdf{}
	if err := pdf.OpenPDFFromStream(&rs, &OpenPDFOption{Box: "/MediaBox"}); err != nil {
		t.Fatalf("OpenPDFFromStream with box: %v", err)
	}
}

func TestCov19_OpenPDF_WithProtection(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := GoPdf{}
	prot := PDFProtectionConfig{
		UseProtection: true,
		Permissions:   PermissionsPrint,
		UserPass:      []byte("user"),
		OwnerPass:     []byte("owner"),
	}
	if err := pdf.OpenPDFFromBytes(data, &OpenPDFOption{Protection: &prot}); err != nil {
		t.Fatalf("OpenPDFFromBytes with protection: %v", err)
	}
}

func TestCov19_OpenPDF_InvalidData(t *testing.T) {
	pdf := GoPdf{}
	err := pdf.OpenPDFFromBytes([]byte("not a pdf"), nil)
	// gofpdi may panic — we just want no crash.
	_ = err
}

func TestCov19_OpenPDF_EmptyData(t *testing.T) {
	pdf := GoPdf{}
	err := pdf.OpenPDFFromBytes([]byte("%%PDF-1.4 invalid"), nil)
	// gofpdi may panic on invalid data — just ensure no hang.
	_ = err
}

func TestCov19_OpenPDF_NonexistentFile(t *testing.T) {
	pdf := GoPdf{}
	err := pdf.OpenPDF("/nonexistent/path/file.pdf", nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// ============================================================
// image_recompress.go — RecompressImages with various options
// ============================================================

func TestCov19_RecompressImages_JPEG(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 200, H: 200}); err != nil {
		t.Skipf("Image: %v", err)
	}
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := RecompressImages(data, RecompressOption{
		Format:      "jpeg",
		JPEGQuality: 50,
	})
	if err != nil {
		t.Fatalf("RecompressImages JPEG: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov19_RecompressImages_PNG(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 200, H: 200}); err != nil {
		t.Skipf("Image: %v", err)
	}
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := RecompressImages(data, RecompressOption{
		Format: "png",
	})
	if err != nil {
		t.Fatalf("RecompressImages PNG: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov19_RecompressImages_WithMaxSize(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 200, H: 200}); err != nil {
		t.Skipf("Image: %v", err)
	}
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := RecompressImages(data, RecompressOption{
		Format:      "jpeg",
		JPEGQuality: 60,
		MaxWidth:    100,
		MaxHeight:   100,
	})
	if err != nil {
		t.Fatalf("RecompressImages with max size: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov19_RecompressImages_InvalidPDF(t *testing.T) {
	result, err := RecompressImages([]byte("not a pdf"), RecompressOption{})
	// May or may not error; just ensure no panic.
	_ = result
	_ = err
}

func TestCov19_RecompressImages_NoPDF(t *testing.T) {
	result, err := RecompressImages([]byte{}, RecompressOption{})
	// May or may not error; just ensure no panic.
	_ = result
	_ = err
}

func TestCov19_RecompressImages_DefaultOptions(t *testing.T) {
	opt := RecompressOption{}
	opt.defaults()
	if opt.Format != "jpeg" {
		t.Errorf("expected jpeg, got %s", opt.Format)
	}
	if opt.JPEGQuality != 75 {
		t.Errorf("expected 75, got %d", opt.JPEGQuality)
	}
}

// ============================================================
// gopdf.go — Text with no font set (error path)
// ============================================================

func TestCov19_Text_NoFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()

	// Text without font should fail or panic — recover.
	func() {
		defer func() { recover() }()
		_ = pdf.Text("hello")
	}()
}

// ============================================================
// gopdf.go — Line with transparency
// ============================================================

func TestCov19_Line_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	tr, err := NewTransparency(0.5, "")
	if err != nil {
		t.Fatalf("NewTransparency: %v", err)
	}
	if err := pdf.SetTransparency(tr); err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	pdf.Line(10, 10, 200, 200)
	pdf.ClearTransparency()
	pdf.Line(10, 20, 200, 20)
}

// ============================================================
// gopdf.go — Read
// ============================================================

func TestCov19_Read_Full(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Read test")

	// Read all bytes.
	var all []byte
	buf := make([]byte, 1024)
	for {
		n, err := pdf.Read(buf)
		if n > 0 {
			all = append(all, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
	}
	if !bytes.HasPrefix(all, []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

func TestCov19_Read_SmallBuffer(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Small buffer read")

	buf := make([]byte, 3)
	n, err := pdf.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Error("expected some bytes")
	}
}

// ============================================================
// gopdf.go — GetBytesPdf (error path)
// ============================================================

func TestCov19_GetBytesPdf_Valid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "GetBytesPdf test")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

// ============================================================
// gopdf.go — CellWithOption with various alignments
// ============================================================

func TestCov19_CellWithOption_AllAlignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	aligns := []int{Left | Top, Center | Middle, Right | Bottom, Left | Bottom, Right | Top}
	for i, a := range aligns {
		pdf.SetXY(50, float64(50+i*30))
		err := pdf.CellWithOption(&Rect{W: 200, H: 25}, "Align test", CellOption{
			Align:  a,
			Border: Left | Right | Top | Bottom,
		})
		if err != nil {
			t.Fatalf("CellWithOption align %d: %v", a, err)
		}
	}
}

func TestCov19_CellWithOption_WithBreakOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 25}, "Cell with break option", CellOption{
		Align:       Center,
		BreakOption: &DefaultBreakOption,
	})
	if err != nil {
		t.Fatalf("CellWithOption with break: %v", err)
	}
}

// ============================================================
// gopdf.go — MeasureTextWidth / MeasureCellHeightByText edge cases
// ============================================================

func TestCov19_MeasureTextWidth_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_, err := pdf.MeasureTextWidth("")
	// Empty string may return error or 0 width.
	_ = err
}

func TestCov19_MeasureCellHeightByText_Long(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	h, err := pdf.MeasureCellHeightByText("A very long text that should have some height measurement")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Error("expected positive height")
	}
}

// ============================================================
// gopdf.go — Polygon / Polyline with transparency
// ============================================================

func TestCov19_Polygon_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	tr, _ := NewTransparency(0.5, "")
	_ = pdf.SetTransparency(tr)

	pdf.Polygon([]Point{
		{X: 100, Y: 100},
		{X: 200, Y: 100},
		{X: 150, Y: 200},
	}, "DF")

	pdf.ClearTransparency()
}

func TestCov19_Polyline_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	tr, _ := NewTransparency(0.5, "")
	_ = pdf.SetTransparency(tr)

	pdf.Polyline([]Point{
		{X: 50, Y: 50},
		{X: 100, Y: 80},
		{X: 150, Y: 50},
		{X: 200, Y: 80},
	})

	pdf.ClearTransparency()
}

// ============================================================
// incremental_save.go — WriteIncrementalPdf
// ============================================================

func TestCov19_WriteIncrementalPdf(t *testing.T) {
	ensureOutDir(t)
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := GoPdf{}
	if err := pdf.OpenPDFFromBytes(data, nil); err != nil {
		t.Skipf("OpenPDFFromBytes: %v", err)
	}

	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.SetPage(1)
	pdf.SetXY(100, 100)
	_ = pdf.Text("Incremental text")

	outPath := resOutDir + "/cov19_incremental.pdf"
	err = pdf.WriteIncrementalPdf(outPath, data, nil)
	if err != nil {
		t.Fatalf("WriteIncrementalPdf: %v", err)
	}
	defer os.Remove(outPath)

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty file")
	}
}

func TestCov19_WriteIncrementalPdf_WithIndices(t *testing.T) {
	ensureOutDir(t)
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := GoPdf{}
	if err := pdf.OpenPDFFromBytes(data, nil); err != nil {
		t.Skipf("OpenPDFFromBytes: %v", err)
	}

	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.SetPage(1)
	pdf.SetXY(100, 100)
	_ = pdf.Text("Incremental with indices")

	outPath := resOutDir + "/cov19_incremental2.pdf"
	err = pdf.WriteIncrementalPdf(outPath, data, []int{0, 1, 2})
	if err != nil {
		t.Fatalf("WriteIncrementalPdf with indices: %v", err)
	}
	defer os.Remove(outPath)
}

// ============================================================
// content_obj.go — write with no compression
// ============================================================

func TestCov19_ContentObj_Write_NoCompression(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetNoCompression()
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "No compression content")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

// ============================================================
// image_obj.go — write with various image types
// ============================================================

func TestCov19_ImageObj_Write_JPEG(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("Image JPEG: %v", err)
	}

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov19_ImageObj_Write_PNG(t *testing.T) {
	if _, err := os.Stat(resPNGPath); os.IsNotExist(err) {
		t.Skip("PNG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("Image PNG: %v", err)
	}

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov19_ImageObj_Write_PNG2(t *testing.T) {
	if _, err := os.Stat(resPNGPath2); os.IsNotExist(err) {
		t.Skip("PNG2 not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Image(resPNGPath2, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("Image PNG2: %v", err)
	}

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// cache_content_text.go — underline path
// ============================================================

func TestCov19_Underline_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set underline style.
	if err := pdf.SetFontWithStyle(fontFamily, Underline, 14); err != nil {
		t.Fatalf("SetFontWithStyle underline: %v", err)
	}
	pdf.SetXY(50, 50)
	if err := pdf.Cell(&Rect{W: 200, H: 20}, "Underlined text"); err != nil {
		t.Fatalf("Cell underline: %v", err)
	}
}

func TestCov19_Bold_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Bold style uses the same font family — just set style flag.
	if err := pdf.SetFontWithStyle(fontFamily, Bold, 14); err != nil {
		// Some fonts don't have bold variant — skip.
		t.Skipf("SetFontWithStyle bold: %v", err)
	}
	pdf.SetXY(50, 80)
	if err := pdf.Cell(&Rect{W: 200, H: 20}, "Bold text"); err != nil {
		t.Fatalf("Cell bold: %v", err)
	}
}

func TestCov19_Underline_MultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	if err := pdf.SetFontWithStyle(fontFamily, Underline, 12); err != nil {
		t.Fatalf("SetFontWithStyle underline: %v", err)
	}
	pdf.SetXY(50, 50)
	if err := pdf.MultiCell(&Rect{W: 150, H: 60}, "Underlined multi cell text that wraps"); err != nil {
		t.Fatalf("MultiCell underline: %v", err)
	}
}

// ============================================================
// cache_content_rotate.go — Rotate with various angles
// ============================================================

func TestCov19_Rotate_Various(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	angles := []float64{0, 30, 45, 90, 180, 270, 360}
	for _, angle := range angles {
		pdf.Rotate(angle, 200, 400)
		pdf.SetXY(180, 390)
		_ = pdf.Text("R")
		pdf.RotateReset()
	}
}

// ============================================================
// form_field.go — AddFormField and findCurrentPageObjID
// ============================================================

func TestCov19_AddFormField_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:  FormFieldText,
		Name:  "name_field",
		X:     50,
		Y:     50,
		W:     200,
		H:     25,
		Value: "Default Name",
	})
	if err != nil {
		t.Fatalf("AddFormField text: %v", err)
	}
}

func TestCov19_AddFormField_Checkbox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:  FormFieldCheckbox,
		Name:  "agree_field",
		X:     50,
		Y:     100,
		W:     20,
		H:     20,
		Value: "Yes",
	})
	if err != nil {
		t.Fatalf("AddFormField checkbox: %v", err)
	}
}

func TestCov19_AddFormField_Choice(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:    FormFieldChoice,
		Name:    "color_field",
		X:       50,
		Y:       150,
		W:       200,
		H:       25,
		Options: []string{"Red", "Green", "Blue"},
		Value:   "Red",
	})
	if err != nil {
		t.Fatalf("AddFormField choice: %v", err)
	}
}

func TestCov19_AddFormField_MultiplePages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "f1",
		X: 50, Y: 50, W: 200, H: 25,
	})

	pdf.AddPage()
	_ = pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "f2",
		X: 50, Y: 50, W: 200, H: 25,
	})

	pdf.AddPage()
	_ = pdf.AddFormField(FormField{
		Type: FormFieldCheckbox, Name: "f3",
		X: 50, Y: 50, W: 20, H: 20,
	})

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// list_cache_content.go — write with protection
// ============================================================

func TestCov19_ListCacheContent_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Protected list cache")
	pdf.Line(10, 10, 100, 100)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// svg_insert.go — ImageSVGFromReader
// ============================================================

func TestCov19_ImageSVGFromReader(t *testing.T) {
	svgData := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="red" stroke="black" stroke-width="2"/>
		<circle cx="50" cy="50" r="30" fill="blue"/>
		<line x1="10" y1="10" x2="90" y2="90" stroke="green" stroke-width="1"/>
	</svg>`

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVGFromReader(strings.NewReader(svgData), SVGOption{
		X:      50,
		Y:      50,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromReader: %v", err)
	}
}

func TestCov19_ImageSVGFromReader_Path(t *testing.T) {
	svgData := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<path d="M10 10 L90 10 L90 90 L10 90 Z" fill="none" stroke="black"/>
		<path d="M50 10 Q90 50 50 90 Q10 50 50 10" fill="yellow" stroke="red"/>
	</svg>`

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVGFromReader(strings.NewReader(svgData), SVGOption{
		X:      50,
		Y:      300,
		Width:  150,
		Height: 150,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromReader path: %v", err)
	}
}

// ============================================================
// digital_signature.go — LoadCertificateFromPEM, LoadPrivateKeyFromPEM
// ============================================================

func TestCov19_LoadCertificateFromPEM_Invalid(t *testing.T) {
	_, err := LoadCertificateFromPEM("nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCov19_LoadPrivateKeyFromPEM_Invalid(t *testing.T) {
	_, err := LoadPrivateKeyFromPEM("nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCov19_LoadCertificateChainFromPEM_Invalid(t *testing.T) {
	_, err := LoadCertificateChainFromPEM("nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCov19_VerifySignatureFromFile_Invalid(t *testing.T) {
	_, err := VerifySignatureFromFile("nonexistent.pdf")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCov19_VerifySignature_InvalidPDF(t *testing.T) {
	_, err := VerifySignature([]byte("not a pdf"))
	// May or may not error, just no panic.
	_ = err
}

func TestCov19_SignPDFToFile_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Sign test")

	err := pdf.SignPDFToFile(SignatureConfig{}, resOutDir+"/cov19_signed.pdf")
	// Expected to fail without valid cert/key.
	_ = err
}

// ============================================================
// gopdf.go — ImageByHolderWithOptions with mask + transparency
// ============================================================

func TestCov19_ImageByHolderWithOptions_MaskAndTransparency(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Fatalf("ImageHolderByPath: %v", err)
	}

	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 200, H: 200},
		Transparency: &Transparency{
			Alpha:         0.7,
			BlendModeType: Multiply,
		},
	})
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions transparency+multiply: %v", err)
	}
}

// ============================================================
// gopdf.go — Sector
// ============================================================

func TestCov19_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Sector(200, 300, 80, 0, 90, "FD")
	pdf.Sector(200, 300, 80, 90, 180, "D")
}

// ============================================================
// gopdf.go — Rectangle with round corners
// ============================================================

func TestCov19_Rectangle_RoundCorners(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetFillColor(200, 200, 255)
	pdf.SetStrokeColor(0, 0, 128)

	if err := pdf.Rectangle(50, 50, 250, 150, "DF", 15, 10); err != nil {
		t.Fatalf("Rectangle round: %v", err)
	}
	if err := pdf.Rectangle(50, 200, 250, 300, "F", 0, 0); err != nil {
		t.Fatalf("Rectangle fill: %v", err)
	}
	if err := pdf.Rectangle(50, 350, 250, 450, "D", 5, 5); err != nil {
		t.Fatalf("Rectangle draw: %v", err)
	}
}

// ============================================================
// gopdf.go — prepare (various PDF features combined)
// ============================================================

func TestCov19_Prepare_FullFeatures(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)

	// Set PDF info.
	pdf.SetInfo(PdfInfo{
		Title:   "Full Features Test",
		Author:  "Test",
		Subject: "Coverage",
	})

	// Add page labels.
	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Style: PageLabelRomanUpper},
		{PageIndex: 2, Style: PageLabelDecimal, Prefix: "Ch-"},
	})

	// Add pages with content.
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Chapter 1 content")

	pdf.AddPage()
	pdf.AddOutline("Chapter 2")
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Chapter 2 content")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Chapter 3 content")

	// Add embedded file.
	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "data.txt",
		Content: []byte("test data"),
	})

	// Write.
	outPath := resOutDir + "/cov19_full_features.pdf"
	if err := pdf.WritePdf(outPath); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
	defer os.Remove(outPath)
}

// ============================================================
// image_holder.go — newImageBuffByReader
// ============================================================

func TestCov19_ImageHolderByReader_JPEG(t *testing.T) {
	// Create a small JPEG in memory.
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})

	holder, err := ImageHolderByReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ImageHolderByReader: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.ImageByHolder(holder, 50, 50, &Rect{W: 50, H: 50}); err != nil {
		t.Fatalf("ImageByHolder: %v", err)
	}
}

// ============================================================
// pdf_parser.go — extractNamedRefs
// ============================================================

func TestCov19_SearchText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Hello World searchable text")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	results, err := SearchText(data, "Hello", false)
	// May or may not find text depending on encoding.
	_ = results
	_ = err
}

func TestCov19_SearchTextOnPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Searchable content on page")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	results, err := SearchTextOnPage(data, 1, "Searchable", false)
	_ = results
	_ = err
}

// ============================================================
// embedded_file.go — UpdateEmbeddedFile, GetEmbeddedFileInfo
// ============================================================

func TestCov19_UpdateEmbeddedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("text")

	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("original"),
	})

	err := pdf.UpdateEmbeddedFile("test.txt", EmbeddedFile{
		Name:        "test.txt",
		Content:     []byte("updated content"),
		MimeType:    "text/plain",
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}

	data, err := pdf.GetEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFile: %v", err)
	}
	if string(data) != "updated content" {
		t.Errorf("expected 'updated content', got %q", string(data))
	}

	info, err := pdf.GetEmbeddedFileInfo("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFileInfo: %v", err)
	}
	if info.Description != "Updated description" {
		t.Errorf("expected 'Updated description', got %q", info.Description)
	}
}

func TestCov19_UpdateEmbeddedFile_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.UpdateEmbeddedFile("nonexistent", EmbeddedFile{
		Content: []byte("data"),
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// ============================================================
// gopdf.go — SetTransparency / getCachedTransparency / saveTransparency
// ============================================================

func TestCov19_SetTransparency_MultipleBlendModes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	modes := []BlendModeType{NormalBlendMode, Multiply, "Screen", "Overlay"}
	for _, mode := range modes {
		tr, err := NewTransparency(0.5, string(mode))
		if err != nil {
			t.Logf("NewTransparency %s: %v", string(mode), err)
			continue
		}
		if err := pdf.SetTransparency(tr); err != nil {
			t.Logf("SetTransparency %s: %v", mode, err)
			continue
		}
		pdf.SetXY(50, 50)
		_ = pdf.Text("Blend: " + string(mode))
		pdf.ClearTransparency()
	}
}

// ============================================================
// clone.go — Clone
// ============================================================

func TestCov19_Clone(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Original")

	cloned, err := pdf.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if cloned.GetNumberOfPages() != pdf.GetNumberOfPages() {
		t.Errorf("cloned pages %d != original %d", cloned.GetNumberOfPages(), pdf.GetNumberOfPages())
	}

	// Re-set font on clone before using it.
	if err := cloned.SetFont(fontFamily, "", 14); err != nil {
		t.Skipf("SetFont on clone: %v", err)
	}

	// Modify clone without affecting original.
	cloned.AddPage()
	cloned.SetXY(50, 50)
	_ = cloned.Cell(nil, "Cloned page")

	if cloned.GetNumberOfPages() == pdf.GetNumberOfPages() {
		t.Error("clone should have more pages than original")
	}
}

// ============================================================
// pdf_lowlevel.go — CopyObject, GetCatalog
// ============================================================

func TestCov19_CopyObject(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Copy test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	copied, newNum, err := CopyObject(data, 1)
	// May or may not succeed depending on object structure.
	_ = copied
	_ = newNum
	_ = err
}

func TestCov19_GetCatalog(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Catalog test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	catalog, err := GetCatalog(data)
	_ = catalog
	_ = err
}

// ============================================================
// doc_stats.go — GetFonts
// ============================================================

func TestCov19_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Font test")

	fonts := pdf.GetFonts()
	_ = fonts
}

// ============================================================
// page_option.go — isTrimBoxSet
// ============================================================

func TestCov19_PageOption_TrimBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 400, H: 600},
		TrimBox: &Box{
			Left:   10,
			Top:    10,
			Right:  390,
			Bottom: 590,
		},
	})
	pdf.SetXY(50, 50)
	_ = pdf.Text("TrimBox page")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// pdf_version.go — String
// ============================================================

func TestCov19_PDFVersion_String(t *testing.T) {
	versions := []PDFVersion{PDFVersion14, PDFVersion17, PDFVersion20}
	for _, v := range versions {
		s := v.String()
		if s == "" {
			t.Errorf("expected non-empty string for version %d", v)
		}
	}
	// Unknown version.
	s := PDFVersion(99).String()
	if s == "" {
		t.Error("expected non-empty string for unknown version")
	}
}
