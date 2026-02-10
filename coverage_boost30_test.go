package gopdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost30_test.go — TestCov30_ prefix
// Targets: cacheContentText, color equal, extractNamedRefs,
// IncrementalSave, SearchText, rebuildXref, content_stream_clean,
// colorspace_convert, page_option, font_extract, pdf_parser
// ============================================================

// ============================================================
// cacheContentTextColorCMYK.equal — !ok branch
// ============================================================

func TestCov30_CacheContentTextColorCMYK_Equal_WrongType(t *testing.T) {
	cmyk := cacheContentTextColorCMYK{c: 10, m: 20, y: 30, k: 40}
	rgb := cacheContentTextColorRGB{r: 100, g: 200, b: 50}
	if cmyk.equal(rgb) {
		t.Error("expected false when comparing CMYK with RGB")
	}
	if cmyk.equal(nil) {
		t.Error("expected false when comparing CMYK with nil")
	}
}

// ============================================================
// cacheContentTextColorRGB.equal — !ok branch
// ============================================================

func TestCov30_CacheContentTextColorRGB_Equal_WrongType(t *testing.T) {
	rgb := cacheContentTextColorRGB{r: 100, g: 200, b: 50}
	cmyk := cacheContentTextColorCMYK{c: 10, m: 20, y: 30, k: 40}
	if rgb.equal(cmyk) {
		t.Error("expected false when comparing RGB with CMYK")
	}
	if rgb.equal(nil) {
		t.Error("expected false when comparing RGB with nil")
	}
}

// ============================================================
// cacheContentText.calX / calY — ErrContentTypeNotFound
// ============================================================

func TestCov30_CacheContentText_CalXY_InvalidContentType(t *testing.T) {
	ct := &cacheContentText{
		contentType: 999,
		pageheight:  841.89,
	}
	_, err := ct.calX()
	if err != ErrContentTypeNotFound {
		t.Errorf("calX: expected ErrContentTypeNotFound, got %v", err)
	}
	_, err = ct.calY()
	if err != ErrContentTypeNotFound {
		t.Errorf("calY: expected ErrContentTypeNotFound, got %v", err)
	}
}

func TestCov30_CacheContentText_Write_CalError(t *testing.T) {
	ct := &cacheContentText{contentType: 999, pageheight: 841.89}
	var buf bytes.Buffer
	err := ct.write(&buf, nil)
	if err != ErrContentTypeNotFound {
		t.Errorf("expected ErrContentTypeNotFound, got %v", err)
	}
}

// ============================================================
// cacheContentText.write — extGStateIndexes error branch
// ============================================================

func TestCov30_CacheContentText_Write_ExtGStateError(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Hello")

	ct := &cacheContentText{
		contentType: ContentTypeText, pageheight: 841.89,
		x: 50, y: 50, fontSubset: sf, fontCountIndex: 1,
		fontSize: 14, text: "Hello",
		cellOpt: CellOption{extGStateIndexes: []int{1, 2}},
	}
	fw := &failWriterAt{n: 5}
	err := ct.write(fw, nil)
	if err == nil {
		t.Error("expected error from failWriterAt")
	}
}

func TestCov30_CacheContentText_Write_BTError(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Hi")

	ct := &cacheContentText{
		contentType: ContentTypeText, pageheight: 841.89,
		x: 50, y: 50, fontSubset: sf, fontCountIndex: 1,
		fontSize: 14, text: "Hi",
	}
	fw := &failWriterAt{n: 0}
	err := ct.write(fw, nil)
	if err == nil {
		t.Error("expected error writing BT")
	}
}

// ============================================================
// cacheContentText.underline — fontSubset == nil
// ============================================================

func TestCov30_CacheContentText_Underline_NilFontSubset(t *testing.T) {
	ct := &cacheContentText{fontSubset: nil}
	var buf bytes.Buffer
	err := ct.underline(&buf)
	if err == nil || !strings.Contains(err.Error(), "not found font") {
		t.Errorf("expected 'not found font' error, got %v", err)
	}
}

func TestCov30_CacheContentText_Underline_WriteError(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Test")

	ct := &cacheContentText{
		fontSubset: sf, fontSize: 14, pageheight: 841.89,
		x: 50, y: 50, cellWidthPdfUnit: 100, cellHeightPdfUnit: 20,
		cellOpt: CellOption{
			CoefLineHeight: 1.5, CoefUnderlinePosition: 1.2,
			CoefUnderlineThickness: 0.8,
		},
	}
	fw := &failWriterAt{n: 0}
	err := ct.underline(fw)
	if err == nil {
		t.Error("expected error from underline write")
	}
}

// ============================================================
// cacheContentText.write — underline error propagation
// ============================================================

func TestCov30_CacheContentText_Write_UnderlineError(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Hi")

	ct := &cacheContentText{
		contentType: ContentTypeText, pageheight: 841.89,
		x: 50, y: 50, fontSubset: sf, fontCountIndex: 1,
		fontSize: 14, fontStyle: Underline, text: "Hi",
	}
	// Measure main body size without underline
	ct2 := *ct
	ct2.fontStyle = 0
	var measure bytes.Buffer
	ct2.write(&measure, nil)
	mainLen := measure.Len()

	// Fail right at underline
	fw := &failWriterAt{n: mainLen}
	err := ct.write(fw, nil)
	if err == nil {
		t.Error("expected error from underline in write")
	}
}

// ============================================================
// cacheContentText.drawBorder — all 4 borders + error paths
// ============================================================

func TestCov30_CacheContentText_DrawBorder_AllBorders(t *testing.T) {
	ct := &cacheContentText{
		pageheight: 841.89, x: 50, y: 50,
		cellWidthPdfUnit: 100, cellHeightPdfUnit: 20, lineWidth: 0.5,
		cellOpt: CellOption{Border: Top | Left | Right | Bottom},
	}
	var buf bytes.Buffer
	if err := ct.drawBorder(&buf); err != nil {
		t.Fatalf("drawBorder: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 border lines, got %d", len(lines))
	}
}

func TestCov30_CacheContentText_DrawBorder_Errors(t *testing.T) {
	for _, border := range []int{Top, Left, Right, Bottom} {
		ct := &cacheContentText{
			pageheight: 841.89, x: 50, y: 50,
			cellWidthPdfUnit: 100, cellHeightPdfUnit: 20, lineWidth: 0.5,
			cellOpt: CellOption{Border: border},
		}
		fw := &failWriterAt{n: 0}
		if err := ct.drawBorder(fw); err == nil {
			t.Errorf("expected error for border %d", border)
		}
	}
}

// ============================================================
// cacheContentText.isSame — various branches
// ============================================================

func TestCov30_CacheContentText_IsSame(t *testing.T) {
	c1 := cacheContentText{rectangle: &Rect{W: 10, H: 10}}
	c2 := cacheContentText{}
	if c1.isSame(c2) {
		t.Error("expected not same when rectangle != nil")
	}

	c3 := cacheContentText{fontCountIndex: 1, fontSize: 14, setXCount: 1, y: 100}
	c4 := cacheContentText{fontCountIndex: 1, fontSize: 14, setXCount: 1, y: 100}
	if !c3.isSame(c4) {
		t.Error("expected same for identical fields")
	}

	c5 := c4
	c5.fontCountIndex = 2
	if c3.isSame(c5) {
		t.Error("expected not same for different fontCountIndex")
	}

	c6 := cacheContentText{
		textColor: cacheContentTextColorRGB{r: 1, g: 2, b: 3},
		fontCountIndex: 1, fontSize: 14, setXCount: 1, y: 100,
	}
	c7 := cacheContentText{fontCountIndex: 1, fontSize: 14, setXCount: 1, y: 100}
	if c6.isSame(c7) {
		t.Error("expected not same when one textColor is nil")
	}

	c8 := cacheContentText{
		textColor: cacheContentTextColorCMYK{c: 1, m: 2, y: 3, k: 4},
		fontCountIndex: 1, fontSize: 14, setXCount: 1, y: 100,
	}
	if c6.isSame(c8) {
		t.Error("expected not same for different textColor types")
	}
}

