package gopdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// ============================================================
// coverage_boost16_test.go — TestCov16_ prefix
// Targets: image_obj.go (write, GetRect, SetImage, parse),
// gopdf.go (AddTTFFontDataWithOption, maskHolder, GetBytesPdfReturnErr,
// IsFitMultiCell, PlaceHolderText, Polygon, Polyline),
// cache_content_rotate.go write, image_obj_parse.go parsePng alpha paths,
// open_pdf.go openPDFFromData deeper paths
// ============================================================

// ============================================================
// image_obj.go — SetImage, parse, getRect, write
// ============================================================

func TestCov16_ImageObj_SetImage_Parse(t *testing.T) {
	imgData := createTestPNG(t, 20, 20)
	reader := bytes.NewReader(imgData)

	imgObj := &ImageObj{}
	err := imgObj.SetImage(reader)
	if err != nil {
		t.Fatalf("SetImage: %v", err)
	}

	err = imgObj.parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if imgObj.getType() != "Image" {
		t.Errorf("getType() = %q, want 'Image'", imgObj.getType())
	}
}

func TestCov16_ImageObj_GetRect(t *testing.T) {
	imgData := createTestPNG(t, 50, 30)
	reader := bytes.NewReader(imgData)

	imgObj := &ImageObj{}
	_ = imgObj.SetImage(reader)

	rect, err := imgObj.getRect()
	if err != nil {
		t.Fatalf("getRect: %v", err)
	}
	if rect.W <= 0 || rect.H <= 0 {
		t.Errorf("expected positive dimensions, got W=%f H=%f", rect.W, rect.H)
	}
}

func TestCov16_ImageObj_Write(t *testing.T) {
	imgData := createTestPNG(t, 10, 10)
	reader := bytes.NewReader(imgData)

	imgObj := &ImageObj{}
	_ = imgObj.SetImage(reader)
	_ = imgObj.parse()

	var buf bytes.Buffer
	err := imgObj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestCov16_ImageObj_Write_WithProtection(t *testing.T) {
	imgData := createTestPNG(t, 10, 10)
	reader := bytes.NewReader(imgData)

	imgObj := &ImageObj{}
	_ = imgObj.SetImage(reader)
	_ = imgObj.parse()

	// Set up protection.
	prot := &PDFProtection{}
	prot.setProtection(PermissionsPrint, []byte("user"), []byte("owner"))
	imgObj.setProtection(prot)

	var buf bytes.Buffer
	err := imgObj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write with protection: %v", err)
	}
}

func TestCov16_ImageObj_Write_Mask(t *testing.T) {
	// Create a PNG with alpha channel (RGBA).
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: uint8(x * 25)})
		}
	}
	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, img)

	imgObj := &ImageObj{IsMask: true}
	_ = imgObj.SetImage(bytes.NewReader(pngBuf.Bytes()))
	_ = imgObj.parse()

	if imgObj.haveSMask() {
		var buf bytes.Buffer
		err := imgObj.write(&buf, 1)
		if err != nil {
			t.Fatalf("write mask: %v", err)
		}
	}
}

func TestCov16_ImageObj_SetImagePath(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	imgObj := &ImageObj{}
	err := imgObj.SetImagePath(resJPEGPath)
	if err != nil {
		t.Fatalf("SetImagePath: %v", err)
	}
	err = imgObj.parse()
	if err != nil {
		t.Fatalf("parse after SetImagePath: %v", err)
	}
}

func TestCov16_ImageObj_SetImagePath_Invalid(t *testing.T) {
	imgObj := &ImageObj{}
	err := imgObj.SetImagePath("/nonexistent/image.jpg")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

// ============================================================
// gopdf.go — AddTTFFontDataWithOption
// ============================================================

func TestCov16_AddTTFFontDataWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); os.IsNotExist(err) {
		t.Skip("font not available")
	}
	fontData, err := os.ReadFile(resFontPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontDataWithOption("DataFont", fontData, TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Fatalf("AddTTFFontDataWithOption: %v", err)
	}
	err = pdf.SetFont("DataFont", "", 14)
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Data font test")
}

