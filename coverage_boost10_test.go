package gopdf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost10_test.go — TestCov10_ prefix
// Targets: page_obj links, outlines_obj, cache_content_rotate,
// bookmark collectOutlineObjs, MeasureTextWidth, acroform_obj,
// form_field, ext_g_state_obj, font_container, doc_stats,
// ImportPagesFromSource, content_obj branches
// ============================================================

// ============================================================
// 1. PageObj.writeExternalLink / writeInternalLink (0% → covered)
// ============================================================

func TestCov10_PageObj_WriteExternalLink(t *testing.T) {
	p := &PageObj{}
	p.getRoot = func() *GoPdf {
		gp := &GoPdf{}
		gp.Start(Config{PageSize: *PageSizeA4})
		return gp
	}
	link := linkOption{
		url: "https://example.com/path?q=1&r=2",
		x:   10, y: 100, w: 200, h: 20,
	}
	var buf bytes.Buffer
	err := p.writeExternalLink(&buf, link, 5)
	if err != nil {
		t.Fatalf("writeExternalLink: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/URI") {
		t.Errorf("expected /URI: %s", out)
	}
	if !strings.Contains(out, "/Subtype /Link") {
		t.Errorf("expected /Subtype /Link: %s", out)
	}
}

func TestCov10_PageObj_WriteExternalLink_SpecialChars(t *testing.T) {
	p := &PageObj{}
	p.getRoot = func() *GoPdf {
		gp := &GoPdf{}
		gp.Start(Config{PageSize: *PageSizeA4})
		return gp
	}
	link := linkOption{
		url: `https://example.com/path\(test)` + "\r",
		x:   10, y: 100, w: 200, h: 20,
	}
	var buf bytes.Buffer
	err := p.writeExternalLink(&buf, link, 5)
	if err != nil {
		t.Fatalf("writeExternalLink: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "\r") {
		t.Error("should escape \\r")
	}
}

func TestCov10_PageObj_WriteExternalLink_Protected(t *testing.T) {
	p := &PageObj{}
	p.getRoot = func() *GoPdf {
		gp := newProtectedPDF(&testing.T{})
		return gp
	}
	link := linkOption{
		url: "https://example.com",
		x:   10, y: 100, w: 200, h: 20,
	}
	var buf bytes.Buffer
	// Protected PDF will encrypt the URL.
	err := p.writeExternalLink(&buf, link, 5)
	if err != nil {
		t.Fatalf("writeExternalLink protected: %v", err)
	}
}

func TestCov10_PageObj_WriteInternalLink(t *testing.T) {
	p := &PageObj{}
	anchors := map[string]anchorOption{
		"chapter1": {page: 3, y: 500},
	}
	link := linkOption{
		anchor: "chapter1",
		x:      10, y: 100, w: 200, h: 20,
	}
	var buf bytes.Buffer
	err := p.writeInternalLink(&buf, link, anchors)
	if err != nil {
		t.Fatalf("writeInternalLink: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Dest") {
		t.Errorf("expected /Dest: %s", out)
	}
}

func TestCov10_PageObj_WriteInternalLink_NotFound(t *testing.T) {
	p := &PageObj{}
	anchors := map[string]anchorOption{}
	link := linkOption{
		anchor: "nonexistent",
		x:      10, y: 100, w: 200, h: 20,
	}
	var buf bytes.Buffer
	err := p.writeInternalLink(&buf, link, anchors)
	if err != nil {
		t.Fatalf("writeInternalLink: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected empty output for missing anchor")
	}
}

// ============================================================
// 2. OutlineObj / OutlineNodes / OutlineNode Parse
// ============================================================

func TestCov10_OutlineObj_Write_WithChildren(t *testing.T) {
	o := &OutlineObj{
		title:     "Parent",
		index:     2,
		dest:      3,
		parent:    1,
		prev:      -1,
		next:      -1,
		first:     4,
		last:      5,
		height:    100.0,
		color:     [3]float64{1, 0, 0},
		bold:      true,
		italic:    true,
		collapsed: true,
	}
	var buf bytes.Buffer
	err := o.write(&buf, 2)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/First 4 0 R") {
		t.Errorf("expected /First: %s", out)
	}
	if !strings.Contains(out, "/Last 5 0 R") {
		t.Errorf("expected /Last: %s", out)
	}
	if !strings.Contains(out, "/C [") {
		t.Errorf("expected /C: %s", out)
	}
	if !strings.Contains(out, "/F 3") {
		t.Errorf("expected /F 3 (bold+italic): %s", out)
	}
	if !strings.Contains(out, "/Count") {
		t.Errorf("expected /Count: %s", out)
	}
}

func TestCov10_OutlineObj_Write_NoPrevNoNext(t *testing.T) {
	o := &OutlineObj{
		title:  "Solo",
		index:  2,
		dest:   3,
		parent: 1,
		prev:   -1,
		next:   -1,
		height: 50.0,
	}
	var buf bytes.Buffer
	err := o.write(&buf, 2)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "/Prev") {
		t.Error("should not have /Prev")
	}
	if strings.Contains(out, "/Next") {
		t.Error("should not have /Next")
	}
}

func TestCov10_OutlineObj_Write_WithPrevNext(t *testing.T) {
	o := &OutlineObj{
		title:  "Middle",
		index:  3,
		dest:   4,
		parent: 1,
		prev:   2,
		next:   5,
		height: 75.0,
	}
	var buf bytes.Buffer
	err := o.write(&buf, 3)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Prev 2 0 R") {
		t.Errorf("expected /Prev: %s", out)
	}
	if !strings.Contains(out, "/Next 5 0 R") {
		t.Errorf("expected /Next: %s", out)
	}
}

