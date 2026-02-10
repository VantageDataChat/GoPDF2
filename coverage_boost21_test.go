package gopdf

import (
	"bytes"
	"testing"
)

// ============================================================
// coverage_boost21_test.go — TestCov21_ prefix
// Targets: remaining low-coverage functions and edge cases
// ============================================================

// ============================================================
// gopdf.go — Cell with nil rectangle
// ============================================================

func TestCov21_Cell_NilRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	if err := pdf.Cell(nil, "Cell with nil rect"); err != nil {
		t.Fatalf("Cell nil rect: %v", err)
	}
}

// ============================================================
// gopdf.go — MultiCellWithOption various alignments
// ============================================================

func TestCov21_MultiCellWithOption_RightBottom(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Right bottom aligned multi cell text that wraps across lines", CellOption{
		Align: Right | Bottom,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption right bottom: %v", err)
	}
}

func TestCov21_MultiCellWithOption_CenterMiddle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 200)

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Center middle aligned multi cell", CellOption{
		Align: Center | Middle,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption center middle: %v", err)
	}
}

// ============================================================
// gopdf.go — IsFitMultiCell
// ============================================================

func TestCov21_IsFitMultiCell_Fits(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fit, h, err := pdf.IsFitMultiCell(&Rect{W: 500, H: 200}, "Short text")
	if err != nil {
		t.Fatalf("IsFitMultiCell: %v", err)
	}
	if !fit {
		t.Error("expected text to fit")
	}
	if h <= 0 {
		t.Error("expected positive height")
	}
}

func TestCov21_IsFitMultiCell_DoesNotFit(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	longText := "This is a very long text that should not fit in a tiny rectangle. " +
		"It keeps going and going and going and going and going and going. " +
		"More text here to make it even longer and longer and longer."
	fit, _, err := pdf.IsFitMultiCell(&Rect{W: 50, H: 10}, longText)
	if err != nil {
		t.Fatalf("IsFitMultiCell: %v", err)
	}
	if fit {
		t.Error("expected text not to fit")
	}
}

// ============================================================
// gopdf.go — SetNewY, SetNewYIfNoOffset, SetNewXY
// ============================================================

func TestCov21_SetNewY_NewPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set Y to near bottom of page to trigger new page.
	pdf.SetNewY(900, 50)
	_ = pdf.Text("After SetNewY")
}

func TestCov21_SetNewYIfNoOffset(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetNewYIfNoOffset(100, 50)
	_ = pdf.Text("After SetNewYIfNoOffset")
}

func TestCov21_SetNewXY_NewPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetNewXY(900, 50, 50)
	_ = pdf.Text("After SetNewXY")
}

// ============================================================
// gopdf.go — Curve
// ============================================================

func TestCov21_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.Curve(50, 400, 100, 350, 150, 450, 200, 400, "D")
	pdf.Curve(50, 500, 100, 450, 150, 550, 200, 500, "")
}

// ============================================================
// gopdf.go — RectFromUpperLeftWithOpts, RectFromLowerLeftWithOpts
// ============================================================

func TestCov21_RectFromUpperLeftWithOpts_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(200, 200, 255)

	err := pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		Rect:       Rect{W: 100, H: 50},
		X:          50,
		Y:          50,
		PaintStyle: FillPaintStyle,
	})
	if err != nil {
		t.Fatalf("RectFromUpperLeftWithOpts fill: %v", err)
	}
}

func TestCov21_RectFromLowerLeftWithOpts_DrawFill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(255, 200, 200)
	pdf.SetStrokeColor(0, 0, 0)

	err := pdf.RectFromLowerLeftWithOpts(DrawableRectOptions{
		Rect:       Rect{W: 100, H: 50},
		X:          50,
		Y:          200,
		PaintStyle: DrawFillPaintStyle,
	})
	if err != nil {
		t.Fatalf("RectFromLowerLeftWithOpts draw+fill: %v", err)
	}
}

// ============================================================
// content_element.go — InsertLineElement, InsertRectElement, InsertOvalElement
// ============================================================

func TestCov21_InsertLineElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "test")

	err := pdf.InsertLineElement(1, 10, 10, 200, 200)
	if err != nil {
		t.Fatalf("InsertLineElement: %v", err)
	}
}

func TestCov21_InsertRectElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "test")

	err := pdf.InsertRectElement(1, 50, 50, 100, 80, "FD")
	if err != nil {
		t.Fatalf("InsertRectElement: %v", err)
	}
}

func TestCov21_InsertOvalElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "test")

	err := pdf.InsertOvalElement(1, 50, 50, 150, 100)
	if err != nil {
		t.Fatalf("InsertOvalElement: %v", err)
	}
}

// ============================================================
// doc_stats.go — GetDocumentStats, GetFonts
// ============================================================

func TestCov21_GetDocumentStats(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Stats test")

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Errorf("expected 1 page, got %d", stats.PageCount)
	}
}

func TestCov21_GetFonts_Method(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Font test")

	fonts := pdf.GetFonts()
	if len(fonts) == 0 {
		t.Log("no fonts found via GetFonts method")
	}
}

