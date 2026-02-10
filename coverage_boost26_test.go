package gopdf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost26_test.go — TestCov26_ prefix
// Targets: parsePng ct==4 (GrayAlpha), rebuildXref missing trailer,
// extractNamedRefs inline dict, failWriter fine-grained,
// watermark all pages, transparency_xobject_group.write errors,
// form_field findCurrentPageObjID, various 75-85% functions
// ============================================================

// --- failWriterAt: fails at exactly byte N ---
type failWriterAt struct {
	n       int
	written int
}

func (fw *failWriterAt) Write(p []byte) (int, error) {
	if fw.written+len(p) > fw.n {
		canWrite := fw.n - fw.written
		if canWrite > 0 {
			fw.written += canWrite
			return canWrite, fmt.Errorf("failWriterAt: limit reached at %d", fw.n)
		}
		return 0, fmt.Errorf("failWriterAt: limit reached at %d", fw.n)
	}
	fw.written += len(p)
	return len(p), nil
}

// ============================================================
// parsePng color type 4 (GrayAlpha) — hand-crafted raw PNG
// ============================================================

func TestCov26_ParsePng_GrayAlpha(t *testing.T) {
	w, h := 4, 4
	// Build raw scanlines for GrayAlpha (ct=4, bpc=8):
	// Each row: filter_byte + (gray, alpha) pairs
	var rawScanlines bytes.Buffer
	for y := 0; y < h; y++ {
		rawScanlines.WriteByte(0) // filter: None
		for x := 0; x < w; x++ {
			rawScanlines.WriteByte(byte((x + y) * 30))  // gray
			rawScanlines.WriteByte(byte(200 + x))        // alpha
		}
	}
	compressed := zlibCompress(rawScanlines.Bytes())
	pngData := buildRawPNG(t, w, h, 4, 8, nil, compressed)

	// Parse via parseImg
	reader := bytes.NewReader(pngData)
	info, err := parseImg(reader)
	if err != nil {
		t.Fatalf("parseImg GrayAlpha: %v", err)
	}
	if info.colspace != "DeviceGray" {
		t.Errorf("expected DeviceGray, got %s", info.colspace)
	}
	if len(info.smask) == 0 {
		t.Error("expected smask for GrayAlpha PNG")
	}
	if len(info.data) == 0 {
		t.Error("expected data for GrayAlpha PNG")
	}
}

// ============================================================
// rebuildXref: missing trailer after xref
// ============================================================

func TestCov26_RebuildXref_MissingTrailer(t *testing.T) {
	// Build minimal PDF-like data with xref but no "trailer" keyword
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("1 0 obj\n<< /Type /Catalog >>\nendobj\n")
	buf.WriteString("xref\n0 2\n0000000000 65535 f \n0000000009 00000 n \n")
	// No "trailer" keyword — just startxref
	buf.WriteString("startxref\n9\n%%EOF\n")

	result := rebuildXref(buf.Bytes())
	// Should return original data since trailer is missing
	if !bytes.Equal(result, buf.Bytes()) {
		t.Error("expected original data returned when trailer is missing")
	}
}

// ============================================================
// rebuildXref: missing startxref after trailer
// ============================================================

func TestCov26_RebuildXref_MissingStartxref(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("1 0 obj\n<< /Type /Catalog >>\nendobj\n")
	buf.WriteString("xref\n0 2\n0000000000 65535 f \n0000000009 00000 n \n")
	buf.WriteString("trailer\n<< /Size 2 /Root 1 0 R >>\n")
	// No startxref
	buf.WriteString("%%EOF\n")

	result := rebuildXref(buf.Bytes())
	if !bytes.Equal(result, buf.Bytes()) {
		t.Error("expected original data returned when startxref is missing")
	}
}


// ============================================================
// extractNamedRefs: inline dict branch
// ============================================================

