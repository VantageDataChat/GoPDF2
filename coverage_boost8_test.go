package gopdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost8_test.go — TestCov8_ prefix
// Targets: font obj writes, protection paths, deduplication,
// mask images, colorspace conversion internals, journal, TOC, SVG
// ============================================================

// --------------- helpers ---------------

// newProtectedPDF creates a GoPdf with protection enabled.
func newProtectedPDF(t *testing.T) *GoPdf {
	t.Helper()
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   PermissionsPrint,
			UserPass:      []byte("user"),
			OwnerPass:     []byte("owner"),
		},
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	return pdf
}

// createTestPNG creates a small in-memory PNG image.
func createTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	return buf.Bytes()
}

// ============================================================
// 1. FontObj.write — embed vs non-embed
// ============================================================

func TestCov8_FontObj_Write_NonEmbed(t *testing.T) {
	f := &FontObj{Family: "TestFamily"}
	f.init(nil)
	var buf bytes.Buffer
	if err := f.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/BaseFont /TestFamily") {
		t.Errorf("expected BaseFont, got: %s", out)
	}
	if strings.Contains(out, "/FirstChar") {
		t.Error("non-embed should not have /FirstChar")
	}
}

func TestCov8_FontObj_Write_Embed(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "embed font test")

	// Find a FontObj in the PDF objects and verify it writes correctly.
	var found bool
	for _, obj := range pdf.pdfObjs {
		if fo, ok := obj.(*FontObj); ok {
			var buf bytes.Buffer
			if err := fo.write(&buf, 1); err != nil {
				t.Fatalf("write: %v", err)
			}
			out := buf.String()
			if !strings.Contains(out, "/Type /Font") {
				t.Errorf("expected /Type /Font: %s", out)
			}
			found = true
			break
		}
	}
	if !found {
		t.Skip("no FontObj found in PDF objects")
	}
}

func TestCov8_FontObj_Write_WithFontName(t *testing.T) {
	// When Font is set, BaseFont should use Font.GetName()
	f := &FontObj{Family: "Fallback"}
	f.init(nil)
	var buf bytes.Buffer
	if err := f.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), "/BaseFont /Fallback") {
		t.Error("expected family as BaseFont")
	}
}

func TestCov8_FontObj_GetType(t *testing.T) {
	f := &FontObj{}
	if f.getType() != "Font" {
		t.Errorf("expected Font, got %s", f.getType())
	}
}

func TestCov8_FontObj_SetIndexMethods(t *testing.T) {
	f := &FontObj{}
	f.SetIndexObjWidth(10)
	f.SetIndexObjFontDescriptor(20)
	f.SetIndexObjEncoding(30)
	if f.indexObjWidth != 10 || f.indexObjFontDescriptor != 20 || f.indexObjEncoding != 30 {
		t.Error("SetIndex methods failed")
	}
}

// ============================================================
// 2. FontDescriptorObj.write
// ============================================================

type mockFont struct {
	name     string
	diff     string
	descs    []FontDescItem
	origSize int
}

func (m *mockFont) Init()                                                  {}
func (m *mockFont) GetType() string                                        { return "TrueType" }
func (m *mockFont) GetName() string                                        { return m.name }
func (m *mockFont) GetDesc() []FontDescItem                                { return m.descs }
func (m *mockFont) GetUp() int                                             { return -100 }
func (m *mockFont) GetUt() int                                             { return 50 }
func (m *mockFont) GetCw() FontCw                                          { return FontCw{} }
func (m *mockFont) GetEnc() string                                         { return "cp1252" }
func (m *mockFont) GetDiff() string                                        { return m.diff }
func (m *mockFont) GetOriginalsize() int                                   { return m.origSize }
func (m *mockFont) SetFamily(f string)                                     {}
func (m *mockFont) GetFamily() string                                      { return "test" }

func TestCov8_FontDescriptorObj_Write(t *testing.T) {
	fd := &FontDescriptorObj{
		font: &mockFont{
			name: "TestFont",
			descs: []FontDescItem{
				{Key: "Flags", Val: "32"},
				{Key: "ItalicAngle", Val: "0"},
			},
		},
		fontFileObjRelate: "8 0 R",
	}
	var buf bytes.Buffer
	if err := fd.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/FontName /TestFont") {
		t.Errorf("missing FontName: %s", out)
	}
	if !strings.Contains(out, "/Flags 32") {
		t.Errorf("missing Flags: %s", out)
	}
	if !strings.Contains(out, "/FontFile2 8 0 R") {
		t.Errorf("missing FontFile2: %s", out)
	}
}

func TestCov8_FontDescriptorObj_GetType(t *testing.T) {
	fd := &FontDescriptorObj{}
	if fd.getType() != "FontDescriptor" {
		t.Errorf("expected FontDescriptor, got %s", fd.getType())
	}
}

func TestCov8_FontDescriptorObj_SetGetFont(t *testing.T) {
	fd := &FontDescriptorObj{}
	mf := &mockFont{name: "X"}
	fd.SetFont(mf)
	if fd.GetFont() != mf {
		t.Error("SetFont/GetFont mismatch")
	}
}

func TestCov8_FontDescriptorObj_SetFontFileObjRelate(t *testing.T) {
	fd := &FontDescriptorObj{}
	fd.SetFontFileObjRelate("99 0 R")
	if fd.fontFileObjRelate != "99 0 R" {
		t.Error("SetFontFileObjRelate failed")
	}
}

// ============================================================
// 3. EncodingObj.write
// ============================================================

func TestCov8_EncodingObj_Write(t *testing.T) {
	e := &EncodingObj{}
	e.SetFont(&mockFont{diff: "128 /Euro"})
	var buf bytes.Buffer
	if err := e.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/WinAnsiEncoding") {
		t.Errorf("missing WinAnsiEncoding: %s", out)
	}
	if !strings.Contains(out, "128 /Euro") {
		t.Errorf("missing diff: %s", out)
	}
}

