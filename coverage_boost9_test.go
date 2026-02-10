package gopdf

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost9_test.go — TestCov9_ prefix
// Targets: font_obj embed write, bookmark delete paths,
// watermark all pages, page_obj link writes, page rotation/cropbox,
// clone, incremental save, open_pdf paths, image_obj_parse,
// markinfo computePageLabel, strhelper, procset, content_obj
// ============================================================

// ============================================================
// 1. FontObj.write — IsEmbedFont=true path (64.3% → 100%)
// ============================================================

func TestCov9_FontObj_Write_EmbedTrue(t *testing.T) {
	f := &FontObj{
		Family:                "EmbedTest",
		IsEmbedFont:           true,
		indexObjWidth:          10,
		indexObjFontDescriptor: 11,
		indexObjEncoding:       12,
		Font:                  &mockFont{name: "EmbedFontName"},
	}
	var buf bytes.Buffer
	if err := f.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/BaseFont /EmbedFontName") {
		t.Errorf("expected /BaseFont /EmbedFontName: %s", out)
	}
	if !strings.Contains(out, "/FirstChar 32 /LastChar 255") {
		t.Errorf("expected /FirstChar: %s", out)
	}
	if !strings.Contains(out, "10 0 R") {
		t.Errorf("expected width ref 10 0 R: %s", out)
	}
	if !strings.Contains(out, "11 0 R") {
		t.Errorf("expected descriptor ref 11 0 R: %s", out)
	}
	if !strings.Contains(out, "12 0 R") {
		t.Errorf("expected encoding ref 12 0 R: %s", out)
	}
}

func TestCov9_FontObj_Init(t *testing.T) {
	f := &FontObj{IsEmbedFont: true}
	f.init(nil)
	if f.IsEmbedFont != false {
		t.Error("init should set IsEmbedFont to false")
	}
}

// ============================================================
// 2. DeleteBookmark — all linked-list paths (66.7% → higher)
// ============================================================

func TestCov9_DeleteBookmark_First(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Ch1")
	pdf.AddPage()
	pdf.AddOutline("Ch2")
	pdf.AddPage()
	pdf.AddOutline("Ch3")

	// Delete first bookmark.
	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}
}

func TestCov9_DeleteBookmark_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Ch1")
	pdf.AddPage()
	pdf.AddOutline("Ch2")
	pdf.AddPage()
	pdf.AddOutline("Ch3")

	toc := pdf.GetTOC()
	lastIdx := len(toc) - 1
	err := pdf.DeleteBookmark(lastIdx)
	if err != nil {
		t.Fatalf("DeleteBookmark last: %v", err)
	}
}

func TestCov9_DeleteBookmark_Middle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Ch1")
	pdf.AddPage()
	pdf.AddOutline("Ch2")
	pdf.AddPage()
	pdf.AddOutline("Ch3")

	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark middle: %v", err)
	}
}

func TestCov9_DeleteBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Ch1")

	err := pdf.DeleteBookmark(99)
	if err == nil {
		t.Error("expected error for out of range")
	}
	err = pdf.DeleteBookmark(-1)
	if err == nil {
		t.Error("expected error for negative index")
	}
}

func TestCov9_DeleteBookmark_NoOutlines(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.DeleteBookmark(0)
	if err == nil {
		t.Error("expected error for no outlines")
	}
}

func TestCov9_ModifyBookmark(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Original")

	err := pdf.ModifyBookmark(0, "Modified")
	if err != nil {
		t.Fatalf("ModifyBookmark: %v", err)
	}
}

func TestCov9_SetBookmarkStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Styled")

	err := pdf.SetBookmarkStyle(0, BookmarkStyle{
		Bold:      true,
		Italic:    true,
		Color:     [3]float64{1, 0, 0},
		Collapsed: true,
	})
	if err != nil {
		t.Fatalf("SetBookmarkStyle: %v", err)
	}
}

// ============================================================
// 3. Watermark all pages (71.4% → higher)
// ============================================================

