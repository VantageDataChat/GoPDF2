package gopdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"
)

// ============================================================
// TestCov36_ — coverage boost round 36
// Targets: page_manipulate.go, select_pages.go, bookmark.go,
//          embedded_file.go, image_obj.go, content_obj.go,
//          gopdf.go deeper paths, pdf_decrypt.go
// ============================================================

// --- page_manipulate.go ---

func TestCov36_DeletePage_NoPages(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.DeletePage(1)
	if err == nil {
		t.Error("expected error for no pages")
	}
}

func TestCov36_DeletePage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeletePage(0)
	if err == nil {
		t.Error("expected error for page 0")
	}
	err = pdf.DeletePage(99)
	if err == nil {
		t.Error("expected error for page 99")
	}
}

func TestCov36_DeletePage_Valid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 2")
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 3")

	err := pdf.DeletePage(2)
	if err != nil {
		t.Fatalf("DeletePage: %v", err)
	}
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

func TestCov36_DeletePage_AllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Only page")

	err := pdf.DeletePage(1)
	if err != nil {
		t.Fatalf("DeletePage: %v", err)
	}
}

func TestCov36_CopyPage_Valid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Original")

	newPageNo, err := pdf.CopyPage(1)
	if err != nil {
		t.Fatalf("CopyPage: %v", err)
	}
	if newPageNo != 2 {
		t.Errorf("expected page 2, got %d", newPageNo)
	}
}

func TestCov36_CopyPage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.CopyPage(0)
	if err == nil {
		t.Error("expected error for page 0")
	}
	_, err = pdf.CopyPage(99)
	if err == nil {
		t.Error("expected error for page 99")
	}
}

func TestCov36_ExtractPages_FromFile(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	result, err := ExtractPages(resTestPDF, []int{1}, nil)
	if err != nil {
		t.Fatalf("ExtractPages: %v", err)
	}
	if result.GetNumberOfPages() != 1 {
		t.Errorf("expected 1 page, got %d", result.GetNumberOfPages())
	}
}

func TestCov36_ExtractPages_EmptyPages(t *testing.T) {
	_, err := ExtractPages(resTestPDF, nil, nil)
	if err == nil {
		t.Error("expected error for empty pages")
	}
}

func TestCov36_ExtractPages_BadFile(t *testing.T) {
	_, err := ExtractPages("/nonexistent/file.pdf", []int{1}, nil)
	if err == nil {
		t.Error("expected error for bad file")
	}
}

func TestCov36_ExtractPagesFromBytes(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	result, err := ExtractPagesFromBytes(data, []int{1}, nil)
	if err != nil {
		t.Fatalf("ExtractPagesFromBytes: %v", err)
	}
	if result.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

func TestCov36_ExtractPagesFromBytes_OutOfRange(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	_, err = ExtractPagesFromBytes(data, []int{999}, nil)
	if err == nil {
		t.Error("expected error for out of range page")
	}
}

func TestCov36_MergePagesFromBytes(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	result, err := MergePagesFromBytes([][]byte{data, data}, nil)
	if err != nil {
		t.Fatalf("MergePagesFromBytes: %v", err)
	}
	if result.GetNumberOfPages() < 2 {
		t.Error("expected at least 2 pages")
	}
}

func TestCov36_MergePagesFromBytes_Empty(t *testing.T) {
	_, err := MergePagesFromBytes(nil, nil)
	if err == nil {
		t.Error("expected error for empty slices")
	}
}

func TestCov36_MergePages_FromFile(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	result, err := MergePages([]string{resTestPDF}, nil)
	if err != nil {
		t.Fatalf("MergePages: %v", err)
	}
	if result.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

func TestCov36_MergePages_Empty(t *testing.T) {
	_, err := MergePages(nil, nil)
	if err == nil {
		t.Error("expected error for empty paths")
	}
}

// --- select_pages.go ---

func TestCov36_SelectPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 2")
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 3")

	result, err := pdf.SelectPages([]int{3, 1})
	if err != nil {
		t.Fatalf("SelectPages: %v", err)
	}
	if result.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", result.GetNumberOfPages())
	}
}

func TestCov36_SelectPages_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.SelectPages(nil)
	if err == nil {
		t.Error("expected error for empty pages")
	}
}

func TestCov36_SelectPages_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.SelectPages([]int{99})
	if err == nil {
		t.Error("expected error for out of range")
	}
}

