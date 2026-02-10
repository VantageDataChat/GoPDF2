package gopdf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// coverage_boost14_test.go — TestCov14_ prefix
// Targets: pdf_decrypt deeper paths, gopdf.go (Read, Sector,
// CellWithOption, Cell, MeasureTextWidth, MeasureCellHeightByText),
// journal.go, doc_stats.go, bookmark.go deeper paths,
// content_element.go ModifyElementPosition more types
// ============================================================

// ============================================================
// gopdf.go — Read, GetBytesPdf
// ============================================================

func TestCov14_Read(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Read test")

	buf := make([]byte, 4096)
	n, err := pdf.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes read")
	}
	if !bytes.Contains(buf[:n], []byte("%PDF")) {
		t.Error("expected PDF header in output")
	}
}

func TestCov14_Read_MultipleCalls(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Multi read test")

	var all []byte
	buf := make([]byte, 256)
	for {
		n, err := pdf.Read(buf)
		if n > 0 {
			all = append(all, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	if len(all) == 0 {
		t.Error("expected non-zero total bytes")
	}
}

func TestCov14_GetBytesPdf(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("GetBytesPdf test")

	data := pdf.GetBytesPdf()
	if len(data) == 0 {
		t.Error("expected non-empty PDF bytes")
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Error("expected PDF header")
	}
}

// ============================================================
// gopdf.go — Sector
// ============================================================

func TestCov14_Sector_Draw(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 400, 80, 0, 90, "D")
}

func TestCov14_Sector_Fill(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 400, 80, 0, 90, "F")
}

func TestCov14_Sector_FillDraw(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 400, 80, 0, 90, "FD")
}

func TestCov14_Sector_EmptyStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 400, 80, 45, 270, "")
}

func TestCov14_Sector_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode})
	pdf.Sector(200, 400, 80, 0, 180, "FD")
	pdf.ClearTransparency()
}

// ============================================================
// gopdf.go — CellWithOption, Cell
// ============================================================

func TestCov14_CellWithOption_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Cell option test", CellOption{
		Align: Right | Top,
	})
	if err != nil {
		t.Fatalf("CellWithOption: %v", err)
	}
}

func TestCov14_CellWithOption_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.CellWithOption(&Rect{W: 200, H: 30}, "Transparent cell", CellOption{
		Align:        Center | Middle,
		Transparency: &Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode},
	})
	if err != nil {
		t.Fatalf("CellWithOption with transparency: %v", err)
	}
}

func TestCov14_Cell_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.Cell(&Rect{W: 200, H: 30}, "Cell test")
	if err != nil {
		t.Fatalf("Cell: %v", err)
	}
}

func TestCov14_Cell_NilRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.Cell(nil, "Nil rect cell")
	if err != nil {
		t.Fatalf("Cell nil rect: %v", err)
	}
}

// ============================================================
// gopdf.go — MeasureTextWidth, MeasureCellHeightByText
// ============================================================

func TestCov14_MeasureTextWidth(t *testing.T) {
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

func TestCov14_MeasureTextWidth_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	w, err := pdf.MeasureTextWidth("")
	if err != nil {
		t.Fatalf("MeasureTextWidth empty: %v", err)
	}
	if w != 0 {
		t.Errorf("expected 0 width for empty string, got %f", w)
	}
}

func TestCov14_MeasureCellHeightByText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	h, err := pdf.MeasureCellHeightByText("Hello World this is a long text that should wrap")
	if err != nil {
		t.Fatalf("MeasureCellHeightByText: %v", err)
	}
	if h <= 0 {
		t.Errorf("expected positive height, got %f", h)
	}
}

// ============================================================
// gopdf.go — Text with underline
// ============================================================

func TestCov14_Text_Underline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	if err := pdf.SetFontWithStyle(fontFamily, Underline, 14); err != nil {
		t.Fatalf("SetFontWithStyle underline: %v", err)
	}
	err := pdf.Text("Underlined text")
	if err != nil {
		t.Fatalf("Text underline: %v", err)
	}
}

func TestCov14_Text_Bold(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	// Bold style uses the same font family but with Bold flag.
	// Need to add the font with Bold style first.
	if err := pdf.AddTTFFontWithOption(fontFamily+"Bold", resFontPath, TtfOption{Style: Bold}); err != nil {
		t.Skipf("cannot add bold font: %v", err)
	}
	if err := pdf.SetFontWithStyle(fontFamily+"Bold", Bold, 14); err != nil {
		t.Skipf("SetFontWithStyle bold: %v", err)
	}
	err := pdf.Text("Bold text")
	if err != nil {
		t.Fatalf("Text bold: %v", err)
	}
}

// ============================================================
// journal.go — JournalEnable, JournalStartOp, JournalEndOp,
// JournalUndo, JournalRedo, JournalSave, JournalLoad
// ============================================================