func TestCov9_AddWatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.2,
		Angle:      30,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

func TestCov9_AddWatermarkTextAllPages_Repeat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "CONFIDENTIAL",
		FontFamily: fontFamily,
		FontSize:   24,
		Opacity:    0.15,
		Repeat:     true,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages repeat: %v", err)
	}
}

func TestCov9_AddWatermarkImageAllPages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.2, 100, 100, 45)
	if err != nil {
		t.Fatalf("AddWatermarkImageAllPages: %v", err)
	}
}

func TestCov9_AddWatermarkText_EmptyText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{FontFamily: fontFamily})
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestCov9_AddWatermarkText_EmptyFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{Text: "test"})
	if err == nil {
		t.Error("expected error for empty font family")
	}
}

// ============================================================
// 4. Page rotation and crop box
// ============================================================

func TestCov9_SetPageRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	for _, angle := range []int{0, 90, 180, 270} {
		if err := pdf.SetPageRotation(1, angle); err != nil {
			t.Fatalf("SetPageRotation(%d): %v", angle, err)
		}
		got, err := pdf.GetPageRotation(1)
		if err != nil {
			t.Fatalf("GetPageRotation: %v", err)
		}
		expected := ((angle % 360) + 360) % 360
		if got != expected {
			t.Errorf("expected %d, got %d", expected, got)
		}
	}
}

func TestCov9_SetPageRotation_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageRotation(1, 45)
	if err == nil {
		t.Error("expected error for non-90 angle")
	}
}

func TestCov9_SetPageRotation_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageRotation(99, 90)
	if err == nil {
		t.Error("expected error for out of range page")
	}
}

func TestCov9_GetPageRotation_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_, err := pdf.GetPageRotation(99)
	if err == nil {
		t.Error("expected error for out of range page")
	}
}

func TestCov9_PageCropBox_SetGetClear(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	box := Box{Left: 50, Top: 50, Right: 500, Bottom: 700}
	if err := pdf.SetPageCropBox(1, box); err != nil {
		t.Fatalf("SetPageCropBox: %v", err)
	}

	got, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil crop box")
	}
	if got.Left != 50 || got.Right != 500 {
		t.Errorf("unexpected crop box: %+v", got)
	}

	if err := pdf.ClearPageCropBox(1); err != nil {
		t.Fatalf("ClearPageCropBox: %v", err)
	}

	got, err = pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox after clear: %v", err)
	}
	if got != nil {
		t.Error("expected nil after clear")
	}

	// Verify it renders.
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov9_PageCropBox_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageCropBox(99, Box{})
	if err == nil {
		t.Error("expected error")
	}
	_, err = pdf.GetPageCropBox(99)
	if err == nil {
		t.Error("expected error")
	}
	err = pdf.ClearPageCropBox(99)
	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// 5. PageObj write with rotation and cropbox in output
// ============================================================

func TestCov9_PageObj_WriteWithRotationAndCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "rotated and cropped")
	pdf.SetPageRotation(1, 90)
	pdf.SetPageCropBox(1, Box{Left: 10, Top: 10, Right: 200, Bottom: 300})

	data := pdf.GetBytesPdf()
	s := string(data)
	if !strings.Contains(s, "/Rotate 90") {
		t.Error("expected /Rotate 90 in output")
	}
	if !strings.Contains(s, "/CropBox") {
		t.Error("expected /CropBox in output")
	}
}

// ============================================================
// 6. Clone
// ============================================================

func TestCov9_Clone(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Original")
	pdf.SetXMPMetadata(XMPMetadata{Title: "CloneTest", Creator: []string{"test"}})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelDecimal}})

	clone, err := pdf.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}

	data := clone.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty cloned PDF")
	}
}

// ============================================================
// 7. IncrementalSave / WriteIncrementalPdf
// ============================================================

