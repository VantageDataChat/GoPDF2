package gopdf

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// ============================================================
// coverage_boost17_test.go — TestCov17_ prefix
// Targets: gopdf.go prepare() paths (embedded files, page labels,
// XMP metadata, OCGs, mark info, form fields in prepare),
// ImageFromWithOption, image_obj_parse.go parsePng tRNS paths,
// open_pdf.go openPDFFromData deeper paths
// ============================================================

// ============================================================
// gopdf.go — prepare() paths via WriteTo with various features
// ============================================================

func TestCov17_Prepare_WithEmbeddedFiles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Embedded file test")

	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "readme.txt",
		Content:  []byte("This is a readme file"),
		MimeType: "text/plain",
	})

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with embedded files: %v", err)
	}
}

func TestCov17_Prepare_WithPageLabels(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")

	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Prefix: "Cover-"},
		{PageIndex: 1, Prefix: "Ch-", Start: 1},
	})

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with page labels: %v", err)
	}
}

func TestCov17_Prepare_WithXMPMetadata(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("XMP test")

	pdf.SetXMPMetadata(XMPMetadata{
		Title:       "Test Document",
		Creator:     []string{"Test Author"},
		Subject:     []string{"test", "pdf", "gopdf"},
		Keywords:    "test, pdf, gopdf",
		CreatorTool: "GoPDF2",
		Producer:    "GoPDF2",
		Description: "A test document",
	})

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with XMP metadata: %v", err)
	}
}

func TestCov17_Prepare_WithOCGs(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("OCG test")

	// Add optional content groups.
	_ = pdf.AddOCG(OCG{Name: "Layer1", On: true})

	var buf bytes.Buffer
	_, writeErr := pdf.WriteTo(&buf)
	if writeErr != nil {
		t.Fatalf("WriteTo with OCGs: %v", writeErr)
	}
}

func TestCov17_Prepare_WithFormFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Form test")

	_ = pdf.AddTextField("name", 50, 100, 200, 25)
	_ = pdf.AddCheckbox("agree", 50, 150, 15, true)

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with form fields: %v", err)
	}
}

func TestCov17_Prepare_WithOutlines(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 2")
	pdf.AddOutline("Chapter 2")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with outlines: %v", err)
	}
}

func TestCov17_Prepare_WithProtection(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Protected content")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with protection: %v", err)
	}
}

func TestCov17_Prepare_AllFeatures(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("All features")
	pdf.AddOutline("Chapter 1")

	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "data.txt",
		Content: []byte("data"),
	})

	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Prefix: "P-"},
	})

	pdf.SetXMPMetadata(XMPMetadata{
		Title:   "All Features",
		Creator: []string{"Test"},
	})

	_ = pdf.AddTextField("field1", 50, 200, 200, 25)

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo all features: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================
// gopdf.go — ImageFromWithOption
// ============================================================

func TestCov17_ImageFromWithOption_PNG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 5), G: uint8(y * 5), B: 128, A: 255})
		}
	}

	err := pdf.ImageFromWithOption(img, ImageFromOption{
		X:      50,
		Y:      50,
		Rect:   &Rect{W: 100, H: 100},
		Format: "png",
	})
	if err != nil {
		t.Fatalf("ImageFromWithOption PNG: %v", err)
	}
}

func TestCov17_ImageFromWithOption_JPEG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: uint8(y * 5), B: uint8(x * 5), A: 255})
		}
	}

	err := pdf.ImageFromWithOption(img, ImageFromOption{
		X:      50,
		Y:      50,
		Rect:   &Rect{W: 100, H: 100},
		Format: "jpeg",
	})
	if err != nil {
		t.Fatalf("ImageFromWithOption JPEG: %v", err)
	}
}

func TestCov17_ImageFromWithOption_NilImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageFromWithOption(nil, ImageFromOption{
		X: 50, Y: 50, Format: "png",
	})
	if err == nil {
		t.Fatal("expected error for nil image")
	}
}

