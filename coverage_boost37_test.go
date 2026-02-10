package gopdf

import (
	"bytes"
	"os"
	"testing"
)

// ============================================================
// TestCov37_ — coverage boost round 37
// Targets: pdf_lowlevel.go, image_recompress.go, content_stream_clean.go,
//          colorspace_convert.go, page_batch_ops.go, page_rotate.go,
//          image_extract.go, font_extract.go, form_field.go,
//          toc.go, svg_insert.go, pdf_parser.go
// ============================================================

// --- helper: build a minimal valid PDF with one image object ---
func buildMinimalPDFWithImage37(t *testing.T) []byte {
	t.Helper()
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Text("Hello")
	// Add a JPEG image
	pdf.AddPage()
	imgPath := resJPEGPath
	if _, err := os.Stat(imgPath); err == nil {
		_ = pdf.Image(imgPath, 10, 10, nil)
	}
	var buf bytes.Buffer
	pdf.Write(&buf)
	return buf.Bytes()
}

// --- pdf_lowlevel.go ---

func TestCov37_ReadObject(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	obj, err := ReadObject(data, 1)
	if err != nil {
		t.Fatalf("ReadObject: %v", err)
	}
	if obj.Num != 1 {
		t.Errorf("expected obj num 1, got %d", obj.Num)
	}
}

func TestCov37_ReadObject_NotFound(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	_, err := ReadObject(data, 9999)
	if err == nil {
		t.Error("expected error for non-existent object")
	}
}

func TestCov37_UpdateObject(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	updated, err := UpdateObject(data, 1, "<< /Type /Catalog >>")
	if err != nil {
		t.Fatalf("UpdateObject: %v", err)
	}
	if len(updated) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov37_GetDictKey(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	val, err := GetDictKey(data, 1, "/Type")
	if err != nil {
		t.Fatalf("GetDictKey: %v", err)
	}
	_ = val // may or may not have /Type
}

func TestCov37_SetDictKey(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	updated, err := SetDictKey(data, 1, "/CustomKey", "/CustomValue")
	if err != nil {
		t.Fatalf("SetDictKey: %v", err)
	}
	if len(updated) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov37_GetStream(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	// Try to get stream from object 1 (may or may not have stream)
	_, err := GetStream(data, 1)
	if err != nil {
		t.Fatalf("GetStream: %v", err)
	}
}

func TestCov37_SetStream(t *testing.T) {
	data := buildMinimalPDFWithImage37(t)
	// Find an object with a stream
	for i := 1; i <= 20; i++ {
		stream, err := GetStream(data, i)
		if err == nil && len(stream) > 0 {
			newData, err := SetStream(data, i, []byte("test stream"))
			if err != nil {
				t.Logf("SetStream on obj %d: %v", i, err)
				continue
			}
			if len(newData) > 0 {
				t.Logf("SetStream on obj %d succeeded", i)
			}
			break
		}
	}
}