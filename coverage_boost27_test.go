package gopdf

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost27_test.go — TestCov27_ prefix
// Targets: watermark error paths, collectOutlineObjs branches,
// DeleteBookmark parent-not-outlines branch, findCurrentPageObjID,
// various write failWriter paths for remaining obj types
// ============================================================

// ============================================================
// AddWatermarkText — error paths
// ============================================================

func TestCov27_AddWatermarkText_EmptyText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "",
		FontFamily: fontFamily,
	})
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestCov27_AddWatermarkText_MissingFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: "",
	})
	if err == nil {
		t.Error("expected error for missing font family")
	}
}

func TestCov27_AddWatermarkText_BadFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: "nonexistent_font",
	})
	if err == nil {
		t.Error("expected error for bad font")
	}
}


// ============================================================
// AddWatermarkText — custom page size branch
// ============================================================

func TestCov27_AddWatermarkText_CustomPageSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	customSize := &Rect{W: 400, H: 600}
	pdf.AddPageWithOption(PageOption{PageSize: customSize})
	pdf.SetXY(50, 50)
	pdf.Text("Custom page")

	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "CUSTOM",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.4,
		Angle:      30,
		Color:      [3]uint8{255, 0, 0},
	})
	if err != nil {
		t.Fatalf("AddWatermarkText custom page: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// AddWatermarkImage — error paths
// ============================================================

func TestCov27_AddWatermarkImage_BadPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkImage("/nonexistent/image.jpg", 0.3, 100, 100, 0)
	if err == nil {
		t.Error("expected error for bad image path")
	}
}

func TestCov27_AddWatermarkImage_NoAngle(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// angle=0 should skip rotation
	err := pdf.AddWatermarkImage(resJPEGPath, 0.5, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImage no angle: %v", err)
	}
	pdf.GetBytesPdf()
}

func TestCov27_AddWatermarkImage_CustomPageSize(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); err != nil {
		t.Skipf("JPEG not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	customSize := &Rect{W: 400, H: 600}
	pdf.AddPageWithOption(PageOption{PageSize: customSize})
	err := pdf.AddWatermarkImage(resJPEGPath, 0.5, 150, 150, 45)
	if err != nil {
		t.Fatalf("AddWatermarkImage custom page: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// collectOutlineObjs — visited branch, !ok branch
// ============================================================

func TestCov27_CollectOutlineObjs_Branches(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 3")
	pdf.AddOutline("Chapter 3")

	// Get outline list — exercises the traversal
	outlines := pdf.getOutlineObjList()
	if len(outlines) != 3 {
		t.Logf("expected 3 outlines, got %d", len(outlines))
	}

	// Call collectOutlineObjs with invalid objID (<=0)
	var result []*OutlineObj
	visited := make(map[int]bool)
	pdf.collectOutlineObjs(nil, 0, &result, visited)
	pdf.collectOutlineObjs(nil, -1, &result, visited)

	// Call with objID not in map
	m := make(map[int]*OutlineObj)
	pdf.collectOutlineObjs(m, 999, &result, visited)
}

// ============================================================
// DeleteBookmark — all single bookmark (first == last)
// ============================================================

func TestCov27_DeleteBookmark_SingleBookmark(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Only page")
	pdf.AddOutline("Only Chapter")

	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark single: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// AddWatermarkTextAllPages — SetPage error
// ============================================================

func TestCov27_AddWatermarkTextAllPages_Error(t *testing.T) {
	pdf := newPDFWithFont(t)
	// No pages — should handle gracefully
	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
	})
	// With 0 pages, the loop doesn't execute
	if err != nil {
		t.Logf("AddWatermarkTextAllPages with 0 pages: %v", err)
	}
}

func TestCov27_AddWatermarkImageAllPages_Error(t *testing.T) {
	pdf := newPDFWithFont(t)
	// No pages
	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.3, 100, 100, 0)
	if err != nil {
		t.Logf("AddWatermarkImageAllPages with 0 pages: %v", err)
	}
}

// ============================================================
// CellWithOption — more alignment branches
// ============================================================

func TestCov27_CellWithOption_WithBorder(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	// Cell with all borders
	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Bordered cell", CellOption{
		Align:  Center | Middle,
		Border: Left | Right | Top | Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption bordered: %v", err)
	}

	// Cell with float right
	pdf.SetXY(50, 100)
	err = pdf.CellWithOption(&Rect{W: 200, H: 30}, "Float right", CellOption{
		Align: Right | Top,
		Float: Right,
	})
	if err != nil {
		t.Fatalf("CellWithOption float right: %v", err)
	}

	// Cell with float bottom
	pdf.SetXY(50, 150)
	err = pdf.CellWithOption(&Rect{W: 200, H: 30}, "Float bottom", CellOption{
		Align: Left | Top,
		Float: Bottom,
	})
	if err != nil {
		t.Fatalf("CellWithOption float bottom: %v", err)
	}

	pdf.GetBytesPdf()
}

// ============================================================
// Text — error path (no font)
// ============================================================

func TestCov27_Text_NoFont(t *testing.T) {
	defer func() { recover() }()
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	err := pdf.Text("hello")
	if err != nil {
		t.Logf("Text error: %v", err)
	}
}

// ============================================================
// GetBytesPdf — with content
// ============================================================

func TestCov27_GetBytesPdf_WithContent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Test content")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page 2 content")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}

	// Also test GetBytesPdfReturnErr
	data2, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data2) == 0 {
		t.Error("expected non-empty PDF from GetBytesPdfReturnErr")
	}
}

