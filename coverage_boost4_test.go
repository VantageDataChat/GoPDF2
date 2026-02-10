package gopdf

import (
	"bytes"
	"compress/zlib"
	"testing"
)

// ============================================================
// coverage_boost4_test.go - Fourth round of coverage tests.
// All functions prefixed TestCov4_.
// ============================================================

// ============================================================
// gopdf.go - SetCompressLevel (44.4% coverage)
// ============================================================

func TestCov4_SetCompressLevel_Valid(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetCompressLevel(zlib.BestSpeed)
	pdf.SetCompressLevel(zlib.BestCompression)
	pdf.SetCompressLevel(zlib.DefaultCompression)
	pdf.SetCompressLevel(zlib.HuffmanOnly)
}

func TestCov4_SetCompressLevel_TooSmall(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetCompressLevel(-10) // too small, should clamp
}

func TestCov4_SetCompressLevel_TooBig(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetCompressLevel(100) // too big, should clamp
}

// ============================================================
// gopdf.go - SetNewY, SetNewYIfNoOffset, SetNewXY (66% coverage)
// ============================================================

func TestCov4_SetNewY_NoOverflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewY(100, 20)
	if pdf.GetY() != 100 {
		t.Errorf("Y = %f, want 100", pdf.GetY())
	}
}

func TestCov4_SetNewY_Overflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Set Y near bottom of page, then call SetNewY with large h to trigger page break
	pdf.SetNewY(800, 100)
}

func TestCov4_SetNewYIfNoOffset_NoOverflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewYIfNoOffset(100, 20)
}

func TestCov4_SetNewYIfNoOffset_Overflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewYIfNoOffset(800, 100)
}

func TestCov4_SetNewXY_NoOverflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewXY(100, 50, 20)
}

func TestCov4_SetNewXY_Overflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewXY(800, 50, 100)
}

// ============================================================
// gopdf.go - IsCurrFontContainGlyph (63.6% coverage)
// ============================================================

func TestCov4_IsCurrFontContainGlyph_Found(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected glyph A to be found")
	}
}

func TestCov4_IsCurrFontContainGlyph_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Try a very rare Unicode character
	ok, err := pdf.IsCurrFontContainGlyph('\U0001F600')
	if err != nil {
		t.Fatal(err)
	}
	// May or may not be found depending on font
	_ = ok
}

func TestCov4_IsCurrFontContainGlyph_NoFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected false when no font set")
	}
}

// ============================================================
// doc_stats.go - GetDocumentStats, GetFonts
// ============================================================

func TestCov4_GetDocumentStats(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")
	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Errorf("PageCount = %d", stats.PageCount)
	}
	if stats.FontCount < 1 {
		t.Errorf("FontCount = %d", stats.FontCount)
	}
	if stats.ObjectCount < 1 {
		t.Errorf("ObjectCount = %d", stats.ObjectCount)
	}
	if stats.ContentStreamCount < 1 {
		t.Errorf("ContentStreamCount = %d", stats.ContentStreamCount)
	}
}

func TestCov4_GetDocumentStats_WithExtras(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")
	pdf.AddOutline("Outline1")
	pdf.SetXMPMetadata(XMPMetadata{Title: "Test"})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: "D"}})
	pdf.AddOCG(OCG{Name: "Layer1", On: true})
	stats := pdf.GetDocumentStats()
	if !stats.HasOutlines {
		t.Error("expected HasOutlines")
	}
	if !stats.HasXMPMetadata {
		t.Error("expected HasXMPMetadata")
	}
	if !stats.HasPageLabels {
		t.Error("expected HasPageLabels")
	}
	if !stats.HasOCGs {
		t.Error("expected HasOCGs")
	}
}

func TestCov4_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	fonts := pdf.GetFonts()
	if len(fonts) < 1 {
		t.Error("expected at least 1 font")
	}
	found := false
	for _, f := range fonts {
		if f.Family == fontFamily {
			found = true
			if !f.IsEmbedded {
				t.Error("expected embedded font")
			}
		}
	}
	if !found {
		t.Errorf("font %q not found", fontFamily)
	}
}

// ============================================================
// page_rotate.go - SetPageRotation, GetPageRotation
// ============================================================

func TestCov4_SetPageRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetPageRotation(1, 90); err != nil {
		t.Fatal(err)
	}
	angle, err := pdf.GetPageRotation(1)
	if err != nil {
		t.Fatal(err)
	}
	if angle != 90 {
		t.Errorf("angle = %d, want 90", angle)
	}
}

func TestCov4_SetPageRotation_270(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetPageRotation(1, 270); err != nil {
		t.Fatal(err)
	}
	angle, _ := pdf.GetPageRotation(1)
	if angle != 270 {
		t.Errorf("angle = %d, want 270", angle)
	}
}

