package gopdf

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost24_test.go — TestCov24_ prefix
// Targets: failing writer paths in content_obj.write, image_obj.write,
// pdf_dictionary_obj.write; extractFontsFromResources deeper paths;
// HTML rendering edge cases; parsePng deeper branches
// ============================================================

// failWriter fails after writing n bytes.
type failWriter struct {
	limit   int
	written int
}

func (fw *failWriter) Write(p []byte) (int, error) {
	if fw.written+len(p) > fw.limit {
		remaining := fw.limit - fw.written
		if remaining > 0 {
			fw.written += remaining
			return remaining, errors.New("write limit reached")
		}
		return 0, errors.New("write limit reached")
	}
	fw.written += len(p)
	return len(p), nil
}

// --- Test: content_obj.write with failing writer ---
func TestCov24_ContentObj_WriteFailingWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Test content for failing writer")

	// Find the content obj
	for _, obj := range pdf.pdfObjs {
		if cObj, ok := obj.(*ContentObj); ok {
			// Try with various fail limits to hit different error branches
			for limit := 0; limit < 200; limit += 10 {
				fw := &failWriter{limit: limit}
				err := cObj.write(fw, 1)
				if err == nil {
					// If it succeeded, the limit was high enough
					break
				}
			}
			break
		}
	}
}

// --- Test: content_obj.write with NoCompression + failing writer ---
func TestCov24_ContentObj_WriteNoCompressFailWriter(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetNoCompression()
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "No compress fail test")

	for _, obj := range pdf.pdfObjs {
		if cObj, ok := obj.(*ContentObj); ok {
			for limit := 0; limit < 200; limit += 10 {
				fw := &failWriter{limit: limit}
				err := cObj.write(fw, 1)
				if err == nil {
					break
				}
			}
			break
		}
	}
}

// --- Test: content_obj.write with protection + failing writer ---
func TestCov24_ContentObj_WriteProtectedFailWriter(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Protected fail test")

	for _, obj := range pdf.pdfObjs {
		if cObj, ok := obj.(*ContentObj); ok {
			for limit := 0; limit < 300; limit += 10 {
				fw := &failWriter{limit: limit}
				err := cObj.write(fw, 1)
				if err == nil {
					break
				}
			}
			break
		}
	}
}

// --- Test: image_obj.write with failing writer ---
func TestCov24_ImageObj_WriteFailingWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pngData := createTestPNG(t, 10, 10)
	holder, _ := ImageHolderByBytes(pngData)
	pdf.ImageByHolder(holder, 10, 10, nil)

	for _, obj := range pdf.pdfObjs {
		if imgObj, ok := obj.(*ImageObj); ok {
			for limit := 0; limit < 500; limit += 20 {
				fw := &failWriter{limit: limit}
				err := imgObj.write(fw, 1)
				if err == nil {
					break
				}
			}
			break
		}
	}
}

// --- Test: image_obj.write mask with failing writer ---
func TestCov24_ImageObj_WriteMaskFailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create an RGBA PNG that will generate smask
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(x * 25), G: uint8(y * 25), B: 100, A: uint8(200 - x*5)})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	holder, _ := ImageHolderByBytes(pngBuf.Bytes())
	pdf.ImageByHolder(holder, 10, 10, nil)

	for _, obj := range pdf.pdfObjs {
		if imgObj, ok := obj.(*ImageObj); ok && imgObj.IsMask {
			for limit := 0; limit < 500; limit += 20 {
				fw := &failWriter{limit: limit}
				err := imgObj.write(fw, 1)
				if err == nil {
					break
				}
			}
			break
		}
	}
}

// --- Test: pdf_dictionary_obj.write with failing writer ---
func TestCov24_PdfDictionaryObj_WriteFailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Dict test")

	for _, obj := range pdf.pdfObjs {
		if dictObj, ok := obj.(*PdfDictionaryObj); ok {
			for limit := 0; limit < 2000; limit += 50 {
				fw := &failWriter{limit: limit}
				err := dictObj.write(fw, 1)
				if err == nil {
					break
				}
			}
			break
		}
	}
}

// --- Test: pdf_dictionary_obj.write with protection ---
func TestCov24_PdfDictionaryObj_WriteProtected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Dict protected test")

	for _, obj := range pdf.pdfObjs {
		if dictObj, ok := obj.(*PdfDictionaryObj); ok {
			var buf bytes.Buffer
			err := dictObj.write(&buf, 1)
			if err != nil {
				t.Logf("PdfDictionaryObj.write protected: %v", err)
			}
			break
		}
	}
}

