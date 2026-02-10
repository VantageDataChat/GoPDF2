package gopdf

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/phpdave11/gofpdi"
)

// ============================================================
// coverage_boost11_test.go — TestCov11_ prefix
// Targets: SVG, SubsetFontObj.SetTTFByPath, PagesObj.test,
// StartWithImporter, CellWithOption, MultiCellWithOption,
// ImageObj.GetRect, maskHolder cached path, content_obj branches,
// embedded_file UpdateEmbeddedFile, page_info, pdf_lowlevel
// ============================================================

// ============================================================
// 1. SVG insert (75% → higher)
// ============================================================

func TestCov11_ImageSVGFromBytes(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="red" stroke="blue" stroke-width="2"/>
		<circle cx="50" cy="50" r="30" fill="green" opacity="0.5"/>
		<line x1="0" y1="0" x2="100" y2="100" stroke="black" stroke-width="1"/>
		<text x="10" y="90" fill="black">Hello</text>
	</svg>`)

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ImageSVGFromBytes(svg, SVGOption{
		X: 50, Y: 50, Width: 200, Height: 200,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes: %v", err)
	}
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov11_ImageSVGFromBytes_Polyline(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<polyline points="10,10 50,50 90,10" fill="none" stroke="red"/>
		<polygon points="20,80 50,20 80,80" fill="blue" stroke="green"/>
	</svg>`)

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ImageSVGFromBytes(svg, SVGOption{X: 10, Y: 10, Width: 100, Height: 100})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes polyline: %v", err)
	}
}

func TestCov11_ImageSVGFromBytes_Path(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<path d="M 10 10 L 90 10 L 90 90 L 10 90 Z" fill="yellow" stroke="black"/>
		<path d="M 50 10 C 80 10 80 90 50 90 S 20 10 50 10" fill="none" stroke="red"/>
	</svg>`)

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ImageSVGFromBytes(svg, SVGOption{X: 10, Y: 10, Width: 100, Height: 100})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes path: %v", err)
	}
}

func TestCov11_ImageSVGFromReader(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="50" height="50">
		<rect x="5" y="5" width="40" height="40" fill="#ff0000"/>
	</svg>`

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ImageSVGFromReader(strings.NewReader(svg), SVGOption{X: 10, Y: 10, Width: 50, Height: 50})
	if err != nil {
		t.Fatalf("ImageSVGFromReader: %v", err)
	}
}

// ============================================================
// 2. SubsetFontObj.SetTTFByPath (0%)
// ============================================================

func TestCov11_SubsetFontObj_SetTTFByPath(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	s := &SubsetFontObj{}
	s.SetTtfFontOption(defaultTtfFontOption())
	err := s.SetTTFByPath(resFontPath)
	if err != nil {
		t.Fatalf("SetTTFByPath: %v", err)
	}
}

func TestCov11_SubsetFontObj_SetTTFByPath_BadPath(t *testing.T) {
	s := &SubsetFontObj{}
	err := s.SetTTFByPath("/nonexistent/font.ttf")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

// ============================================================
// 3. PagesObj.test (0%)
// ============================================================

func TestCov11_PagesObj_Test(t *testing.T) {
	p := &PagesObj{}
	p.getRoot = func() *GoPdf {
		gp := &GoPdf{}
		gp.Start(Config{PageSize: *PageSizeA4})
		return gp
	}
	// Just call it — it prints to stdout.
	p.test()
}

// ============================================================
// 4. StartWithImporter (0%)
// ============================================================

func TestCov11_StartWithImporter(t *testing.T) {
	pdf := &GoPdf{}
	importer := gofpdi.NewImporter()
	pdf.StartWithImporter(Config{PageSize: *PageSizeA4}, importer)
	pdf.AddPage()
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 5. CellWithOption — transparency, border, alignment
// ============================================================

func TestCov11_CellWithOption_Border(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Bordered cell", CellOption{
		Border: Left | Right | Top | Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption border: %v", err)
	}
}

func TestCov11_CellWithOption_Align(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 100)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Right aligned", CellOption{
		Align: Right | Middle,
	})
	if err != nil {
		t.Fatalf("CellWithOption align: %v", err)
	}
}

func TestCov11_CellWithOption_Transparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 150)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Transparent", CellOption{
		Transparency: &Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode},
	})
	if err != nil {
		t.Fatalf("CellWithOption transparency: %v", err)
	}
}

// ============================================================
// 6. MultiCellWithOption
// ============================================================

func TestCov11_MultiCellWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 20}, "This is a long text that should wrap across multiple lines in the cell", CellOption{
		Align: Left,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}
}

func TestCov11_MultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 200)

	err := pdf.MultiCell(&Rect{W: 200, H: 20}, "Another long text for multi cell testing with word wrap")
	if err != nil {
		t.Fatalf("MultiCell: %v", err)
	}
}

// ============================================================
// 7. maskHolder — cached image path (63.9%)
// ============================================================