// ============================================================
// image_obj_parse.go — parsePng with tRNS chunk (indexed color)
// ============================================================

func TestCov17_ParsePng_Indexed(t *testing.T) {
	// Create a paletted PNG to trigger the Indexed color path.
	palette := color.Palette{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
		color.RGBA{R: 255, G: 255, B: 0, A: 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
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
	err = pdf.ImageByHolder(holder, 50, 50, nil)
	if err != nil {
		t.Fatalf("ImageByHolder indexed: %v", err)
	}
}

// ============================================================
// gopdf.go — ImageByHolder with various image types
// ============================================================

func TestCov17_ImageByHolder_JPEG(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Fatalf("ImageHolderByPath: %v", err)
	}
	err = pdf.ImageByHolder(holder, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Fatalf("ImageByHolder JPEG: %v", err)
	}
}

func TestCov17_ImageByHolder_PNG(t *testing.T) {
	if _, err := os.Stat(resPNGPath); os.IsNotExist(err) {
		t.Skip("PNG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resPNGPath)
	if err != nil {
		t.Fatalf("ImageHolderByPath: %v", err)
	}
	err = pdf.ImageByHolder(holder, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Fatalf("ImageByHolder PNG: %v", err)
	}
}

func TestCov17_ImageHolderByReader(t *testing.T) {
	imgData := createTestPNG(t, 20, 20)
	reader := bufio.NewReader(bytes.NewReader(imgData))

	holder, err := ImageHolderByReader(reader)
	if err != nil {
		t.Fatalf("ImageHolderByReader: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err = pdf.ImageByHolder(holder, 50, 50, nil)
	if err != nil {
		t.Fatalf("ImageByHolder from reader: %v", err)
	}
}

// ============================================================
// gopdf.go — SetFillColor, SetStrokeColor, RectFromUpperLeftWithStyle
// ============================================================

func TestCov17_RectStyles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetFillColor(255, 0, 0)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 50, "F")

	pdf.SetStrokeColor(0, 0, 255)
	pdf.RectFromUpperLeftWithStyle(50, 120, 100, 50, "D")

	pdf.SetFillColor(0, 255, 0)
	pdf.SetStrokeColor(0, 0, 0)
	pdf.RectFromUpperLeftWithStyle(50, 190, 100, 50, "FD")
}

// ============================================================
// gopdf.go — SetFillColorCMYK
// ============================================================

func TestCov17_SetFillColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetFillColorCMYK(100, 0, 0, 0) // cyan fill
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 50, "F")

	pdf.SetFillColorCMYK(0, 100, 0, 0) // magenta fill
	pdf.RectFromUpperLeftWithStyle(50, 120, 100, 50, "F")
}

// ============================================================
// gopdf.go — Br, SetX, SetY, GetX, GetY
// ============================================================

func TestCov17_PositionMethods(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetX(100)
	pdf.SetY(200)
	if pdf.GetX() != 100 {
		t.Errorf("GetX() = %f, want 100", pdf.GetX())
	}
	if pdf.GetY() != 200 {
		t.Errorf("GetY() = %f, want 200", pdf.GetY())
	}

	pdf.Br(20)
	if pdf.GetY() != 220 {
		t.Errorf("after Br(20), GetY() = %f, want 220", pdf.GetY())
	}
}

// ============================================================
// gopdf.go — SetCharSpacing
// ============================================================

func TestCov17_SetCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	pdf.SetCharSpacing(2.0)
	_ = pdf.Text("Spaced text")

	pdf.SetCharSpacing(0)
	_ = pdf.Text("Normal text")
}

// ============================================================
// gopdf.go — SetCustomLineType
// ============================================================

func TestCov17_SetCustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(10, 10, 200, 200)
}

// ============================================================
// gopdf.go — AddMarkInfo
// ============================================================

func TestCov17_AddMarkInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Mark info test")

	pdf.SetMarkInfo(MarkInfo{Marked: true})

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with mark info: %v", err)
	}
}
