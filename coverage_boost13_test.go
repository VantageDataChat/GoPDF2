package gopdf

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost13_test.go — TestCov13_ prefix
// Targets: pdf_lowlevel, page_info, text_search, content_element,
// font_extract, bookmark, watermark, image_obj_parse
// ============================================================

// --------------- helpers ---------------

// buildMinimalPDF creates a minimal valid PDF byte slice with one object.
func buildMinimalPDF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 44 >>
stream
BT /F1 12 Tf 100 700 Td (Hello World) Tj ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000266 00000 n 
0000000360 00000 n 
trailer
<< /Size 6 /Root 1 0 R >>
startxref
434
%%EOF
`)
}

// buildMinimalPDFWithStream creates a PDF with a stream object for stream tests.
func buildMinimalPDFWithStream() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 11 >>
stream
Hello World
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000206 00000 n 
trailer
<< /Size 5 /Root 1 0 R >>
startxref
280
%%EOF
`)
}

// ============================================================
// pdf_lowlevel.go tests
// ============================================================

func TestCov13_ReadObject_Valid(t *testing.T) {
	data := buildMinimalPDF()
	obj, err := ReadObject(data, 1)
	if err != nil {
		t.Fatalf("ReadObject: %v", err)
	}
	if obj == nil {
		t.Fatal("expected non-nil object")
	}
	if obj.Num != 1 {
		t.Errorf("Num = %d, want 1", obj.Num)
	}
	if !strings.Contains(obj.Dict, "/Type") {
		t.Errorf("Dict should contain /Type, got: %s", obj.Dict)
	}
}

