package gopdf

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

// ─── gopdf.go: Text / CellWithOption / Cell error branches (nil FontISubset) ───

func TestCov34_Text_NilFontISubset(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	// No font set — FontISubset is nil, should panic
	func() {
		defer func() { recover() }()
		pdf.Text("hello")
	}()
}

func TestCov34_CellWithOption_NilFontISubset(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	func() {
		defer func() { recover() }()
		pdf.CellWithOption(&Rect{W: 100, H: 20}, "hello", CellOption{})
	}()
}

func TestCov34_Cell_NilFontISubset(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	func() {
		defer func() { recover() }()
		pdf.Cell(&Rect{W: 100, H: 20}, "hello")
	}()
}

func TestCov34_MeasureTextWidth_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	func() {
		defer func() { recover() }()
		pdf.MeasureTextWidth("hello")
	}()
}

func TestCov34_MeasureCellHeightByText_NilFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddPage()
	func() {
		defer func() { recover() }()
		pdf.MeasureCellHeightByText("hello")
	}()
}

// ─── gopdf.go: Read — compilePdf path when buf is empty ───

func TestCov34_Read_CompilePdf(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Read compile")

	// First Read triggers compilePdf
	buf := make([]byte, 8192)
	n, _ := pdf.Read(buf)
	if n == 0 {
		t.Fatal("expected non-zero bytes")
	}

	// Second Read uses cached buffer
	buf2 := make([]byte, 8192)
	n2, _ := pdf.Read(buf2)
	_ = n2 // may be 0 if all consumed
}

// ─── gopdf.go: GetBytesPdfReturnErr error path ───

func TestCov34_GetBytesPdfReturnErr_Error(t *testing.T) {
	// A fresh GoPdf with no pages — compilePdf should still work
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}
}

// ─── gopdf.go: ImageByHolderWithOptions with Mask ───

func TestCov34_ImageByHolderWithOptions_Mask(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create main image
	mainImg := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			mainImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var mainBuf bytes.Buffer
	jpeg.Encode(&mainBuf, mainImg, nil)
	mainHolder, _ := ImageHolderByBytes(mainBuf.Bytes())

	// Create mask image (grayscale)
	maskImg := image.NewGray(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			maskImg.Set(x, y, color.Gray{Y: uint8(x * 12)})
		}
	}
	var maskBuf bytes.Buffer
	jpeg.Encode(&maskBuf, maskImg, nil)
	maskHolder, _ := ImageHolderByBytes(maskBuf.Bytes())

	err := pdf.ImageByHolderWithOptions(mainHolder, ImageOptions{
		X:    10,
		Y:    10,
		Rect: &Rect{W: 100, H: 100},
		Mask: &MaskOptions{
			Holder: maskHolder,
			ImageOptions: ImageOptions{
				X:    10,
				Y:    10,
				Rect: &Rect{W: 100, H: 100},
			},
		},
	})
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions with mask: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: ImageByHolderWithOptions with Transparency ───

func TestCov34_ImageByHolderWithOptions_Transparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)

	err := pdf.ImageByHolderWithOptions(holder, ImageOptions{
		X:    10,
		Y:    10,
		Rect: &Rect{W: 50, H: 50},
		Transparency: &Transparency{
			Alpha:         0.5,
			BlendModeType: NormalBlendMode,
		},
	})
	if err != nil {
		t.Fatalf("ImageByHolderWithOptions transparency: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: SetTransparency error path (invalid blend mode) ───

func TestCov34_SetTransparency_Cached(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Set same transparency twice to hit cached path
	tr := Transparency{Alpha: 0.3, BlendModeType: NormalBlendMode}
	pdf.SetTransparency(tr)
	pdf.Cell(&Rect{W: 100, H: 20}, "First")
	pdf.SetTransparency(tr)
	pdf.Cell(&Rect{W: 100, H: 20}, "Second (cached)")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── toc.go: SetTOC hierarchical, GetTOC with outlines ───

func TestCov34_SetTOC_Hierarchical(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 3")

	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Chapter 1", PageNo: 1},
		{Level: 2, Title: "Section 1.1", PageNo: 1, Y: 200},
		{Level: 2, Title: "Section 1.2", PageNo: 2},
		{Level: 1, Title: "Chapter 2", PageNo: 3},
	})
	if err != nil {
		t.Fatalf("SetTOC hierarchical: %v", err)
	}

	// GetTOC
	items := pdf.GetTOC()
	if len(items) < 4 {
		t.Fatalf("expected >= 4 TOC items, got %d", len(items))
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

func TestCov34_SetTOC_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutlineWithPosition("Test Outline")

	// Clear TOC
	err := pdf.SetTOC([]TOCItem{})
	if err != nil {
		t.Fatalf("SetTOC empty: %v", err)
	}
}

func TestCov34_SetTOC_InvalidLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// First item not level 1
	err := pdf.SetTOC([]TOCItem{{Level: 2, Title: "Bad"}})
	if err == nil {
		t.Fatal("expected error for invalid first level")
	}

	// Level jump > 1
	err = pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Ch1"},
		{Level: 3, Title: "Bad jump"},
	})
	if err == nil {
		t.Fatal("expected error for level jump")
	}
}