// ============================================================
// cacheContentText.calY — Cell alignment branches
// ============================================================

func TestCov30_CacheContentText_CalY_CellAlignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset

	base := cacheContentText{
		contentType: ContentTypeCell, pageheight: 841.89,
		y: 100, cellHeightPdfUnit: 20, fontSubset: sf, fontSize: 14,
	}
	for _, align := range []int{Bottom, Middle, Top} {
		ct := base
		ct.cellOpt = CellOption{Align: align}
		y, err := ct.calY()
		if err != nil {
			t.Fatalf("calY align=%d: %v", align, err)
		}
		if y <= 0 {
			t.Errorf("calY align=%d: expected positive y", align)
		}
	}
}

func TestCov30_CacheContentText_CalX_CellAlignments(t *testing.T) {
	base := cacheContentText{
		contentType: ContentTypeCell, x: 50,
		cellWidthPdfUnit: 200, textWidthPdfUnit: 80,
	}
	for _, tc := range []struct {
		align    int
		expected float64
	}{
		{Right, 50 + 200 - 80},
		{Center, 50 + 200*0.5 - 80*0.5},
		{Left, 50},
	} {
		ct := base
		ct.cellOpt = CellOption{Align: tc.align}
		x, err := ct.calX()
		if err != nil {
			t.Fatalf("calX align=%d: %v", tc.align, err)
		}
		if x != tc.expected {
			t.Errorf("calX align=%d: expected %f, got %f", tc.align, tc.expected, x)
		}
	}
}

// ============================================================
// extractNamedRefs — inline dict branch
// ============================================================

func TestCov30_ExtractNamedRefs_InlineDict(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>\nendobj\n4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n5 0 obj\n<< /Length 44 >>\nstream\nBT /F1 12 Tf 100 700 Td (Hello World) Tj ET\nendstream\nendobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000306 00000 n \n0000000383 00000 n \ntrailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n479\n%%EOF\n"
	data := []byte(pdfContent)
	parser, err := newRawPDFParser(data)
	if err != nil {
		t.Fatalf("newRawPDFParser: %v", err)
	}
	if len(parser.pages) == 0 {
		t.Fatal("expected at least 1 page")
	}
	page := parser.pages[0]
	if len(page.resources.fonts) == 0 {
		t.Error("expected fonts from inline dict")
	}
}

// ============================================================
// WriteIncrementalPdf — error path (bad path)
// ============================================================

func TestCov30_WriteIncrementalPdf_BadPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("test")
	err := pdf.WriteIncrementalPdf("/nonexistent/dir/file.pdf", []byte("%PDF-1.4"), nil)
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// ============================================================
// SearchTextOnPage — various paths
// ============================================================

func TestCov30_SearchTextOnPage_InvalidPage(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 44 >>\nstream\nBT /F1 12 Tf 100 700 Td (Hello World) Tj ET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000232 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n328\n%%EOF\n"
	data := []byte(pdfContent)
	results, _ := SearchTextOnPage(data, 999, "Hello", false)
	if len(results) != 0 {
		t.Error("expected no results for invalid page")
	}
	results2, _ := SearchTextOnPage(data, -1, "Hello", false)
	if len(results2) != 0 {
		t.Error("expected no results for negative page")
	}
}

func TestCov30_SearchTextOnPage_CaseInsensitive(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 44 >>\nstream\nBT /F1 12 Tf 100 700 Td (Hello World) Tj ET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000232 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n328\n%%EOF\n"
	data := []byte(pdfContent)
	results, _ := SearchTextOnPage(data, 0, "hello", true)
	_ = results
}

func TestCov30_SearchText_MultiPage(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 44 >>\nstream\nBT /F1 12 Tf 100 700 Td (Hello World) Tj ET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000232 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n328\n%%EOF\n"
	data := []byte(pdfContent)
	results, _ := SearchText(data, "Hello", false)
	_ = results
}

func TestCov30_SearchText_BadData(t *testing.T) {
	_, err := SearchText([]byte("not a pdf"), "test", false)
	_ = err
	_, err = SearchTextOnPage([]byte("not a pdf"), 0, "test", false)
	_ = err
}

// ============================================================
// extractOperator — empty string
// ============================================================

func TestCov30_ExtractOperator_Empty(t *testing.T) {
	if extractOperator("") != "" {
		t.Error("expected empty")
	}
	if extractOperator("   ") != "" {
		t.Error("expected empty for whitespace")
	}
}

// ============================================================
// convertColorLine — all operator branches
// ============================================================

func TestCov30_ConvertColorLine_AllOps(t *testing.T) {
	tests := []struct {
		line   string
		target ColorspaceTarget
	}{
		{"0.5 0.3 0.1 rg", ColorspaceGray},
		{"0.5 0.3 0.1 RG", ColorspaceCMYK},
		{"0.1 0.2 0.3 0.4 k", ColorspaceRGB},
		{"0.1 0.2 0.3 0.4 K", ColorspaceGray},
		{"0.5 g", ColorspaceRGB},
		{"0.5 G", ColorspaceCMYK},
		{"0.5 0.3 0.1 rg", ColorspaceRGB},
		{"0.1 0.2 0.3 0.4 k", ColorspaceCMYK},
		{"0.5 g", ColorspaceGray},
	}
	for _, tt := range tests {
		result := convertColorLine(tt.line, tt.target)
		if result == "" {
			t.Errorf("convertColorLine(%q, %d) returned empty", tt.line, tt.target)
		}
	}
}

// ============================================================
// convertRGBOp / convertCMYKOp / convertGrayOp — stroking
// ============================================================

func TestCov30_ConvertOps_Stroking(t *testing.T) {
	r := convertRGBOp(0.5, 0.3, 0.1, true, ColorspaceGray)
	if !strings.HasSuffix(r, "G") {
		t.Errorf("expected stroking gray, got %q", r)
	}
	r2 := convertRGBOp(0.5, 0.3, 0.1, true, ColorspaceCMYK)
	if !strings.HasSuffix(r2, "K") {
		t.Errorf("expected stroking CMYK, got %q", r2)
	}
	r3 := convertCMYKOp(0.1, 0.2, 0.3, 0.4, true, ColorspaceGray)
	if !strings.HasSuffix(r3, "G") {
		t.Errorf("expected stroking gray, got %q", r3)
	}
	r4 := convertCMYKOp(0.1, 0.2, 0.3, 0.4, true, ColorspaceRGB)
	if !strings.HasSuffix(r4, "RG") {
		t.Errorf("expected stroking RGB, got %q", r4)
	}
	r5 := convertGrayOp(0.5, true, ColorspaceRGB)
	if !strings.HasSuffix(r5, "RG") {
		t.Errorf("expected stroking RGB, got %q", r5)
	}
	r6 := convertGrayOp(0.5, true, ColorspaceCMYK)
	if !strings.HasSuffix(r6, "K") {
		t.Errorf("expected stroking CMYK, got %q", r6)
	}
	r7 := convertGrayOp(0.5, true, ColorspaceGray)
	if !strings.HasSuffix(r7, "G") {
		t.Errorf("expected stroking gray, got %q", r7)
	}
}

// ============================================================
// rebuildXref — various branches
// ============================================================