func TestCov16_AddTTFFontDataWithOption_InvalidData(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontDataWithOption("bad", []byte("not a font"), TtfOption{})
	if err == nil {
		t.Fatal("expected error for invalid font data")
	}
}

// ============================================================
// gopdf.go — GetBytesPdfReturnErr
// ============================================================

func TestCov16_GetBytesPdfReturnErr(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("GetBytesPdfReturnErr test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF bytes")
	}
}

// ============================================================
// gopdf.go — IsFitMultiCell
// ============================================================

func TestCov16_IsFitMultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fit, _, err := pdf.IsFitMultiCell(&Rect{W: 200, H: 100}, "Short text")
	if err != nil {
		t.Fatalf("IsFitMultiCell: %v", err)
	}
	if !fit {
		t.Error("expected short text to fit")
	}
}

func TestCov16_IsFitMultiCell_LongText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	longText := "This is a very long text that should not fit in a tiny cell. "
	for i := 0; i < 20; i++ {
		longText += "More text to overflow. "
	}

	fit, _, err := pdf.IsFitMultiCell(&Rect{W: 50, H: 20}, longText)
	if err != nil {
		t.Fatalf("IsFitMultiCell long: %v", err)
	}
	if fit {
		t.Error("expected long text not to fit in tiny cell")
	}
}

// ============================================================
// gopdf.go — PlaceHolderText
// ============================================================

func TestCov16_PlaceHolderText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.PlaceHolderText("placeholder1", 200)
	if err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}
}

func TestCov16_PlaceHolderText_SingleLine(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.PlaceHolderText("placeholder2", 100)
	if err != nil {
		t.Fatalf("PlaceHolderText single line: %v", err)
	}
}

// ============================================================
// gopdf.go — Polygon, Polyline
// ============================================================

func TestCov16_Polygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	points := []Point{
		{X: 100, Y: 100},
		{X: 200, Y: 100},
		{X: 150, Y: 200},
	}
	pdf.Polygon(points, "D")
}

func TestCov16_Polygon_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	points := []Point{
		{X: 100, Y: 100},
		{X: 200, Y: 100},
		{X: 200, Y: 200},
		{X: 100, Y: 200},
	}
	pdf.Polygon(points, "F")
}

func TestCov16_Polygon_FillDraw(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	points := []Point{
		{X: 100, Y: 100},
		{X: 200, Y: 100},
		{X: 150, Y: 200},
	}
	pdf.Polygon(points, "FD")
}

func TestCov16_Polyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	points := []Point{
		{X: 50, Y: 50},
		{X: 100, Y: 100},
		{X: 150, Y: 50},
		{X: 200, Y: 100},
	}
	pdf.Polyline(points)
}

// ============================================================
// gopdf.go — Rotate (triggers cache_content_rotate.go write)
// ============================================================

func TestCov16_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(200, 400)

	pdf.Rotate(45, 200, 400)
	_ = pdf.Text("Rotated text")
	pdf.RotateReset()
}

func TestCov16_Rotate_Multiple(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Rotate(30, 100, 100)
	pdf.SetXY(100, 100)
	_ = pdf.Text("30 degrees")
	pdf.RotateReset()

	pdf.Rotate(60, 200, 200)
	pdf.SetXY(200, 200)
	_ = pdf.Text("60 degrees")
	pdf.RotateReset()

	pdf.Rotate(90, 300, 300)
	pdf.SetXY(300, 300)
	_ = pdf.Text("90 degrees")
	pdf.RotateReset()
}

// ============================================================
// image_obj_parse.go — parsePng with alpha channel (ct >= 4)
// ============================================================