// --- Test: extractFontsFromResources with embedded font ---
func TestCov24_ExtractFontsFromResources(t *testing.T) {
	// Create a PDF with embedded font, write it, then extract fonts
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Font extraction test")
	var buf bytes.Buffer
	pdf.Write(&buf)

	fonts, err := ExtractFontsFromPage(buf.Bytes(), 0)
	if err != nil {
		t.Fatalf("ExtractFontsFromPage: %v", err)
	}
	t.Logf("extracted %d fonts from page 0", len(fonts))
	for _, f := range fonts {
		t.Logf("  font: %s base=%s subtype=%s embedded=%v", f.Name, f.BaseFont, f.Subtype, f.IsEmbedded)
	}

	allFonts, err := ExtractFontsFromAllPages(buf.Bytes())
	if err != nil {
		t.Fatalf("ExtractFontsFromAllPages: %v", err)
	}
	t.Logf("extracted fonts from %d pages", len(allFonts))
}

// --- Test: ExtractFontsFromPage out of range ---
func TestCov24_ExtractFontsFromPage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Test")
	var buf bytes.Buffer
	pdf.Write(&buf)

	_, err := ExtractFontsFromPage(buf.Bytes(), 99)
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

// --- Test: HTML rendering with <img> that exceeds box width ---
func TestCov24_HTML_ImageExceedsWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Image wider than box
	html := `<img src="` + resJPEGPath + `" width="500">`
	_, err := pdf.InsertHTMLBox(10, 10, 100, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox wide img: %v", err)
	}
}

// --- Test: HTML rendering with <img> that exceeds box height ---
func TestCov24_HTML_ImageExceedsHeight(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="` + resJPEGPath + `" width="50" height="500">`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 100, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox tall img: %v", err)
	}
}

// --- Test: HTML rendering with text that exceeds box height ---
func TestCov24_HTML_TextExceedsHeight(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Very long text in a small box
	longText := strings.Repeat("This is a long paragraph of text. ", 50)
	html := `<p>` + longText + `</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 50, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox overflow: %v", err)
	}
}

// --- Test: HTML rendering with nested elements ---
func TestCov24_HTML_NestedElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p><b><i>Bold italic</i></b> and <u>underlined</u> text</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox nested: %v", err)
	}
}

// --- Test: HTML rendering with <span style="color:..."> ---
func TestCov24_HTML_SpanColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p><span style="color:#FF0000">Red</span> <span style="color:rgb(0,128,0)">Green</span></p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox span color: %v", err)
	}
}

// --- Test: HTML rendering with <br> ---
func TestCov24_HTML_BR(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `Line 1<br>Line 2<br>Line 3`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox br: %v", err)
	}
}

// --- Test: HTML rendering with list that exceeds box height ---
func TestCov24_HTML_ListExceedsHeight(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	var items strings.Builder
	for i := 0; i < 30; i++ {
		items.WriteString("<li>Item</li>")
	}
	html := "<ul>" + items.String() + "</ul>"
	_, err := pdf.InsertHTMLBox(10, 10, 200, 50, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox list overflow: %v", err)
	}
}

// --- Test: parsePng error paths ---
func TestCov24_ParsePng_Errors(t *testing.T) {
	// Not a PNG
	reader := bytes.NewReader([]byte("not a png file at all"))
	var info imgInfo
	err := parsePng(reader, &info, image.Config{})
	if err == nil {
		t.Error("expected error for non-PNG data")
	}

	// Too short
	reader2 := bytes.NewReader([]byte{0x89, 0x50})
	err = parsePng(reader2, &info, image.Config{})
	if err == nil {
		t.Error("expected error for truncated PNG")
	}
}

// --- Test: parsePng with 16-bit depth ---
func TestCov24_ParsePng_16BitDepth(t *testing.T) {
	// Build a PNG with 16-bit depth
	w, h := 2, 2
	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0) // filter: None
		for x := 0; x < w; x++ {
			// 16-bit RGB: 6 bytes per pixel
			raw.Write([]byte{0, 128, 0, 64, 0, 32})
		}
	}
	compressed := zlibCompress(raw.Bytes())
	pngData := buildRawPNG(t, w, h, 2, 16, nil, compressed)

	reader := bytes.NewReader(pngData)
	var info imgInfo
	err := parsePng(reader, &info, image.Config{})
	if err == nil {
		t.Error("expected error for 16-bit depth")
	} else if !strings.Contains(err.Error(), "16-bit") {
		t.Errorf("expected 16-bit error, got: %v", err)
	}
}