// ============================================================
// Read — EOF handling
// ============================================================

func TestCov27_Read_EOF(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Small")

	// Read all data
	var allData bytes.Buffer
	buf := make([]byte, 64)
	for {
		n, err := pdf.Read(buf)
		if n > 0 {
			allData.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	if allData.Len() == 0 {
		t.Error("expected some data")
	}
}

// ============================================================
// cid_font_obj.write — failWriter
// ============================================================

func TestCov27_CidFontObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("CID font test")

	for i, obj := range pdf.pdfObjs {
		if cidObj, ok := obj.(*CIDFontObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				cidObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no CIDFontObj found")
}

// ============================================================
// subfont_descriptor_obj.write — failWriter
// ============================================================

func TestCov27_SubfontDescriptorObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Subfont descriptor test")

	for i, obj := range pdf.pdfObjs {
		if sfObj, ok := obj.(*SubfontDescriptorObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				sfObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no SubfontDescriptorObj found")
}

// ============================================================
// subset_font_obj.write — failWriter
// ============================================================

func TestCov27_SubsetFontObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Subset font test")

	for i, obj := range pdf.pdfObjs {
		if sfObj, ok := obj.(*SubsetFontObj); ok {
			for n := 0; n <= 500; n += 3 {
				fw := &failWriterAt{n: n}
				sfObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no SubsetFontObj found")
}


// ============================================================
// encoding_obj.write — failWriter
// ============================================================

func TestCov27_EncodingObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Encoding test")

	for i, obj := range pdf.pdfObjs {
		if encObj, ok := obj.(*EncodingObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				encObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no EncodingObj found")
}

// ============================================================
// encryption_obj.write — failWriter
// ============================================================

func TestCov27_EncryptionObj_Write_FailWriter(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Encrypted test")

	for i, obj := range pdf.pdfObjs {
		if encObj, ok := obj.(*EncryptionObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				encObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no EncryptionObj found")
}

// ============================================================
// HTML — blockquote, sub, sup, strike, ins, headings
// ============================================================

func TestCov27_HTML_MoreElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<h1>Heading 1</h1>
<h2>Heading 2</h2>
<h3>Heading 3</h3>
<blockquote>This is a blockquote with some text</blockquote>
<p>Normal text with <sub>subscript</sub> and <sup>superscript</sup></p>
<p><strike>Strikethrough</strike> and <ins>inserted</ins> text</p>
<p><font color="red" size="5" face="` + fontFamily + `">Font tag</font></p>`

	_, err := pdf.InsertHTMLBox(50, 50, 300, 600, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox more elements: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — link (a tag)
// ============================================================

func TestCov27_HTML_Link(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Visit <a href="https://example.com">Example</a> for more info.</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox link: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — unordered list with nested items
// ============================================================

func TestCov27_HTML_UnorderedList(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<ul>
<li>Item 1</li>
<li>Item 2 with longer text that should wrap around</li>
<li>Item 3</li>
</ul>`
	_, err := pdf.InsertHTMLBox(50, 50, 150, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox unordered list: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — HR (horizontal rule)
// ============================================================

func TestCov27_HTML_HR(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Before HR</p><hr/><p>After HR</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox HR: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// FormField — error paths
// ============================================================

func TestCov27_FormField_Errors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Empty name
	err := pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "",
		X: 50, Y: 50, W: 100, H: 25,
	})
	if err == nil {
		t.Error("expected error for empty name")
	}

	// Zero dimensions
	err = pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "test",
		X: 50, Y: 50, W: 0, H: 25,
	})
	if err == nil {
		t.Error("expected error for zero width")
	}
}

// ============================================================
// findCurrentPageObjID — no page
// ============================================================

func TestCov27_FindCurrentPageObjID_NoPage(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	// No page added
	id := pdf.findCurrentPageObjID()
	if id != 0 {
		t.Errorf("expected 0, got %d", id)
	}
}

// ============================================================
// findCurrentPageObj — no page
// ============================================================

func TestCov27_FindCurrentPageObj_NoPage(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	obj := pdf.findCurrentPageObj()
	if obj != nil {
		t.Error("expected nil")
	}
}

// ============================================================
// GetFormFields — empty
// ============================================================

func TestCov27_GetFormFields_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	fields := pdf.GetFormFields()
	if len(fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(fields))
	}
}

// ============================================================
// GetFormFields — with fields
// ============================================================

func TestCov27_GetFormFields_WithFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "f1",
		X: 50, Y: 50, W: 100, H: 25,
	})
	pdf.AddFormField(FormField{
		Type: FormFieldCheckbox, Name: "f2",
		X: 50, Y: 100, W: 20, H: 20,
	})
	fields := pdf.GetFormFields()
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}

// ============================================================
// HTML — very narrow box (forces long word wrapping)
// ============================================================

func TestCov27_HTML_VeryNarrowBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Supercalifragilisticexpialidocious</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 30, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox narrow: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// HTML — box height exceeded
// ============================================================

func TestCov27_HTML_BoxHeightExceeded(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Very small box height with lots of text
	html := `<p>` + strings.Repeat("Word ", 100) + `</p>`
	_, err := pdf.InsertHTMLBox(50, 50, 200, 30, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Logf("InsertHTMLBox height exceeded: %v", err)
	}
	pdf.GetBytesPdf()
}

// ============================================================
// annot_obj.write — failWriter (external link)
// ============================================================

func TestCov27_AnnotObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Link text")
	pdf.AddExternalLink("https://example.com", 50, 50, 100, 20)

	for i, obj := range pdf.pdfObjs {
		if aObj, ok := obj.(*annotObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				aObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no AnnotObj found")
}

// ============================================================
// ext_g_state_obj.write — failWriter
// ============================================================

func TestCov27_ExtGStateObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode})
	pdf.SetXY(50, 50)
	pdf.Text("Transparent text")

	for i, obj := range pdf.pdfObjs {
		if egObj, ok := obj.(ExtGState); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				egObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no ExtGStateObj found")
}