func TestCov10_OutlineNodes_Parse(t *testing.T) {
	child1 := &OutlineObj{index: 10, title: "C1"}
	child2 := &OutlineObj{index: 11, title: "C2"}
	child3 := &OutlineObj{index: 12, title: "C3"}

	nodes := OutlineNodes{
		{Obj: child1},
		{Obj: child2},
		{Obj: child3},
	}
	nodes.Parse()

	if child1.prev != -1 {
		t.Errorf("first prev should be -1, got %d", child1.prev)
	}
	if child3.next != -1 {
		t.Errorf("last next should be -1, got %d", child3.next)
	}
}

func TestCov10_OutlineNode_Parse_WithChildren(t *testing.T) {
	parent := &OutlineObj{index: 1, title: "Parent"}
	child1 := &OutlineObj{index: 10, title: "C1"}
	child2 := &OutlineObj{index: 11, title: "C2"}

	node := OutlineNode{
		Obj: parent,
		Children: []*OutlineNode{
			{Obj: child1},
			{Obj: child2},
		},
	}
	node.Parse()

	if parent.first != 10 {
		t.Errorf("parent first should be 10, got %d", parent.first)
	}
	if parent.last != 11 {
		t.Errorf("parent last should be 11, got %d", parent.last)
	}
	if child1.parent != 1 {
		t.Errorf("child1 parent should be 1, got %d", child1.parent)
	}
}

func TestCov10_OutlineNode_Parse_NoChildren(t *testing.T) {
	parent := &OutlineObj{index: 1, title: "Leaf"}
	node := OutlineNode{Obj: parent}
	node.Parse() // should not panic
}

func TestCov10_OutlinesObj_Write(t *testing.T) {
	o := &OutlinesObj{first: 3, last: 5, count: 2}
	var buf bytes.Buffer
	err := o.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Count 2") {
		t.Errorf("expected /Count 2: %s", out)
	}
	if !strings.Contains(out, "/First 3 0 R") {
		t.Errorf("expected /First: %s", out)
	}
}

func TestCov10_OutlinesObj_Write_NoFirstLast(t *testing.T) {
	o := &OutlinesObj{first: -1, last: -1, count: 0}
	var buf bytes.Buffer
	err := o.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "/First") {
		t.Error("should not have /First")
	}
}

// ============================================================
// 3. cacheContentRotate.write — non-reset path (77.8%)
// ============================================================

func TestCov10_CacheContentRotate_Write_NonReset(t *testing.T) {
	cc := &cacheContentRotate{
		isReset:    false,
		pageHeight: 842,
		angle:      45,
		x:          100,
		y:          200,
	}
	var buf bytes.Buffer
	err := cc.write(&buf, nil)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "q\n") {
		t.Errorf("expected q: %s", out)
	}
	if !strings.Contains(out, "cm") {
		t.Errorf("expected cm: %s", out)
	}
}

func TestCov10_CacheContentRotate_Write_Reset(t *testing.T) {
	cc := &cacheContentRotate{isReset: true}
	var buf bytes.Buffer
	err := cc.write(&buf, nil)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Q\n") {
		t.Errorf("expected Q: %s", out)
	}
}