func TestCov30_RebuildXref_MissingTrailer(t *testing.T) {
	data := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\nxref\n0 2\n0000000000 65535 f \n0000000009 00000 n \nno_trailer_here\nstartxref\n100\n%%EOF\n")
	result := rebuildXref(data)
	if !bytes.Equal(result, data) {
		t.Error("expected original data when trailer is missing")
	}
}

func TestCov30_RebuildXref_MissingStartxref(t *testing.T) {
	// trailer content without startxref — the function searches for "startxref"
	// within the trailer section. We need trailer present but no startxref after it.
	data := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\nxref\n0 2\n0000000000 65535 f \n0000000009 00000 n \ntrailer\n<< /Size 2 /Root 1 0 R >>\n%%EOF\n")
	result := rebuildXref(data)
	if !bytes.Equal(result, data) {
		t.Error("expected original data when startxref is missing")
	}
}

func TestCov30_RebuildXref_NoXref(t *testing.T) {
	data := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")
	result := rebuildXref(data)
	if !bytes.Equal(result, data) {
		t.Error("expected original data when no xref")
	}
}

func TestCov30_RebuildXref_NoObjects(t *testing.T) {
	data := []byte("%PDF-1.4\nxref\n0 1\n0000000000 65535 f \ntrailer\n<< /Size 1 /Root 1 0 R >>\nstartxref\n15\n%%EOF\n")
	result := rebuildXref(data)
	if !bytes.Equal(result, data) {
		t.Error("expected original data when no objects found")
	}
}

// ============================================================
// replaceObjectStream — object not found
// ============================================================

func TestCov30_ReplaceObjectStream_NotFound(t *testing.T) {
	data := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n")
	result := replaceObjectStream(data, 999, "<< /New >>", []byte("new stream"))
	if !bytes.Equal(result, data) {
		t.Error("expected original data when object not found")
	}
}

// ============================================================
// buildCleanedDict
// ============================================================

func TestCov30_BuildCleanedDict(t *testing.T) {
	dict := "<< /Length 100 >>"
	result := buildCleanedDict(dict, 50)
	if !strings.Contains(result, "/FlateDecode") {
		t.Error("expected /FlateDecode to be added")
	}
	if !strings.Contains(result, "/Length 50") {
		t.Error("expected /Length to be updated")
	}

	dict2 := "<< /Filter /FlateDecode /Length 100 >>"
	result2 := buildCleanedDict(dict2, 75)
	if strings.Count(result2, "FlateDecode") != 1 {
		t.Error("expected exactly one /FlateDecode")
	}
}

// ============================================================
// cleanContentStream / CleanContentStreams
// ============================================================

func TestCov30_CleanContentStream(t *testing.T) {
	stream := []byte("1.0 w\n2.0 w\nq\nQ\nBT\n/F1 12 Tf\n(Hello) Tj\nET\n")
	result := cleanContentStream(stream)
	if strings.Contains(string(result), "1.0 w") {
		t.Error("expected redundant 1.0 w to be removed")
	}
}

func TestCov30_RemoveEmptyQBlocks(t *testing.T) {
	lines := []string{"q", "Q", "BT", "ET"}
	result := removeEmptyQBlocks(lines)
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d", len(result))
	}
}

func TestCov30_RemoveRedundantStateChanges(t *testing.T) {
	lines := []string{"1.0 w", "2.0 w", "3.0 w", "BT", "ET"}
	result := removeRedundantStateChanges(lines)
	count := 0
	for _, l := range result {
		if strings.HasSuffix(l, "w") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 line width op, got %d", count)
	}
}

func TestCov30_NormalizeWhitespace(t *testing.T) {
	lines := []string{"  hello   world  ", "  foo  bar  "}
	result := normalizeWhitespace(lines)
	if result[0] != "hello world" {
		t.Errorf("expected 'hello world', got %q", result[0])
	}
}

func TestCov30_CleanContentStreams_FullPipeline(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 60 >>\nstream\n1.0 w\n2.0 w\nq\nQ\nBT /F1 12 Tf 100 700 Td (Hello) Tj ET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000232 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n344\n%%EOF\n"
	data := []byte(pdfContent)
	result, err := CleanContentStreams(data)
	if err != nil {
		t.Fatalf("CleanContentStreams: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// ============================================================
// GetFonts
// ============================================================

func TestCov30_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	fonts := pdf.GetFonts()
	if len(fonts) == 0 {
		t.Error("expected at least one font")
	}
}

// ============================================================
// cacheContentText.write — with txtColorMode="color"
// ============================================================

func TestCov30_CacheContentText_Write_WithColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Test")

	ct := &cacheContentText{
		contentType: ContentTypeText, pageheight: 841.89,
		x: 50, y: 50, fontSubset: sf, fontCountIndex: 1,
		fontSize: 14, text: "Test",
		txtColorMode: "color",
		textColor:    cacheContentTextColorRGB{r: 255, g: 0, b: 0},
	}
	var buf bytes.Buffer
	if err := ct.write(&buf, nil); err != nil {
		t.Fatalf("write with color: %v", err)
	}
	if !strings.Contains(buf.String(), "rg") {
		t.Error("expected color operator in output")
	}
}

// ============================================================
// cacheContentText.write — with kerning
// ============================================================

func TestCov30_CacheContentText_Write_WithKerning(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{UseKerning: true})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("AV")

	ct := &cacheContentText{
		contentType: ContentTypeText, pageheight: 841.89,
		x: 50, y: 50, fontSubset: sf, fontCountIndex: 1,
		fontSize: 14, text: "AV",
	}
	var buf bytes.Buffer
	if err := ct.write(&buf, nil); err != nil {
		t.Fatalf("write with kerning: %v", err)
	}
}

// ============================================================
// cacheContentText.write — Cell type with all alignments
// ============================================================

func TestCov30_CacheContentText_Write_CellType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Test")

	for _, align := range []int{Left | Top, Right | Top, Center | Top, Left | Bottom, Left | Middle} {
		ct := &cacheContentText{
			contentType: ContentTypeCell, pageheight: 841.89,
			x: 50, y: 50, fontSubset: sf, fontCountIndex: 1,
			fontSize: 14, text: "Test",
			cellWidthPdfUnit: 200, cellHeightPdfUnit: 20, textWidthPdfUnit: 60,
			cellOpt: CellOption{Align: align},
		}
		var buf bytes.Buffer
		if err := ct.write(&buf, nil); err != nil {
			t.Fatalf("write Cell align=%d: %v", align, err)
		}
	}
}

// ============================================================
// cacheContentTextColor write
// ============================================================

func TestCov30_CacheContentTextColor_Write(t *testing.T) {
	rgb := cacheContentTextColorRGB{r: 255, g: 128, b: 0}
	var buf bytes.Buffer
	if err := rgb.write(&buf, nil); err != nil {
		t.Fatalf("RGB write: %v", err)
	}

	cmyk := cacheContentTextColorCMYK{c: 10, m: 20, y: 30, k: 40}
	var buf2 bytes.Buffer
	if err := cmyk.write(&buf2, nil); err != nil {
		t.Fatalf("CMYK write: %v", err)
	}
}

// ============================================================
// FormatFloatTrim
// ============================================================

func TestCov30_FormatFloatTrim(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.0, "0"}, {1.0, "1"}, {1.5, "1.5"},
		{1.123, "1.123"}, {-0.5, "-0.5"},
	}
	for _, tt := range tests {
		if r := FormatFloatTrim(tt.input); r != tt.expected {
			t.Errorf("FormatFloatTrim(%f): expected %q, got %q", tt.input, tt.expected, r)
		}
	}
}

// ============================================================
// convertTypoUnit
// ============================================================

func TestCov30_ConvertTypoUnit(t *testing.T) {
	if convertTypoUnit(800, 2048, 14) <= 0 {
		t.Error("expected positive result")
	}
}