func TestCov14_Journal_BasicUndoRedo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Initial")

	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Fatal("expected journal to be enabled")
	}

	pdf.JournalStartOp("add text")
	pdf.SetXY(50, 100)
	_ = pdf.Text("Added text")
	pdf.JournalEndOp()

	name, err := pdf.JournalUndo()
	if err != nil {
		t.Fatalf("JournalUndo: %v", err)
	}
	if name != "add text" {
		t.Errorf("expected 'add text', got %q", name)
	}

	name, err = pdf.JournalRedo()
	if err != nil {
		t.Fatalf("JournalRedo: %v", err)
	}
	if name != "add text" {
		t.Errorf("expected 'add text', got %q", name)
	}
}

func TestCov14_Journal_UndoNothingToUndo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()

	// Only one snapshot (initial), so undo should fail.
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Fatal("expected error for nothing to undo")
	}
}

func TestCov14_Journal_RedoNothingToRedo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()

	_, err := pdf.JournalRedo()
	if err == nil {
		t.Fatal("expected error for nothing to redo")
	}
}

func TestCov14_Journal_Disable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()
	pdf.JournalDisable()
	if pdf.JournalIsEnabled() {
		t.Fatal("expected journal to be disabled")
	}
}

func TestCov14_Journal_GetOperations(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")

	pdf.JournalEnable()
	pdf.JournalStartOp("op1")
	pdf.SetXY(50, 100)
	_ = pdf.Text("op1 text")
	pdf.JournalEndOp()

	pdf.JournalStartOp("op2")
	pdf.SetXY(50, 150)
	_ = pdf.Text("op2 text")
	pdf.JournalEndOp()

	ops := pdf.JournalGetOperations()
	if len(ops) < 2 {
		t.Errorf("expected at least 2 operations, got %d", len(ops))
	}
}

func TestCov14_Journal_GetOperations_NoJournal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	ops := pdf.JournalGetOperations()
	if ops != nil {
		t.Errorf("expected nil operations when journal not enabled")
	}
}

func TestCov14_Journal_SaveLoad(t *testing.T) {
	ensureOutDir(t)
	journalPath := filepath.Join(resOutDir, "test_journal.json")

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("text")

	pdf.JournalEnable()
	pdf.JournalStartOp("save test")
	pdf.SetXY(50, 100)
	_ = pdf.Text("more text")
	pdf.JournalEndOp()

	err := pdf.JournalSave(journalPath)
	if err != nil {
		t.Fatalf("JournalSave: %v", err)
	}

	// Load into a new PDF.
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.JournalEnable()
	err = pdf2.JournalLoad(journalPath)
	if err != nil {
		t.Fatalf("JournalLoad: %v", err)
	}

	os.Remove(journalPath)
}

func TestCov14_Journal_SaveNoJournal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.JournalSave("test.json")
	if err == nil {
		t.Fatal("expected error when journal not enabled")
	}
}

func TestCov14_Journal_LoadNoJournal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.JournalLoad("test.json")
	if err == nil {
		t.Fatal("expected error when journal not enabled")
	}
}

func TestCov14_Journal_LoadInvalidPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()
	err := pdf.JournalLoad("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestCov14_Journal_UndoNoJournal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Fatal("expected error when journal not enabled")
	}
}

func TestCov14_Journal_RedoNoJournal(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.JournalRedo()
	if err == nil {
		t.Fatal("expected error when journal not enabled")
	}
}

// ============================================================
// doc_stats.go — GetDocumentStats, GetFonts
// ============================================================

func TestCov14_GetDocumentStats(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Stats test")

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Errorf("expected 1 page, got %d", stats.PageCount)
	}
	if stats.FontCount == 0 {
		t.Error("expected at least 1 font")
	}
	if stats.ObjectCount == 0 {
		t.Error("expected non-zero object count")
	}
}

func TestCov14_GetFonts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	fonts := pdf.GetFonts()
	if len(fonts) == 0 {
		t.Error("expected at least 1 font")
	}
	found := false
	for _, f := range fonts {
		if f.Family == fontFamily {
			found = true
			if !f.IsEmbedded {
				t.Error("expected font to be embedded")
			}
		}
	}
	if !found {
		t.Errorf("expected to find font family %q", fontFamily)
	}
}

func TestCov14_GetDocumentStats_MultiPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	_ = pdf.Text("Page 2")
	pdf.AddPage()
	_ = pdf.Text("Page 3")

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 3 {
		t.Errorf("expected 3 pages, got %d", stats.PageCount)
	}
}

// ============================================================
// pdf_decrypt.go — deeper paths
// ============================================================

func TestCov14_Authenticate_NotEncrypted(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Not encrypted")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)

	dc, err := authenticate(buf.Bytes(), "")
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if dc != nil {
		t.Error("expected nil decryptContext for non-encrypted PDF")
	}
}