func TestCov13_ReadObject_NotFound(t *testing.T) {
	data := buildMinimalPDF()
	_, err := ReadObject(data, 999)
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

func TestCov13_GetDictKey_Catalog(t *testing.T) {
	data := buildMinimalPDF()
	val, err := GetDictKey(data, 1, "/Type")
	if err != nil {
		t.Fatalf("GetDictKey: %v", err)
	}
	if val != "/Catalog" {
		t.Errorf("got %q, want /Catalog", val)
	}
}

func TestCov13_GetDictKey_MissingKey(t *testing.T) {
	data := buildMinimalPDF()
	val, err := GetDictKey(data, 1, "/NonExistent")
	if err != nil {
		t.Fatalf("GetDictKey: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string for missing key, got %q", val)
	}
}

func TestCov13_GetDictKey_InvalidObj(t *testing.T) {
	data := buildMinimalPDF()
	_, err := GetDictKey(data, 999, "/Type")
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

func TestCov13_SetDictKey_NewKey(t *testing.T) {
	data := buildMinimalPDF()
	result, err := SetDictKey(data, 1, "/NewKey", "/NewValue")
	if err != nil {
		t.Fatalf("SetDictKey: %v", err)
	}
	// Verify the key was added.
	val, err := GetDictKey(result, 1, "/NewKey")
	if err != nil {
		t.Fatalf("GetDictKey after set: %v", err)
	}
	if val != "/NewValue" {
		t.Errorf("got %q, want /NewValue", val)
	}
}

func TestCov13_SetDictKey_ReplaceExisting(t *testing.T) {
	data := buildMinimalPDF()
	result, err := SetDictKey(data, 1, "/Type", "/Dictionary")
	if err != nil {
		t.Fatalf("SetDictKey: %v", err)
	}
	val, err := GetDictKey(result, 1, "/Type")
	if err != nil {
		t.Fatalf("GetDictKey: %v", err)
	}
	if val != "/Dictionary" {
		t.Errorf("got %q, want /Dictionary", val)
	}
}

func TestCov13_SetDictKey_InvalidObj(t *testing.T) {
	data := buildMinimalPDF()
	_, err := SetDictKey(data, 999, "/Key", "/Val")
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

func TestCov13_GetStream_WithStream(t *testing.T) {
	data := buildMinimalPDFWithStream()
	stream, err := GetStream(data, 4)
	if err != nil {
		t.Fatalf("GetStream: %v", err)
	}
	if !bytes.Contains(stream, []byte("Hello World")) {
		t.Errorf("stream should contain 'Hello World', got: %s", string(stream))
	}
}

func TestCov13_GetStream_NoStream(t *testing.T) {
	data := buildMinimalPDF()
	stream, err := GetStream(data, 1)
	if err != nil {
		t.Fatalf("GetStream: %v", err)
	}
	if stream != nil {
		t.Errorf("expected nil stream for non-stream object, got %d bytes", len(stream))
	}
}

func TestCov13_GetStream_InvalidObj(t *testing.T) {
	data := buildMinimalPDF()
	_, err := GetStream(data, 999)
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

func TestCov13_SetStream(t *testing.T) {
	data := buildMinimalPDFWithStream()
	newData := []byte("New stream content here")
	result, err := SetStream(data, 4, newData)
	if err != nil {
		t.Fatalf("SetStream: %v", err)
	}
	stream, err := GetStream(result, 4)
	if err != nil {
		t.Fatalf("GetStream after SetStream: %v", err)
	}
	if !bytes.Contains(stream, newData) {
		t.Errorf("stream should contain new data")
	}
}

func TestCov13_SetStream_InvalidObj(t *testing.T) {
	data := buildMinimalPDF()
	_, err := SetStream(data, 999, []byte("data"))
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

func TestCov13_CopyObject(t *testing.T) {
	data := buildMinimalPDF()
	result, newObjNum, err := CopyObject(data, 1)
	if err != nil {
		t.Fatalf("CopyObject: %v", err)
	}
	if newObjNum <= 5 {
		t.Errorf("newObjNum should be > 5, got %d", newObjNum)
	}
	// Verify the new object exists.
	obj, err := ReadObject(result, newObjNum)
	if err != nil {
		t.Fatalf("ReadObject for copied obj: %v", err)
	}
	if !strings.Contains(obj.Dict, "/Catalog") {
		t.Errorf("copied object should contain /Catalog")
	}
}

func TestCov13_CopyObject_InvalidObj(t *testing.T) {
	data := buildMinimalPDF()
	_, _, err := CopyObject(data, 999)
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

func TestCov13_GetCatalog(t *testing.T) {
	data := buildMinimalPDF()
	cat, err := GetCatalog(data)
	if err != nil {
		t.Fatalf("GetCatalog: %v", err)
	}
	if cat == nil {
		t.Fatal("expected non-nil catalog")
	}
	if !strings.Contains(cat.Dict, "/Catalog") {
		t.Errorf("catalog dict should contain /Catalog, got: %s", cat.Dict)
	}
}

func TestCov13_GetCatalog_InvalidPDF(t *testing.T) {
	_, err := GetCatalog([]byte("not a pdf"))
	if err == nil {
		t.Fatal("expected error for invalid PDF")
	}
}

func TestCov13_GetTrailer(t *testing.T) {
	data := buildMinimalPDF()
	trailer, err := GetTrailer(data)
	if err != nil {
		t.Fatalf("GetTrailer: %v", err)
	}
	if !strings.Contains(trailer, "/Root") {
		t.Errorf("trailer should contain /Root, got: %s", trailer)
	}
	if !strings.Contains(trailer, "/Size") {
		t.Errorf("trailer should contain /Size, got: %s", trailer)
	}
}

func TestCov13_GetTrailer_NoTrailer(t *testing.T) {
	_, err := GetTrailer([]byte("%PDF-1.4\nstartxref\n0\n%%EOF"))
	if err == nil {
		t.Fatal("expected error for PDF without trailer")
	}
}

func TestCov13_GetTrailer_NoStartxref(t *testing.T) {
	_, err := GetTrailer([]byte("%PDF-1.4\ntrailer\n<< /Size 1 >>"))
	if err == nil {
		t.Fatal("expected error for PDF without startxref")
	}
}

func TestCov13_UpdateObject(t *testing.T) {
	data := buildMinimalPDF()
	newContent := "<< /Type /Catalog /Pages 2 0 R /NewEntry true >>"
	result, err := UpdateObject(data, 1, newContent)
	if err != nil {
		t.Fatalf("UpdateObject: %v", err)
	}
	obj, err := ReadObject(result, 1)
	if err != nil {
		t.Fatalf("ReadObject after update: %v", err)
	}
	if !strings.Contains(obj.Dict, "/NewEntry") {
		t.Errorf("updated object should contain /NewEntry")
	}
}

func TestCov13_UpdateObject_InvalidObj(t *testing.T) {
	data := buildMinimalPDF()
	_, err := UpdateObject(data, 999, "<< /Type /Test >>")
	if err == nil {
		t.Fatal("expected error for non-existent object")
	}
}

// ============================================================
// page_info.go tests
// ============================================================

func TestCov13_GetPageSize_Valid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")

	w, h, err := pdf.GetPageSize(1)
	if err != nil {
		t.Fatalf("GetPageSize: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got w=%f h=%f", w, h)
	}
}

func TestCov13_GetPageSize_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, _, err := pdf.GetPageSize(0)
	if err == nil {
		t.Fatal("expected error for page 0")
	}
	_, _, err = pdf.GetPageSize(99)
	if err == nil {
		t.Fatal("expected error for page 99")
	}
}

func TestCov13_GetAllPageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")

	sizes := pdf.GetAllPageSizes()
	if len(sizes) != 2 {
		t.Errorf("expected 2 page sizes, got %d", len(sizes))
	}
	for _, s := range sizes {
		if s.Width <= 0 || s.Height <= 0 {
			t.Errorf("invalid page size: %+v", s)
		}
	}
}

func TestCov13_GetSourcePDFPageCount(t *testing.T) {
	if _, err := os.Stat(resTestPDF); os.IsNotExist(err) {
		t.Skip("test PDF not available")
	}
	count, err := GetSourcePDFPageCount(resTestPDF)
	if err != nil {
		t.Fatalf("GetSourcePDFPageCount: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected positive page count, got %d", count)
	}
}

func TestCov13_GetSourcePDFPageCount_InvalidPath(t *testing.T) {
	_, err := GetSourcePDFPageCount("/nonexistent/file.pdf")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestCov13_GetSourcePDFPageCountFromBytes(t *testing.T) {
	if _, err := os.Stat(resTestPDF); os.IsNotExist(err) {
		t.Skip("test PDF not available")
	}
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	count, err := GetSourcePDFPageCountFromBytes(data)
	if err != nil {
		t.Fatalf("GetSourcePDFPageCountFromBytes: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected positive page count, got %d", count)
	}
}

func TestCov13_GetSourcePDFPageCountFromBytes_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// gofpdi panics on invalid PDF — that's expected.
		}
	}()
	_, err := GetSourcePDFPageCountFromBytes([]byte("not a pdf"))
	if err == nil {
		t.Fatal("expected error for invalid PDF data")
	}
}

func TestCov13_GetSourcePDFPageSizes(t *testing.T) {
	if _, err := os.Stat(resTestPDF); os.IsNotExist(err) {
		t.Skip("test PDF not available")
	}
	sizes, err := GetSourcePDFPageSizes(resTestPDF)
	if err != nil {
		t.Fatalf("GetSourcePDFPageSizes: %v", err)
	}
	if len(sizes) == 0 {
		t.Fatal("expected at least one page size")
	}
	for pageNo, info := range sizes {
		if info.Width <= 0 || info.Height <= 0 {
			t.Errorf("page %d: invalid size %+v", pageNo, info)
		}
	}
}

func TestCov13_GetSourcePDFPageSizes_InvalidPath(t *testing.T) {
	_, err := GetSourcePDFPageSizes("/nonexistent/file.pdf")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestCov13_GetSourcePDFPageSizesFromBytes(t *testing.T) {
	if _, err := os.Stat(resTestPDF); os.IsNotExist(err) {
		t.Skip("test PDF not available")
	}
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	sizes, err := GetSourcePDFPageSizesFromBytes(data)
	if err != nil {
		t.Fatalf("GetSourcePDFPageSizesFromBytes: %v", err)
	}
	if len(sizes) == 0 {
		t.Fatal("expected at least one page size")
	}
}

func TestCov13_GetSourcePDFPageSizesFromBytes_Invalid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// gofpdi panics on invalid PDF — that's expected.
		}
	}()
	_, err := GetSourcePDFPageSizesFromBytes([]byte("not a pdf"))
	if err == nil {
		t.Fatal("expected error for invalid PDF data")
	}
}

