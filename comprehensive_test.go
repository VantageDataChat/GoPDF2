package gopdf

import (
	"bytes"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

// ============================================================
// Geometry: RectFrom - comprehensive edge cases
// ============================================================

func TestComp_RectFrom_Contains(t *testing.T) {
	r := RectFrom{X: 10, Y: 20, W: 100, H: 50}
	tests := []struct {
		px, py float64
		want   bool
	}{
		{10, 20, true},   // top-left corner
		{110, 70, true},  // bottom-right corner
		{60, 45, true},   // center
		{9, 20, false},   // just outside left
		{111, 70, false}, // just outside right
		{10, 19, false},  // just above
		{10, 71, false},  // just below
	}
	for _, tt := range tests {
		got := r.Contains(tt.px, tt.py)
		if got != tt.want {
			t.Errorf("Contains(%.0f, %.0f) = %v, want %v", tt.px, tt.py, got, tt.want)
		}
	}
}

func TestComp_RectFrom_ContainsRect(t *testing.T) {
	outer := RectFrom{X: 0, Y: 0, W: 100, H: 100}
	inner := RectFrom{X: 10, Y: 10, W: 50, H: 50}
	bigger := RectFrom{X: -1, Y: 0, W: 102, H: 100}

	if !outer.ContainsRect(inner) {
		t.Error("outer should contain inner")
	}
	if outer.ContainsRect(bigger) {
		t.Error("outer should not contain bigger")
	}
	if !outer.ContainsRect(outer) {
		t.Error("rect should contain itself")
	}
}

func TestComp_RectFrom_Intersects(t *testing.T) {
	a := RectFrom{X: 0, Y: 0, W: 50, H: 50}
	b := RectFrom{X: 25, Y: 25, W: 50, H: 50}
	c := RectFrom{X: 100, Y: 100, W: 10, H: 10}

	if !a.Intersects(b) {
		t.Error("a and b should intersect")
	}
	if a.Intersects(c) {
		t.Error("a and c should not intersect")
	}
}

func TestComp_RectFrom_Intersection(t *testing.T) {
	a := RectFrom{X: 0, Y: 0, W: 50, H: 50}
	b := RectFrom{X: 25, Y: 25, W: 50, H: 50}
	inter := a.Intersection(b)
	if inter.X != 25 || inter.Y != 25 || inter.W != 25 || inter.H != 25 {
		t.Errorf("Intersection = %+v, want {25 25 25 25}", inter)
	}
	c := RectFrom{X: 100, Y: 100, W: 10, H: 10}
	noInter := a.Intersection(c)
	if !noInter.IsEmpty() {
		t.Errorf("non-overlapping intersection should be empty, got %+v", noInter)
	}
}

func TestComp_RectFrom_Union(t *testing.T) {
	a := RectFrom{X: 10, Y: 20, W: 30, H: 40}
	b := RectFrom{X: 50, Y: 60, W: 10, H: 10}
	u := a.Union(b)
	if u.X != 10 || u.Y != 20 || u.W != 50 || u.H != 50 {
		t.Errorf("Union = %+v, want {10 20 50 50}", u)
	}
}

func TestComp_RectFrom_IsEmpty(t *testing.T) {
	if !(RectFrom{W: 0, H: 10}).IsEmpty() {
		t.Error("zero width should be empty")
	}
	if !(RectFrom{W: 10, H: -1}).IsEmpty() {
		t.Error("negative height should be empty")
	}
	if (RectFrom{W: 1, H: 1}).IsEmpty() {
		t.Error("1x1 should not be empty")
	}
}

func TestComp_RectFrom_Area(t *testing.T) {
	r := RectFrom{W: 10, H: 20}
	if r.Area() != 200 {
		t.Errorf("Area() = %f, want 200", r.Area())
	}
	empty := RectFrom{W: -1, H: 10}
	if empty.Area() != 0 {
		t.Errorf("negative width Area() = %f, want 0", empty.Area())
	}
}

func TestComp_RectFrom_Center(t *testing.T) {
	r := RectFrom{X: 10, Y: 20, W: 100, H: 50}
	c := r.Center()
	if c.X != 60 || c.Y != 45 {
		t.Errorf("Center() = %+v, want {60 45}", c)
	}
}

func TestComp_RectFrom_Normalize(t *testing.T) {
	r := RectFrom{X: 50, Y: 50, W: -30, H: -20}
	n := r.Normalize()
	if n.X != 20 || n.Y != 30 || n.W != 30 || n.H != 20 {
		t.Errorf("Normalize() = %+v, want {20 30 30 20}", n)
	}
}

// ============================================================
// Geometry: Matrix - comprehensive
// ============================================================

func TestComp_IdentityMatrix(t *testing.T) {
	m := IdentityMatrix()
	if !m.IsIdentity() {
		t.Error("IdentityMatrix should be identity")
	}
}

func TestComp_TranslateMatrix(t *testing.T) {
	m := TranslateMatrix(10, 20)
	x, y := m.TransformPoint(0, 0)
	if x != 10 || y != 20 {
		t.Errorf("TranslateMatrix(10,20).TransformPoint(0,0) = (%f,%f), want (10,20)", x, y)
	}
}

func TestComp_ScaleMatrix(t *testing.T) {
	m := ScaleMatrix(2, 3)
	x, y := m.TransformPoint(5, 10)
	if x != 10 || y != 30 {
		t.Errorf("ScaleMatrix(2,3).TransformPoint(5,10) = (%f,%f), want (10,30)", x, y)
	}
}

func TestComp_RotateMatrix(t *testing.T) {
	m := RotateMatrix(90)
	x, y := m.TransformPoint(1, 0)
	if math.Abs(x) > 1e-10 || math.Abs(y-1) > 1e-10 {
		t.Errorf("RotateMatrix(90).TransformPoint(1,0) = (%f,%f), want (0,1)", x, y)
	}
}

func TestComp_MatrixMultiply(t *testing.T) {
	translate := TranslateMatrix(10, 0)
	scale := ScaleMatrix(2, 2)
	combined := scale.Multiply(translate)
	x, y := combined.TransformPoint(5, 5)
	if math.Abs(x-30) > 1e-10 || math.Abs(y-10) > 1e-10 {
		t.Errorf("combined.TransformPoint(5,5) = (%f,%f), want (30,10)", x, y)
	}
}

func TestComp_MatrixIsIdentity(t *testing.T) {
	if !IdentityMatrix().IsIdentity() {
		t.Error("identity should be identity")
	}
	if TranslateMatrix(1, 0).IsIdentity() {
		t.Error("translate should not be identity")
	}
}

func TestComp_Distance(t *testing.T) {
	d := Distance(Point{X: 0, Y: 0}, Point{X: 3, Y: 4})
	if math.Abs(d-5) > 1e-10 {
		t.Errorf("Distance = %f, want 5", d)
	}
}

// ============================================================
// Config & Units
// ============================================================

func TestComp_Config_getUnit(t *testing.T) {
	c := Config{Unit: UnitMM}
	if c.getUnit() != UnitMM {
		t.Errorf("getUnit() = %d, want %d", c.getUnit(), UnitMM)
	}
}

func TestComp_UnitsToPoints(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{Unit: UnitMM, PageSize: *PageSizeA4})
	pts := pdf.UnitsToPoints(25.4)
	if math.Abs(pts-72) > 0.01 {
		t.Errorf("UnitsToPoints(25.4mm) = %f, want 72", pts)
	}
}

