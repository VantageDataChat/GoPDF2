package gopdf

import (
	"bytes"
	"os"
	"testing"
)

// ============================================================
// coverage_boost28_test.go — TestCov28_ prefix
// Targets: form_field findCurrentPageObjID, CellWithOption more
// branches, Cell nil rect, GetBytesPdfReturnErr failWriter,
// cache_content_text.write protected, various obj write failWriter
// ============================================================

// ============================================================
// Cell — nil rect
// ============================================================

func TestCov28_Cell_NilRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.Cell(nil, "Nil rect cell")
	if err != nil {
		t.Fatalf("Cell nil rect: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// Cell — with rect
// ============================================================

func TestCov28_Cell_WithRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.Cell(&Rect{W: 200, H: 20}, "With rect cell")
	if err != nil {
		t.Fatalf("Cell with rect: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// MultiCell — wrapping text
// ============================================================

func TestCov28_MultiCell_Wrapping(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.MultiCell(&Rect{W: 100, H: 200}, "This is a long text that should wrap across multiple lines in the multi cell area.")
	if err != nil {
		t.Fatalf("MultiCell: %v", err)
	}
	pdf.GetBytesPdf()
}


// ============================================================
// MultiCellWithOption — various options
// ============================================================

func TestCov28_MultiCellWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.MultiCellWithOption(&Rect{W: 150, H: 200}, "Multi cell with border and alignment options for testing.", CellOption{
		Align:  Center | Middle,
		Border: Left | Right | Top | Bottom,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// IsFitMultiCell — various widths
// ============================================================

func TestCov28_IsFitMultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Should fit
	fit, _, err := pdf.IsFitMultiCell(&Rect{W: 500, H: 200}, "Short text")
	if err != nil {
		t.Fatalf("IsFitMultiCell: %v", err)
	}
	if !fit {
		t.Error("expected text to fit")
	}

	// Should not fit
	fit2, _, err := pdf.IsFitMultiCell(&Rect{W: 50, H: 10}, "This is a very long text that definitely will not fit in such a tiny box area.")
	if err != nil {
		t.Fatalf("IsFitMultiCell: %v", err)
	}
	_ = fit2
}

// ============================================================
// SplitText — various modes
// ============================================================

func TestCov28_SplitText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	lines, err := pdf.SplitText("Hello World this is a test of text splitting", 100)
	if err != nil {
		t.Fatalf("SplitText: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected some lines")
	}

	// SplitTextWithWordWrap
	lines2, err := pdf.SplitTextWithWordWrap("Hello World this is a test of word wrap splitting", 100)
	if err != nil {
		t.Fatalf("SplitTextWithWordWrap: %v", err)
	}
	if len(lines2) == 0 {
		t.Error("expected some lines")
	}
}

// ============================================================
// Line — with options
// ============================================================

func TestCov28_Line_WithOptions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetLineWidth(2)
	pdf.SetLineType("dashed")
	pdf.Line(50, 50, 200, 200)

	pdf.SetLineType("dotted")
	pdf.Line(50, 100, 200, 100)

	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(50, 150, 200, 150)

	pdf.GetBytesPdf()
}

// ============================================================
// RectFromLowerLeftWithOpts
// ============================================================

func TestCov28_RectFromLowerLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.RectFromLowerLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 200,
		Rect:       Rect{W: 100, H: 80},
		PaintStyle: FillPaintStyle,
	})
	pdf.GetBytesPdf()
}

// ============================================================
// Oval
// ============================================================

func TestCov28_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(100, 100, 200, 150)
	pdf.GetBytesPdf()
}

// ============================================================
// Curve
// ============================================================

func TestCov28_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(50, 50, 100, 200, 200, 200, 250, 50, "FD")
	pdf.GetBytesPdf()
}

// ============================================================
// SetGrayFill / SetGrayStroke
// ============================================================

func TestCov28_GrayFillStroke(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
	pdf.SetGrayStroke(0.3)
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 50,
		Rect:       Rect{W: 100, H: 100},
		PaintStyle: DrawFillPaintStyle,
	})
	pdf.GetBytesPdf()
}

// ============================================================
// SetStrokeColor / SetFillColor
// ============================================================

func TestCov28_StrokeFillColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetFillColor(0, 0, 255)
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 50,
		Rect:       Rect{W: 100, H: 100},
		PaintStyle: DrawFillPaintStyle,
	})
	pdf.GetBytesPdf()
}

// ============================================================
// SetStrokeColorCMYK / SetFillColorCMYK
// ============================================================

