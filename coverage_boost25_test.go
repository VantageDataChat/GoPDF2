package gopdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost25_test.go — TestCov25_ prefix
// Targets: remaining uncovered branches in DeleteBookmark,
// collectOutlineObjs, cache_content_text underline,
// cache_content_text_color_cmyk equal, renderXObject,
// various 75% functions
// ============================================================

// --- Test: DeleteBookmark with multiple bookmarks to exercise all branches ---
func TestCov25_DeleteBookmark_AllBranches(t *testing.T) {
	pdf := newPDFWithFont(t)

	// Create 5 bookmarks
	for i := 0; i < 5; i++ {
		pdf.AddPage()
		pdf.AddOutline("Bookmark " + string(rune('A'+i)))
	}

	// Delete middle (index 2) — exercises prev>0 and next>0 branches
	err := pdf.DeleteBookmark(2)
	if err != nil {
		t.Fatalf("DeleteBookmark(2): %v", err)
	}

	// Delete first (index 0) — exercises prev<=0 branch (update parent's first)
	err = pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark(0): %v", err)
	}

	// Now remaining bookmarks are at indices 0 and 1 (after re-indexing)
	// Delete last — exercises next<=0 branch (update parent's last)
	outlines := pdf.getOutlineObjList()
	if len(outlines) > 0 {
		err = pdf.DeleteBookmark(len(outlines) - 1)
		if err != nil {
			t.Fatalf("DeleteBookmark(last): %v", err)
		}
	}

	var buf bytes.Buffer
	pdf.Write(&buf)
	if buf.Len() == 0 {
		t.Error("empty PDF")
	}
}

// --- Test: Underline text with CMYK color ---
func TestCov25_UnderlineTextCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set CMYK text color
	pdf.SetTextColorCMYK(0, 100, 100, 0) // red in CMYK

	// Set underline style
	pdf.SetFontWithStyle(fontFamily, Underline, 14)
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Underlined CMYK text")

	var buf bytes.Buffer
	err := pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// --- Test: Text with char spacing ---
func TestCov25_TextWithCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCharSpacing(2.0)
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Spaced text")

	// Also test MeasureTextWidth with char spacing
	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth: %v", err)
	}
	t.Logf("text width with spacing: %f", w)

	// MeasureCellHeightByText
	h, err := pdf.MeasureCellHeightByText("Hello World Multiline")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	t.Logf("cell height: %f", h)

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: CellWithOption with various alignments ---
func TestCov25_CellWithOption_Alignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	rect := &Rect{W: 200, H: 30}

	// Right + Bottom
	pdf.SetXY(10, 10)
	pdf.CellWithOption(rect, "Right Bottom", CellOption{Align: Right | Bottom})

	// Center + Middle
	pdf.SetXY(10, 50)
	pdf.CellWithOption(rect, "Center Middle", CellOption{Align: Center | Middle})

	// Left + Top (default)
	pdf.SetXY(10, 90)
	pdf.CellWithOption(rect, "Left Top", CellOption{Align: Left | Top})

	// With border
	pdf.SetXY(10, 130)
	pdf.CellWithOption(rect, "With Border", CellOption{
		Align:  Center | Middle,
		Border: AllBorders,
	})

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: Text function ---
func TestCov25_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	err := pdf.Text("Direct text output")
	if err != nil {
		t.Fatalf("Text: %v", err)
	}

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: Read function ---
func TestCov25_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Read test")

	// First write to get bytes
	var buf bytes.Buffer
	pdf.Write(&buf)

	// Now test Read
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.SetXY(10, 10)
	pdf2.Cell(nil, "Read test 2")

	p := make([]byte, 4096)
	n, err := pdf2.Read(p)
	if err != nil {
		t.Logf("Read: %v (n=%d)", err, n)
	}
	if n > 0 {
		t.Logf("Read %d bytes", n)
	}
}

// --- Test: GetBytesPdf ---
func TestCov25_GetBytesPdf(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "GetBytesPdf test")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("empty PDF bytes")
	}
}