func TestComp_PointsToUnits(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{Unit: UnitIN, PageSize: *PageSizeA4})
	u := pdf.PointsToUnits(72)
	if math.Abs(u-1) > 0.01 {
		t.Errorf("PointsToUnits(72pt) = %f, want 1 inch", u)
	}
}

func TestComp_UnitsToPointsVar(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{Unit: UnitCM, PageSize: *PageSizeA4})
	v := 2.54
	pdf.UnitsToPointsVar(&v)
	if math.Abs(v-72) > 0.01 {
		t.Errorf("UnitsToPointsVar(2.54cm) = %f, want 72", v)
	}
}

// ============================================================
// Watermark edge cases
// ============================================================

func TestComp_WatermarkText_EmptyText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{Text: "", FontFamily: fontFamily})
	if err != ErrEmptyString {
		t.Errorf("expected ErrEmptyString, got %v", err)
	}
}

func TestComp_WatermarkText_MissingFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{Text: "test", FontFamily: ""})
	if err != ErrMissingFontFamily {
		t.Errorf("expected ErrMissingFontFamily, got %v", err)
	}
}

func TestComp_WatermarkText_Repeat(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text: "DRAFT", FontFamily: fontFamily, FontSize: 36,
		Repeat: true, Opacity: 0.2,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText repeat: %v", err)
	}
}

func TestComp_WatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")
	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text: "CONFIDENTIAL", FontFamily: fontFamily, FontSize: 48, Opacity: 0.3,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

func TestComp_WatermarkImage_DefaultSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkImage(resJPEGPath, 0.5, 0, 0, 0)
	if err != nil {
		t.Skipf("image not available: %v", err)
	}
}

// ============================================================
// Embedded Files CRUD
// ============================================================

func TestComp_EmbeddedFile_AddAndRetrieve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	content := []byte("Hello, embedded world!")
	err := pdf.AddEmbeddedFile(EmbeddedFile{Name: "test.txt", Content: content, MimeType: "text/plain"})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}
	got, err := pdf.GetEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFile: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestComp_EmbeddedFile_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.GetEmbeddedFile("nonexistent.txt")
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected ErrEmbeddedFileNotFound, got %v", err)
	}
}

func TestComp_EmbeddedFile_Delete(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddEmbeddedFile(EmbeddedFile{Name: "del.txt", Content: []byte("data")})
	if err := pdf.DeleteEmbeddedFile("del.txt"); err != nil {
		t.Fatalf("DeleteEmbeddedFile: %v", err)
	}
	_, err := pdf.GetEmbeddedFile("del.txt")
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected not found after delete, got %v", err)
	}
}

func TestComp_EmbeddedFile_DeleteNotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeleteEmbeddedFile("nope.txt")
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected ErrEmbeddedFileNotFound, got %v", err)
	}
}

func TestComp_EmbeddedFile_Update(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddEmbeddedFile(EmbeddedFile{Name: "upd.txt", Content: []byte("original")})
	newContent := []byte("updated content")
	err := pdf.UpdateEmbeddedFile("upd.txt", EmbeddedFile{Name: "upd.txt", Content: newContent})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}
	got, _ := pdf.GetEmbeddedFile("upd.txt")
	if !bytes.Equal(got, newContent) {
		t.Errorf("content after update: got %q, want %q", got, newContent)
	}
}

func TestComp_EmbeddedFile_UpdateNotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.UpdateEmbeddedFile("nope.txt", EmbeddedFile{Content: []byte("x")})
	if err != ErrEmbeddedFileNotFound {
		t.Errorf("expected ErrEmbeddedFileNotFound, got %v", err)
	}
}

func TestComp_EmbeddedFile_Info(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	content := []byte("info test data")
	pdf.AddEmbeddedFile(EmbeddedFile{
		Name: "info.txt", Content: content, MimeType: "text/plain", Description: "A test file",
	})
	info, err := pdf.GetEmbeddedFileInfo("info.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFileInfo: %v", err)
	}
	if info.Size != len(content) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}
	if info.MimeType != "text/plain" {
		t.Errorf("MimeType = %q, want %q", info.MimeType, "text/plain")
	}
}

func TestComp_EmbeddedFile_Names(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddEmbeddedFile(EmbeddedFile{Name: "a.txt", Content: []byte("a")})
	pdf.AddEmbeddedFile(EmbeddedFile{Name: "b.txt", Content: []byte("b")})
	names := pdf.GetEmbeddedFileNames()
	if len(names) != 2 || names[0] != "a.txt" || names[1] != "b.txt" {
		t.Errorf("names = %v, want [a.txt b.txt]", names)
	}
	if pdf.GetEmbeddedFileCount() != 2 {
		t.Errorf("count = %d, want 2", pdf.GetEmbeddedFileCount())
	}
}

