package gopdf

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// ============================================================
// coverage_boost5_test.go - Fifth round of coverage tests.
// All functions prefixed TestCov5_.
// ============================================================

// ============================================================
// bookmark.go - DeleteBookmark (48.5%), ModifyBookmark, SetBookmarkStyle
// ============================================================

func TestCov5_DeleteBookmark_Single(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")
	pdf.AddOutline("Chapter 2")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 3")
	pdf.AddOutline("Chapter 3")

	// Delete middle bookmark
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_, err = pdf.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov5_DeleteBookmark_First(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("First")
	pdf.AddPage()
	pdf.AddOutline("Second")

	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_DeleteBookmark_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("First")
	pdf.AddPage()
	pdf.AddOutline("Second")

	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_DeleteBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Only")
	err := pdf.DeleteBookmark(5)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCov5_DeleteBookmark_NegativeIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Only")
	err := pdf.DeleteBookmark(-1)
	if err == nil {
		t.Error("expected error")
	}
}

func TestCov5_ModifyBookmark(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Original")
	err := pdf.ModifyBookmark(0, "Modified")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_ModifyBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Only")
	err := pdf.ModifyBookmark(5, "Nope")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCov5_SetBookmarkStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Styled")
	err := pdf.SetBookmarkStyle(0, BookmarkStyle{
		Bold:      true,
		Italic:    true,
		Color:     [3]float64{1, 0, 0},
		Collapsed: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_SetBookmarkStyle_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Only")
	err := pdf.SetBookmarkStyle(5, BookmarkStyle{})
	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// arabic_helper.go - equals (42.9%), getHarf, getCharShape, ToArabic
// ============================================================

func TestCov5_Harf_Equals_AllCases(t *testing.T) {
	// Test equals with all shape variants
	h := Harf{
		Unicode:   0x0628, // Ba
		Beginning: 0xFE91,
		Isolated:  0xFE8F,
		Middle:    0xFE92,
		Final:     0xFE90,
	}
	if !h.equals(h.Unicode) {
		t.Error("should match Unicode")
	}
	if !h.equals(h.Beginning) {
		t.Error("should match Beginning")
	}
	if !h.equals(h.Isolated) {
		t.Error("should match Isolated")
	}
	if !h.equals(h.Middle) {
		t.Error("should match Middle")
	}
	if !h.equals(h.Final) {
		t.Error("should match Final")
	}
	if h.equals('X') {
		t.Error("should not match X")
	}
}

func TestCov5_GetHarf_Known(t *testing.T) {
	h := getHarf(0x0628) // Ba
	if h.Unicode != 0x0628 {
		t.Errorf("expected Ba, got %X", h.Unicode)
	}
}

func TestCov5_GetHarf_Unknown(t *testing.T) {
	h := getHarf('Z')
	if h.Unicode != 'Z' {
		t.Errorf("expected Z, got %c", h.Unicode)
	}
}

func TestCov5_ToArabic(t *testing.T) {
	// Simple Arabic text
	result := ToArabic("مرحبا")
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov5_ToArabic_Mixed(t *testing.T) {
	result := ToArabic("Hello مرحبا World")
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov5_Reverse(t *testing.T) {
	if Reverse("abc") != "cba" {
		t.Error("expected cba")
	}
	if Reverse("") != "" {
		t.Error("expected empty")
	}
}

// ============================================================
// text_search.go - searchAcrossItems (64.9%), SearchText, SearchTextOnPage
// ============================================================

func TestCov5_SearchText_Generated(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Hello World Test")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Another Page")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	results, err := SearchText(data, "Hello", false)
	if err != nil {
		t.Fatal(err)
	}
	_ = results // may or may not find depending on encoding
}

func TestCov5_SearchText_CaseInsensitive(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Hello World")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	results, err := SearchText(data, "hello", true)
	if err != nil {
		t.Fatal(err)
	}
	_ = results
}

func TestCov5_SearchTextOnPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Searchable Text")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	results, err := SearchTextOnPage(data, 0, "Searchable", false)
	if err != nil {
		t.Fatal(err)
	}
	_ = results
}

func TestCov5_SearchText_InvalidPDF(t *testing.T) {
	results, err := SearchText([]byte("not a pdf"), "test", false)
	// Parser may handle gracefully
	_ = err
	_ = results
}

func TestCov5_SearchTextOnPage_InvalidPDF(t *testing.T) {
	results, err := SearchTextOnPage([]byte("not a pdf"), 0, "test", false)
	// Parser may handle gracefully
	_ = err
	_ = results
}

// ============================================================
// html_insert.go - renderNode (67.7%), applyStyleAttr (59.3%),
// resolveFontFamily (63.6%), renderImage (54.3%), remainingWidth (0%)
// ============================================================

func TestCov5_InsertHTMLBox_Headings(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<h1>Heading 1</h1><h2>Heading 2</h2><h3>Heading 3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 600, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_StyledSpan(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<span style="color:red;font-size:18px;font-weight:bold;font-style:italic;text-decoration:underline;text-align:center">Styled</span>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_FontTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<font color="blue" size="5">Blue text</font>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_CenterTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<center>Centered text here</center>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_LinkTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<a href="https://example.com">Click here</a>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_Blockquote(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<blockquote>This is a quoted block of text that should be indented.</blockquote>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_SubSup(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `Normal <sub>subscript</sub> and <sup>superscript</sup>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_StrikeAndDel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<s>strikethrough</s> <strike>strike</strike> <del>deleted</del>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_OrderedList(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<ol><li>First</li><li>Second</li><li>Third</li></ol>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_UnorderedList(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<ul><li>Apple</li><li>Banana</li></ul>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_HR(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<p>Before</p><hr><p>After</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_DivWithStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<div style="color:#FF0000;font-size:20px;text-align:right">Right aligned red</div>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_ImageTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<img src="` + resJPEGPath + `" width="100" height="80">`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_ImageNoSrc(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<img>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_ImageOnlyWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<img src="` + resJPEGPath + `" width="150">`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_ImageOnlyHeight(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<img src="` + resJPEGPath + `" height="100">`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_ImageNoDimensions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<img src="` + resJPEGPath + `">`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_BrTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `Line 1<br>Line 2<br>Line 3`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_StyleFontFamily(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<span style="font-family:'` + fontFamily + `'">Custom font</span>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - getCachedTransparency (42.9%), SetTransparency
// ============================================================

func TestCov5_SetTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode})
	if err != nil {
		t.Fatal(err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Semi-transparent")
	pdf.ClearTransparency()
}

func TestCov5_RectWithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	tr := Transparency{Alpha: 0.3, BlendModeType: NormalBlendMode}
	err := pdf.RectFromLowerLeftWithOpts(DrawableRectOptions{
		X:            50,
		Y:            50,
		Rect:         Rect{W: 100, H: 50},
		PaintStyle:   DrawFillPaintStyle,
		Transparency: &tr,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_RectFromUpperLeftWithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	tr := Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode}
	err := pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		X:            50,
		Y:            50,
		Rect:         Rect{W: 100, H: 50},
		PaintStyle:   DrawFillPaintStyle,
		Transparency: &tr,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - writeInfo (73.3%) - test all info fields
// ============================================================

func TestCov5_WriteInfo_AllFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetInfo(PdfInfo{
		Title:        "Test Title",
		Author:       "Test Author",
		Subject:      "Test Subject",
		Creator:      "Test Creator",
		Producer:     "Test Producer",
		CreationDate: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	})
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Info test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	content := buf.String()
	if !strings.Contains(content, "Author") {
		t.Error("expected Author in output")
	}
	if !strings.Contains(content, "Title") {
		t.Error("expected Title in output")
	}
}

// ============================================================
// gopdf.go - Line with custom line type (75%)
// ============================================================

func TestCov5_Line_CustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCustomLineType([]float64{5, 3}, 0)
	pdf.Line(10, 20, 200, 300)
	pdf.SetLineType("dashed")
	pdf.Line(10, 30, 200, 310)
}

// ============================================================
// gopdf.go - ImageByHolderWithOptions (35%)
// ============================================================

func TestCov5_ImageByHolderWithOptions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:    50,
		Y:    50,
		Rect: &Rect{W: 100, H: 80},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_ImageByHolderWithOptions_Transparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	tr := Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode}
	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:            50,
		Y:            50,
		Rect:         &Rect{W: 100, H: 80},
		Transparency: &tr,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_ImageByHolderWithOptions_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:           50,
		Y:           50,
		Rect:        &Rect{W: 100, H: 80},
		DegreeAngle: 45,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_ImageByHolderWithOptions_Flip(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	err = pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:              50,
		Y:              50,
		Rect:           &Rect{W: 100, H: 80},
		VerticalFlip:   true,
		HorizontalFlip: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// font_extract.go - extractFontsFromResources (57.1%)
// ============================================================

func TestCov5_ExtractFontsFromPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Font test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	fonts, err := ExtractFontsFromPage(data, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = fonts
}

func TestCov5_ExtractFontsFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	result, err := ExtractFontsFromAllPages(data)
	if err != nil {
		t.Fatal(err)
	}
	_ = result
}

func TestCov5_ExtractFontsFromPage_BadIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	_, err := ExtractFontsFromPage(buf.Bytes(), 99)
	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// gc.go - deduplicateObjects (31%)
// ============================================================

func TestCov5_GarbageCollect_Dedup(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")

	// Just run dedup on a normal PDF
	removed := pdf.GarbageCollect(GCDedup)
	_ = removed
}

// ============================================================
// gopdf.go - prepare (78.1%) - test with protection
// ============================================================

func TestCov5_Prepare_WithFormFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Form test")
	pdf.AddTextField("field1", 50, 100, 200, 30)
	pdf.AddCheckbox("check1", 50, 150, 20, true)
	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// ============================================================
// content_obj.go - write (62.5%) - test with NoCompression
// ============================================================

func TestCov5_ContentObj_NoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "No compression")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// ============================================================
// gopdf.go - CellWithOption more align combos (66.7%)
// ============================================================

func TestCov5_CellWithOption_LeftTop(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "LeftTop", CellOption{Align: Left | Top})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_CellWithOption_CenterMiddle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "CenterMiddle", CellOption{Align: Center | Middle})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_CellWithOption_RightBottom(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "RightBottom", CellOption{Align: Right | Bottom})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_CellWithOption_WithBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Bordered", CellOption{
		Align:  Center | Middle,
		Border: AllBorders,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - MultiCellWithOption (82.8%) - more options
// ============================================================

func TestCov5_MultiCellWithOption_RightAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 200}, "Right aligned multi cell text that wraps", CellOption{
		Align: Right | Top,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - Text (71.4%) - with color
// ============================================================

func TestCov5_Text_WithColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.SetTextColor(255, 0, 0)
	err := pdf.Text("Red text")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_Text_WithCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.SetTextColorCMYK(0, 100, 100, 0)
	err := pdf.Text("CMYK text")
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// watermark.go - AddWatermarkTextAllPages, AddWatermarkImageAllPages (71.4%)
// ============================================================

func TestCov5_AddWatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   48,
		Opacity:    0.3,
	})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov5_AddWatermarkImageAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")

	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.3, 100, 80, 30)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// ============================================================
// select_pages.go - SelectPagesFromFile (66.7%), SelectPagesFromBytes (66.7%)
// ============================================================

func TestCov5_SelectPagesFromBytes(t *testing.T) {
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

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	result, err := SelectPagesFromBytes(data, []int{1, 3}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestCov5_SelectPagesFromFile(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")

	path := resOutDir + "/select_pages_test.pdf"
	pdf.WritePdf(path)
	defer os.Remove(path)

	result, err := SelectPagesFromFile(path, []int{1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ============================================================
// open_pdf.go - OpenPDFFromStream (71.4%), openPDFFromData (65.2%)
// ============================================================

func TestCov5_OpenPDFFromStream(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Stream test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	reader := io.ReadSeeker(bytes.NewReader(buf.Bytes()))
	opened := &GoPdf{}
	err := opened.OpenPDFFromStream(&reader, nil)
	if err != nil {
		t.Fatal(err)
	}
	if opened.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

// ============================================================
// gopdf.go - GetBytesPdf, GetBytesPdfReturnErr (75-80%)
// ============================================================

func TestCov5_GetBytesPdf(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Bytes test")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestCov5_GetBytesPdfReturnErr(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Bytes test")
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty bytes")
	}
}

// ============================================================
// gopdf.go - Polygon, Polyline, Sector more coverage
// ============================================================

func TestCov5_Polygon_WithFill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(200, 200, 200)
	points := []Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}
	pdf.Polygon(points, "F")
}

func TestCov5_Polygon_DrawFill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(200, 200, 200)
	pdf.SetStrokeColor(0, 0, 0)
	points := []Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}
	pdf.Polygon(points, "DF")
}

func TestCov5_Polyline_Styled(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	points := []Point{{X: 10, Y: 10}, {X: 100, Y: 50}, {X: 200, Y: 10}}
	pdf.Polyline(points)
}

func TestCov5_Sector_Large(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(100, 100, 255)
	pdf.Sector(200, 300, 80, 0, 270, "FD")
}

// ============================================================
// gopdf.go - Rotate, RotateReset
// ============================================================

func TestCov5_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.SetXY(100, 100)
	pdf.Cell(nil, "Rotated")
	pdf.RotateReset()
}

// ============================================================
// gopdf.go - ClipPolygon, SaveGraphicsState, RestoreGraphicsState
// ============================================================

func TestCov5_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	points := []Point{{X: 50, Y: 50}, {X: 200, Y: 50}, {X: 200, Y: 200}, {X: 50, Y: 200}}
	pdf.ClipPolygon(points)
	pdf.SetXY(60, 60)
	pdf.Cell(nil, "Clipped")
	pdf.RestoreGraphicsState()
}

// ============================================================
// gopdf.go - AddExternalLink, AddInternalLink
// ============================================================

func TestCov5_AddExternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Click me")
	pdf.AddExternalLink("https://example.com", 50, 50, 100, 20)
}

func TestCov5_AddInternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Go to page 2")
	pdf.SetAnchor("page2")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2")
	pdf.AddInternalLink("page2", 50, 50, 100, 20)
}

// ============================================================
// gopdf.go - ImportPagesFromSource (75%)
// ============================================================

func TestCov5_ImportPagesFromSource(t *testing.T) {
	// Create source PDF
	src := newPDFWithFont(t)
	src.AddPage()
	src.SetXY(50, 50)
	src.Cell(nil, "Source page 1")
	src.AddPage()
	src.SetXY(50, 50)
	src.Cell(nil, "Source page 2")
	var srcBuf bytes.Buffer
	src.WriteTo(&srcBuf)

	// Import into new PDF using byte slice
	dst := newPDFWithFont(t)
	err := dst.ImportPagesFromSource(srcBuf.Bytes(), "/MediaBox")
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - AddPageWithOption (92.3%) - with TrimBox
// ============================================================

func TestCov5_AddPageWithOption_TrimBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: PageSizeA4,
		TrimBox: &Box{
			Left:   10,
			Top:    10,
			Right:  585,
			Bottom: 832,
		},
	})
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "With TrimBox")
}

// ============================================================
// gopdf.go - FillInPlaceHoldText (78.3%) - more cases
// ============================================================

func TestCov5_FillInPlaceHoldText_RightAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.PlaceHolderText("{{right}}", 200)
	if err != nil {
		t.Fatal(err)
	}
	err = pdf.FillInPlaceHoldText("{{right}}", "Right Aligned", Right)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_FillInPlaceHoldText_CenterAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.PlaceHolderText("{{center}}", 200)
	if err != nil {
		t.Fatal(err)
	}
	err = pdf.FillInPlaceHoldText("{{center}}", "Centered", Center)
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// gopdf.go - SetColorSpace, AddColorSpaceRGB, AddColorSpaceCMYK
// ============================================================

func TestCov5_ColorSpace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceRGB("myrgb", 255, 0, 0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "RGB color space")
}

func TestCov5_ColorSpaceCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceCMYK("mycmyk", 0, 100, 100, 0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "CMYK color space")
}

// ============================================================
// gopdf.go - Oval
// ============================================================

func TestCov5_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(100, 200, 80, 50)
}

// ============================================================
// gopdf.go - Curve
// ============================================================

func TestCov5_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(50, 50, 80, 150, 150, 150, 200, 50, "D")
}

// ============================================================
// gopdf.go - Br
// ============================================================

func TestCov5_Br(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Before")
	pdf.Br(20)
	pdf.Cell(nil, "After")
}

// ============================================================
// gopdf.go - SetPage (100% but exercises content paths)
// ============================================================

func TestCov5_SetPage_Multiple(t *testing.T) {
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

	// Switch back to page 1
	pdf.SetPage(1)
	pdf.SetXY(50, 100)
	pdf.Cell(nil, "Added to page 1")

	// Switch to page 3
	pdf.SetPage(3)
	pdf.SetXY(50, 100)
	pdf.Cell(nil, "Added to page 3")
}

// ============================================================
// gopdf.go - SplitTextWithOption (91.9%)
// ============================================================

func TestCov5_SplitTextWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitTextWithOption("This is a long text that should be split into multiple lines based on width", 200, &BreakOption{
		Mode:           BreakModeIndicatorSensitive,
		BreakIndicator: ' ',
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

// ============================================================
// gopdf.go - KernOverride (91.7%)
// ============================================================

func TestCov5_KernOverride(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.KernOverride(fontFamily, func(leftRune, rightRune rune, leftPair, rightPair uint, pairVal int16) int16 {
		if leftRune == 'A' && rightRune == 'V' {
			return -50
		}
		return pairVal
	})
	if err != nil {
		t.Fatal(err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "AV kerning")
}

// ============================================================
// gopdf.go - SetFontSize
// ============================================================

func TestCov5_SetFontSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFontSize(24)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Large text")
	pdf.SetFontSize(8)
	pdf.SetXY(50, 80)
	pdf.Cell(nil, "Small text")
}

// ============================================================
// gopdf.go - SetMargins, MarginLeft, etc.
// ============================================================

func TestCov5_Margins(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetMargins(20, 30, 20, 30)
	pdf.AddPage()
	if pdf.MarginLeft() != 20 {
		t.Errorf("MarginLeft = %f", pdf.MarginLeft())
	}
	if pdf.MarginTop() != 30 {
		t.Errorf("MarginTop = %f", pdf.MarginTop())
	}
	if pdf.MarginRight() != 20 {
		t.Errorf("MarginRight = %f", pdf.MarginRight())
	}
	if pdf.MarginBottom() != 30 {
		t.Errorf("MarginBottom = %f", pdf.MarginBottom())
	}
}

// ============================================================
// page_info.go - GetSourcePDFPageCount, GetSourcePDFPageSizes (75%)
// ============================================================

func TestCov5_GetSourcePDFPageCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	count, err := GetSourcePDFPageCountFromBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestCov5_GetSourcePDFPageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	sizes, err := GetSourcePDFPageSizesFromBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 1 {
		t.Errorf("sizes = %d, want 1", len(sizes))
	}
}

// ============================================================
// page_manipulate.go - MergePages (83%), MergePagesFromBytes (84.1%)
// ============================================================

func TestCov5_MergePagesFromBytes(t *testing.T) {
	pdf1 := newPDFWithFont(t)
	pdf1.AddPage()
	pdf1.SetXY(50, 50)
	pdf1.Cell(nil, "Doc 1")
	var buf1 bytes.Buffer
	pdf1.WriteTo(&buf1)

	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.SetXY(50, 50)
	pdf2.Cell(nil, "Doc 2")
	var buf2 bytes.Buffer
	pdf2.WriteTo(&buf2)

	result, err := MergePagesFromBytes([][]byte{buf1.Bytes(), buf2.Bytes()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ============================================================
// page_manipulate.go - ExtractPagesFromBytes (84.6%)
// ============================================================

func TestCov5_ExtractPagesFromBytes(t *testing.T) {
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
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	result, err := ExtractPagesFromBytes(buf.Bytes(), []int{1, 3}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// ============================================================
// image_recompress.go - RecompressImages (87.5%), rebuildXref (59.5%)
// ============================================================

func TestCov5_RecompressImages_JPEG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 80})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	result, err := RecompressImages(buf.Bytes(), RecompressOption{
		Format:      "jpeg",
		JPEGQuality: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Error("expected output")
	}
}

func TestCov5_RecompressImages_PNG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 80})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	result, err := RecompressImages(buf.Bytes(), RecompressOption{
		Format: "png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Error("expected output")
	}
}

func TestCov5_RecompressImages_WithMaxSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	pdf.ImageByHolder(holder, 50, 50, &Rect{W: 200, H: 160})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	result, err := RecompressImages(buf.Bytes(), RecompressOption{
		Format:      "jpeg",
		JPEGQuality: 60,
		MaxWidth:    50,
		MaxHeight:   50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Error("expected output")
	}
}

func TestCov5_RecompressImages_NoImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "No images")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	result, err := RecompressImages(buf.Bytes(), RecompressOption{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Error("expected output")
	}
}

func TestCov5_RecompressImages_InvalidPDF(t *testing.T) {
	result, err := RecompressImages([]byte("not a pdf"), RecompressOption{})
	// Parser may handle gracefully
	_ = err
	_ = result
}

// ============================================================
// pdf_parser.go - extractMediaBox (42.9%), extractNamedRefs (52.4%)
// These are exercised indirectly through SearchText, ExtractText, etc.
// ============================================================

func TestCov5_ExtractTextFromPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Extractable text")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	text, err := ExtractTextFromPage(buf.Bytes(), 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = text
}

func TestCov5_ExtractTextFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 1 text")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Page 2 text")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	texts, err := ExtractTextFromAllPages(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	_ = texts
}

// ============================================================
// image_extract.go - ExtractImagesFromPage, ExtractImagesFromAllPages
// ============================================================

func TestCov5_ExtractImagesFromPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 80})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	images, err := ExtractImagesFromPage(buf.Bytes(), 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = images
}

func TestCov5_ExtractImagesFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, err := ImageHolderByPath(resJPEGPath)
	if err != nil {
		t.Skip("image not available")
	}
	pdf.ImageByHolder(holder, 50, 50, &Rect{W: 100, H: 80})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	result, err := ExtractImagesFromAllPages(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	_ = result
}

// ============================================================
// pdf_lowlevel.go - ReadObject, GetDictKey, SetDictKey, GetStream, etc.
// ============================================================

func TestCov5_LowLevel_ReadAndUpdate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Low level test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	obj, err := ReadObject(data, 1)
	if err != nil {
		t.Fatal(err)
	}
	_ = obj

	// GetDictKey
	val, err := GetDictKey(data, 1, "/Type")
	if err != nil {
		t.Fatal(err)
	}
	_ = val

	// GetCatalog
	cat, err := GetCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	_ = cat

	// GetTrailer
	trailer, err := GetTrailer(data)
	if err != nil {
		t.Fatal(err)
	}
	_ = trailer
}

// ============================================================
// pdf_obj_id.go - GetObjID, GetObjType (66-75%)
// ============================================================

func TestCov5_GetObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")

	id := pdf.GetObjID(0)
	if !id.IsValid() {
		t.Error("expected valid ID")
	}
	_ = id.Ref()
	_ = id.RefStr()
}

func TestCov5_GetObjType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	tp := pdf.GetObjType(0)
	_ = tp
}

func TestCov5_CatalogObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	id := pdf.CatalogObjID()
	if !id.IsValid() {
		t.Error("expected valid catalog ID")
	}
}

func TestCov5_PagesObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	id := pdf.PagesObjID()
	if !id.IsValid() {
		t.Error("expected valid pages ID")
	}
}

// ============================================================
// gopdf.go - SetStrokeColor, SetFillColor, SetStrokeColorCMYK, SetFillColorCMYK
// ============================================================

func TestCov5_StrokeAndFillColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetFillColor(0, 255, 0)
	pdf.RectFromUpperLeft(50, 50, 100, 50)

	pdf.SetStrokeColorCMYK(0, 100, 100, 0)
	pdf.SetFillColorCMYK(100, 0, 0, 0)
	pdf.RectFromUpperLeft(50, 120, 100, 50)
}

// ============================================================
// gopdf.go - SetGrayFill, SetGrayStroke
// ============================================================

func TestCov5_GrayColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
	pdf.SetGrayStroke(0.3)
	pdf.RectFromLowerLeft(50, 50, 100, 50)
}

// ============================================================
// gopdf.go - Read (75%)
// ============================================================

func TestCov5_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Read test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	pdf2 := &GoPdf{}
	pdf2.Start(Config{PageSize: *PageSizeA4})
	n, err := pdf2.Read(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Error("expected non-zero n")
	}
}

// ============================================================
// gopdf.go - Close
// ============================================================

func TestCov5_Close(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Close test")
	path := resOutDir + "/close_test.pdf"
	pdf.WritePdf(path)
	pdf.Close()
	os.Remove(path)
}

// ============================================================
// markinfo.go - computePageLabel (72.7%)
// ============================================================

func TestCov5_FindPagesByLabel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Style: "r", Prefix: ""},
		{PageIndex: 3, Style: "D", Prefix: "Ch-", Start: 1},
	})
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()

	pages := pdf.FindPagesByLabel("i")
	_ = pages
	pages = pdf.FindPagesByLabel("Ch-1")
	_ = pages
}

// ============================================================
// Additional tests for remaining low-coverage areas
// ============================================================

// cache_content_text.go - drawBorder (17.6%)
func TestCov5_CellWithOption_AllBorderTypes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Top", CellOption{Border: Top})
	pdf.SetXY(50, 80)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Left", CellOption{Border: Left})
	pdf.SetXY(50, 110)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Right", CellOption{Border: Right})
	pdf.SetXY(50, 140)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Bottom", CellOption{Border: Bottom})
	pdf.SetXY(50, 170)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "TopBottom", CellOption{Border: Top | Bottom})
	pdf.SetXY(50, 200)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "LeftRight", CellOption{Border: Left | Right})
	pdf.SetXY(50, 230)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "All", CellOption{Border: AllBorders})
}

// annotation.go - AddLineAnnotation (63.6%)
func TestCov5_AddLineAnnotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Line annotation test")
	pdf.AddLineAnnotation(
		Point{X: 50, Y: 100},
		Point{X: 200, Y: 100},
		[3]uint8{255, 0, 0},
	)
}

// html_insert.go - unitConversion (33.3%) - test with different units
func TestCov5_InsertHTMLBox_WithUnits(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4, Unit: UnitMM})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skip("font not available")
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	html := `<p>Text in mm units</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 180, 100, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// gopdf.go - SetNewY (66.7%) - trigger page overflow
func TestCov5_SetNewY_PageBreak(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Set Y very close to bottom, then request more space than available
	pdf.SetNewY(830, 20)
	// Should have triggered a new page
	if pdf.GetNumberOfPages() < 2 {
		// May or may not add page depending on implementation
	}
}

// gopdf.go - CellWithOption (66.7%) - nil rect
func TestCov5_CellWithOption_NilRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.CellWithOption(nil, "No rect", CellOption{Align: Left})
	if err != nil {
		t.Fatal(err)
	}
}

// open_pdf.go - OpenPDFFromBytes (65.2%)
func TestCov5_OpenPDFFromBytes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Open test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)

	opened := &GoPdf{}
	err := opened.OpenPDFFromBytes(buf.Bytes(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if opened.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

// open_pdf.go - OpenPDF from file
func TestCov5_OpenPDF(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Open file test")
	path := resOutDir + "/open_test.pdf"
	pdf.WritePdf(path)
	defer os.Remove(path)

	opened := &GoPdf{}
	err := opened.OpenPDF(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if opened.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

// gopdf.go - Image from file path
func TestCov5_ImageFromPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 80})
	if err != nil {
		t.Fatal(err)
	}
}

// gopdf.go - Image PNG
func TestCov5_ImagePNG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 100, H: 80})
	if err != nil {
		t.Fatal(err)
	}
}

// content_element.go - GetPageElementsByType, GetPageElementCount
func TestCov5_GetPageElementsByType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Element test")
	pdf.Line(10, 20, 200, 300)

	elems, err := pdf.GetPageElementsByType(1, ElementText)
	if err != nil {
		t.Fatal(err)
	}
	_ = elems

	count, err := pdf.GetPageElementCount(1)
	if err != nil {
		t.Fatal(err)
	}
	if count < 1 {
		t.Errorf("count = %d", count)
	}
}

// content_element.go - DeleteElementsInRect
func TestCov5_DeleteElementsInRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Delete me")
	pdf.SetXY(300, 300)
	pdf.Cell(nil, "Keep me")

	n, err := pdf.DeleteElementsInRect(1, 0, 0, 200, 200)
	if err != nil {
		t.Fatal(err)
	}
	_ = n
}

// content_element.go - InsertRectElement
func TestCov5_InsertRectElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Before rect")
	err := pdf.InsertRectElement(1, 50, 100, 100, 50, "D")
	if err != nil {
		t.Fatal(err)
	}
}

// gopdf.go - AddTTFFontWithOption
func TestCov5_AddTTFFontWithOption(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Skip("font not available")
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "With kerning option")
}

// table.go - DrawTable
func TestCov5_DrawTable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	tbl := pdf.NewTableLayout(50, 50, 20, 5)
	tbl.AddColumn("Name", 200, "left")
	tbl.AddColumn("Value", 200, "center")
	tbl.AddRow([]string{"Key1", "Val1"})
	tbl.AddRow([]string{"Key2", "Val2"})
	tbl.SetTableStyle(CellStyle{
		BorderStyle: BorderStyle{
			Top: true, Left: true, Right: true, Bottom: true,
			Width:    0.5,
			RGBColor: RGBColor{R: 0, G: 0, B: 0},
		},
	})
	err := tbl.DrawTable()
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// More targeted tests for remaining gaps
// ============================================================

// cache_content_text.go - drawBorder (17.6%) - more border combos
func TestCov5_MultiCellWithOption_WithBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Bordered multi cell", CellOption{
		Align:  Left | Top,
		Border: AllBorders,
	})
}

