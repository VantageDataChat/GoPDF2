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
// coverage_boost18_test.go — TestCov18_ prefix
// Targets: gopdf.go IsCurrFontContainGlyph, ImageByHolderWithOptions
// with transparency and mask, GetEmbeddedFile, FillInPlaceHoldText,
// IsFitMultiCellWithNewline, SplitTextWithOption
// ============================================================

// ============================================================
// gopdf.go — IsCurrFontContainGlyph
// ============================================================

func TestCov18_IsCurrFontContainGlyph_ASCII(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph: %v", err)
	}
	if !ok {
		t.Error("expected 'A' to be in font")
	}
}

func TestCov18_IsCurrFontContainGlyph_Missing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Try a rare Unicode character that's unlikely to be in the font.
	ok, err := pdf.IsCurrFontContainGlyph('\U0001F600') // emoji
	if err != nil {
		t.Logf("IsCurrFontContainGlyph emoji: %v", err)
	}
	_ = ok
}

func TestCov18_IsCurrFontContainGlyph_NoFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()

	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph no font: %v", err)
	}
	if ok {
		t.Error("expected false when no font set")
	}
}

// ============================================================
// gopdf.go — ImageByHolderWithOptions with transparency
// ============================================================

func TestCov18_ImageByHolderWithOptions_Transparency(t *testing.T) {
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
		Transparency: &Transparency{
			Alpha:         0.5,
			BlendModeType: NormalBlendMode,
		},
	})
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions with transparency: %v", err)
	}
}

func TestCov18_ImageByHolderWithOptions_Mask(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}

	// Create a mask image.
	maskImg := image.NewGray(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			maskImg.Set(x, y, color.Gray{Y: uint8(x * 5)})
		}
	}
	var maskBuf bytes.Buffer
	_ = png.Encode(&maskBuf, maskImg)

	maskHolder, err := ImageHolderByBytes(maskBuf.Bytes())
	if err != nil {
		t.Fatalf("ImageHolderByBytes mask: %v", err)
	}

	pdf := newPDFWithFont(t)
	pdf.AddPage()

	mainHolder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Fatalf("ImageHolderByPath: %v", err)
	}

	err = pdf.ImageByHolderWithOptions(mainHolder, ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 200, H: 200},
		Mask: &MaskOptions{
			Holder: maskHolder,
			ImageOptions: ImageOptions{
				X:    50,
				Y:    50,
				Rect: &Rect{W: 200, H: 200},
			},
		},
	})
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions with mask: %v", err)
	}
}

// ============================================================
// embedded_file.go — GetEmbeddedFile
// ============================================================

func TestCov18_GetEmbeddedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("text")

	_ = pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "test.txt",
		Content:  []byte("Hello World"),
		MimeType: "text/plain",
	})

	data, err := pdf.GetEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestCov18_GetEmbeddedFile_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_, err := pdf.GetEmbeddedFile("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

// ============================================================
// gopdf.go — PlaceHolderText + FillInPlaceHoldText
// ============================================================

func TestCov18_FillInPlaceHoldText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.PlaceHolderText("name", 200)
	if err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	err = pdf.FillInPlaceHoldText("name", "John Doe", Left)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText: %v", err)
	}
}

func TestCov18_FillInPlaceHoldText_Center(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	_ = pdf.PlaceHolderText("title", 300)
	err := pdf.FillInPlaceHoldText("title", "Centered Title", Center)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText center: %v", err)
	}
}

func TestCov18_FillInPlaceHoldText_Right(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	_ = pdf.PlaceHolderText("amount", 150)
	err := pdf.FillInPlaceHoldText("amount", "$1,234.56", Right)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText right: %v", err)
	}
}

func TestCov18_FillInPlaceHoldText_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.FillInPlaceHoldText("nonexistent", "text", Left)
	if err == nil {
		t.Fatal("expected error for non-existent placeholder")
	}
}

// ============================================================
// gopdf.go — IsFitMultiCellWithNewline
// ============================================================

func TestCov18_IsFitMultiCellWithNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fit, _, err := pdf.IsFitMultiCellWithNewline(&Rect{W: 200, H: 100}, "Line 1\nLine 2\nLine 3")
	if err != nil {
		t.Fatalf("IsFitMultiCellWithNewline: %v", err)
	}
	_ = fit
}

func TestCov18_IsFitMultiCellWithNewline_Overflow(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	text := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10"
	fit, _, err := pdf.IsFitMultiCellWithNewline(&Rect{W: 50, H: 20}, text)
	if err != nil {
		t.Fatalf("IsFitMultiCellWithNewline overflow: %v", err)
	}
	if fit {
		t.Error("expected text not to fit")
	}
}

// ============================================================
// gopdf.go — SplitTextWithOption
// ============================================================

func TestCov18_SplitTextWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	splits, err := pdf.SplitTextWithOption("Hello World this is a long text that should be split", 100, nil)
	if err != nil {
		t.Fatalf("SplitTextWithOption: %v", err)
	}
	if len(splits) == 0 {
		t.Error("expected at least one split")
	}
}

func TestCov18_SplitTextWithOption_WithBreakOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	splits, err := pdf.SplitTextWithOption("Hello World this is a long text", 100, &DefaultBreakOption)
	if err != nil {
		t.Fatalf("SplitTextWithOption with break option: %v", err)
	}
	if len(splits) == 0 {
		t.Error("expected at least one split")
	}
}

// ============================================================
// gopdf.go — SetPage with multiple pages
// ============================================================

func TestCov18_SetPage_Multiple(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 3")

	// Navigate back to page 1.
	err := pdf.SetPage(1)
	if err != nil {
		t.Fatalf("SetPage(1): %v", err)
	}
	pdf.SetXY(50, 100)
	_ = pdf.Text("Added to page 1")

	// Navigate to page 3.
	err = pdf.SetPage(3)
	if err != nil {
		t.Fatalf("SetPage(3): %v", err)
	}
	pdf.SetXY(50, 100)
	_ = pdf.Text("Added to page 3")
}

func TestCov18_SetPage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPage(0)
	if err == nil {
		t.Fatal("expected error for page 0")
	}
	err = pdf.SetPage(99)
	if err == nil {
		t.Fatal("expected error for page 99")
	}
}

// ============================================================
// gopdf.go — AddPageWithOption (custom page size)
// ============================================================

func TestCov18_AddPageWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)

	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 400, H: 600},
	})
	pdf.SetXY(50, 50)
	_ = pdf.Text("Custom page size")

	w, h, err := pdf.GetPageSize(1)
	if err != nil {
		t.Fatalf("GetPageSize: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got w=%f h=%f", w, h)
	}
}

// ============================================================
// gopdf.go — SetNumberOfPages, GetNumberOfPages
// ============================================================

func TestCov18_GetNumberOfPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	if pdf.GetNumberOfPages() != 0 {
		t.Errorf("expected 0 pages initially, got %d", pdf.GetNumberOfPages())
	}

	pdf.AddPage()
	if pdf.GetNumberOfPages() != 1 {
		t.Errorf("expected 1 page, got %d", pdf.GetNumberOfPages())
	}

	pdf.AddPage()
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 3 {
		t.Errorf("expected 3 pages, got %d", pdf.GetNumberOfPages())
	}
}