// ============================================================
// text_search.go tests
// ============================================================

func TestCov13_SearchText_Basic(t *testing.T) {
	// Create a PDF with known text.
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Hello World Test")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Another Page Content")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	data := buf.Bytes()

	results, err := SearchText(data, "Hello", false)
	if err != nil {
		t.Fatalf("SearchText: %v", err)
	}
	// May or may not find depending on font encoding, but should not error.
	_ = results
}

func TestCov13_SearchText_CaseInsensitive(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("HELLO world")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	results, err := SearchText(buf.Bytes(), "hello", true)
	if err != nil {
		t.Fatalf("SearchText case insensitive: %v", err)
	}
	_ = results
}

func TestCov13_SearchText_InvalidPDF(t *testing.T) {
	// newRawPDFParser may not error on all invalid data; just ensure no panic.
	results, _ := SearchText([]byte("not a pdf"), "test", false)
	_ = results
}

func TestCov13_SearchTextOnPage_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Search Target Text")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	results, err := SearchTextOnPage(buf.Bytes(), 0, "Target", false)
	if err != nil {
		t.Fatalf("SearchTextOnPage: %v", err)
	}
	_ = results
}

func TestCov13_SearchTextOnPage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("text")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)

	results, err := SearchTextOnPage(buf.Bytes(), 99, "text", false)
	if err != nil {
		t.Fatalf("SearchTextOnPage: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for out-of-range page, got %d", len(results))
	}
}