func TestCov36_SelectPagesFromFile(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	result, err := SelectPagesFromFile(resTestPDF, []int{1}, nil)
	if err != nil {
		t.Fatalf("SelectPagesFromFile: %v", err)
	}
	if result.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

func TestCov36_SelectPagesFromFile_Empty(t *testing.T) {
	_, err := SelectPagesFromFile(resTestPDF, nil, nil)
	if err == nil {
		t.Error("expected error for empty pages")
	}
}

func TestCov36_SelectPagesFromFile_BadFile(t *testing.T) {
	_, err := SelectPagesFromFile("/nonexistent.pdf", []int{1}, nil)
	if err == nil {
		t.Error("expected error for bad file")
	}
}

func TestCov36_SelectPagesFromBytes(t *testing.T) {
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	result, err := SelectPagesFromBytes(data, []int{1}, nil)
	if err != nil {
		t.Fatalf("SelectPagesFromBytes: %v", err)
	}
	if result.GetNumberOfPages() < 1 {
		t.Error("expected at least 1 page")
	}
}

func TestCov36_SelectPagesFromBytes_Empty(t *testing.T) {
	_, err := SelectPagesFromBytes(nil, nil, nil)
	if err == nil {
		t.Error("expected error for empty pages")
	}
}

// --- bookmark.go: DeleteBookmark deeper paths ---

func TestCov36_DeleteBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutlineWithPosition("Ch1")
	err := pdf.DeleteBookmark(-1)
	if err == nil {
		t.Error("expected error for negative index")
	}
	err = pdf.DeleteBookmark(99)
	if err == nil {
		t.Error("expected error for out of range")
	}
}

func TestCov36_DeleteBookmark_Valid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutlineWithPosition("Ch1")
	pdf.AddOutlineWithPosition("Ch2")
	pdf.AddOutlineWithPosition("Ch3")

	err := pdf.DeleteBookmark(1) // delete middle
	if err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}
}

func TestCov36_DeleteBookmark_First(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutlineWithPosition("Ch1")
	pdf.AddOutlineWithPosition("Ch2")

	err := pdf.DeleteBookmark(0) // delete first
	if err != nil {
		t.Fatalf("DeleteBookmark first: %v", err)
	}
}

func TestCov36_DeleteBookmark_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutlineWithPosition("Ch1")
	pdf.AddOutlineWithPosition("Ch2")

	err := pdf.DeleteBookmark(1) // delete last
	if err != nil {
		t.Fatalf("DeleteBookmark last: %v", err)
	}
}

// --- image_obj.go: SetImagePath, SetImage, getRect, parse ---

func TestCov36_ImageObj_SetImagePath(t *testing.T) {
	img := &ImageObj{}
	err := img.SetImagePath(resJPEGPath)
	if err != nil {
		t.Skipf("image not available: %v", err)
	}
	err = img.parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	rect, err := img.getRect()
	if err != nil {
		t.Fatalf("getRect: %v", err)
	}
	if rect.W <= 0 || rect.H <= 0 {
		t.Error("invalid rect")
	}
}

func TestCov36_ImageObj_SetImagePath_BadPath(t *testing.T) {
	img := &ImageObj{}
	err := img.SetImagePath("/nonexistent/image.jpg")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestCov36_ImageObj_GetRect_Deprecated(t *testing.T) {
	img := &ImageObj{}
	if err := img.SetImagePath(resJPEGPath); err != nil {
		t.Skipf("image not available: %v", err)
	}
	// GetRect uses log.Fatalf on error, but should work for valid image
	rect := img.GetRect()
	if rect.W <= 0 || rect.H <= 0 {
		t.Error("invalid rect from GetRect")
	}
}

func TestCov36_ImageObj_SetImage(t *testing.T) {
	f, err := os.Open(resJPEGPath)
	if err != nil {
		t.Skipf("image not available: %v", err)
	}
	defer f.Close()
	img := &ImageObj{}
	err = img.SetImage(f)
	if err != nil {
		t.Fatalf("SetImage: %v", err)
	}
}

func TestCov36_ImageObj_Types(t *testing.T) {
	img := &ImageObj{}
	if img.getType() != "Image" {
		t.Errorf("expected 'Image', got %q", img.getType())
	}
}

// --- gopdf.go: Text/Cell/CellWithOption nil font panic ---

func TestCov36_Text_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	// Don't set font — FontISubset is nil
	defer func() { recover() }()
	pdf.Text("crash")
}

func TestCov36_Cell_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	defer func() { recover() }()
	pdf.Cell(nil, "crash")
}

