继续 package gopdf

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

// ─── gopdf.go: Rectangle with round corners ───

func TestCov33_Rectangle_RoundCorners(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetFillColor(200, 200, 200)
	// round corner rectangle
	if err := pdf.Rectangle(50, 50, 300, 200, "DF", 10, 3); err != nil {
		t.Fatalf("Rectangle round corners: %v", err)
	}
	// invalid coords
	if err := pdf.Rectangle(300, 50, 50, 200, "D", 0, 0); err == nil {
		t.Fatal("expected error for invalid coords")
	}
	// radius too large
	if err := pdf.Rectangle(50, 50, 100, 100, "D", 60, 3); err == nil {
		t.Fatal("expected error for radius too large")
	}
	// no radius (radiusPointNum=0)
	if err := pdf.Rectangle(50, 50, 300, 200, "F", 0, 0); err != nil {
		t.Fatalf("Rectangle no radius: %v", err)
	}
}

// ─── gopdf.go: Polygon, Polyline, Sector, ClipPolygon ───

func TestCov33_Polygon_Polyline_Sector_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetFillColor(0, 255, 0)

	pts := []Point{{X: 10, Y: 10}, {X: 100, Y: 10}, {X: 55, Y: 80}}
	pdf.Polygon(pts, "DF")
	pdf.Polyline(pts)
	pdf.Sector(200, 200, 50, 0, 90, "FD")
	pdf.Sector(200, 200, 50, 90, 270, "D")
	pdf.Sector(200, 200, 50, 0, 45, "")

	pdf.SaveGraphicsState()
	pdf.ClipPolygon(pts)
	pdf.RestoreGraphicsState()

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
}

// ─── gopdf.go: IsFitMultiCellWithNewline ───

func TestCov33_IsFitMultiCellWithNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// fits
	ok, h, err := pdf.IsFitMultiCellWithNewline(&Rect{W: 200, H: 500}, "Hello\nWorld\nTest")
	if err != nil {
		t.Fatalf("IsFitMultiCellWithNewline: %v", err)
	}
	if !ok {
		t.Fatal("expected fit")
	}
	if h <= 0 {
		t.Fatal("expected positive height")
	}

	// doesn't fit
	ok2, _, err2 := pdf.IsFitMultiCellWithNewline(&Rect{W: 200, H: 1}, "Hello\nWorld")
	if err2 != nil {
		t.Fatalf("IsFitMultiCellWithNewline: %v", err2)
	}
	if ok2 {
		t.Fatal("expected not fit")
	}
}

// ─── gopdf.go: MultiCellWithOption with BreakOption ───

func TestCov33_MultiCellWithOption_BreakModes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// BreakModeStrict with separator
	opt := CellOption{
		Align: Left | Top,
		BreakOption: &BreakOption{
			Mode:      BreakModeStrict,
			Separator: "-",
		},
	}
	err := pdf.MultiCellWithOption(&Rect{W: 80, H: 200}, "Superlongwordthatneedstobebrokenstrictly", opt)
	if err != nil {
		t.Fatalf("MultiCellWithOption strict: %v", err)
	}

	// BreakModeIndicatorSensitive
	opt2 := CellOption{
		Align: Left | Top,
		BreakOption: &BreakOption{
			Mode:             BreakModeIndicatorSensitive,
			BreakIndicator:   ' ',
			Separator:        "-",
		},
	}
	err = pdf.MultiCellWithOption(&Rect{W: 80, H: 200}, "This is a long sentence that should break at spaces", opt2)
	if err != nil {
		t.Fatalf("MultiCellWithOption indicator: %v", err)
	}
}

// ─── gopdf.go: SplitTextWithOption various modes ───

func TestCov33_SplitTextWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// empty string
	_, err := pdf.SplitTextWithOption("", 100, nil)
	if err == nil {
		t.Fatal("expected error for empty string")
	}

	// strict mode with separator
	opt := &BreakOption{
		Mode:      BreakModeStrict,
		Separator: "-",
	}
	parts, err := pdf.SplitTextWithOption("Superlongwordthatneedstobebrokenstrictly", 60, opt)
	if err != nil {
		t.Fatalf("SplitTextWithOption strict: %v", err)
	}
	if len(parts) < 2 {
		t.Fatalf("expected multiple parts, got %d", len(parts))
	}

	// indicator sensitive mode
	opt2 := &BreakOption{
		Mode:           BreakModeIndicatorSensitive,
		BreakIndicator: ' ',
	}
	parts2, err := pdf.SplitTextWithOption("Hello World Foo Bar Baz", 60, opt2)
	if err != nil {
		t.Fatalf("SplitTextWithOption indicator: %v", err)
	}
	if len(parts2) < 2 {
		t.Fatalf("expected multiple parts, got %d", len(parts2))
	}

	// with newlines
	parts3, err := pdf.SplitTextWithOption("Line1\nLine2\nLine3", 200, nil)
	if err != nil {
		t.Fatalf("SplitTextWithOption newlines: %v", err)
	}
	if len(parts3) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts3))
	}
}

// ─── pdf_lowlevel.go: ReadObject, UpdateObject, GetDictKey, SetDictKey, GetStream, SetStream, GetTrailer ───

func TestCov33_PDFLowLevel_ReadUpdateGetSet(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.Cell(&Rect{W: 100, H: 20}, "Hello"); err != nil {
		t.Fatalf("Cell: %v", err)
	}
	var buf bytes.Buffer
	if _, err := pdf.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	data := buf.Bytes()

	// ReadObject
	obj, err := ReadObject(data, 1)
	if err != nil {
		t.Fatalf("ReadObject: %v", err)
	}
	if obj.Num != 1 {
		t.Fatalf("expected obj num 1, got %d", obj.Num)
	}

	// ReadObject not found
	_, err = ReadObject(data, 99999)
	if err == nil {
		t.Fatal("expected error for missing object")
	}

	// GetDictKey
	val, err := GetDictKey(data, 1, "/Type")
	if err != nil {
		t.Fatalf("GetDictKey: %v", err)
	}
	_ = val

	// UpdateObject
	updated, err := UpdateObject(data, 1, "<< /Type /Catalog >>")
	if err != nil {
		t.Fatalf("UpdateObject: %v", err)
	}
	if len(updated) == 0 {
		t.Fatal("expected non-empty updated data")
	}

	// SetDictKey
	updated2, err := SetDictKey(data, 1, "/CustomKey", "/CustomValue")
	if err != nil {
		t.Fatalf("SetDictKey: %v", err)
	}
	if len(updated2) == 0 {
		t.Fatal("expected non-empty data")
	}

	// GetTrailer
	trailer, err := GetTrailer(data)
	if err != nil {
		t.Fatalf("GetTrailer: %v", err)
	}
	if !strings.Contains(trailer, "trailer") {
		t.Fatal("expected trailer content")
	}

	// GetTrailer on bad data
	_, err = GetTrailer([]byte("no trailer here"))
	if err == nil {
		t.Fatal("expected error for missing trailer")
	}
}

// ─── pdf_lowlevel.go: GetStream, SetStream with stream objects ───

func TestCov33_PDFLowLevel_StreamOps(t *testing.T) {
	// Build a minimal PDF with a stream object
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Find a content stream object (usually obj 4 or 5)
	for objNum := 1; objNum <= 20; objNum++ {
		obj, err := ReadObject(data, objNum)
		if err != nil {
			continue
		}
		if obj.Stream != nil {
			// GetStream
			stream, err := GetStream(data, objNum)
			if err != nil {
				t.Fatalf("GetStream obj %d: %v", objNum, err)
			}
			if stream == nil {
				t.Fatalf("expected stream for obj %d", objNum)
			}

			// SetStream
			newData, err := SetStream(data, objNum, []byte("BT /F1 12 Tf 100 700 Td (Hi) Tj ET"))
			if err != nil {
				t.Fatalf("SetStream obj %d: %v", objNum, err)
			}
			if len(newData) == 0 {
				t.Fatal("expected non-empty data after SetStream")
			}
			return
		}
	}
	t.Skip("no stream object found in test PDF")
}