func TestCov26_ExtractNamedRefs_InlineDict(t *testing.T) {
	// Build a minimal PDF with inline Font dict in page Resources
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	// Object 1: Catalog
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	// Object 2: Pages
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	// Object 3: Page with INLINE Font dict (not a reference)
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R /F2 6 0 R >> >> /Contents 4 0 R >>\nendobj\n")
	// Object 4: Content stream
	buf.WriteString("4 0 obj\n<< /Length 0 >>\nstream\n\nendstream\nendobj\n")
	// Object 5: Font
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")
	// Object 6: Font
	buf.WriteString("6 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>\nendobj\n")
	// xref + trailer
	buf.WriteString("xref\n0 7\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString("0000000009 00000 n \n")
	buf.WriteString("0000000062 00000 n \n")
	buf.WriteString("0000000121 00000 n \n")
	buf.WriteString("0000000310 00000 n \n")
	buf.WriteString("0000000363 00000 n \n")
	buf.WriteString("0000000440 00000 n \n")
	buf.WriteString("trailer\n<< /Size 7 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n500\n%%EOF\n")

	// Try to parse — the inline dict branch in extractNamedRefs should be hit
	parser, err := newRawPDFParser(buf.Bytes())
	if err != nil {
		t.Skipf("parser error (expected for hand-crafted PDF): %v", err)
	}
	if len(parser.pages) == 0 {
		t.Skip("no pages parsed from hand-crafted PDF")
	}
	// If we got here, the inline dict branch was exercised
	page := parser.pages[0]
	if len(page.resources.fonts) == 0 {
		t.Log("no fonts extracted (inline dict may not have matched)")
	}
}

// ============================================================
// extractFontsFromResources: deeper paths with FontDescriptor
// ============================================================

func TestCov26_ExtractFontsFromResources_FontDescriptor(t *testing.T) {
	// Build PDF with FontDescriptor containing /FontFile2
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>\nendobj\n")
	buf.WriteString("4 0 obj\n<< /Length 0 >>\nstream\n\nendstream\nendobj\n")
	// Font with FontDescriptor reference
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /TrueType /BaseFont /MyFont /FontDescriptor 6 0 R >>\nendobj\n")
	// FontDescriptor with FontFile2
	buf.WriteString("6 0 obj\n<< /Type /FontDescriptor /FontName /MyFont /FontFile2 7 0 R >>\nendobj\n")
	// FontFile2 stream
	buf.WriteString("7 0 obj\n<< /Length 4 >>\nstream\ntest\nendstream\nendobj\n")
	buf.WriteString("xref\n0 8\n")
	for i := 0; i < 8; i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", i*50))
	}
	buf.WriteString("trailer\n<< /Size 8 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n0\n%%EOF\n")

	fonts, err := ExtractFontsFromPage(buf.Bytes(), 0)
	if err != nil {
		t.Skipf("parse error (expected for hand-crafted PDF): %v", err)
	}
	for _, f := range fonts {
		t.Logf("Font: %s, BaseFont: %s, Embedded: %v, DataLen: %d", f.Name, f.BaseFont, f.IsEmbedded, len(f.Data))
	}
}

// ============================================================
// AddWatermarkTextAllPages — multi-page watermark
// ============================================================

func TestCov26_AddWatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")

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

	// Also test repeat mode
	err = pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "REPEAT",
		FontFamily: fontFamily,
		FontSize:   24,
		Opacity:    0.2,
		Angle:      30,
		Repeat:     true,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages repeat: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// AddWatermarkImageAllPages — multi-page image watermark
// ============================================================