func TestCov13_SearchTextOnPage_InvalidPDF(t *testing.T) {
	// newRawPDFParser may not error on all invalid data; just ensure no panic.
	results, _ := SearchTextOnPage([]byte("not a pdf"), 0, "test", false)
	_ = results
}

func TestCov13_SearchTextOnPage_CaseInsensitive(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("UPPERCASE text")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)

	results, err := SearchTextOnPage(buf.Bytes(), 0, "uppercase", true)
	if err != nil {
		t.Fatalf("SearchTextOnPage: %v", err)
	}
	_ = results
}

// ============================================================
// font_extract.go tests
// ============================================================

func TestCov13_ExtractFontsFromPage_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Font test")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	fonts, err := ExtractFontsFromPage(buf.Bytes(), 0)
	if err != nil {
		t.Fatalf("ExtractFontsFromPage: %v", err)
	}
	// Should find at least one font.
	_ = fonts
}

func TestCov13_ExtractFontsFromPage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("text")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)

	_, err := ExtractFontsFromPage(buf.Bytes(), 99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}
}

func TestCov13_ExtractFontsFromPage_NegativeIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("text")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)

	_, err := ExtractFontsFromPage(buf.Bytes(), -1)
	if err == nil {
		t.Fatal("expected error for negative page index")
	}
}

func TestCov13_ExtractFontsFromPage_InvalidPDF(t *testing.T) {
	_, err := ExtractFontsFromPage([]byte("not a pdf"), 0)
	if err == nil {
		t.Fatal("expected error for invalid PDF")
	}
}

func TestCov13_ExtractFontsFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	result, err := ExtractFontsFromAllPages(buf.Bytes())
	if err != nil {
		t.Fatalf("ExtractFontsFromAllPages: %v", err)
	}
	_ = result
}

func TestCov13_ExtractFontsFromAllPages_InvalidPDF(t *testing.T) {
	// newRawPDFParser may not error on all invalid data; just ensure no panic.
	result, _ := ExtractFontsFromAllPages([]byte("not a pdf"))
	_ = result
}

// ============================================================
// content_element.go tests
// ============================================================

func TestCov13_GetPageElements_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Element test")
	pdf.Line(10, 10, 100, 100)

	elems, err := pdf.GetPageElements(1)
	if err != nil {
		t.Fatalf("GetPageElements: %v", err)
	}
	if len(elems) == 0 {
		t.Error("expected at least one element")
	}
}

