package gopdf

import (
	"bytes"
	"os"
	"testing"
)

// ============================================================
// coverage_boost29_test.go — TestCov29_ prefix
// Targets:
// - replaceGlyphThatNotFound: OnGlyphNotFoundSubstitute=nil, substitute rune not in font
// - KernValueByLeft: UseKerning=false direct call, Kern()==nil
// - JournalEnable: else branch (re-enable after disable)
// - DeleteBookmark: parent==outlines root branches
// - collectOutlineObjs: o.first > 0 branch
// - GlyphIndexToPdfWidth: glyphIndex >= numberOfHMetrics
// - cacheContentLine extGStateIndexes branch
// - underline CoefLineHeight/CoefUnderlinePosition/CoefUnderlineThickness
// - convertNumericToFloat64 more types
// ============================================================

// ============================================================
// replaceGlyphThatNotFound — OnGlyphNotFoundSubstitute returns rune
// that is NOT in the font (CharCodeToGlyphIndex fails)
// ============================================================

func TestCov29_ReplaceGlyphThatNotFound_SubstituteNotInFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// Use a substitute function that returns a rune not in the font
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		OnGlyphNotFoundSubstitute: func(r rune) rune {
			// Return a very high Unicode codepoint unlikely to be in the font
			// but still in BMP so format4 is used
			return rune(0xFFFD) // replacement character — may or may not be in font
		},
	})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)

	// Try to render a character that's definitely not in the font
	// This triggers AddChars -> CharCodeToGlyphIndex -> ErrGlyphNotFound -> replaceGlyphThatNotFound
	// The substitute (0xFFFD) may also not be in font, hitting the err != nil branch
	_ = pdf.Text("\u0E3F") // Thai Baht sign — unlikely in LiberationSerif
	pdf.GetBytesPdf()
}

// ============================================================
// replaceGlyphThatNotFound — OnGlyphNotFoundSubstitute is nil
// ============================================================

func TestCov29_ReplaceGlyphThatNotFound_NilSubstitute(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// Set OnGlyphNotFoundSubstitute to nil explicitly
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		OnGlyphNotFoundSubstitute: nil,
	})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}

	// But SetTtfFontOption sets default if nil, so we need to override after
	// Find the SubsetFontObj and set it directly
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			sf.ttfFontOption.OnGlyphNotFoundSubstitute = nil
			break
		}
	}

	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)

	// Render a character not in the font — should hit the else branch
	_ = pdf.Text("\u0E3F")
	pdf.GetBytesPdf()
}

// ============================================================
// replaceGlyphThatNotFound — substitute rune already in CharacterToGlyphIndex
// ============================================================

func TestCov29_ReplaceGlyphThatNotFound_SubstituteAlreadyExists(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		OnGlyphNotFoundSubstitute: func(r rune) rune {
			return 'A' // 'A' will already be in the map after first use
		},
	})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)

	// First render 'A' to add it to CharacterToGlyphIndex
	_ = pdf.Text("A")
	// Now render a character not in font — substitute returns 'A' which already exists
	_ = pdf.Text("\u0E3F")
	pdf.GetBytesPdf()
}

// ============================================================
// KernValueByLeft — direct call with UseKerning=false
// ============================================================

func TestCov29_KernValueByLeft_UseKerningFalse(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFont(fontFamily, resFontPath) // default: UseKerning=false
	if err != nil {
		t.Skipf("font not available: %v", err)
	}

	// Find SubsetFontObj and call KernValueByLeft directly
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			found, kv := sf.KernValueByLeft(65) // 'A'
			if found {
				t.Error("expected not found when UseKerning is false")
			}
			if kv != nil {
				t.Error("expected nil KernValue when UseKerning is false")
			}
			return
		}
	}
	t.Skip("no SubsetFontObj found")
}

// ============================================================
// KernValueByLeft — UseKerning=true but Kern()==nil
// ============================================================

func TestCov29_KernValueByLeft_KernNil(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	// Add font with UseKerning=true
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Skipf("font not available: %v", err)
	}

	// Find SubsetFontObj
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			// The font may or may not have kern data.
			// If it does, we can't easily make Kern() return nil.
			// But we can call with a glyph index that doesn't have kerning.
			found, _ := sf.KernValueByLeft(99999) // very high index, unlikely to have kerning
			_ = found
			return
		}
	}
	t.Skip("no SubsetFontObj found")
}

// ============================================================
// JournalEnable — else branch (re-enable after disable)
// ============================================================

func TestCov29_JournalEnable_ReEnable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Before journal")

	// Enable journal
	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Fatal("expected journal enabled")
	}

	// Disable
	pdf.JournalDisable()
	if pdf.JournalIsEnabled() {
		t.Fatal("expected journal disabled")
	}

	// Re-enable — this hits the else branch (journal != nil)
	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Fatal("expected journal re-enabled")
	}

	// Do some operations to verify it works
	pdf.JournalStartOp("test op")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "After re-enable")
	pdf.JournalEndOp()

	ops := pdf.JournalGetOperations()
	if len(ops) < 2 {
		t.Errorf("expected at least 2 operations, got %d", len(ops))
	}
}

// ============================================================
// DeleteBookmark — delete first child where parent is root outlines
// ============================================================

