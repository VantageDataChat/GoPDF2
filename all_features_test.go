package gopdf

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

// ============================================================
// Test resources
// ============================================================

const (
	resFontPath     = "./test/res/LiberationSerif-Regular.ttf"
	resFontPath2    = "./examples/outline_example/Ubuntu-L.ttf"
	resJPEGPath     = "./test/res/gopher01.jpg"
	resPNGPath      = "./test/res/PNG_transparency_demonstration_1.png"
	resPNGPath2     = "./test/res/gopher02.png"
	resTestPDF      = "./examples/outline_example/outline_demo.pdf"
	resOutDir       = "./test/out"
	fontFamily      = "LiberationSerif"
	fontFamily2     = "Ubuntu-L"
)

// helper: ensure output directory exists
func ensureOutDir(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(resOutDir, 0777); err != nil {
		t.Fatalf("cannot create output dir: %v", err)
	}
}

// helper: create a basic A4 PDF with font loaded
func newPDFWithFont(t *testing.T) *GoPdf {
	t.Helper()
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	return pdf
}

// ============================================================
// 1. Document Lifecycle
// ============================================================

func TestAllFeatures_DocumentLifecycle(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	if err := pdf.Cell(nil, "Document lifecycle test"); err != nil {
		t.Fatalf("Cell: %v", err)
	}

	// WritePdf
	if err := pdf.WritePdf(resOutDir + "/all_lifecycle.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}

	// GetBytesPdfReturnErr
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.SetXY(50, 50)
	pdf2.Cell(nil, "bytes test")
	out, err := pdf2.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("output does not start with %PDF-")
	}

	// Write to buffer
	pdf3 := newPDFWithFont(t)
	pdf3.AddPage()
	pdf3.Cell(nil, "write test")
	var buf bytes.Buffer
	if err := pdf3.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("Write produced empty output")
	}

	// WriteTo
	pdf4 := newPDFWithFont(t)
	pdf4.AddPage()
	pdf4.Cell(nil, "writeto test")
	var buf2 bytes.Buffer
	n, err := pdf4.WriteTo(&buf2)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n == 0 {
		t.Fatal("WriteTo wrote 0 bytes")
	}

	// Read
	pdf5 := newPDFWithFont(t)
	pdf5.AddPage()
	pdf5.Cell(nil, "read test")
	p := make([]byte, 5)
	_, err = pdf5.Read(p)
	if err != nil && err != io.EOF {
		t.Fatalf("Read: %v", err)
	}

	// Close
	pdf6 := newPDFWithFont(t)
	pdf6.AddPage()
	if err := pdf6.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// ============================================================
// 2. Page Management
// ============================================================

func TestAllFeatures_PageManagement(t *testing.T) {
	pdf := newPDFWithFont(t)

	if pdf.GetNumberOfPages() != 0 {
		t.Fatal("expected 0 pages initially")
	}

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "page 1")
	if pdf.GetNumberOfPages() != 1 {
		t.Fatal("expected 1 page after AddPage")
	}

	// AddPageWithOption
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 400, H: 600},
	})
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "page 2")
	if pdf.GetNumberOfPages() != 2 {
		t.Fatal("expected 2 pages")
	}

	// SetPage
	if err := pdf.SetPage(1); err != nil {
		t.Fatalf("SetPage(1): %v", err)
	}
	if err := pdf.SetPage(2); err != nil {
		t.Fatalf("SetPage(2): %v", err)
	}

	// SetPage invalid
	if err := pdf.SetPage(99); err == nil {
		t.Fatal("expected error for invalid page number")
	}
}

// ============================================================
// 3. Font Management
// ============================================================