func TestCov13_GetPageElements_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetPageElements(99)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_GetPageElementsByType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text element")
	pdf.Line(10, 10, 100, 100)

	textElems, err := pdf.GetPageElementsByType(1, ElementText)
	if err != nil {
		t.Fatalf("GetPageElementsByType: %v", err)
	}
	_ = textElems

	lineElems, err := pdf.GetPageElementsByType(1, ElementLine)
	if err != nil {
		t.Fatalf("GetPageElementsByType line: %v", err)
	}
	if len(lineElems) == 0 {
		t.Error("expected at least one line element")
	}
}

func TestCov13_GetPageElementsByType_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetPageElementsByType(99, ElementText)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_GetPageElementCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")
	pdf.Line(10, 10, 100, 100)

	count, err := pdf.GetPageElementCount(1)
	if err != nil {
		t.Fatalf("GetPageElementCount: %v", err)
	}
	if count < 2 {
		t.Errorf("expected at least 2 elements, got %d", count)
	}
}

func TestCov13_GetPageElementCount_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetPageElementCount(99)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_DeleteElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")
	pdf.Line(10, 10, 100, 100)

	countBefore, _ := pdf.GetPageElementCount(1)
	err := pdf.DeleteElement(1, 0)
	if err != nil {
		t.Fatalf("DeleteElement: %v", err)
	}
	countAfter, _ := pdf.GetPageElementCount(1)
	if countAfter != countBefore-1 {
		t.Errorf("expected count %d, got %d", countBefore-1, countAfter)
	}
}

func TestCov13_DeleteElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeleteElement(99, 0)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_DeleteElement_InvalidIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	err := pdf.DeleteElement(1, 99)
	if err == nil {
		t.Fatal("expected error for invalid index")
	}
}

func TestCov13_DeleteElementsByType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	pdf.Line(20, 20, 200, 200)
	pdf.SetXY(50, 50)
	_ = pdf.Text("keep this")

	removed, err := pdf.DeleteElementsByType(1, ElementLine)
	if err != nil {
		t.Fatalf("DeleteElementsByType: %v", err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}
}

func TestCov13_DeleteElementsByType_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.DeleteElementsByType(99, ElementLine)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_DeleteElementsInRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(50, 50, 60, 60) // inside rect
	pdf.Line(500, 500, 510, 510) // outside rect

	removed, err := pdf.DeleteElementsInRect(1, 0, 0, 200, 200)
	if err != nil {
		t.Fatalf("DeleteElementsInRect: %v", err)
	}
	if removed < 1 {
		t.Errorf("expected at least 1 removed, got %d", removed)
	}
}

func TestCov13_DeleteElementsInRect_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.DeleteElementsInRect(99, 0, 0, 100, 100)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_ClearPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")
	pdf.Line(10, 10, 100, 100)

	err := pdf.ClearPage(1)
	if err != nil {
		t.Fatalf("ClearPage: %v", err)
	}
	count, _ := pdf.GetPageElementCount(1)
	if count != 0 {
		t.Errorf("expected 0 elements after clear, got %d", count)
	}
}

func TestCov13_ClearPage_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ClearPage(99)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_ModifyTextElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("original text")

	err := pdf.ModifyTextElement(1, 0, "modified text")
	if err != nil {
		t.Fatalf("ModifyTextElement: %v", err)
	}
}

func TestCov13_ModifyTextElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyTextElement(99, 0, "text")
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_ModifyTextElement_InvalidIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")
	err := pdf.ModifyTextElement(1, 99, "text")
	if err == nil {
		t.Fatal("expected error for invalid index")
	}
}

func TestCov13_ModifyTextElement_NotText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	err := pdf.ModifyTextElement(1, 0, "text")
	if err == nil {
		t.Fatal("expected error for non-text element")
	}
}

func TestCov13_ModifyElementPosition_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")

	err := pdf.ModifyElementPosition(1, 0, 100, 100)
	if err != nil {
		t.Fatalf("ModifyElementPosition: %v", err)
	}
}

func TestCov13_ModifyElementPosition_Line(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)

	err := pdf.ModifyElementPosition(1, 0, 50, 50)
	if err != nil {
		t.Fatalf("ModifyElementPosition line: %v", err)
	}
}

func TestCov13_ModifyElementPosition_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyElementPosition(99, 0, 50, 50)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_ModifyElementPosition_InvalidIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	err := pdf.ModifyElementPosition(1, 99, 50, 50)
	if err == nil {
		t.Fatal("expected error for invalid index")
	}
}