func TestCov11_MaskHolder_CachedPath(t *testing.T) {
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	holder, err := ImageHolderByPath(resPNGPath)
	if err != nil {
		t.Skipf("ImageHolderByPath: %v", err)
	}

	maskHolder, err := ImageHolderByPath(resPNGPath)
	if err != nil {
		t.Skipf("mask ImageHolderByPath: %v", err)
	}

	// First call — creates the mask.
	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X: 50, Y: 50,
		Rect: &Rect{W: 100, H: 100},
		Mask: &MaskOptions{
			Holder: maskHolder,
			ImageOptions: ImageOptions{
				X: 50, Y: 50,
				Rect: &Rect{W: 100, H: 100},
			},
		},
	})
	if err != nil {
		t.Logf("first mask call: %v", err)
	}

	// Second call with same holder — should use cached path.
	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X: 200, Y: 50,
		Rect: &Rect{W: 100, H: 100},
		Mask: &MaskOptions{
			Holder: maskHolder,
			ImageOptions: ImageOptions{
				X: 200, Y: 50,
				Rect: &Rect{W: 100, H: 100},
			},
		},
	})
	if err != nil {
		t.Logf("second mask call (cached): %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 8. UpdateEmbeddedFile (77.8%)
// ============================================================

func TestCov11_UpdateEmbeddedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "With attachment")

	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("original content"),
	})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}

	err = pdf.UpdateEmbeddedFile("test.txt", EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("updated content"),
	})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov11_UpdateEmbeddedFile_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.UpdateEmbeddedFile("nonexistent.txt", EmbeddedFile{
		Name:    "nonexistent.txt",
		Content: []byte("data"),
	})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ============================================================
// 9. Cell with underline
// ============================================================

func TestCov11_Cell_Underline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Underlined text", CellOption{
		Align: Left,
		Border: Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption underline: %v", err)
	}
}

// ============================================================
// 10. Read function (75%)
// ============================================================

func TestCov11_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Read test")

	buf := make([]byte, 64*1024)
	n, err := pdf.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes")
	}
}

// ============================================================
// 11. GetBytesPdf (75%)
// ============================================================

func TestCov11_GetBytesPdf_WithProtection(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.Cell(nil, "Protected PDF bytes")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 12. Line with options (75%)
// ============================================================

func TestCov11_Line_WithOptions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetLineWidth(2)
	pdf.SetStrokeColor(255, 0, 0)
	pdf.Line(10, 10, 200, 200)

	pdf.SetLineType("dashed")
	pdf.Line(10, 200, 200, 10)

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 13. ImageObj.getRect
// ============================================================

func TestCov11_ImageObj_GetRect(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	imgObj := &ImageObj{}
	imgObj.init(nil)
	err := imgObj.SetImagePath(resJPEGPath)
	if err != nil {
		t.Fatalf("SetImagePath: %v", err)
	}
	rect, err := imgObj.getRect()
	if err != nil {
		t.Fatalf("getRect: %v", err)
	}
	if rect.W <= 0 || rect.H <= 0 {
		t.Errorf("expected positive dimensions: %+v", rect)
	}
}

// ============================================================
// 14. Sector with style variations (77.8%)
// ============================================================

func TestCov11_Sector_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(255, 0, 0)
	pdf.Sector(200, 200, 50, 0, 120, "F")
	pdf.Sector(200, 200, 50, 120, 240, "FD")
	pdf.Sector(200, 200, 50, 240, 360, "D")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 15. Cell — various alignment combinations
// ============================================================

func TestCov11_Cell_AllAlignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	aligns := []int{Left, Center, Right, Top, Middle, Bottom, Left | Top, Center | Middle, Right | Bottom}
	for i, a := range aligns {
		pdf.SetXY(50, float64(50+i*30))
		pdf.CellWithOption(&Rect{W: 200, H: 25}, "Aligned", CellOption{Align: a})
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 16. AddTTFFontByReaderWithOption on GoPdf (77.8%)
// ============================================================

func TestCov11_GoPdf_AddTTFFontByReaderWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	data, _ := os.ReadFile(resFontPath)
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontByReaderWithOption("ReaderFont", bytes.NewReader(data), TtfOption{UseKerning: true})
	if err != nil {
		t.Fatalf("AddTTFFontByReaderWithOption: %v", err)
	}
}

func TestCov11_GoPdf_AddTTFFontDataWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	data, _ := os.ReadFile(resFontPath)
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontDataWithOption("DataFont", data, TtfOption{UseKerning: true})
	if err != nil {
		t.Fatalf("AddTTFFontDataWithOption: %v", err)
	}
}

// ============================================================
// 17. ClipPolygon
// ============================================================

func TestCov11_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	points := []Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}
	pdf.ClipPolygon(points)
	pdf.Cell(nil, "Clipped")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 18. SetNoCompression / SetCompressLevel
// ============================================================

func TestCov11_SetNoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.Cell(nil, "No compression")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 19. RectFromUpperLeftWithStyle variations
// ============================================================

func TestCov11_RectFromUpperLeft_Styles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(200, 200, 200)
	pdf.SetStrokeColor(0, 0, 0)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 50, "D")
	pdf.RectFromUpperLeftWithStyle(50, 120, 100, 50, "F")
	pdf.RectFromUpperLeftWithStyle(50, 190, 100, 50, "FD")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}