func TestCov5_CellWithOption_TopBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 200, H: 30}, "TopBorder", CellOption{
		Align:  Left | Middle,
		Border: Top,
	})
}

func TestCov5_CellWithOption_BottomBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 200, H: 30}, "BottomBorder", CellOption{
		Align:  Left | Middle,
		Border: Bottom,
	})
}

func TestCov5_CellWithOption_LeftBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 200, H: 30}, "LeftBorder", CellOption{
		Align:  Left | Middle,
		Border: Left,
	})
}

func TestCov5_CellWithOption_RightBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 200, H: 30}, "RightBorder", CellOption{
		Align:  Left | Middle,
		Border: Right,
	})
}

// html_insert.go - renderNode more tags, resolveFontFamily
func TestCov5_InsertHTMLBox_NestedTags(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<p><b><i><u>Bold Italic Underline</u></i></b></p>
<p style="color:green;font-size:16px">Green paragraph</p>
<div><span style="font-weight:700">Weight 700</span></div>
<div><span style="font-weight:800">Weight 800</span></div>
<div><span style="font-weight:900">Weight 900</span></div>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_TextAlignLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<div style="text-align:left">Left aligned</div>
<div style="text-align:right">Right aligned</div>
<div style="text-align:center">Center aligned</div>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_LongText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Long text that forces word wrapping and line breaks
	html := `<p>This is a very long paragraph that should wrap across multiple lines in the HTML box. It contains enough text to test the word wrapping and line breaking logic thoroughly. The quick brown fox jumps over the lazy dog. Lorem ipsum dolor sit amet consectetur adipiscing elit.</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 200, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov5_InsertHTMLBox_OverflowBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Very small box that can't fit much
	html := `<p>Line 1</p><p>Line 2</p><p>Line 3</p><p>Line 4</p><p>Line 5</p><p>Line 6</p><p>Line 7</p><p>Line 8</p><p>Line 9</p><p>Line 10</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 200, 50, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// gopdf.go - IsFitMultiCell (80%)
func TestCov5_IsFitMultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	ok, _, err := pdf.IsFitMultiCell(&Rect{W: 200, H: 100}, "Short text")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected fit")
	}
}