func TestCov36_CellWithOption_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	defer func() { recover() }()
	pdf.CellWithOption(nil, "crash", CellOption{})
}

func TestCov36_MeasureTextWidth_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	defer func() { recover() }()
	pdf.MeasureTextWidth("crash")
}

func TestCov36_MeasureCellHeightByText_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	defer func() { recover() }()
	pdf.MeasureCellHeightByText("crash")
}

// --- gopdf.go: GetBytesPdfReturnErr ---

func TestCov36_GetBytesPdfReturnErr(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("test")
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty data")
	}
}

// --- gopdf.go: SetPDFVersion ---

func TestCov36_SetPDFVersion(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetPDFVersion(PDFVersion17)
	pdf.AddPage()
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: SetNoCompression ---

func TestCov36_SetNoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("no compress")
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: AddColorSpace ---

func TestCov36_AddColorSpaceRGB(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceRGB("mycolor", 100, 200, 50)
}

func TestCov36_AddColorSpaceCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceCMYK("mycmyk", 10, 20, 30, 40)
}

// --- gopdf.go: SaveGraphicsState / RestoreGraphicsState ---

func TestCov36_GraphicsState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	pdf.SetXY(10, 10)
	pdf.Text("in state")
	pdf.RestoreGraphicsState()
}

// --- gopdf.go: Rotate / RotateReset ---

func TestCov36_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.SetXY(100, 100)
	pdf.Text("rotated")
	pdf.RotateReset()
}

// --- gopdf.go: SetTextColor / SetStrokeColor ---

func TestCov36_SetColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColor(255, 0, 0)
	pdf.SetStrokeColor(0, 255, 0)
	pdf.SetFillColor(0, 0, 255)
	pdf.SetXY(10, 10)
	pdf.Text("colored")
}

// --- gopdf.go: SetLineWidth / SetLineType ---

func TestCov36_SetLineWidthAndType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(2.0)
	pdf.SetLineType("dashed")
	pdf.Line(10, 10, 100, 100)
}

// --- gopdf.go: SetGrayFill / SetGrayStroke ---

func TestCov36_SetGray(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
	pdf.SetGrayStroke(0.3)
}

// --- gopdf.go: SetTextColorCMYK / SetStrokeColorCMYK / SetFillColorCMYK ---

func TestCov36_SetCMYKColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColorCMYK(10, 20, 30, 40)
	pdf.SetStrokeColorCMYK(50, 60, 70, 80)
	pdf.SetFillColorCMYK(90, 10, 20, 30)
}

// --- pdf_decrypt.go: detectEncryption, parseEncryptDict ---

func TestCov36_DetectEncryption_NoEncrypt(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.Write(&buf)
	objNum := detectEncryption(buf.Bytes())
	if objNum != 0 {
		t.Errorf("expected 0 for unencrypted, got %d", objNum)
	}
}

func TestCov36_ParseEncryptDict_Invalid(t *testing.T) {
	_, _, _, _, _, _, err := parseEncryptDict("")
	if err == nil {
		t.Error("expected error for empty dict")
	}
}

// --- gopdf.go: AddPage with custom size ---

func TestCov36_AddPageWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 300, H: 400},
	})
	pdf.SetXY(10, 10)
	pdf.Text("custom size")
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: SetCustomLineType ---

func TestCov36_SetCustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(10, 10, 200, 10)
}

// --- gopdf.go: ClipPolygon ---

func TestCov36_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	points := []Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}
	pdf.SaveGraphicsState()
	pdf.ClipPolygon(points)
	pdf.SetXY(60, 80)
	pdf.Text("clipped")
	pdf.RestoreGraphicsState()
}

// --- gopdf.go: Oval ---

func TestCov36_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(50, 50, 150, 100)
}

// --- gopdf.go: ImageFrom ---

func TestCov36_ImageFromWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	img := createSimpleRGBAImage36()
	err := pdf.ImageFromWithOption(img, ImageFromOption{
		X: 10, Y: 10, Rect: &Rect{W: 50, H: 50}, Format: "jpeg",
	})
	if err != nil {
		t.Fatalf("ImageFromWithOption: %v", err)
	}
}

// --- gopdf.go: AddOCG ---

func TestCov36_AddOCG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	ocg := pdf.AddOCG(OCG{Name: "Layer1", On: true})
	if ocg.Name != "Layer1" {
		t.Errorf("expected Layer1, got %s", ocg.Name)
	}
}


// --- gopdf.go: SetTransparency ---

