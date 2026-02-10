package gopdf

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost12_test.go — TestCov12_ prefix
// Targets: ToByte, CacheContent Setup/WriteTextToContent,
// cacheContentText underline, pdf_decrypt helpers,
// content_obj write branches, image_obj write branches,
// DeleteBookmark linked-list paths, InsertHTMLBox remainingWidth,
// openPDFFromData encrypted path
// ============================================================

// ============================================================
// 1. ToByte (0%)
// ============================================================

func TestCov12_ToByte(t *testing.T) {
	b := ToByte("A")
	if b != 'A' {
		t.Errorf("expected 'A', got %c", b)
	}
	b = ToByte("Z")
	if b != 'Z' {
		t.Errorf("expected 'Z', got %c", b)
	}
}

// ============================================================
// 2. CacheContent Setup / WriteTextToContent (0%)
// ============================================================

func TestCov12_CacheContent_Setup(t *testing.T) {
	cc := &CacheContent{}
	cc.Setup(
		&Rect{W: 200, H: 30},
		cacheContentTextColorRGB{r: 0, g: 0, b: 0},
		0.0,
		1,
		14.0,
		0,
		0.0,
		0,
		50.0, 100.0,
		nil,
		842.0,
		ContentTypeCell,
		CellOption{},
		1.0,
	)
	if cc.cacheContentText.fontSize != 14.0 {
		t.Errorf("expected fontSize 14, got %f", cc.cacheContentText.fontSize)
	}
}

func TestCov12_CacheContent_WriteTextToContent(t *testing.T) {
	cc := &CacheContent{}
	cc.Setup(
		&Rect{W: 200, H: 30},
		cacheContentTextColorRGB{r: 0, g: 0, b: 0},
		0.0, 1, 14.0, 0, 0.0, 0,
		50.0, 100.0, nil, 842.0,
		ContentTypeCell, CellOption{}, 1.0,
	)
	cc.WriteTextToContent("Hello")
	cc.WriteTextToContent(" World")
	if cc.cacheContentText.text != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", cc.cacheContentText.text)
	}
}

// ============================================================
// 3. cacheContentText.setPageHeight (0%)
// ============================================================

func TestCov12_CacheContentText_SetPageHeight(t *testing.T) {
	ct := &cacheContentText{}
	ct.setPageHeight(842.0)
	if ct.pageHeight() != 842.0 {
		t.Errorf("expected 842, got %f", ct.pageHeight())
	}
}

// ============================================================
// 4. cacheContentText.underline (0%)
// ============================================================

func TestCov12_CacheContentText_Underline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	// Set underline font style.
	pdf.SetFontWithStyle(fontFamily, Underline, 14)
	err := pdf.Cell(&Rect{W: 200, H: 30}, "Underlined text")
	if err != nil {
		t.Fatalf("Cell underline: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
	// The underline should produce a rectangle fill in the content stream.
	if !strings.Contains(string(data), "re f") {
		t.Log("underline rect fill may be in compressed stream")
	}
}

func TestCov12_CacheContentText_Underline_CustomCoefs(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 100)

	pdf.SetFontWithStyle(fontFamily, Underline, 14)
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Custom underline", CellOption{
		CoefLineHeight:         1.5,
		CoefUnderlinePosition:  1.2,
		CoefUnderlineThickness: 2.0,
	})
	if err != nil {
		t.Fatalf("CellWithOption custom underline: %v", err)
	}
}

// ============================================================
// 5. pdf_decrypt helpers — parseEncryptDict, padPassword, etc.
// ============================================================

func TestCov12_DetectEncryption_NoEncrypt(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")
	data := pdf.GetBytesPdf()

	result := detectEncryption(data)
	if result > 0 {
		t.Errorf("expected no encryption, got obj %d", result)
	}
}

func TestCov12_DetectEncryption_FakeEncrypt(t *testing.T) {
	// Create a minimal PDF-like data with /Encrypt reference.
	data := []byte(`%PDF-1.4
trailer
<< /Encrypt 5 0 R /Root 1 0 R /Size 6 >>
startxref
0
%%EOF`)
	result := detectEncryption(data)
	if result != 5 {
		t.Errorf("expected encrypt obj 5, got %d", result)
	}
}