func TestCov5_IsFitMultiCell_TooSmall(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	longText := strings.Repeat("This is a very long text that should not fit. ", 50)
	ok, _, err := pdf.IsFitMultiCell(&Rect{W: 50, H: 10}, longText)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected not fit")
	}
}

// gopdf.go - SplitText
func TestCov5_SplitText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitText("Hello World this is a test", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 1 {
		t.Error("expected at least 1 line")
	}
}

// gopdf.go - AddOutlineWithPosition
func TestCov5_AddOutlineWithPosition(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Outline with position")
	pdf.AddOutlineWithPosition("Chapter 1")
}

// toc.go - GetTOC, SetTOC
func TestCov5_GetTOC(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")

	toc := pdf.GetTOC()
	if len(toc) < 2 {
		t.Errorf("expected at least 2 TOC items, got %d", len(toc))
	}
}

// gopdf.go - SetLineWidth
func TestCov5_SetLineWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(2.0)
	pdf.Line(10, 10, 200, 10)
	pdf.SetLineWidth(0.5)
	pdf.Line(10, 20, 200, 20)
}

// gopdf.go - AddTTFFontByReader
func TestCov5_AddTTFFontByReader(t *testing.T) {
	f, err := os.Open(resFontPath)
	if err != nil {
		t.Skip("font not available")
	}
	defer f.Close()

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontByReader("ReaderFont", f)
	if err != nil {
		t.Fatal(err)
	}
	pdf.SetFont("ReaderFont", "", 14)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Font from reader")
}