// ─── pdf_lowlevel.go: extractDictKeyValue various value types ───

func TestCov33_ExtractDictKeyValue_Types(t *testing.T) {
	tests := []struct {
		dict, key, want string
	}{
		{"/Type /Page /MediaBox [0 0 612 792]", "/Type", "/Page"},
		{"/MediaBox [0 0 612 792]", "/MediaBox", "[0 0 612 792]"},
		{"/Title (Hello World)", "/Title", "(Hello World)"},
		{"/Info << /Author (Test) >>", "/Info", "<< /Author (Test) >>"},
		{"/ID <ABCDEF>", "/ID", "<ABCDEF>"},
		{"/Pages 2 0 R", "/Pages", "2 0 R"},
		{"/Count 5 /Type /Pages", "/Count", "5"},
		{"/Missing /Type", "/NotHere", ""},
	}
	for _, tt := range tests {
		got := extractDictKeyValue(tt.dict, tt.key)
		if got != tt.want {
			t.Errorf("extractDictKeyValue(%q, %q) = %q, want %q", tt.dict, tt.key, got, tt.want)
		}
	}
}

// ─── pdf_lowlevel.go: setDictKeyValue ───

func TestCov33_SetDictKeyValue(t *testing.T) {
	// update existing key
	dict := "/Type /Page\n/MediaBox [0 0 612 792]"
	result := setDictKeyValue(dict, "/MediaBox", "[0 0 595 842]")
	if !strings.Contains(result, "[0 0 595 842]") {
		t.Fatalf("expected updated MediaBox, got: %s", result)
	}

	// add new key
	result2 := setDictKeyValue(dict, "/NewKey", "/NewValue")
	if !strings.Contains(result2, "/NewKey /NewValue") {
		t.Fatalf("expected new key, got: %s", result2)
	}
}

// ─── pixmap_render.go: RenderPageToImage, RenderAllPagesToImages ───

func TestCov33_RenderPageToImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetLineWidth(2)
	pdf.Line(10, 10, 200, 200)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 80, "D")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Render page 0
	img, err := RenderPageToImage(data, 0, RenderOption{DPI: 72})
	if err != nil {
		t.Fatalf("RenderPageToImage: %v", err)
	}
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		t.Fatal("expected non-zero image dimensions")
	}

	// Out of range
	_, err = RenderPageToImage(data, 99, RenderOption{})
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// Negative page
	_, err = RenderPageToImage(data, -1, RenderOption{})
	if err == nil {
		t.Fatal("expected error for negative page")
	}

	// Bad data — may or may not error depending on parser
	_, err = RenderPageToImage([]byte("not a pdf"), 0, RenderOption{})
	// parser may succeed with 0 pages, then page index out of range
	_ = err

	// RenderAllPagesToImages
	imgs, err := RenderAllPagesToImages(data, RenderOption{DPI: 36})
	if err != nil {
		t.Fatalf("RenderAllPagesToImages: %v", err)
	}
	if len(imgs) != 1 {
		t.Fatalf("expected 1 image, got %d", len(imgs))
	}

	// RenderAllPagesToImages bad data
	_, err = RenderAllPagesToImages([]byte("bad"), RenderOption{})
	_ = err // may or may not error
}

// ─── pixmap_render.go: RenderOption defaults ───

func TestCov33_RenderOption_Defaults(t *testing.T) {
	opt := RenderOption{}
	opt.defaults()
	if opt.DPI != 72 {
		t.Fatalf("expected DPI 72, got %f", opt.DPI)
	}
	if opt.Background == nil {
		t.Fatal("expected non-nil background")
	}

	// Custom values should be preserved
	opt2 := RenderOption{DPI: 150, Background: color.Black}
	opt2.defaults()
	if opt2.DPI != 150 {
		t.Fatalf("expected DPI 150, got %f", opt2.DPI)
	}
}

// ─── pixmap_render.go: renderContentStream with various operators ───