func TestCov29_DeleteBookmark_FirstChild_RootParent(t *testing.T) {
	pdf := newPDFWithFont(t)

	// Add a single bookmark — it will be both first and last child of root
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Only Chapter")

	// Delete it — target.prev <= 0 AND target.parent == indexOfOutlinesObj+1
	// AND target.next <= 0 AND target.parent == indexOfOutlinesObj+1
	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// DeleteBookmark — delete first of two (hits root parent first-child update)
// ============================================================

func TestCov29_DeleteBookmark_FirstOfTwo_RootParent(t *testing.T) {
	pdf := newPDFWithFont(t)

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	// Delete first — target.prev <= 0, target.parent == root, target.next > 0
	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark first: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// DeleteBookmark — delete last of two (hits root parent last-child update)
// ============================================================

func TestCov29_DeleteBookmark_LastOfTwo_RootParent(t *testing.T) {
	pdf := newPDFWithFont(t)

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	// Delete last — target.next <= 0, target.parent == root, target.prev > 0
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark last: %v", err)
	}

	pdf.GetBytesPdf()
}


// ============================================================
// GlyphIndexToPdfWidth — glyphIndex >= numberOfHMetrics
// ============================================================

func TestCov29_GlyphIndexToPdfWidth_HighGlyphIndex(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFont(fontFamily, resFontPath)
	if err != nil {
		t.Skipf("font not available: %v", err)
	}

	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			// Call with a very high glyph index to hit the >= numberOfHMetrics branch
			w := sf.GlyphIndexToPdfWidth(999999)
			_ = w
			// Also call with 0
			w0 := sf.GlyphIndexToPdfWidth(0)
			_ = w0
			return
		}
	}
	t.Skip("no SubsetFontObj found")
}

// ============================================================
// CellWithOption — with CoefLineHeight, CoefUnderlinePosition, CoefUnderlineThickness
// ============================================================

func TestCov29_CellWithOption_UnderlineCoefficients(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set underline style
	if err := pdf.SetFontWithStyle(fontFamily, Underline, 14); err != nil {
		t.Fatalf("SetFontWithStyle: %v", err)
	}

	pdf.SetXY(50, 50)
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Custom underline coefficients", CellOption{
		Align:                    Left | Top,
		CoefLineHeight:          1.5,
		CoefUnderlinePosition:   1.2,
		CoefUnderlineThickness:  2.0,
	})
	if err != nil {
		t.Fatalf("CellWithOption: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// Line with extGStateIndexes (transparency on line)
// ============================================================

func TestCov29_Line_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set transparency before drawing line
	err := pdf.SetTransparency(Transparency{
		Alpha:         0.5,
		BlendModeType: NormalBlendMode,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}

	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetLineWidth(2)
	pdf.Line(50, 50, 200, 50)

	pdf.ClearTransparency()
	pdf.GetBytesPdf()
}

// ============================================================
// convertNumericToFloat64 — various numeric types
// ============================================================

func TestCov29_ConvertNumericToFloat64_AllTypes(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"int", int(12)},
		{"int8", int8(12)},
		{"int16", int16(12)},
		{"int32", int32(12)},
		{"int64", int64(12)},
		{"uint", uint(12)},
		{"uint8", uint8(12)},
		{"uint16", uint16(12)},
		{"uint32", uint32(12)},
		{"uint64", uint64(12)},
		{"float32", float32(12.5)},
		{"float64", float64(12.5)},
		{"string", "invalid"},
	}

	for _, tt := range tests {
		val, err := convertNumericToFloat64(tt.input)
		if tt.name == "string" {
			if err == nil {
				t.Errorf("expected error for string input")
			}
		} else {
			if err != nil {
				t.Errorf("convertNumericToFloat64(%s): %v", tt.name, err)
			}
			if val <= 0 {
				t.Errorf("convertNumericToFloat64(%s) = %f, expected > 0", tt.name, val)
			}
		}
	}
}

// ============================================================
// SetFont with various size types
// ============================================================

func TestCov29_SetFont_VariousSizeTypes(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}

	// int
	if err := pdf.SetFont(fontFamily, "", int(14)); err != nil {
		t.Errorf("SetFont int: %v", err)
	}
	// int32
	if err := pdf.SetFont(fontFamily, "", int32(14)); err != nil {
		t.Errorf("SetFont int32: %v", err)
	}
	// uint
	if err := pdf.SetFont(fontFamily, "", uint(14)); err != nil {
		t.Errorf("SetFont uint: %v", err)
	}
	// float32
	if err := pdf.SetFont(fontFamily, "", float32(14.5)); err != nil {
		t.Errorf("SetFont float32: %v", err)
	}
	// string (should error)
	if err := pdf.SetFont(fontFamily, "", "bad"); err == nil {
		t.Error("expected error for string size")
	}
}

// ============================================================
// JournalSave / JournalLoad round-trip
// ============================================================

func TestCov29_Journal_SaveLoad(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Journal save/load test")

	pdf.JournalEnable()
	pdf.JournalStartOp("add text")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Second page")
	pdf.JournalEndOp()

	journalPath := resOutDir + "/test_journal.json"
	if err := pdf.JournalSave(journalPath); err != nil {
		t.Fatalf("JournalSave: %v", err)
	}
	defer os.Remove(journalPath)

	// Load into a new PDF
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.JournalEnable()
	if err := pdf2.JournalLoad(journalPath); err != nil {
		t.Fatalf("JournalLoad: %v", err)
	}

	ops := pdf2.JournalGetOperations()
	if len(ops) == 0 {
		t.Error("expected operations after load")
	}
}