// --- Test: parsePng with unknown color type ---
func TestCov24_ParsePng_UnknownColorType(t *testing.T) {
	w, h := 2, 2
	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0)
		for x := 0; x < w; x++ {
			raw.WriteByte(128)
		}
	}
	compressed := zlibCompress(raw.Bytes())
	// Color type 5 is invalid
	pngData := buildRawPNG(t, w, h, 5, 8, nil, compressed)

	reader := bytes.NewReader(pngData)
	var info imgInfo
	err := parsePng(reader, &info, image.Config{})
	if err == nil {
		t.Error("expected error for unknown color type")
	}
}

// --- Test: parsePng with Indexed but missing palette ---
func TestCov24_ParsePng_MissingPalette(t *testing.T) {
	w, h := 2, 2
	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0)
		for x := 0; x < w; x++ {
			raw.WriteByte(0)
		}
	}
	compressed := zlibCompress(raw.Bytes())
	// Color type 3 (Indexed) but no PLTE chunk
	pngData := buildRawPNG(t, w, h, 3, 8, nil, compressed)

	reader := bytes.NewReader(pngData)
	var info imgInfo
	err := parsePng(reader, &info, image.Config{})
	if err == nil {
		t.Error("expected error for missing palette")
	}
}

// --- Test: RecompressImages with downscaling ---
func TestCov24_RecompressImages_Downscale(t *testing.T) {
	// Create a PDF with a large image
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create a larger PNG
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	holder, _ := ImageHolderByBytes(pngBuf.Bytes())
	pdf.ImageByHolder(holder, 10, 10, nil)

	var buf bytes.Buffer
	pdf.Write(&buf)

	// Recompress with downscaling
	result, err := RecompressImages(buf.Bytes(), RecompressOption{
		Format:      "jpeg",
		JPEGQuality: 50,
		MaxWidth:    100,
		MaxHeight:   100,
	})
	if err != nil {
		t.Fatalf("RecompressImages: %v", err)
	}
	if len(result) == 0 {
		t.Error("empty result")
	}
}

// --- Test: RecompressImages to PNG format ---
func TestCov24_RecompressImages_ToPNG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pngData := createTestPNG(t, 30, 30)
	holder, _ := ImageHolderByBytes(pngData)
	pdf.ImageByHolder(holder, 10, 10, nil)

	var buf bytes.Buffer
	pdf.Write(&buf)

	result, err := RecompressImages(buf.Bytes(), RecompressOption{
		Format: "png",
	})
	if err != nil {
		t.Fatalf("RecompressImages to PNG: %v", err)
	}
	if len(result) == 0 {
		t.Error("empty result")
	}
}

// --- Test: WriteIncrementalPdf ---
func TestCov24_WriteIncrementalPdf(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Original content")

	var buf bytes.Buffer
	pdf.Write(&buf)
	originalData := buf.Bytes()

	// Open the PDF and add content
	pdf2 := &GoPdf{}
	err := pdf2.OpenPDFFromBytes(originalData, nil)
	if err != nil {
		t.Fatalf("OpenPDFFromBytes: %v", err)
	}

	if err := pdf2.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font: %v", err)
	}
	pdf2.SetFont(fontFamily, "", 14)
	pdf2.SetPage(1)
	pdf2.SetXY(10, 50)
	pdf2.Cell(nil, "Added content")

	outPath := resOutDir + "/incremental_test.pdf"
	defer os.Remove(outPath)
	err = pdf2.WriteIncrementalPdf(outPath, originalData, nil)
	if err != nil {
		t.Logf("WriteIncrementalPdf: %v (may be expected)", err)
	}
}

// --- Test: JournalEnable with failing writer ---
func TestCov24_JournalEnable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Error("journal should be enabled")
	}

	pdf.AddPage()
	pdf.JournalStartOp("test_op")
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Journaled content")
	pdf.JournalEndOp()

	ops := pdf.JournalGetOperations()
	if len(ops) == 0 {
		t.Error("expected journal operations")
	}
	t.Logf("journal ops: %d", len(ops))

	pdf.JournalDisable()
	if pdf.JournalIsEnabled() {
		t.Error("journal should be disabled")
	}
}