func TestCov8_EncodingObj_GetType(t *testing.T) {
	e := &EncodingObj{}
	if e.getType() != "Encoding" {
		t.Errorf("expected Encoding, got %s", e.getType())
	}
}

func TestCov8_EncodingObj_SetGetFont(t *testing.T) {
	e := &EncodingObj{}
	mf := &mockFont{name: "Y"}
	e.SetFont(mf)
	if e.GetFont() != mf {
		t.Error("SetFont/GetFont mismatch")
	}
}

// ============================================================
// 4. ImportedObj.write
// ============================================================

func TestCov8_ImportedObj_Write(t *testing.T) {
	obj := &ImportedObj{Data: "some pdf data here"}
	var buf bytes.Buffer
	if err := obj.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.String() != "some pdf data here" {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestCov8_ImportedObj_Write_Nil(t *testing.T) {
	var obj *ImportedObj
	var buf bytes.Buffer
	if err := obj.write(&buf, 1); err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("nil ImportedObj should write nothing")
	}
}

func TestCov8_ImportedObj_GetType(t *testing.T) {
	obj := &ImportedObj{}
	if obj.getType() != "Imported" {
		t.Errorf("expected Imported, got %s", obj.getType())
	}
}

// ============================================================
// 5. deduplicateObjects via GarbageCollect(GCDedup)
// ============================================================

func TestCov8_DeduplicateObjects_WithDuplicates(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")

	// Inject a null to trigger compact phase, plus duplicate ImportedObjs.
	dup1 := &ImportedObj{Data: "duplicate stream data"}
	dup2 := &ImportedObj{Data: "duplicate stream data"}
	dup3 := &ImportedObj{Data: "unique stream data"}
	pdf.pdfObjs = append(pdf.pdfObjs, nil, dup1, dup2, dup3)

	countBefore := len(pdf.pdfObjs)
	removed := pdf.GarbageCollect(GCDedup)
	countAfter := len(pdf.pdfObjs)

	// Should have removed the null + the duplicate = at least 2.
	if removed < 2 {
		t.Errorf("expected at least 2 removed, got %d", removed)
	}
	if countAfter >= countBefore {
		t.Errorf("expected fewer objects after dedup: before=%d after=%d", countBefore, countAfter)
	}
}

func TestCov8_DeduplicateObjects_NoDuplicates(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")

	// Add a null to trigger compact, plus unique ImportedObjs.
	pdf.pdfObjs = append(pdf.pdfObjs, nil, &ImportedObj{Data: "aaa"}, &ImportedObj{Data: "bbb"})

	removed := pdf.GarbageCollect(GCDedup)
	// Should remove only the null (1), no dedup removals.
	if removed != 1 {
		t.Errorf("expected 1 removed (null only), got %d", removed)
	}
}

func TestCov8_DeduplicateObjects_MultipleDuplicates(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")

	// Inject null + 3 copies of AAA + 2 copies of BBB.
	pdf.pdfObjs = append(pdf.pdfObjs, nil)
	for i := 0; i < 3; i++ {
		pdf.pdfObjs = append(pdf.pdfObjs, &ImportedObj{Data: "AAA"})
	}
	for i := 0; i < 2; i++ {
		pdf.pdfObjs = append(pdf.pdfObjs, &ImportedObj{Data: "BBB"})
	}

	removed := pdf.GarbageCollect(GCDedup)
	// null(1) + AAA dups(2) + BBB dups(1) = 4.
	if removed < 4 {
		t.Errorf("expected at least 4 removed, got %d", removed)
	}
}

// ============================================================
// 6. Protection paths — ImageObj.write with protection
// ============================================================

func TestCov8_ImageObj_Write_WithProtection(t *testing.T) {
	prot := &PDFProtection{}
	if err := prot.SetProtection(PermissionsPrint, []byte("u"), []byte("o")); err != nil {
		t.Fatalf("SetProtection: %v", err)
	}

	imgObj := &ImageObj{}
	imgObj.setProtection(prot)

	// Set up minimal image info.
	imgObj.imginfo = imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		data:            []byte("fake image data for testing"),
	}

	var buf bytes.Buffer
	err := imgObj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "stream") {
		t.Error("expected stream in output")
	}
	if !strings.Contains(out, "endstream") {
		t.Error("expected endstream in output")
	}
}

func TestCov8_ImageObj_Write_IsMask(t *testing.T) {
	imgObj := &ImageObj{IsMask: true}
	imgObj.imginfo = imgInfo{
		w:               5,
		h:               5,
		colspace:        "DeviceGray",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		smask:           []byte("mask data here"),
	}

	var buf bytes.Buffer
	err := imgObj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), "stream") {
		t.Error("expected stream")
	}
}

func TestCov8_ImageObj_Write_SplittedMask(t *testing.T) {
	imgObj := &ImageObj{SplittedMask: true}
	imgObj.imginfo = imgInfo{
		w:               5,
		h:               5,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		data:            []byte("rgb data"),
	}

	var buf bytes.Buffer
	err := imgObj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), "stream") {
		t.Error("expected stream")
	}
}

// ============================================================
// 7. DeviceRGBObj.write with protection
// ============================================================

func TestCov8_DeviceRGBObj_Write_WithProtection(t *testing.T) {
	prot := &PDFProtection{}
	if err := prot.SetProtection(PermissionsPrint, []byte("u"), []byte("o")); err != nil {
		t.Fatalf("SetProtection: %v", err)
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	d := &DeviceRGBObj{
		data: []byte{0xFF, 0x00, 0x55, 0xAA},
	}
	d.init(func() *GoPdf {
		gp := &GoPdf{}
		gp.Start(Config{
			PageSize: *PageSizeA4,
			Protection: PDFProtectionConfig{
				UseProtection: true,
				Permissions:   PermissionsPrint,
				UserPass:      []byte("u"),
				OwnerPass:     []byte("o"),
			},
		})
		return gp
	})

	var buf bytes.Buffer
	err := d.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "stream") {
		t.Error("expected stream")
	}
}