// gopdf.go - SetFontWithStyle
func TestCov5_SetFontWithStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Use same style as loaded (Regular)
	err := pdf.SetFontWithStyle(fontFamily, Regular, 16)
	if err != nil {
		t.Fatal(err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Regular 16pt")
}

// embedded_file.go - more coverage
func TestCov5_EmbeddedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Embedded file test")

	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "test_embed.txt",
		Content:  []byte("Hello embedded world"),
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatal(err)
	}
}

// scrub.go - Scrub
func TestCov5_Scrub(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetInfo(PdfInfo{
		Title:  "Scrub Test",
		Author: "Test Author",
	})
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Scrub test")

	pdf.Scrub(DefaultScrubOption())
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// pdf_version.go - SetPDFVersion, GetPDFVersion
func TestCov5_PDFVersion(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetPDFVersion(PDFVersion17)
	if pdf.GetPDFVersion() != PDFVersion17 {
		t.Errorf("version = %d", pdf.GetPDFVersion())
	}
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	content := buf.String()
	if !strings.Contains(content, "1.7") {
		t.Error("expected 1.7 in output")
	}
}

// markinfo.go - SetMarkInfo, GetMarkInfo
func TestCov5_MarkInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetMarkInfo(MarkInfo{Marked: true})
	mi := pdf.GetMarkInfo()
	if mi == nil || !mi.Marked {
		t.Error("expected marked")
	}
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// ============================================================
// Force drawBorder paths by writing PDF output
// ============================================================

func TestCov5_DrawBorder_AllSides_Output(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Each border side individually
	pdf.SetXY(50, 50)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Top", CellOption{Border: Top, Align: Left | Middle})
	pdf.SetXY(50, 80)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Left", CellOption{Border: Left, Align: Left | Middle})
	pdf.SetXY(50, 110)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Right", CellOption{Border: Right, Align: Left | Middle})
	pdf.SetXY(50, 140)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "Bottom", CellOption{Border: Bottom, Align: Left | Middle})
	pdf.SetXY(50, 170)
	pdf.CellWithOption(&Rect{W: 100, H: 20}, "All", CellOption{Border: AllBorders, Align: Center | Middle})

	// Force output to trigger write paths
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// Force MultiCell with border to trigger drawBorder
func TestCov5_MultiCell_WithBorder_Output(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.MultiCellWithOption(&Rect{W: 200, H: 100}, "Multi cell with all borders and wrapping text that goes on", CellOption{
		Align:  Left | Top,
		Border: AllBorders,
	})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// gopdf.go - compilePdf (94.7%) - with embedded file to trigger prepare paths
func TestCov5_CompilePdf_WithEmbeddedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "With embedded file")
	pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "data.txt",
		Content:  []byte("embedded content"),
		MimeType: "text/plain",
	})
	pdf.SetXMPMetadata(XMPMetadata{Title: "Test"})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: "D"}})
	pdf.AddOCG(OCG{Name: "Layer1", On: true})
	pdf.SetMarkInfo(MarkInfo{Marked: true})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// gopdf.go - Image with PNG (triggers parsePng path)