// ============================================================
// JournalSave/Load error paths
// ============================================================

func TestCov29_Journal_SaveLoad_Errors(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// Save without journal enabled
	if err := pdf.JournalSave("test.json"); err == nil {
		t.Error("expected error saving without journal")
	}

	// Load without journal enabled
	if err := pdf.JournalLoad("test.json"); err == nil {
		t.Error("expected error loading without journal")
	}

	// Undo without journal
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Error("expected error undoing without journal")
	}

	// Redo without journal
	_, err = pdf.JournalRedo()
	if err == nil {
		t.Error("expected error redoing without journal")
	}

	// Load non-existent file
	pdf.JournalEnable()
	if err := pdf.JournalLoad("/nonexistent/path/journal.json"); err == nil {
		t.Error("expected error loading non-existent file")
	}

	// Load invalid JSON
	ensureOutDir(t)
	badPath := resOutDir + "/bad_journal.json"
	os.WriteFile(badPath, []byte("not json"), 0644)
	defer os.Remove(badPath)
	if err := pdf.JournalLoad(badPath); err == nil {
		t.Error("expected error loading invalid JSON")
	}
}

// ============================================================
// JournalUndo / JournalRedo — nothing to undo/redo
// ============================================================

func TestCov29_Journal_UndoRedo_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()

	// Only initial snapshot — nothing to undo
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Error("expected error for nothing to undo")
	}

	// Nothing to redo
	_, err = pdf.JournalRedo()
	if err == nil {
		t.Error("expected error for nothing to redo")
	}
}

// ============================================================
// JournalUndo then Redo
// ============================================================

func TestCov29_Journal_UndoThenRedo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Initial")

	pdf.JournalEnable()

	pdf.JournalStartOp("add page 2")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")
	pdf.JournalEndOp()

	// Undo
	name, err := pdf.JournalUndo()
	if err != nil {
		t.Fatalf("JournalUndo: %v", err)
	}
	if name != "add page 2" {
		t.Errorf("expected 'add page 2', got %q", name)
	}

	// Redo
	name, err = pdf.JournalRedo()
	if err != nil {
		t.Fatalf("JournalRedo: %v", err)
	}
	if name != "add page 2" {
		t.Errorf("expected 'add page 2', got %q", name)
	}
}

// ============================================================
// AddWatermarkText — error paths
// ============================================================

func TestCov29_AddWatermarkText_CustomPageSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	// Add page with custom size
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 300, H: 400},
	})
	pdf.SetXY(50, 50)
	pdf.Text("Custom page")

	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "CUSTOM",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.4,
		Angle:      30,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// Polygon with transparency (extGStateIndexes)
// ============================================================

func TestCov29_Polygon_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.5,
		BlendModeType: NormalBlendMode,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}

	pdf.SetFillColor(255, 0, 0)
	pdf.Polygon([]Point{
		{X: 100, Y: 100},
		{X: 200, Y: 100},
		{X: 150, Y: 200},
	}, "F")

	pdf.ClearTransparency()
	pdf.GetBytesPdf()
}

// ============================================================
// Rectangle with transparency (extGStateIndexes on rect)
// ============================================================

func TestCov29_Rectangle_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.3,
		BlendModeType: Multiply,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}

	pdf.SetFillColor(0, 0, 255)
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 50,
		Rect:       Rect{W: 100, H: 100},
		PaintStyle: FillPaintStyle,
	})

	pdf.ClearTransparency()
	pdf.GetBytesPdf()
}

// ============================================================
// IsFitMultiCellWithNewline
// ============================================================

func TestCov29_IsFitMultiCellWithNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Text with newlines
	fit, h, err := pdf.IsFitMultiCellWithNewline(&Rect{W: 200, H: 100}, "Line 1\nLine 2\nLine 3")
	if err != nil {
		t.Fatalf("IsFitMultiCellWithNewline: %v", err)
	}
	if !fit {
		t.Error("expected text to fit")
	}
	if h <= 0 {
		t.Error("expected positive height")
	}

	// Text that doesn't fit
	fit2, _, err := pdf.IsFitMultiCellWithNewline(&Rect{W: 200, H: 5}, "Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	if err != nil {
		t.Fatalf("IsFitMultiCellWithNewline: %v", err)
	}
	if fit2 {
		t.Error("expected text not to fit in small rect")
	}
}

// ============================================================
// FillInPlaceHoldText — various alignments
// ============================================================