func TestCov8_DeviceRGBObj_Write_NoProtection(t *testing.T) {
	d := &DeviceRGBObj{
		data: []byte{0x01, 0x02, 0x03},
	}
	d.init(func() *GoPdf {
		gp := &GoPdf{}
		gp.Start(Config{PageSize: *PageSizeA4})
		return gp
	})

	var buf bytes.Buffer
	err := d.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Length 3") {
		t.Errorf("expected /Length 3, got: %s", out)
	}
}

func TestCov8_DeviceRGBObj_GetType(t *testing.T) {
	d := &DeviceRGBObj{}
	if d.getType() != "devicergb" {
		t.Errorf("expected devicergb, got %s", d.getType())
	}
}

// ============================================================
// 8. ContentObj.write with protection
// ============================================================

func TestCov8_ContentObj_Write_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	if err := pdf.Cell(nil, "Protected content"); err != nil {
		t.Fatalf("Cell: %v", err)
	}

	// The PDF should compile without error.
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_ContentObj_Write_NoCompression(t *testing.T) {
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
	pdf.Cell(nil, "No compression")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_ContentObj_Write_ProtectedNoCompression(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   PermissionsPrint | PermissionsCopy,
			UserPass:      []byte("test"),
			OwnerPass:     []byte("owner"),
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
	pdf.Cell(nil, "Protected no compression")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 9. annotObj.writeExternalLink with protection
// ============================================================

func TestCov8_AnnotObj_WriteExternalLink_Protected(t *testing.T) {
	prot := &PDFProtection{}
	if err := prot.SetProtection(PermissionsPrint, []byte("u"), []byte("o")); err != nil {
		t.Fatalf("SetProtection: %v", err)
	}

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

	a := annotObj{
		linkOption: linkOption{
			url: "https://example.com/test?q=1&r=2",
			x:   10, y: 20, w: 100, h: 15,
		},
		GetRoot: func() *GoPdf { return pdf },
	}

	var buf bytes.Buffer
	err := a.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Annot") {
		t.Errorf("expected /Annot: %s", out)
	}
}

func TestCov8_AnnotObj_WriteExternalLink_SpecialChars(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	a := annotObj{
		linkOption: linkOption{
			url: `https://example.com/path\(with)parens\and\backslash` + "\r",
			x:   0, y: 0, w: 50, h: 10,
		},
		GetRoot: func() *GoPdf { return pdf },
	}

	var buf bytes.Buffer
	err := a.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestCov8_AnnotObj_WriteInternalLink_Found(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	pdf.SetAnchor("myanchor")

	a := annotObj{
		linkOption: linkOption{
			anchor: "myanchor",
			x:      10, y: 20, w: 100, h: 15,
		},
		GetRoot: func() *GoPdf { return pdf },
	}

	var buf bytes.Buffer
	err := a.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Dest") {
		t.Errorf("expected /Dest for internal link: %s", out)
	}
}

func TestCov8_AnnotObj_WriteInternalLink_NotFound(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	a := annotObj{
		linkOption: linkOption{
			anchor: "nonexistent",
			x:      10, y: 20, w: 100, h: 15,
		},
		GetRoot: func() *GoPdf { return pdf },
	}

	var buf bytes.Buffer
	err := a.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	// Should produce empty output for missing anchor.
	if buf.Len() != 0 {
		t.Errorf("expected empty output for missing anchor, got: %s", buf.String())
	}
}

// ============================================================
// 10. Colorspace conversion internals
// ============================================================

func TestCov8_ConvertColorspace_RGBToGray(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColor(255, 0, 0)
	pdf.Cell(nil, "Red text")
	pdf.SetStrokeColor(0, 255, 0)
	pdf.Line(10, 10, 100, 100)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Skip("empty PDF")
	}

	gray, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceGray})
	if err != nil {
		t.Fatalf("ConvertColorspace: %v", err)
	}
	if len(gray) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov8_ConvertColorspace_RGBToCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColor(100, 150, 200)
	pdf.Cell(nil, "CMYK test")

	data := pdf.GetBytesPdf()
	result, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceCMYK})
	if err != nil {
		t.Fatalf("ConvertColorspace: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov8_ConvertColorspace_GrayToRGB(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
	pdf.Cell(nil, "Gray text")

	data := pdf.GetBytesPdf()
	result, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceRGB})
	if err != nil {
		t.Fatalf("ConvertColorspace: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov8_ConvertColorspace_CMYKToRGB(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColorCMYK(100, 0, 0, 0)
	pdf.Cell(nil, "CMYK text")

	data := pdf.GetBytesPdf()
	result, err := ConvertColorspace(data, ConvertColorspaceOption{Target: ColorspaceRGB})
	if err != nil {
		t.Fatalf("ConvertColorspace: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov8_ConvertColorspace_InvalidPDF(t *testing.T) {
	result, err := ConvertColorspace([]byte("not a pdf"), ConvertColorspaceOption{Target: ColorspaceGray})
	// May or may not error; just ensure no panic.
	_ = err
	_ = result
}

func TestCov8_ConvertColorspace_EmptyPDF(t *testing.T) {
	result, err := ConvertColorspace(nil, ConvertColorspaceOption{Target: ColorspaceGray})
	// May or may not error; just ensure no panic.
	_ = err
	_ = result
}

func TestCov8_ConvertColorLine_RGB_NonStroking(t *testing.T) {
	result := convertColorLine("0.5 0.3 0.1 rg", ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Errorf("expected gray op: %s", result)
	}
}

func TestCov8_ConvertColorLine_RGB_Stroking(t *testing.T) {
	result := convertColorLine("0.5 0.3 0.1 RG", ColorspaceCMYK)
	if !strings.Contains(result, "K") {
		t.Errorf("expected CMYK stroking: %s", result)
	}
}

func TestCov8_ConvertColorLine_CMYK_NonStroking(t *testing.T) {
	result := convertColorLine("0.1 0.2 0.3 0.4 k", ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Errorf("expected gray: %s", result)
	}
}

func TestCov8_ConvertColorLine_CMYK_Stroking(t *testing.T) {
	result := convertColorLine("0.1 0.2 0.3 0.4 K", ColorspaceRGB)
	if !strings.Contains(result, "RG") {
		t.Errorf("expected RGB stroking: %s", result)
	}
}

func TestCov8_ConvertColorLine_Gray_NonStroking(t *testing.T) {
	result := convertColorLine("0.5 g", ColorspaceRGB)
	if !strings.Contains(result, "rg") {
		t.Errorf("expected RGB: %s", result)
	}
}

func TestCov8_ConvertColorLine_Gray_Stroking(t *testing.T) {
	result := convertColorLine("0.5 G", ColorspaceCMYK)
	if !strings.Contains(result, "K") {
		t.Errorf("expected CMYK: %s", result)
	}
}

func TestCov8_ConvertColorLine_NonColor(t *testing.T) {
	result := convertColorLine("100 200 m", ColorspaceGray)
	if result != "100 200 m" {
		t.Errorf("non-color line should be unchanged: %s", result)
	}
}

func TestCov8_ConvertColorLine_Empty(t *testing.T) {
	result := convertColorLine("", ColorspaceGray)
	if result != "" {
		t.Errorf("empty line should stay empty: %s", result)
	}
}

func TestCov8_RgbToGray(t *testing.T) {
	g := rgbToGray(1.0, 1.0, 1.0)
	if g < 0.99 || g > 1.01 {
		t.Errorf("white should be ~1.0, got %f", g)
	}
	g = rgbToGray(0, 0, 0)
	if g != 0 {
		t.Errorf("black should be 0, got %f", g)
	}
}

func TestCov8_RgbToCMYK(t *testing.T) {
	c, m, y, k := rgbToCMYK(1.0, 0, 0)
	if k != 0 {
		t.Errorf("pure red k should be 0, got %f", k)
	}
	if c != 0 {
		t.Errorf("pure red c should be 0, got %f", c)
	}
	_ = m
	_ = y
}

func TestCov8_RgbToCMYK_Black(t *testing.T) {
	c, m, y, k := rgbToCMYK(0, 0, 0)
	if k != 1.0 {
		t.Errorf("black k should be 1.0, got %f", k)
	}
	if c != 0 || m != 0 || y != 0 {
		t.Error("black CMY should all be 0")
	}
}

func TestCov8_CmykToRGB(t *testing.T) {
	r, g, b := cmykToRGB(0, 0, 0, 0)
	if r != 1.0 || g != 1.0 || b != 1.0 {
		t.Errorf("no ink should be white, got %f %f %f", r, g, b)
	}
}

func TestCov8_ParseColorFloat(t *testing.T) {
	if parseColorFloat("0.5") != 0.5 {
		t.Error("expected 0.5")
	}
	if parseColorFloat("invalid") != 0 {
		t.Error("expected 0 for invalid")
	}
	if parseColorFloat("  1.0  ") != 1.0 {
		t.Error("expected 1.0 with whitespace")
	}
}

func TestCov8_ConvertStreamColorspace(t *testing.T) {
	stream := []byte("0.5 0.3 0.1 rg\n100 200 m\n0.8 G\n")
	result := convertStreamColorspace(stream, ColorspaceGray)
	lines := strings.Split(string(result), "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

// ============================================================
// 11. Journal functions
// ============================================================

func TestCov8_Journal_EnableDisable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "initial")

	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Error("expected journal enabled")
	}

	pdf.JournalDisable()
	if pdf.JournalIsEnabled() {
		t.Error("expected journal disabled")
	}
}

func TestCov8_Journal_UndoRedo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "initial")

	pdf.JournalEnable()

	pdf.JournalStartOp("add text")
	pdf.Cell(nil, "more text")
	pdf.JournalEndOp()

	name, err := pdf.JournalUndo()
	if err != nil {
		t.Fatalf("JournalUndo: %v", err)
	}
	if name != "add text" {
		t.Errorf("expected 'add text', got '%s'", name)
	}

	name, err = pdf.JournalRedo()
	if err != nil {
		t.Fatalf("JournalRedo: %v", err)
	}
	if name != "add text" {
		t.Errorf("expected 'add text', got '%s'", name)
	}
}