// ============================================================
// IncrementalSave — various branches
// ============================================================

func TestCov30_IncrementalSave_SpecificIndices(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello")
	original := pdf.GetBytesPdf()
	result, err := pdf.IncrementalSave(original, []int{0, 1})
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov30_IncrementalSave_OutOfRangeIndices(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello")
	original := pdf.GetBytesPdf()
	result, err := pdf.IncrementalSave(original, []int{-1, 0, 9999})
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov30_IncrementalSave_NoTrailingNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello")
	original := bytes.TrimRight(pdf.GetBytesPdf(), "\n\r")
	result, err := pdf.IncrementalSave(original, nil)
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov30_IncrementalSave_WithEncryption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello")
	pdf.encryptionObjID = 5
	original := pdf.GetBytesPdf()
	result, err := pdf.IncrementalSave(original, nil)
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if !bytes.Contains(result, []byte("/Encrypt")) {
		t.Error("expected /Encrypt in output")
	}
}

// ============================================================
// Read
// ============================================================

func TestCov30_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello")
	buf := make([]byte, 4096)
	n, err := pdf.Read(buf)
	if err != nil && n == 0 {
		t.Fatalf("Read: %v", err)
	}
	// Second read uses cached buffer
	buf2 := make([]byte, 4096)
	n2, _ := pdf.Read(buf2)
	_ = n2
}

// ============================================================
// extractDict — edge cases
// ============================================================

func TestCov30_ExtractDict_Unclosed(t *testing.T) {
	data := []byte("<< /Type /Catalog /Pages 2 0 R")
	result := extractDict(data)
	if result == "" {
		t.Error("expected partial dict for unclosed dict")
	}
}

func TestCov30_ExtractDict_NoDict(t *testing.T) {
	data := []byte("no dict here")
	result := extractDict(data)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

// ============================================================
// extractContentRefs — various forms
// ============================================================

func TestCov30_ExtractContentRefs_Array(t *testing.T) {
	refs := extractContentRefs("<< /Contents [4 0 R 5 0 R] >>")
	if len(refs) != 2 {
		t.Errorf("expected 2 refs, got %d", len(refs))
	}
}

func TestCov30_ExtractContentRefs_Single(t *testing.T) {
	refs := extractContentRefs("<< /Contents 4 0 R >>")
	if len(refs) != 1 {
		t.Errorf("expected 1 ref, got %d", len(refs))
	}
}

func TestCov30_ExtractContentRefs_None(t *testing.T) {
	refs := extractContentRefs("<< /Type /Page >>")
	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}

// ============================================================
// extractMediaBox — default
// ============================================================

func TestCov30_ExtractMediaBox_Default(t *testing.T) {
	box := extractMediaBox("<< /Type /Page >>")
	if box[2] != 612 || box[3] != 792 {
		t.Errorf("expected default letter size, got %v", box)
	}
}

// ============================================================
// extractRefArray / extractRef — edge cases
// ============================================================

func TestCov30_ExtractRefArray_NoKey(t *testing.T) {
	if extractRefArray("<< /Type /Page >>", "/Kids") != nil {
		t.Error("expected nil")
	}
}

func TestCov30_ExtractRefArray_NoBrackets(t *testing.T) {
	if extractRefArray("<< /Kids 3 0 R >>", "/Kids") != nil {
		t.Error("expected nil (no brackets)")
	}
}

func TestCov30_ExtractRef_NoKey(t *testing.T) {
	if extractRef("<< /Type /Page >>", "/Root") != 0 {
		t.Error("expected 0")
	}
}

// ============================================================
// isTrimBoxSet
// ============================================================

func TestCov30_IsTrimBoxSet(t *testing.T) {
	p1 := PageOption{}
	if p1.isTrimBoxSet() {
		t.Error("expected false for nil TrimBox")
	}
	p2 := PageOption{TrimBox: &Box{}}
	if p2.isTrimBoxSet() {
		t.Error("expected false for all-zero TrimBox")
	}
	p3 := PageOption{TrimBox: &Box{Top: 10, Left: 5, Bottom: 10, Right: 5}}
	if !p3.isTrimBoxSet() {
		t.Error("expected true for non-zero TrimBox")
	}
}

func TestCov30_PageOption_IsEmpty(t *testing.T) {
	if !(PageOption{}).isEmpty() {
		t.Error("expected empty")
	}
	if (PageOption{PageSize: &Rect{W: 100, H: 200}}).isEmpty() {
		t.Error("expected not empty")
	}
}

// ============================================================
// ExtractFontsFromPage — out of range
// ============================================================

func TestCov30_ExtractFontsFromPage_OutOfRange(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \ntrailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n200\n%%EOF\n"
	data := []byte(pdfContent)
	if _, err := ExtractFontsFromPage(data, 999); err == nil {
		t.Error("expected error for out of range page")
	}
	if _, err := ExtractFontsFromPage(data, -1); err == nil {
		t.Error("expected error for negative page")
	}
}

func TestCov30_ExtractFontsFromAllPages(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> >>\nendobj\n4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000260 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n340\n%%EOF\n"
	data := []byte(pdfContent)
	result, err := ExtractFontsFromAllPages(data)
	if err != nil {
		t.Fatalf("ExtractFontsFromAllPages: %v", err)
	}
	_ = result
}

// ============================================================
// cacheContentText.createContent
// ============================================================

func TestCov30_CacheContentText_CreateContent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("Hello World")

	ct := &cacheContentText{
		fontSubset: sf, fontSize: 14, text: "Hello World",
		rectangle: &Rect{W: 200, H: 20},
	}
	_, _, err := ct.createContent()
	if err != nil {
		t.Fatalf("createContent: %v", err)
	}
}

// ============================================================
// kern function
// ============================================================

func TestCov30_Kern(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{UseKerning: true})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	sf := pdf.curr.FontISubset
	sf.AddChars("AV")
	aIdx, _ := sf.CharCodeToGlyphIndex('A')
	vIdx, _ := sf.CharCodeToGlyphIndex('V')
	_ = kern(sf, 'A', 'V', aIdx, vIdx)
}

// ============================================================
// convertTTFUnit2PDFUnit / fixRange10 / ContentObjCalTextHeight
// ============================================================

func TestCov30_ConvertTTFUnit2PDFUnit(t *testing.T) {
	if convertTTFUnit2PDFUnit(100, 2048) == 0 {
		t.Error("expected non-zero")
	}
}

func TestCov30_FixRange10(t *testing.T) {
	if fixRange10(-0.5) != 0 {
		t.Error("expected 0 for negative")
	}
	if fixRange10(1.5) != 1 {
		t.Error("expected 1 for > 1")
	}
	if fixRange10(0.5) != 0.5 {
		t.Error("expected 0.5")
	}
}

func TestCov30_ContentObjCalTextHeight(t *testing.T) {
	if ContentObjCalTextHeight(14) <= 0 {
		t.Error("expected positive height")
	}
	if ContentObjCalTextHeightPrecise(14.5) <= 0 {
		t.Error("expected positive height")
	}
}

// ============================================================
// extractFilterValue / parseColorFloat / color conversions
// ============================================================

func TestCov30_ExtractFilterValue(t *testing.T) {
	if extractFilterValue("<< /Filter /FlateDecode /Length 100 >>") != "FlateDecode" {
		t.Error("expected FlateDecode")
	}
	if extractFilterValue("<< /Length 100 >>") != "" {
		t.Error("expected empty")
	}
}

func TestCov30_ParseColorFloat(t *testing.T) {
	if parseColorFloat("0.5") != 0.5 {
		t.Error("expected 0.5")
	}
	if parseColorFloat("invalid") != 0 {
		t.Error("expected 0")
	}
}