func TestComp_EmbeddedFile_EmptyName(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddEmbeddedFile(EmbeddedFile{Name: "", Content: []byte("x")})
	if err != ErrEmptyString {
		t.Errorf("expected ErrEmptyString, got %v", err)
	}
}

func TestComp_EmbeddedFile_EmptyContent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddEmbeddedFile(EmbeddedFile{Name: "empty.txt", Content: nil})
	if err != ErrEmptyString {
		t.Errorf("expected ErrEmptyString, got %v", err)
	}
}

// ============================================================
// Clone & Scrub
// ============================================================

func TestComp_Clone_Independent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Original document")
	clone, err := pdf.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if clone.GetNumberOfPages() != pdf.GetNumberOfPages() {
		t.Errorf("clone pages = %d, original = %d", clone.GetNumberOfPages(), pdf.GetNumberOfPages())
	}
	data, err := clone.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("clone GetBytesPdfReturnErr: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("clone output doesn't start with %PDF-")
	}
}

func TestComp_Clone_PreservesMetadata(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Content for clone")
	pdf.SetPDFVersion(PDFVersion17)
	pdf.SetXMPMetadata(XMPMetadata{Title: "Test"})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: "D"}})
	clone, err := pdf.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	if clone.GetPDFVersion() != PDFVersion17 {
		t.Errorf("clone version = %v, want %v", clone.GetPDFVersion(), PDFVersion17)
	}
	if clone.GetXMPMetadata() == nil || clone.GetXMPMetadata().Title != "Test" {
		t.Error("clone XMP metadata mismatch")
	}
}

func TestComp_Scrub_AllOptions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetInfo(PdfInfo{Title: "Secret", Author: "Agent"})
	pdf.SetXMPMetadata(XMPMetadata{Title: "Secret XMP"})
	pdf.AddEmbeddedFile(EmbeddedFile{Name: "secret.txt", Content: []byte("data")})
	pdf.SetPageLabels([]PageLabel{{PageIndex: 0, Style: "D"}})
	pdf.Scrub(DefaultScrubOption())
	if pdf.GetXMPMetadata() != nil {
		t.Error("XMP should be nil after scrub")
	}
	if pdf.GetEmbeddedFileCount() != 0 {
		t.Error("embedded files should be empty after scrub")
	}
}

// ============================================================
// Garbage Collection
// ============================================================

func TestComp_GC_None(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.GarbageCollect(GCNone) != 0 {
		t.Error("GCNone should remove 0")
	}
}

func TestComp_GC_Compact(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	pdf.Cell(nil, "Page 3")
	before := pdf.GetObjectCount()
	pdf.DeletePage(2)
	removed := pdf.GarbageCollect(GCCompact)
	if removed == 0 {
		t.Error("GCCompact should remove objects after DeletePage")
	}
	if pdf.GetObjectCount() >= before {
		t.Errorf("object count should decrease: before=%d, after=%d", before, pdf.GetObjectCount())
	}
}

func TestComp_GC_LiveObjectCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	total := pdf.GetObjectCount()
	live := pdf.GetLiveObjectCount()
	if live != total {
		t.Errorf("live=%d should equal total=%d for fresh doc", live, total)
	}
}

func TestComp_GC_NoNulls(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.GarbageCollect(GCCompact) != 0 {
		t.Error("no nulls to remove")
	}
}

// ============================================================
// PDF Version
// ============================================================

func TestComp_PDFVersion_SetAndGet(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	versions := []PDFVersion{PDFVersion14, PDFVersion15, PDFVersion16, PDFVersion17, PDFVersion20}
	for _, v := range versions {
		pdf.SetPDFVersion(v)
		if pdf.GetPDFVersion() != v {
			t.Errorf("SetPDFVersion(%v) -> GetPDFVersion() = %v", v, pdf.GetPDFVersion())
		}
	}
}

// ============================================================
// Page Manipulation edge cases
// ============================================================

func TestComp_DeletePage_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.DeletePage(0) == nil {
		t.Error("DeletePage(0) should return error")
	}
	if pdf.DeletePage(5) == nil {
		t.Error("DeletePage(5) on 1-page doc should return error")
	}
}

func TestComp_DeletePage_LastPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")
	if err := pdf.DeletePage(2); err != nil {
		t.Fatalf("DeletePage(2): %v", err)
	}
	if pdf.GetNumberOfPages() != 1 {
		t.Errorf("pages after delete = %d, want 1", pdf.GetNumberOfPages())
	}
}

func TestComp_CopyPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Original")
	before := pdf.GetNumberOfPages()
	_, err := pdf.CopyPage(1)
	if err != nil {
		t.Fatalf("CopyPage: %v", err)
	}
	if pdf.GetNumberOfPages() != before+1 {
		t.Errorf("pages after copy = %d, want %d", pdf.GetNumberOfPages(), before+1)
	}
}

func TestComp_CopyPage_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if _, err := pdf.CopyPage(0); err == nil {
		t.Error("CopyPage(0) should return error")
	}
	if _, err := pdf.CopyPage(5); err == nil {
		t.Error("CopyPage(5) on 1-page doc should return error")
	}
}

// ============================================================
// Content Element CRUD
// ============================================================

func TestComp_GetPageElements_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if _, err := pdf.GetPageElements(0); err == nil {
		t.Error("GetPageElements(0) should return error")
	}
	if _, err := pdf.GetPageElements(99); err == nil {
		t.Error("GetPageElements(99) should return error")
	}
}

func TestComp_GetPageElementCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Hello")
	count, err := pdf.GetPageElementCount(1)
	if err != nil {
		t.Fatalf("GetPageElementCount: %v", err)
	}
	if count == 0 {
		t.Error("expected at least 1 element after Cell")
	}
}

func TestComp_DeleteElement_InvalidIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "text")
	if pdf.DeleteElement(1, -1) == nil {
		t.Error("DeleteElement with negative index should error")
	}
	if pdf.DeleteElement(1, 9999) == nil {
		t.Error("DeleteElement with out-of-range index should error")
	}
}