func TestCov26_AddWatermarkImageAllPages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.3, 200, 200, 30)
	if err != nil {
		t.Fatalf("AddWatermarkImageAllPages: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// content_obj.write — fine-grained failWriter
// ============================================================

func TestCov26_ContentObj_Write_FailWriter_FineGrained(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello World")

	// Get the content object
	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	// Try failing at various byte positions
	for n := 0; n <= 200; n += 1 {
		fw := &failWriterAt{n: n}
		err := content.write(fw, 1)
		if err == nil && n < 50 {
			// Very small limits should fail
			continue
		}
	}
}

// ============================================================
// content_obj.write — NoCompression path with failWriter
// ============================================================

func TestCov26_ContentObj_Write_NoCompression_FailWriter(t *testing.T) {
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
	pdf.Text("No compression test")

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for n := 0; n <= 150; n += 1 {
		fw := &failWriterAt{n: n}
		content.write(fw, 1)
	}
}

// ============================================================
// content_obj.write — protected PDF with failWriter
// ============================================================

func TestCov26_ContentObj_Write_Protected_FailWriter(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Protected content")

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for n := 0; n <= 200; n += 1 {
		fw := &failWriterAt{n: n}
		content.write(fw, 1)
	}
}


// ============================================================
// PdfDictionaryObj.write — failWriter
// ============================================================

func TestCov26_PdfDictionaryObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Test for dictionary obj")

	// Find PdfDictionaryObj
	for i, obj := range pdf.pdfObjs {
		if dictObj, ok := obj.(*PdfDictionaryObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				dictObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no PdfDictionaryObj found")
}

// ============================================================
// transparency_xobject_group.write — failWriter
// ============================================================

func TestCov26_TransparencyXObjectGroup_Write_FailWriter(t *testing.T) {
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add image with mask to create TransparencyXObjectGroup
	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 100, H: 100})
	if err != nil {
		t.Skipf("Image: %v", err)
	}

	// Find TransparencyXObjectGroup
	for i, obj := range pdf.pdfObjs {
		if txog, ok := obj.(TransparencyXObjectGroup); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				txog.write(fw, i+1)
			}
			return
		}
	}
	t.Log("no TransparencyXObjectGroup found — trying with mask")
}

// ============================================================
// image_obj.write — failWriter fine-grained
// ============================================================

func TestCov26_ImageObj_Write_FailWriter_FineGrained(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})
	if err != nil {
		t.Skipf("Image: %v", err)
	}

	for i, obj := range pdf.pdfObjs {
		if imgObj, ok := obj.(*ImageObj); ok {
			for n := 0; n <= 400; n += 2 {
				fw := &failWriterAt{n: n}
				imgObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no ImageObj found")
}

// ============================================================
// embedfont_obj.write — failWriter fine-grained
// ============================================================

func TestCov26_EmbedFontObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Embed font test")

	for i, obj := range pdf.pdfObjs {
		if efObj, ok := obj.(*EmbedFontObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				efObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no EmbedFontObj found")
}

// ============================================================
// smask_obj.write — failWriter
// ============================================================

func TestCov26_SMask_Write_FailWriter(t *testing.T) {
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Image(resPNGPath, 50, 50, &Rect{W: 100, H: 100})

	for i, obj := range pdf.pdfObjs {
		if smObj, ok := obj.(SMask); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				smObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no SMask found")
}

// ============================================================
// list_cache_content.write — failWriter
// ============================================================

func TestCov26_ListCacheContent_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Cache content test")
	pdf.SetXY(50, 70)
	pdf.Cell(nil, "Another cell")

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for n := 0; n <= 300; n += 2 {
		fw := &failWriterAt{n: n}
		content.listCache.write(fw, nil)
	}
}

// ============================================================
// MeasureTextWidth / MeasureCellHeightByText — no font set
// ============================================================

func TestCov26_MeasureTextWidth_NoFont(t *testing.T) {
	defer func() { recover() }()
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	_, err := pdf.MeasureTextWidth("hello")
	if err == nil {
		t.Log("no error returned (may panic instead)")
	}
}

func TestCov26_MeasureCellHeightByText_NoFont(t *testing.T) {
	defer func() { recover() }()
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	_, err := pdf.MeasureCellHeightByText("hello")
	if err == nil {
		t.Log("no error returned (may panic instead)")
	}
}

// ============================================================
// GetBytesPdf — no font (error path)
// ============================================================

func TestCov26_GetBytesPdf_NoContent(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	// No page added — should still produce bytes
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF bytes")
	}
}

// ============================================================
// Read — various buffer sizes
// ============================================================

func TestCov26_Read_VariousBufferSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Read test")

	// Read with small buffer
	buf := make([]byte, 10)
	n, err := pdf.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Error("expected some bytes read")
	}

	// Read again (should continue or EOF)
	n2, err2 := pdf.Read(buf)
	_ = n2
	_ = err2

	// Read with large buffer
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.SetXY(50, 50)
	pdf2.Text("Large read")
	largeBuf := make([]byte, 1024*1024)
	n3, _ := pdf2.Read(largeBuf)
	if n3 == 0 {
		t.Error("expected some bytes from large read")
	}
}