func TestCov30_ColorConversions(t *testing.T) {
	gray := rgbToGray(1.0, 0.0, 0.0)
	if gray < 0.29 || gray > 0.31 {
		t.Errorf("expected ~0.299, got %f", gray)
	}
	_, _, _, k := rgbToCMYK(0, 0, 0)
	if k != 1.0 {
		t.Errorf("expected k=1 for black, got %f", k)
	}
	r, g, b := cmykToRGB(0, 0, 0, 0)
	if r != 1.0 || g != 1.0 || b != 1.0 {
		t.Errorf("expected white, got %f %f %f", r, g, b)
	}
}

// ============================================================
// Suppress unused import warning
// ============================================================

var _ = fmt.Sprintf

// ============================================================
// extractNamedRefs — reference branch (else)
// ============================================================

func TestCov30_ExtractNamedRefs_ReferenceObj(t *testing.T) {
	// PDF where /Font is a reference to another object (not inline)
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources 6 0 R /Contents 5 0 R >>\nendobj\n4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n5 0 obj\n<< /Length 44 >>\nstream\nBT /F1 12 Tf 100 700 Td (Hello World) Tj ET\nendstream\nendobj\n6 0 obj\n<< /Font << /F1 4 0 R >> >>\nendobj\nxref\n0 7\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000250 00000 n \n0000000327 00000 n \n0000000423 00000 n \ntrailer\n<< /Size 7 /Root 1 0 R >>\nstartxref\n470\n%%EOF\n"
	data := []byte(pdfContent)
	parser, err := newRawPDFParser(data)
	if err != nil {
		t.Fatalf("newRawPDFParser: %v", err)
	}
	if len(parser.pages) == 0 {
		t.Fatal("expected at least 1 page")
	}
	page := parser.pages[0]
	if len(page.resources.fonts) == 0 {
		t.Error("expected fonts from reference obj")
	}
}

// ============================================================
// extractNamedRefs — empty rest after key
// ============================================================

func TestCov30_ExtractNamedRefs_EmptyRest(t *testing.T) {
	p := &rawPDFParser{objects: make(map[int]rawPDFObject)}
	out := make(map[string]int)
	// Dict where /Font is at the very end with nothing after
	p.extractNamedRefs("<< /Font", "/Font", out)
	if len(out) != 0 {
		t.Error("expected empty output for empty rest")
	}
}

// ============================================================
// getPageContentStream — various branches
// ============================================================

func TestCov30_GetPageContentStream_OutOfRange(t *testing.T) {
	p := &rawPDFParser{objects: make(map[int]rawPDFObject)}
	result := p.getPageContentStream(-1)
	if result != nil {
		t.Error("expected nil for negative index")
	}
	result2 := p.getPageContentStream(999)
	if result2 != nil {
		t.Error("expected nil for out of range index")
	}
}

// ============================================================
// parsePages — no root
// ============================================================

func TestCov30_ParsePages_NoRoot(t *testing.T) {
	p := &rawPDFParser{objects: make(map[int]rawPDFObject)}
	p.parsePages() // should not panic
	if len(p.pages) != 0 {
		t.Error("expected no pages")
	}
}

// ============================================================
// collectPages — missing object
// ============================================================

func TestCov30_CollectPages_MissingObj(t *testing.T) {
	p := &rawPDFParser{objects: make(map[int]rawPDFObject)}
	p.collectPages(999) // should not panic
	if len(p.pages) != 0 {
		t.Error("expected no pages")
	}
}

// ============================================================
// findRoot — no root ref
// ============================================================

func TestCov30_FindRoot_NoRoot(t *testing.T) {
	p := &rawPDFParser{
		data:    []byte("no root here"),
		objects: make(map[int]rawPDFObject),
	}
	p.findRoot()
	if p.root != 0 {
		t.Error("expected root=0")
	}
}

// ============================================================
// GetFonts — with FontObj (not just SubsetFontObj)
// ============================================================

func TestCov30_GetFonts_WithFontObj(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Manually add a FontObj to pdfObjs
	fontObj := &FontObj{
		Family:      "TestFont",
		IsEmbedFont: false,
	}
	pdf.pdfObjs = append(pdf.pdfObjs, fontObj)

	fonts := pdf.GetFonts()
	foundFontObj := false
	for _, f := range fonts {
		if f.Family == "TestFont" {
			foundFontObj = true
		}
	}
	if !foundFontObj {
		t.Error("expected to find FontObj in GetFonts")
	}
}

// ============================================================
// GetDocumentStats — comprehensive
// ============================================================

func TestCov30_GetDocumentStats(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Hello")

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Errorf("expected 1 page, got %d", stats.PageCount)
	}
	if stats.FontCount == 0 {
		t.Error("expected at least 1 font")
	}
	if stats.ContentStreamCount == 0 {
		t.Error("expected at least 1 content stream")
	}
}

// ============================================================
// FontObj.write — both branches (embedded and non-embedded)
// ============================================================

func TestCov30_FontObj_Write(t *testing.T) {
	// Non-embedded
	f := &FontObj{Family: "TestFont", IsEmbedFont: false}
	var buf bytes.Buffer
	if err := f.write(&buf, 1); err != nil {
		t.Fatalf("FontObj write: %v", err)
	}
	if !strings.Contains(buf.String(), "/BaseFont /TestFont") {
		t.Error("expected /BaseFont /TestFont")
	}

	// Embedded
	f2 := &FontObj{
		Family: "EmbedFont", IsEmbedFont: true,
		indexObjWidth: 5, indexObjFontDescriptor: 6, indexObjEncoding: 7,
	}
	var buf2 bytes.Buffer
	if err := f2.write(&buf2, 2); err != nil {
		t.Fatalf("FontObj write embedded: %v", err)
	}
	out := buf2.String()
	if !strings.Contains(out, "/FirstChar 32") {
		t.Error("expected /FirstChar 32")
	}
	if !strings.Contains(out, "/Widths 5 0 R") {
		t.Error("expected /Widths reference")
	}
}

// ============================================================
// FontObj.getType
// ============================================================

func TestCov30_FontObj_GetType(t *testing.T) {
	f := &FontObj{}
	if f.getType() != "Font" {
		t.Errorf("expected 'Font', got %q", f.getType())
	}
}

// ============================================================
// FontObj setters
// ============================================================

func TestCov30_FontObj_Setters(t *testing.T) {
	f := &FontObj{}
	f.SetIndexObjWidth(10)
	f.SetIndexObjFontDescriptor(20)
	f.SetIndexObjEncoding(30)
	if f.indexObjWidth != 10 || f.indexObjFontDescriptor != 20 || f.indexObjEncoding != 30 {
		t.Error("setters did not work")
	}
}

// ============================================================
// zlibDecompress — error path
// ============================================================

func TestCov30_ZlibDecompress_BadData(t *testing.T) {
	_, err := zlibDecompress([]byte("not zlib data"))
	if err == nil {
		t.Error("expected error for bad zlib data")
	}
}

// ============================================================
// splitContentLines
// ============================================================

func TestCov30_SplitContentLines(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n\n(Hello) Tj\nET\n")
	lines := splitContentLines(stream)
	// Empty lines should be removed
	for _, l := range lines {
		if l == "" {
			t.Error("expected no empty lines")
		}
	}
}

// ============================================================
// joinContentLines
// ============================================================

func TestCov30_JoinContentLines(t *testing.T) {
	lines := []string{"BT", "/F1 12 Tf", "ET"}
	result := joinContentLines(lines)
	if !strings.HasSuffix(string(result), "\n") {
		t.Error("expected trailing newline")
	}
}