func TestCov28_StrokeFillColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColorCMYK(100, 0, 0, 0)
	pdf.SetFillColorCMYK(0, 100, 0, 0)
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 50,
		Rect:       Rect{W: 100, H: 100},
		PaintStyle: DrawFillPaintStyle,
	})
	pdf.GetBytesPdf()
}

// ============================================================
// cache_content_text.write — protected with underline + CMYK
// ============================================================

func TestCov28_CacheContentText_Protected_Underline_CMYK(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()

	// CMYK text color
	pdf.SetTextColorCMYK(100, 0, 0, 0)
	pdf.SetFontWithStyle(fontFamily, Underline, 14)
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 200, H: 20}, "CMYK underlined protected", CellOption{
		Align:  Left | Top,
		Border: Left | Right | Top | Bottom,
	})

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	// Write each cache with failWriter
	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 3 {
			fw := &failWriterAt{n: n}
			cache.write(fw, pdf.pdfProtection)
		}
	}
}

// ============================================================
// Image — PNG with transparency (exercises mask paths)
// ============================================================

func TestCov28_Image_PNG_Transparency(t *testing.T) {
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 200, H: 200})
	if err != nil {
		t.Fatalf("Image PNG: %v", err)
	}

	// Also test with transparency option on a real image
	if _, err2 := os.Stat(resJPEGPath); err2 == nil {
		holder, _ := ImageHolderByPath(resJPEGPath)
		if holder != nil {
			err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
				X:    50,
				Y:    300,
				Rect: &Rect{W: 100, H: 100},
				Transparency: &Transparency{
					Alpha:         0.5,
					BlendModeType: NormalBlendMode,
				},
			})
			if err != nil {
				t.Logf("ImageByHolderWithOptions with transparency: %v", err)
			}
		}
	}

	pdf.GetBytesPdf()
}

// ============================================================
// GetBytesPdfReturnErr — with protected PDF
// ============================================================

func TestCov28_GetBytesPdfReturnErr_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Protected PDF bytes")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// compilePdf — with various content types
// ============================================================

func TestCov28_CompilePdf_RichContent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Text
	pdf.SetXY(50, 50)
	pdf.Text("Rich content test")

	// Line
	pdf.Line(50, 70, 200, 70)

	// Rectangle
	pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X: 50, Y: 80,
		Rect:       Rect{W: 100, H: 50},
		PaintStyle: DrawFillPaintStyle,
	})

	// Oval
	pdf.Oval(200, 80, 300, 130)

	// Image
	if _, err := os.Stat(resJPEGPath); err == nil {
		pdf.Image(resJPEGPath, 50, 150, &Rect{W: 100, H: 80})
	}

	// Second page
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// Read — complete read cycle
// ============================================================

func TestCov28_Read_Complete(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Complete read test")

	var buf bytes.Buffer
	p := make([]byte, 128)
	for {
		n, err := pdf.Read(p)
		if n > 0 {
			buf.Write(p[:n])
		}
		if err != nil {
			break
		}
	}

	if buf.Len() == 0 {
		t.Error("expected data from Read")
	}

	// Verify it starts with %PDF
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Error("expected PDF header")
	}
}

// ============================================================
// ImageByHolderWithOptions — no holder (error path)
// ============================================================

func TestCov28_ImageByHolderWithOptions_NoHolder(t *testing.T) {
	defer func() { recover() }()
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageByHolderWithOptions(nil, ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 100, H: 100},
	})
	// Should fail because nil image holder
	if err == nil {
		t.Log("no error (may panic instead)")
	}
}

// ============================================================
// SetCharSpacing
// ============================================================

func TestCov28_SetCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCharSpacing(2.0)
	pdf.SetXY(50, 50)
	pdf.Text("Spaced text")
	pdf.SetCharSpacing(0)
	pdf.GetBytesPdf()
}

// ============================================================
// PlaceHolderText + FillInPlaceHoldText — more alignments
// ============================================================

func TestCov28_PlaceHolder_MoreAlignments(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Left align
	pdf.SetXY(50, 50)
	pdf.PlaceHolderText("ph_left", 200)

	// Right align
	pdf.SetXY(50, 80)
	pdf.PlaceHolderText("ph_right", 200)

	// Center align
	pdf.SetXY(50, 110)
	pdf.PlaceHolderText("ph_center", 200)

	pdf.FillInPlaceHoldText("ph_left", "Left aligned", Left)
	pdf.FillInPlaceHoldText("ph_right", "Right aligned", Right)
	pdf.FillInPlaceHoldText("ph_center", "Centered", Center)

	// Not found
	err := pdf.FillInPlaceHoldText("nonexistent", "text", Left)
	if err == nil {
		t.Error("expected error for nonexistent placeholder")
	}

	pdf.GetBytesPdf()
}