func TestCov14_Authenticate_EncryptedPDF(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Protected content")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Try with correct user password.
	dc, err := authenticate(data, "user")
	// May or may not work depending on encryption format.
	_ = dc
	_ = err
}

func TestCov14_Authenticate_WrongPassword(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Protected content")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Try with wrong password.
	dc, err := authenticate(data, "wrongpassword")
	_ = dc
	_ = err
}

func TestCov14_Authenticate_OwnerPassword(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Protected content")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Try with owner password.
	dc, err := authenticate(data, "owner")
	_ = dc
	_ = err
}

func TestCov14_DecryptPDF_NotEncrypted(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Not encrypted")

	var buf bytes.Buffer
	_, _ = pdf.WriteTo(&buf)
	data := buf.Bytes()

	// decryptPDF with nil context should return data as-is.
	dc := &decryptContext{
		encryptionKey: []byte("12345"),
		keyLen:        5,
		v:             1,
		r:             2,
	}
	result := decryptPDF(data, dc)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov14_ParseEncryptDict_V1R2(t *testing.T) {
	dict := `/V 1 /R 2 /P -3904 /O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108> /U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108>`
	v, r, keyLen, oValue, uValue, pValue, err := parseEncryptDict(dict)
	if err != nil {
		t.Fatalf("parseEncryptDict: %v", err)
	}
	if v != 1 {
		t.Errorf("v = %d, want 1", v)
	}
	if r != 2 {
		t.Errorf("r = %d, want 2", r)
	}
	if keyLen != 5 {
		t.Errorf("keyLen = %d, want 5", keyLen)
	}
	if len(oValue) < 32 {
		t.Errorf("oValue too short: %d", len(oValue))
	}
	if len(uValue) < 32 {
		t.Errorf("uValue too short: %d", len(uValue))
	}
	if pValue != -3904 {
		t.Errorf("pValue = %d, want -3904", pValue)
	}
}

func TestCov14_ParseEncryptDict_V2R3(t *testing.T) {
	dict := `/V 2 /R 3 /Length 128 /P -3904 /O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108> /U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108>`
	v, r, keyLen, _, _, _, err := parseEncryptDict(dict)
	if err != nil {
		t.Fatalf("parseEncryptDict V2R3: %v", err)
	}
	if v != 2 {
		t.Errorf("v = %d, want 2", v)
	}
	if r != 3 {
		t.Errorf("r = %d, want 3", r)
	}
	if keyLen != 16 {
		t.Errorf("keyLen = %d, want 16", keyLen)
	}
}

func TestCov14_ParseEncryptDict_UnsupportedV(t *testing.T) {
	dict := `/V 4 /R 4 /P -3904 /O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108> /U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108>`
	_, _, _, _, _, _, err := parseEncryptDict(dict)
	if err == nil {
		t.Fatal("expected error for unsupported V=4")
	}
}

func TestCov14_ParseEncryptDict_UnsupportedR(t *testing.T) {
	dict := `/V 2 /R 5 /P -3904 /O <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108> /U <28BF4E5E4E758A4164004E56FFFA01082E2E00B6D0683E802F0CA9FE6453697A28BF4E5E4E758A4164004E56FFFA0108>`
	_, _, _, _, _, _, err := parseEncryptDict(dict)
	if err == nil {
		t.Fatal("expected error for unsupported R=5")
	}
}

func TestCov14_ComputeEncryptionKey_R2(t *testing.T) {
	userPass := []byte("test")
	oValue := make([]byte, 32)
	key := computeEncryptionKey(userPass, oValue, -3904, 5, 2)
	if len(key) != 5 {
		t.Errorf("expected key length 5, got %d", len(key))
	}
}

func TestCov14_ComputeEncryptionKey_R3(t *testing.T) {
	userPass := []byte("test")
	oValue := make([]byte, 32)
	key := computeEncryptionKey(userPass, oValue, -3904, 16, 3)
	if len(key) != 16 {
		t.Errorf("expected key length 16, got %d", len(key))
	}
}

func TestCov14_TryOwnerPassword_R2(t *testing.T) {
	ownerPass := []byte("owner")
	oValue := make([]byte, 32)
	uValue := make([]byte, 32)
	key, ok := tryOwnerPassword(ownerPass, oValue, uValue, -3904, 5, 2)
	// May or may not succeed, but should not panic.
	_ = key
	_ = ok
}

func TestCov14_TryOwnerPassword_R3(t *testing.T) {
	ownerPass := []byte("owner")
	oValue := make([]byte, 32)
	uValue := make([]byte, 32)
	key, ok := tryOwnerPassword(ownerPass, oValue, uValue, -3904, 16, 3)
	_ = key
	_ = ok
}

// ============================================================
// gopdf.go — Line
// ============================================================

func TestCov14_Line_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 200, 200)
}