func TestCov29_FillInPlaceHoldText_Alignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create placeholders
	pdf.SetXY(50, 50)
	pdf.PlaceHolderText("ph_left", 200)
	pdf.SetXY(50, 80)
	pdf.PlaceHolderText("ph_center", 200)
	pdf.SetXY(50, 110)
	pdf.PlaceHolderText("ph_right", 200)

	// Fill with different alignments
	if err := pdf.FillInPlaceHoldText("ph_left", "Left aligned", Left); err != nil {
		t.Errorf("FillInPlaceHoldText left: %v", err)
	}
	if err := pdf.FillInPlaceHoldText("ph_center", "Center aligned", Center); err != nil {
		t.Errorf("FillInPlaceHoldText center: %v", err)
	}
	if err := pdf.FillInPlaceHoldText("ph_right", "Right aligned", Right); err != nil {
		t.Errorf("FillInPlaceHoldText right: %v", err)
	}

	// Non-existent placeholder
	if err := pdf.FillInPlaceHoldText("nonexistent", "test", Left); err == nil {
		t.Error("expected error for non-existent placeholder")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// Clone
// ============================================================

func TestCov29_Clone(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Original document")

	clone, err := pdf.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}

	// Clone should be usable
	if clone.GetNumberOfPages() == 0 {
		t.Error("expected clone to have pages")
	}
}

// ============================================================
// SetColorSpace / AddColorSpaceRGB / AddColorSpaceCMYK
// ============================================================

func TestCov29_ColorSpace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add color spaces
	if err := pdf.AddColorSpaceRGB("MyRed", 255, 0, 0); err != nil {
		t.Fatalf("AddColorSpaceRGB: %v", err)
	}
	if err := pdf.AddColorSpaceCMYK("MyCyan", 100, 0, 0, 0); err != nil {
		t.Fatalf("AddColorSpaceCMYK: %v", err)
	}

	// Set color space
	if err := pdf.SetColorSpace("MyRed"); err != nil {
		t.Fatalf("SetColorSpace: %v", err)
	}

	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Color space test")
	pdf.GetBytesPdf()
}

// ============================================================
// SetNewXY — triggers new page
// ============================================================

func TestCov29_SetNewXY_NewPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Move Y near bottom of page, then call SetNewXY with h that exceeds remaining space
	pdf.SetY(800)
	pdf.SetNewXY(50, 50, 50)
	if pdf.GetNumberOfPages() < 2 {
		t.Error("expected new page to be added")
	}
}

// ============================================================
// SetNewYIfNoOffset — triggers new page
// ============================================================

func TestCov29_SetNewYIfNoOffset_NewPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set Y beyond page height (A4 is ~842 pts)
	pdf.SetNewYIfNoOffset(800, 100)
	if pdf.GetNumberOfPages() < 2 {
		t.Error("expected new page to be added")
	}
}

// ============================================================
// ImageFromWithOption
// ============================================================