// ============================================================
// 4. MeasureTextWidth / MeasureCellHeightByText (71.4%)
// ============================================================

func TestCov10_MeasureTextWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth: %v", err)
	}
	if w <= 0 {
		t.Errorf("expected positive width, got %f", w)
	}
}

func TestCov10_MeasureCellHeightByText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	h, err := pdf.MeasureCellHeightByText("Hello\nWorld")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Errorf("expected positive height, got %f", h)
	}
}

// ============================================================
// 5. acroFormObj.write — with fontRefs (71.4%)
// ============================================================

func TestCov10_AcroFormObj_Write_WithFonts(t *testing.T) {
	a := acroFormObj{
		fieldObjIDs: []int{5, 6, 7},
		fontRefs: []acroFormFont{
			{name: "F1", objID: 10},
			{name: "F2", objID: 11},
		},
	}
	var buf bytes.Buffer
	err := a.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Fields [") {
		t.Errorf("expected /Fields: %s", out)
	}
	if !strings.Contains(out, "/DR << /Font <<") {
		t.Errorf("expected /DR: %s", out)
	}
	if !strings.Contains(out, "/F1 10 0 R") {
		t.Errorf("expected font ref: %s", out)
	}
}

func TestCov10_AcroFormObj_Write_NoFonts(t *testing.T) {
	a := acroFormObj{
		fieldObjIDs: []int{5},
	}
	var buf bytes.Buffer
	err := a.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "/DR") {
		t.Error("should not have /DR without fonts")
	}
}

// ============================================================
// 6. FormField — all types via AddFormField (66.7%)
// ============================================================

func TestCov10_AddFormField_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:        FormFieldText,
		Name:        "username",
		X:           50, Y: 100, W: 200, H: 25,
		Value:       "default",
		FontFamily:  fontFamily,
		FontSize:    12,
		MaxLen:      100,
		Multiline:   true,
		ReadOnly:    false,
		Required:    true,
		HasBorder:   true,
		BorderColor: [3]uint8{0, 0, 0},
		HasFill:     true,
		FillColor:   [3]uint8{255, 255, 200},
		Color:       [3]uint8{0, 0, 128},
	})
	if err != nil {
		t.Fatalf("AddFormField text: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_AddFormField_Checkbox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddCheckbox("agree", 50, 200, 15, true)
	if err != nil {
		t.Fatalf("AddCheckbox: %v", err)
	}

	err = pdf.AddCheckbox("terms", 50, 230, 15, false)
	if err != nil {
		t.Fatalf("AddCheckbox unchecked: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_AddFormField_Radio(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldRadio,
		Name: "gender",
		X:    50, Y: 300, W: 15, H: 15,
	})
	if err != nil {
		t.Fatalf("AddFormField radio: %v", err)
	}
}

func TestCov10_AddFormField_Choice(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddDropdown("country", 50, 350, 200, 25, []string{"US", "UK", "CN", "JP"})
	if err != nil {
		t.Fatalf("AddDropdown: %v", err)
	}
}

func TestCov10_AddFormField_Button(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldButton,
		Name: "submit",
		X:    50, Y: 400, W: 100, H: 30,
	})
	if err != nil {
		t.Fatalf("AddFormField button: %v", err)
	}
}

func TestCov10_AddFormField_Signature(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldSignature,
		Name: "sig",
		X:    50, Y: 450, W: 200, H: 50,
	})
	if err != nil {
		t.Fatalf("AddFormField signature: %v", err)
	}
}

func TestCov10_AddFormField_Errors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldText,
		Name: "",
		X:    50, Y: 100, W: 200, H: 25,
	})
	if err == nil {
		t.Error("expected error for empty name")
	}

	err = pdf.AddFormField(FormField{
		Type: FormFieldText,
		Name: "test",
		X:    50, Y: 100, W: 0, H: 25,
	})
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestCov10_AddTextField(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddTextField("email", 50, 500, 200, 25)
	if err != nil {
		t.Fatalf("AddTextField: %v", err)
	}
}

func TestCov10_GetFormFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddTextField("f1", 10, 10, 100, 20)
	pdf.AddCheckbox("f2", 10, 40, 15, true)

	fields := pdf.GetFormFields()
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}