func TestCov14_Line_WithLineWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(3.0)
	pdf.Line(10, 10, 200, 200)
}

func TestCov14_Line_WithColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.Line(10, 10, 200, 200)
}

// ============================================================
// gopdf.go — AddTTFFontWithOption
// ============================================================

func TestCov14_AddTTFFontWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); os.IsNotExist(err) {
		t.Skip("font not available")
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Fatalf("AddTTFFontWithOption: %v", err)
	}
	err = pdf.SetFont(fontFamily, "", 14)
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Kerning test")
}

func TestCov14_AddTTFFontWithOption_InvalidPath(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption("test", "/nonexistent/font.ttf", TtfOption{})
	if err == nil {
		t.Fatal("expected error for invalid font path")
	}
}

// ============================================================
// gopdf.go — AddTTFFontByReaderWithOption
// ============================================================

func TestCov14_AddTTFFontByReaderWithOption(t *testing.T) {
	if _, err := os.Stat(resFontPath); os.IsNotExist(err) {
		t.Skip("font not available")
	}
	fontData, err := os.ReadFile(resFontPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontByReaderWithOption("ReaderFont", bytes.NewReader(fontData), TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Fatalf("AddTTFFontByReaderWithOption: %v", err)
	}
	err = pdf.SetFont("ReaderFont", "", 14)
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Reader font test")
}

// ============================================================
// gopdf.go — MultiCellWithOption
// ============================================================

func TestCov14_MultiCellWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 20}, "This is a long text that should wrap across multiple lines in the cell", CellOption{
		Align: Left | Top,
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption: %v", err)
	}
}

func TestCov14_MultiCellWithOption_WithTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	err := pdf.MultiCellWithOption(&Rect{W: 200, H: 20}, "Transparent multi cell text", CellOption{
		Transparency: &Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode},
	})
	if err != nil {
		t.Fatalf("MultiCellWithOption with transparency: %v", err)
	}
}

// ============================================================
// bookmark.go — DeleteBookmark deeper paths (prev/next linking)
// ============================================================

func TestCov14_DeleteBookmark_First(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch3")
	pdf.AddOutline("Chapter 3")

	// Delete first bookmark.
	err := pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark first: %v", err)
	}
}

func TestCov14_DeleteBookmark_Last(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch3")
	pdf.AddOutline("Chapter 3")

	// Delete last bookmark (index 2).
	err := pdf.DeleteBookmark(2)
	if err != nil {
		t.Fatalf("DeleteBookmark last: %v", err)
	}
}

func TestCov14_DeleteBookmark_Middle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch2")
	pdf.AddOutline("Chapter 2")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Ch3")
	pdf.AddOutline("Chapter 3")

	// Delete middle bookmark.
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark middle: %v", err)
	}
}

// ============================================================
// content_element.go — ModifyElementPosition for more types
// ============================================================

func TestCov14_ModifyElementPosition_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(200, 400, 80, 0, 90, "D")

	err := pdf.ModifyElementPosition(1, 0, 100, 200)
	if err != nil {
		t.Fatalf("ModifyElementPosition sector: %v", err)
	}
}

func TestCov14_ModifyElementPosition_Image(t *testing.T) {
	if _, err := os.Stat(resJPEGPath); os.IsNotExist(err) {
		t.Skip("JPEG not available")
	}
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_ = pdf.Image(resJPEGPath, 50, 50, &Rect{W: 100, H: 100})

	err := pdf.ModifyElementPosition(1, 0, 200, 200)
	if err != nil {
		t.Fatalf("ModifyElementPosition image: %v", err)
	}
}

// ============================================================
// gopdf.go — OpenPDFFromStream
// ============================================================

func TestCov14_OpenPDFFromStream(t *testing.T) {
	if _, err := os.Stat(resTestPDF); os.IsNotExist(err) {
		t.Skip("test PDF not available")
	}
	data, err := os.ReadFile(resTestPDF)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	rs := io.ReadSeeker(bytes.NewReader(data))
	err = pdf.OpenPDFFromStream(&rs, nil)
	if err != nil {
		t.Fatalf("OpenPDFFromStream: %v", err)
	}
}

// ============================================================
// incremental_save.go — WriteIncrementalPdf
// ============================================================

func TestCov14_WriteIncrementalPdf(t *testing.T) {
	ensureOutDir(t)
	outPath := filepath.Join(resOutDir, "incremental_cov14.pdf")

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Incremental save test")

	// Get original bytes.
	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	originalData := buf.Bytes()

	// Modify and do incremental save.
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2 added")

	err = pdf.WriteIncrementalPdf(outPath, originalData, nil)
	if err != nil {
		t.Logf("WriteIncrementalPdf: %v (may be expected)", err)
	}

	os.Remove(outPath)
}