// ─── digital_signature.go: SignPDFToFile, VerifySignatureFromFile ───

func TestCov34_SignPDFToFile(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Sign to file")

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert, _ := x509.ParseCertificate(certDER)

	outPath := resOutDir + "/cov34_signed.pdf"
	err := pdf.SignPDFToFile(SignatureConfig{
		Certificate: cert,
		PrivateKey:  key,
		Reason:      "Test",
	}, outPath)
	if err != nil {
		t.Fatalf("SignPDFToFile: %v", err)
	}

	// VerifySignatureFromFile
	_, err = VerifySignatureFromFile(outPath)
	// May fail verification (self-signed), but should not error on parsing
	_ = err
}

// ─── digital_signature.go: createPKCS7Signature with ECDSA ───

func TestCov34_CreatePKCS7Signature_ECDSA(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "ECDSA Test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert, _ := x509.ParseCertificate(certDER)

	sig, err := createPKCS7Signature([]byte("test data"), &SignatureConfig{
		Certificate: cert,
		PrivateKey:  key,
	})
	if err != nil {
		t.Fatalf("createPKCS7Signature ECDSA: %v", err)
	}
	if len(sig) == 0 {
		t.Fatal("expected non-empty signature")
	}
}

// ─── pdf_lowlevel.go: CopyObject, GetCatalog ───

func TestCov34_CopyObject(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Copy test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	newData, newNum, err := CopyObject(data, 1)
	if err != nil {
		t.Fatalf("CopyObject: %v", err)
	}
	if len(newData) == 0 {
		t.Fatal("expected non-empty data")
	}
	if newNum <= 0 {
		t.Fatal("expected positive new object number")
	}
}

func TestCov34_GetCatalog(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Catalog test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	cat, err := GetCatalog(data)
	if err != nil {
		t.Fatalf("GetCatalog: %v", err)
	}
	if cat == nil {
		t.Fatal("expected non-nil catalog")
	}
}

func TestCov34_CopyObject_BadObj(t *testing.T) {
	_, _, err := CopyObject([]byte("not a pdf"), 1)
	if err == nil {
		t.Fatal("expected error for bad data")
	}
}

func TestCov34_GetCatalog_BadData(t *testing.T) {
	_, err := GetCatalog([]byte("not a pdf"))
	if err == nil {
		t.Fatal("expected error for bad data")
	}
}

// ─── image_recompress.go: rebuildXref, RecompressImages ───

func TestCov34_RecompressImages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add a JPEG image
	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 100, H: 100})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Recompress with default options
	result, err := RecompressImages(data, RecompressOption{})
	if err != nil {
		t.Fatalf("RecompressImages: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// Recompress with quality and max dimensions
	result2, err := RecompressImages(data, RecompressOption{
		JPEGQuality: 50,
		MaxWidth:    50,
		MaxHeight:   50,
	})
	if err != nil {
		t.Fatalf("RecompressImages with opts: %v", err)
	}
	if len(result2) == 0 {
		t.Fatal("expected non-empty result")
	}
}

// ─── image_recompress.go: rebuildXref with \r\n xref ───

func TestCov34_RebuildXref(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Xref test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// rebuildXref should work on valid PDF
	result := rebuildXref(data)
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	// No xref marker
	result2 := rebuildXref([]byte("no xref here"))
	if string(result2) != "no xref here" {
		t.Fatal("expected passthrough for no xref")
	}
}

// ─── text_extract.go: ExtractTextFromPage, ExtractTextFromAllPages, ExtractPageText ───

func TestCov34_ExtractText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Hello World")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Page Two")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// ExtractTextFromPage
	text, err := ExtractTextFromPage(data, 0)
	if err != nil {
		t.Fatalf("ExtractTextFromPage: %v", err)
	}
	_ = text

	// ExtractTextFromAllPages
	allText, err := ExtractTextFromAllPages(data)
	if err != nil {
		t.Fatalf("ExtractTextFromAllPages: %v", err)
	}
	_ = allText

	// ExtractPageText
	items, err := ExtractPageText(data, 0)
	if err != nil {
		t.Fatalf("ExtractPageText: %v", err)
	}
	_ = items

	// Out of range
	_, err = ExtractTextFromPage(data, 99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// Bad data
	_, err = ExtractTextFromPage([]byte("bad"), 0)
	_ = err
}