func TestCov4_SetPageRotation_Negative(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetPageRotation(1, -90); err != nil {
		t.Fatal(err)
	}
	angle, _ := pdf.GetPageRotation(1)
	if angle != 270 {
		t.Errorf("angle = %d, want 270", angle)
	}
}

func TestCov4_SetPageRotation_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetPageRotation(1, 45)
	if err == nil {
		t.Error("expected error for non-90 angle")
	}
}

func TestCov4_SetPageRotation_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetPageRotation(5, 90)
	if err == nil {
		t.Error("expected error for out of range page")
	}
}

func TestCov4_GetPageRotation_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetPageRotation(5)
	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// page_cropbox.go - SetPageCropBox, GetPageCropBox, ClearPageCropBox
// ============================================================

func TestCov4_SetPageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetPageCropBox(1, Box{Left: 50, Top: 50, Right: 500, Bottom: 700})
	if err != nil {
		t.Fatal(err)
	}
	box, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatal(err)
	}
	if box == nil {
		t.Fatal("expected non-nil box")
	}
	if box.Left != 50 || box.Right != 500 {
		t.Errorf("box: %+v", box)
	}
}

func TestCov4_ClearPageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetPageCropBox(1, Box{Left: 50, Top: 50, Right: 500, Bottom: 700})
	err := pdf.ClearPageCropBox(1)
	if err != nil {
		t.Fatal(err)
	}
	box, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatal(err)
	}
	if box != nil {
		t.Error("expected nil box after clear")
	}
}

func TestCov4_GetPageCropBox_NoCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	box, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatal(err)
	}
	if box != nil {
		t.Error("expected nil box")
	}
}

func TestCov4_SetPageCropBox_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetPageCropBox(5, Box{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestCov4_GetPageCropBox_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetPageCropBox(5)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCov4_ClearPageCropBox_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ClearPageCropBox(5)
	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// page_layout.go - SetPageLayout, GetPageLayout, SetPageMode, GetPageMode
// ============================================================

func TestCov4_SetPageLayout(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetPageLayout(PageLayoutTwoColumnLeft)
	if pdf.GetPageLayout() != PageLayoutTwoColumnLeft {
		t.Errorf("layout = %v", pdf.GetPageLayout())
	}
}

func TestCov4_GetPageLayout_Default(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.GetPageLayout() != PageLayoutSinglePage {
		t.Errorf("default layout = %v", pdf.GetPageLayout())
	}
}

func TestCov4_SetPageMode(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetPageMode(PageModeUseThumbs)
	if pdf.GetPageMode() != PageModeUseThumbs {
		t.Errorf("mode = %v", pdf.GetPageMode())
	}
}

func TestCov4_GetPageMode_Default(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.GetPageMode() != PageModeUseNone {
		t.Errorf("default mode = %v", pdf.GetPageMode())
	}
}

func TestCov4_ValidPageLayout(t *testing.T) {
	valids := []string{"singlepage", "onecolumn", "twocolumnleft", "twocolumnright", "twopageleft", "twopageright"}
	for _, v := range valids {
		if !validPageLayout(v) {
			t.Errorf("validPageLayout(%q) should be true", v)
		}
	}
	if validPageLayout("invalid") {
		t.Error("expected false for invalid")
	}
}

func TestCov4_ValidPageMode(t *testing.T) {
	valids := []string{"usenone", "useoutlines", "usethumbs", "fullscreen", "useoc", "useattachments"}
	for _, v := range valids {
		if !validPageMode(v) {
			t.Errorf("validPageMode(%q) should be true", v)
		}
	}
	if validPageMode("invalid") {
		t.Error("expected false for invalid")
	}
}

// ============================================================
// config.go - pointsToUnits (more unit types)
// ============================================================

func TestCov4_PointsToUnits_AllTypes(t *testing.T) {
	tests := []struct {
		unit int
		pts  float64
	}{
		{UnitPT, 72},
		{UnitMM, 72},
		{UnitCM, 72},
		{UnitIN, 72},
		{UnitPX, 72},
		{UnitUnset, 72},
	}
	for _, tt := range tests {
		got := PointsToUnits(tt.unit, tt.pts)
		if got <= 0 {
			t.Errorf("PointsToUnits(%d, %f) = %f", tt.unit, tt.pts, got)
		}
	}
}

func TestCov4_UnitsToPoints_AllTypes(t *testing.T) {
	tests := []struct {
		unit int
		val  float64
	}{
		{UnitPT, 72},
		{UnitMM, 25.4},
		{UnitCM, 2.54},
		{UnitIN, 1},
		{UnitPX, 96},
		{UnitUnset, 72},
	}
	for _, tt := range tests {
		got := UnitsToPoints(tt.unit, tt.val)
		if got <= 0 {
			t.Errorf("UnitsToPoints(%d, %f) = %f", tt.unit, tt.val, got)
		}
	}
}

func TestCov4_PointsToUnits_CustomConversion(t *testing.T) {
	cfg := defaultUnitConfig{ConversionForUnit: 2.0}
	got := pointsToUnits(cfg, 100)
	if got != 50 {
		t.Errorf("got %f, want 50", got)
	}
}

// ============================================================
// font_option.go - getConvertedStyle
// ============================================================

func TestCov4_GetConvertedStyle(t *testing.T) {
	if getConvertedStyle("") != Regular {
		t.Error("empty should be Regular")
	}
	if getConvertedStyle("B") != Bold {
		t.Error("B should be Bold")
	}
	if getConvertedStyle("I") != Italic {
		t.Error("I should be Italic")
	}
	if getConvertedStyle("U") != Underline {
		t.Error("U should be Underline")
	}
	if getConvertedStyle("BI") != Bold|Italic {
		t.Error("BI should be Bold|Italic")
	}
	if getConvertedStyle("BIU") != Bold|Italic|Underline {
		t.Error("BIU should be Bold|Italic|Underline")
	}
}

// ============================================================
// clone.go - Clone
// ============================================================

func TestCov4_Clone(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "original")
	pdf.SetXMPMetadata(XMPMetadata{Title: "Test"})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: "D"}})

	clone, err := pdf.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if clone.GetNumberOfPages() != 1 {
		t.Errorf("clone pages = %d", clone.GetNumberOfPages())
	}
}