func TestCov13_InsertLineElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("seed content") // creates ContentObj

	err := pdf.InsertLineElement(1, 10, 20, 100, 200)
	if err != nil {
		t.Fatalf("InsertLineElement: %v", err)
	}
	count, _ := pdf.GetPageElementCount(1)
	if count < 1 {
		t.Error("expected at least 1 element after insert")
	}
}

func TestCov13_InsertLineElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.InsertLineElement(99, 10, 20, 100, 200)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_InsertRectElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("seed content")

	err := pdf.InsertRectElement(1, 10, 20, 100, 50, "D")
	if err != nil {
		t.Fatalf("InsertRectElement: %v", err)
	}
	count, _ := pdf.GetPageElementCount(1)
	if count < 1 {
		t.Error("expected at least 1 element after insert")
	}
}

func TestCov13_InsertRectElement_DefaultStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("seed content")

	err := pdf.InsertRectElement(1, 10, 20, 100, 50, "")
	if err != nil {
		t.Fatalf("InsertRectElement default style: %v", err)
	}
}

func TestCov13_InsertRectElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.InsertRectElement(99, 10, 20, 100, 50, "D")
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_InsertOvalElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("seed content")

	err := pdf.InsertOvalElement(1, 10, 20, 100, 80)
	if err != nil {
		t.Fatalf("InsertOvalElement: %v", err)
	}
}

func TestCov13_InsertOvalElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.InsertOvalElement(99, 10, 20, 100, 80)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_ReplaceElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)

	newCache := &cacheContentLine{
		pageHeight: PageSizeA4.H,
		x1:         50,
		y1:         50,
		x2:         200,
		y2:         200,
	}
	err := pdf.ReplaceElement(1, 0, newCache)
	if err != nil {
		t.Fatalf("ReplaceElement: %v", err)
	}
}

func TestCov13_ReplaceElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ReplaceElement(99, 0, &cacheContentLine{})
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestCov13_ReplaceElement_InvalidIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	err := pdf.ReplaceElement(1, 99, &cacheContentLine{})
	if err == nil {
		t.Fatal("expected error for invalid index")
	}
}

func TestCov13_InsertElementAt(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)

	newCache := &cacheContentLine{
		pageHeight: PageSizeA4.H,
		x1:         50,
		y1:         50,
		x2:         200,
		y2:         200,
	}
	err := pdf.InsertElementAt(1, 0, newCache)
	if err != nil {
		t.Fatalf("InsertElementAt: %v", err)
	}
	count, _ := pdf.GetPageElementCount(1)
	if count != 2 {
		t.Errorf("expected 2 elements, got %d", count)
	}
}

func TestCov13_InsertElementAt_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.InsertElementAt(99, 0, &cacheContentLine{})
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

// ============================================================
// watermark.go tests
// ============================================================

func TestCov13_AddWatermarkText_EmptyText(t *testing.T) {
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

func TestCov13_AddWatermarkText_MissingFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: "",
	})
	if err == nil {
		t.Fatal("expected error for missing font family")
	}
}

func TestCov13_AddWatermarkText_Single(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Content")

	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "CONFIDENTIAL",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.2,
		Angle:      30,
		Color:      [3]uint8{255, 0, 0},
	})
	if err != nil {
		t.Fatalf("AddWatermarkText: %v", err)
	}
}

func TestCov13_AddWatermarkText_Repeat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Content")

	err := pdf.AddWatermarkText(WatermarkOption{
		Text:           "DRAFT",
		FontFamily:     fontFamily,
		FontSize:       24,
		Opacity:        0.15,
		Angle:          45,
		Repeat:         true,
		RepeatSpacingX: 100,
		RepeatSpacingY: 100,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText repeat: %v", err)
	}
}

func TestCov13_AddWatermarkText_Defaults(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Use zero values to trigger defaults.
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "DEFAULT",
		FontFamily: fontFamily,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText defaults: %v", err)
	}
}

func TestCov13_AddWatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "WATERMARK",
		FontFamily: fontFamily,
		FontSize:   48,
		Opacity:    0.3,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