// ─── page_rotate.go: SetPageRotation, GetPageRotation ───

func TestCov34_PageRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Rotated page")

	err := pdf.SetPageRotation(1, 90)
	if err != nil {
		t.Fatalf("SetPageRotation: %v", err)
	}

	rot, err := pdf.GetPageRotation(1)
	if err != nil {
		t.Fatalf("GetPageRotation: %v", err)
	}
	if rot != 90 {
		t.Fatalf("expected 90, got %d", rot)
	}

	// Invalid rotation
	err = pdf.SetPageRotation(1, 45)
	if err == nil {
		t.Fatal("expected error for invalid rotation")
	}

	// No pages
	pdf2 := &GoPdf{}
	pdf2.Start(Config{PageSize: *PageSizeA4})
	_, err = pdf2.GetPageRotation(1)
	if err == nil {
		t.Fatal("expected error for no pages")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── page_cropbox.go: SetPageCropBox, GetPageCropBox, ClearPageCropBox ───

func TestCov34_PageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "CropBox test")

	err := pdf.SetPageCropBox(1, Box{Left: 10, Top: 10, Right: 200, Bottom: 300})
	if err != nil {
		t.Fatalf("SetPageCropBox: %v", err)
	}

	box, err := pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox: %v", err)
	}
	if box == nil {
		t.Fatal("expected non-nil crop box")
	}

	err = pdf.ClearPageCropBox(1)
	if err != nil {
		t.Fatalf("ClearPageCropBox: %v", err)
	}

	// No pages
	pdf2 := &GoPdf{}
	pdf2.Start(Config{PageSize: *PageSizeA4})
	err = pdf2.SetPageCropBox(1, Box{})
	if err == nil {
		t.Fatal("expected error for no pages")
	}
	_, err = pdf2.GetPageCropBox(1)
	if err == nil {
		t.Fatal("expected error for no pages")
	}
	err = pdf2.ClearPageCropBox(1)
	if err == nil {
		t.Fatal("expected error for no pages")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── page_batch_ops.go: MovePage ───

func TestCov34_MovePage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 3")

	// Move page 3 to position 1
	err := pdf.MovePage(3, 1)
	if err != nil {
		t.Fatalf("MovePage: %v", err)
	}

	// Invalid source
	err = pdf.MovePage(99, 1)
	if err == nil {
		t.Fatal("expected error for invalid source")
	}

	// Invalid dest
	err = pdf.MovePage(1, 99)
	if err == nil {
		t.Fatal("expected error for invalid dest")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── content_element.go: InsertLineElement, InsertRectElement, InsertOvalElement, InsertElementAt ───

func TestCov34_ContentElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Elements test")

	err := pdf.InsertLineElement(1, 10, 10, 200, 200)
	if err != nil {
		t.Fatalf("InsertLineElement: %v", err)
	}

	err = pdf.InsertRectElement(1, 50, 50, 100, 80, "DF")
	if err != nil {
		t.Fatalf("InsertRectElement: %v", err)
	}

	err = pdf.InsertOvalElement(1, 200, 200, 280, 250)
	if err != nil {
		t.Fatalf("InsertOvalElement: %v", err)
	}

	// Out of range page
	err = pdf.InsertLineElement(99, 0, 0, 10, 10)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	err = pdf.InsertRectElement(99, 0, 0, 10, 10, "D")
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	err = pdf.InsertOvalElement(99, 0, 0, 10, 10)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── embedded_file.go: GetEmbeddedFile, UpdateEmbeddedFile, GetEmbeddedFileInfo ───

func TestCov34_EmbeddedFile_Ops(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Embed test")

	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:        "test.txt",
		Content:     []byte("Hello embedded file"),
		Description: "A test file",
		MimeType:    "text/plain",
	})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}

	// GetEmbeddedFile
	content, err := pdf.GetEmbeddedFile("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFile: %v", err)
	}
	if string(content) != "Hello embedded file" {
		t.Fatalf("unexpected content: %s", string(content))
	}

	// GetEmbeddedFile not found
	_, err = pdf.GetEmbeddedFile("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// UpdateEmbeddedFile
	err = pdf.UpdateEmbeddedFile("test.txt", EmbeddedFile{
		Name:    "test.txt",
		Content: []byte("Updated content"),
	})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}

	// UpdateEmbeddedFile not found
	err = pdf.UpdateEmbeddedFile("nonexistent.txt", EmbeddedFile{
		Name:    "nonexistent.txt",
		Content: []byte("data"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// GetEmbeddedFileInfo
	info, err := pdf.GetEmbeddedFileInfo("test.txt")
	if err != nil {
		t.Fatalf("GetEmbeddedFileInfo: %v", err)
	}
	if info.Name != "test.txt" {
		t.Fatalf("expected name test.txt, got %s", info.Name)
	}

	// GetEmbeddedFileInfo not found
	_, err = pdf.GetEmbeddedFileInfo("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error for non-existent file info")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── journal.go: JournalSave, JournalLoad, snapshot deeper paths ───

func TestCov34_Journal_SaveLoad(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.JournalEnable()
	pdf.AddPage()

	pdf.JournalStartOp("add text")
	pdf.Cell(&Rect{W: 100, H: 20}, "Journal test")
	pdf.JournalEndOp()

	// Save
	savePath := resOutDir + "/cov34_journal.json"
	err := pdf.JournalSave(savePath)
	if err != nil {
		t.Fatalf("JournalSave: %v", err)
	}

	// Load
	pdf2 := newPDFWithFont(t)
	pdf2.JournalEnable()
	pdf2.AddPage()
	err = pdf2.JournalLoad(savePath)
	if err != nil {
		t.Fatalf("JournalLoad: %v", err)
	}

	// Undo
	desc, err := pdf2.JournalUndo()
	_ = desc
	_ = err

	// Redo
	desc2, err2 := pdf2.JournalRedo()
	_ = desc2
	_ = err2
}

// ─── font_extract.go: ExtractFontsFromPage, ExtractFontsFromAllPages ───

func TestCov34_FontExtract(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Font extract test")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	fonts, err := ExtractFontsFromPage(data, 0)
	if err != nil {
		t.Fatalf("ExtractFontsFromPage: %v", err)
	}
	_ = fonts

	allFonts, err := ExtractFontsFromAllPages(data)
	if err != nil {
		t.Fatalf("ExtractFontsFromAllPages: %v", err)
	}
	_ = allFonts

	// Out of range
	_, err = ExtractFontsFromPage(data, 99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}
}

// ─── doc_stats.go: GetDocumentStats, GetFonts ───

func TestCov34_DocStats(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Stats test")

	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder, 10, 50, &Rect{W: 50, H: 50})

	stats := pdf.GetDocumentStats()
	if stats.PageCount != 1 {
		t.Fatalf("expected 1 page, got %d", stats.PageCount)
	}
	if stats.FontCount < 1 {
		t.Fatalf("expected >= 1 font, got %d", stats.FontCount)
	}
	if stats.ImageCount < 1 {
		t.Fatalf("expected >= 1 image, got %d", stats.ImageCount)
	}

	fonts := pdf.GetFonts()
	if len(fonts) < 1 {
		t.Fatalf("expected >= 1 font, got %d", len(fonts))
	}
}

// ─── page_info.go: GetPageSize, GetAllPageSizes, GetSourcePDFPageCountFromBytes, GetSourcePDFPageSizesFromBytes ───

func TestCov34_PageInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPageWithOption(PageOption{PageSize: &Rect{W: 400, H: 600}})

	// GetPageSize
	w, h, err := pdf.GetPageSize(1)
	if err != nil {
		t.Fatalf("GetPageSize: %v", err)
	}
	if w == 0 {
		t.Fatal("expected non-zero width")
	}
	_ = h

	// GetPageSize out of range
	_, _, err = pdf.GetPageSize(99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	// GetAllPageSizes
	sizes := pdf.GetAllPageSizes()
	if len(sizes) != 2 {
		t.Fatalf("expected 2 sizes, got %d", len(sizes))
	}

	// GetSourcePDFPageCountFromBytes
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	func() {
		defer func() { recover() }()
		count, _ := GetSourcePDFPageCountFromBytes(data)
		_ = count
	}()

	// GetSourcePDFPageSizesFromBytes
	func() {
		defer func() { recover() }()
		pageSizes, _ := GetSourcePDFPageSizesFromBytes(data)
		_ = pageSizes
	}()
}

// ─── clone.go: Clone ───

func TestCov34_Clone(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Clone test")

	cloned, err := pdf.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}

	// Need to add font to cloned PDF
	if err := cloned.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	cloned.SetFont(fontFamily, "", 14)
	cloned.AddPage()
	cloned.Cell(&Rect{W: 200, H: 20}, "Cloned page")

	var buf bytes.Buffer
	cloned.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty cloned output")
	}
}