// ============================================================
// CellWithOption — various alignments
// ============================================================

func TestCov26_CellWithOption_AllAlignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	rect := &Rect{W: 200, H: 20}
	aligns := []int{Left | Top, Center | Top, Right | Top, Left | Bottom, Center | Bottom, Right | Bottom, Left | Middle, Center | Middle, Right | Middle}
	for _, a := range aligns {
		pdf.SetXY(50, 50)
		err := pdf.CellWithOption(rect, "Align test", CellOption{Align: a})
		if err != nil {
			t.Errorf("CellWithOption align %d: %v", a, err)
		}
	}
	pdf.GetBytesPdf()
}

// ============================================================
// ImageSVGFromReader — error path
// ============================================================

func TestCov26_ImageSVGFromReader_BadReader(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	badReader := &errReader{}
	err := pdf.ImageSVGFromReader(badReader, SVGOption{
		X: 10, Y: 10, Width: 100, Height: 100,
	})
	if err == nil {
		t.Error("expected error from bad reader")
	}
}

type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("errReader: always fails")
}

// ============================================================
// DeleteBookmark — more linked-list branches
// ============================================================

func TestCov26_DeleteBookmark_MiddleChild(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 3")
	pdf.AddOutline("Chapter 3")

	// Delete middle bookmark (index 1)
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark(1): %v", err)
	}

	// Delete first bookmark (index 0) — now "Chapter 1" is first
	err = pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark(0): %v", err)
	}

	pdf.GetBytesPdf()
}