// ============================================================
// TransparencyXObjectGroup — write error paths
// ============================================================

func TestCov30_TransparencyXObjectGroup_Write(t *testing.T) {
	group := TransparencyXObjectGroup{
		BBox: [4]float64{0, 0, 100, 100},
	}
	var buf bytes.Buffer
	if err := group.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), "/FormType 1") {
		t.Error("expected /FormType 1")
	}
}

func TestCov30_TransparencyXObjectGroup_WriteError(t *testing.T) {
	group := TransparencyXObjectGroup{
		BBox: [4]float64{0, 0, 100, 100},
	}
	fw := &failWriterAt{n: 5}
	err := group.write(fw, 1)
	if err == nil {
		t.Error("expected error from failWriterAt")
	}
}

func TestCov30_TransparencyXObjectGroup_GetType(t *testing.T) {
	g := TransparencyXObjectGroup{}
	if g.getType() != "XObject" {
		t.Errorf("expected 'XObject', got %q", g.getType())
	}
}

func TestCov30_TransparencyXObjectGroup_Protection(t *testing.T) {
	g := TransparencyXObjectGroup{}
	if g.protection() != nil {
		t.Error("expected nil protection")
	}
	p := &PDFProtection{}
	g.setProtection(p)
	if g.protection() != p {
		t.Error("expected set protection")
	}
}

// ============================================================
// SMask write — value receiver
// ============================================================

func TestCov30_SMask_Write_WithGroup(t *testing.T) {
	s := SMask{
		TransparencyXObjectGroupIndex: 5,
		S: SMaskAlphaSubtype,
	}
	var buf bytes.Buffer
	if err := s.write(&buf, 1); err != nil {
		t.Fatalf("SMask write: %v", err)
	}
	if !strings.Contains(buf.String(), "/Type /Mask") {
		t.Error("expected /Type /Mask")
	}
}

func TestCov30_SMask_Write_WithGroupError(t *testing.T) {
	s := SMask{
		TransparencyXObjectGroupIndex: 5,
		S: SMaskAlphaSubtype,
	}
	fw := &failWriterAt{n: 5}
	err := s.write(fw, 1)
	if err == nil {
		t.Error("expected error from failWriterAt")
	}
}

func TestCov30_SMask_Write_WithData(t *testing.T) {
	s := SMask{
		data: []byte("test image data"),
	}
	var buf bytes.Buffer
	if err := s.write(&buf, 1); err != nil {
		t.Fatalf("SMask write with data: %v", err)
	}
}

func TestCov30_SMask_GetType(t *testing.T) {
	s := SMask{}
	if s.getType() != "Mask" {
		t.Errorf("expected 'Mask', got %q", s.getType())
	}
}

func TestCov30_SMaskOptions_GetId(t *testing.T) {
	opts := SMaskOptions{
		TransparencyXObjectGroupIndex: 3,
		Subtype: SMaskAlphaSubtype,
	}
	id := opts.GetId()
	if id == "" {
		t.Error("expected non-empty id")
	}
}

// ============================================================
// ExtGState write
// ============================================================

func TestCov30_ExtGState_Write(t *testing.T) {
	alpha := 0.5
	bm := NormalBlendMode
	smaskIdx := 3
	g := ExtGState{
		Index: 1,
		ca:    &alpha,
		CA:    &alpha,
		BM:    &bm,
		SMaskIndex: &smaskIdx,
	}
	var buf bytes.Buffer
	if err := g.write(&buf, 1); err != nil {
		t.Fatalf("ExtGState write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/CA") {
		t.Error("expected /CA in output")
	}
	if !strings.Contains(out, "/ca") {
		t.Error("expected /ca in output")
	}
	if !strings.Contains(out, "/BM") {
		t.Error("expected /BM in output")
	}
	if !strings.Contains(out, "/SMask") {
		t.Error("expected /SMask in output")
	}
}

func TestCov30_ExtGState_WriteError(t *testing.T) {
	alpha := 0.5
	g := ExtGState{ca: &alpha, CA: &alpha}
	fw := &failWriterAt{n: 5}
	err := g.write(fw, 1)
	if err == nil {
		t.Error("expected error from failWriterAt")
	}
}

func TestCov30_ExtGState_GetType(t *testing.T) {
	g := ExtGState{}
	if g.getType() != "ExtGState" {
		t.Errorf("expected 'ExtGState', got %q", g.getType())
	}
}

func TestCov30_ExtGStateOptions_GetId(t *testing.T) {
	alpha := 0.5
	bm := NormalBlendMode
	smaskIdx := 3
	opts := ExtGStateOptions{
		StrokingCA:    &alpha,
		NonStrokingCa: &alpha,
		BlendMode:     &bm,
		SMaskIndex:    &smaskIdx,
	}
	id := opts.GetId()
	if id == "" {
		t.Error("expected non-empty id")
	}
}

// ============================================================
// digital_signature — LoadCertificateFromPEM error path
// ============================================================

func TestCov30_LoadCertificateFromPEM_BadPath(t *testing.T) {
	_, err := LoadCertificateFromPEM("/nonexistent/cert.pem")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestCov30_LoadPrivateKeyFromPEM_BadPath(t *testing.T) {
	_, err := LoadPrivateKeyFromPEM("/nonexistent/key.pem")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestCov30_LoadCertificateChainFromPEM_BadPath(t *testing.T) {
	_, err := LoadCertificateChainFromPEM("/nonexistent/chain.pem")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

// ============================================================
// ParseCertificatePEM — bad data
// ============================================================

func TestCov30_ParseCertificatePEM_NoPEM(t *testing.T) {
	_, err := ParseCertificatePEM([]byte("not pem data"))
	if err == nil {
		t.Error("expected error for non-PEM data")
	}
}

func TestCov30_ParsePrivateKeyPEM_NoPEM(t *testing.T) {
	_, err := ParsePrivateKeyPEM([]byte("not pem data"))
	if err == nil {
		t.Error("expected error for non-PEM data")
	}
}

func TestCov30_ParseCertificateChainPEM_NoPEM(t *testing.T) {
	_, err := ParseCertificateChainPEM([]byte("not pem data"))
	if err == nil {
		t.Error("expected error for non-PEM data")
	}
}

// ============================================================
// VerifySignatureFromFile — bad path
// ============================================================

func TestCov30_VerifySignatureFromFile_BadPath(t *testing.T) {
	_, err := VerifySignatureFromFile("/nonexistent/file.pdf")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

// ============================================================
// VerifySignature — bad data
// ============================================================

func TestCov30_VerifySignature_BadData(t *testing.T) {
	_, err := VerifySignature([]byte("not a pdf"))
	if err == nil {
		t.Error("expected error for bad PDF data")
	}
}

// ============================================================
// cacheContentPolyline.write — empty points
// ============================================================

func TestCov30_CacheContentPolyline_Write_Empty(t *testing.T) {
	p := &cacheContentPolyline{}
	var buf bytes.Buffer
	if err := p.write(&buf, nil); err != nil {
		t.Fatalf("write empty polyline: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected empty output for empty polyline")
	}
}

func TestCov30_CacheContentPolyline_Write_OnePoint(t *testing.T) {
	p := &cacheContentPolyline{
		points: []Point{{X: 10, Y: 20}},
	}
	var buf bytes.Buffer
	if err := p.write(&buf, nil); err != nil {
		t.Fatalf("write one-point polyline: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected empty output for single-point polyline")
	}
}

// ============================================================
// cacheContentSector.write — various styles
// ============================================================

func TestCov30_CacheContentSector_Write_Styles(t *testing.T) {
	for _, style := range []string{"F", "FD", "DF", "S"} {
		s := &cacheContentSector{
			pageHeight: 841.89, cx: 100, cy: 100, r: 50,
			startDeg: 0, endDeg: 90, style: style,
		}
		var buf bytes.Buffer
		if err := s.write(&buf, nil); err != nil {
			t.Fatalf("sector write style=%s: %v", style, err)
		}
	}
}

// ============================================================
// cacheContentOval.write
// ============================================================

func TestCov30_CacheContentOval_Write(t *testing.T) {
	o := &cacheContentOval{
		pageHeight: 841.89, x1: 50, y1: 50, x2: 150, y2: 100,
	}
	var buf bytes.Buffer
	if err := o.write(&buf, nil); err != nil {
		t.Fatalf("oval write: %v", err)
	}
}

// ============================================================
// cacheContentClipPolygon.write
// ============================================================

func TestCov30_CacheContentClipPolygon_Write(t *testing.T) {
	c := &cacheContentClipPolygon{
		pageHeight: 841.89,
		points:     []Point{{X: 10, Y: 10}, {X: 100, Y: 10}, {X: 100, Y: 100}},
	}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("clip polygon write: %v", err)
	}
}

// ============================================================
// cacheContentLineType.write — various types
// ============================================================

func TestCov30_CacheContentLineType_Write(t *testing.T) {
	for _, lt := range []string{"dashed", "dotted", "solid"} {
		c := &cacheContentLineType{lineType: lt}
		var buf bytes.Buffer
		if err := c.write(&buf, nil); err != nil {
			t.Fatalf("line type write %s: %v", lt, err)
		}
	}
}

// ============================================================
// cacheContentGray.write
// ============================================================

func TestCov30_CacheContentGray_Write(t *testing.T) {
	c := &cacheContentGray{scale: 0.5, grayType: "g"}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("gray write: %v", err)
	}
}

// ============================================================
// cacheContentLineWidth.write
// ============================================================

func TestCov30_CacheContentLineWidth_Write(t *testing.T) {
	c := &cacheContentLineWidth{width: 2.0}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("line width write: %v", err)
	}
}

// ============================================================
// cacheContentCustomLineType.write
// ============================================================

func TestCov30_CacheContentCustomLineType_Write(t *testing.T) {
	c := &cacheContentCustomLineType{
		dashArray: []float64{3, 2},
		dashPhase: 0,
	}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("custom line type write: %v", err)
	}
}

// ============================================================
// cacheContentColorCMYK.write
// ============================================================

func TestCov30_CacheContentColorCMYK_Write(t *testing.T) {
	c := &cacheContentColorCMYK{c: 10, m: 20, y: 30, k: 40, colorType: "k"}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("CMYK write: %v", err)
	}
}