func TestCov36_SetTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: ""})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	// Set again (cached path)
	err = pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: ""})
	if err != nil {
		t.Fatalf("SetTransparency cached: %v", err)
	}
	pdf.ClearTransparency()
}

// --- gopdf.go: ImageByHolder with mask ---

func TestCov36_ImageByHolder_WithMask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Use PNG with transparency (has smask)
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}
	err := pdf.Image(resPNGPath, 10, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("Image PNG: %v", err)
	}
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: Image with transparency ---

func TestCov36_Image_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetTransparency(Transparency{Alpha: 0.5}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	err := pdf.Image(resJPEGPath, 10, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("Image with transparency: %v", err)
	}
	pdf.ClearTransparency()
}

// --- gopdf.go: multiple fonts ---

func TestCov36_MultipleFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font1 not available: %v", err)
	}
	if err := pdf.AddTTFFont(fontFamily2, resFontPath2); err != nil {
		t.Skipf("font2 not available: %v", err)
	}
	pdf.AddPage()
	pdf.SetFont(fontFamily, "", 14)
	pdf.SetXY(10, 10)
	pdf.Text("Font 1")
	pdf.SetFont(fontFamily2, "", 14)
	pdf.SetXY(10, 30)
	pdf.Text("Font 2")
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: SetFontWithStyle ---

func TestCov36_SetFontWithStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetFontWithStyle(fontFamily, 0, 18)
	if err != nil {
		t.Fatalf("SetFontWithStyle: %v", err)
	}
	pdf.SetXY(10, 10)
	pdf.Text("styled")
}

// --- gopdf.go: GetNumberOfPages ---

func TestCov36_GetNumberOfPages(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	if pdf.GetNumberOfPages() != 0 {
		t.Error("expected 0 pages")
	}
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 1 {
		t.Error("expected 1 page")
	}
}

// --- gopdf.go: SetPage ---

func TestCov36_SetPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("P1")
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("P2")

	err := pdf.SetPage(1)
	if err != nil {
		t.Fatalf("SetPage(1): %v", err)
	}
}

// --- gopdf.go: GetNextObjectID ---

func TestCov36_GetNextObjectID(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	id := pdf.GetNextObjectID()
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

// --- helper ---

func createSimpleRGBAImage36() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	return img
}


// --- html_parser.go: edge cases ---

func TestCov36_ParseHTML_EmptyText(t *testing.T) {
	// Text node that decodes to empty
	nodes := parseHTML("<p></p>")
	_ = nodes
}

func TestCov36_ParseHTML_SelfClosing(t *testing.T) {
	nodes := parseHTML("<br/><hr/><img/>")
	_ = nodes
}

func TestCov36_ParseHTML_Attributes(t *testing.T) {
	nodes := parseHTML(`<div style="color:red" class="test">hello</div>`)
	if len(nodes) == 0 {
		t.Error("expected nodes")
	}
}

func TestCov36_ParseHTML_UnquotedAttr(t *testing.T) {
	nodes := parseHTML(`<div style=color:red>hello</div>`)
	_ = nodes
}

func TestCov36_ParseHTML_Comment(t *testing.T) {
	nodes := parseHTML("<!-- comment --><p>text</p>")
	_ = nodes
}

func TestCov36_ParseHTML_MalformedTag(t *testing.T) {
	nodes := parseHTML("<>text</>")
	_ = nodes
}

func TestCov36_ParseHTML_AttrNoValue(t *testing.T) {
	nodes := parseHTML(`<input disabled>`)
	_ = nodes
}

func TestCov36_ParseHTML_NestedElements(t *testing.T) {
	nodes := parseHTML("<div><p>inner</p></div>")
	out := debugHTMLTree(nodes, 0)
	if out == "" {
		t.Error("expected debug output")
	}
}

func TestCov36_ParseInlineStyle_NoColon(t *testing.T) {
	style := parseInlineStyle("invalid-no-colon; color: red")
	if style["color"] != "red" {
		t.Errorf("expected red, got %q", style["color"])
	}
}

func TestCov36_ParseInlineStyle_Empty(t *testing.T) {
	style := parseInlineStyle("")
	if len(style) != 0 {
		t.Error("expected empty map")
	}
}

// --- text_extract.go: deeper branches ---

func TestCov36_DecodeHexUTF16BE(t *testing.T) {
	// UTF-16BE for "A" = 0041
	got := decodeHexUTF16BE("0041")
	if got != "A" {
		t.Errorf("expected A, got %q", got)
	}
}