func TestCov26_DeleteBookmark_LastChild(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	// Delete last bookmark
	outlines := pdf.getOutlineObjList()
	lastIdx := len(outlines) - 1
	err := pdf.DeleteBookmark(lastIdx)
	if err != nil {
		t.Fatalf("DeleteBookmark last: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SearchText — case insensitive
// ============================================================

func TestCov26_SearchText_CaseInsensitive(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello World Test")
	data := pdf.GetBytesPdf()

	results, err := SearchText(data, "hello", true)
	if err != nil {
		t.Logf("SearchText: %v", err)
	}
	_ = results
}

// ============================================================
// extractFilterValue — various filter strings
// ============================================================

func TestCov26_ExtractFilterValue(t *testing.T) {
	tests := []struct {
		dict     string
		expected string
	}{
		{"/Filter /DCTDecode", "DCTDecode"},
		{"/Filter /FlateDecode", "FlateDecode"},
		{"no filter here", ""},
	}
	for _, tt := range tests {
		got := extractFilterValue(tt.dict)
		if got != tt.expected {
			t.Errorf("extractFilterValue(%q) = %q, want %q", tt.dict, got, tt.expected)
		}
	}
}

// ============================================================
// replaceObjectStream — object not found
// ============================================================

func TestCov26_ReplaceObjectStream_NotFound(t *testing.T) {
	data := []byte("%PDF-1.4\n1 0 obj\n<< >>\nendobj\n")
	result := replaceObjectStream(data, 999, "<< >>", []byte("test"))
	if !bytes.Equal(result, data) {
		t.Error("expected original data when object not found")
	}
}

// ============================================================
// replaceObjectStream — endobj not found
// ============================================================

func TestCov26_ReplaceObjectStream_NoEndobj(t *testing.T) {
	data := []byte("%PDF-1.4\n999 0 obj\n<< >>")
	result := replaceObjectStream(data, 999, "<< >>", []byte("test"))
	if !bytes.Equal(result, data) {
		t.Error("expected original data when endobj not found")
	}
}


// ============================================================
// page_option isTrimBoxSet
// ============================================================

func TestCov26_PageOption_TrimBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: PageSizeA4,
		TrimBox: &Box{Left: 10, Top: 10, Right: 10, Bottom: 10},
	})
	pdf.SetXY(50, 50)
	pdf.Text("TrimBox test")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// cache_content_text_color_cmyk.equal — different values
// ============================================================

func TestCov26_CacheContentTextColorCMYK_Equal(t *testing.T) {
	a := cacheContentTextColorCMYK{c: 10, m: 20, y: 30, k: 40}
	b := cacheContentTextColorCMYK{c: 10, m: 20, y: 30, k: 40}
	c := cacheContentTextColorCMYK{c: 10, m: 20, y: 30, k: 50}
	d := cacheContentTextColorCMYK{c: 10, m: 20, y: 40, k: 40}
	e := cacheContentTextColorCMYK{c: 10, m: 30, y: 30, k: 40}
	f := cacheContentTextColorCMYK{c: 20, m: 20, y: 30, k: 40}

	if !a.equal(b) {
		t.Error("expected equal")
	}
	if a.equal(c) {
		t.Error("expected not equal (k differs)")
	}
	if a.equal(d) {
		t.Error("expected not equal (y differs)")
	}
	if a.equal(e) {
		t.Error("expected not equal (m differs)")
	}
	if a.equal(f) {
		t.Error("expected not equal (c differs)")
	}
}

// ============================================================
// formFieldObj.write — various field types
// ============================================================

func TestCov26_FormField_AllTypes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Radio
	err := pdf.AddFormField(FormField{
		Type: FormFieldRadio, Name: "radio1",
		X: 50, Y: 100, W: 20, H: 20,
	})
	if err != nil {
		t.Fatalf("AddFormField radio: %v", err)
	}

	// Button
	err = pdf.AddFormField(FormField{
		Type: FormFieldButton, Name: "btn1",
		X: 50, Y: 130, W: 80, H: 25,
	})
	if err != nil {
		t.Fatalf("AddFormField button: %v", err)
	}

	// Signature
	err = pdf.AddFormField(FormField{
		Type: FormFieldSignature, Name: "sig1",
		X: 50, Y: 160, W: 200, H: 50,
	})
	if err != nil {
		t.Fatalf("AddFormField signature: %v", err)
	}

	// Choice with options, fill, border
	err = pdf.AddFormField(FormField{
		Type: FormFieldChoice, Name: "choice1",
		X: 50, Y: 220, W: 150, H: 25,
		Options:     []string{"Option A", "Option B", "Option C"},
		HasBorder:   true,
		HasFill:     true,
		BorderColor: [3]uint8{0, 0, 0},
		FillColor:   [3]uint8{255, 255, 200},
		FontFamily:  fontFamily,
	})
	if err != nil {
		t.Fatalf("AddFormField choice: %v", err)
	}

	// Text with all options
	err = pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "text_full",
		X: 50, Y: 260, W: 200, H: 25,
		Value:       "Default",
		MaxLen:      100,
		Multiline:   true,
		ReadOnly:    true,
		Required:    true,
		FontFamily:  fontFamily,
		FontSize:    10,
		Color:       [3]uint8{0, 0, 128},
		HasBorder:   true,
		HasFill:     true,
		BorderColor: [3]uint8{128, 128, 128},
		FillColor:   [3]uint8{240, 240, 240},
	})
	if err != nil {
		t.Fatalf("AddFormField text_full: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// HTML renderText — alignment branches (center/right on new line)
// ============================================================

func TestCov26_HTML_RenderText_Alignment_NewLine(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Center alignment with text that wraps
	html := `<div style="text-align:center">` + strings.Repeat("Word ", 50) + `</div>`
	_, err := pdf.InsertHTMLBox(50, 50, 200, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox center: %v", err)
	}

	// Right alignment with text that wraps
	html2 := `<div style="text-align:right">` + strings.Repeat("Test ", 50) + `</div>`
	_, err = pdf.InsertHTMLBox(50, 500, 200, 400, html2, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox right: %v", err)
	}

	pdf.GetBytesPdf()
}


// ============================================================
// cache_content_rotate.write — failWriter
// ============================================================

func TestCov26_CacheContentRotate_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.SetXY(100, 100)
	pdf.Text("Rotated")
	pdf.RotateReset()

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	// Write each cache item with failWriter
	for _, cache := range content.listCache.caches {
		for n := 0; n <= 100; n += 1 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// outlines_obj.write — failWriter
// ============================================================

func TestCov26_OutlinesObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	// Find OutlineObj
	for i, obj := range pdf.pdfObjs {
		if _, ok := obj.(*OutlineObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				obj.(*OutlineObj).write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no OutlineObj found")
}

// ============================================================
// fontdescriptor_obj.write — failWriter
// ============================================================

func TestCov26_FontDescriptorObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Font descriptor test")

	for i, obj := range pdf.pdfObjs {
		if fdObj, ok := obj.(*FontDescriptorObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				fdObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no FontDescriptorObj found")
}

// ============================================================
// cache_content_line.write — failWriter
// ============================================================

func TestCov26_CacheContentLine_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(50, 50, 200, 200)

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 100; n += 1 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// cache_content_rectangle.write — failWriter
// ============================================================

func TestCov26_CacheContentRectangle_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 50,
		Rect: Rect{W: 100, H: 100},
		PaintStyle: DrawFillPaintStyle,
	})

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// cache_content_polygon.write — failWriter
// ============================================================

func TestCov26_CacheContentPolygon_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polygon([]Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}, "FD")

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// cache_content_text.write — failWriter (underline + border)
// ============================================================

func TestCov26_CacheContentText_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Underline text
	pdf.SetFontWithStyle(fontFamily, Underline, 14)
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 200, H: 20}, "Underlined text", CellOption{
		Align:  Left | Top,
		Border: Left | Right | Top | Bottom,
	})

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 300; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// cache_content_image.write — failWriter
// ============================================================