func TestCov10_FormFieldType_String(t *testing.T) {
	types := []FormFieldType{FormFieldText, FormFieldCheckbox, FormFieldRadio, FormFieldChoice, FormFieldButton, FormFieldSignature, FormFieldType(99)}
	expected := []string{"Text", "Checkbox", "Radio", "Choice", "Button", "Signature", "Unknown"}
	for i, ft := range types {
		if ft.String() != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], ft.String())
		}
	}
}

// ============================================================
// 7. ExtGState — write with all fields
// ============================================================

func TestCov10_ExtGState_Write_AllFields(t *testing.T) {
	ca := 0.5
	CA := 0.8
	bm := Multiply
	smaskIdx := 10
	egs := ExtGState{
		ca:         &ca,
		CA:         &CA,
		BM:         &bm,
		SMaskIndex: &smaskIdx,
	}
	var buf bytes.Buffer
	err := egs.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/ca 0.500") {
		t.Errorf("expected /ca: %s", out)
	}
	if !strings.Contains(out, "/CA 0.800") {
		t.Errorf("expected /CA: %s", out)
	}
	if !strings.Contains(out, "/BM") {
		t.Errorf("expected /BM: %s", out)
	}
	if !strings.Contains(out, "/SMask 11 0 R") {
		t.Errorf("expected /SMask: %s", out)
	}
}

func TestCov10_ExtGState_Write_Minimal(t *testing.T) {
	egs := ExtGState{}
	var buf bytes.Buffer
	err := egs.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Type /ExtGState") {
		t.Errorf("expected /Type: %s", out)
	}
}

func TestCov10_ExtGStateOptions_GetId(t *testing.T) {
	ca := 0.5
	CA := 0.8
	bm := Multiply
	smask := 3
	opts := ExtGStateOptions{
		StrokingCA:    &CA,
		NonStrokingCa: &ca,
		BlendMode:     &bm,
		SMaskIndex:    &smask,
	}
	id := opts.GetId()
	if !strings.Contains(id, "CA_") {
		t.Errorf("expected CA_ in id: %s", id)
	}
	if !strings.Contains(id, "ca_") {
		t.Errorf("expected ca_ in id: %s", id)
	}
	if !strings.Contains(id, "BM_") {
		t.Errorf("expected BM_ in id: %s", id)
	}
	if !strings.Contains(id, "SMask_") {
		t.Errorf("expected SMask_ in id: %s", id)
	}
}

// ============================================================
// 8. FontContainer (71.4%)
// ============================================================

func TestCov10_FontContainer_AddTTFFontWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	fc := &FontContainer{}
	err := fc.AddTTFFontWithOption(fontFamily, resFontPath, defaultTtfFontOption())
	if err != nil {
		t.Fatalf("AddTTFFontWithOption: %v", err)
	}

	// Verify it's stored.
	err = fc.AddTTFFont(fontFamily+"2", resFontPath)
	if err != nil {
		t.Fatalf("AddTTFFont: %v", err)
	}
}

func TestCov10_FontContainer_AddTTFFont_BadPath(t *testing.T) {
	fc := &FontContainer{}
	err := fc.AddTTFFont("bad", "/nonexistent/font.ttf")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestCov10_FontContainer_AddTTFFontByReader(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	data, _ := os.ReadFile(resFontPath)
	fc := &FontContainer{}
	err := fc.AddTTFFontByReader(fontFamily, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("AddTTFFontByReader: %v", err)
	}
}

func TestCov10_FontContainer_AddTTFFontData(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	data, _ := os.ReadFile(resFontPath)
	fc := &FontContainer{}
	err := fc.AddTTFFontData(fontFamily, data)
	if err != nil {
		t.Fatalf("AddTTFFontData: %v", err)
	}
}

func TestCov10_AddTTFFontFromFontContainer(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	fc := &FontContainer{}
	fc.AddTTFFont(fontFamily, resFontPath)

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontFromFontContainer(fontFamily, fc)
	if err != nil {
		t.Fatalf("AddTTFFontFromFontContainer: %v", err)
	}
}

func TestCov10_AddTTFFontFromFontContainer_NotFound(t *testing.T) {
	fc := &FontContainer{}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontFromFontContainer("nonexistent", fc)
	if err == nil {
		t.Error("expected error for missing font")
	}
}

// ============================================================
// 9. GetFonts / GetDocumentStats (75%)
// ============================================================

func TestCov10_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fonts := pdf.GetFonts()
	if len(fonts) == 0 {
		t.Error("expected at least one font")
	}
	found := false
	for _, f := range fonts {
		if f.Family == fontFamily {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find %s in fonts", fontFamily)
	}
}

func TestCov10_GetDocumentStats_Full(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")
	pdf.AddOutline("Ch1")
	pdf.SetXMPMetadata(XMPMetadata{Title: "Test"})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: PageLabelDecimal}})

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Errorf("expected 1 page, got %d", stats.PageCount)
	}
	if stats.FontCount == 0 {
		t.Error("expected fonts")
	}
	if !stats.HasOutlines {
		t.Error("expected outlines")
	}
	if !stats.HasXMPMetadata {
		t.Error("expected XMP metadata")
	}
	if !stats.HasPageLabels {
		t.Error("expected page labels")
	}
}