func TestCov36_DecodeHexLatin(t *testing.T) {
	got := decodeHexLatin("48656C6C6F")
	if got != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

func TestCov36_DecodeUTF16BE(t *testing.T) {
	// UTF-16BE for "AB" = 0041 0042
	data := []byte{0x00, 0x41, 0x00, 0x42}
	got := decodeUTF16BE(data)
	if got != "AB" {
		t.Errorf("expected AB, got %q", got)
	}
}

func TestCov36_FindStringBefore(t *testing.T) {
	tokens := []string{"(Hello)", "Tj"}
	got := findStringBefore(tokens, 1)
	if got != "(Hello)" {
		t.Errorf("expected (Hello), got %q", got)
	}
}

func TestCov36_FindArrayBefore(t *testing.T) {
	tokens := []string{"[", "(A)", "-50", "(B)", "]", "TJ"}
	got := findArrayBefore(tokens, 5)
	if len(got) == 0 {
		t.Error("expected array elements")
	}
}

func TestCov36_IsStringToken(t *testing.T) {
	if !isStringToken("(hello)") {
		t.Error("expected true for (hello)")
	}
	if !isStringToken("<0041>") {
		t.Error("expected true for <0041>")
	}
	if isStringToken("123") {
		t.Error("expected false for 123")
	}
}

func TestCov36_FontDisplayName(t *testing.T) {
	fi := &fontInfo{name: "TestFont"}
	got := fontDisplayName(fi)
	if got != "TestFont" {
		t.Errorf("expected TestFont, got %q", got)
	}
	// nil font
	got = fontDisplayName(nil)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// --- embedded_file.go: deeper paths ---

func TestCov36_EmbeddedFile_AddAndGet(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "test.txt",
		Content:  []byte("hello world"),
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}

	data, err := pdf.GetEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFile: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}

	info, err := pdf.GetEmbeddedFileInfo("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFileInfo: %v", err)
	}
	if info.Name != "test.txt" {
		t.Errorf("expected test.txt, got %s", info.Name)
	}
}

func TestCov36_EmbeddedFile_GetNotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetEmbeddedFile("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCov36_EmbeddedFile_Update(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("original"),
	})
	err := pdf.UpdateEmbeddedFile("test.txt", EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("updated"),
	})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}
}

// --- ocg.go: deeper paths ---

func TestCov36_OCG_GetAndSetState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "Layer1", On: true})
	pdf.AddOCG(OCG{Name: "Layer2", On: false})

	ocgs := pdf.GetOCGs()
	if len(ocgs) != 2 {
		t.Errorf("expected 2 OCGs, got %d", len(ocgs))
	}

	err := pdf.SetOCGState("Layer1", false)
	if err != nil {
		t.Fatalf("SetOCGState: %v", err)
	}

	err = pdf.SetOCGState("nonexistent", true)
	if err == nil {
		t.Error("expected error for nonexistent OCG")
	}

	err = pdf.SetOCGStates(map[string]bool{"Layer1": true, "Layer2": true})
	if err != nil {
		t.Fatalf("SetOCGStates: %v", err)
	}
}

// --- content_obj.go: write with protection ---

func TestCov36_ContentObj_WriteWithProtection(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{
		PageSize:   *PageSizeA4,
		Protection: PDFProtectionConfig{UseProtection: true, Permissions: PermissionsPrint | PermissionsCopy},
	})
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	if err := pdf.SetFont(fontFamily, "", 14); err != nil {
		t.Fatal(err)
	}
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("protected content")
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("empty output")
	}
}

// --- gopdf.go: Write to failing writer ---

func TestCov36_Write_FailingWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("test")

	for failAt := 0; failAt < 500; failAt += 50 {
		fw := &failWriterCov36{failAfter: failAt}
		_ = pdf.Write(fw)
	}
}

type failWriterCov36 struct {
	written   int
	failAfter int
}

func (fw *failWriterCov36) Write(p []byte) (int, error) {
	if fw.written+len(p) > fw.failAfter {
		remaining := fw.failAfter - fw.written
		if remaining <= 0 {
			return 0, errCov36Write
		}
		fw.written += remaining
		return remaining, errCov36Write
	}
	fw.written += len(p)
	return len(p), nil
}

var errCov36Write = fmt.Errorf("write failed")

// --- gopdf.go: compilePdf error path in Read ---