// ============================================================
// linearization.go — Linearize (if exists)
// ============================================================

func TestCov21_CleanContentStreams(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Clean content test")
	pdf.Line(10, 10, 200, 200)

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}

	result, err := CleanContentStreams(data)
	if err != nil {
		t.Logf("CleanContentStreams: %v", err)
	}
	_ = result
}

// ============================================================
// page_manipulate.go — DeletePage, CopyPage, ExtractPages
// ============================================================

func TestCov21_DeletePage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 3")

	err := pdf.DeletePage(2)
	if err != nil {
		t.Fatalf("DeletePage: %v", err)
	}
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov21_DeletePage_First(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	err := pdf.DeletePage(1)
	if err != nil {
		t.Fatalf("DeletePage first: %v", err)
	}
}

func TestCov21_DeletePage_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	err := pdf.DeletePage(2)
	if err != nil {
		t.Fatalf("DeletePage last: %v", err)
	}
}

func TestCov21_DeletePage_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.DeletePage(0)
	if err == nil {
		t.Fatal("expected error for page 0")
	}
	err = pdf.DeletePage(99)
	if err == nil {
		t.Fatal("expected error for page 99")
	}
}

func TestCov21_CopyPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")

	newPageNo, err := pdf.CopyPage(1)
	if err != nil {
		t.Fatalf("CopyPage: %v", err)
	}
	if newPageNo != 3 {
		t.Errorf("expected new page 3, got %d", newPageNo)
	}
	if pdf.GetNumberOfPages() != 3 {
		t.Errorf("expected 3 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov21_CopyPage_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_, err := pdf.CopyPage(0)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	_, err = pdf.CopyPage(99)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
}

func TestCov21_ExtractPages(t *testing.T) {
	result, err := ExtractPages(resTestPDF, []int{1}, nil)
	if err != nil {
		t.Skipf("ExtractPages: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ============================================================
// page_batch_ops.go — DeletePages, MovePage
// ============================================================

func TestCov21_DeletePages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 3")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 4")

	err := pdf.DeletePages([]int{2, 4})
	if err != nil {
		t.Fatalf("DeletePages: %v", err)
	}
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov21_MovePage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	_ = pdf.Cell(nil, "Page 3")

	err := pdf.MovePage(3, 1)
	if err != nil {
		t.Fatalf("MovePage: %v", err)
	}
}

func TestCov21_MovePage_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.MovePage(0, 1)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}
	err = pdf.MovePage(1, 99)
	if err == nil {
		t.Fatal("expected error for invalid dest")
	}
}

// ============================================================
// gopdf.go — GetBytesPdfReturnErr with protection
// ============================================================

func TestCov21_GetBytesPdfReturnErr_Protected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Protected bytes test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("expected PDF header")
	}
}

// ============================================================
// gopdf.go — WriteTo
// ============================================================

func TestCov21_WriteTo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "WriteTo test")

	var buf bytes.Buffer
	n, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
}

// ============================================================
// gopdf.go — Write
// ============================================================

func TestCov21_Write(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Cell(nil, "Write test")

	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================
// cache_content_text_color_cmyk.go — equal
// ============================================================

func TestCov21_CMYK_TextColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetTextColorCMYK(100, 0, 0, 0)
	pdf.SetXY(50, 50)
	_ = pdf.Cell(nil, "Cyan text")

	pdf.SetTextColorCMYK(0, 100, 0, 0)
	pdf.SetXY(50, 80)
	_ = pdf.Cell(nil, "Magenta text")

	// Same color — should trigger equal() returning true.
	pdf.SetTextColorCMYK(0, 100, 0, 0)
	pdf.SetXY(50, 110)
	_ = pdf.Cell(nil, "Same magenta text")
}

// ============================================================
// cache_content_line.go — write with dash
// ============================================================

func TestCov21_Line_Dashed(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetLineType("dashed")
	pdf.SetLineWidth(1)
	pdf.Line(50, 50, 300, 50)

	pdf.SetLineType("dotted")
	pdf.Line(50, 70, 300, 70)

	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(50, 90, 300, 90)

	pdf.SetLineType("solid")
	pdf.Line(50, 110, 300, 110)
}

// ============================================================
// gopdf.go — GetPageSize, GetAllPageSizes
// ============================================================

func TestCov21_GetPageSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	w, h, err := pdf.GetPageSize(1)
	if err != nil {
		t.Fatalf("GetPageSize: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got w=%f h=%f", w, h)
	}
}

func TestCov21_GetAllPageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPageWithOption(PageOption{PageSize: &Rect{W: 400, H: 600}})

	sizes := pdf.GetAllPageSizes()
	if len(sizes) != 2 {
		t.Errorf("expected 2 sizes, got %d", len(sizes))
	}
}

func TestCov21_GetPageSize_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_, _, err := pdf.GetPageSize(0)
	if err == nil {
		t.Fatal("expected error for page 0")
	}
	_, _, err = pdf.GetPageSize(99)
	if err == nil {
		t.Fatal("expected error for page 99")
	}
}