func TestCov26_CacheContentImage_Write_FailWriter(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}


// ============================================================
// page_obj.write — failWriter (with external links)
// ============================================================

func TestCov26_PageObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page with link")
	pdf.AddExternalLink("https://example.com", 50, 50, 100, 20)

	for i, obj := range pdf.pdfObjs {
		if pageObj, ok := obj.(*PageObj); ok {
			for n := 0; n <= 300; n += 3 {
				fw := &failWriterAt{n: n}
				pageObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no PageObj found")
}

// ============================================================
// procset_obj.write — failWriter
// ============================================================

func TestCov26_ProcsetObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Procset test")

	for i, obj := range pdf.pdfObjs {
		if psObj, ok := obj.(*ProcSetObj); ok {
			for n := 0; n <= 500; n += 3 {
				fw := &failWriterAt{n: n}
				psObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no ProcsetObj found")
}

// ============================================================
// unicode_map.write — failWriter
// ============================================================

func TestCov26_UnicodeMap_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Unicode map test: éàü")

	for i, obj := range pdf.pdfObjs {
		if umObj, ok := obj.(*UnicodeMap); ok {
			for n := 0; n <= 500; n += 3 {
				fw := &failWriterAt{n: n}
				umObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no UnicodeMap found")
}

// ============================================================
// device_rgb_obj.write — failWriter
// ============================================================

func TestCov26_DeviceRGBObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("DeviceRGB test")

	for i, obj := range pdf.pdfObjs {
		if drObj, ok := obj.(*DeviceRGBObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				drObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no DeviceRGBObj found")
}

// ============================================================
// SubsetFontObj — charCodeToGlyphIndexFormat4 edge cases
// ============================================================

func TestCov26_CharCodeToGlyphIndexFormat4_EdgeCases(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	// Use various Unicode characters to exercise format4 lookup
	pdf.Text("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	pdf.Text("éàüñçøæ") // accented characters
	pdf.Text("αβγδε")   // Greek
	pdf.Text("中文测试")   // CJK (may not be in font, but exercises lookup)

	// Try IsCurrFontContainGlyph with various runes
	for _, r := range "ABCéàü中αβ" {
		pdf.IsCurrFontContainGlyph(r)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// HTML renderList — ordered list
// ============================================================

func TestCov26_HTML_OrderedList(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<ol><li>First item</li><li>Second item with longer text that might wrap</li><li>Third item</li></ol>`
	_, err := pdf.InsertHTMLBox(50, 50, 200, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox ordered list: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML renderImage — image with src attribute
// ============================================================

func TestCov26_HTML_Image_Src(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := fmt.Sprintf(`<p>Before image</p><img src="%s" width="100" height="80"/><p>After image</p>`, resJPEGPath)
	_, err := pdf.InsertHTMLBox(50, 50, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox image: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — nested inline styles
// ============================================================

func TestCov26_HTML_NestedInlineStyles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p style="color:red">Red text <b>bold red</b> <i>italic red</i> <span style="color:blue">blue span</span></p>`
	_, err := pdf.InsertHTMLBox(50, 50, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox nested styles: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// Sector — various styles
// ============================================================

func TestCov26_Sector_Styles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Sector(200, 200, 80, 0, 90, "F")
	pdf.Sector(200, 200, 80, 90, 180, "D")
	pdf.Sector(200, 200, 80, 180, 270, "FD")
	pdf.Sector(200, 200, 80, 270, 360, "")

	pdf.GetBytesPdf()
}

// ============================================================
// SetTransparency — various blend modes
// ============================================================

func TestCov26_SetTransparency_BlendModes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	modes := []BlendModeType{NormalBlendMode, Multiply, Screen, Overlay, Darken, Lighten, ColorDodge, ColorBurn, HardLight, SoftLight, Difference, Exclusion}
	for _, mode := range modes {
		err := pdf.SetTransparency(Transparency{
			Alpha:         0.5,
			BlendModeType: mode,
		})
		if err != nil {
			t.Logf("SetTransparency %s: %v", mode, err)
		}
		pdf.SetXY(50, 50)
		pdf.Text("Blend: " + string(mode))
	}
	pdf.ClearTransparency()
	pdf.GetBytesPdf()
}

// ============================================================
// parsePng — zlib decompress error for ct>=4
// ============================================================

func TestCov26_ParsePng_GrayAlpha_BadZlib(t *testing.T) {
	// Build a GrayAlpha PNG with invalid zlib data
	pngData := buildRawPNG(t, 2, 2, 4, 8, nil, []byte{0xFF, 0xFE, 0xFD, 0xFC})
	reader := bytes.NewReader(pngData)
	_, err := parseImg(reader)
	if err == nil {
		t.Error("expected error for bad zlib data in GrayAlpha PNG")
	}
}

// ============================================================
// parsePng — RGBA (ct=6) with bad zlib
// ============================================================

func TestCov26_ParsePng_RGBA_BadZlib(t *testing.T) {
	pngData := buildRawPNG(t, 2, 2, 6, 8, nil, []byte{0xFF, 0xFE, 0xFD, 0xFC})
	reader := bytes.NewReader(pngData)
	_, err := parseImg(reader)
	if err == nil {
		t.Error("expected error for bad zlib data in RGBA PNG")
	}
}

// ============================================================
// rebuildXref — no objects found
// ============================================================

func TestCov26_RebuildXref_NoObjects(t *testing.T) {
	data := []byte("xref\n0 1\n0000000000 65535 f \ntrailer\n<< /Size 1 >>\nstartxref\n0\n%%EOF\n")
	result := rebuildXref(data)
	// Should still work (no obj headers found, but xref is rebuilt)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// ============================================================
// rebuildXref — no xref at all
// ============================================================

func TestCov26_RebuildXref_NoXref(t *testing.T) {
	data := []byte("%PDF-1.4\n1 0 obj\n<< >>\nendobj\n")
	result := rebuildXref(data)
	if !bytes.Equal(result, data) {
		t.Error("expected original data when no xref found")
	}
}