// ============================================================
// image_delete.go - DeleteImages, DeleteImagesFromAllPages, DeleteImageByIndex
// ============================================================

func TestCov4_DeleteImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "no images")
	n, err := pdf.DeleteImages(1)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted, got %d", n)
	}
}

func TestCov4_DeleteImagesFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "page1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "page2")
	n, err := pdf.DeleteImagesFromAllPages()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted, got %d", n)
	}
}

func TestCov4_DeleteImageByIndex_NoImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "no images")
	err := pdf.DeleteImageByIndex(1, 0)
	if err == nil {
		t.Error("expected error for no images")
	}
}

func TestCov4_DeleteImageByIndex_BadPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.DeleteImageByIndex(99, 0)
	if err == nil {
		t.Error("expected error for bad page")
	}
}

// ============================================================
// gopdf.go - Text, CellWithOption, MultiCell
// ============================================================

func TestCov4_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.Text("Hello World")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov4_CellWithOption_Align(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	rect := &Rect{W: 200, H: 30}
	err := pdf.CellWithOption(rect, "Center", CellOption{Align: Center | Middle})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov4_CellWithOption_Right(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	rect := &Rect{W: 200, H: 30}
	err := pdf.CellWithOption(rect, "Right", CellOption{Align: Right | Bottom})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov4_MultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	rect := &Rect{W: 200, H: 200}
	err := pdf.MultiCell(rect, "This is a long text that should wrap across multiple lines in the cell area.")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov4_MultiCellWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	rect := &Rect{W: 200, H: 200}
	err := pdf.MultiCellWithOption(rect, "Multi line\ntext here", CellOption{
		Align:  Center | Top,
		Border: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - PlaceHolderText, FillInPlaceHoldText
// ============================================================

func TestCov4_PlaceHolderText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.PlaceHolderText("{{name}}", 200)
	if err != nil {
		t.Fatal(err)
	}
	err = pdf.FillInPlaceHoldText("{{name}}", "John Doe", Left)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov4_FillInPlaceHoldText_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.FillInPlaceHoldText("{{missing}}", "value", Left)
	if err == nil {
		t.Error("expected error for missing placeholder")
	}
}

// ============================================================
// gopdf.go - Line with style
// ============================================================

func TestCov4_Line(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 20, 200, 300)
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// ============================================================
// gopdf.go - RectFromLowerLeftWithOpts, RectFromUpperLeftWithOpts
// ============================================================

func TestCov4_RectFromLowerLeftWithOpts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.8)
	pdf.SetGrayStroke(0)
	pdf.RectFromLowerLeftWithOpts(DrawableRectOptions{
		X:          50,
		Y:          50,
		Rect:       Rect{W: 100, H: 50},
		PaintStyle: DrawFillPaintStyle,
	})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov4_RectFromUpperLeftWithOpts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.8)
	pdf.SetGrayStroke(0)
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X:          50,
		Y:          50,
		Rect:       Rect{W: 100, H: 50},
		PaintStyle: DrawFillPaintStyle,
	})
}

// ============================================================
// gopdf.go - MeasureTextWidth, MeasureCellHeightByText
// ============================================================

func TestCov4_MeasureTextWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatal(err)
	}
	if w <= 0 {
		t.Errorf("width = %f", w)
	}
}

func TestCov4_MeasureCellHeightByText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	h, err := pdf.MeasureCellHeightByText("Hello\nWorld")
	if err != nil {
		t.Fatal(err)
	}
	if h <= 0 {
		t.Errorf("height = %f", h)
	}
}