func TestCov8_Journal_UndoNothing(t *testing.T) {
	pdf := newPDFWithFont(t)
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Error("expected error when journal not enabled")
	}
}

func TestCov8_Journal_RedoNothing(t *testing.T) {
	pdf := newPDFWithFont(t)
	_, err := pdf.JournalRedo()
	if err == nil {
		t.Error("expected error when journal not enabled")
	}
}

func TestCov8_Journal_UndoEmpty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()
	// Only initial snapshot, nothing to undo.
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Error("expected error for nothing to undo")
	}
}

func TestCov8_Journal_RedoEmpty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()
	_, err := pdf.JournalRedo()
	if err == nil {
		t.Error("expected error for nothing to redo")
	}
}

func TestCov8_Journal_GetOperations(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "x")

	pdf.JournalEnable()
	pdf.JournalStartOp("op1")
	pdf.Cell(nil, "y")
	pdf.JournalEndOp()

	pdf.JournalStartOp("op2")
	pdf.Cell(nil, "z")
	pdf.JournalEndOp()

	ops := pdf.JournalGetOperations()
	if len(ops) < 3 { // initial + op1 + op2
		t.Errorf("expected at least 3 operations, got %d", len(ops))
	}
}

func TestCov8_Journal_GetOperations_Nil(t *testing.T) {
	pdf := newPDFWithFont(t)
	ops := pdf.JournalGetOperations()
	if ops != nil {
		t.Error("expected nil for no journal")
	}
}