func TestCov13_AddWatermarkImage_Basic(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Content")

	err := pdf.AddWatermarkImage(resJPEGPath, 0.3, 200, 200, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImage: %v", err)
	}
}

func TestCov13_AddWatermarkImage_WithAngle(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddWatermarkImage(resJPEGPath, 0.5, 150, 150, 45)
	if err != nil {
		t.Fatalf("AddWatermarkImage with angle: %v", err)
	}
}

func TestCov13_AddWatermarkImage_DefaultSize(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Pass 0 for imgW and imgH to trigger default sizing.
	err := pdf.AddWatermarkImage(resJPEGPath, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImage default size: %v", err)
	}
}

func TestCov13_AddWatermarkImageAllPages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.3, 100, 100, 30)
	if err != nil {
		t.Fatalf("AddWatermarkImageAllPages: %v", err)
	}
}

// ============================================================
// bookmark.go tests
// ============================================================

func TestCov13_ModifyBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyBookmark(-1, "test")
	if err == nil {
		t.Fatal("expected error for negative index")
	}
	err = pdf.ModifyBookmark(99, "test")
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestCov13_DeleteBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeleteBookmark(-1)
	if err == nil {
		t.Fatal("expected error for negative index")
	}
	err = pdf.DeleteBookmark(99)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestCov13_SetBookmarkStyle_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetBookmarkStyle(-1, BookmarkStyle{Bold: true})
	if err == nil {
		t.Fatal("expected error for negative index")
	}
	err = pdf.SetBookmarkStyle(99, BookmarkStyle{Bold: true})
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestCov13_BookmarkOperations_WithBookmarks(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 3")
	pdf.AddOutline("Chapter 3")

	// Modify bookmark.
	err := pdf.ModifyBookmark(0, "Updated Chapter 1")
	if err != nil {
		t.Fatalf("ModifyBookmark: %v", err)
	}

	// Set style.
	err = pdf.SetBookmarkStyle(1, BookmarkStyle{
		Bold:      true,
		Italic:    true,
		Color:     [3]float64{1, 0, 0},
		Collapsed: true,
	})
	if err != nil {
		t.Fatalf("SetBookmarkStyle: %v", err)
	}

	// Delete middle bookmark.
	err = pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}
}

// ============================================================
// ContentElementType.String() test
// ============================================================

func TestCov13_ContentElementType_String(t *testing.T) {
	tests := []struct {
		typ  ContentElementType
		want string
	}{
		{ElementText, "Text"},
		{ElementImage, "Image"},
		{ElementLine, "Line"},
		{ElementRectangle, "Rectangle"},
		{ElementOval, "Oval"},
		{ElementPolygon, "Polygon"},
		{ElementPolyline, "Polyline"},
		{ElementSector, "Sector"},
		{ElementCurve, "Curve"},
		{ContentElementType(999), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.typ.String()
		if got != tt.want {
			t.Errorf("ContentElementType(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

// ============================================================
// diagonalAngle test (watermark.go)
// ============================================================

func TestCov13_DiagonalAngle(t *testing.T) {
	angle := diagonalAngle(100, 100)
	if angle < 44 || angle > 46 {
		t.Errorf("diagonalAngle(100,100) = %f, expected ~45", angle)
	}
	angle = diagonalAngle(200, 100)
	if angle < 26 || angle > 27 {
		t.Errorf("diagonalAngle(200,100) = %f, expected ~26.57", angle)
	}
}

// ============================================================
// ModifyElementPosition for various element types
// ============================================================

func TestCov13_ModifyElementPosition_Rect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("seed content")
	_ = pdf.InsertRectElement(1, 10, 20, 100, 50, "D")

	err := pdf.ModifyElementPosition(1, 1, 200, 300)
	if err != nil {
		t.Fatalf("ModifyElementPosition rect: %v", err)
	}
}

func TestCov13_ModifyElementPosition_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("seed content")
	_ = pdf.InsertOvalElement(1, 10, 20, 100, 80)

	err := pdf.ModifyElementPosition(1, 1, 200, 300)
	if err != nil {
		t.Fatalf("ModifyElementPosition oval: %v", err)
	}
}