// --- Test: RenderPageToImage with XObject ---
func TestCov25_RenderPageToImage_WithImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Image render test")

	// Add an image
	pngData := createTestPNG(t, 30, 30)
	holder, _ := ImageHolderByBytes(pngData)
	pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 100})

	var buf bytes.Buffer
	pdf.Write(&buf)

	img, err := RenderPageToImage(buf.Bytes(), 0, RenderOption{DPI: 72})
	if err != nil {
		t.Logf("RenderPageToImage: %v", err)
		return
	}
	bounds := img.Bounds()
	t.Logf("rendered image: %dx%d", bounds.Dx(), bounds.Dy())
}

// --- Test: cache_content_text_color_cmyk equal ---
func TestCov25_CacheContentTextColorCMYK_Equal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set CMYK colors and draw text to exercise the equal function
	pdf.SetTextColorCMYK(100, 0, 0, 0) // cyan
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Cyan text")

	// Same color again — should trigger equal() returning true
	pdf.SetTextColorCMYK(100, 0, 0, 0)
	pdf.SetXY(10, 30)
	pdf.Cell(nil, "Same cyan")

	// Different color — equal() returns false
	pdf.SetTextColorCMYK(0, 100, 0, 0) // magenta
	pdf.SetXY(10, 50)
	pdf.Cell(nil, "Magenta text")

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: HTML with <font face="..."> ---
func TestCov25_HTML_FontFace(t *testing.T) {
	pdf := newPDFWithFont(t)
	// Add a second font
	if err := pdf.AddTTFFont(fontFamily2, resFontPath2); err != nil {
		t.Skipf("second font not available: %v", err)
	}
	pdf.AddPage()

	html := `<font face="` + fontFamily2 + `">Different font</font>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox font face: %v", err)
	}
}

// --- Test: HTML with <p style="font-size:..."> ---
func TestCov25_HTML_ParagraphFontSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p style="font-size:20pt">Large text</p><p style="font-size:8pt">Small text</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox font-size: %v", err)
	}
}

// --- Test: HTML with <p style="background-color:..."> ---
func TestCov25_HTML_BackgroundColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p style="background-color:#FFFF00">Yellow background</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox bg color: %v", err)
	}
}

// --- Test: Render page with image XObject to exercise renderXObject ---
func TestCov25_RenderXObject(t *testing.T) {
	// Create a PDF with an image, then render it
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create a colorful PNG
	img := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 6), G: uint8(y * 6), B: 128, A: 255})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	holder, _ := ImageHolderByBytes(pngBuf.Bytes())
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 200, H: 200})

	var buf bytes.Buffer
	pdf.Write(&buf)

	rendered, err := RenderPageToImage(buf.Bytes(), 0, RenderOption{DPI: 72})
	if err != nil {
		t.Logf("RenderPageToImage with XObject: %v", err)
		return
	}
	t.Logf("rendered: %dx%d", rendered.Bounds().Dx(), rendered.Bounds().Dy())
}

// --- Test: Cell with nil rect ---
func TestCov25_Cell_NilRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	err := pdf.Cell(nil, "Nil rect cell")
	if err != nil {
		t.Fatalf("Cell nil rect: %v", err)
	}
}

// --- Test: Cell with rect ---
func TestCov25_Cell_WithRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	err := pdf.Cell(&Rect{W: 200, H: 30}, "Rect cell")
	if err != nil {
		t.Fatalf("Cell with rect: %v", err)
	}
}

// --- Test: cache_content_rotate write ---
func TestCov25_CacheContentRotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Rotate and draw
	pdf.Rotate(45, 100, 100)
	pdf.SetXY(80, 90)
	pdf.Cell(nil, "Rotated text")
	pdf.RotateReset()

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: cache_content_rotate with protection ---
func TestCov25_CacheContentRotate_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()

	pdf.Rotate(30, 100, 100)
	pdf.SetXY(80, 90)
	pdf.Cell(nil, "Protected rotated")
	pdf.RotateReset()

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: Underline with border ---
func TestCov25_UnderlineWithBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetFontWithStyle(fontFamily, Underline, 14)
	pdf.SetXY(10, 10)
	pdf.CellWithOption(&Rect{W: 200, H: 30}, "Underlined with border", CellOption{
		Align:  Left | Top,
		Border: AllBorders,
	})

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: Multiple pages with different sizes ---
func TestCov25_MultiplePageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)

	// A4 page
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "A4 page")

	// Letter page
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 612, H: 792},
	})
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Letter page")

	// Custom small page
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 200, H: 200},
	})
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Small page")

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: IsCurrFontContainGlyph ---
func TestCov25_IsCurrFontContainGlyph(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Test with ASCII char
	found, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph('A'): %v", err)
	}
	t.Logf("'A' found: %v", found)

	// Test with rare Unicode char
	found2, err := pdf.IsCurrFontContainGlyph(0x1F600) // emoji
	if err != nil {
		t.Logf("IsCurrFontContainGlyph(emoji): %v", err)
	} else {
		t.Logf("emoji found: %v", found2)
	}
}

// --- Test: PlaceHolderText + FillInPlaceHoldText ---
func TestCov25_PlaceHolderText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add placeholder
	pdf.SetXY(10, 10)
	err := pdf.PlaceHolderText("ph1", 200)
	if err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	// Fill with left align (default)
	err = pdf.FillInPlaceHoldText("ph1", "Left aligned", Left)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText left: %v", err)
	}

	// Add another placeholder
	pdf.SetXY(10, 40)
	pdf.PlaceHolderText("ph2", 200)
	pdf.FillInPlaceHoldText("ph2", "Right aligned", Right)

	// Add another placeholder
	pdf.SetXY(10, 70)
	pdf.PlaceHolderText("ph3", 200)
	pdf.FillInPlaceHoldText("ph3", "Center aligned", Center)

	// Try to fill non-existent placeholder
	err = pdf.FillInPlaceHoldText("nonexistent", "text", Left)
	if err == nil {
		t.Error("expected error for non-existent placeholder")
	}

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: Polygon and Polyline ---
func TestCov25_PolygonPolyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Polygon
	points := []Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}
	pdf.Polygon(points, "D")

	// Polyline
	linePoints := []Point{{X: 200, Y: 50}, {X: 250, Y: 100}, {X: 300, Y: 50}, {X: 350, Y: 100}}
	pdf.Polyline(linePoints)

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: ClipPolygon ---
func TestCov25_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SaveGraphicsState()
	points := []Point{{X: 50, Y: 50}, {X: 200, Y: 50}, {X: 200, Y: 200}, {X: 50, Y: 200}}
	pdf.ClipPolygon(points)
	pdf.SetXY(60, 60)
	pdf.Cell(nil, "Clipped text")
	pdf.RestoreGraphicsState()

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: HTML with very narrow box (forces word wrapping) ---
func TestCov25_HTML_NarrowBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Word wrap test with narrow box</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 30, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox narrow: %v", err)
	}
}

// --- Test: HTML with center alignment and multiple words ---
func TestCov25_HTML_CenterMultipleWords(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<center>` + strings.Repeat("Word ", 20) + `</center>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox center multi: %v", err)
	}
}

// --- Test: HTML with right alignment and multiple words ---
func TestCov25_HTML_RightMultipleWords(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p style="text-align:right">` + strings.Repeat("Word ", 20) + `</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox right multi: %v", err)
	}
}

// --- Test: Sector ---
func TestCov25_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Sector(200, 200, 100, 0, 90, "FD")
	pdf.Sector(200, 200, 100, 90, 180, "D")
	pdf.Sector(200, 200, 100, 180, 270, "F")

	var buf bytes.Buffer
	pdf.Write(&buf)
}

// --- Test: SetTransparency with various blend modes ---
func TestCov25_SetTransparency_BlendModes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	modes := []BlendModeType{NormalBlendMode, Multiply, Screen, Overlay, Darken, Lighten}
	for i, mode := range modes {
		err := pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: mode})
		if err != nil {
			t.Fatalf("SetTransparency(%v): %v", mode, err)
		}
		pdf.SetXY(10, float64(10+i*30))
		pdf.Cell(nil, "Transparent text")
		pdf.ClearTransparency()
	}

	var buf bytes.Buffer
	pdf.Write(&buf)
}