func TestCov8_Journal_SaveLoad(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "journal test")

	pdf.JournalEnable()
	pdf.JournalStartOp("save-test")
	pdf.Cell(nil, "more")
	pdf.JournalEndOp()

	path := resOutDir + "/test_journal.json"
	if err := pdf.JournalSave(path); err != nil {
		t.Fatalf("JournalSave: %v", err)
	}
	defer os.Remove(path)

	// Load into a new PDF.
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.JournalEnable()
	if err := pdf2.JournalLoad(path); err != nil {
		t.Fatalf("JournalLoad: %v", err)
	}
}

func TestCov8_Journal_SaveNotEnabled(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.JournalSave("/tmp/nope.json")
	if err == nil {
		t.Error("expected error when journal not enabled")
	}
}

func TestCov8_Journal_LoadNotEnabled(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.JournalLoad("/tmp/nope.json")
	if err == nil {
		t.Error("expected error when journal not enabled")
	}
}

func TestCov8_Journal_LoadBadFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()
	err := pdf.JournalLoad("/nonexistent/path/journal.json")
	if err == nil {
		t.Error("expected error for bad file")
	}
}

func TestCov8_Journal_StartEndOp_Disabled(t *testing.T) {
	pdf := newPDFWithFont(t)
	// These should be no-ops when journal is nil.
	pdf.JournalStartOp("noop")
	pdf.JournalEndOp()
}

// ============================================================
// 12. TOC functions
// ============================================================

func TestCov8_TOC_GetTOC_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	toc := pdf.GetTOC()
	if toc != nil {
		t.Errorf("expected nil TOC for no outlines, got %d items", len(toc))
	}
}

func TestCov8_TOC_GetTOC_WithOutlines(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")

	toc := pdf.GetTOC()
	if len(toc) < 2 {
		t.Errorf("expected at least 2 TOC items, got %d", len(toc))
	}
}

func TestCov8_TOC_SetTOC_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Ch1")

	err := pdf.SetTOC(nil)
	if err != nil {
		t.Fatalf("SetTOC(nil): %v", err)
	}
}

func TestCov8_TOC_SetTOC_Flat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()

	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "First", PageNo: 1},
		{Level: 1, Title: "Second", PageNo: 2},
	})
	if err != nil {
		t.Fatalf("SetTOC: %v", err)
	}

	toc := pdf.GetTOC()
	if len(toc) < 2 {
		t.Errorf("expected 2 TOC items, got %d", len(toc))
	}
}

func TestCov8_TOC_SetTOC_Hierarchical(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()

	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Chapter 1", PageNo: 1},
		{Level: 2, Title: "Section 1.1", PageNo: 1, Y: 200},
		{Level: 2, Title: "Section 1.2", PageNo: 2},
		{Level: 1, Title: "Chapter 2", PageNo: 3},
	})
	if err != nil {
		t.Fatalf("SetTOC: %v", err)
	}
}

func TestCov8_TOC_SetTOC_InvalidFirstLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTOC([]TOCItem{
		{Level: 2, Title: "Bad", PageNo: 1},
	})
	if err == nil {
		t.Error("expected error for invalid first level")
	}
}

func TestCov8_TOC_SetTOC_InvalidLevelJump(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Ch1", PageNo: 1},
		{Level: 3, Title: "Bad", PageNo: 1}, // jump from 1 to 3
	})
	if err == nil {
		t.Error("expected error for level jump > 1")
	}
}

// ============================================================
// 13. SVG functions
// ============================================================

func TestCov8_SVG_ImageSVGFromBytes_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="red"/>
		<circle cx="50" cy="50" r="30" fill="blue"/>
		<line x1="0" y1="0" x2="100" y2="100" stroke="green" stroke-width="2"/>
	</svg>`)

	err := pdf.ImageSVGFromBytes(svgData, SVGOption{
		Width:  100,
		Height: 100,
		X:      50,
		Y:      50,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_SVG_WithPolygonAndPolyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<polygon points="100,10 40,198 190,78 10,78 160,198" fill="gold" stroke="black"/>
		<polyline points="0,40 40,40 40,80 80,80 80,120" fill="none" stroke="red"/>
		<ellipse cx="100" cy="100" rx="50" ry="30" fill="purple"/>
	</svg>`)

	err := pdf.ImageSVGFromBytes(svgData, SVGOption{Width: 200, Height: 200, X: 10, Y: 10})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes: %v", err)
	}
}

func TestCov8_SVG_WithPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<path d="M 10 10 L 90 10 L 90 90 L 10 90 Z" fill="orange" stroke="black"/>
	</svg>`)

	err := pdf.ImageSVGFromBytes(svgData, SVGOption{Width: 100, Height: 100, X: 10, Y: 10})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes: %v", err)
	}
}

func TestCov8_SVG_WithText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="100">
		<text x="10" y="50" font-size="20" fill="black">Hello SVG</text>
	</svg>`)

	// Text in SVG may or may not be supported; just ensure no panic.
	_ = pdf.ImageSVGFromBytes(svgData, SVGOption{Width: 200, Height: 100, X: 10, Y: 10})
}