func TestComp_ClearPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "text")
	pdf.Line(0, 0, 100, 100)
	if err := pdf.ClearPage(1); err != nil {
		t.Fatalf("ClearPage: %v", err)
	}
	count, _ := pdf.GetPageElementCount(1)
	if count != 0 {
		t.Errorf("expected 0 elements after ClearPage, got %d", count)
	}
}

func TestComp_ClearPage_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.ClearPage(0) == nil {
		t.Error("ClearPage(0) should return error")
	}
}

func TestComp_InsertLineElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "content") // ensure content obj exists
	if err := pdf.InsertLineElement(1, 10, 10, 100, 100); err != nil {
		t.Fatalf("InsertLineElement: %v", err)
	}
	elems, _ := pdf.GetPageElements(1)
	found := false
	for _, e := range elems {
		if e.Type == ElementLine {
			found = true
		}
	}
	if !found {
		t.Error("expected to find a line element")
	}
}

func TestComp_InsertRectElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "content")
	if err := pdf.InsertRectElement(1, 10, 10, 50, 30, "D"); err != nil {
		t.Fatalf("InsertRectElement: %v", err)
	}
}

func TestComp_InsertOvalElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "content")
	if err := pdf.InsertOvalElement(1, 10, 10, 60, 40); err != nil {
		t.Fatalf("InsertOvalElement: %v", err)
	}
}

func TestComp_InsertLineElement_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.InsertLineElement(0, 10, 10, 100, 100) == nil {
		t.Error("InsertLineElement(0,...) should return error")
	}
}

// ============================================================
// Annotations edge cases
// ============================================================

func TestComp_Annotation_PolygonAnnotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPolygonAnnotation(50, 50, 100, 100,
		[]Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}},
		[3]uint8{255, 0, 0})
	if len(pdf.GetAnnotations()) == 0 {
		t.Error("expected at least 1 annotation")
	}
}

func TestComp_Annotation_LineAnnotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddLineAnnotation(Point{X: 10, Y: 10}, Point{X: 200, Y: 200}, [3]uint8{0, 0, 255})
	found := false
	for _, a := range pdf.GetAnnotations() {
		if a.Type == AnnotLine {
			found = true
		}
	}
	if !found {
		t.Error("expected Line annotation")
	}
}

func TestComp_Annotation_DeleteInvalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.DeleteAnnotation(-1) {
		t.Error("DeleteAnnotation(-1) should return false")
	}
	if pdf.DeleteAnnotation(999) {
		t.Error("DeleteAnnotation(999) should return false")
	}
}

// ============================================================
// Form Fields edge cases
// ============================================================

func TestComp_FormField_RadioButton(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{Type: FormFieldRadio, Name: "radio1", X: 50, Y: 50, W: 20, H: 20})
	if err != nil {
		t.Fatalf("AddFormField radio: %v", err)
	}
	fields := pdf.GetFormFields()
	if len(fields) != 1 || fields[0].Type != FormFieldRadio {
		t.Error("expected 1 radio field")
	}
}

func TestComp_FormField_Button(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{Type: FormFieldButton, Name: "submit", X: 50, Y: 50, W: 100, H: 30, Value: "Submit"})
	if err != nil {
		t.Fatalf("AddFormField button: %v", err)
	}
}

func TestComp_FormField_Signature(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{Type: FormFieldSignature, Name: "sig1", X: 50, Y: 50, W: 200, H: 50})
	if err != nil {
		t.Fatalf("AddFormField signature: %v", err)
	}
}

func TestComp_FormField_ReadOnlyRequired(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "ro", X: 50, Y: 50, W: 200, H: 25,
		ReadOnly: true, Required: true, Value: "Cannot edit",
	})
	if err != nil {
		t.Fatalf("AddFormField readonly: %v", err)
	}
}

func TestComp_FormField_MultilineMaxLen(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "ml", X: 50, Y: 50, W: 200, H: 100,
		Multiline: true, MaxLen: 500,
	})
	if err != nil {
		t.Fatalf("AddFormField multiline: %v", err)
	}
}

func TestComp_FormField_WithColors(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "colored", X: 50, Y: 50, W: 200, H: 25,
		Color: [3]uint8{255, 0, 0}, BorderColor: [3]uint8{0, 0, 255},
		FillColor: [3]uint8{255, 255, 200}, HasBorder: true, HasFill: true,
	})
	if err != nil {
		t.Fatalf("AddFormField colored: %v", err)
	}
}