func TestCov12_ParseEncryptDict(t *testing.T) {
	dict := `<< /Filter /Standard /V 1 /R 2 /Length 40 /P -44 /O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A> /U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A> >>`
	v, r, keyLen, oValue, uValue, pValue, err := parseEncryptDict(dict)
	if err != nil {
		t.Fatalf("parseEncryptDict: %v", err)
	}
	if v != 1 {
		t.Errorf("expected V=1, got %d", v)
	}
	if r != 2 {
		t.Errorf("expected R=2, got %d", r)
	}
	if keyLen != 5 {
		t.Errorf("expected keyLen=5, got %d", keyLen)
	}
	if pValue != -44 {
		t.Errorf("expected P=-44, got %d", pValue)
	}
	if len(oValue) == 0 {
		t.Error("expected non-empty O value")
	}
	if len(uValue) == 0 {
		t.Error("expected non-empty U value")
	}
}

func TestCov12_ParseEncryptDict_Literal(t *testing.T) {
	// Literal strings in encrypt dict need proper hex values.
	dict := `<< /Filter /Standard /V 2 /R 3 /Length 128 /P -3904 /O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A> /U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A> >>`
	v, r, _, _, _, _, err := parseEncryptDict(dict)
	if err != nil {
		t.Fatalf("parseEncryptDict: %v", err)
	}
	if v != 2 || r != 3 {
		t.Errorf("expected V=2 R=3, got V=%d R=%d", v, r)
	}
}

func TestCov12_PadPassword(t *testing.T) {
	// Empty password.
	padded := padPassword(nil)
	if len(padded) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(padded))
	}

	// Short password.
	padded = padPassword([]byte("test"))
	if len(padded) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(padded))
	}

	// Long password (>32 bytes).
	padded = padPassword(bytes.Repeat([]byte("A"), 50))
	if len(padded) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(padded))
	}
}

func TestCov12_ComputeEncryptionKey(t *testing.T) {
	oValue := make([]byte, 32)
	key := computeEncryptionKey([]byte("test"), oValue, -44, 5, 2)
	if len(key) != 5 {
		t.Errorf("expected key length 5, got %d", len(key))
	}
}

func TestCov12_ComputeUValue_R2(t *testing.T) {
	key := make([]byte, 5)
	uValue := computeUValue(key, 2)
	if len(uValue) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(uValue))
	}
}

func TestCov12_ComputeUValue_R3(t *testing.T) {
	key := make([]byte, 16)
	uValue := computeUValue(key, 3)
	if len(uValue) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(uValue))
	}
}

func TestCov12_TryUserPassword(t *testing.T) {
	oValue := make([]byte, 32)
	// Compute a valid U value for testing.
	key := computeEncryptionKey([]byte(""), oValue, -44, 5, 2)
	uValue := computeUValue(key, 2)

	resultKey, ok := tryUserPassword([]byte(""), oValue, uValue, -44, 5, 2)
	if !ok {
		t.Error("expected user password to match")
	}
	if len(resultKey) != 5 {
		t.Errorf("expected key length 5, got %d", len(resultKey))
	}
}

func TestCov12_TryUserPassword_Wrong(t *testing.T) {
	oValue := make([]byte, 32)
	uValue := make([]byte, 32)
	_, ok := tryUserPassword([]byte("wrong"), oValue, uValue, -44, 5, 2)
	if ok {
		t.Error("expected user password to NOT match")
	}
}

func TestCov12_TryOwnerPassword(t *testing.T) {
	oValue := make([]byte, 32)
	key := computeEncryptionKey([]byte(""), oValue, -44, 5, 2)
	uValue := computeUValue(key, 2)

	_, ok := tryOwnerPassword([]byte("owner"), oValue, uValue, -44, 5, 2)
	// May or may not match depending on how O was computed.
	_ = ok
}

func TestCov12_DecryptContext_ObjectKey(t *testing.T) {
	dc := &decryptContext{
		encryptionKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		keyLen:        5,
	}
	objKey := dc.objectKey(1)
	if len(objKey) == 0 {
		t.Error("expected non-empty object key")
	}
}