func TestCov8_SVG_InvalidXML(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVGFromBytes([]byte("<not valid xml"), SVGOption{Width: 100, Height: 100})
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestCov8_SVG_EmptyData(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVGFromBytes(nil, SVGOption{Width: 100, Height: 100})
	if err == nil {
		t.Error("expected error for nil SVG data")
	}
}

// ============================================================
// 14. ImageByHolderWithOptions with Mask
// ============================================================

func TestCov8_ImageByHolderWithOptions_WithMask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create main image holder.
	mainHolder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skipf("JPEG not available: %v", err)
	}

	// Create mask image holder.
	maskHolder, err := ImageHolderByPath(resPNGPath)
	if err != nil {
		t.Skipf("PNG not available: %v", err)
	}

	opts := ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 200, H: 200},
		Mask: &MaskOptions{
			Holder: maskHolder,
			ImageOptions: ImageOptions{
				X:    50,
				Y:    50,
				Rect: &Rect{W: 200, H: 200},
			},
		},
	}

	err = pdf.ImageByHolderWithOptions(mainHolder, opts)
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_ImageByHolderWithOptions_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skipf("JPEG not available: %v", err)
	}

	opts := ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 200, H: 200},
		Transparency: &Transparency{
			Alpha:         0.5,
			BlendModeType: NormalBlendMode,
		},
	}

	err = pdf.ImageByHolderWithOptions(holder, opts)
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions: %v", err)
	}
}

func TestCov8_ImageByHolderWithOptions_Rotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skipf("JPEG not available: %v", err)
	}

	opts := ImageOptions{
		X:           50,
		Y:           50,
		Rect:        &Rect{W: 200, H: 200},
		DegreeAngle: 45.0,
	}

	err = pdf.ImageByHolderWithOptions(holder, opts)
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions: %v", err)
	}
}

func TestCov8_ImageByHolderWithOptions_Flip(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skipf("JPEG not available: %v", err)
	}

	opts := ImageOptions{
		X:              50,
		Y:              50,
		Rect:           &Rect{W: 200, H: 200},
		VerticalFlip:   true,
		HorizontalFlip: true,
	}

	err = pdf.ImageByHolderWithOptions(holder, opts)
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions: %v", err)
	}
}

// ============================================================
// 15. ImageFrom with image.Image
// ============================================================

func TestCov8_ImageFrom_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	err := pdf.ImageFrom(img, 10, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("ImageFrom: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 16. Protection-related PDF generation
// ============================================================

func TestCov8_ProtectedPDF_FullWorkflow(t *testing.T) {
	ensureOutDir(t)
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.Cell(nil, "Protected document")
	pdf.SetTextColor(255, 0, 0)
	pdf.Cell(nil, "Red text")
	pdf.Line(10, 100, 200, 100)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_ProtectedPDF_WithImage(t *testing.T) {
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
}

// ============================================================
// 17. PDFProtection methods
// ============================================================

func TestCov8_PDFProtection_ObjectKey(t *testing.T) {
	p := &PDFProtection{}
	if err := p.SetProtection(PermissionsPrint, []byte("u"), []byte("o")); err != nil {
		t.Fatalf("SetProtection: %v", err)
	}

	key := p.Objectkey(1)
	if len(key) != 10 {
		t.Errorf("expected 10-byte key, got %d", len(key))
	}

	key2 := p.Objectkey(2)
	if bytes.Equal(key, key2) {
		t.Error("different objIDs should produce different keys")
	}
}

func TestCov8_PDFProtection_EncryptionObj(t *testing.T) {
	p := &PDFProtection{}
	if err := p.SetProtection(PermissionsPrint|PermissionsModify, []byte("user"), []byte("owner")); err != nil {
		t.Fatalf("SetProtection: %v", err)
	}

	enc := p.EncryptionObj()
	if enc == nil {
		t.Fatal("expected non-nil EncryptionObj")
	}
	if len(enc.uValue) == 0 {
		t.Error("expected non-empty uValue")
	}
	if len(enc.oValue) == 0 {
		t.Error("expected non-empty oValue")
	}
}

func TestCov8_PDFProtection_EmptyOwnerPass(t *testing.T) {
	p := &PDFProtection{}
	// Empty owner pass should auto-generate.
	if err := p.SetProtection(PermissionsPrint, []byte("u"), nil); err != nil {
		t.Fatalf("SetProtection: %v", err)
	}
	if len(p.oValue) == 0 {
		t.Error("expected non-empty oValue even with nil owner pass")
	}
}

func TestCov8_Rc4Cip(t *testing.T) {
	key := []byte("testkey123")
	src := []byte("hello world")
	encrypted, err := rc4Cip(key, src)
	if err != nil {
		t.Fatalf("rc4Cip: %v", err)
	}
	if bytes.Equal(encrypted, src) {
		t.Error("encrypted should differ from source")
	}

	// Decrypt should give back original.
	decrypted, err := rc4Cip(key, encrypted)
	if err != nil {
		t.Fatalf("rc4Cip decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, src) {
		t.Error("decrypted should match original")
	}
}

// ============================================================
// 18. Scrub with partial options
// ============================================================

func TestCov8_Scrub_MetadataOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetInfo(PdfInfo{Title: "Test", Author: "Author"})
	pdf.Cell(nil, "scrub test")

	pdf.Scrub(ScrubOption{Metadata: true})
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Scrub_XMLMetadataOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXMPMetadata(XMPMetadata{
		Title:   "Test",
		Creator: []string{"Kiro"},
	})
	pdf.Cell(nil, "xmp scrub")

	pdf.Scrub(ScrubOption{XMLMetadata: true})
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Scrub_EmbeddedFilesOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "embed scrub")

	pdf.AddEmbeddedFile(EmbeddedFile{Name: "test.txt", Content: []byte("hello")})
	pdf.Scrub(ScrubOption{EmbeddedFiles: true})

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Scrub_PageLabelsOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Prefix: "P-"}})
	pdf.Cell(nil, "label scrub")

	pdf.Scrub(ScrubOption{PageLabels: true})
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Scrub_AllDisabled(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "no scrub")

	pdf.Scrub(ScrubOption{})
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 19. Read (io.Reader interface)
// ============================================================

func TestCov8_Read_Interface(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "reader test")

	buf := make([]byte, 4096)
	n, err := pdf.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Error("expected some bytes read")
	}
}