func TestComp_FormFieldType_String_All(t *testing.T) {
	tests := []struct {
		ft   FormFieldType
		want string
	}{
		{FormFieldText, "Text"}, {FormFieldCheckbox, "Checkbox"},
		{FormFieldRadio, "Radio"}, {FormFieldChoice, "Choice"},
		{FormFieldButton, "Button"}, {FormFieldSignature, "Signature"},
		{FormFieldType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.ft.String(); got != tt.want {
			t.Errorf("FormFieldType(%d).String() = %q, want %q", tt.ft, got, tt.want)
		}
	}
}

// ============================================================
// HTML Parser - additional coverage
// ============================================================

func TestComp_ParseHTML_BasicTags(t *testing.T) {
	nodes := parseHTML("<b>bold</b> <i>italic</i>")
	if len(nodes) == 0 {
		t.Fatal("parseHTML returned no nodes")
	}
}

func TestComp_ParseHTML_VoidElements(t *testing.T) {
	nodes := parseHTML("before<br/>after<hr>end")
	if len(nodes) == 0 {
		t.Fatal("parseHTML returned no nodes")
	}
}

func TestComp_ParseHTML_Entities(t *testing.T) {
	result := decodeHTMLEntities("&amp; &lt; &gt; &quot; &apos; &#65; &#x41;")
	if !strings.Contains(result, "&") || !strings.Contains(result, "<") {
		t.Errorf("entity decoding failed: %q", result)
	}
}

func TestComp_HeadingFontSize(t *testing.T) {
	sizes := map[string]float64{"h1": 24, "h2": 20, "h3": 16, "h4": 14, "h5": 12, "h6": 10}
	for tag, want := range sizes {
		got := headingFontSize(tag)
		if math.Abs(got-want) > 0.01 {
			t.Errorf("headingFontSize(%q) = %f, want %f", tag, got, want)
		}
	}
}

func TestComp_IsVoidElement(t *testing.T) {
	for _, tag := range []string{"br", "hr", "img"} {
		if !isVoidElement(tag) {
			t.Errorf("isVoidElement(%q) = false, want true", tag)
		}
	}
	if isVoidElement("div") {
		t.Error("div should not be void element")
	}
	if isVoidElement("input") {
		t.Error("input is not void in this implementation")
	}
}

func TestComp_ParseInlineStyle(t *testing.T) {
	style := parseInlineStyle("color: red; font-size: 14pt; font-weight: bold")
	if style["color"] != "red" {
		t.Errorf("color = %q, want %q", style["color"], "red")
	}
	if style["font-size"] != "14pt" {
		t.Errorf("font-size = %q, want %q", style["font-size"], "14pt")
	}
}

func TestComp_SplitWords(t *testing.T) {
	words := splitWords("hello world  foo")
	if len(words) != 3 {
		t.Errorf("splitWords returned %d words, want 3: %v", len(words), words)
	}
}

// ============================================================
// XMP Metadata
// ============================================================

func TestComp_XMPMetadata_SetAndGet(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	meta := XMPMetadata{
		Title: "Test Document", Creator: []string{"Author1", "Author2"},
		Description: "A test", Subject: []string{"testing", "pdf"},
		Rights: "Copyright 2025", Language: "en-US",
		CreatorTool: "GoPDF2", Producer: "GoPDF2",
		CreateDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ModifyDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		Keywords: "test,pdf", Trapped: "False",
		PDFAPart: 1, PDFAConformance: "B",
	}
	pdf.SetXMPMetadata(meta)
	got := pdf.GetXMPMetadata()
	if got == nil {
		t.Fatal("GetXMPMetadata returned nil")
	}
	if got.Title != "Test Document" {
		t.Errorf("Title = %q, want %q", got.Title, "Test Document")
	}
	if len(got.Creator) != 2 {
		t.Errorf("Creator count = %d, want 2", len(got.Creator))
	}
}

func TestComp_XMPMetadata_BuildXMP(t *testing.T) {
	meta := &XMPMetadata{
		Title: "Test <Doc> & \"Quotes\"", Creator: []string{"Author"},
		Description: "Desc", CreatorTool: "Tool",
		CreateDate: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	obj := xmpMetadataObj{meta: meta}
	xmp := obj.buildXMP()
	if !strings.Contains(xmp, "Test &lt;Doc&gt; &amp; &quot;Quotes&quot;") {
		t.Error("XML escaping failed in title")
	}
	if !strings.Contains(xmp, "2025-01-01T12:00:00Z") {
		t.Error("missing CreateDate")
	}
}

func TestComp_XmlEscape(t *testing.T) {
	tests := []struct{ input, want string }{
		{"hello", "hello"},
		{"<>&\"'", "&lt;&gt;&amp;&quot;&apos;"},
	}
	for _, tt := range tests {
		if got := xmlEscape(tt.input); got != tt.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ============================================================
// Page Info, Select Pages, Incremental Save, Output
// ============================================================

func TestComp_GetNumberOfPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	if pdf.GetNumberOfPages() != 0 {
		t.Error("new doc should have 0 pages")
	}
	pdf.AddPage()
	if pdf.GetNumberOfPages() != 1 {
		t.Error("after AddPage should have 1 page")
	}
}

func TestComp_SetPage_Valid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	err := pdf.SetPage(1)
	if err != nil {
		t.Fatalf("SetPage(1): %v", err)
	}
	err = pdf.SetPage(2)
	if err != nil {
		t.Fatalf("SetPage(2): %v", err)
	}
}

func TestComp_SetPage_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if pdf.SetPage(0) == nil {
		t.Error("SetPage(0) should return error")
	}
	if pdf.SetPage(5) == nil {
		t.Error("SetPage(5) on 1-page doc should return error")
	}
}

func TestComp_SelectPages_Reverse(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")
	pdf.AddPage()
	pdf.Cell(nil, "Page 3")
	result, err := pdf.SelectPages([]int{3, 2, 1})
	if err != nil {
		t.Fatalf("SelectPages: %v", err)
	}
	if result.GetNumberOfPages() != 3 {
		t.Errorf("pages = %d, want 3", result.GetNumberOfPages())
	}
}

func TestComp_SelectPages_Duplicate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	result, err := pdf.SelectPages([]int{1, 1, 1})
	if err != nil {
		t.Fatalf("SelectPages: %v", err)
	}
	if result.GetNumberOfPages() != 3 {
		t.Errorf("pages = %d, want 3", result.GetNumberOfPages())
	}
}

func TestComp_SelectPages_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if _, err := pdf.SelectPages([]int{}); err == nil {
		t.Error("SelectPages([]) should return error")
	}
}

func TestComp_SelectPages_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if _, err := pdf.SelectPages([]int{5}); err == nil {
		t.Error("SelectPages([5]) on 1-page doc should return error")
	}
}

func TestComp_IncrementalSave_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Original content")
	original, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	pdf2 := &GoPdf{}
	if err := pdf2.OpenPDFFromBytes(original, nil); err != nil {
		t.Fatalf("OpenPDFFromBytes: %v", err)
	}
	if err := pdf2.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf2.SetFont(fontFamily, "", 14)
	pdf2.SetPage(1)
	pdf2.SetXY(50, 100)
	pdf2.Cell(nil, "Added text")
	result, err := pdf2.IncrementalSave(original, nil)
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if !bytes.HasPrefix(result, []byte("%PDF-")) {
		t.Error("incremental save output doesn't start with %PDF-")
	}
	if len(result) <= len(original) {
		t.Error("incremental save should be larger than original")
	}
}

func TestComp_WritePdf_ToFile(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "File output test")
	path := resOutDir + "/comp_write_test.pdf"
	if err := pdf.WritePdf(path); err != nil {
		t.Fatalf("WritePdf: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestComp_WriteTo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "WriteTo test")
	var buf bytes.Buffer
	n, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n == 0 || buf.Len() == 0 {
		t.Error("WriteTo produced empty output")
	}
}

// ============================================================
// Drawing, Color, Text, Position, Graphics State, Info, Compression
// ============================================================

func TestComp_Line(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 200, 200)
}