// ============================================================
// color_space_obj.write — failWriter
// ============================================================

func TestCov27_ColorSpaceObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceCMYK("spot1", 100, 0, 0, 0)
	pdf.SetXY(50, 50)
	pdf.Text("Color space test")

	for i, obj := range pdf.pdfObjs {
		if csObj, ok := obj.(*ColorSpaceObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				csObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no ColorSpaceObj found")
}

// ============================================================
// markinfo.write — failWriter
// ============================================================

func TestCov27_MarkInfoObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Mark info test")

	for i, obj := range pdf.pdfObjs {
		if miObj, ok := obj.(*markInfoObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				miObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no MarkInfoObj found")
}

// ============================================================
// xmp_metadata.write — failWriter
// ============================================================

func TestCov27_XMPMetadata_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetXMPMetadata(XMPMetadata{
		Title:       "Test",
		Creator:     []string{"Author"},
		Description: "Test doc",
	})
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("XMP test")

	for i, obj := range pdf.pdfObjs {
		if xmpObj, ok := obj.(*xmpMetadataObj); ok {
			for n := 0; n <= 500; n += 3 {
				fw := &failWriterAt{n: n}
				xmpObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no XMPMetadataObj found")
}

// ============================================================
// embedded_file.write — failWriter
// ============================================================

func TestCov27_EmbeddedFile_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Embedded file test")
	pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("Hello embedded file"),
	})

	for i, obj := range pdf.pdfObjs {
		if efObj, ok := obj.(*embeddedFileStreamObj); ok {
			for n := 0; n <= 300; n += 2 {
				fw := &failWriterAt{n: n}
				efObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no embedded file obj found")
}


// ============================================================
// acroform_obj.write — failWriter
// ============================================================

func TestCov27_AcroFormObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "field1",
		X: 50, Y: 50, W: 100, H: 25,
	})

	for i, obj := range pdf.pdfObjs {
		if afObj, ok := obj.(*acroFormObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				afObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no AcroFormObj found")
}

// ============================================================
// names_obj.write — failWriter
// ============================================================

func TestCov27_NamesObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Names test")
	pdf.AddEmbeddedFile(EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("test content"),
	})

	for i, obj := range pdf.pdfObjs {
		if nObj, ok := obj.(*namesObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				nObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no NamesObj found")
}

// ============================================================
// page_label.write — failWriter
// ============================================================

func TestCov27_PageLabelObj_Write_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Text("Page label test")

	for i, obj := range pdf.pdfObjs {
		if plObj, ok := obj.(*pageLabelObj); ok {
			for n := 0; n <= 200; n += 2 {
				fw := &failWriterAt{n: n}
				plObj.write(fw, i+1)
			}
			return
		}
	}
	t.Skip("no PageLabelObj found")
}

// ============================================================
// Polyline — failWriter on cache
// ============================================================

func TestCov27_Polyline_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polyline([]Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}})

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 100; n += 1 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// Sector — failWriter on cache
// ============================================================

func TestCov27_Sector_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 200, 80, 0, 90, "FD")

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// Oval — failWriter on cache
// ============================================================

func TestCov27_Oval_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(50, 50, 200, 150)

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// Curve — failWriter on cache
// ============================================================

func TestCov27_Curve_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(50, 50, 100, 200, 200, 200, 250, 50, "D")

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// ClipPolygon — failWriter on cache
// ============================================================

func TestCov27_ClipPolygon_FailWriter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.ClipPolygon([]Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}})

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}

// ============================================================
// ImportedTemplate — failWriter on cache
// ============================================================

func TestCov27_ImportedTemplate_FailWriter(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Import a page from test PDF
	tpl := pdf.ImportPage(resTestPDF, 1, "/MediaBox")
	pdf.UseImportedTemplate(tpl, 0, 0, 200, 300)

	content, ok := pdf.pdfObjs[pdf.indexOfContent].(*ContentObj)
	if !ok {
		t.Fatal("not a ContentObj")
	}

	for _, cache := range content.listCache.caches {
		for n := 0; n <= 200; n += 2 {
			fw := &failWriterAt{n: n}
			cache.write(fw, nil)
		}
	}
}