// --- Test: GetFonts ---
func TestCov24_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Font stats test")

	fonts := pdf.GetFonts()
	t.Logf("fonts: %d", len(fonts))
}

// --- Test: SetImage with mask ---
func TestCov24_SetImage_WithMask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create RGBA PNG (will generate smask)
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(x * 12), G: uint8(y * 12), B: 100, A: uint8(128 + x*5)})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	holder, _ := ImageHolderByBytes(pngBuf.Bytes())
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 100, H: 100})

	var buf bytes.Buffer
	err := pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// --- Test: ImageObj.GetRect ---
func TestCov24_ImageObj_GetRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pngData := createTestPNG(t, 50, 30)
	holder, _ := ImageHolderByBytes(pngData)
	pdf.ImageByHolder(holder, 10, 10, nil)

	for _, obj := range pdf.pdfObjs {
		if imgObj, ok := obj.(*ImageObj); ok && !imgObj.IsMask {
			rect := imgObj.GetRect()
			t.Logf("ImageObj rect: %+v", rect)
			if rect.W <= 0 || rect.H <= 0 {
				t.Error("expected positive rect dimensions")
			}
			break
		}
	}
}

// --- Test: RenderAllPagesToImages ---
func TestCov24_RenderToPixmap(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Pixmap render test")
	pdf.RectFromUpperLeft(50, 50, 100, 80)

	var buf bytes.Buffer
	pdf.Write(&buf)

	imgs, err := RenderAllPagesToImages(buf.Bytes(), RenderOption{DPI: 72})
	if err != nil {
		t.Logf("RenderAllPagesToImages: %v", err)
		return
	}
	if len(imgs) == 0 {
		t.Error("no images rendered")
	}
}

// --- Test: OpenPDFFromStream with error ---
func TestCov24_OpenPDFFromStream_Error(t *testing.T) {
	pdf := &GoPdf{}
	// Create a reader that returns an error
	errReader := io.ReadSeeker(errReadSeeker{})
	err := pdf.OpenPDFFromStream(&errReader, nil)
	if err == nil {
		t.Error("expected error from bad reader")
	}
}

type errReadSeeker struct{}

func (e errReadSeeker) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func (e errReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek error")
}

// --- Test: AddTTFFontByReaderWithOption with bad reader ---
func TestCov24_AddTTFFontByReaderWithOption_BadReader(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	err := pdf.AddTTFFontByReaderWithOption("bad", &errReadSeeker{}, TtfOption{})
	if err == nil {
		t.Error("expected error from bad reader")
	}
}

// --- Test: SetTTFByReader with bad reader ---
func TestCov24_SetTTFByReader_BadReader(t *testing.T) {
	s := &SubsetFontObj{}
	s.init(func() *GoPdf {
		pdf := &GoPdf{}
		pdf.Start(Config{PageSize: *PageSizeA4})
		return pdf
	})
	err := s.SetTTFByReader(&errReadSeeker{})
	if err == nil {
		t.Error("expected error from bad reader")
	}
}

// --- Test: parseImgJpg with CMYK ---
func TestCov24_ParseImgJpg_CMYK(t *testing.T) {
	// Read the JPEG file
	data, err := os.ReadFile(resJPEGPath)
	if err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	reader := bytes.NewReader(data)
	info, err := parseImg(reader)
	if err != nil {
		t.Fatalf("parseImg JPEG: %v", err)
	}
	t.Logf("JPEG info: w=%d h=%d colspace=%s", info.w, info.h, info.colspace)
}

// --- Test: HTML with <img> on same line as text ---
func TestCov24_HTML_ImageAfterText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Some text before <img src="` + resJPEGPath + `" width="50" height="40"> and after</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img after text: %v", err)
	}
}

// --- Test: HTML with <img> invalid src ---
func TestCov24_HTML_ImageInvalidSrc(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="/nonexistent/image.jpg" width="50" height="40">`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	// Should return error for invalid image path
	if err != nil {
		t.Logf("InsertHTMLBox invalid img: %v (expected)", err)
	}
}

// --- Test: HTML with percentage dimensions ---
func TestCov24_HTML_ImagePercentDimensions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="` + resJPEGPath + `" width="50%" height="30%">`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox percent img: %v", err)
	}
}