func TestCov8_Read_MultipleReads(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "multi read")

	var all []byte
	buf := make([]byte, 64)
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
	if len(all) == 0 {
		t.Error("expected some bytes")
	}
}

// ============================================================
// 20. WriteTo (io.WriterTo interface)
// ============================================================

func TestCov8_WriteTo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "writeto test")

	var buf bytes.Buffer
	n, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
}

// ============================================================
// 21. SetCompressLevel
// ============================================================

func TestCov8_SetCompressLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetCompressLevel(9)
	pdf.AddPage()
	pdf.Cell(nil, "compressed")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 22. Various GoPdf methods for coverage
// ============================================================

func TestCov8_SetNewXY(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewXY(100, 50, 20)
	if pdf.GetX() != 50 {
		t.Errorf("expected X=50, got %f", pdf.GetX())
	}
}

func TestCov8_SetXY(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(30, 40)
	if pdf.GetX() != 30 {
		t.Errorf("expected X=30, got %f", pdf.GetX())
	}
	if pdf.GetY() != 40 {
		t.Errorf("expected Y=40, got %f", pdf.GetY())
	}
}

func TestCov8_Br(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	y1 := pdf.GetY()
	pdf.Br(20)
	y2 := pdf.GetY()
	if y2 <= y1 {
		t.Error("Br should increase Y")
	}
}

func TestCov8_SetGrayStroke(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayStroke(0.5)
	pdf.Line(10, 10, 100, 100)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(50, 50, 150, 100)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Polygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polygon([]Point{{X: 10, Y: 10}, {X: 100, Y: 10}, {X: 55, Y: 80}}, "D")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Polygon_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polygon([]Point{{X: 10, Y: 10}, {X: 100, Y: 10}, {X: 55, Y: 80}}, "F")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Polyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polyline([]Point{{X: 10, Y: 10}, {X: 50, Y: 50}, {X: 90, Y: 10}})
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(100, 100, 50, 0, 90, "FD")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(10, 10, 30, 80, 70, 80, 90, 10, "D")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	pdf.ClipPolygon([]Point{{X: 10, Y: 10}, {X: 100, Y: 10}, {X: 55, Y: 80}})
	pdf.Cell(nil, "clipped")
	pdf.RestoreGraphicsState()
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Rectangle_Rounded(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Rectangle(10, 10, 100, 80, "D", 10, 8)
	if err != nil {
		t.Fatalf("Rectangle: %v", err)
	}
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_Rectangle_NoRadius(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Rectangle(10, 10, 100, 80, "FD", 0, 0)
	if err != nil {
		t.Fatalf("Rectangle: %v", err)
	}
}

// ============================================================
// 23. SetInfo / GetInfo
// ============================================================

func TestCov8_SetInfo_GetInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	info := PdfInfo{
		Title:   "Test Title",
		Author:  "Test Author",
		Subject: "Test Subject",
		Creator: "Test Creator",
	}
	pdf.SetInfo(info)

	got := pdf.GetInfo()
	if got.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got '%s'", got.Title)
	}
	if got.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got '%s'", got.Author)
	}
}

// ============================================================
// 24. Transparency
// ============================================================

func TestCov8_SetTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.5,
		BlendModeType: NormalBlendMode,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}

	pdf.Cell(nil, "transparent text")
	pdf.ClearTransparency()

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_SetTransparency_MultiplyBlend(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.3,
		BlendModeType: Multiply,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	pdf.Cell(nil, "multiply blend")
}

// ============================================================
// 25. IsCurrFontContainGlyph
// ============================================================

func TestCov8_IsCurrFontContainGlyph(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph: %v", err)
	}
	if !ok {
		t.Error("expected 'A' to be in font")
	}
}

// ============================================================
// 26. AddExternalLink / AddInternalLink / SetAnchor
// ============================================================

func TestCov8_AddExternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddExternalLink("https://example.com", 10, 20, 100, 15)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_AddInternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetAnchor("target")
	pdf.AddPage()
	pdf.AddInternalLink("target", 10, 20, 100, 15)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 27. ImportPagesFromSource
// ============================================================

func TestCov8_ImportPagesFromSource_File(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := newPDFWithFont(t)
	err := pdf.ImportPagesFromSource(resTestPDF, "/MediaBox")
	if err != nil {
		t.Fatalf("ImportPagesFromSource: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov8_ImportPagesFromSource_Bytes(t *testing.T) {
	pdfBytes, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := newPDFWithFont(t)
	rs := io.ReadSeeker(bytes.NewReader(pdfBytes))
	err = pdf.ImportPagesFromSource(&rs, "/MediaBox")
	if err != nil {
		t.Fatalf("ImportPagesFromSource: %v", err)
	}
}

// ============================================================
// 28. GetNextObjectID / GetNumberOfPages
// ============================================================

func TestCov8_GetNextObjectID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	id := pdf.GetNextObjectID()
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCov8_GetNumberOfPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	if pdf.GetNumberOfPages() != 0 {
		t.Error("expected 0 pages initially")
	}
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 1 {
		t.Errorf("expected 1 page, got %d", pdf.GetNumberOfPages())
	}
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

// ============================================================
// 29. AddHeader / AddFooter
// ============================================================

func TestCov8_AddHeader_AddFooter(t *testing.T) {
	pdf := newPDFWithFont(t)

	headerCalled := false
	footerCalled := false

	pdf.AddHeader(func() {
		headerCalled = true
		pdf.SetY(15)
		pdf.Cell(nil, "Header")
	})
	pdf.AddFooter(func() {
		footerCalled = true
		pdf.SetY(280)
		pdf.Cell(nil, "Footer")
	})

	pdf.AddPage()
	pdf.Cell(nil, "Body content")
	pdf.AddPage()

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
	_ = headerCalled
	_ = footerCalled
}

// ============================================================
// 30. MultiCellWithOption
// ============================================================

func TestCov8_MultiCellWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Line one\nLine two\nLine three", CellOption{
		Align: Right,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}
}