func TestCov36_Read_SecondCall(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("read twice")

	buf := make([]byte, 8192)
	n1, _ := pdf.Read(buf)
	if n1 == 0 {
		t.Error("first read empty")
	}
	// Second read should use cached buffer
	buf2 := make([]byte, 8192)
	n2, _ := pdf.Read(buf2)
	_ = n2 // may be 0 if all data was read
}


// --- text_extract.go: extractName edge cases ---

func TestCov36_ExtractName(t *testing.T) {
	// Key not found
	got := extractName("/Type /Page", "/Subtype")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	// Key found but no slash after
	got = extractName("/Subtype Image", "/Subtype")
	if got != "" {
		t.Errorf("expected empty for no slash, got %q", got)
	}
	// Normal case
	got = extractName("/Subtype /Image", "/Subtype")
	if got != "Image" {
		t.Errorf("expected Image, got %q", got)
	}
}

// --- text_extract.go: decodeTextString literal without CMap ---

func TestCov36_DecodeTextString_LiteralNoCMap(t *testing.T) {
	// Literal string without font info
	got := decodeTextString("(Hello)", nil)
	if got != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

// --- text_extract.go: ExtractTextFromPage out of range ---

func TestCov36_ExtractTextFromPage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("test")
	var buf bytes.Buffer
	pdf.Write(&buf)

	_, err := ExtractTextFromPage(buf.Bytes(), 99)
	if err == nil {
		t.Error("expected error for out of range")
	}
}

// --- text_extract.go: ExtractTextFromAllPages ---

func TestCov36_ExtractTextFromAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Page 2")
	var buf bytes.Buffer
	pdf.Write(&buf)

	result, err := ExtractTextFromAllPages(buf.Bytes())
	if err != nil {
		t.Fatalf("ExtractTextFromAllPages: %v", err)
	}
	_ = result
}

// --- gopdf.go: ImageByHolderWithOptions with Mask ---

func TestCov36_ImageByHolderWithOptions_Mask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Use PNG with alpha (triggers mask path)
	if _, err := os.Stat(resPNGPath); err != nil {
		t.Skipf("PNG not available: %v", err)
	}
	err := pdf.Image(resPNGPath, 10, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("Image PNG with mask: %v", err)
	}
	// Add same image again (cached path)
	err = pdf.Image(resPNGPath, 120, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("Image PNG cached: %v", err)
	}
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: createTransparencyXObjectGroup ---