// ============================================================
// cacheContentSaveGraphicsState / RestoreGraphicsState
// ============================================================

func TestCov30_CacheContentGraphicsState_Write(t *testing.T) {
	save := &cacheContentSaveGraphicsState{}
	var buf bytes.Buffer
	if err := save.write(&buf, nil); err != nil {
		t.Fatalf("save graphics state: %v", err)
	}
	if buf.String() != "q\n" {
		t.Errorf("expected 'q\\n', got %q", buf.String())
	}

	restore := &cacheContentRestoreGraphicsState{}
	var buf2 bytes.Buffer
	if err := restore.write(&buf2, nil); err != nil {
		t.Fatalf("restore graphics state: %v", err)
	}
	if buf2.String() != "Q\n" {
		t.Errorf("expected 'Q\\n', got %q", buf2.String())
	}
}

// ============================================================
// cacheContentImportedTemplate.write
// ============================================================

func TestCov30_CacheContentImportedTemplate_Write(t *testing.T) {
	c := &cacheContentImportedTemplate{
		pageHeight: 841.89, tplName: "/TPL1",
		scaleX: 1, scaleY: 1, tX: 0, tY: 0,
	}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("imported template write: %v", err)
	}
}

// ============================================================
// OpenPDFFromStream — error path (bad reader)
// ============================================================

func TestCov30_OpenPDFFromStream_BadReader(t *testing.T) {
	pdf := &GoPdf{}
	// errReadSeeker is defined in coverage_boost24_test.go
	er := &errReadSeeker{}
	rs := io.ReadSeeker(er)
	err := pdf.OpenPDFFromStream(&rs, nil)
	if err == nil {
		t.Error("expected error from bad reader")
	}
}

// ============================================================
// cacheContentPolygon.write — various styles
// ============================================================

func TestCov30_CacheContentPolygon_Write(t *testing.T) {
	for _, style := range []string{"F", "FD", "DF", "D"} {
		p := &cacheContentPolygon{
			pageHeight: 841.89,
			points:     []Point{{X: 10, Y: 10}, {X: 100, Y: 10}, {X: 100, Y: 100}},
			style:      style,
		}
		var buf bytes.Buffer
		if err := p.write(&buf, nil); err != nil {
			t.Fatalf("polygon write style=%s: %v", style, err)
		}
	}
}

// ============================================================
// cacheContentLine.write — with extGStateIndexes
// ============================================================

func TestCov30_CacheContentLine_Write_WithExtGState(t *testing.T) {
	c := &cacheContentLine{
		pageHeight: 841.89,
		x1: 10, y1: 10, x2: 100, y2: 100,
		opts: lineOptions{extGStateIndexes: []int{1, 2}},
	}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("line write: %v", err)
	}
	if !strings.Contains(buf.String(), "/GS1 gs") {
		t.Error("expected /GS1 gs in output")
	}
}

// ============================================================
// cacheContentPolyline.write — with extGStateIndexes
// ============================================================

func TestCov30_CacheContentPolyline_Write_WithExtGState(t *testing.T) {
	p := &cacheContentPolyline{
		pageHeight: 841.89,
		points:     []Point{{X: 10, Y: 10}, {X: 100, Y: 100}},
		opts:       polylineOptions{extGStateIndexes: []int{1}},
	}
	var buf bytes.Buffer
	if err := p.write(&buf, nil); err != nil {
		t.Fatalf("polyline write: %v", err)
	}
	if !strings.Contains(buf.String(), "/GS1 gs") {
		t.Error("expected /GS1 gs in output")
	}
}

// ============================================================
// cacheContentSector.write — with extGStateIndexes
// ============================================================

func TestCov30_CacheContentSector_Write_WithExtGState(t *testing.T) {
	s := &cacheContentSector{
		pageHeight: 841.89, cx: 100, cy: 100, r: 50,
		startDeg: 0, endDeg: 90, style: "F",
		opts: sectorOptions{extGStateIndexes: []int{1}},
	}
	var buf bytes.Buffer
	if err := s.write(&buf, nil); err != nil {
		t.Fatalf("sector write: %v", err)
	}
	if !strings.Contains(buf.String(), "/GS1 gs") {
		t.Error("expected /GS1 gs in output")
	}
}

// ============================================================
// cacheContentImage.write — basic
// ============================================================

func TestCov30_CacheContentImage_Write(t *testing.T) {
	c := &cacheContentImage{
		pageHeight: 841.89,
		rect:       Rect{W: 100, H: 100},
		x:          50, y: 50,
		index:      1,
	}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("image write: %v", err)
	}
}

// ============================================================
// cacheContentColorSpace.write
// ============================================================

func TestCov30_CacheColorSpace_Write(t *testing.T) {
	c := &cacheColorSpace{countOfSpaceColor: 1}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("color space write: %v", err)
	}
}

// ============================================================
// cacheContentColorRGB.write
// ============================================================

func TestCov30_CacheContentColorRGB_Write(t *testing.T) {
	c := &cacheContentColorRGB{r: 255, g: 128, b: 0, colorType: colorTypeFillRGB}
	var buf bytes.Buffer
	if err := c.write(&buf, nil); err != nil {
		t.Fatalf("color RGB write: %v", err)
	}
}