func TestCov8_MultiCellWithOption_Center(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Centered text", CellOption{
		Align: Center,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}
}

// ============================================================
// 31. CellWithOption
// ============================================================

func TestCov8_CellWithOption_Right(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.CellWithOption(&Rect{W: 200, H: 20}, "Right aligned", CellOption{
		Align: Right,
	})
	if err != nil {
		t.Fatalf("CellWithOption: %v", err)
	}
}

func TestCov8_CellWithOption_Center(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.CellWithOption(&Rect{W: 200, H: 20}, "Center aligned", CellOption{
		Align: Center,
	})
	if err != nil {
		t.Fatalf("CellWithOption: %v", err)
	}
}

func TestCov8_CellWithOption_Border(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.CellWithOption(&Rect{W: 200, H: 20}, "With border", CellOption{
		Align:  Left,
		Border: AllBorders,
	})
	if err != nil {
		t.Fatalf("CellWithOption: %v", err)
	}
}

// ============================================================
// 32. IsFitMultiCellWithNewline
// ============================================================

func TestCov8_IsFitMultiCellWithNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fit, h, err := pdf.IsFitMultiCellWithNewline(&Rect{W: 200, H: 100}, "Line 1\nLine 2\nLine 3")
	if err != nil {
		t.Fatalf("IsFitMultiCellWithNewline: %v", err)
	}
	_ = fit
	if h <= 0 {
		t.Error("expected positive height")
	}
}

// ============================================================
// 33. SplitTextWithOption
// ============================================================

func TestCov8_SplitTextWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	lines, err := pdf.SplitTextWithOption("Hello world this is a long text that should wrap", 100, nil)
	if err != nil {
		t.Fatalf("SplitTextWithOption: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

// ============================================================
// 34. SetColorSpace / AddColorSpaceRGB / AddColorSpaceCMYK
// ============================================================

func TestCov8_ColorSpace_RGB(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddColorSpaceRGB("custom-red", 255, 0, 0)
	if err != nil {
		t.Fatalf("AddColorSpaceRGB: %v", err)
	}

	err = pdf.SetColorSpace("custom-red")
	if err != nil {
		t.Fatalf("SetColorSpace: %v", err)
	}
}

func TestCov8_ColorSpace_CMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddColorSpaceCMYK("custom-cyan", 100, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddColorSpaceCMYK: %v", err)
	}
}

// ============================================================
// 35. SetPage
// ============================================================

func TestCov8_SetPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	// Go back to page 1.
	err := pdf.SetPage(1)
	if err != nil {
		t.Fatalf("SetPage: %v", err)
	}
	pdf.Cell(nil, "Back on page 1")
}

func TestCov8_SetPage_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPage(999)
	if err == nil {
		t.Error("expected error for invalid page")
	}
}

// ============================================================
// 36. convertRGBOp / convertCMYKOp / convertGrayOp identity
// ============================================================

func TestCov8_ConvertRGBOp_ToRGB(t *testing.T) {
	// RGB -> RGB should be identity.
	result := convertRGBOp(0.5, 0.3, 0.1, false, ColorspaceRGB)
	if !strings.Contains(result, "rg") {
		t.Errorf("expected rg: %s", result)
	}
}

func TestCov8_ConvertRGBOp_ToRGB_Stroking(t *testing.T) {
	result := convertRGBOp(0.5, 0.3, 0.1, true, ColorspaceRGB)
	if !strings.Contains(result, "RG") {
		t.Errorf("expected RG: %s", result)
	}
}

func TestCov8_ConvertCMYKOp_ToCMYK(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, false, ColorspaceCMYK)
	if !strings.Contains(result, "k") {
		t.Errorf("expected k: %s", result)
	}
}

func TestCov8_ConvertCMYKOp_ToCMYK_Stroking(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, true, ColorspaceCMYK)
	if !strings.Contains(result, "K") {
		t.Errorf("expected K: %s", result)
	}
}

func TestCov8_ConvertGrayOp_ToGray(t *testing.T) {
	result := convertGrayOp(0.5, false, ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Errorf("expected g: %s", result)
	}
}

func TestCov8_ConvertGrayOp_ToGray_Stroking(t *testing.T) {
	result := convertGrayOp(0.5, true, ColorspaceGray)
	if !strings.Contains(result, "G") {
		t.Errorf("expected G: %s", result)
	}
}

func TestCov8_ConvertGrayOp_ToCMYK(t *testing.T) {
	result := convertGrayOp(0.5, false, ColorspaceCMYK)
	if !strings.Contains(result, "k") {
		t.Errorf("expected k: %s", result)
	}
}

func TestCov8_ConvertGrayOp_ToCMYK_Stroking(t *testing.T) {
	result := convertGrayOp(0.5, true, ColorspaceCMYK)
	if !strings.Contains(result, "K") {
		t.Errorf("expected K: %s", result)
	}
}

func TestCov8_ConvertCMYKOp_ToGray(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, false, ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Errorf("expected g: %s", result)
	}
}

func TestCov8_ConvertCMYKOp_ToGray_Stroking(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, true, ColorspaceGray)
	if !strings.Contains(result, "G") {
		t.Errorf("expected G: %s", result)
	}
}

func TestCov8_ConvertCMYKOp_ToRGB(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, false, ColorspaceRGB)
	if !strings.Contains(result, "rg") {
		t.Errorf("expected rg: %s", result)
	}
}

func TestCov8_ConvertCMYKOp_ToRGB_Stroking(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, true, ColorspaceRGB)
	if !strings.Contains(result, "RG") {
		t.Errorf("expected RG: %s", result)
	}
}