func TestComp_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(50, 50, 150, 100)
}

func TestComp_Polygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polygon([]Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}}, "D")
}

func TestComp_Polyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Polyline([]Point{{X: 50, Y: 50}, {X: 100, Y: 100}, {X: 150, Y: 50}})
}

func TestComp_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(50, 50, 100, 10, 150, 90, 200, 50, "D")
}

func TestComp_RectFromLowerLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromLowerLeft(50, 50, 100, 50)
}

func TestComp_RectFromUpperLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromUpperLeft(50, 50, 100, 50)
}

func TestComp_SetLineWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(2.0)
	pdf.Line(10, 10, 200, 10)
}

func TestComp_SetLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineType("dashed")
	pdf.Line(10, 10, 200, 10)
}

func TestComp_SetCustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCustomLineType([]float64{5, 3, 1, 3}, 0)
	pdf.Line(10, 10, 200, 10)
}

func TestComp_SetTextColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColor(255, 0, 0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Red text")
}

func TestComp_SetTextColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColorCMYK(0, 100, 100, 0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "CMYK text")
}

func TestComp_SetStrokeColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(0, 0, 255)
	pdf.Line(10, 10, 200, 10)
}

func TestComp_SetFillColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColor(0, 255, 0)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 50, "F")
}

func TestComp_SetGrayFill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
}

func TestComp_SetGrayStroke(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayStroke(0.5)
}

func TestComp_MeasureTextWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth: %v", err)
	}
	if w <= 0 {
		t.Errorf("width = %f, expected > 0", w)
	}
	w2, _ := pdf.MeasureTextWidth("Hello World, this is longer")
	if w2 <= w {
		t.Errorf("longer text width %f should be > %f", w2, w)
	}
}

func TestComp_SplitText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitText("This is a long text that should be split into multiple lines when the width is limited", 100)
	if err != nil {
		t.Fatalf("SplitText: %v", err)
	}
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

func TestComp_MultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	err := pdf.MultiCell(&Rect{W: 200, H: 300}, "This is a multi-cell text that should wrap.")
	if err != nil {
		t.Fatalf("MultiCell: %v", err)
	}
}

func TestComp_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	if err := pdf.Text("Direct text output"); err != nil {
		t.Fatalf("Text: %v", err)
	}
}

func TestComp_Br(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	y1 := pdf.GetY()
	pdf.Br(20)
	y2 := pdf.GetY()
	if y2-y1 != 20 {
		t.Errorf("Br(20) moved Y by %f, want 20", y2-y1)
	}
}

func TestComp_SetGetXY(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetX(100)
	pdf.SetY(200)
	if pdf.GetX() != 100 || pdf.GetY() != 200 {
		t.Errorf("X=%f, Y=%f, want 100, 200", pdf.GetX(), pdf.GetY())
	}
	pdf.SetXY(50, 75)
	if pdf.GetX() != 50 || pdf.GetY() != 75 {
		t.Errorf("after SetXY(50,75): X=%f, Y=%f", pdf.GetX(), pdf.GetY())
	}
}

func TestComp_SaveRestoreGraphicsState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.Line(10, 10, 100, 100)
	pdf.RestoreGraphicsState()
}

func TestComp_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.SetXY(100, 100)
	pdf.Cell(nil, "Rotated")
	pdf.RotateReset()
}

func TestComp_Transparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetTransparency(Transparency{Alpha: 0.5}); err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Semi-transparent")
	pdf.ClearTransparency()
}

func TestComp_SetGetInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	info := PdfInfo{Title: "Test Title", Author: "Test Author", Subject: "Test Subject"}
	pdf.SetInfo(info)
	got := pdf.GetInfo()
	if got.Title != "Test Title" || got.Author != "Test Author" {
		t.Errorf("Info mismatch: %+v", got)
	}
}

func TestComp_SetCompressLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetCompressLevel(9)
	pdf.AddPage()
	pdf.Cell(nil, "Compressed")
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("compressed output is empty")
	}
}

func TestComp_SetNoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.Cell(nil, "Uncompressed")
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("uncompressed output is empty")
	}
}

// ============================================================
// Utility functions, ObjID, Buffer Pool, ContentElementType
// ============================================================

func TestComp_EscapeAnnotString(t *testing.T) {
	tests := []struct{ input, want string }{
		{"hello", "hello"},
		{"(parens)", "\\(parens\\)"},
		{"back\\slash", "back\\\\slash"},
	}
	for _, tt := range tests {
		if got := escapeAnnotString(tt.input); got != tt.want {
			t.Errorf("escapeAnnotString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestComp_EscapeMimeType(t *testing.T) {
	if got := escapeMimeType("application/pdf"); got != "application#2Fpdf" {
		t.Errorf("escapeMimeType = %q, want %q", got, "application#2Fpdf")
	}
}

func TestComp_ContentElementType_String(t *testing.T) {
	tests := []struct {
		t    ContentElementType
		want string
	}{
		{ElementText, "Text"}, {ElementLine, "Line"},
		{ElementRectangle, "Rectangle"}, {ElementOval, "Oval"},
		{ElementImage, "Image"}, {ElementPolygon, "Polygon"},
		{ElementCurve, "Curve"},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("ContentElementType(%d).String() = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestComp_DiagonalAngle(t *testing.T) {
	angle := diagonalAngle(100, 100)
	if math.Abs(angle-45) > 0.01 {
		t.Errorf("diagonalAngle(100,100) = %f, want 45", angle)
	}
}

func TestComp_BufferPool(t *testing.T) {
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}
	buf.WriteString("test data")
	PutBuffer(buf)
	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Errorf("recycled buffer should be empty, got %d bytes", buf2.Len())
	}
	PutBuffer(buf2)
}

func TestComp_ObjID_Methods(t *testing.T) {
	id := ObjID(10)
	if id.Index() != 10 {
		t.Errorf("Index() = %d, want 10", id.Index())
	}
	if id.Ref() != 11 {
		t.Errorf("Ref() = %d, want 11", id.Ref())
	}
	if id.RefStr() != "11 0 R" {
		t.Errorf("RefStr() = %q, want %q", id.RefStr(), "11 0 R")
	}
	if !id.IsValid() {
		t.Error("ObjID(10) should be valid")
	}
}

func TestComp_ObjID_Invalid(t *testing.T) {
	if invalidObjID.IsValid() {
		t.Error("invalidObjID should not be valid")
	}
}

// ============================================================
// Sector drawing
// ============================================================

func TestComp_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 200, 100, 0, 90, "FD")
}

// ============================================================
// Rectangle with rounded corners
// ============================================================

func TestComp_Rectangle_Rounded(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Rectangle(50, 50, 200, 150, "D", 10, 20)
	if err != nil {
		t.Fatalf("Rectangle: %v", err)
	}
}

func TestComp_Rectangle_InvalidCoords(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	// x0 == x1 should be invalid
	err := pdf.Rectangle(50, 50, 50, 150, "D", 0, 0)
	if err == nil {
		t.Error("Rectangle with x0==x1 should return error")
	}
}

// ============================================================
// AddPage with options
// ============================================================

func TestComp_AddPageWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPageWithOption(PageOption{
		PageSize: &Rect{W: 612, H: 792}, // Letter size
	})
	if pdf.GetNumberOfPages() != 1 {
		t.Error("expected 1 page")
	}
}