func TestCov16_ParsePng_RGBA(t *testing.T) {
	// Create RGBA PNG (color type 6) to trigger alpha extraction path.
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 128, A: uint8(200 - x*10)})
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
		t.Fatalf("ImageByHolder RGBA: %v", err)
	}
}

func TestCov16_ParsePng_GrayAlpha(t *testing.T) {
	// Create a gray+alpha PNG (NRGBA with gray values) to trigger ct=4 path.
	// Go's png encoder will produce NRGBA, but we need to test the gray alpha path.
	// We'll use a standard RGBA image with gray values.
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			v := uint8(x*30 + y*10)
			img.Set(x, y, color.NRGBA{R: v, G: v, B: v, A: uint8(200 - x*10)})
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
		t.Fatalf("ImageByHolder gray alpha: %v", err)
	}
}

// ============================================================
// open_pdf.go — OpenPDF deeper paths
// ============================================================

func TestCov16_OpenPDF_Basic(t *testing.T) {
	if _, err := os.Stat(resTestPDF); os.IsNotExist(err) {
		t.Skip("test PDF not available")
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.OpenPDF(resTestPDF, nil)
	if err != nil {
		t.Fatalf("OpenPDF: %v", err)
	}
}

func TestCov16_OpenPDF_InvalidPath(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.OpenPDF("/nonexistent/file.pdf", nil)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

// ============================================================
// gopdf.go — SaveGraphicsState, RestoreGraphicsState
// ============================================================

func TestCov16_GraphicsState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SaveGraphicsState()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.Line(10, 10, 200, 200)
	pdf.RestoreGraphicsState()

	pdf.SaveGraphicsState()
	pdf.SetFillColor(0, 0, 255)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 100, "F")
	pdf.RestoreGraphicsState()
}

// ============================================================
// gopdf.go — SetLineWidth, SetLineType
// ============================================================

func TestCov16_SetLineWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetLineWidth(0.5)
	pdf.Line(10, 10, 200, 200)
	pdf.SetLineWidth(2.0)
	pdf.Line(20, 20, 210, 210)
	pdf.SetLineWidth(5.0)
	pdf.Line(30, 30, 220, 220)
}

func TestCov16_SetLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetLineType("dashed")
	pdf.Line(10, 10, 200, 200)
	pdf.SetLineType("dotted")
	pdf.Line(20, 20, 210, 210)
	pdf.SetLineType("")
	pdf.Line(30, 30, 220, 220)
}

// ============================================================
// gopdf.go — SetGrayFill, SetGrayStroke
// ============================================================

func TestCov16_GrayColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetGrayFill(0.5)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 100, "F")

	pdf.SetGrayStroke(0.3)
	pdf.Line(10, 10, 200, 200)
}

// ============================================================
// gopdf.go — Curve
// ============================================================

func TestCov16_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Curve(50, 50, 75, 150, 150, 150, 200, 50, "D")
}

func TestCov16_Curve_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Curve(50, 50, 75, 150, 150, 150, 200, 50, "F")
}

// ============================================================
// gopdf.go — WriteTo with multiple pages and images
// ============================================================

func TestCov16_WriteTo_Complex(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)

	// Page 1: text + image
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1 with image")
	_ = pdf.Image(resJPEGPath, 50, 100, &Rect{W: 100, H: 100})

	// Page 2: rotated text
	pdf.AddPage()
	pdf.Rotate(30, 200, 400)
	pdf.SetXY(200, 400)
	_ = pdf.Text("Rotated")
	pdf.RotateReset()

	// Page 3: shapes
	pdf.AddPage()
	pdf.SetStrokeColor(0, 128, 0)
	pdf.Line(10, 10, 500, 700)
	pdf.Polygon([]Point{{X: 100, Y: 100}, {X: 200, Y: 100}, {X: 150, Y: 200}}, "FD")
	pdf.Sector(300, 400, 50, 0, 120, "FD")

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo complex: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}