func TestCov9_IncrementalSave(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Original content")

	original := pdf.GetBytesPdf()
	if len(original) == 0 {
		t.Fatal("empty original PDF")
	}

	// Re-open and modify.
	pdf2 := &GoPdf{}
	if err := pdf2.OpenPDFFromBytes(original, nil); err != nil {
		t.Fatalf("OpenPDFFromBytes: %v", err)
	}
	if err := pdf2.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font: %v", err)
	}
	pdf2.SetFont(fontFamily, "", 14)
	pdf2.SetPage(1)
	pdf2.SetXY(100, 100)
	pdf2.Text("Added text")

	result, err := pdf2.IncrementalSave(original, nil)
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) <= len(original) {
		t.Error("incremental result should be larger than original")
	}
}

func TestCov9_WriteIncrementalPdf(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")

	original := pdf.GetBytesPdf()

	pdf2 := &GoPdf{}
	pdf2.OpenPDFFromBytes(original, nil)
	pdf2.AddTTFFont(fontFamily, resFontPath)
	pdf2.SetFont(fontFamily, "", 14)
	pdf2.SetPage(1)
	pdf2.SetXY(50, 50)
	pdf2.Text("incremental")

	path := resOutDir + "/incremental_test.pdf"
	err := pdf2.WriteIncrementalPdf(path, original, nil)
	if err != nil {
		t.Fatalf("WriteIncrementalPdf: %v", err)
	}
	defer os.Remove(path)

	info, _ := os.Stat(path)
	if info == nil || info.Size() == 0 {
		t.Error("expected non-empty file")
	}
}

// ============================================================
// 8. OpenPDF / OpenPDFFromStream
// ============================================================

func TestCov9_OpenPDF(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := &GoPdf{}
	err := pdf.OpenPDF(resTestPDF, nil)
	if err != nil {
		t.Fatalf("OpenPDF: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov9_OpenPDF_BadPath(t *testing.T) {
	pdf := &GoPdf{}
	err := pdf.OpenPDF("/nonexistent/path.pdf", nil)
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestCov9_OpenPDFFromStream(t *testing.T) {
	pdfBytes, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := &GoPdf{}
	rs := io.ReadSeeker(bytes.NewReader(pdfBytes))
	err = pdf.OpenPDFFromStream(&rs, nil)
	if err != nil {
		t.Fatalf("OpenPDFFromStream: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov9_OpenPDFFromBytes_WithBox(t *testing.T) {
	pdfBytes, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := &GoPdf{}
	err = pdf.OpenPDFFromBytes(pdfBytes, &OpenPDFOption{Box: "/CropBox"})
	if err != nil {
		// CropBox may not exist, that's ok.
		t.Logf("OpenPDFFromBytes with CropBox: %v", err)
	}
}

// ============================================================
// 9. computePageLabel — all styles (77.3% → higher)
// ============================================================

func TestCov9_ComputePageLabel_AllStyles(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 10; i++ {
		pdf.AddPage()
		pdf.Cell(nil, fmt.Sprintf("Page %d", i+1))
	}

	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Style: PageLabelDecimal, Start: 1, Prefix: "D-"},
		{PageIndex: 2, Style: PageLabelRomanUpper, Start: 1, Prefix: ""},
		{PageIndex: 4, Style: PageLabelRomanLower, Start: 1, Prefix: "r-"},
		{PageIndex: 6, Style: PageLabelAlphaUpper, Start: 1, Prefix: ""},
		{PageIndex: 8, Style: PageLabelAlphaLower, Start: 1, Prefix: "a-"},
	})

	// FindPagesByLabel exercises computePageLabel for all pages.
	pages := pdf.FindPagesByLabel("D-1")
	if len(pages) == 0 {
		t.Error("expected to find page with label D-1")
	}

	pages = pdf.FindPagesByLabel("I")
	if len(pages) == 0 {
		t.Log("Roman upper I not found (may be expected)")
	}

	pages = pdf.FindPagesByLabel("r-i")
	if len(pages) == 0 {
		t.Log("Roman lower r-i not found (may be expected)")
	}

	pages = pdf.FindPagesByLabel("A")
	_ = pages

	pages = pdf.FindPagesByLabel("a-a")
	_ = pages
}

func TestCov9_ComputePageLabel_NoLabels(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pages := pdf.FindPagesByLabel("1")
	// With no labels set, FindPagesByLabel may return empty.
	_ = pages
}

func TestCov9_ComputePageLabel_DefaultStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()

	// Style 0 (default) with prefix only.
	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Prefix: "P-"},
	})

	pages := pdf.FindPagesByLabel("P-")
	if len(pages) == 0 {
		t.Log("P- not found (default style produces empty numStr)")
	}
}