func TestAllFeatures_FontManagement(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// AddTTFFont
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}

	// AddTTFFontWithOption
	var glyphNotFound []rune
	if err := pdf.AddTTFFontWithOption("WithOption", resFontPath, TtfOption{
		OnGlyphNotFound: func(r rune) {
			glyphNotFound = append(glyphNotFound, r)
		},
	}); err != nil {
		t.Fatalf("AddTTFFontWithOption: %v", err)
	}

	// AddTTFFontByReader
	fontData, err := os.ReadFile(resFontPath)
	if err != nil {
		t.Skipf("cannot read font: %v", err)
	}
	if err := pdf.AddTTFFontByReader("ByReader", bytes.NewReader(fontData)); err != nil {
		t.Fatalf("AddTTFFontByReader: %v", err)
	}

	// AddTTFFontData
	if err := pdf.AddTTFFontData("ByData", fontData); err != nil {
		t.Fatalf("AddTTFFontData: %v", err)
	}

	// SetFont
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}

	// SetFontWithStyle
	if err := pdf.SetFontWithStyle(fontFamily, Regular, 16); err != nil {
		t.Fatalf("SetFontWithStyle: %v", err)
	}

	// SetFontSize
	if err := pdf.SetFontSize(18); err != nil {
		t.Fatalf("SetFontSize: %v", err)
	}

	// SetCharSpacing
	if err := pdf.SetCharSpacing(1.5); err != nil {
		t.Fatalf("SetCharSpacing: %v", err)
	}

	// IsCurrFontContainGlyph
	pdf.AddPage()
	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph: %v", err)
	}
	if !ok {
		t.Fatal("expected 'A' to be in font")
	}

	// KernOverride
	if err := pdf.KernOverride(fontFamily, func(leftRune, rightRune rune, leftPair, rightPair uint, pairVal int16) int16 {
		return pairVal
	}); err != nil {
		t.Fatalf("KernOverride: %v", err)
	}
}

// ============================================================
// 4. FontContainer
// ============================================================

func TestAllFeatures_FontContainer(t *testing.T) {
	fc := &FontContainer{}

	// AddTTFFont
	if err := fc.AddTTFFont("fc-font", resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}

	fontData, err := os.ReadFile(resFontPath)
	if err != nil {
		t.Skipf("cannot read font: %v", err)
	}

	// AddTTFFontByReader
	if err := fc.AddTTFFontByReader("fc-reader", bytes.NewReader(fontData)); err != nil {
		t.Fatalf("FontContainer.AddTTFFontByReader: %v", err)
	}

	// AddTTFFontData
	if err := fc.AddTTFFontData("fc-data", fontData); err != nil {
		t.Fatalf("FontContainer.AddTTFFontData: %v", err)
	}

	// AddTTFFontFromFontContainer
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	if err := pdf.AddTTFFontFromFontContainer("fc-font", fc); err != nil {
		t.Fatalf("AddTTFFontFromFontContainer: %v", err)
	}

	// ErrFontNotFound
	if err := pdf.AddTTFFontFromFontContainer("nonexistent", fc); err != ErrFontNotFound {
		t.Fatalf("expected ErrFontNotFound, got %v", err)
	}
}

// ============================================================
// 5. Text Rendering
// ============================================================

func TestAllFeatures_TextRendering(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Text
	pdf.SetXY(50, 50)
	if err := pdf.Text("Hello Text()"); err != nil {
		t.Fatalf("Text: %v", err)
	}

	// Cell
	pdf.SetXY(50, 80)
	if err := pdf.Cell(nil, "Hello Cell()"); err != nil {
		t.Fatalf("Cell: %v", err)
	}

	// Cell with Rect
	pdf.SetXY(50, 110)
	if err := pdf.Cell(&Rect{W: 200, H: 20}, "Cell with Rect"); err != nil {
		t.Fatalf("Cell with Rect: %v", err)
	}

	// CellWithOption
	pdf.SetXY(50, 140)
	if err := pdf.CellWithOption(&Rect{W: 200, H: 20}, "CellWithOption centered", CellOption{
		Align: Center | Middle,
	}); err != nil {
		t.Fatalf("CellWithOption: %v", err)
	}

	// MultiCell
	pdf.SetXY(50, 170)
	if err := pdf.MultiCell(&Rect{W: 150, H: 60}, "This is a long text that should wrap across multiple lines in the MultiCell."); err != nil {
		t.Fatalf("MultiCell: %v", err)
	}

	// MultiCellWithOption
	pdf.SetXY(50, 250)
	if err := pdf.MultiCellWithOption(&Rect{W: 150, H: 60}, "MultiCellWithOption right aligned text here.", CellOption{
		Align: Right | Top,
	}); err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}

	if err := pdf.WritePdf(resOutDir + "/all_text.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 6. Text Measurement & Splitting
// ============================================================

func TestAllFeatures_TextMeasurement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// MeasureTextWidth
	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth: %v", err)
	}
	if w <= 0 {
		t.Fatal("expected positive text width")
	}

	// MeasureCellHeightByText
	h, err := pdf.MeasureCellHeightByText("Hello")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Fatal("expected positive cell height")
	}

	// SplitText
	lines, err := pdf.SplitText("Lorem ipsum dolor sit amet consetetur", 100)
	if err != nil {
		t.Fatalf("SplitText: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("SplitText returned no lines")
	}

	// SplitTextWithWordWrap
	lines2, err := pdf.SplitTextWithWordWrap("Lorem ipsum dolor sit amet consetetur", 100)
	if err != nil {
		t.Fatalf("SplitTextWithWordWrap: %v", err)
	}
	if len(lines2) == 0 {
		t.Fatal("SplitTextWithWordWrap returned no lines")
	}

	// SplitTextWithOption
	lines3, err := pdf.SplitTextWithOption("Lorem ipsum dolor sit amet", 100, &BreakOption{
		Mode:           BreakModeIndicatorSensitive,
		BreakIndicator: ' ',
	})
	if err != nil {
		t.Fatalf("SplitTextWithOption: %v", err)
	}
	if len(lines3) == 0 {
		t.Fatal("SplitTextWithOption returned no lines")
	}
}