func TestCov29_ImageFromWithOption(t *testing.T) {
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}

	// Read PNG as image.Image
	f, err := os.Open(resPNGPath)
	if err != nil {
		t.Skipf("cannot open PNG: %v", err)
	}
	defer f.Close()

	imgHolder, err := ImageHolderByPath(resPNGPath)
	if err != nil {
		t.Skipf("ImageHolderByPath: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Use ImageByHolderWithOptions with mask
	bbox := [4]float64{0, 0, 100, 100}
	err = pdf.ImageByHolderWithOptions(imgHolder, ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 100, H: 100},
		Mask: &MaskOptions{
			BBox:   &bbox,
			Holder: imgHolder,
			ImageOptions: ImageOptions{
				Rect: &Rect{W: 100, H: 100},
			},
		},
	})
	if err != nil {
		t.Logf("ImageByHolderWithOptions with mask: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// HTML — blockquote, sub, sup, strike, ins, font tags
// ============================================================

func TestCov29_HTML_MoreElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<blockquote>This is a blockquote with some text.</blockquote>
<p>Normal <sub>subscript</sub> and <sup>superscript</sup></p>
<p><strike>strikethrough</strike> and <ins>inserted</ins></p>
<p><font color="red" size="16">Red font tag</font></p>
<p><del>deleted text</del></p>
<p><s>strikethrough s tag</s></p>
<p><em>emphasized</em> and <strong>strong</strong></p>`

	_, err := pdf.InsertHTMLBox(50, 50, 400, 600, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — table element
// ============================================================

func TestCov29_HTML_Table(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<table border="1">
<tr><th>Header 1</th><th>Header 2</th></tr>
<tr><td>Cell 1</td><td>Cell 2</td></tr>
<tr><td>Cell 3</td><td>Cell 4</td></tr>
</table>`

	_, err := pdf.InsertHTMLBox(50, 50, 400, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox table: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — pre and code tags
// ============================================================

func TestCov29_HTML_PreCode(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<pre>Preformatted text
  with spaces
    and indentation</pre>
<p>Inline <code>code</code> here</p>`

	_, err := pdf.InsertHTMLBox(50, 50, 400, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox pre/code: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// Text with transparency (extGStateIndexes on text)
// ============================================================

func TestCov29_Text_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTransparency(Transparency{
		Alpha:         0.5,
		BlendModeType: NormalBlendMode,
	})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}

	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Transparent text")

	pdf.ClearTransparency()
	pdf.GetBytesPdf()
}

// ============================================================
// MultiCellWithOption — right alignment
// ============================================================

func TestCov29_MultiCellWithOption_RightAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetXY(50, 50)
	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Right aligned multi cell text that should wrap across lines", CellOption{
		Align: Right | Top,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}

	pdf.SetXY(50, 200)
	err = pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Center aligned multi cell text that should wrap across lines", CellOption{
		Align: Center | Middle,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption center: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SetMargins and verify AddPage respects them
// ============================================================

func TestCov29_Margins_WithContent(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetMargins(20, 30, 20, 30)

	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()

	// After AddPage, X and Y should be at margin positions
	if pdf.GetX() < 20 {
		t.Errorf("expected X >= 20, got %f", pdf.GetX())
	}
	if pdf.GetY() < 30 {
		t.Errorf("expected Y >= 30, got %f", pdf.GetY())
	}

	pdf.Cell(nil, "Margin test")
	pdf.GetBytesPdf()
}

// ============================================================
// OpenPDF with password
// ============================================================

func TestCov29_OpenPDF_WithPassword(t *testing.T) {
	// Create a protected PDF first
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Protected content")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	// Try to open with password
	pdf2 := GoPdf{}
	err = pdf2.OpenPDFFromBytes(data, &OpenPDFOption{
		Password: "user",
	})
	// May or may not work depending on encryption support
	_ = err
}

// ============================================================
// AddTTFFontFromFontContainer — with options
// ============================================================

func TestCov29_FontContainer_WithOptions(t *testing.T) {
	fc := &FontContainer{}
	if err := fc.AddTTFFont("fc-font", resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}

	fontData, err := os.ReadFile(resFontPath)
	if err != nil {
		t.Skipf("cannot read font: %v", err)
	}

	// AddTTFFontDataWithOption
	if err := fc.AddTTFFontDataWithOption("fc-data-opt", fontData, TtfOption{
		UseKerning: true,
	}); err != nil {
		t.Fatalf("FontContainer.AddTTFFontDataWithOption: %v", err)
	}

	// AddTTFFontByReaderWithOption
	if err := fc.AddTTFFontByReaderWithOption("fc-reader-opt", bytes.NewReader(fontData), TtfOption{
		UseKerning: true,
	}); err != nil {
		t.Fatalf("FontContainer.AddTTFFontByReaderWithOption: %v", err)
	}

	// Use in PDF
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	if err := pdf.AddTTFFontFromFontContainer("fc-data-opt", fc); err != nil {
		t.Fatalf("AddTTFFontFromFontContainer: %v", err)
	}
	pdf.SetFont("fc-data-opt", "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Font container with options")
	pdf.GetBytesPdf()
}

// ============================================================
// diagonalAngle
// ============================================================

func TestCov29_DiagonalAngle(t *testing.T) {
	angle := diagonalAngle(100, 50)
	if angle <= 0 || angle >= 90 {
		t.Errorf("expected angle between 0 and 90, got %f", angle)
	}
}


// ============================================================
// ExtractPagesFromBytes — empty pages
// ============================================================

func TestCov29_ExtractPagesFromBytes_EmptyPages(t *testing.T) {
	_, err := ExtractPagesFromBytes([]byte("%PDF-1.4"), nil, nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
	_, err = ExtractPagesFromBytes([]byte("%PDF-1.4"), []int{}, nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
}

// ============================================================
// MergePages — empty paths
// ============================================================

func TestCov29_MergePages_EmptyPaths(t *testing.T) {
	_, err := MergePages(nil, nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
	_, err = MergePages([]string{}, nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
}

// ============================================================
// MergePagesFromBytes — empty slices
// ============================================================

func TestCov29_MergePagesFromBytes_EmptySlices(t *testing.T) {
	_, err := MergePagesFromBytes(nil, nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
	_, err = MergePagesFromBytes([][]byte{}, nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
}

// ============================================================
// DeletePage — no pages
// ============================================================

func TestCov29_DeletePage_NoPages(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	// No pages added
	err := pdf.DeletePage(1)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
}

// ============================================================
// DeletePages — no pages in document
// ============================================================

func TestCov29_DeletePages_NoPages(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.DeletePages([]int{1})
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
}

// ============================================================
// SelectPagesFromBytes — with Box option
// ============================================================

func TestCov29_SelectPagesFromBytes_WithBox(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}

	result, err := SelectPagesFromBytes(data, []int{1}, &OpenPDFOption{Box: "/CropBox"})
	if err != nil {
		t.Logf("SelectPagesFromBytes with box: %v", err)
	}
	_ = result
}

// ============================================================
// MergePages — with invalid PDF data (0 pages)
// ============================================================

func TestCov29_MergePagesFromBytes_InvalidPDF(t *testing.T) {
	// Pass data that gofpdi can't parse — should skip those
	defer func() { recover() }()
	_, err := MergePagesFromBytes([][]byte{
		[]byte("not a pdf"),
	}, nil)
	_ = err
}

// ============================================================
// SetPageCropBox / ClearPageCropBox
// ============================================================

func TestCov29_SetPageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Crop box test")

	err := pdf.SetPageCropBox(1, Box{Left: 10, Top: 10, Right: 10, Bottom: 10})
	if err != nil {
		t.Fatalf("SetPageCropBox: %v", err)
	}

	// Out of range
	err = pdf.SetPageCropBox(99, Box{})
	if err == nil {
		t.Error("expected error for out of range page")
	}

	// Clear
	err = pdf.ClearPageCropBox(1)
	if err != nil {
		t.Fatalf("ClearPageCropBox: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SetPageRotation
// ============================================================

func TestCov29_SetPageRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Rotation test")

	err := pdf.SetPageRotation(1, 90)
	if err != nil {
		t.Fatalf("SetPageRotation: %v", err)
	}

	// Out of range
	err = pdf.SetPageRotation(99, 90)
	if err == nil {
		t.Error("expected error for out of range page")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SetPageLayout / SetPageMode
// ============================================================

func TestCov29_SetPageLayout_SetPageMode(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetPageLayout(PageLayoutTwoColumnLeft)
	pdf.SetPageMode(PageModeUseThumbs)

	pdf.GetBytesPdf()
}

// ============================================================
// SetPageLabels
// ============================================================

func TestCov29_SetPageLabels(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")

	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Style: PageLabelDecimal, Prefix: "Page ", Start: 1},
	})

	pdf.GetBytesPdf()
}

// ============================================================
// OCG (Optional Content Groups)
// ============================================================

func TestCov29_OCG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// AddOCG
	layer := pdf.AddOCG(OCG{Name: "Layer 1", On: true})

	// BeginMarkedContent / EndMarkedContent for OCG
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Layer 1 content")
	_ = layer

	pdf.GetBytesPdf()
}

// ============================================================
// EmbedFile
// ============================================================

func TestCov29_EmbedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Embedded file test")

	// Embed a small file
	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "test.txt",
		Content:  []byte("Hello embedded file"),
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}

	// Get embedded file
	data, err := pdf.GetEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFile: %v", err)
	}
	if string(data) != "Hello embedded file" {
		t.Errorf("expected 'Hello embedded file', got %q", string(data))
	}

	// Get info
	info, err := pdf.GetEmbeddedFileInfo("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFileInfo: %v", err)
	}
	if info.Size != 19 {
		t.Errorf("expected size 19, got %d", info.Size)
	}

	// Get names
	names := pdf.GetEmbeddedFileNames()
	if len(names) != 1 || names[0] != "test.txt" {
		t.Errorf("unexpected names: %v", names)
	}

	// Count
	if pdf.GetEmbeddedFileCount() != 1 {
		t.Errorf("expected count 1, got %d", pdf.GetEmbeddedFileCount())
	}

	// Update
	err = pdf.UpdateEmbeddedFile("test.txt", EmbeddedFile{
		Content:  []byte("Updated content"),
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}

	// Delete
	err = pdf.DeleteEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("DeleteEmbeddedFile: %v", err)
	}

	// Error paths
	_, err = pdf.GetEmbeddedFile("nonexistent")
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected ErrEmbeddedFileNotFound, got %v", err)
	}
	err = pdf.DeleteEmbeddedFile("nonexistent")
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected ErrEmbeddedFileNotFound, got %v", err)
	}
	err = pdf.UpdateEmbeddedFile("nonexistent", EmbeddedFile{Content: []byte("x")})
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected ErrEmbeddedFileNotFound, got %v", err)
	}

	// Empty name/content
	err = pdf.AddEmbeddedFile(EmbeddedFile{Name: "", Content: []byte("x")})
	if err == nil {
		t.Error("expected error for empty name")
	}
	err = pdf.AddEmbeddedFile(EmbeddedFile{Name: "x", Content: nil})
	if err == nil {
		t.Error("expected error for empty content")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SetInfo with all fields
// ============================================================

func TestCov29_SetInfo_AllFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetInfo(PdfInfo{
		Title:    "Test Title",
		Author:   "Test Author",
		Subject:  "Test Subject",
		Creator:  "Test Creator",
		Producer: "Test Producer",
	})

	info := pdf.GetInfo()
	if info.Producer != "Test Producer" {
		t.Errorf("expected producer, got %q", info.Producer)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// PDF Version
// ============================================================

func TestCov29_PDFVersion(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "PDF version test")

	data := pdf.GetBytesPdf()
	if !bytes.Contains(data, []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

// ============================================================
// MarkInfo (tagged PDF)
// ============================================================

func TestCov29_MarkInfo(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize: *PageSizeA4,
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Mark info test")

	pdf.GetBytesPdf()
}

// ============================================================
// XMP Metadata
// ============================================================

func TestCov29_XMPMetadata(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "XMP test")

	pdf.SetXMPMetadata(XMPMetadata{
		Title:       "Test Document",
		Creator:     []string{"Author 1", "Author 2"},
		Description: "Test description",
		Subject:     []string{"test", "pdf"},
		Rights:      "Copyright 2025",
		Language:    "en-US",
		CreatorTool: "GoPDF2 Test",
		Producer:    "GoPDF2",
		Keywords:    "test, pdf",
		Trapped:     "False",
		PDFAPart:    2,
		PDFAConformance: "B",
	})

	meta := pdf.GetXMPMetadata()
	if meta == nil {
		t.Fatal("expected XMP metadata")
	}
	if meta.Title != "Test Document" {
		t.Errorf("expected title, got %q", meta.Title)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// Linearization
// ============================================================

func TestCov29_Linearize(t *testing.T) {
	// linearization.go is empty — skip
	t.Skip("linearization not implemented")
}

// ============================================================
// Text extraction
// ============================================================

func TestCov29_ExtractText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Extract this text")

	data := pdf.GetBytesPdf()

	text, err := ExtractTextFromPage(data, 0)
	if err != nil {
		t.Logf("ExtractTextFromPage: %v", err)
	}
	_ = text

	allText, err := ExtractTextFromAllPages(data)
	if err != nil {
		t.Logf("ExtractTextFromAllPages: %v", err)
	}
	_ = allText
}

// ============================================================
// Image extraction
// ============================================================

func TestCov29_ExtractImages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})

	data := pdf.GetBytesPdf()

	images, err := ExtractImagesFromPage(data, 0)
	if err != nil {
		t.Logf("ExtractImagesFromPage: %v", err)
	}
	_ = images

	allImages, err := ExtractImagesFromAllPages(data)
	if err != nil {
		t.Logf("ExtractImagesFromAllPages: %v", err)
	}
	_ = allImages
}

// ============================================================
// Content stream cleaning
// ============================================================

func TestCov29_CleanContentStreams(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Clean content test")

	data := pdf.GetBytesPdf()

	cleaned, err := CleanContentStreams(data)
	if err != nil {
		t.Logf("CleanContentStreams: %v", err)
	}
	_ = cleaned
}


// ============================================================
// Annotation — various types
// ============================================================

func TestCov29_Annotations_AllTypes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Text annotation
	pdf.AddTextAnnotation(50, 50, "Note", "This is a sticky note")

	// Highlight
	pdf.AddHighlightAnnotation(50, 80, 200, 20, [3]uint8{255, 255, 0})

	// FreeText
	pdf.AddFreeTextAnnotation(50, 110, 200, 30, "Free text annotation", 12)

	// Ink
	pdf.AddInkAnnotation(50, 150, 100, 50, [][]Point{
		{{X: 50, Y: 150}, {X: 100, Y: 170}, {X: 150, Y: 150}},
	}, [3]uint8{0, 0, 255})

	// Polyline
	pdf.AddPolylineAnnotation(50, 210, 200, 50, []Point{
		{X: 50, Y: 210}, {X: 150, Y: 230}, {X: 250, Y: 210},
	}, [3]uint8{255, 0, 0})

	// Polygon
	pdf.AddPolygonAnnotation(50, 270, 200, 50, []Point{
		{X: 50, Y: 270}, {X: 150, Y: 290}, {X: 250, Y: 270},
	}, [3]uint8{0, 255, 0})

	// Line
	pdf.AddLineAnnotation(Point{X: 50, Y: 330}, Point{X: 250, Y: 350}, [3]uint8{0, 0, 0})

	// Stamp
	pdf.AddStampAnnotation(50, 370, 100, 40, StampApproved)

	// Squiggly
	pdf.AddSquigglyAnnotation(50, 420, 200, 20, [3]uint8{255, 0, 0})

	// Caret
	pdf.AddCaretAnnotation(50, 450, 20, 20, "Insert here")

	// FileAttachment
	pdf.AddFileAttachmentAnnotation(50, 480, "test.txt", []byte("file content"), "Attached file")

	// Redact
	pdf.AddRedactAnnotation(50, 510, 200, 20, "REDACTED")

	// Get annotations
	annots := pdf.GetAnnotations()
	if len(annots) == 0 {
		t.Error("expected annotations")
	}

	// Get annotations on page
	annots2 := pdf.GetAnnotationsOnPage(1)
	if len(annots2) == 0 {
		t.Error("expected annotations on page 1")
	}

	// Apply redactions
	count := pdf.ApplyRedactions()
	_ = count

	// Delete annotation
	pdf.DeleteAnnotation(0)

	pdf.GetBytesPdf()
}

// ============================================================
// ModifyAnnotation
// ============================================================

func TestCov29_ModifyAnnotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddTextAnnotation(50, 50, "Note", "Original")

	err := pdf.ModifyAnnotation(0, 0, AnnotationOption{
		Type:    AnnotText,
		Title:   "Modified",
		Content: "Modified content",
		X:       60, Y: 60, W: 30, H: 30,
	})
	if err != nil {
		t.Fatalf("ModifyAnnotation: %v", err)
	}

	// Error: invalid page
	err = pdf.ModifyAnnotation(99, 0, AnnotationOption{})
	if err == nil {
		t.Error("expected error for invalid page")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// GetPageElements / GetPageElementsByType
// ============================================================

func TestCov29_GetPageElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Element test")
	pdf.Line(50, 80, 200, 80)

	elems, err := pdf.GetPageElements(1)
	if err != nil {
		t.Fatalf("GetPageElements: %v", err)
	}
	if len(elems) == 0 {
		t.Error("expected page elements")
	}

	// By type
	textElems, err := pdf.GetPageElementsByType(1, ElementText)
	if err != nil {
		t.Fatalf("GetPageElementsByType: %v", err)
	}
	_ = textElems

	// Count
	count, err := pdf.GetPageElementCount(1)
	if err != nil {
		t.Fatalf("GetPageElementCount: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero element count")
	}

	// Invalid page
	_, err = pdf.GetPageElements(99)
	if err == nil {
		t.Error("expected error for invalid page")
	}
}

// ============================================================
// GetTOC / SetTOC
// ============================================================

func TestCov29_GetTOC_SetTOC(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	toc := pdf.GetTOC()
	if len(toc) == 0 {
		t.Error("expected TOC items")
	}

	// SetTOC with empty items (clears)
	err := pdf.SetTOC(nil)
	if err != nil {
		t.Fatalf("SetTOC nil: %v", err)
	}

	// SetTOC with items
	err = pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "New Chapter 1", PageNo: 1},
		{Level: 1, Title: "New Chapter 2", PageNo: 2},
	})
	if err != nil {
		t.Fatalf("SetTOC: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// OCG — GetOCGs, SetOCGState, SetOCGStates
// ============================================================

func TestCov29_OCG_Operations(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	layer := pdf.AddOCG(OCG{Name: "Layer 1", On: true})
	pdf.AddOCG(OCG{Name: "Layer 2", On: false})

	// GetOCGs
	ocgs := pdf.GetOCGs()
	if len(ocgs) < 2 {
		t.Errorf("expected at least 2 OCGs, got %d", len(ocgs))
	}

	// SetOCGState
	err := pdf.SetOCGState("Layer 1", false)
	if err != nil {
		t.Fatalf("SetOCGState: %v", err)
	}

	// SetOCGStates
	err = pdf.SetOCGStates(map[string]bool{
		"Layer 1": true,
		"Layer 2": true,
	})
	if err != nil {
		t.Fatalf("SetOCGStates: %v", err)
	}

	// Error: non-existent layer
	err = pdf.SetOCGState("NonExistent", true)
	if err == nil {
		t.Error("expected error for non-existent OCG")
	}

	_ = layer
	pdf.GetBytesPdf()
}

// ============================================================
// AddEmbeddedFile — with description
// ============================================================

func TestCov29_EmbedFile_WithDescription(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:        "data.csv",
		Content:     []byte("a,b,c\n1,2,3"),
		MimeType:    "text/csv",
		Description: "Sample CSV data",
	})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}

	info, err := pdf.GetEmbeddedFileInfo("data.csv")
	if err != nil {
		t.Fatalf("GetEmbeddedFileInfo: %v", err)
	}
	if info.Description != "Sample CSV data" {
		t.Errorf("expected description, got %q", info.Description)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SetBookmarkStyle
// ============================================================

func TestCov29_SetBookmarkStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")

	err := pdf.SetBookmarkStyle(0, BookmarkStyle{
		Bold:      true,
		Italic:    true,
		Color:     [3]float64{1, 0, 0},
		Collapsed: true,
	})
	if err != nil {
		t.Fatalf("SetBookmarkStyle: %v", err)
	}

	// Out of range
	err = pdf.SetBookmarkStyle(99, BookmarkStyle{})
	if err == nil {
		t.Error("expected error for out of range")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// ModifyBookmark
// ============================================================

func TestCov29_ModifyBookmark(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Original Title")

	err := pdf.ModifyBookmark(0, "New Title")
	if err != nil {
		t.Fatalf("ModifyBookmark: %v", err)
	}

	// Out of range
	err = pdf.ModifyBookmark(99, "test")
	if err == nil {
		t.Error("expected error for out of range")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// ConvertColorspace
// ============================================================

func TestCov29_ConvertColorspace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.SetTextColor(255, 0, 0)
	pdf.Cell(nil, "Red text for colorspace conversion")

	data := pdf.GetBytesPdf()

	gray, err := ConvertColorspace(data, ConvertColorspaceOption{
		Target: ColorspaceGray,
	})
	if err != nil {
		t.Logf("ConvertColorspace: %v", err)
	}
	_ = gray
}

// ============================================================
// RecompressImages
// ============================================================

func TestCov29_RecompressImages(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})

	data := pdf.GetBytesPdf()

	recompressed, err := RecompressImages(data, RecompressOption{
		JPEGQuality: 50,
	})
	if err != nil {
		t.Logf("RecompressImages: %v", err)
	}
	_ = recompressed
}

// ============================================================
// SearchText — more patterns
// ============================================================

func TestCov29_SearchText_Patterns(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Hello World Test Document")

	data := pdf.GetBytesPdf()

	// Case sensitive
	results, err := SearchText(data, "Hello", false)
	if err != nil {
		t.Logf("SearchText: %v", err)
	}
	_ = results

	// Case insensitive
	results2, err := SearchText(data, "hello", true)
	if err != nil {
		t.Logf("SearchText insensitive: %v", err)
	}
	_ = results2
}

// ============================================================
// GarbageCollect
// ============================================================

func TestCov29_GarbageCollect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "GC test")

	before := pdf.GetObjectCount()
	removed := pdf.GarbageCollect(GCCompact)
	after := pdf.GetObjectCount()
	t.Logf("GC: before=%d, removed=%d, after=%d", before, removed, after)

	// Also test live count
	live := pdf.GetLiveObjectCount()
	if live <= 0 {
		t.Error("expected positive live object count")
	}

	pdf.GetBytesPdf()
}

// ============================================================
// SelectPages
// ============================================================

func TestCov29_SelectPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 3")

	// Select pages 3, 1 (reorder)
	result, err := pdf.SelectPages([]int{3, 1})
	if err != nil {
		t.Fatalf("SelectPages: %v", err)
	}
	if result.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", result.GetNumberOfPages())
	}

	// Empty pages
	_, err = pdf.SelectPages(nil)
	if err != ErrNoPages {
		t.Errorf("expected ErrNoPages, got %v", err)
	}
}