// ============================================================
// ParsePrivateKeyPEM — unsupported PEM type
// ============================================================

func TestCov30_ParsePrivateKeyPEM_UnsupportedType(t *testing.T) {
	// Create a PEM block with unsupported type
	pemData := []byte("-----BEGIN UNKNOWN KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\n-----END UNKNOWN KEY-----\n")
	_, err := ParsePrivateKeyPEM(pemData)
	if err == nil {
		t.Error("expected error for unsupported PEM type")
	}
	if !strings.Contains(err.Error(), "unsupported PEM block type") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================
// ParseCertificatePEM — bad certificate data
// ============================================================

func TestCov30_ParseCertificatePEM_BadCert(t *testing.T) {
	pemData := []byte("-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\n-----END CERTIFICATE-----\n")
	_, err := ParseCertificatePEM(pemData)
	if err == nil {
		t.Error("expected error for bad certificate data")
	}
}

// ============================================================
// ConvertColorspace — more branches
// ============================================================

func TestCov30_ConvertColorspace_NoModification(t *testing.T) {
	// PDF with no color operators — should return original data
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 30 >>\nstream\nBT /F1 12 Tf (Hi) Tj ET\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000232 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n314\n%%EOF\n"
	data := []byte(pdfContent)
	result, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceGray})
	if err != nil {
		t.Fatalf("ConvertColorspace: %v", err)
	}
	_ = result
}

// ============================================================
// ConvertColorspace — with color operators
// ============================================================

func TestCov30_ConvertColorspace_WithColors(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>\nendobj\n4 0 obj\n<< /Length 80 >>\nstream\n0.5 0.3 0.1 rg\nBT /F1 12 Tf 100 700 Td (Hello) Tj ET\n0.8 0.2 0.1 RG\n0.5 g\nendstream\nendobj\nxref\n0 5\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000232 00000 n \ntrailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n364\n%%EOF\n"
	data := []byte(pdfContent)
	for _, target := range []ColorspaceTarget{ColorspaceGray, ColorspaceCMYK, ColorspaceRGB} {
		result, err := ConvertColorspace(data, ConvertColorspaceOption{Target: target})
		if err != nil {
			t.Fatalf("ConvertColorspace target=%d: %v", target, err)
		}
		if len(result) == 0 {
			t.Error("expected non-empty result")
		}
	}
}

// ============================================================
// RecompressImages — no images
// ============================================================

func TestCov30_RecompressImages_NoImages(t *testing.T) {
	pdfContent := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \ntrailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n200\n%%EOF\n"
	data := []byte(pdfContent)
	result, err := RecompressImages(data, RecompressOption{})
	if err != nil {
		t.Fatalf("RecompressImages: %v", err)
	}
	if !bytes.Equal(result, data) {
		t.Error("expected original data when no images")
	}
}

// ============================================================
// RecompressOption.defaults
// ============================================================

func TestCov30_RecompressOption_Defaults(t *testing.T) {
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
// WatermarkOption.defaults
// ============================================================

func TestCov30_WatermarkOption_Defaults(t *testing.T) {
	opt := WatermarkOption{}
	opt.defaults()
	if opt.FontSize != 48 {
		t.Errorf("expected 48, got %f", opt.FontSize)
	}
	if opt.Angle != 45 {
		t.Errorf("expected 45, got %f", opt.Angle)
	}
	if opt.Opacity != 0.3 {
		t.Errorf("expected 0.3, got %f", opt.Opacity)
	}
}

// ============================================================
// AddWatermarkText — error paths
// ============================================================

func TestCov30_AddWatermarkText_EmptyText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{Text: "", FontFamily: fontFamily})
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestCov30_AddWatermarkText_NoFontFamily(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{Text: "test"})
	if err == nil {
		t.Error("expected error for missing font family")
	}
}

// ============================================================
// AddWatermarkImage — error paths
// ============================================================

func TestCov30_AddWatermarkImage_BadPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkImage("/nonexistent/image.jpg", 0.3, 0, 0, 0)
	if err == nil {
		t.Error("expected error for bad image path")
	}
}

// ============================================================
// extractName — various branches
// ============================================================

func TestCov30_ExtractName(t *testing.T) {
	dict := "<< /BaseFont /Helvetica /Subtype /Type1 >>"
	if extractName(dict, "/BaseFont") != "Helvetica" {
		t.Error("expected Helvetica")
	}
	if extractName(dict, "/Missing") != "" {
		t.Error("expected empty for missing key")
	}
}

// ============================================================
// html_parser — parseAttributes
// ============================================================

func TestCov30_ParseHTML_Attributes(t *testing.T) {
	nodes := parseHTML(`<p style="color:red" class="test">Hello</p>`)
	if len(nodes) == 0 {
		t.Error("expected at least one node")
	}
}

// ============================================================
// html_parser — skipComment
// ============================================================

func TestCov30_ParseHTML_SkipComment(t *testing.T) {
	nodes := parseHTML("<!-- comment --><p>Hello</p><!-- another -->")
	_ = nodes
}

func TestCov30_ParseHTML_UnclosedComment(t *testing.T) {
	nodes := parseHTML("<!-- unclosed comment")
	_ = nodes
}

// ============================================================
// html_parser — parseText
// ============================================================

func TestCov30_ParseText(t *testing.T) {
	nodes := parseHTML("<p>Hello <b>World</b></p>")
	if len(nodes) == 0 {
		t.Error("expected at least one node")
	}
}

func TestCov30_ParseHTML_Comment(t *testing.T) {
	nodes := parseHTML("<!-- comment --><p>Hello</p>")
	_ = nodes
}

func TestCov30_ParseHTML_SelfClosing(t *testing.T) {
	nodes := parseHTML("<br/><hr/><img src='test.jpg'/>")
	_ = nodes
}

// ============================================================
// collapseWhitespace / splitWords
// ============================================================

func TestCov30_CollapseWhitespace(t *testing.T) {
	if collapseWhitespace("  hello   world  ") != "hello world" {
		t.Error("unexpected result")
	}
	if collapseWhitespace("") != "" {
		t.Error("expected empty")
	}
}

func TestCov30_SplitWords(t *testing.T) {
	words := splitWords("hello world  foo")
	if len(words) != 3 {
		t.Errorf("expected 3 words, got %d", len(words))
	}
}

// ============================================================
// OpenPDFOption.box
// ============================================================

func TestCov30_OpenPDFOption_Box(t *testing.T) {
	var opt *OpenPDFOption
	if opt.box() != "/MediaBox" {
		t.Error("expected /MediaBox for nil opt")
	}
	opt2 := &OpenPDFOption{Box: "/CropBox"}
	if opt2.box() != "/CropBox" {
		t.Error("expected /CropBox")
	}
	opt3 := &OpenPDFOption{}
	if opt3.box() != "/MediaBox" {
		t.Error("expected /MediaBox for empty box")
	}
}

// ============================================================
// OpenPDF — bad path
// ============================================================

func TestCov30_OpenPDF_BadPath(t *testing.T) {
	pdf := &GoPdf{}
	err := pdf.OpenPDF("/nonexistent/file.pdf", nil)
	if err == nil {
		t.Error("expected error for bad path")
	}
}

// ============================================================
// OpenPDFFromBytes — bad data
// ============================================================

func TestCov30_OpenPDFFromBytes_BadData(t *testing.T) {
	pdf := &GoPdf{}
	err := pdf.OpenPDFFromBytes([]byte("not a pdf at all"), nil)
	if err == nil {
		t.Error("expected error for bad PDF data")
	}
}