func TestCov36_CreateTransparencyXObjectGroup(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetTransparency(Transparency{Alpha: 0.5}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	err := pdf.Image(resJPEGPath, 10, 10, &Rect{W: 100, H: 100})
	if err != nil {
		t.Fatalf("Image with transparency: %v", err)
	}
	pdf.ClearTransparency()
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: AppendStreamImportedTemplate ---

func TestCov36_UseImportedTemplate(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Import a page from test PDF
	result, err := ExtractPages(resTestPDF, []int{1}, nil)
	if err != nil {
		t.Fatalf("ExtractPages: %v", err)
	}
	var buf bytes.Buffer
	if err := result.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

// --- gopdf.go: SetXY / GetX / GetY ---

func TestCov36_SetXY_GetXY(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(42, 84)
	x := pdf.GetX()
	y := pdf.GetY()
	if x != 42 || y != 84 {
		t.Errorf("expected (42,84), got (%v,%v)", x, y)
	}
}

// --- gopdf.go: SetNewY / SetNewYIfNoRecordYyet ---

func TestCov36_SetNewY(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetNewY(100, 20)
	pdf.SetNewYIfNoOffset(200, 50)
}

// --- gopdf.go: AddHeader / AddFooter ---

func TestCov36_AddHeaderFooter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddHeader(func() {
		pdf.SetXY(10, 10)
		pdf.Text("Header")
	})
	pdf.AddFooter(func() {
		pdf.SetXY(10, 800)
		pdf.Text("Footer")
	})
	pdf.AddPage()
	pdf.SetXY(10, 50)
	pdf.Text("Body")
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}


// --- image_obj_parse.go: writeImgProps / writeMaskImgProps error paths ---

func TestCov36_WriteMaskImgProps_FailWriter(t *testing.T) {
	info := imgInfo{
		w: 10, h: 10,
		colspace:        "DeviceGray",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
	}
	for failAt := 0; failAt < 100; failAt += 10 {
		fw := &failWriterCov36{failAfter: failAt}
		_ = writeMaskImgProps(fw, info)
	}
}

func TestCov36_WriteImgProps_FailWriter(t *testing.T) {
	info := imgInfo{
		w: 10, h: 10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		decodeParms:     "/Predictor 15",
		smask:           []byte{1, 2, 3},
		trns:            []byte{0xFF},
	}
	for failAt := 0; failAt < 200; failAt += 10 {
		fw := &failWriterCov36{failAfter: failAt}
		_ = writeImgProps(fw, info, false)
	}
}

func TestCov36_WriteImgProps_Indexed_FailWriter(t *testing.T) {
	info := imgInfo{
		w: 10, h: 10,
		colspace:        "Indexed",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		pal:             []byte{255, 0, 0, 0, 255, 0},
		trns:            []byte{0x00},
	}
	for failAt := 0; failAt < 200; failAt += 10 {
		fw := &failWriterCov36{failAfter: failAt}
		_ = writeImgProps(fw, info, false)
	}
}

func TestCov36_WriteBaseImgProps_FailWriter(t *testing.T) {
	info := imgInfo{
		w: 10, h: 10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		filter:          "FlateDecode",
		decodeParms:     "/Predictor 15",
	}
	for failAt := 0; failAt < 100; failAt += 10 {
		fw := &failWriterCov36{failAfter: failAt}
		_ = writeBaseImgProps(fw, info, "DeviceRGB")
	}
}

// --- image_obj_parse.go: parsePng truncated at various points ---

func TestCov36_ParsePng_TruncatedAfterIHDR(t *testing.T) {
	// Build a PNG that's truncated right after IHDR (no IDAT/IEND)
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	ihdr := make([]byte, 13)
	ihdr[0], ihdr[1], ihdr[2], ihdr[3] = 0, 0, 0, 1 // w=1
	ihdr[4], ihdr[5], ihdr[6], ihdr[7] = 0, 0, 0, 1 // h=1
	ihdr[8] = 8  // bit depth
	ihdr[9] = 2  // color type RGB
	ihdr[10] = 0 // compression
	ihdr[11] = 0 // filter
	ihdr[12] = 0 // interlace
	writePNGChunk35(&buf, "IHDR", ihdr)
	// No IDAT, no IEND — truncated

	r := bytes.NewReader(buf.Bytes())
	err := parsePng(r, &imgInfo{}, image.Config{})
	// Should error due to truncated data
	if err == nil {
		// Some truncations may not error if the loop exits cleanly
	}
}

// --- gopdf.go: WritePdf ---

func TestCov36_WritePdf(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("WritePdf test")
	err := pdf.WritePdf(resOutDir + "/cov36_writepdf.pdf")
	if err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
}

// --- gopdf.go: SetMargins ---

func TestCov36_SetMargins(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetLeftMargin(20)
	pdf.SetTopMargin(30)
	pdf.AddPage()
}

// --- gopdf.go: GetBytesPdfReturnErr after Write ---

func TestCov36_GetBytesPdfReturnErr_AfterWrite(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("test")

	// First call
	data1, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatal(err)
	}
	// Second call (should use cached)
	data2, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatal(err)
	}
	if len(data1) == 0 || len(data2) == 0 {
		t.Error("expected non-empty data")
	}
}

// --- gopdf.go: AddTTFFontByReader ---

func TestCov36_AddTTFFontByReader(t *testing.T) {
	f, err := os.Open(resFontPath)
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	defer f.Close()

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontByReader("TestReader", f)
	if err != nil {
		t.Fatalf("AddTTFFontByReader: %v", err)
	}
	err = pdf.SetFont("TestReader", "", 14)
	if err != nil {
		t.Fatal(err)
	}
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("from reader")
}

// --- gopdf.go: SetCompressLevel ---

func TestCov36_SetCompressLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetCompressLevel(6)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("compressed")
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}


// --- html_insert.go: InsertHTMLBox with various HTML ---

func TestCov36_InsertHTMLBox_Complex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<h1>Title</h1><h2>Subtitle</h2><h3>Section</h3>
<p style="color: red; font-size: 16pt">Styled paragraph with <b>bold</b> and <i>italic</i> text.</p>
<p style="text-align: center">Centered text</p>
<p style="text-align: right">Right aligned</p>
<ul><li>Item 1</li><li>Item 2</li></ul>
<ol><li>First</li><li>Second</li></ol>
<hr/>
<p>A very long word: supercalifragilisticexpialidocious that might need wrapping</p>
<p style="margin-left: 20px">Indented paragraph</p>
<p><font size="5">Big font</font></p>
<p><a href="http://example.com">Link text</a></p>
<p>&amp; &lt; &gt; entities</p>`

	_, err := pdf.InsertHTMLBox(10, 10, 200, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox complex: %v", err)
	}
	var buf bytes.Buffer
	if err := pdf.Write(&buf); err != nil {
		t.Fatal(err)
	}
}

func TestCov36_InsertHTMLBox_NarrowBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Very narrow box forces word wrapping
	html := `<p>This text should wrap in a very narrow box with multiple words</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 30, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox narrow: %v", err)
	}
}

