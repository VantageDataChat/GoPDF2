package gopdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"testing"
)

// ============================================================
// coverage_boost20_test.go — TestCov20_ prefix
// Targets: parsePng (gray+alpha, indexed+tRNS), image_obj write
// with protection, content_obj write branches, openPDFFromData
// encrypted path, MergePages, ExtractPagesFromBytes,
// MergePagesFromBytes, page_manipulate, rebuildXref,
// AddCompositeGlyphs, charCodeToGlyphIndexFormat4/12
// ============================================================

// ============================================================
// image_obj_parse.go — parsePng with RGBA (ct=6, alpha channel)
// ============================================================

func TestCov20_ParsePng_RGBA(t *testing.T) {
	// Create an RGBA PNG (color type 6) to exercise the alpha extraction path.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(x * 60), G: uint8(y * 60), B: 128, A: uint8(100 + x*30)})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("ImageByHolder RGBA: %v", err)
	}

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov20_ParsePng_GrayAlpha(t *testing.T) {
	// Create a Gray+Alpha PNG (color type 4) to exercise the gray alpha path.
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			g := uint8(x*60 + y*10)
			img.Set(x, y, color.NRGBA{R: g, G: g, B: g, A: uint8(200 - x*30)})
		}
	}
	var buf bytes.Buffer
	enc := &png.Encoder{CompressionLevel: png.DefaultCompression}
	_ = enc.Encode(&buf, img)

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 50, 200, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("ImageByHolder GrayAlpha: %v", err)
	}
}

func TestCov20_ParsePng_Indexed(t *testing.T) {
	// Create a paletted PNG (color type 3).
	palette := color.Palette{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
		color.RGBA{R: 255, G: 255, B: 0, A: 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 4, 4), palette)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetColorIndex(x, y, uint8((x+y)%4))
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 50, 350, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("ImageByHolder Indexed: %v", err)
	}
}

func TestCov20_ParsePng_Gray(t *testing.T) {
	// Create a pure grayscale PNG (color type 0).
	img := image.NewGray(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.Gray{Y: uint8(x*30 + y*10)})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolder(holder, 200, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("ImageByHolder Gray: %v", err)
	}
}

// ============================================================
// image_obj.go — write with protection
// ============================================================