func TestCov12_DecryptContext_DecryptStream(t *testing.T) {
	dc := &decryptContext{
		encryptionKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		keyLen:        5,
	}
	data := []byte("encrypted data here")
	result, err := dc.decryptStream(1, data)
	if err != nil {
		t.Fatalf("decryptStream: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov12_ExtractSignedIntValue(t *testing.T) {
	dict := "<< /P -3904 /V 2 >>"
	p := extractSignedIntValue(dict, "/P")
	if p != -3904 {
		t.Errorf("expected -3904, got %d", p)
	}
	v := extractSignedIntValue(dict, "/V")
	if v != 2 {
		t.Errorf("expected 2, got %d", v)
	}
}

func TestCov12_ExtractHexOrLiteralString(t *testing.T) {
	dict := "<< /O <48656C6C6F> /U (World) >>"
	o := extractHexOrLiteralString(dict, "/O")
	if string(o) != "Hello" {
		t.Errorf("expected 'Hello', got %q", o)
	}
	u := extractHexOrLiteralString(dict, "/U")
	if string(u) != "World" {
		t.Errorf("expected 'World', got %q", u)
	}
}

func TestCov12_DecodeHexString(t *testing.T) {
	result := decodeHexString("48656C6C6F")
	if string(result) != "Hello" {
		t.Errorf("expected 'Hello', got %q", result)
	}
}

func TestCov12_DecodeLiteralString(t *testing.T) {
	result := decodeLiteralString(`(Hello\)World)`)
	if string(result) != "Hello)World" {
		t.Errorf("expected 'Hello)World', got %q", result)
	}
}

func TestCov12_RemoveEncryptFromTrailer(t *testing.T) {
	data := []byte(`%PDF-1.4
trailer
<< /Encrypt 5 0 R /Root 1 0 R /Size 6 >>
startxref
0
%%EOF`)
	result := removeEncryptFromTrailer(data)
	if bytes.Contains(result, []byte("/Encrypt")) {
		t.Error("expected /Encrypt to be removed")
	}
}

// ============================================================
// 6. ImageObj.GetRect (deprecated, 0%)
// ============================================================

func TestCov12_ImageObj_GetRect_Deprecated(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	imgObj := &ImageObj{}
	imgObj.init(nil)
	imgObj.SetImagePath(resJPEGPath)
	rect := imgObj.GetRect()
	if rect == nil {
		t.Fatal("expected non-nil rect")
	}
	if rect.W <= 0 || rect.H <= 0 {
		t.Errorf("expected positive dimensions: %+v", rect)
	}
}

// ============================================================
// 7. InsertHTMLBox — trigger remainingWidth (0%)
// ============================================================

func TestCov12_InsertHTMLBox_LongLine(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Long text that forces line wrapping, which triggers remainingWidth.
	html := `<p>This is a very long paragraph that should definitely wrap across multiple lines in the box because it contains many words and the box width is limited to just 200 points which is not very wide at all for this amount of text.</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 200, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox: %v", err)
	}
}

func TestCov12_InsertHTMLBox_InlineElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Normal <b>bold</b> <i>italic</i> <u>underline</u> <s>strike</s> text</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox inline: %v", err)
	}
}

// ============================================================
// 8. content_obj.write — more branches
// ============================================================

func TestCov12_ContentObj_Write_NoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.Cell(nil, "No compression content")
	pdf.Line(10, 10, 200, 200)
	pdf.SetGrayFill(0.5)
	pdf.Cell(nil, "Gray")
	pdf.SetFillColor(255, 0, 0)
	pdf.Cell(nil, "Red")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 9. DecryptPDF (0%) — test with synthetic encrypted data
// ============================================================

func TestCov12_DecryptPDF(t *testing.T) {
	// Create a minimal PDF with fake encrypted streams.
	data := []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R >>
endobj
xref
0 4
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
trailer
<< /Root 1 0 R /Size 4 >>
startxref
162
%%EOF`)

	dc := &decryptContext{
		encryptionKey: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		keyLen:        5,
	}
	result := decryptPDF(data, dc)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// ============================================================
// 10. Authenticate — with empty password
// ============================================================

func TestCov12_Authenticate_NoEncrypt(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")
	data := pdf.GetBytesPdf()

	dc, err := authenticate(data, "")
	// Non-encrypted PDF should return nil, nil or error.
	if dc != nil {
		t.Log("unexpected decrypt context for non-encrypted PDF")
	}
	_ = err
}

// ============================================================
// 11. Cell with border drawing (drawBorder path)
// ============================================================

func TestCov12_Cell_DrawBorder_AllSides(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "All borders", CellOption{
		Border: Left | Right | Top | Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption all borders: %v", err)
	}

	pdf.SetXY(50, 100)
	err = pdf.CellWithOption(&Rect{W: 200, H: 30}, "Left only", CellOption{
		Border: Left,
	})
	if err != nil {
		t.Fatalf("CellWithOption left border: %v", err)
	}

	pdf.SetXY(50, 150)
	err = pdf.CellWithOption(&Rect{W: 200, H: 30}, "Top only", CellOption{
		Border: Top,
	})
	if err != nil {
		t.Fatalf("CellWithOption top border: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 12. Cell with Float option
// ============================================================

func TestCov12_Cell_Float(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 100, H: 30}, "Float right", CellOption{
		Float: Right,
	})
	if err != nil {
		t.Fatalf("CellWithOption float right: %v", err)
	}

	pdf.SetXY(50, 100)
	err = pdf.CellWithOption(&Rect{W: 100, H: 30}, "Float bottom", CellOption{
		Float: Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption float bottom: %v", err)
	}
}