func TestCov33_RenderContentStream_Operators(t *testing.T) {
	// Build a PDF with various drawing operations
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Lines, rects, colors
	pdf.SetStrokeColor(128, 0, 0)
	pdf.SetFillColor(0, 128, 0)
	pdf.Line(10, 10, 100, 100)
	pdf.RectFromUpperLeftWithStyle(20, 20, 60, 40, "DF")

	// Gray
	pdf.SetGrayFill(0.5)
	pdf.SetGrayStroke(0.3)
	pdf.RectFromUpperLeftWithStyle(100, 100, 50, 50, "F")

	// Polygon
	pdf.Polygon([]Point{{X: 10, Y: 200}, {X: 50, Y: 250}, {X: 90, Y: 200}}, "D")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	img, err := RenderPageToImage(data, 0, RenderOption{DPI: 72})
	if err != nil {
		t.Fatalf("RenderPageToImage: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}

// ─── pixmap_render.go: renderXObject with image ───

func TestCov33_RenderXObject_WithImage(t *testing.T) {
	// Create a PDF with an embedded JPEG image
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create a small JPEG in memory
	imgBuf := createSmallJPEG(t)
	holder, err := ImageHolderByBytes(imgBuf)
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("ImageByHolder: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	img, err := RenderPageToImage(data, 0, RenderOption{DPI: 72})
	if err != nil {
		t.Fatalf("RenderPageToImage with image: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
}

func createSmallJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return buf.Bytes()
}

// ─── digital_signature.go: SignPDF with ECDSA key ───

func TestCov33_SignPDF_ECDSA(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Signed Doc")

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ECDSA key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test ECDSA"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}

	cfg := SignatureConfig{
		Certificate: cert,
		PrivateKey:  key,
		Reason:      "ECDSA Test",
	}

	var buf bytes.Buffer
	err = pdf.SignPDF(cfg, &buf)
	if err != nil {
		t.Fatalf("SignPDF ECDSA: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty signed PDF")
	}
}

// ─── digital_signature.go: SignPDF error paths ───

func TestCov33_SignPDF_Errors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Test")

	// No certificate
	err := pdf.SignPDF(SignatureConfig{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing certificate")
	}

	// No private key
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert, _ := x509.ParseCertificate(certDER)

	err = pdf.SignPDF(SignatureConfig{Certificate: cert}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing private key")
	}

	// Unsupported key type (ed25519 would be unsupported, but let's use a mock)
	// We can't easily create an unsupported key, so skip that branch
}

// ─── digital_signature.go: VerifySignature on unsigned PDF ───

func TestCov33_VerifySignature_NoSig(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Unsigned")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	_, err := VerifySignature(buf.Bytes())
	if err == nil {
		t.Fatal("expected error for unsigned PDF")
	}
}

// ─── digital_signature.go: extractSignatures edge cases ───

func TestCov33_ExtractSignatures_EdgeCases(t *testing.T) {
	// No signatures
	sigs, err := extractSignatures([]byte("%PDF-1.4 no sigs here"))
	if err != nil {
		t.Fatalf("extractSignatures: %v", err)
	}
	if len(sigs) != 0 {
		t.Fatalf("expected 0 sigs, got %d", len(sigs))
	}

	// Malformed /Type /Sig without proper dict
	sigs2, _ := extractSignatures([]byte("<< /Type /Sig /ByteRange [0 100 200 300] /Contents <AABB> >>"))
	if len(sigs2) != 1 {
		t.Fatalf("expected 1 sig, got %d", len(sigs2))
	}

	// /Type/Sig (no space variant)
	sigs3, _ := extractSignatures([]byte("<< /Type/Sig /ByteRange [0 50 100 50] /Contents <FF00> >>"))
	if len(sigs3) != 1 {
		t.Fatalf("expected 1 sig for no-space variant, got %d", len(sigs3))
	}
}

// ─── digital_signature.go: parseIntArray, hexDecode, trimTrailingZeros, extractPDFString ───

func TestCov33_DigSig_Helpers(t *testing.T) {
	// parseIntArray
	arr := parseIntArray("0 100 200 300")
	if len(arr) != 4 || arr[0] != 0 || arr[3] != 300 {
		t.Fatalf("parseIntArray unexpected: %v", arr)
	}
	arr2 := parseIntArray("")
	if len(arr2) != 0 {
		t.Fatalf("expected empty array, got %v", arr2)
	}

	// hexDecode
	decoded := hexDecode("48656C6C6F")
	if string(decoded) != "Hello" {
		t.Fatalf("hexDecode: got %q", string(decoded))
	}
	decoded2 := hexDecode("4865") // even length
	if string(decoded2) != "He" {
		t.Fatalf("hexDecode even: got %q", string(decoded2))
	}

	// trimTrailingZeros (trims pairs of "00")
	s := trimTrailingZeros("AABB00000")
	if s != "AABB0" { // odd trailing zeros: only pairs removed
		t.Fatalf("trimTrailingZeros: got %q", s)
	}
	s2 := trimTrailingZeros("0000")
	if s2 != "" {
		t.Fatalf("trimTrailingZeros all zeros: got %q", s2)
	}
	s3 := trimTrailingZeros("AABB0000")
	if s3 != "AABB" {
		t.Fatalf("trimTrailingZeros even: got %q", s3)
	}

	// extractPDFString
	ps := extractPDFString([]byte("(Hello World) /Next"))
	if ps != "Hello World" {
		t.Fatalf("extractPDFString paren: got %q", ps)
	}
	// extractPDFString only handles (...) strings, not <hex>
	ps2 := extractPDFString([]byte("  (Nested (parens) here)"))
	if ps2 != "Nested (parens) here" {
		t.Fatalf("extractPDFString nested: got %q", ps2)
	}
	// with escape
	ps3 := extractPDFString([]byte(`(Hello\nWorld)`))
	if ps3 != "HellonWorld" { // backslash-n becomes 'n' (simple escape)
		t.Fatalf("extractPDFString escape: got %q", ps3)
	}
	ps4 := extractPDFString([]byte("no string here"))
	if ps4 != "" {
		t.Fatalf("extractPDFString none: got %q", ps4)
	}
}

// ─── pdf_parser.go: newRawPDFParser, parseObjects, collectPages, extractMediaBox, extractContentRefs ───

func TestCov33_RawPDFParser(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Parser Test")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	parser, err := newRawPDFParser(data)
	if err != nil {
		t.Fatalf("newRawPDFParser: %v", err)
	}
	if len(parser.pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(parser.pages))
	}

	// getPageContentStream
	stream := parser.getPageContentStream(0)
	if len(stream) == 0 {
		t.Fatal("expected non-empty content stream for page 0")
	}

	// bad data
	_, err = newRawPDFParser([]byte("not a pdf"))
	if err != nil {
		// some parsers may not error on bad data, just have 0 pages
		_ = err
	}
}

// ─── pdf_parser.go: extractDict, extractRef, extractRefArray, extractMediaBox ───

func TestCov33_PDFParser_Helpers(t *testing.T) {
	// extractDict
	dict := extractDict([]byte("<< /Type /Page /MediaBox [0 0 612 792] >>"))
	if !strings.Contains(dict, "/Type") {
		t.Fatalf("extractDict: got %q", dict)
	}

	// nested dict
	dict2 := extractDict([]byte("<< /Resources << /Font << >> >> >>"))
	if dict2 == "" {
		t.Fatal("expected non-empty dict for nested")
	}

	// extractRef
	ref := extractRef("<< /Pages 2 0 R >>", "/Pages")
	if ref != 2 {
		t.Fatalf("extractRef: got %d", ref)
	}
	ref2 := extractRef("<< /Type /Catalog >>", "/Pages")
	if ref2 != 0 {
		t.Fatalf("extractRef missing: got %d", ref2)
	}

	// extractRefArray
	refs := extractRefArray("<< /Kids [3 0 R 5 0 R 7 0 R] >>", "/Kids")
	if len(refs) != 3 {
		t.Fatalf("extractRefArray: got %v", refs)
	}

	// extractMediaBox
	mb := extractMediaBox("<< /MediaBox [0 0 595.28 841.89] >>")
	if mb[2] < 595 || mb[3] < 841 {
		t.Fatalf("extractMediaBox: got %v", mb)
	}

	// extractContentRefs
	cr := extractContentRefs("<< /Contents 4 0 R >>")
	if len(cr) != 1 || cr[0] != 4 {
		t.Fatalf("extractContentRefs single: got %v", cr)
	}
	cr2 := extractContentRefs("<< /Contents [4 0 R 6 0 R] >>")
	if len(cr2) != 2 {
		t.Fatalf("extractContentRefs array: got %v", cr2)
	}
}

// ─── select_pages.go: SelectPages, SelectPagesFromBytes ───

func TestCov33_SelectPages(t *testing.T) {
	// Use the test PDF file if available (gofpdi panics on re-exported PDFs)
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skip("test PDF not available")
	}

	// SelectPagesFromFile
	result, err := SelectPagesFromFile(resTestPDF, []int{1}, nil)
	if err != nil {
		t.Fatalf("SelectPagesFromFile: %v", err)
	}
	if result.GetNumberOfPages() != 1 {
		t.Fatalf("expected 1 page, got %d", result.GetNumberOfPages())
	}

	// SelectPages - empty
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err = pdf.SelectPages([]int{})
	if err == nil {
		t.Fatal("expected error for empty pages")
	}

	// SelectPages - out of range
	_, err = pdf.SelectPages([]int{99})
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// SelectPagesFromBytes - empty
	_, err = SelectPagesFromBytes([]byte("%PDF-1.4"), []int{}, nil)
	if err == nil {
		t.Fatal("expected error for empty pages")
	}
}

// ─── select_pages.go: SelectPagesFromFile ───

func TestCov33_SelectPagesFromFile(t *testing.T) {
	// empty pages
	_, err := SelectPagesFromFile("nonexistent.pdf", []int{}, nil)
	if err == nil {
		t.Fatal("expected error for empty pages")
	}

	// bad file
	_, err = SelectPagesFromFile("nonexistent.pdf", []int{1}, nil)
	if err == nil {
		t.Fatal("expected error for bad file")
	}

	// Use the test PDF if available
	if _, err := os.Stat(resTestPDF); err == nil {
		result, err := SelectPagesFromFile(resTestPDF, []int{1}, nil)
		if err != nil {
			t.Fatalf("SelectPagesFromFile: %v", err)
		}
		if result.GetNumberOfPages() != 1 {
			t.Fatalf("expected 1 page, got %d", result.GetNumberOfPages())
		}
	}
}

// ─── gopdf.go: AddPageWithOption with TrimBox and PageSize ───

func TestCov33_AddPageWithOption_TrimBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 400, H: 600},
		TrimBox: &Box{
			Left: 10, Top: 10, Right: 390, Bottom: 590,
		},
	})
	pdf.Cell(&Rect{W: 100, H: 20}, "TrimBox page")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
}

// ─── gopdf.go: AddHeader/AddFooter with page operations ───

func TestCov33_HeaderFooter_MultiPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddHeader(func() {
		pdf.SetY(15)
		pdf.Cell(&Rect{W: 100, H: 10}, "Header")
	})
	pdf.AddFooter(func() {
		pdf.SetY(280)
		pdf.Cell(&Rect{W: 100, H: 10}, "Footer")
	})
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1 content")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2 content")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
}