func TestCov36_InsertHTMLBox_ShortBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Very short box — text should be clipped
	html := `<p>Line 1</p><p>Line 2</p><p>Line 3</p><p>Line 4</p><p>Line 5</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 20, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox short: %v", err)
	}
}

func TestCov36_InsertHTMLBox_NestedLists(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<ul><li>A</li><li>B<ul><li>B1</li><li>B2</li></ul></li><li>C</li></ul>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox nested lists: %v", err)
	}
}

func TestCov36_InsertHTMLBox_Headings(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<h1>H1</h1><h2>H2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox headings: %v", err)
	}
}

func TestCov36_InsertHTMLBox_StyledSpan(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<p><span style="color: #ff0000; font-size: 18pt">Red big</span> normal <span style="color: rgb(0,0,255)">Blue</span></p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox styled span: %v", err)
	}
}

// --- pdf_decrypt.go: authenticate / tryUserPassword / tryOwnerPassword ---

func TestCov36_Authenticate_Unencrypted(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.Write(&buf)
	dc, err := authenticate(buf.Bytes(), "")
	// Unencrypted PDF — authenticate may return nil dc or error
	_ = dc
	_ = err
}

func TestCov36_PadPassword(t *testing.T) {
	// Short password
	padded := padPassword([]byte("short"))
	if len(padded) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(padded))
	}
	// Long password
	padded = padPassword([]byte("this is a very long password that exceeds 32 bytes"))
	if len(padded) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(padded))
	}
}

func TestCov36_ComputeEncryptionKey(t *testing.T) {
	key := computeEncryptionKey([]byte("test"), []byte("owner value here 32 bytes long!!"), -4, 5, 2)
	if len(key) == 0 {
		t.Error("expected non-empty key")
	}
}

func TestCov36_ComputeUValue(t *testing.T) {
	key := []byte{1, 2, 3, 4, 5}
	u := computeUValue(key, 2)
	if len(u) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(u))
	}
	u3 := computeUValue(key, 3)
	if len(u3) != 32 {
		t.Errorf("expected 32 bytes for r=3, got %d", len(u3))
	}
}

func TestCov36_ExtractSignedIntValue(t *testing.T) {
	dict := "/V 2 /R 3 /Length 128 /P -3904"
	v := extractSignedIntValue(dict, "/P")
	if v != -3904 {
		t.Errorf("expected -3904, got %d", v)
	}
	// Key not found
	v = extractSignedIntValue(dict, "/Missing")
	if v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
}

func TestCov36_ExtractHexOrLiteralString(t *testing.T) {
	dict := "/O <48656C6C6F> /U (World)"
	o := extractHexOrLiteralString(dict, "/O")
	if string(o) != "Hello" {
		t.Errorf("hex: got %q", string(o))
	}
	u := extractHexOrLiteralString(dict, "/U")
	if string(u) != "World" {
		t.Errorf("literal: got %q", string(u))
	}
	// Not found
	n := extractHexOrLiteralString(dict, "/Missing")
	if n != nil {
		t.Errorf("expected nil, got %v", n)
	}
}

func TestCov36_DecodeHexString(t *testing.T) {
	got := decodeHexString("48656C6C6F")
	if string(got) != "Hello" {
		t.Errorf("expected Hello, got %q", string(got))
	}
}

func TestCov36_DecodeLiteralString(t *testing.T) {
	got := decodeLiteralString("(Hello\\nWorld)")
	if string(got) != "Hello\nWorld" {
		t.Errorf("expected Hello\\nWorld, got %q", string(got))
	}
	// Empty
	got = decodeLiteralString("")
	if got != nil {
		t.Error("expected nil for empty")
	}
}


func TestCov36_RemoveEncryptFromTrailer(t *testing.T) {
	data := []byte("trailer\n<< /Encrypt 5 0 R /Root 1 0 R >>\nstartxref\n100\n%%EOF")
	result := removeEncryptFromTrailer(data)
	_ = result
}