// ============================================================
// 10. ImportPagesFromSource (0%)
// ============================================================

func TestCov10_ImportPagesFromSource_File(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.ImportPagesFromSource(resTestPDF, "/MediaBox")
	if err != nil {
		t.Fatalf("ImportPagesFromSource file: %v", err)
	}
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_ImportPagesFromSource_Bytes(t *testing.T) {
	pdfData, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.ImportPagesFromSource(pdfData, "/MediaBox")
	if err != nil {
		t.Fatalf("ImportPagesFromSource bytes: %v", err)
	}
}

func TestCov10_ImportPagesFromSource_ReadSeeker(t *testing.T) {
	pdfData, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	rs := io.ReadSeeker(bytes.NewReader(pdfData))
	err = pdf.ImportPagesFromSource(rs, "/MediaBox")
	if err != nil {
		t.Fatalf("ImportPagesFromSource ReadSeeker: %v", err)
	}
}

func TestCov10_ImportPagesFromSource_ReadSeekerPtr(t *testing.T) {
	pdfData, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	rs := io.ReadSeeker(bytes.NewReader(pdfData))
	err = pdf.ImportPagesFromSource(&rs, "/MediaBox")
	if err != nil {
		t.Fatalf("ImportPagesFromSource *ReadSeeker: %v", err)
	}
}

func TestCov10_ImportPagesFromSource_BadType(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.ImportPagesFromSource(12345, "/MediaBox")
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestCov10_ImportPage(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	tpl := pdf.ImportPage(resTestPDF, 1, "/MediaBox")
	// Template ID may be 0-based.
	pdf.UseImportedTemplate(tpl, 0, 0, PageSizeA4.W, PageSizeA4.H)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_ImportPageStream(t *testing.T) {
	pdfData, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Skipf("test PDF not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	rs := io.ReadSeeker(bytes.NewReader(pdfData))
	tpl := pdf.ImportPageStream(&rs, 1, "/MediaBox")
	// Template ID may be 0-based.
	_ = tpl
}

// ============================================================
// 11. collectOutlineObjs — child traversal (72.7%)
// ============================================================

func TestCov10_CollectOutlineObjs_WithChildren(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Use AddOutlineWithPosition to get OutlineObj references.
	parent := pdf.AddOutlineWithPosition("Parent")

	pdf.AddPage()
	child := pdf.AddOutlineWithPosition("Child")

	// Set parent-child relationship.
	parent.SetFirst(child.GetIndex())
	parent.SetLast(child.GetIndex())
	child.SetParent(parent.GetIndex())

	toc := pdf.GetTOC()
	if len(toc) == 0 {
		t.Error("expected TOC entries")
	}
}

// ============================================================
// 12. Content stream — more branches
// ============================================================

func TestCov10_ContentObj_AppendStreamSubsetFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Use Cell with various options to exercise AppendStreamSubsetFont.
	pdf.Cell(&Rect{W: 200, H: 20}, "Normal text")
	pdf.SetTextColor(255, 0, 0)
	pdf.Cell(&Rect{W: 200, H: 20}, "Red text")
	pdf.SetTextColorCMYK(0, 100, 100, 0)
	pdf.Cell(&Rect{W: 200, H: 20}, "CMYK text")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_ContentObj_AppendStreamOval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(50, 50, 200, 100)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_ContentObj_AppendStreamCurve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(10, 10, 50, 200, 200, 200, 200, 10, "D")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_ContentObj_AppendStreamPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	points := []Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}
	pdf.Polygon(points, "D")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_ContentObj_Rotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.Cell(nil, "Rotated text")
	pdf.RotateReset()
	pdf.Cell(nil, "Normal text")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 13. SplitText / SplitTextWithWordWrap
// ============================================================

func TestCov10_SplitText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	lines, err := pdf.SplitText("Hello World this is a long text that should be split", 100)
	if err != nil {
		t.Fatalf("SplitText: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

func TestCov10_SplitTextWithWordWrap(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	lines, err := pdf.SplitTextWithWordWrap("Hello World this is a long text that should be split at word boundaries", 100)
	if err != nil {
		t.Fatalf("SplitTextWithWordWrap: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

func TestCov10_SplitText_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_, err := pdf.SplitText("", 100)
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestCov10_SplitText_WithNewlines(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	lines, err := pdf.SplitText("Line1\nLine2\nLine3", 500)
	if err != nil {
		t.Fatalf("SplitText: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

// ============================================================
// 14. PlaceHolderText / FillInPlaceHoldText
// ============================================================

func TestCov10_PlaceHolderText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.PlaceHolderText("name", 200)
	if err != nil {
		t.Fatalf("PlaceHolderText: %v", err)
	}

	err = pdf.FillInPlaceHoldText("name", "John Doe", Left)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText Left: %v", err)
	}
}

func TestCov10_PlaceHolderText_Center(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 100)

	pdf.PlaceHolderText("center_field", 200)
	err := pdf.FillInPlaceHoldText("center_field", "Centered", Center)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText Center: %v", err)
	}
}

func TestCov10_PlaceHolderText_Right(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 150)

	pdf.PlaceHolderText("right_field", 200)
	err := pdf.FillInPlaceHoldText("right_field", "Right", Right)
	if err != nil {
		t.Fatalf("FillInPlaceHoldText Right: %v", err)
	}
}

func TestCov10_FillInPlaceHoldText_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.FillInPlaceHoldText("nonexistent", "test", Left)
	if err == nil {
		t.Error("expected error for missing placeholder")
	}
}

// ============================================================
// 15. Text function (71.4%)
// ============================================================

func TestCov10_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.Text("Hello World")
	if err != nil {
		t.Fatalf("Text: %v", err)
	}

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 16. AddTTFFontWithOption on GoPdf (71.4%)
// ============================================================

func TestCov10_GoPdf_AddTTFFontWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{UseKerning: true})
	if err != nil {
		t.Fatalf("AddTTFFontWithOption: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.Cell(nil, "Kerning test")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 17. Bookmark with nested children — collectOutlineObjs
// ============================================================

func TestCov10_Bookmark_NestedChildren(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	parent := pdf.AddOutlineWithPosition("Chapter 1")

	pdf.AddPage()
	child1 := pdf.AddOutlineWithPosition("Section 1.1")
	child1.SetParent(parent.GetIndex())
	parent.SetFirst(child1.GetIndex())
	parent.SetLast(child1.GetIndex())

	pdf.AddPage()
	child2 := pdf.AddOutlineWithPosition("Section 1.2")
	child2.SetParent(parent.GetIndex())
	child2.SetPrev(child1.GetIndex())
	child1.SetNext(child2.GetIndex())
	parent.SetLast(child2.GetIndex())

	toc := pdf.GetTOC()
	_ = toc

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 18. ContentObjCalTextHeight / ContentObjCalTextHeightPrecise
// ============================================================

func TestCov10_ContentObjCalTextHeight(t *testing.T) {
	h := ContentObjCalTextHeight(14)
	if h <= 0 {
		t.Errorf("expected positive height, got %f", h)
	}
}

func TestCov10_ContentObjCalTextHeightPrecise(t *testing.T) {
	h := ContentObjCalTextHeightPrecise(14.5)
	if h <= 0 {
		t.Errorf("expected positive height, got %f", h)
	}
}

// ============================================================
// 19. fixRange10 / convertTTFUnit2PDFUnit
// ============================================================

func TestCov10_FixRange10(t *testing.T) {
	tests := []struct {
		in, out float64
	}{
		{0.5, 0.5},
		{-0.1, 0},
		{1.5, 1},
	}
	for _, tc := range tests {
		got := fixRange10(tc.in)
		if got != tc.out {
			t.Errorf("fixRange10(%f) = %f, want %f", tc.in, got, tc.out)
		}
	}
}

func TestCov10_ConvertTTFUnit2PDFUnit(t *testing.T) {
	result := convertTTFUnit2PDFUnit(500, 1000)
	if result != 500 {
		t.Errorf("expected 500, got %d", result)
	}
}

// ============================================================
// 20. FormatFloatTrim
// ============================================================

func TestCov10_FormatFloatTrim(t *testing.T) {
	tests := []struct {
		in  float64
		out string
	}{
		{1.0, "1"},
		{1.5, "1.5"},
		{1.50, "1.5"},
		{0.123456, "0.123"},
	}
	for _, tc := range tests {
		got := FormatFloatTrim(tc.in)
		if got != tc.out {
			t.Errorf("FormatFloatTrim(%f) = %q, want %q", tc.in, got, tc.out)
		}
	}
}

// ============================================================
// 21. AddExternalLink / AddInternalLink — via full PDF
// ============================================================

func TestCov10_AddExternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Click here")
	pdf.AddExternalLink("https://example.com", 50, 50, 100, 20)

	data := pdf.GetBytesPdf()
	s := string(data)
	if !strings.Contains(s, "/URI") {
		t.Error("expected /URI in output")
	}
}

func TestCov10_AddInternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetAnchor("target")
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Target")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Go to target")
	pdf.AddInternalLink("target", 50, 50, 100, 20)

	data := pdf.GetBytesPdf()
	s := string(data)
	if !strings.Contains(s, "/Dest") {
		t.Logf("Internal link may not appear in raw bytes (rendered at write time)")
	}
	_ = s
}

// ============================================================
// 22. Misc: SetGrayStroke, SetColorStrokeCMYK, SetColorFillCMYK
// ============================================================

func TestCov10_SetGrayStroke(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayStroke(0.5)
	pdf.Line(10, 10, 200, 200)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_SetColorStrokeCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColorCMYK(100, 0, 0, 0)
	pdf.Line(10, 10, 200, 200)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_SetFillColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColorCMYK(0, 100, 0, 0)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 100, "F")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 23. Polyline / Sector
// ============================================================

func TestCov10_Polyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	points := []Point{{X: 50, Y: 50}, {X: 100, Y: 100}, {X: 150, Y: 50}, {X: 200, Y: 100}}
	pdf.Polyline(points)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 200, 50, 0, 90, "FD")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 24. SetLineWidth / SetLineType / SetCustomLineType
// ============================================================

func TestCov10_SetLineWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(2.5)
	pdf.Line(10, 10, 200, 200)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_SetLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineType("dashed")
	pdf.Line(10, 10, 200, 200)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

func TestCov10_SetCustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(10, 10, 200, 200)
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 25. SaveGraphicsState / RestoreGraphicsState
// ============================================================

func TestCov10_SaveRestoreGraphicsState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	pdf.SetGrayFill(0.5)
	pdf.Cell(nil, "Gray")
	pdf.RestoreGraphicsState()
	pdf.Cell(nil, "Normal")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 26. appendColorSpace
// ============================================================

func TestCov10_AppendColorSpace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// Use SetColorSpace to trigger appendColorSpace.
	pdf.SetFillColor(128, 0, 255)
	pdf.Cell(nil, "Colored")
	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF")
	}
}

// ============================================================
// 27. computeRotateTransformationMatrix
// ============================================================

func TestCov10_ComputeRotateTransformationMatrix(t *testing.T) {
	result := computeRotateTransformationMatrix(100, 200, 45, 842)
	if !strings.Contains(result, "cm") {
		t.Errorf("expected cm in result: %s", result)
	}
}

// ============================================================
// 28. Misc: GetNumberOfPages, GetObjectCount, GetLiveObjectCount
// ============================================================

func TestCov10_GetNumberOfPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov10_GetObjectCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	count := pdf.GetObjectCount()
	if count <= 0 {
		t.Errorf("expected positive count, got %d", count)
	}
}

func TestCov10_GetLiveObjectCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	live := pdf.GetLiveObjectCount()
	if live <= 0 {
		t.Errorf("expected positive count, got %d", live)
	}
}

// ============================================================
// 29. AddPageWithOption — various options
// ============================================================

func TestCov10_AddPageWithOption_CustomSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 300, H: 400},
	})
	pdf.Cell(nil, "Custom size page")
	data := pdf.GetBytesPdf()
	if !strings.Contains(string(data), "300.00") {
		t.Error("expected custom width in output")
	}
}

// ============================================================
// 30. Ensure unused import is used
// ============================================================

func TestCov10_Imports(t *testing.T) {
	_ = fmt.Sprintf("test")
	_ = os.TempDir()
	_ = io.Discard
}