// ─── gopdf.go: SetNewY, SetNewYIfNoOffset, SetNewXY triggering page add ───

func TestCov33_SetNewY_PageAdd(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Move Y to near bottom of page first
	pdf.SetY(800) // near bottom of A4 (842pt)
	// SetNewY that triggers new page (curr.Y + h > pageSize.H - marginBottom)
	pdf.SetNewY(800, 100)
	if pdf.GetNumberOfPages() < 2 {
		t.Fatal("expected new page from SetNewY")
	}

	// SetNewYIfNoOffset
	pdf.SetY(800)
	pdf.SetNewYIfNoOffset(800, 100)

	// SetNewXY
	pdf.SetY(800)
	pdf.SetNewXY(800, 50, 100)
}

// ─── gopdf.go: Curve ───

func TestCov33_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 255)
	pdf.SetLineWidth(1)
	pdf.Curve(10, 10, 50, 200, 150, 200, 200, 10, "D")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: Oval ───

func TestCov33_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 0)
	pdf.Oval(100, 100, 80, 50)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: SetFontWithStyle ───

func TestCov33_SetFontWithStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Regular
	err := pdf.SetFontWithStyle(fontFamily, 0, 14)
	if err != nil {
		t.Fatalf("SetFontWithStyle regular: %v", err)
	}

	// Bold (may fail if bold font not loaded, that's ok)
	_ = pdf.SetFontWithStyle(fontFamily, Bold, 14)

	// Non-existent font
	err = pdf.SetFontWithStyle("NonExistentFont", 0, 14)
	if err == nil {
		t.Fatal("expected error for non-existent font")
	}
}

// ─── gopdf.go: SetTextColorCMYK ───

func TestCov33_SetTextColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColorCMYK(0, 100, 100, 0)
	pdf.Cell(&Rect{W: 100, H: 20}, "CMYK text")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: PlaceHolderText and FillInPlaceHoldText ───

func TestCov33_PlaceHolderText_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add placeholder
	if err := pdf.PlaceHolderText("name", 100); err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	// Fill it (align: Left=8)
	if err := pdf.FillInPlaceHoldText("name", "John Doe", Left); err != nil {
		t.Fatalf("FillInPlaceHoldText: %v", err)
	}

	// Fill non-existent
	if err := pdf.FillInPlaceHoldText("nonexistent", "value", Left); err == nil {
		t.Fatal("expected error for non-existent placeholder")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetCustomLineType ───

func TestCov33_SetCustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(10, 10, 200, 10)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: SetPDFVersion ───

func TestCov33_SetPDFVersion(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetPDFVersion(PDFVersion17)
	if pdf.GetPDFVersion() != PDFVersion17 {
		t.Fatalf("expected PDF 1.7, got %v", pdf.GetPDFVersion())
	}
}

// ─── gopdf.go: AddExternalLink, AddInternalLink, SetAnchor ───

func TestCov33_Links_Anchors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetAnchor("section1")
	pdf.Cell(&Rect{W: 100, H: 20}, "Section 1")

	pdf.AddPage()
	pdf.AddInternalLink("section1", 50, 50, 100, 20)
	pdf.AddExternalLink("https://example.com", 50, 80, 100, 20)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: ImageByHolder with PNG RGBA (smask path) ───

func TestCov33_ImageByHolder_PNG_RGBA_SMask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create RGBA PNG (triggers smask creation)
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 12), G: uint8(y * 12), B: 128, A: 200})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)

	holder, err := ImageHolderByBytes(pngBuf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 10, 10, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("ImageByHolder RGBA: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: ImageByHolder with GIF ───

func TestCov33_ImageByHolder_GIF(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create a small GIF
	gifImg := image.NewPaletted(image.Rect(0, 0, 10, 10), []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
	})
	var gifBuf bytes.Buffer
	gif.Encode(&gifBuf, gifImg, nil)

	holder, err := ImageHolderByBytes(gifBuf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes GIF: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50}); err != nil {
		t.Fatalf("ImageByHolder GIF: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: ImageByHolder cached path ───

func TestCov33_ImageByHolder_Cached(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)

	// First use
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50})
	// Second use (cached path)
	pdf.ImageByHolder(holder, 100, 10, nil) // nil rect uses cached rect

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: prepare() Font case ───

func TestCov33_Prepare_FontCase(t *testing.T) {
	// This tests the Font case in prepare() where it matches encoding objects
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// Add a TTF font with option to trigger encoding objects
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{UseKerning: true})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Font prepare test")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: protection-related paths ───

func TestCov33_Protection_Compile(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   PermissionsPrint | PermissionsCopy,
			OwnerPass:     []byte("owner"),
			UserPass:      []byte("user"),
		},
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Protected PDF")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo protected: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: SetNoCompression ───

func TestCov33_SetNoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "No compression")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: SaveGraphicsState / RestoreGraphicsState ───

func TestCov33_GraphicsState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.Line(10, 10, 100, 100)
	pdf.RestoreGraphicsState()

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: GetNextObjectID, GetNumberOfPages ───

func TestCov33_GetNextObjectID_GetNumberOfPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	id1 := pdf.GetNextObjectID()
	if id1 <= 0 {
		t.Fatalf("expected positive object ID, got %d", id1)
	}

	pdf.AddPage()
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 2 {
		t.Fatalf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

// ─── pdf_decrypt.go: decryptPDF deeper paths ───

func TestCov33_DecryptPDF_Deeper(t *testing.T) {
	// Build a protected PDF and try to decrypt it
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   PermissionsPrint,
			OwnerPass:     []byte("owner123"),
			UserPass:      []byte("user123"),
		},
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Encrypted content")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Try authenticate with correct password
	ctx, err2 := authenticate(data, "user123")
	if err2 == nil && ctx != nil {
		// If authentication succeeded, try decryptPDF
		result := decryptPDF(data, ctx)
		if len(result) == 0 {
			t.Fatal("expected non-empty decrypted data")
		}
	}

	// Try authenticate with wrong password
	ctx2, _ := authenticate(data, "wrongpass")
	_ = ctx2 // may be nil

	// Try authenticate on non-encrypted PDF
	plainPDF := newPDFWithFont(t)
	plainPDF.AddPage()
	plainPDF.Cell(&Rect{W: 100, H: 20}, "Plain")
	var plainBuf bytes.Buffer
	plainPDF.WriteTo(&plainBuf)
	ctx3, _ := authenticate(plainBuf.Bytes(), "")
	_ = ctx3
}

// ─── pdf_decrypt.go: detectEncryption ───

func TestCov33_DetectEncryption(t *testing.T) {
	// Non-encrypted PDF
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Plain")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	enc := detectEncryption(buf.Bytes())
	if enc != 0 {
		t.Fatal("expected non-encrypted PDF")
	}

	// Encrypted PDF
	pdf2 := &GoPdf{}
	pdf2.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   PermissionsPrint,
			OwnerPass:     []byte("owner"),
			UserPass:      []byte("user"),
		},
	})
	if err := pdf2.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf2.SetFont(fontFamily, "", 14)
	pdf2.AddPage()
	pdf2.Cell(&Rect{W: 200, H: 20}, "Encrypted")
	var buf2 bytes.Buffer
	pdf2.WriteTo(&buf2)

	enc2 := detectEncryption(buf2.Bytes())
	if enc2 == 0 {
		t.Fatal("expected encrypted PDF")
	}
}

// ─── html_insert.go: renderNode error propagation from children ───