// ============================================================
// Outlines
// ============================================================

func TestComp_AddOutline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("Chapter 1")
	pdf.AddPage()
	pdf.AddOutline("Chapter 2")
	// Should not panic
}

// ============================================================
// Color Space
// ============================================================

func TestComp_AddColorSpaceRGB(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddColorSpaceRGB("myred", 255, 0, 0)
	if err != nil {
		t.Fatalf("AddColorSpaceRGB: %v", err)
	}
	err = pdf.SetColorSpace("myred")
	if err != nil {
		t.Fatalf("SetColorSpace: %v", err)
	}
}

func TestComp_AddColorSpaceCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddColorSpaceCMYK("mycyan", 100, 0, 0, 0)
	if err != nil {
		t.Fatalf("AddColorSpaceCMYK: %v", err)
	}
}

func TestComp_SetColorSpace_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetColorSpace("nonexistent")
	if err == nil {
		t.Error("SetColorSpace for nonexistent should return error")
	}
}

// ============================================================
// IsFitMultiCell
// ============================================================

func TestComp_IsFitMultiCell(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	fits, height, err := pdf.IsFitMultiCell(&Rect{W: 200, H: 100}, "Short text")
	if err != nil {
		t.Fatalf("IsFitMultiCell: %v", err)
	}
	if !fits {
		t.Error("short text should fit in 200x100")
	}
	if height <= 0 {
		t.Errorf("height = %f, expected > 0", height)
	}
}

// ============================================================
// Margins
// ============================================================

func TestComp_Margins(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetLeftMargin(20)
	pdf.SetTopMargin(30)
	l := pdf.MarginLeft()
	top := pdf.MarginTop()
	if l != 20 {
		t.Errorf("left margin = %f, want 20", l)
	}
	if top != 30 {
		t.Errorf("top margin = %f, want 30", top)
	}
}

// ============================================================
// Close
// ============================================================

func TestComp_Close(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")
	if err := pdf.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// ============================================================
// OCG (Optional Content Groups / Layers)
// ============================================================

func TestComp_OCG_AddAndList(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	ocg1 := pdf.AddOCG(OCG{Name: "Layer1", On: true})
	ocg2 := pdf.AddOCG(OCG{Name: "Layer2", On: false})

	if ocg1.Name != "Layer1" {
		t.Errorf("ocg1.Name = %q, want %q", ocg1.Name, "Layer1")
	}
	if ocg2.Name != "Layer2" {
		t.Errorf("ocg2.Name = %q, want %q", ocg2.Name, "Layer2")
	}

	ocgs := pdf.GetOCGs()
	if len(ocgs) != 2 {
		t.Fatalf("expected 2 OCGs, got %d", len(ocgs))
	}
}

func TestComp_OCG_SetState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "TestLayer", On: true})

	err := pdf.SetOCGState("TestLayer", false)
	if err != nil {
		t.Fatalf("SetOCGState: %v", err)
	}

	err = pdf.SetOCGState("NonExistent", true)
	if err == nil {
		t.Error("SetOCGState for nonexistent layer should return error")
	}
}

func TestComp_OCG_SetStates(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "L1", On: true})
	pdf.AddOCG(OCG{Name: "L2", On: true})

	err := pdf.SetOCGStates(map[string]bool{"L1": false, "L2": true})
	if err != nil {
		t.Fatalf("SetOCGStates: %v", err)
	}
}

func TestComp_OCG_SwitchLayer(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "Exclusive", On: true})
	pdf.AddOCG(OCG{Name: "Other", On: true})

	err := pdf.SwitchLayer("Exclusive", true)
	if err != nil {
		t.Fatalf("SwitchLayer: %v", err)
	}
}

func TestComp_OCG_LayerConfig(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "L1", On: true})

	pdf.AddLayerConfig(LayerConfig{Name: "Config1"})
	configs := pdf.GetLayerConfigs()
	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}
}

// ============================================================
// Page Layout & Page Mode
// ============================================================

func TestComp_PageLayout(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetPageLayout("TwoColumnLeft")
	got := pdf.GetPageLayout()
	if got != "TwoColumnLeft" {
		t.Errorf("GetPageLayout() = %q, want %q", got, "TwoColumnLeft")
	}
}

func TestComp_PageMode(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetPageMode("UseOutlines")
	got := pdf.GetPageMode()
	if got != "UseOutlines" {
		t.Errorf("GetPageMode() = %q, want %q", got, "UseOutlines")
	}
}

// ============================================================
// Document Stats
// ============================================================

func TestComp_DocumentStats(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Stats test")

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Errorf("PageCount = %d, want 1", stats.PageCount)
	}
	if stats.ObjectCount == 0 {
		t.Error("ObjectCount should be > 0")
	}
}