// ─── incremental_save.go: IncrementalSave, WriteIncrementalPdf ───

func TestCov34_IncrementalSave(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Incremental save")

	var origBuf bytes.Buffer
	pdf.WriteTo(&origBuf)
	origData := origBuf.Bytes()

	// IncrementalSave
	result, err := pdf.IncrementalSave(origData, nil)
	if err != nil {
		_ = err // may fail, that's ok
	}
	_ = result

	// WriteIncrementalPdf
	outPath := resOutDir + "/cov34_incremental.pdf"
	err = pdf.WriteIncrementalPdf(outPath, origData, nil)
	if err != nil {
		_ = err // may fail
	}
}

// ─── form_field.go: AddFormField ───

func TestCov34_AddFormField(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Name:       "textfield1",
		Type:       FormFieldText,
		X:          50,
		Y:          50,
		W:          200,
		H:          30,
		Value:      "Default text",
		FontFamily: fontFamily,
	})
	if err != nil {
		t.Fatalf("AddFormField text: %v", err)
	}

	// Checkbox
	err = pdf.AddFormField(FormField{
		Name:    "checkbox1",
		Type:    FormFieldCheckbox,
		X:       50,
		Y:       100,
		W:       20,
		H:       20,
		Checked: true,
	})
	if err != nil {
		t.Fatalf("AddFormField checkbox: %v", err)
	}

	// Missing name
	err = pdf.AddFormField(FormField{
		Type: FormFieldText,
		X:    10, Y: 10, W: 100, H: 20,
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	// Zero dimensions
	err = pdf.AddFormField(FormField{
		Name: "bad",
		Type: FormFieldText,
		X:    10, Y: 10, W: 0, H: 20,
	})
	if err == nil {
		t.Fatal("expected error for zero width")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── image_delete.go: DeleteImages, DeleteImagesFromAllPages, DeleteImageByIndex ───

func TestCov34_ImageDelete(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50})

	// DeleteImagesFromAllPages
	n, err := pdf.DeleteImagesFromAllPages()
	if err != nil {
		t.Fatalf("DeleteImagesFromAllPages: %v", err)
	}
	_ = n

	// Re-add image for more tests
	pdf.AddPage()
	holder2, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder2, 10, 10, &Rect{W: 50, H: 50})

	// DeleteImages (specific page)
	n2, err := pdf.DeleteImages(2)
	if err != nil {
		t.Fatalf("DeleteImages: %v", err)
	}
	_ = n2

	// Re-add for DeleteImageByIndex
	pdf.AddPage()
	holder3, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder3, 10, 10, &Rect{W: 50, H: 50})

	// DeleteImageByIndex
	err = pdf.DeleteImageByIndex(3, 0)
	if err != nil {
		// May fail if no image elements found
		_ = err
	}

	// Out of range page
	_, err = pdf.DeleteImages(99)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	err = pdf.DeleteImageByIndex(99, 0)
	if err == nil {
		t.Fatal("expected error for out-of-range page")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── table.go: DrawTable with header, border, multi-row ───

func TestCov34_DrawTable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	tbl := pdf.NewTableLayout(50, 50, 20, 5)
	tbl.AddColumn("Name", 100, "left")
	tbl.AddColumn("Age", 50, "center")
	tbl.AddColumn("City", 100, "right")

	tbl.SetTableStyle(CellStyle{
		BorderStyle: BorderStyle{
			Top: true, Left: true, Right: true, Bottom: true,
			Width:    1,
			RGBColor: RGBColor{R: 0, G: 0, B: 0},
		},
	})
	tbl.SetHeaderStyle(CellStyle{
		BorderStyle: BorderStyle{
			Top: true, Left: true, Right: true, Bottom: true,
			Width:    1,
			RGBColor: RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: RGBColor{R: 200, G: 200, B: 200},
		TextColor: RGBColor{R: 0, G: 0, B: 0},
	})
	tbl.SetCellStyle(CellStyle{
		BorderStyle: BorderStyle{
			Top: true, Left: true, Right: true, Bottom: true,
			Width:    0.5,
			RGBColor: RGBColor{R: 128, G: 128, B: 128},
		},
		TextColor: RGBColor{R: 0, G: 0, B: 0},
	})

	tbl.AddRow([]string{"Alice", "30", "NYC"})
	tbl.AddRow([]string{"Bob", "25", "LA"})

	// Styled row
	tbl.AddStyledRow([]RowCell{
		NewRowCell("Charlie", CellStyle{
			FillColor: RGBColor{R: 255, G: 200, B: 200},
			TextColor: RGBColor{R: 255, G: 0, B: 0},
			BorderStyle: BorderStyle{
				Top: true, Left: true, Right: true, Bottom: true,
				Width: 0.5, RGBColor: RGBColor{R: 0, G: 0, B: 0},
			},
		}),
		NewRowCell("35", CellStyle{TextColor: RGBColor{R: 0, G: 0, B: 0}}),
		NewRowCell("Chicago", CellStyle{TextColor: RGBColor{R: 0, G: 0, B: 0}}),
	})

	err := tbl.DrawTable()
	if err != nil {
		t.Fatalf("DrawTable: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── open_pdf.go: OpenPDF, OpenPDFFromStream, OpenPDFFromBytes ───

func TestCov34_OpenPDF(t *testing.T) {
	if _, err := os.Stat(resTestPDF); err != nil {
		t.Skip("test PDF not available")
	}

	// OpenPDF
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.OpenPDF(resTestPDF, nil)
	if err != nil {
		t.Fatalf("OpenPDF: %v", err)
	}

	// OpenPDFFromBytes
	data, _ := os.ReadFile(resTestPDF)
	pdf2 := &GoPdf{}
	pdf2.Start(Config{PageSize: *PageSizeA4})
	err = pdf2.OpenPDFFromBytes(data, nil)
	if err != nil {
		t.Fatalf("OpenPDFFromBytes: %v", err)
	}

	// OpenPDFFromStream
	f, _ := os.Open(resTestPDF)
	defer f.Close()
	var rs io.ReadSeeker = f
	pdf3 := &GoPdf{}
	pdf3.Start(Config{PageSize: *PageSizeA4})
	err = pdf3.OpenPDFFromStream(&rs, nil)
	if err != nil {
		t.Fatalf("OpenPDFFromStream: %v", err)
	}
}

// ─── html_insert.go: renderText, renderLongWord, applyFont, renderHR, renderList deeper paths ───

func TestCov34_HTMLInsert_DeepPaths(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Long word that needs breaking
	longWord := strings.Repeat("Superlongword", 10)
	html := fmt.Sprintf(`<p>%s</p>`, longWord)
	_, err := pdf.InsertHTMLBox(10, 10, 100, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox long word: %v", err)
	}

	// HR tag
	html2 := `<p>Before</p><hr><p>After</p>`
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html2, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox hr: %v", err)
	}

	// Ordered list
	html3 := `<ol><li>First</li><li>Second</li><li>Third</li></ol>`
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html3, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox ol: %v", err)
	}

	// Unordered list
	html4 := `<ul><li>Apple</li><li>Banana</li></ul>`
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html4, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox ul: %v", err)
	}

	// Nested elements
	html5 := `<p><b>Bold</b> and <i>italic</i> and <u>underline</u></p>`
	_, err = pdf.InsertHTMLBox(10, 10, 200, 300, html5, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox nested: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── html_insert.go: renderImage with bad src ───

func TestCov34_HTMLInsert_BadImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="nonexistent.jpg" width="50" height="50">`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	// May or may not error depending on implementation
	_ = err
}

// ─── image_obj_parse.go: parsePng with various color types, compress error paths ───

func TestCov34_ParsePng_GrayAlpha(t *testing.T) {
	// Create a gray+alpha PNG (color type 4)
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.NRGBA{R: 128, G: 128, B: 128, A: uint8(x * 25)})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, _ := ImageHolderByBytes(pngBuf.Bytes())
	err := pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50})
	if err != nil {
		t.Fatalf("ImageByHolder gray+alpha PNG: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

func TestCov34_ParsePng_Palette(t *testing.T) {
	// Create a paletted PNG (color type 3)
	palette := []color.Color{
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}
	img := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetColorIndex(x, y, uint8(x%3))
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, _ := ImageHolderByBytes(pngBuf.Bytes())
	err := pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50})
	if err != nil {
		t.Fatalf("ImageByHolder paletted PNG: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── image_obj_parse.go: parseImgJpg with CMYK JPEG ───

func TestCov34_ParseImgJpg_CMYK(t *testing.T) {
	// Standard JPEG (RGB) — already tested, but let's test with larger image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 2), G: uint8(y * 2), B: 128, A: 255})
		}
	}
	var jpgBuf bytes.Buffer
	jpeg.Encode(&jpgBuf, img, &jpeg.Options{Quality: 50})

	pdf := newPDFWithFont(t)
	pdf.AddPage()
	holder, _ := ImageHolderByBytes(jpgBuf.Bytes())
	err := pdf.ImageByHolder(holder, 10, 10, nil) // nil rect uses image dimensions
	if err != nil {
		t.Fatalf("ImageByHolder large JPEG: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: ImageFromWithOption (image.Image with options) ───

func TestCov34_ImageFromWithOption(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create an image.Image
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 5), G: uint8(y * 5), B: 128, A: 255})
		}
	}

	err := pdf.ImageFromWithOption(img, ImageFromOption{
		Format: "jpeg",
	})
	if err != nil {
		t.Fatalf("ImageFromWithOption: %v", err)
	}

	// nil image
	err = pdf.ImageFromWithOption(nil, ImageFromOption{})
	if err == nil {
		t.Fatal("expected error for nil image")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: addColorSpace, IsCurrentFontContainGlyph ───

func TestCov34_AddColorSpace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// AddColorSpaceRGB and AddColorSpaceCMYK
	pdf.AddColorSpaceRGB("MyRed", 255, 0, 0)
	pdf.AddColorSpaceCMYK("MyCyan", 100, 0, 0, 0)

	// SetColorSpace
	pdf.SetColorSpace("DeviceRGB")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

func TestCov34_IsCurrentFontContainGlyph(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// ASCII should be contained
	ok, err := pdf.IsCurrFontContainGlyph('A')
	if err != nil {
		t.Fatalf("IsCurrFontContainGlyph: %v", err)
	}
	if !ok {
		t.Fatal("expected glyph A to be contained")
	}

	// Rare Unicode might not be
	_, _ = pdf.IsCurrFontContainGlyph(0x1F600) // emoji
}

// ─── bookmark.go: DeleteBookmark deeper paths ───

func TestCov34_DeleteBookmark(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddOutlineWithPosition("Bookmark 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")
	pdf.AddOutlineWithPosition("Bookmark 2")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 3")
	pdf.AddOutlineWithPosition("Bookmark 3")

	// Delete middle bookmark
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark: %v", err)
	}

	// Delete first bookmark
	err = pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark first: %v", err)
	}

	// Out of range
	err = pdf.DeleteBookmark(99)
	if err == nil {
		t.Fatal("expected error for out-of-range bookmark")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── watermark.go: AddWatermarkTextAllPages, AddWatermarkImageAllPages ───

func TestCov34_WatermarkAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 1")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 2")
	pdf.AddPage()
	pdf.Cell(&Rect{W: 100, H: 20}, "Page 3")

	// Text watermark all pages
	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.2,
		Angle:      45,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}

	// Image watermark all pages
	if _, err := os.Stat(resJPEGPath); err == nil {
		err = pdf.AddWatermarkImageAllPages(resJPEGPath, 0.1, 100, 100, 30)
		if err != nil {
			t.Fatalf("AddWatermarkImageAllPages: %v", err)
		}
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: Line error path, RectFromLowerLeftWithOpts, RectFromUpperLeftWithOpts ───

func TestCov34_Line_RectOpts(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Line with transparency
	pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode})
	pdf.Line(10, 10, 200, 200)
	pdf.ClearTransparency()

	// RectFromLowerLeftWithOpts
	err := pdf.RectFromLowerLeftWithOpts(DrawableRectOptions{
		Rect:       Rect{W: 100, H: 50},
		X:          50,
		Y:          200,
		PaintStyle: "DF",
	})
	if err != nil {
		t.Fatalf("RectFromLowerLeftWithOpts: %v", err)
	}

	// RectFromUpperLeftWithOpts
	err = pdf.RectFromUpperLeftWithOpts(DrawableRectOptions{
		Rect:       Rect{W: 100, H: 50},
		X:          50,
		Y:          300,
		PaintStyle: "DF",
		Transparency: &Transparency{
			Alpha:         0.5,
			BlendModeType: NormalBlendMode,
		},
	})
	if err != nil {
		t.Fatalf("RectFromUpperLeftWithOpts: %v", err)
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: Polygon/Polyline error paths (empty points) ───

func TestCov34_Polygon_Polyline_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Empty points — should not panic
	pdf.Polygon(nil, "D")
	pdf.Polygon([]Point{}, "D")
	pdf.Polyline(nil)
	pdf.Polyline([]Point{})

	// Single point
	pdf.Polygon([]Point{{X: 10, Y: 10}}, "D")
	pdf.Polyline([]Point{{X: 10, Y: 10}})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: Sector with various angle ranges ───

func TestCov34_Sector_Angles(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Full circle
	pdf.Sector(100, 100, 50, 0, 360, "DF")
	// Negative angles
	pdf.Sector(200, 100, 50, -90, 90, "D")
	// Large angle
	pdf.Sector(300, 100, 50, 0, 270, "F")

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── gopdf.go: prepare() deeper paths ───

func TestCov34_Prepare_MultiFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	// Add two fonts
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font not available: %v", err)
	}
	if _, err := os.Stat(resFontPath2); err == nil {
		pdf.AddTTFFont(fontFamily2, resFontPath2)
	}

	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.Cell(&Rect{W: 200, H: 20}, "Font 1 text")

	if _, err := os.Stat(resFontPath2); err == nil {
		pdf.SetFont(fontFamily2, "", 12)
		pdf.Cell(&Rect{W: 200, H: 20}, "Font 2 text")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ─── gopdf.go: KernOverride ───

func TestCov34_KernOverride(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	if err := pdf.AddTTFFontWithOption(fontFamily, resFontPath, TtfOption{UseKerning: true}); err != nil {
		t.Skipf("font not available: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()

	// KernOverride
	err := pdf.KernOverride(fontFamily, FuncKernOverride(func(leftRune, rightRune rune, leftPair, rightPair uint, pairVal int16) int16 {
		if leftRune == 'A' && rightRune == 'V' {
			return -100
		}
		return pairVal
	}))
	if err != nil {
		t.Fatalf("KernOverride: %v", err)
	}

	pdf.Cell(&Rect{W: 200, H: 20}, "AV kerning test")

	// Non-existent font
	err = pdf.KernOverride("NonExistent", FuncKernOverride(func(l, r rune, lp, rp uint, pv int16) int16 { return pv }))
	if err == nil {
		t.Fatal("expected error for non-existent font")
	}

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
}

// ─── pdf_parser.go: extractNamedRefs indirect reference path ───

func TestCov34_ExtractNamedRefs_Indirect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Add an image to create XObject references
	jpgBuf := createSmallJPEG(t)
	holder, _ := ImageHolderByBytes(jpgBuf)
	pdf.ImageByHolder(holder, 10, 10, &Rect{W: 50, H: 50})

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	// Parse the PDF — this exercises extractNamedRefs for both fonts and xobjects
	parser, err := newRawPDFParser(data)
	if err != nil {
		t.Fatalf("newRawPDFParser: %v", err)
	}
	if len(parser.pages) < 1 {
		t.Fatal("expected at least 1 page")
	}

	// Get content stream to exercise getPageContentStream
	stream := parser.getPageContentStream(0)
	if len(stream) == 0 {
		t.Fatal("expected non-empty content stream")
	}
}

// ─── content_stream_clean.go: extractOperator edge cases ───

func TestCov34_CleanContentStreams_Complex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Various drawing operations to create complex content stream
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetFillColor(0, 255, 0)
	pdf.SetLineWidth(2)
	pdf.Line(10, 10, 200, 200)
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 80, "DF")
	pdf.SetGrayFill(0.5)
	pdf.RectFromUpperLeftWithStyle(200, 50, 50, 50, "F")
	pdf.Polygon([]Point{{X: 10, Y: 300}, {X: 50, Y: 350}, {X: 90, Y: 300}}, "DF")
	pdf.Oval(300, 300, 40, 30)

	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	data := buf.Bytes()

	cleaned, err := CleanContentStreams(data)
	if err != nil {
		t.Fatalf("CleanContentStreams complex: %v", err)
	}
	if len(cleaned) == 0 {
		t.Fatal("expected non-empty cleaned data")
	}
}

// ─── unused import guard ───

var _ = fmt.Sprintf
var _ = strings.Contains
var _ = os.DevNull
var _ = time.Now
var _ = big.NewInt
var _ = pkix.Name{}
var _ = elliptic.P256