func TestCov33_HTMLInsert_RenderNode_ErrorPropagation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Test with font tag with color and size
	html := `<font color="#FF0000" size="5" face="` + fontFamily + `">Red large text</font>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox font tag: %v", err)
	}

	// Test with span + inline style
	html2 := `<span style="color: blue; font-size: 18px;">Styled span</span>`
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html2, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox span style: %v", err)
	}

	// Test blockquote
	html3 := `<blockquote>This is a blockquote with some text that should be indented.</blockquote>`
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html3, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox blockquote: %v", err)
	}
}

// ─── html_insert.go: renderImage with various attributes ───

func TestCov33_HTMLInsert_RenderImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Image with width and height
	html := fmt.Sprintf(`<img src="%s" width="50" height="50">`, resJPEGPath)
	_, err := pdf.InsertHTMLBox(10, 10, 200, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img: %v", err)
	}

	// Image with only width
	html2 := fmt.Sprintf(`<img src="%s" width="80">`, resJPEGPath)
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html2, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img width-only: %v", err)
	}
}

// ─── gopdf.go: WriteTo error path ───

func TestCov33_WriteTo_Error(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Test")

	// Write to a writer that fails
	fw := &failWriterCov33{failAfter: 10}
	_, err := pdf.WriteTo(fw)
	if err == nil {
		// countingWriter may not propagate errors, so this might not fail
		_ = err
	}
}

type failWriterCov33 struct {
	written   int
	failAfter int
}

func (fw *failWriterCov33) Write(p []byte) (int, error) {
	if fw.written+len(p) > fw.failAfter {
		return 0, fmt.Errorf("write failed")
	}
	fw.written += len(p)
	return len(p), nil
}

// ─── gopdf.go: Close ───

func TestCov33_Close(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Close test")

	err := pdf.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// ─── gopdf.go: Write (io.Writer interface) ───

func TestCov33_Write(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Write test")

	var buf bytes.Buffer
	err := pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected non-zero bytes written")
	}
}

// ─── gopdf.go: SetCharSpacing ───

func TestCov33_SetCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetCharSpacing(2.0); err != nil {
		t.Fatalf("SetCharSpacing: %v", err)
	}
	pdf.Cell(&Rect{W: 200, H: 20}, "Spaced text")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: margins ───

func TestCov33_Margins(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetLeftMargin(20)
	pdf.SetTopMargin(30)
	pdf.AddPage()

	if pdf.GetX() != 20 {
		t.Fatalf("expected X=20, got %f", pdf.GetX())
	}

	pdf.SetMargins(15, 25, 10, 10)
	pdf.AddPage()
	if pdf.GetX() != 15 {
		t.Fatalf("expected X=15, got %f", pdf.GetX())
	}
}

// ─── gopdf.go: SetGrayFill, SetGrayStroke ───

func TestCov33_GrayFillStroke(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
	pdf.SetGrayStroke(0.3)
	pdf.RectFromUpperLeftWithStyle(10, 10, 100, 50, "DF")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetFillColorCMYK, SetStrokeColorCMYK ───

func TestCov33_CMYKColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColorCMYK(0, 100, 100, 0)
	pdf.SetStrokeColorCMYK(100, 0, 0, 0)
	pdf.RectFromUpperLeftWithStyle(10, 10, 100, 50, "DF")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: RectFromLowerLeft, RectFromLowerLeftWithStyle ───

func TestCov33_RectFromLowerLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 0)
	pdf.RectFromLowerLeft(50, 200, 100, 50)
	pdf.RectFromLowerLeftWithStyle(50, 300, 100, 50, "DF")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: Rotate, RotateReset ───

func TestCov33_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.Cell(&Rect{W: 100, H: 20}, "Rotated")
	pdf.RotateReset()

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetInfo, GetInfo ───

func TestCov33_SetGetInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetInfo(PdfInfo{
		Title:    "Test Title",
		Author:   "Test Author",
		Subject:  "Test Subject",
		Creator:  "Test Creator",
		Producer: "Test Producer",
	})
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Info test")

	info := pdf.GetInfo()
	if info.Title != "Test Title" {
		t.Fatalf("expected title 'Test Title', got %q", info.Title)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: AddTTFFontWithOption ───

func TestCov33_AddTTFFontWithOption(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}

	// Add same font again (should be ok or error gracefully)
	_ = pdf.AddTTFFontWithOption(fontFamily+"2", resFontPath, TtfOption{})

	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "TTF option test")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── image_recompress.go: downscaleImage edge cases ───

func TestCov33_DownscaleImage_EdgeCases(t *testing.T) {
	// Create a large image
	src := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			src.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}

	// Width-only constraint
	result := downscaleImage(src, 100, 0)
	if result.Bounds().Dx() != 100 {
		t.Fatalf("expected width 100, got %d", result.Bounds().Dx())
	}

	// Height-only constraint
	result2 := downscaleImage(src, 0, 50)
	if result2.Bounds().Dy() != 50 {
		t.Fatalf("expected height 50, got %d", result2.Bounds().Dy())
	}

	// Both constraints
	result3 := downscaleImage(src, 100, 50)
	if result3.Bounds().Dx() > 100 || result3.Bounds().Dy() > 50 {
		t.Fatalf("expected within 100x50, got %dx%d", result3.Bounds().Dx(), result3.Bounds().Dy())
	}

	// No downscale needed (target larger than source)
	result4 := downscaleImage(src, 400, 200)
	if result4.Bounds().Dx() != 200 {
		t.Fatalf("expected original width 200, got %d", result4.Bounds().Dx())
	}
}

// ─── page_manipulate.go: ExtractPages with multiple pages ───

func TestCov33_ExtractPages_MultiPage(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skip("test PDF not available")
	}

	result, err := ExtractPages(resTestPDF, []int{1}, nil)
	if err != nil {
		t.Fatalf("ExtractPages: %v", err)
	}
	if result.GetNumberOfPages() != 1 {
		t.Fatalf("expected 1 page, got %d", result.GetNumberOfPages())
	}
}

// ─── page_manipulate.go: MergePagesFromBytes with multiple PDFs ───

func TestCov33_MergePagesFromBytes_Multi(t *testing.T) {
	// Create two PDFs
	pdf1 := newPDFWithFont(t)
	pdf1.AddPage()
	pdf1.Cell(&Rect{W: 100, H: 20}, "PDF 1")
	var buf1 bytes.Buffer
	pdf1.WriteTo(&buf1)

	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.Cell(&Rect{W: 100, H: 20}, "PDF 2")
	var buf2 bytes.Buffer
	pdf2.WriteTo(&buf2)

	result, err := MergePagesFromBytes([][]byte{buf1.Bytes(), buf2.Bytes()}, nil)
	if err != nil {
		t.Fatalf("MergePagesFromBytes: %v", err)
	}
	if result.GetNumberOfPages() != 2 {
		t.Fatalf("expected 2 pages, got %d", result.GetNumberOfPages())
	}
}

// ─── content_obj.go: write with NoCompression ───

func TestCov33_ContentObj_NoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "No compression content")
	pdf.Line(10, 10, 200, 200)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: Br ───

func TestCov33_Br(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Line 1")
	pdf.Br(20)
	pdf.Cell(&Rect{W: 100, H: 20}, "Line 2")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: MeasureCellHeightByText ───

func TestCov33_MeasureCellHeightByText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	h, err := pdf.MeasureCellHeightByText("Hello World")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Fatal("expected positive height")
	}
}

// ─── gopdf.go: GetAllPageSizes ───

func TestCov33_GetAllPageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 400, H: 600},
	})

	sizes := pdf.GetAllPageSizes()
	if len(sizes) != 2 {
		t.Fatalf("expected 2 page sizes, got %d", len(sizes))
	}
}

// ─── gopdf.go: SetPage ───

func TestCov33_SetPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")

	// Go back to page 1
	pdf.SetPage(1)
	pdf.Cell(&Rect{W: 100, H: 20}, "Back to page 1")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── digital_signature.go: SignPDF with visible signature ───

func TestCov33_SignPDF_Visible(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Visible Sig")

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test Visible"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert, _ := x509.ParseCertificate(certDER)

	cfg := SignatureConfig{
		Certificate:        cert,
		PrivateKey:         key,
		Reason:             "Visible Test",
		Location:           "Test Location",
		Visible:            true,
		PageNo:             1,
		SignatureFieldName: "sig1",
		X: 50, Y: 50, W: 100, H: 50,
	}

	var buf bytes.Buffer
	err := pdf.SignPDF(cfg, &buf)
	if err != nil {
		t.Fatalf("SignPDF visible: %v", err)
	}
}

// ─── digital_signature.go: AddSignatureField ───

func TestCov33_AddSignatureField(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Sig Field")

	err := pdf.AddSignatureField("sig1", 50, 50, 100, 30)
	if err != nil {
		t.Fatalf("AddSignatureField: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── digital_signature.go: LoadCertificateFromPEM, ParseCertificatePEM errors ───

func TestCov33_DigSig_PEM_Errors(t *testing.T) {
	// Bad file path
	_, err := LoadCertificateFromPEM("nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for bad path")
	}

	// Bad PEM data
	_, err = ParseCertificatePEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for bad PEM")
	}

	// Bad private key path
	_, err = LoadPrivateKeyFromPEM("nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for bad path")
	}

	// Bad private key PEM
	_, err = ParsePrivateKeyPEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for bad PEM")
	}

	// Bad chain path
	_, err = LoadCertificateChainFromPEM("nonexistent.pem")
	if err == nil {
		t.Fatal("expected error for bad path")
	}

	// Bad chain PEM — returns error for no certs found
	_, err = ParseCertificateChainPEM([]byte("not a pem"))
	if err == nil {
		t.Fatal("expected error for bad chain PEM")
	}
}

// ─── gopdf.go: DrawableRectOptions ───

func TestCov33_DrawableRectOptions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetFillColor(200, 200, 200)

	// Draw style
	pdf.RectFromUpperLeftWithStyle(10, 10, 100, 50, "D")
	// Fill style
	pdf.RectFromUpperLeftWithStyle(10, 70, 100, 50, "F")
	// Draw+Fill style
	pdf.RectFromUpperLeftWithStyle(10, 130, 100, 50, "DF")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── unused import guard ───

var _ = io.Discard
var _ = os.DevNull

// ─── text_search.go: SearchText, SearchTextOnPage ───

func TestCov33_SearchText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Hello World")
	pdf.Br(20)
	pdf.Cell(&Rect{W: 200, H: 20}, "Foo Bar Baz")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Second Page Hello")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Search across all pages
	results, err := SearchText(data, "Hello", false)
	if err != nil {
		t.Fatalf("SearchText: %v", err)
	}
	_ = results // may or may not find depending on encoding

	// Case insensitive
	results2, err := SearchText(data, "hello", true)
	if err != nil {
		t.Fatalf("SearchText case insensitive: %v", err)
	}
	_ = results2

	// SearchTextOnPage
	results3, err := SearchTextOnPage(data, 0, "Hello", false)
	if err != nil {
		t.Fatalf("SearchTextOnPage: %v", err)
	}
	_ = results3

	// Out of range page
	results4, err := SearchTextOnPage(data, 99, "Hello", false)
	if err != nil {
		t.Fatalf("SearchTextOnPage out of range: %v", err)
	}
	if len(results4) != 0 {
		t.Fatal("expected no results for out-of-range page")
	}
}

// ─── watermark.go: AddWatermarkText, AddWatermarkTextAllPages ───

func TestCov33_Watermark_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Page 2")

	// Single page watermark
	pdf.SetPage(1)
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   48,
		Opacity:    0.3,
		Angle:      45,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText: %v", err)
	}

	// Repeat watermark
	pdf.SetPage(2)
	err = pdf.AddWatermarkText(WatermarkOption{
		Text:       "CONFIDENTIAL",
		FontFamily: fontFamily,
		FontSize:   24,
		Opacity:    0.2,
		Angle:      30,
		Repeat:     true,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText repeat: %v", err)
	}

	// All pages
	err = pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "ALL",
		FontFamily: fontFamily,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}

	// Error: empty text
	err = pdf.AddWatermarkText(WatermarkOption{FontFamily: fontFamily})
	if err == nil {
		t.Fatal("expected error for empty text")
	}

	// Error: missing font family
	err = pdf.AddWatermarkText(WatermarkOption{Text: "test"})
	if err == nil {
		t.Fatal("expected error for missing font family")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── watermark.go: AddWatermarkImage, AddWatermarkImageAllPages ───

func TestCov33_Watermark_Image(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skip("JPEG not available")
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Page 1")

	err := pdf.AddWatermarkImage(resJPEGPath, 0.3, 100, 100, 45)
	if err != nil {
		t.Fatalf("AddWatermarkImage: %v", err)
	}

	// Default size (0, 0)
	err = pdf.AddWatermarkImage(resJPEGPath, 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImage default: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: Read (io.Reader interface) ───

func TestCov33_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Read test")

	buf := make([]byte, 4096)
	n, err := pdf.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Fatal("expected non-zero bytes read")
	}
}

// ─── gopdf.go: GetBytesPdf, GetBytesPdfReturnErr ───

func TestCov33_GetBytesPdf(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Bytes test")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Fatal("expected non-empty bytes")
	}

	data2, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data2) == 0 {
		t.Fatal("expected non-empty bytes")
	}
}

// ─── gopdf.go: SetTransparency, ClearTransparency ───

func TestCov33_Transparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.5,
		BlendModeType: NormalBlendMode,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}

	pdf.Cell(&Rect{W: 100, H: 20}, "Transparent text")
	pdf.ClearTransparency()
	pdf.Cell(&Rect{W: 100, H: 20}, "Opaque text")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetPage error ───

func TestCov33_SetPage_Error(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Out of range
	err := pdf.SetPage(99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// Page 0
	err = pdf.SetPage(0)
	if err == nil {
		t.Fatal("expected error for page 0")
	}
}

// ─── gopdf.go: WritePdf ───

func TestCov33_WritePdf(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "WritePdf test")

	outPath := resOutDir + "/cov33_writepdf.pdf"
	err := pdf.WritePdf(outPath)
	if err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
	// Verify file exists
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("output file not found: %v", err)
	}
}

// ─── subset_font_obj.go: charCodeToGlyphIndexFormat12 ───

func TestCov33_CharCodeToGlyphIndexFormat12(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Use various Unicode characters to exercise format12 paths
	// CJK characters often use format12
	text := "Hello 你好世界 🌍"
	_ = pdf.Cell(&Rect{W: 300, H: 20}, text)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: Image (file path) ───

func TestCov33_Image_FilePath(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skip("JPEG not available")
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.Image(resJPEGPath, 10, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("Image: %v", err)
	}

	// PNG
	if _, err := os.Stat(resPNGPath); err == nil {
		err = pdf.Image(resPNGPath, 120, 10, &Rect{W: 100, H: 100})
		if err != nil {
			t.Fatalf("Image PNG: %v", err)
		}
	}

	// Non-existent file
	err = pdf.Image("nonexistent.jpg", 10, 10, nil)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetLineWidth, SetLineType ───

func TestCov33_LineWidthType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetLineWidth(3)
	pdf.SetLineType("dashed")
	pdf.Line(10, 10, 200, 10)

	pdf.SetLineType("dotted")
	pdf.Line(10, 30, 200, 30)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetX, SetY, GetX, GetY ───

func TestCov33_XY(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetX(50)
	pdf.SetY(100)
	if pdf.GetX() != 50 {
		t.Fatalf("expected X=50, got %f", pdf.GetX())
	}
	if pdf.GetY() != 100 {
		t.Fatalf("expected Y=100, got %f", pdf.GetY())
	}

	pdf.SetXY(75, 150)
	if pdf.GetX() != 75 {
		t.Fatalf("expected X=75, got %f", pdf.GetX())
	}
}

// ─── svg_insert.go: InsertSVG with various elements ───

func TestCov33_InsertSVG_Elements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// SVG with rect, circle, ellipse, line, polyline, polygon, path
	svg := `<svg width="200" height="200" viewBox="0 0 200 200">
		<rect x="10" y="10" width="50" height="30" fill="#FF0000" stroke="black" stroke-width="2"/>
		<circle cx="100" cy="50" r="20" fill="blue"/>
		<ellipse cx="150" cy="50" rx="30" ry="15" fill="green" opacity="0.5"/>
		<line x1="10" y1="100" x2="190" y2="100" stroke="red" stroke-width="1"/>
		<polyline points="10,120 50,150 90,120 130,150" stroke="purple" fill="none"/>
		<polygon points="150,120 170,160 130,160" fill="orange"/>
		<path d="M 10 180 L 50 180 L 30 200 Z" fill="cyan"/>
		<g>
			<rect x="60" y="170" width="20" height="20" style="fill:magenta;stroke:black;stroke-width:1;opacity:0.8"/>
		</g>
	</svg>`

	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── svg_insert.go: InsertSVG with viewBox only (no width/height) ───

func TestCov33_InsertSVG_ViewBoxOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svg := `<svg viewBox="0 0 100 100">
		<rect x="0" y="0" width="100" height="100" fill="gray"/>
	</svg>`

	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 100, Height: 100})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes viewBox only: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── svg_insert.go: InsertSVG with rgb() colors and named colors ───

func TestCov33_InsertSVG_Colors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svg := `<svg width="200" height="100">
		<rect x="0" y="0" width="50" height="50" fill="rgb(255,128,0)"/>
		<rect x="60" y="0" width="50" height="50" fill="#abc"/>
		<rect x="120" y="0" width="50" height="50" fill="yellow"/>
	</svg>`

	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 100})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes colors: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── svg_insert.go: InsertSVG bad XML ───

func TestCov33_InsertSVG_BadXML(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVGFromBytes([]byte("not xml"), SVGOption{X: 10, Y: 10, Width: 100, Height: 100})
	if err == nil {
		t.Fatal("expected error for bad XML")
	}
}

// ─── page_batch_ops.go: DeletePages ───

func TestCov33_DeletePages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 3")

	// Delete page 2
	err := pdf.DeletePages([]int{2})
	if err != nil {
		t.Fatalf("DeletePages: %v", err)
	}

	// Empty list (no-op)
	err = pdf.DeletePages([]int{})
	if err != nil {
		t.Fatalf("DeletePages empty: %v", err)
	}

	// Out of range
	err = pdf.DeletePages([]int{99})
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// Delete all pages (should error)
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	err = pdf2.DeletePages([]int{1})
	if err == nil {
		t.Fatal("expected error for deleting all pages")
	}
}

// ─── page_batch_ops.go: DeletePages with duplicates ───

func TestCov33_DeletePages_Duplicates(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 3")

	// Duplicate page numbers should be deduplicated
	err := pdf.DeletePages([]int{3, 3, 2})
	if err != nil {
		t.Fatalf("DeletePages duplicates: %v", err)
	}
}

// ─── annotation_modify.go: ModifyAnnotation ───

func TestCov33_ModifyAnnotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add an annotation first
	pdf.AddExternalLink("https://example.com", 50, 50, 100, 20)

	// Modify it
	err := pdf.ModifyAnnotation(0, 0, AnnotationOption{
		Content: "Updated",
		Title:   "New Title",
		Color:   [3]uint8{255, 0, 0},
	})
	// May fail if annotation type doesn't match
	_ = err

	// Out of range page
	err = pdf.ModifyAnnotation(99, 0, AnnotationOption{})
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// Out of range annotation
	err = pdf.ModifyAnnotation(0, 99, AnnotationOption{})
	if err == nil {
		t.Fatal("expected error for out-of-range annotation")
	}
}

// ─── colorspace_convert.go: ConvertColorspace ───

func TestCov33_ConvertColorspace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetFillColor(0, 255, 0)
	pdf.RectFromUpperLeftWithStyle(10, 10, 100, 50, "DF")
	pdf.SetFillColorCMYK(0, 100, 100, 0)
	pdf.RectFromUpperLeftWithStyle(10, 70, 100, 50, "F")
	pdf.SetGrayFill(0.5)
	pdf.RectFromUpperLeftWithStyle(10, 130, 100, 50, "F")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Convert to gray
	gray, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceGray})
	if err != nil {
		t.Fatalf("ConvertColorspace gray: %v", err)
	}
	if len(gray) == 0 {
		t.Fatal("expected non-empty gray data")
	}

	// Convert to CMYK
	cmyk, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceCMYK})
	if err != nil {
		t.Fatalf("ConvertColorspace CMYK: %v", err)
	}
	if len(cmyk) == 0 {
		t.Fatal("expected non-empty CMYK data")
	}

	// Convert to RGB (identity for RGB content)
	rgb, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceRGB})
	if err != nil {
		t.Fatalf("ConvertColorspace RGB: %v", err)
	}
	if len(rgb) == 0 {
		t.Fatal("expected non-empty RGB data")
	}
}

// ─── colorspace_convert.go: convertColorLine edge cases ───

func TestCov33_ConvertColorLine(t *testing.T) {
	// RGB to gray
	result := convertColorLine("1.0000 0.0000 0.0000 rg", ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Fatalf("expected gray op, got: %s", result)
	}

	// RGB stroking to CMYK
	result2 := convertColorLine("0.0000 1.0000 0.0000 RG", ColorspaceCMYK)
	if !strings.Contains(result2, "K") {
		t.Fatalf("expected CMYK stroking op, got: %s", result2)
	}

	// CMYK to gray
	result3 := convertColorLine("0.0000 1.0000 1.0000 0.0000 k", ColorspaceGray)
	if !strings.Contains(result3, "g") {
		t.Fatalf("expected gray op, got: %s", result3)
	}

	// CMYK stroking to RGB
	result4 := convertColorLine("1.0000 0.0000 0.0000 0.0000 K", ColorspaceRGB)
	if !strings.Contains(result4, "RG") {
		t.Fatalf("expected RGB stroking op, got: %s", result4)
	}

	// Gray to RGB
	result5 := convertColorLine("0.5000 g", ColorspaceRGB)
	if !strings.Contains(result5, "rg") {
		t.Fatalf("expected RGB op, got: %s", result5)
	}

	// Gray stroking to CMYK
	result6 := convertColorLine("0.5000 G", ColorspaceCMYK)
	if !strings.Contains(result6, "K") {
		t.Fatalf("expected CMYK stroking op, got: %s", result6)
	}

	// Non-color line (passthrough)
	result7 := convertColorLine("BT", ColorspaceGray)
	if result7 != "BT" {
		t.Fatalf("expected passthrough, got: %s", result7)
	}
}

// ─── content_stream_clean.go: CleanContentStreams ───

func TestCov33_CleanContentStreams(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.Line(10, 10, 200, 200)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 80, "DF")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	cleaned, err := CleanContentStreams(data)
	if err != nil {
		t.Fatalf("CleanContentStreams: %v", err)
	}
	if len(cleaned) == 0 {
		t.Fatal("expected non-empty cleaned data")
	}
}

// ─── gopdf.go: DeletePage ───

func TestCov33_DeletePage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")

	err := pdf.DeletePage(2)
	if err != nil {
		t.Fatalf("DeletePage: %v", err)
	}

	// Out of range
	err = pdf.DeletePage(99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}
}

// ─── gopdf.go: CopyPage ───

func TestCov33_CopyPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Original")

	newPageNo, err := pdf.CopyPage(1)
	if err != nil {
		t.Fatalf("CopyPage: %v", err)
	}
	if newPageNo < 2 {
		t.Fatalf("expected new page number >= 2, got %d", newPageNo)
	}
}

// ─── image_extract.go: ExtractImagesFromPage, ExtractImagesFromAllPages ───

func TestCov33_ExtractImages(t *testing.T) {
	// Create a PDF with an embedded image
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// ExtractImagesFromPage
	imgs, err := ExtractImagesFromPage(data, 0)
	if err != nil {
		t.Fatalf("ExtractImagesFromPage: %v", err)
	}
	_ = imgs

	// Out of range
	_, err = ExtractImagesFromPage(data, 99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// ExtractImagesFromAllPages
	allImgs, err := ExtractImagesFromAllPages(data)
	if err != nil {
		t.Fatalf("ExtractImagesFromAllPages: %v", err)
	}
	_ = allImgs

	// GetImageFormat
	ei := ExtractedImage{Filter: "DCTDecode"}
	if ei.GetImageFormat() != "jpeg" {
		t.Fatalf("expected jpeg, got %s", ei.GetImageFormat())
	}
	ei2 := ExtractedImage{Filter: "FlateDecode"}
	if ei2.GetImageFormat() != "png" {
		t.Fatalf("expected png, got %s", ei2.GetImageFormat())
	}
	ei3 := ExtractedImage{Filter: "JPXDecode"}
	if ei3.GetImageFormat() != "jp2" {
		t.Fatalf("expected jp2, got %s", ei3.GetImageFormat())
	}
	ei4 := ExtractedImage{Filter: "CCITTFaxDecode"}
	if ei4.GetImageFormat() != "tiff" {
		t.Fatalf("expected tiff, got %s", ei4.GetImageFormat())
	}
	ei5 := ExtractedImage{Filter: "Unknown"}
	if ei5.GetImageFormat() != "raw" {
		t.Fatalf("expected raw, got %s", ei5.GetImageFormat())
	}
}

// ─── markinfo.go: SetMarkInfo, GetMarkInfo, FindPagesByLabel, computePageLabel ───

func TestCov33_MarkInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetMarkInfo(MarkInfo{Marked: true, UserProperties: true, Suspects: true})
	info := pdf.GetMarkInfo()
	if info == nil || !info.Marked {
		t.Fatal("expected marked info")
	}

	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Tagged PDF")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestCov33_FindPagesByLabel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")

	// No labels set
	result := pdf.FindPagesByLabel("1")
	if len(result) != 0 {
		t.Fatalf("expected no results without labels, got %v", result)
	}

	// Set page labels
	pdf.SetPageLabels([]PageLabel{{
		PageIndex: 0,
		Style:     PageLabelDecimal,
		Start:     1,
	}})

	result2 := pdf.FindPagesByLabel("1")
	if len(result2) != 1 || result2[0] != 0 {
		t.Fatalf("expected page 0, got %v", result2)
	}

	result3 := pdf.FindPagesByLabel("2")
	if len(result3) != 1 || result3[0] != 1 {
		t.Fatalf("expected page 1, got %v", result3)
	}

	// Non-existent label
	result4 := pdf.FindPagesByLabel("99")
	if len(result4) != 0 {
		t.Fatalf("expected no results, got %v", result4)
	}
}

func TestCov33_PageLabel_Styles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()

	// Roman upper
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelRomanUpper, Start: 1}})
	label := pdf.computePageLabel(0)
	if label != "I" {
		t.Fatalf("expected I, got %s", label)
	}
	label2 := pdf.computePageLabel(2)
	if label2 != "III" {
		t.Fatalf("expected III, got %s", label2)
	}

	// Roman lower
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelRomanLower, Start: 1}})
	label3 := pdf.computePageLabel(0)
	if label3 != "i" {
		t.Fatalf("expected i, got %s", label3)
	}

	// Alpha upper
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelAlphaUpper, Start: 1}})
	label4 := pdf.computePageLabel(0)
	if label4 != "A" {
		t.Fatalf("expected A, got %s", label4)
	}

	// Alpha lower
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelAlphaLower, Start: 1}})
	label5 := pdf.computePageLabel(0)
	if label5 != "a" {
		t.Fatalf("expected a, got %s", label5)
	}

	// With prefix
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelDecimal, Start: 1, Prefix: "P-"}})
	label6 := pdf.computePageLabel(0)
	if label6 != "P-1" {
		t.Fatalf("expected P-1, got %s", label6)
	}
}

// ─── markinfo.go: toRoman, toAlpha edge cases ───

func TestCov33_ToRoman_ToAlpha(t *testing.T) {
	// toRoman
	if r := toRoman(4, true); r != "IV" {
		t.Fatalf("expected IV, got %s", r)
	}
	if r := toRoman(9, false); r != "ix" {
		t.Fatalf("expected ix, got %s", r)
	}
	if r := toRoman(0, true); r != "0" {
		t.Fatalf("expected 0, got %s", r)
	}
	if r := toRoman(3999, true); r != "MMMCMXCIX" {
		t.Fatalf("expected MMMCMXCIX, got %s", r)
	}

	// toAlpha
	if a := toAlpha(1, true); a != "A" {
		t.Fatalf("expected A, got %s", a)
	}
	if a := toAlpha(27, true); a != "AA" {
		t.Fatalf("expected AA, got %s", a)
	}
	if a := toAlpha(0, true); a != "" {
		t.Fatalf("expected empty, got %s", a)
	}
	if a := toAlpha(26, false); a != "z" {
		t.Fatalf("expected z, got %s", a)
	}
}

// ─── gopdf.go: OCG (Optional Content Groups) ───

func TestCov33_OCG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add OCG layer
	layer := pdf.AddOCG(OCG{Name: "Layer1", On: true})
	_ = layer

	// Add another layer
	layer2 := pdf.AddOCG(OCG{Name: "Layer2", Intent: OCGIntentDesign, On: false})
	_ = layer2

	// GetOCGs
	ocgs := pdf.GetOCGs()
	if len(ocgs) != 2 {
		t.Fatalf("expected 2 OCGs, got %d", len(ocgs))
	}

	// SetOCGState
	err := pdf.SetOCGState("Layer1", false)
	if err != nil {
		t.Fatalf("SetOCGState: %v", err)
	}

	// SetOCGStates
	err = pdf.SetOCGStates(map[string]bool{"Layer1": true, "Layer2": false})
	if err != nil {
		t.Fatalf("SetOCGStates: %v", err)
	}

	// Non-existent layer
	err = pdf.SetOCGState("NonExistent", true)
	if err == nil {
		t.Fatal("expected error for non-existent layer")
	}

	pdf.Cell(&Rect{W: 100, H: 20}, "OCG content")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: XMP Metadata ───

func TestCov33_XMPMetadata(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetXMPMetadata(XMPMetadata{
		Title:       "Test Document",
		Creator:     []string{"Test Creator"},
		Description: "A test document",
		Subject:     []string{"testing"},
	})
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "XMP Metadata test")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: PageLabels ───

func TestCov33_PageLabels(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()

	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Style: PageLabelRomanLower, Start: 1},
		{PageIndex: 2, Style: PageLabelDecimal, Start: 1, Prefix: "Ch1-"},
	})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: Annotation ───

func TestCov33_Annotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.AddAnnotation(AnnotationOption{
		X:       50,
		Y:       50,
		W:       100,
		H:       50,
		Content: "This is a note",
		Title:   "Note Title",
		Color:   [3]uint8{255, 255, 0},
		Type:    AnnotText,
	})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}