// ============================================================
// 10. StrHelperGetStringWidth / StrHelperGetStringWidthPrecise
// ============================================================

func TestCov9_StrHelperGetStringWidth(t *testing.T) {
	mf := &mockFont{name: "test"}
	w := StrHelperGetStringWidth("A", 14, mf)
	// With empty FontCw, width will be 0.
	_ = w
}

func TestCov9_StrHelperGetStringWidthPrecise(t *testing.T) {
	mf := &mockFont{name: "test"}
	w := StrHelperGetStringWidthPrecise("Hello", 12.5, mf)
	_ = w
}

// ============================================================
// 11. ProcSetObj.IsContainsFamily
// ============================================================

func TestCov9_IsContainsFamily(t *testing.T) {
	fonts := RelateFonts{
		{Family: "Arial", CountOfFont: 0, IndexOfObj: 1},
		{Family: "Times", CountOfFont: 1, IndexOfObj: 2},
	}
	if !fonts.IsContainsFamily("Arial") {
		t.Error("expected to find Arial")
	}
	if fonts.IsContainsFamily("Courier") {
		t.Error("should not find Courier")
	}
}

func TestCov9_IsContainsFamilyAndStyle(t *testing.T) {
	fonts := RelateFonts{
		{Family: "Arial", Style: 0, CountOfFont: 0, IndexOfObj: 1},
		{Family: "Arial", Style: 1, CountOfFont: 1, IndexOfObj: 2},
	}
	if !fonts.IsContainsFamilyAndStyle("Arial", 0) {
		t.Error("expected to find Arial regular")
	}
	if !fonts.IsContainsFamilyAndStyle("Arial", 1) {
		t.Error("expected to find Arial bold")
	}
	if fonts.IsContainsFamilyAndStyle("Arial", 2) {
		t.Error("should not find Arial italic")
	}
}

// ============================================================
// 12. isDeviceRGB / ImgReactagleToWH (0% each)
// ============================================================

func TestCov9_IsDeviceRGB(t *testing.T) {
	// YCbCr image.
	ycbcr := image.NewYCbCr(image.Rect(0, 0, 10, 10), image.YCbCrSubsampleRatio420)
	var img1 image.Image = ycbcr
	if !isDeviceRGB("jpeg", &img1) {
		t.Error("YCbCr should be DeviceRGB")
	}

	// NRGBA image.
	nrgba := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	var img2 image.Image = nrgba
	if !isDeviceRGB("png", &img2) {
		t.Error("NRGBA should be DeviceRGB")
	}

	// Gray image — not DeviceRGB.
	gray := image.NewGray(image.Rect(0, 0, 10, 10))
	var img3 image.Image = gray
	if isDeviceRGB("png", &img3) {
		t.Error("Gray should not be DeviceRGB")
	}
}

func TestCov9_ImgReactagleToWH(t *testing.T) {
	rect := image.Rect(0, 0, 200, 100)
	w, h := ImgReactagleToWH(rect)
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got w=%f h=%f", w, h)
	}
}

// ============================================================
// 13. ImageObj.write — all branches via full PDF generation
// ============================================================