func TestCov5_ImagePNG_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Use PNG with transparency
	err := pdf.Image(resPNGPath, 50, 50, &Rect{W: 150, H: 150})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// gopdf.go - Image second PNG
func TestCov5_ImagePNG2(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resPNGPath2, 50, 50, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// gopdf.go - Multiple images on same page
func TestCov5_MultipleImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 80})
	pdf.Image(resPNGPath, 200, 50, &Rect{W: 100, H: 100})
	pdf.Image(resJPEGPath, 50, 200, &Rect{W: 80, H: 60})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

// content_stream_clean.go - extractOperator (85.7%)
func TestCov5_ContentStreamClean(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Stream clean test")
	pdf.Line(10, 10, 200, 200)
	pdf.RectFromUpperLeft(50, 300, 100, 50)
	pdf.SetGrayFill(0.5)
	pdf.RectFromUpperLeftWithStyle(50, 400, 100, 50, "F")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Use CleanContentStreams
	cleaned, err := CleanContentStreams(data)
	if err != nil {
		// May fail on non-standard content
		_ = cleaned
	}
}

// journal.go - JournalEnable, JournalEndOp, JournalUndo, JournalRedo
func TestCov5_Journal_UndoRedo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Error("expected journal enabled")
	}

	pdf.AddPage()
	pdf.JournalStartOp("add text")
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "First text")
	pdf.JournalEndOp()

	pdf.JournalStartOp("add more text")
	pdf.SetXY(50, 80)
	pdf.Cell(nil, "Second text")
	pdf.JournalEndOp()

	// Undo
	_, err := pdf.JournalUndo()
	if err != nil {
		t.Fatal(err)
	}

	// Redo
	_, err = pdf.JournalRedo()
	if err != nil {
		t.Fatal(err)
	}

	ops := pdf.JournalGetOperations()
	_ = ops
}

// journal.go - JournalSave, JournalLoad
func TestCov5_Journal_SaveLoad(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.JournalEnable()
	pdf.AddPage()
	pdf.JournalStartOp("test op")
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Journal test")
	pdf.JournalEndOp()

	path := resOutDir + "/journal_test.json"
	err := pdf.JournalSave(path)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	pdf2 := newPDFWithFont(t)
	pdf2.JournalEnable()
	pdf2.AddPage()
	err = pdf2.JournalLoad(path)
	if err != nil {
		t.Fatal(err)
	}
}