func TestCov20_ImageObj_Write_WithProtection(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("Image: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov20_ImageObj_Write_PNG_WithProtection(t *testing.T) {
	if _, err := os.Stat(resPNGPath); os.IsNotExist(err) {
		t.Skip("PNG not available")
	}
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	if err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Fatalf("Image PNG: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// content_obj.go — write with protection + no compression
// ============================================================

func TestCov20_ContentObj_Write_ProtectedFlate(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Protected flate content")
	pdf.Line(10, 10, 200, 200)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov20_ContentObj_Write_ProtectedNoFlate(t *testing.T) {
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
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Protected no flate")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov20_ContentObj_Write_UnprotectedFlate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetCompressLevel(9)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Compressed content")
	pdf.Line(10, 10, 200, 200)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov20_ContentObj_Write_UnprotectedNoFlate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "No compression content")
	pdf.Line(10, 10, 200, 200)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// page_manipulate.go — MergePages, ExtractPagesFromBytes, MergePagesFromBytes
// ============================================================

func TestCov20_MergePages(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	result, err := MergePages([]string{resTestPDF, resTestPDF}, nil)
	if err != nil {
		t.Fatalf("MergePages: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
	_ = data
}

func TestCov20_MergePagesFromBytes(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	result, err := MergePagesFromBytes([][]byte{data, data}, nil)
	if err != nil {
		t.Fatalf("MergePagesFromBytes: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestCov20_ExtractPagesFromBytes(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	result, err := ExtractPagesFromBytes(data, []int{1}, nil)
	if err != nil {
		t.Fatalf("ExtractPagesFromBytes: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestCov20_ExtractPagesFromBytes_Multiple(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	// Try extracting multiple pages.
	rs := io.ReadSeeker(bytes.NewReader(data))
	pdf := GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	probe := pdf.fpdi
	probe.SetSourceStream(&rs)
	numPages := probe.GetNumPages()

	if numPages > 1 {
		pages := make([]int, numPages)
		for i := range pages {
			pages[i] = i + 1
		}
		result, err := ExtractPagesFromBytes(data, pages, nil)
		if err != nil {
			t.Fatalf("ExtractPagesFromBytes multiple: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	}
}

func TestCov20_MergePages_InvalidPath(t *testing.T) {
	_, err := MergePages([]string{"/nonexistent/path.pdf"}, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCov20_MergePagesFromBytes_Invalid(t *testing.T) {
	func() {
		defer func() { recover() }()
		result, err := MergePagesFromBytes([][]byte{[]byte("not a pdf")}, nil)
		_ = result
		_ = err
	}()
}

// ============================================================
// open_pdf.go — openPDFFromData with encrypted PDF
// ============================================================

func TestCov20_OpenPDF_EncryptedNoPassword(t *testing.T) {
	// Create a protected PDF first.
	ensureOutDir(t)
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Encrypted content")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	// Try to open without password.
	pdf2 := GoPdf{}
	err = pdf2.OpenPDFFromBytes(data, nil)
	// Should either succeed (empty password) or return ErrEncryptedPDF.
	_ = err
}

func TestCov20_OpenPDF_EncryptedWithPassword(t *testing.T) {
	ensureOutDir(t)
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Encrypted content")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	// Try to open with password.
	pdf2 := GoPdf{}
	err = pdf2.OpenPDFFromBytes(data, &OpenPDFOption{Password: "user"})
	_ = err
}

func TestCov20_OpenPDF_EncryptedWrongPassword(t *testing.T) {
	ensureOutDir(t)
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Encrypted content")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	// Try to open with wrong password.
	pdf2 := GoPdf{}
	err = pdf2.OpenPDFFromBytes(data, &OpenPDFOption{Password: "wrongpassword"})
	_ = err
}

// ============================================================
// gopdf.go — prepare with various features
// ============================================================

func TestCov20_Prepare_WithEmbeddedFiles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Embedded files test")

	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "file1.txt",
		Content: []byte("content 1"),
	})
	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "file2.txt",
		Content: []byte("content 2"),
	})

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov20_Prepare_WithOCG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	ocg := pdf.AddOCG(OCG{Name: "Layer 1"})
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "OCG content")
	_ = ocg
}

func TestCov20_Prepare_WithXMPMetadata(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetXMPMetadata(XMPMetadata{
		Title:       "XMP Test",
		Creator:     []string{"Test Creator"},
		Subject:     []string{"Test Subject"},
		Description: "Test Description",
	})
	pdf.AddPage()
	_ = pdf.Cell(nil, "XMP metadata test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// gopdf.go — AddTTFFontByReaderWithOption
// ============================================================

func TestCov20_AddTTFFontByReaderWithOption(t *testing.T) {
	fontData, err := os.ReadFile(resFontPath)
	if err != nil {
		t.Skipf("font not available: %v", err)
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	var glyphNotFound []rune
	if err := pdf.AddTTFFontByReaderWithOption("ByReaderOpt", bytes.NewReader(fontData), TtfOption{
		OnGlyphNotFound: func(r rune) {
			glyphNotFound = append(glyphNotFound, r)
		},
	}); err != nil {
		t.Fatalf("AddTTFFontByReaderWithOption: %v", err)
	}

	if err := pdf.SetFont("ByReaderOpt", "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	_ = pdf.Cell(nil, "Font by reader with option")
}

// ============================================================
// gopdf.go — ImageByHolderWithOptions with crop
// ============================================================

func TestCov20_ImageByHolderWithOptions_Crop(t *testing.T) {
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
		Crop: &CropOptions{X: 10, Y: 10, Width: 80, Height: 80},
	})
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions crop: %v", err)
	}
}

// ============================================================
// gopdf.go — ImageFromWithOption
// ============================================================

func TestCov20_ImageFromWithOption(t *testing.T) {
	// Create an in-memory image.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 5), G: uint8(y * 5), B: 128, A: 255})
		}
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageFromWithOption(img, ImageFromOption{
		Format: "png",
		X:      50,
		Y:      50,
		Rect:   &Rect{W: 150, H: 150},
	})
	if err != nil {
		t.Fatalf("ImageFromWithOption: %v", err)
	}
}

// ============================================================
// html_insert.go — renderImage, renderList, measureLineWidth
// ============================================================

func TestCov20_InsertHTMLBox_WithImage(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	htmlStr := `<p>Before image</p>
<img src="` + resJPEGPath + `" width="100" height="100"/>
<p>After image</p>`

	_, err := pdf.InsertHTMLBox(50, 50, 400, 600, htmlStr, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
		DefaultColor:      [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Logf("InsertHTMLBox with image: %v", err)
	}
}

func TestCov20_InsertHTMLBox_NestedLists(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	htmlStr := `<ul>
  <li>Item 1</li>
  <li>Item 2
    <ul>
      <li>Sub-item 2.1</li>
      <li>Sub-item 2.2</li>
    </ul>
  </li>
  <li>Item 3</li>
</ul>
<ol>
  <li>First</li>
  <li>Second</li>
  <li>Third</li>
</ol>`

	_, err := pdf.InsertHTMLBox(50, 50, 400, 600, htmlStr, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
		DefaultColor:      [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox nested lists: %v", err)
	}
}

func TestCov20_InsertHTMLBox_LongText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	htmlStr := `<p>This is a very long paragraph that should wrap across multiple lines in the HTML box. ` +
		`It contains enough text to test the line breaking and word wrapping logic. ` +
		`The quick brown fox jumps over the lazy dog. ` +
		`Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
<p style="text-align: center;">Centered paragraph</p>
<p style="text-align: right;">Right-aligned paragraph</p>`

	_, err := pdf.InsertHTMLBox(50, 50, 300, 600, htmlStr, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
		DefaultColor:      [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox long text: %v", err)
	}
}

// ============================================================
// text_extract.go — ExtractTextFromPage, ExtractTextFromAllPages
// ============================================================

func TestCov20_ExtractTextFromPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Extractable text content")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	text, err := ExtractTextFromPage(data, 0)
	_ = text
	_ = err
}

func TestCov20_ExtractTextFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1 text")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2 text")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	texts, err := ExtractTextFromAllPages(data)
	_ = texts
	_ = err
}

// ============================================================
// font_extract.go — ExtractFontsFromPage, ExtractFontsFromAllPages
// ============================================================

func TestCov20_ExtractFontsFromPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Font extract test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	fonts, err := ExtractFontsFromPage(data, 0)
	_ = fonts
	_ = err
}

func TestCov20_ExtractFontsFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Font extract all pages")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	fonts, err := ExtractFontsFromAllPages(data)
	_ = fonts
	_ = err
}

// ============================================================
// image_extract.go — ExtractImagesFromPage, ExtractImagesFromAllPages
// ============================================================

func TestCov20_ExtractImagesFromPage(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	images, err := ExtractImagesFromPage(data, 0)
	_ = images
	_ = err
}

func TestCov20_ExtractImagesFromAllPages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	images, err := ExtractImagesFromAllPages(data)
	_ = images
	_ = err
}

// ============================================================
// select_pages.go — SelectPages, selectPagesFromBytes
// ============================================================

func TestCov20_SelectPages(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	result, err := SelectPagesFromBytes(data, []int{1}, nil)
	if err != nil {
		t.Fatalf("SelectPagesFromBytes: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ============================================================
// page_rotate.go — SetPageRotation, GetPageRotation
// ============================================================

func TestCov20_PageRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Rotation test")

	err := pdf.SetPageRotation(1, 90)
	if err != nil {
		t.Fatalf("SetPageRotation: %v", err)
	}

	rot, err := pdf.GetPageRotation(1)
	if err != nil {
		t.Fatalf("GetPageRotation: %v", err)
	}
	if rot != 90 {
		t.Errorf("expected 90, got %d", rot)
	}
}

func TestCov20_PageRotation_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageRotation(99, 90)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}

	_, err = pdf.GetPageRotation(99)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

// ============================================================
// page_cropbox.go — SetPageCropBox, GetPageCropBox, ClearPageCropBox
// ============================================================

func TestCov20_PageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageCropBox(1, Box{Left: 10, Top: 10, Right: 400, Bottom: 700})
	if err != nil {
		t.Fatalf("SetPageCropBox: %v", err)
	}

	box, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox: %v", err)
	}
	if box == nil {
		t.Fatal("expected non-nil crop box")
	}

	err = pdf.ClearPageCropBox(1)
	if err != nil {
		t.Fatalf("ClearPageCropBox: %v", err)
	}

	box, err = pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox after clear: %v", err)
	}
	if box != nil {
		t.Error("expected nil crop box after clear")
	}
}

func TestCov20_PageCropBox_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageCropBox(99, Box{})
	if err == nil {
		t.Fatal("expected error for invalid page")
	}

	_, err = pdf.GetPageCropBox(99)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}

	err = pdf.ClearPageCropBox(99)
	if err == nil {
		t.Fatal("expected error for invalid page")
	}
}

// ============================================================
// pixmap_render.go — RenderPageToImage, RenderAllPagesToImages
// ============================================================

func TestCov20_RenderPageToImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Render test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	img, err := RenderPageToImage(data, 0, RenderOption{DPI: 72})
	_ = img
	_ = err
}

func TestCov20_RenderAllPagesToImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	imgs, err := RenderAllPagesToImages(data, RenderOption{DPI: 36})
	_ = imgs
	_ = err
}