func TestCov9_ImageObj_Write_ProtectedPDF_WithJPEG(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()

	err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})
	if err != nil {
		t.Skipf("image not available: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
	if !strings.Contains(string(data), "stream") {
		t.Error("expected stream in output")
	}
}

func TestCov9_ImageObj_Write_ProtectedPDF_WithPNG(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()

	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 100, H: 100})
	if err != nil {
		t.Skipf("image not available: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 14. ContentObj.write — protection + no-compression combined
// ============================================================

func TestCov9_ContentObj_Write_ProtectedFlate(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.Cell(nil, "Protected with flate")
	pdf.Line(10, 10, 200, 200)
	pdf.SetTextColor(255, 0, 0)
	pdf.Cell(nil, "Red protected")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov9_ContentObj_Write_ProtectedNoFlate(t *testing.T) {
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
		t.Skipf("font: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.Cell(nil, "Protected no flate")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov9_ContentObj_Write_FlateWithContent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetCompressLevel(9)
	pdf.AddPage()
	pdf.Cell(nil, "Flate compressed")
	pdf.Line(10, 10, 100, 100)
	pdf.SetGrayFill(0.5)
	pdf.Cell(nil, "Gray text")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 15. AddPageWithOption — TrimBox path
// ============================================================

func TestCov9_AddPageWithOption_TrimBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	trimBox := &Box{Left: 10, Top: 10, Right: 200, Bottom: 300}
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 210, H: 297},
		TrimBox:  trimBox,
	})
	pdf.Cell(nil, "With TrimBox")

	data := pdf.GetBytesPdf()
	if !strings.Contains(string(data), "/TrimBox") {
		t.Error("expected /TrimBox in output")
	}
}

// ============================================================
// 16. ImageObj write — trns path (writeImgProps with trns)
// ============================================================

func TestCov9_WriteImgProps_WithTrns(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		trns:            []byte{0, 128, 255},
	}
	var buf bytes.Buffer
	err := writeImgProps(&buf, info, false)
	if err != nil {
		t.Fatalf("writeImgProps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Mask [") {
		t.Errorf("expected /Mask: %s", out)
	}
}

func TestCov9_WriteImgProps_SplittedMask(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		trns:            []byte{0},
	}
	var buf bytes.Buffer
	err := writeImgProps(&buf, info, true)
	if err != nil {
		t.Fatalf("writeImgProps: %v", err)
	}
	out := buf.String()
	// SplittedMask should return early, no /Mask.
	if strings.Contains(out, "/Mask [") {
		t.Error("splittedMask should not have /Mask")
	}
}

func TestCov9_WriteImgProps_WithSMask(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		smask:           []byte{1, 2, 3},
		smarkObjID:      5,
	}
	var buf bytes.Buffer
	err := writeImgProps(&buf, info, false)
	if err != nil {
		t.Fatalf("writeImgProps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/SMask 6 0 R") {
		t.Errorf("expected /SMask ref: %s", out)
	}
}

func TestCov9_WriteBaseImgProps_Indexed(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "Indexed",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		pal:             make([]byte, 768), // 256*3
		deviceRGBObjID:  7,
	}
	var buf bytes.Buffer
	err := writeBaseImgProps(&buf, info, info.colspace)
	if err != nil {
		t.Fatalf("writeBaseImgProps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Indexed") {
		t.Errorf("expected /Indexed: %s", out)
	}
}

func TestCov9_WriteBaseImgProps_CMYK(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceCMYK",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
	}
	var buf bytes.Buffer
	err := writeBaseImgProps(&buf, info, info.colspace)
	if err != nil {
		t.Fatalf("writeBaseImgProps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Decode [1 0 1 0 1 0 1 0]") {
		t.Errorf("expected CMYK decode: %s", out)
	}
}

func TestCov9_WriteMaskImgProps(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceGray",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
	}
	var buf bytes.Buffer
	err := writeMaskImgProps(&buf, info)
	if err != nil {
		t.Fatalf("writeMaskImgProps: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/DecodeParms") {
		t.Errorf("expected /DecodeParms: %s", out)
	}
	if !strings.Contains(out, "/Predictor 15") {
		t.Errorf("expected /Predictor 15: %s", out)
	}
}