// ============================================================
// 7. Position & Navigation
// ============================================================

func TestAllFeatures_Position(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// SetX / GetX
	pdf.SetX(100)
	if pdf.GetX() != 100 {
		t.Fatalf("expected X=100, got %f", pdf.GetX())
	}

	// SetY / GetY
	pdf.SetY(200)
	if pdf.GetY() != 200 {
		t.Fatalf("expected Y=200, got %f", pdf.GetY())
	}

	// SetXY
	pdf.SetXY(150, 250)
	if pdf.GetX() != 150 || pdf.GetY() != 250 {
		t.Fatalf("SetXY failed: got (%f, %f)", pdf.GetX(), pdf.GetY())
	}

	// Br
	y0 := pdf.GetY()
	pdf.Br(30)
	if pdf.GetY() != y0+30 {
		t.Fatalf("Br(30) failed: expected Y=%f, got %f", y0+30, pdf.GetY())
	}

	// SetNewY (adds a new page if y+h exceeds page height)
	pdf.SetNewY(800, 50)

	// SetNewYIfNoOffset
	pdf.SetNewYIfNoOffset(100, 50)

	// SetNewXY
	pdf.SetNewXY(800, 50, 50)
}

// ============================================================
// 8. Colors
// ============================================================

func TestAllFeatures_Colors(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// SetTextColor
	pdf.SetTextColor(255, 0, 0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Red text")

	// SetTextColorCMYK
	pdf.SetTextColorCMYK(0, 100, 100, 0)
	pdf.SetXY(50, 80)
	pdf.Cell(nil, "CMYK text")

	// SetStrokeColor
	pdf.SetStrokeColor(0, 0, 255)

	// SetFillColor
	pdf.SetFillColor(200, 200, 200)

	// SetStrokeColorCMYK
	pdf.SetStrokeColorCMYK(100, 0, 0, 0)

	// SetFillColorCMYK
	pdf.SetFillColorCMYK(0, 0, 100, 0)

	// SetGrayFill
	pdf.SetGrayFill(0.5)

	// SetGrayStroke
	pdf.SetGrayStroke(0.3)

	if err := pdf.WritePdf(resOutDir + "/all_colors.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 9. Drawing Primitives
// ============================================================

func TestAllFeatures_Drawing(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Line
	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetLineWidth(1)
	pdf.Line(50, 50, 200, 50)

	// SetLineType
	pdf.SetLineType("dashed")
	pdf.Line(50, 60, 200, 60)

	// SetCustomLineType
	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(50, 70, 200, 70)

	// SetLineWidth
	pdf.SetLineWidth(2)
	pdf.SetLineType("solid")
	pdf.Line(50, 80, 200, 80)

	// Oval
	pdf.Oval(50, 100, 150, 150)

	// Polygon
	pdf.Polygon([]Point{
		{X: 200, Y: 100},
		{X: 250, Y: 150},
		{X: 200, Y: 150},
	}, "D")

	// ClipPolygon
	pdf.SaveGraphicsState()
	pdf.ClipPolygon([]Point{
		{X: 300, Y: 100},
		{X: 350, Y: 100},
		{X: 350, Y: 150},
		{X: 300, Y: 150},
	})
	pdf.RestoreGraphicsState()

	// Rectangle
	pdf.SetFillColor(255, 255, 255)
	if err := pdf.Rectangle(50, 200, 150, 250, "DF", 0, 0); err != nil {
		t.Fatalf("Rectangle: %v", err)
	}

	// Rectangle with round corners
	if err := pdf.Rectangle(200, 200, 300, 250, "DF", 10, 5); err != nil {
		t.Fatalf("Rectangle round: %v", err)
	}

	// RectFromUpperLeft
	pdf.RectFromUpperLeft(50, 270, 100, 30)

	// RectFromLowerLeft
	pdf.RectFromLowerLeft(50, 330, 100, 30)

	// RectFromUpperLeftWithStyle
	pdf.RectFromUpperLeftWithStyle(200, 270, 100, 30, "D")

	// RectFromLowerLeftWithStyle
	pdf.RectFromLowerLeftWithStyle(200, 330, 100, 30, "D")

	// RectFromUpperLeftWithOpts
	if err := pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		Rect:       Rect{W: 80, H: 25},
		X:          350, Y: 270,
		PaintStyle: DrawPaintStyle,
	}); err != nil {
		t.Fatalf("RectFromUpperLeftWithOpts: %v", err)
	}

	// RectFromLowerLeftWithOpts
	if err := pdf.RectFromLowerLeftWithOpts(DrawableRectOptions{
		Rect:       Rect{W: 80, H: 25},
		X:          350, Y: 330,
		PaintStyle: DrawPaintStyle,
	}); err != nil {
		t.Fatalf("RectFromLowerLeftWithOpts: %v", err)
	}

	// Curve
	pdf.Curve(50, 400, 100, 350, 150, 450, 200, 400, "D")

	if err := pdf.WritePdf(resOutDir + "/all_drawing.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 10. Images
// ============================================================

func TestAllFeatures_Images(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Image (JPEG)
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Skipf("JPEG image not available: %v", err)
	}

	// Image (PNG)
	if err := pdf.Image(resPNGPath, 200, 50, &Rect{W: 100, H: 100}); err != nil {
		t.Skipf("PNG image not available: %v", err)
	}

	// ImageHolderByPath -> ImageByHolder
	imgH, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skipf("ImageHolderByPath: %v", err)
	}
	if err := pdf.ImageByHolder(imgH, 50, 200, &Rect{W: 80, H: 80}); err != nil {
		t.Fatalf("ImageByHolder: %v", err)
	}

	// ImageHolderByBytes -> ImageByHolderWithOptions
	imgBytes, err := os.ReadFile(resJPEGPath)
	if err != nil {
		t.Skipf("cannot read image: %v", err)
	}
	imgH2, err := ImageHolderByBytes(imgBytes)
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	if err := pdf.ImageByHolderWithOptions(imgH2, ImageOptions{
		Rect: &Rect{W: 80, H: 80},
		X:    200, Y: 200,
	}); err != nil {
		t.Fatalf("ImageByHolderWithOptions: %v", err)
	}

	// ImageHolderByReader
	imgFile, err := os.Open(resJPEGPath)
	if err != nil {
		t.Skipf("cannot open image: %v", err)
	}
	defer imgFile.Close()
	imgH3, err := ImageHolderByReader(imgFile)
	if err != nil {
		t.Fatalf("ImageHolderByReader: %v", err)
	}
	if err := pdf.ImageByHolder(imgH3, 350, 200, &Rect{W: 80, H: 80}); err != nil {
		t.Fatalf("ImageByHolder from reader: %v", err)
	}

	// ImageByHolderWithOptions with Crop
	imgH4, _ := ImageHolderByBytes(imgBytes)
	if err := pdf.ImageByHolderWithOptions(imgH4, ImageOptions{
		Rect: &Rect{W: 100, H: 100},
		Crop: &CropOptions{X: 0, Y: 0, Width: 50, Height: 50},
	}); err != nil {
		t.Fatalf("ImageByHolderWithOptions with Crop: %v", err)
	}

	if err := pdf.WritePdf(resOutDir + "/all_images.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 11. Rotation & Transparency
// ============================================================

func TestAllFeatures_RotationTransparency(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Rotate
	pdf.Rotate(45, 150, 150)
	pdf.SetXY(100, 140)
	pdf.Cell(nil, "Rotated 45 degrees")
	pdf.RotateReset()

	// SetTransparency
	tr, err := NewTransparency(0.5, "")
	if err != nil {
		t.Fatalf("NewTransparency: %v", err)
	}
	if err := pdf.SetTransparency(tr); err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	pdf.SetXY(50, 200)
	pdf.Cell(nil, "50% transparent text")

	// ClearTransparency
	pdf.ClearTransparency()
	pdf.SetXY(50, 230)
	pdf.Cell(nil, "Opaque text again")

	if err := pdf.WritePdf(resOutDir + "/all_rotation_transparency.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 12. Graphics State
// ============================================================

func TestAllFeatures_GraphicsState(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SaveGraphicsState()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetLineWidth(3)
	pdf.Line(50, 50, 200, 50)
	pdf.RestoreGraphicsState()

	// After restore, stroke should be back to default
	pdf.Line(50, 70, 200, 70)

	if err := pdf.WritePdf(resOutDir + "/all_graphics_state.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 13. Links & Anchors
// ============================================================

func TestAllFeatures_LinksAnchors(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// AddExternalLink
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Click here for Google")
	pdf.AddExternalLink("https://www.google.com", 50, 50, 200, 20)

	// SetAnchor + AddInternalLink
	pdf.AddPage()
	pdf.SetAnchor("page2top")
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "This is page 2 anchor")

	if err := pdf.SetPage(1); err != nil {
		t.Fatalf("SetPage: %v", err)
	}
	pdf.SetXY(50, 80)
	pdf.Cell(nil, "Go to page 2")
	pdf.AddInternalLink("page2top", 50, 80, 200, 20)

	if err := pdf.WritePdf(resOutDir + "/all_links.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 14. Headers & Footers
// ============================================================

func TestAllFeatures_HeaderFooter(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)

	pdf.AddHeader(func() {
		pdf.SetY(5)
		pdf.Cell(nil, "== Header ==")
	})
	pdf.AddFooter(func() {
		pdf.SetY(825)
		pdf.Cell(nil, "== Footer ==")
	})

	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("Page 1 content")

	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("Page 2 content")

	if err := pdf.WritePdf(resOutDir + "/all_header_footer.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 15. Outlines
// ============================================================

func TestAllFeatures_Outlines(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)

	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Chapter 1 content")

	pdf.AddPage()
	outlineObj := pdf.AddOutlineWithPosition("Chapter 2")
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Chapter 2 content")
	if outlineObj == nil {
		t.Fatal("AddOutlineWithPosition returned nil")
	}

	if err := pdf.WritePdf(resOutDir + "/all_outlines.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 16. PDF Info
// ============================================================

func TestAllFeatures_PdfInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	now := time.Now()
	pdf.SetInfo(PdfInfo{
		Title:        "Test Title",
		Author:       "Test Author",
		Subject:      "Test Subject",
		Creator:      "GoPDF2 Test",
		Producer:     "GoPDF2",
		CreationDate: now,
	})

	info := pdf.GetInfo()
	if info.Title != "Test Title" {
		t.Fatalf("expected title 'Test Title', got %q", info.Title)
	}
	if info.Author != "Test Author" {
		t.Fatalf("expected author 'Test Author', got %q", info.Author)
	}
}

// ============================================================
// 17. Margins
// ============================================================

func TestAllFeatures_Margins(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// SetMargins
	pdf.SetMargins(10, 20, 10, 20)
	l, top, r, bot := pdf.Margins()
	if l != 10 || top != 20 || r != 10 || bot != 20 {
		t.Fatalf("Margins mismatch: got (%f, %f, %f, %f)", l, top, r, bot)
	}

	// Individual margin setters
	pdf.SetMarginLeft(15)
	pdf.SetMarginTop(25)
	pdf.SetMarginRight(15)
	pdf.SetMarginBottom(25)
	l, top, r, bot = pdf.Margins()
	if l != 15 || top != 25 || r != 15 || bot != 25 {
		t.Fatalf("Individual margins mismatch: got (%f, %f, %f, %f)", l, top, r, bot)
	}
}

// ============================================================
// 18. Unit Conversion
// ============================================================

func TestAllFeatures_UnitConversion(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4, Unit: UnitMM})

	// UnitsToPoints / PointsToUnits (instance methods)
	pts := pdf.UnitsToPoints(25.4) // 25.4mm = 1 inch = 72 points
	if pts < 71.9 || pts > 72.1 {
		t.Fatalf("expected ~72 points, got %f", pts)
	}
	mm := pdf.PointsToUnits(72)
	if mm < 25.3 || mm > 25.5 {
		t.Fatalf("expected ~25.4 mm, got %f", mm)
	}

	// Package-level functions
	pts2 := UnitsToPoints(UnitIN, 1.0)
	if pts2 != 72 {
		t.Fatalf("expected 72 points for 1 inch, got %f", pts2)
	}
	in := PointsToUnits(UnitIN, 72)
	if in != 1.0 {
		t.Fatalf("expected 1 inch for 72 points, got %f", in)
	}
}

// ============================================================
// 19. Compression
// ============================================================

func TestAllFeatures_Compression(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Compressed PDF")

	// SetCompressLevel
	pdf.SetCompressLevel(9)
	compressed, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	// SetNoCompression
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.SetXY(50, 50)
	pdf2.Cell(nil, "Compressed PDF")
	pdf2.SetNoCompression()
	uncompressed, err := pdf2.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	// Uncompressed should generally be larger
	if len(uncompressed) <= len(compressed) {
		t.Logf("warning: uncompressed (%d) not larger than compressed (%d)", len(uncompressed), len(compressed))
	}
}

// ============================================================
// 20. Protection
// ============================================================

func TestAllFeatures_Protection(t *testing.T) {
	ensureOutDir(t)
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
		Protection: PDFProtectionConfig{
			UseProtection: true,
			Permissions:   0,
			UserPass:      []byte("user123"),
			OwnerPass:     []byte("owner456"),
		},
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Protected PDF")

	if err := pdf.WritePdf(resOutDir + "/all_protected.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 21. Import PDF Pages
// ============================================================

func TestAllFeatures_ImportPages(t *testing.T) {
	ensureOutDir(t)
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// ImportPagesFromSource
	if err := pdf.ImportPagesFromSource(resTestPDF, "/MediaBox"); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)

	if err := pdf.SetPage(1); err != nil {
		t.Fatalf("SetPage(1): %v", err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Overlay on imported page 1")

	// GetPageSizes
	sizes := pdf.GetPageSizes(resTestPDF)
	if len(sizes) == 0 {
		t.Fatal("GetPageSizes returned empty map")
	}

	if err := pdf.WritePdf(resOutDir + "/all_import_pages.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

func TestAllFeatures_ImportPageStream(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	rs := io.ReadSeeker(bytes.NewReader(data))
	tpl := pdf.ImportPageStream(&rs, 1, "/MediaBox")
	pdf.AddPage()
	pdf.UseImportedTemplate(tpl, 0, 0, PageSizeA4.W, PageSizeA4.H)

	// GetStreamPageSizes
	rs2 := io.ReadSeeker(bytes.NewReader(data))
	streamSizes := pdf.GetStreamPageSizes(&rs2)
	if len(streamSizes) == 0 {
		t.Fatal("GetStreamPageSizes returned empty map")
	}

	out, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) {
		t.Fatal("output does not start with %PDF-")
	}
}

// ============================================================
// 22. OpenPDF (overlay approach)
// ============================================================

func TestAllFeatures_OpenPDF(t *testing.T) {
	ensureOutDir(t)

	// OpenPDF from file
	pdf := GoPdf{}
	if err := pdf.OpenPDF(resTestPDF, nil); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	if pdf.GetNumberOfPages() == 0 {
		t.Fatal("expected at least 1 page")
	}

	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.SetPage(1)
	pdf.SetXY(100, 100)
	pdf.Cell(nil, "OpenPDF overlay")

	if err := pdf.WritePdf(resOutDir + "/all_open_pdf.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}

	// OpenPDFFromBytes
	data, _ := os.ReadFile(resTestPDF)
	pdf2 := GoPdf{}
	if err := pdf2.OpenPDFFromBytes(data, nil); err != nil {
		t.Fatalf("OpenPDFFromBytes: %v", err)
	}
	if pdf2.GetNumberOfPages() == 0 {
		t.Fatal("expected pages from bytes")
	}

	// OpenPDFFromStream
	rs := io.ReadSeeker(bytes.NewReader(data))
	pdf3 := GoPdf{}
	if err := pdf3.OpenPDFFromStream(&rs, nil); err != nil {
		t.Fatalf("OpenPDFFromStream: %v", err)
	}
	if pdf3.GetNumberOfPages() == 0 {
		t.Fatal("expected pages from stream")
	}

	// OpenPDF with custom box option
	pdf4 := GoPdf{}
	if err := pdf4.OpenPDF(resTestPDF, &OpenPDFOption{Box: "/MediaBox"}); err != nil {
		t.Fatalf("OpenPDF with box: %v", err)
	}
}

// ============================================================
// 23. HTML Rendering (InsertHTMLBox)
// ============================================================

func TestAllFeatures_InsertHTMLBox(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	htmlStr := `<h1>Title</h1>
<p>This is a <b>bold</b> and <i>italic</i> paragraph with <u>underline</u>.</p>
<p style="color: red;">Red paragraph.</p>
<ul>
  <li>Item 1</li>
  <li>Item 2</li>
</ul>
<ol>
  <li>First</li>
  <li>Second</li>
</ol>
<hr/>
<center>Centered text</center>
<p>Link: <a href="https://example.com">Example</a></p>
<p>Sub<sub>script</sub> and Super<sup>script</sup></p>`

	endY, err := pdf.InsertHTMLBox(50, 50, 495, 700, htmlStr, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
		DefaultColor:      [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox: %v", err)
	}
	if endY <= 50 {
		t.Fatal("InsertHTMLBox did not advance Y position")
	}

	if err := pdf.WritePdf(resOutDir + "/all_html.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 24. Table Layout
// ============================================================

func TestAllFeatures_TableLayout(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	table := pdf.NewTableLayout(10, 10, 25, 5)

	table.AddColumn("ID", 40, "center")
	table.AddColumn("Name", 150, "left")
	table.AddColumn("Price", 80, "right")

	table.AddRow([]string{"1", "Widget A", "9.99"})
	table.AddRow([]string{"2", "Widget B", "19.99"})

	// AddStyledRow
	table.AddStyledRow([]RowCell{
		NewRowCell("3", CellStyle{TextColor: RGBColor{R: 255, G: 0, B: 0}}),
		NewRowCell("Widget C (red)", CellStyle{TextColor: RGBColor{R: 255, G: 0, B: 0}}),
		NewRowCell("29.99", CellStyle{TextColor: RGBColor{R: 255, G: 0, B: 0}}),
	})

	table.SetTableStyle(CellStyle{
		BorderStyle: BorderStyle{
			Top: true, Left: true, Bottom: true, Right: true,
			Width: 1.0,
		},
		FillColor: RGBColor{R: 255, G: 255, B: 255},
		TextColor: RGBColor{R: 0, G: 0, B: 0},
		FontSize:  10,
	})

	table.SetHeaderStyle(CellStyle{
		BorderStyle: BorderStyle{
			Top: true, Left: true, Bottom: true, Right: true,
			Width:    2.0,
			RGBColor: RGBColor{R: 0, G: 0, B: 200},
		},
		FillColor: RGBColor{R: 220, G: 220, B: 255},
		TextColor: RGBColor{R: 0, G: 0, B: 150},
		FontSize:  12,
	})

	table.SetCellStyle(CellStyle{
		BorderStyle: BorderStyle{
			Right: true, Bottom: true,
			Width: 0.5,
		},
		TextColor: RGBColor{R: 0, G: 0, B: 0},
		FontSize:  10,
	})

	if err := table.DrawTable(); err != nil {
		t.Fatalf("DrawTable: %v", err)
	}

	if err := pdf.WritePdf(resOutDir + "/all_table.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 25. Placeholder Text
// ============================================================

func TestAllFeatures_PlaceholderText(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// PlaceHolderText
	pdf.SetXY(50, 50)
	if err := pdf.PlaceHolderText("name_field", 200); err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	// FillInPlaceHoldText
	if err := pdf.FillInPlaceHoldText("name_field", "John Doe", Left); err != nil {
		t.Fatalf("FillInPlaceHoldText: %v", err)
	}

	if err := pdf.WritePdf(resOutDir + "/all_placeholder.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 26. Color Spaces
// ============================================================

func TestAllFeatures_ColorSpaces(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// AddColorSpaceRGB
	if err := pdf.AddColorSpaceRGB("myRGB", 255, 0, 0); err != nil {
		t.Fatalf("AddColorSpaceRGB: %v", err)
	}

	// AddColorSpaceCMYK
	if err := pdf.AddColorSpaceCMYK("myCMYK", 0, 100, 100, 0); err != nil {
		t.Fatalf("AddColorSpaceCMYK: %v", err)
	}

	// SetColorSpace
	if err := pdf.SetColorSpace("myRGB"); err != nil {
		t.Fatalf("SetColorSpace: %v", err)
	}

	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Color space test")

	if err := pdf.WritePdf(resOutDir + "/all_colorspace.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// ============================================================
// 27. Page Sizes
// ============================================================

func TestAllFeatures_PageSizes(t *testing.T) {
	// Verify predefined page sizes are non-nil and have positive dimensions
	sizes := []*Rect{
		PageSizeA0, PageSizeA1, PageSizeA2, PageSizeA3, PageSizeA3Landscape,
		PageSizeA4, PageSizeA4Landscape, PageSizeA4Small, PageSizeA5,
		PageSizeB4, PageSizeB5,
		PageSizeLetter, PageSizeLetterSmall, PageSizeLegal,
		PageSizeTabloid, PageSizeLedger,
		PageSizeExecutive, PageSizeFolio, PageSizeQuarto,
		PageSizeStatement, PageSize10x14,
	}
	for i, s := range sizes {
		if s == nil {
			t.Fatalf("page size %d is nil", i)
		}
		if s.W <= 0 || s.H <= 0 {
			t.Fatalf("page size %d has invalid dimensions: %fx%f", i, s.W, s.H)
		}
	}
}

// ============================================================
// 28. Error Cases
// ============================================================

func TestAllFeatures_ErrorCases(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// SetFont without adding font
	if err := pdf.SetFont("nonexistent", "", 14); err == nil {
		t.Fatal("expected error for nonexistent font family")
	}

	// OpenPDF with nonexistent file
	pdf2 := &GoPdf{}
	if err := pdf2.OpenPDF("nonexistent.pdf", nil); err == nil {
		t.Fatal("expected error for nonexistent PDF file")
	}

	// OpenPDFFromBytes with invalid data
	pdf3 := &GoPdf{}
	if err := pdf3.OpenPDFFromBytes([]byte("not a pdf"), nil); err == nil {
		t.Fatal("expected error for invalid PDF data")
	}

	// SetPage on empty document
	pdf4 := &GoPdf{}
	pdf4.Start(Config{PageSize: *PageSizeA4})
	if err := pdf4.SetPage(1); err == nil {
		t.Fatal("expected error for SetPage on empty document")
	}

	// SetColorSpace with unknown name
	pdf5 := newPDFWithFont(t)
	pdf5.AddPage()
	if err := pdf5.SetColorSpace("unknown"); err == nil {
		t.Fatal("expected error for unknown color space")
	}
}

// ============================================================
// 29. Start resets state (clear value)
// ============================================================

func TestAllFeatures_StartResetsState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "before reset")
	pdf.SetInfo(PdfInfo{Title: "Before"})

	// Re-start should reset state
	pdf.Start(Config{PageSize: *PageSizeA4})

	// After Start(), adding a page and writing should work cleanly
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "after reset")
	_, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr after reset: %v", err)
	}
}

// ============================================================
// 30. Multiple pages with mixed content
// ============================================================

func TestAllFeatures_MultiPageMixed(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)

	// Page 1: text
	pdf.AddPage()
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1: Text content")

	// Page 2: drawing
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 0)
	pdf.SetLineWidth(1)
	pdf.Line(50, 50, 300, 50)
	pdf.Oval(100, 100, 200, 200)
	if err := pdf.Rectangle(250, 100, 400, 200, "D", 0, 0); err != nil {
		t.Fatalf("Rectangle: %v", err)
	}

	// Page 3: image
	pdf.AddPage()
	if err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 200, H: 200}); err != nil {
		t.Logf("image not available, skipping image on page 3: %v", err)
	}

	// Page 4: HTML
	pdf.AddPage()
	pdf.InsertHTMLBox(50, 50, 495, 700, "<h2>HTML Page</h2><p>Content here.</p>", HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})

	if err := pdf.WritePdf(resOutDir + "/all_multipage_mixed.pdf"); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}