func TestComp_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fonts := pdf.GetFonts()
	if len(fonts) == 0 {
		t.Error("expected at least 1 font")
	}
}

// ============================================================
// TOC
// ============================================================

func TestComp_TOC_GetEmpty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	toc := pdf.GetTOC()
	if len(toc) != 0 {
		t.Errorf("expected empty TOC, got %d items", len(toc))
	}
}

func TestComp_TOC_SetFlat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Chapter 1", PageNo: 1},
		{Level: 1, Title: "Chapter 2", PageNo: 2},
	})
	if err != nil {
		t.Fatalf("SetTOC: %v", err)
	}

	toc := pdf.GetTOC()
	if len(toc) != 2 {
		t.Errorf("expected 2 TOC items, got %d", len(toc))
	}
}

func TestComp_TOC_SetInvalidLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetTOC([]TOCItem{
		{Level: 2, Title: "Bad", PageNo: 1}, // first item must be level 1
	})
	if err == nil {
		t.Error("SetTOC with first item level 2 should return error")
	}
}

// ============================================================
// Page Crop Box
// ============================================================

func TestComp_PageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Crop box test")

	err := pdf.SetPageCropBox(1, Box{Left: 50, Top: 50, Right: 500, Bottom: 700})
	if err != nil {
		t.Fatalf("SetPageCropBox: %v", err)
	}

	box, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox: %v", err)
	}
	if box == nil {
		t.Fatal("GetPageCropBox returned nil")
	}
	if box.Left != 50 || box.Top != 50 || box.Right != 500 || box.Bottom != 700 {
		t.Errorf("crop box = %+v, want {50 50 500 700}", box)
	}

	err = pdf.ClearPageCropBox(1)
	if err != nil {
		t.Fatalf("ClearPageCropBox: %v", err)
	}
}

func TestComp_PageCropBox_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageCropBox(0, Box{})
	if err == nil {
		t.Error("SetPageCropBox(0,...) should return error")
	}
}

// ============================================================
// Page Rotation
// ============================================================

func TestComp_PageRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Rotation test")

	err := pdf.SetPageRotation(1, 90)
	if err != nil {
		t.Fatalf("SetPageRotation: %v", err)
	}

	rot, err := pdf.GetPageRotation(1)
	if err != nil {
		t.Fatalf("GetPageRotation: %v", err)
	}
	if rot != 90 {
		t.Errorf("rotation = %d, want 90", rot)
	}
}

func TestComp_PageRotation_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.SetPageRotation(1, 45) // only 0, 90, 180, 270 are valid
	if err == nil {
		t.Error("SetPageRotation(45) should return error")
	}
}

// ============================================================
// Page Labels
// ============================================================

func TestComp_PageLabels(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPage()
	pdf.Cell(nil, "Page 2")

	labels := []PageLabel{
		{PageIndex: 0, Style: "r", Prefix: ""},
		{PageIndex: 1, Style: "D", Prefix: "Page "},
	}
	pdf.SetPageLabels(labels)

	got := pdf.GetPageLabels()
	if len(got) != 2 {
		t.Errorf("expected 2 labels, got %d", len(got))
	}
}

// ============================================================
// Page Size
// ============================================================

func TestComp_GetPageSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "test")

	w, h, err := pdf.GetPageSize(1)
	if err != nil {
		t.Fatalf("GetPageSize: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("page size = (%f, %f), expected positive", w, h)
	}
}

func TestComp_GetAllPageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(nil, "Page 1")
	pdf.AddPageWithOption(PageOption{PageSize: &Rect{W: 612, H: 792}})
	pdf.Cell(nil, "Page 2")

	sizes := pdf.GetAllPageSizes()
	if len(sizes) != 2 {
		t.Errorf("expected 2 page sizes, got %d", len(sizes))
	}
}

// ============================================================
// Stoke/Fill CMYK
// ============================================================

func TestComp_SetStrokeColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColorCMYK(100, 0, 0, 0)
	pdf.Line(10, 10, 200, 10)
}

func TestComp_SetFillColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetFillColorCMYK(0, 100, 0, 0)
}

// ============================================================
// Image operations
// ============================================================

func TestComp_Image_JPEG(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.Image(resJPEGPath, 50, 50, &Rect{W: 200, H: 150})
	if err != nil {
		t.Skipf("image not available: %v", err)
	}
}

// ============================================================
// External/Internal Links
// ============================================================

func TestComp_AddExternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Click here")
	pdf.AddExternalLink("https://example.com", 50, 50, 100, 20)
}

func TestComp_AddInternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetAnchor("target")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Go to target")
	pdf.AddInternalLink("target", 50, 50, 100, 20)
}

// ============================================================
// SplitTextWithWordWrap
// ============================================================

func TestComp_SplitTextWithWordWrap(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitTextWithWordWrap("This is a test of word wrapping functionality", 100)
	if err != nil {
		t.Fatalf("SplitTextWithWordWrap: %v", err)
	}
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

// ============================================================
// MeasureCellHeightByText
// ============================================================

func TestComp_MeasureCellHeightByText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	h, err := pdf.MeasureCellHeightByText("Hello")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Errorf("height = %f, expected > 0", h)
	}
}

// ============================================================
// IsCurrFontContainGlyph
// ============================================================

func TestComp_IsCurrFontContainGlyph(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph: %v", err)
	}
	if !ok {
		t.Error("font should contain 'A'")
	}
}

// ============================================================
// GetNextObjectID / GetNumberOfPages
// ============================================================

func TestComp_GetNextObjectID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	id := pdf.GetNextObjectID()
	if id <= 0 {
		t.Errorf("GetNextObjectID() = %d, expected > 0", id)
	}
}

// ============================================================
// SetCharSpacing
// ============================================================

func TestComp_SetCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetCharSpacing(2.0)
	if err != nil {
		t.Fatalf("SetCharSpacing: %v", err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Spaced text")
}

// ============================================================
// SetFontSize
// ============================================================

func TestComp_SetFontSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetFontSize(24)
	if err != nil {
		t.Fatalf("SetFontSize: %v", err)
	}
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "Large text")
}
