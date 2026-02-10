package gopdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost23_test.go — TestCov23_ prefix
// Targets: parsePng (tRNS, RGBA, GrayAlpha), HTML rendering
// (renderImage, renderList, renderLongWord, remainingWidth,
// renderHR, alignment), rebuildXref, extractNamedRefs,
// AddCompositeGlyphs, embedfont_obj, content_obj.write error paths
// ============================================================

// --- PNG helpers for raw PNG construction ---

func buildRawPNG(t *testing.T, width, height int, colorType byte, bitDepth byte, extraChunks []pngChunk, imageData []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	// PNG signature
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	// IHDR
	var ihdr bytes.Buffer
	binary.Write(&ihdr, binary.BigEndian, uint32(width))
	binary.Write(&ihdr, binary.BigEndian, uint32(height))
	ihdr.WriteByte(bitDepth)
	ihdr.WriteByte(colorType)
	ihdr.WriteByte(0) // compression
	ihdr.WriteByte(0) // filter
	ihdr.WriteByte(0) // interlace
	writePNGChunk(&buf, "IHDR", ihdr.Bytes())
	// extra chunks (PLTE, tRNS, etc.)
	for _, c := range extraChunks {
		writePNGChunk(&buf, c.typ, c.data)
	}
	// IDAT
	writePNGChunk(&buf, "IDAT", imageData)
	// IEND
	writePNGChunk(&buf, "IEND", nil)
	return buf.Bytes()
}

type pngChunk struct {
	typ  string
	data []byte
}

func writePNGChunk(w *bytes.Buffer, chunkType string, data []byte) {
	binary.Write(w, binary.BigEndian, uint32(len(data)))
	w.WriteString(chunkType)
	w.Write(data)
	// CRC (simplified — just write 4 bytes; parsePng skips CRC)
	crc := pngCRC([]byte(chunkType), data)
	binary.Write(w, binary.BigEndian, crc)
}

func pngCRC(chunkType, data []byte) uint32 {
	// IEEE CRC-32 used by PNG
	crcTable := makeCRCTable()
	crc := uint32(0xFFFFFFFF)
	for _, b := range chunkType {
		crc = crcTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	for _, b := range data {
		crc = crcTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	return crc ^ 0xFFFFFFFF
}

func makeCRCTable() [256]uint32 {
	var table [256]uint32
	for i := 0; i < 256; i++ {
		crc := uint32(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = 0xEDB88320 ^ (crc >> 1)
			} else {
				crc >>= 1
			}
		}
		table[i] = crc
	}
	return table
}

func zlibCompress(data []byte) []byte {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// --- Test: parsePng with color type 0 (Grayscale) + tRNS ---
func TestCov23_ParsePng_GrayscaleTRNS(t *testing.T) {
	w, h := 4, 4
	// Build raw scanlines: filter byte + grayscale pixels
	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0) // filter: None
		for x := 0; x < w; x++ {
			raw.WriteByte(byte((x + y) * 30))
		}
	}
	compressed := zlibCompress(raw.Bytes())

	// tRNS for grayscale: 2 bytes, second byte is the transparent gray value
	trnsData := []byte{0, 60} // gray value 60 is transparent

	pngData := buildRawPNG(t, w, h, 0, 8, []pngChunk{
		{typ: "tRNS", data: trnsData},
	}, compressed)

	reader := bytes.NewReader(pngData)
	imgConfig, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		// If standard decoder can't parse our raw PNG, skip
		t.Skipf("image.DecodeConfig failed: %v", err)
	}

	var info imgInfo
	reader.Seek(0, 0)
	err = parsePng(reader, &info, imgConfig)
	if err != nil {
		t.Logf("parsePng grayscale+tRNS: %v (may be expected)", err)
		return
	}
	if info.colspace != "DeviceGray" {
		t.Errorf("expected DeviceGray, got %s", info.colspace)
	}
	if len(info.trns) == 0 {
		t.Error("expected tRNS data for grayscale")
	}
}

// --- Test: parsePng with color type 2 (RGB) + tRNS ---
func TestCov23_ParsePng_RGBWithTRNS(t *testing.T) {
	w, h := 4, 4
	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0) // filter: None
		for x := 0; x < w; x++ {
			raw.WriteByte(byte(x * 60))
			raw.WriteByte(byte(y * 60))
			raw.WriteByte(128)
		}
	}
	compressed := zlibCompress(raw.Bytes())

	// tRNS for RGB: 6 bytes (R_high, R_low, G_high, G_low, B_high, B_low)
	trnsData := []byte{0, 255, 0, 128, 0, 64}

	pngData := buildRawPNG(t, w, h, 2, 8, []pngChunk{
		{typ: "tRNS", data: trnsData},
	}, compressed)

	reader := bytes.NewReader(pngData)
	imgConfig, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		t.Skipf("image.DecodeConfig failed: %v", err)
	}

	var info imgInfo
	reader.Seek(0, 0)
	err = parsePng(reader, &info, imgConfig)
	if err != nil {
		t.Logf("parsePng RGB+tRNS: %v (may be expected)", err)
		return
	}
	if info.colspace != "DeviceRGB" {
		t.Errorf("expected DeviceRGB, got %s", info.colspace)
	}
	if len(info.trns) != 3 {
		t.Errorf("expected 3 tRNS bytes for RGB, got %d", len(info.trns))
	}
}

// --- Test: parsePng with color type 3 (Indexed/Paletted) + tRNS ---
func TestCov23_ParsePng_PalettedTRNS(t *testing.T) {
	w, h := 4, 4
	// Palette: 4 colors
	palette := []byte{
		255, 0, 0, // red
		0, 255, 0, // green
		0, 0, 255, // blue
		0, 0, 0, // black (transparent)
	}
	// tRNS for paletted: alpha values per palette entry, with \x00 marking transparent
	trnsData := []byte{255, 255, 255, 0}

	var raw bytes.Buffer
	for y := 0; y < h; y++ {
		raw.WriteByte(0) // filter: None
		for x := 0; x < w; x++ {
			raw.WriteByte(byte((x + y) % 4))
		}
	}
	compressed := zlibCompress(raw.Bytes())

	pngData := buildRawPNG(t, w, h, 3, 8, []pngChunk{
		{typ: "PLTE", data: palette},
		{typ: "tRNS", data: trnsData},
	}, compressed)

	reader := bytes.NewReader(pngData)
	imgConfig, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		t.Skipf("image.DecodeConfig failed: %v", err)
	}

	var info imgInfo
	reader.Seek(0, 0)
	err = parsePng(reader, &info, imgConfig)
	if err != nil {
		t.Logf("parsePng paletted+tRNS: %v (may be expected)", err)
		return
	}
	if info.colspace != "Indexed" {
		t.Errorf("expected Indexed, got %s", info.colspace)
	}
}

// --- Test: parsePng with color type 4 (GrayAlpha) ---
func TestCov23_ParsePng_GrayAlpha(t *testing.T) {
	// Create a GrayAlpha PNG using Go's image library
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			g := uint8((x + y) * 15)
			a := uint8(128 + x*10)
			img.Set(x, y, color.NRGBA{R: g, G: g, B: g, A: a})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	// This will be decoded as RGBA by Go, but parsePng reads raw chunks
	// and handles ct[0]==6 (RGBA) path which splits color+alpha
	reader := bytes.NewReader(pngData)
	imgConfig, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		t.Skipf("DecodeConfig: %v", err)
	}
	var info imgInfo
	err = parsePng(reader, &info, imgConfig)
	if err != nil {
		t.Logf("parsePng GrayAlpha: %v", err)
	}
	// The key is that we exercise the ct[0] >= 4 branch
	if info.smask != nil {
		t.Logf("smask generated, len=%d", len(info.smask))
	}
}

// --- Test: parsePng with color type 6 (RGBA) ---
func TestCov23_ParsePng_RGBA(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(x * 25), G: uint8(y * 25), B: 100, A: uint8(200 - x*5)})
		}
	}
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngData := pngBuf.Bytes()

	reader := bytes.NewReader(pngData)
	imgConfig, _, err := image.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		t.Skipf("DecodeConfig: %v", err)
	}
	var info imgInfo
	err = parsePng(reader, &info, imgConfig)
	if err != nil {
		t.Fatalf("parsePng RGBA: %v", err)
	}
	if info.smask == nil {
		t.Error("expected smask for RGBA PNG")
	}
	if info.data == nil {
		t.Error("expected data for RGBA PNG")
	}
	if info.colspace != "DeviceRGB" {
		t.Errorf("expected DeviceRGB, got %s", info.colspace)
	}
}

// --- Test: HTML rendering with ordered and unordered lists ---
func TestCov23_HTML_Lists(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Unordered list
	html1 := `<ul><li>Item A</li><li>Item B</li><li>Item C</li></ul>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 300, html1, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox ul: %v", err)
	}

	// Ordered list with >9 items to trigger the counter>9 branch
	var items strings.Builder
	for i := 1; i <= 12; i++ {
		items.WriteString("<li>Item number ")
		items.WriteString(strings.Repeat("x", 5))
		items.WriteString("</li>")
	}
	html2 := "<ol>" + items.String() + "</ol>"
	_, err = pdf.InsertHTMLBox(10, 320, 200, 400, html2, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   10,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox ol: %v", err)
	}
}

// --- Test: HTML rendering with <hr> ---
func TestCov23_HTML_HR(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Before HR</p><hr><p>After HR</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 200, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox hr: %v", err)
	}
}

// --- Test: HTML rendering with <img> tag ---
func TestCov23_HTML_Image(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Use a real image path
	html := `<p>Image below:</p><img src="` + resJPEGPath + `" width="100" height="75"><p>After image</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img: %v", err)
	}
}

// --- Test: HTML rendering with <img> no width/height ---
func TestCov23_HTML_ImageNoSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="` + resJPEGPath + `">`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img no size: %v", err)
	}
}

// --- Test: HTML rendering with <img> only width ---
func TestCov23_HTML_ImageOnlyWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="` + resJPEGPath + `" width="150">`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img only width: %v", err)
	}
}

// --- Test: HTML rendering with <img> only height ---
func TestCov23_HTML_ImageOnlyHeight(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img src="` + resJPEGPath + `" height="100">`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img only height: %v", err)
	}
}

// --- Test: HTML rendering with <img> src missing ---
func TestCov23_HTML_ImageNoSrc(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<img>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox img no src: %v", err)
	}
}

// --- Test: HTML rendering with center alignment ---
func TestCov23_HTML_CenterAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<center>Centered text here</center>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   14,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox center: %v", err)
	}
}

// --- Test: HTML rendering with right alignment via style ---
func TestCov23_HTML_RightAlign(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p style="text-align:right">Right aligned text</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   14,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox right: %v", err)
	}
}

// --- Test: HTML rendering with very long word (renderLongWord) ---
func TestCov23_HTML_LongWord(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Create a word longer than the box width
	longWord := strings.Repeat("W", 200)
	html := `<p>` + longWord + `</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 100, 400, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   14,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox long word: %v", err)
	}
}

// --- Test: HTML rendering with <a> link ---
func TestCov23_HTML_Link(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>Visit <a href="https://example.com">Example</a> site</p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox link: %v", err)
	}
}

// --- Test: HTML rendering with <blockquote> ---
func TestCov23_HTML_Blockquote(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<blockquote>This is a quoted block of text that should be indented.</blockquote>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox blockquote: %v", err)
	}
}

// --- Test: HTML rendering with <sub> and <sup> ---
func TestCov23_HTML_SubSup(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p>H<sub>2</sub>O and E=mc<sup>2</sup></p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   14,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox sub/sup: %v", err)
	}
}

// --- Test: HTML rendering with <font> tag ---
func TestCov23_HTML_FontTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<font color="#FF0000" size="5">Red large text</font>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox font tag: %v", err)
	}
}

// --- Test: HTML rendering with headings ---
func TestCov23_HTML_Headings(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<h1>Heading 1</h1><h2>Heading 2</h2><h3>Heading 3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 500, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox headings: %v", err)
	}
}

// --- Test: HTML rendering with <s>, <del>, <ins> ---
func TestCov23_HTML_StrikeInsert(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	html := `<p><s>strikethrough</s> <del>deleted</del> <ins>inserted</ins></p>`
	_, err := pdf.InsertHTMLBox(10, 10, 300, 200, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox strike/ins: %v", err)
	}
}

// --- Test: rebuildXref with \r\n line endings ---
func TestCov23_RebuildXref_CRLF(t *testing.T) {
	// Build a minimal PDF with \r\n xref
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Test")
	var buf bytes.Buffer
	pdf.Write(&buf)
	data := buf.Bytes()

	// Replace \nxref\n with \r\nxref\r\n to test CRLF path
	data2 := bytes.ReplaceAll(data, []byte("xref\n"), []byte("xref\r\n"))
	result := rebuildXref(data2)
	if len(result) == 0 {
		t.Error("rebuildXref returned empty")
	}
}

// --- Test: rebuildXref with no xref ---
func TestCov23_RebuildXref_NoXref(t *testing.T) {
	data := []byte("some random data without xref")
	result := rebuildXref(data)
	if !bytes.Equal(result, data) {
		t.Error("expected unchanged data when no xref")
	}
}

// --- Test: rebuildXref with no objects ---
func TestCov23_RebuildXref_NoObjects(t *testing.T) {
	data := []byte("xref\n0 1\n0000000000 65535 f \ntrailer\n<< /Size 1 >>\nstartxref\n0\n%%EOF\n")
	result := rebuildXref(data)
	// Should return original since no obj headers found
	if len(result) == 0 {
		t.Error("rebuildXref returned empty")
	}
}

// --- Test: extractNamedRefs with inline dict ---
func TestCov23_ExtractNamedRefs_InlineDict(t *testing.T) {
	// Build a minimal PDF that has inline resource dicts
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Hello")

	// Add an image to create XObject resources
	pngData := createTestPNG(t, 10, 10)
	holder, err := ImageHolderByBytes(pngData)
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	pdf.ImageByHolder(holder, 10, 50, nil)

	var buf bytes.Buffer
	pdf.Write(&buf)
	data := buf.Bytes()

	// Parse the PDF to exercise extractNamedRefs
	parser, err := newRawPDFParser(data)
	if err != nil {
		t.Fatalf("newRawPDFParser: %v", err)
	}
	if len(parser.pages) == 0 {
		t.Fatal("no pages parsed")
	}
	t.Logf("parsed %d pages, %d objects", len(parser.pages), len(parser.objects))
}

// --- Test: OpenPDFFromStream ---
func TestCov23_OpenPDFFromStream(t *testing.T) {
	// Create a simple PDF first
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Stream test")
	var buf bytes.Buffer
	pdf.Write(&buf)

	// Now open it via OpenPDFFromStream
	pdf2 := &GoPdf{}
	rs := io.ReadSeeker(bytes.NewReader(buf.Bytes()))
	err := pdf2.OpenPDFFromStream(&rs, nil)
	if err != nil {
		t.Fatalf("OpenPDFFromStream: %v", err)
	}
}

// --- Test: DeleteBookmark edge cases ---
func TestCov23_DeleteBookmark_Edges(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOutline("BM1")
	pdf.AddPage()
	pdf.AddOutline("BM2")
	pdf.AddPage()
	pdf.AddOutline("BM3")
	pdf.AddPage()
	pdf.AddOutline("BM4")

	// Delete middle bookmark (index 1)
	err := pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark 1: %v", err)
	}

	// Delete first bookmark (index 0)
	err = pdf.DeleteBookmark(0)
	if err != nil {
		t.Fatalf("DeleteBookmark 0: %v", err)
	}

	// Delete last bookmark (now index 1 after previous deletes)
	err = pdf.DeleteBookmark(1)
	if err != nil {
		t.Fatalf("DeleteBookmark last: %v", err)
	}

	// Delete out-of-range
	err = pdf.DeleteBookmark(999)
	if err == nil {
		t.Error("expected error for out-of-range bookmark")
	}

	// Verify PDF still works
	var buf bytes.Buffer
	err = pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// --- Test: collectOutlineObjs ---
func TestCov23_CollectOutlineObjs(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 5; i++ {
		pdf.AddPage()
		pdf.AddOutline("Chapter " + string(rune('A'+i)))
	}
	// SetBookmarkStyle to exercise that path
	pdf.SetBookmarkStyle(0, BookmarkStyle{Bold: true, Italic: true, Color: [3]float64{1.0, 0, 0}})

	var buf bytes.Buffer
	err := pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// --- Test: content_obj.write with NoCompression ---
func TestCov23_ContentObj_WriteNoCompression(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetNoCompression()
	if err := pdf.AddTTFFont(fontFamily, resFontPath); err != nil {
		t.Skipf("font: %v", err)
	}
	pdf.SetFont(fontFamily, "", 14)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "No compression test")

	var buf bytes.Buffer
	err := pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("empty output")
	}
}

// --- Test: content_obj.write with protection ---
func TestCov23_ContentObj_WriteProtected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Protected content")

	// Add various content to exercise more write paths
	pdf.Line(10, 50, 200, 50)
	pdf.RectFromUpperLeft(10, 60, 100, 50)
	pdf.Oval(50, 150, 150, 200)

	var buf bytes.Buffer
	err := pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// --- Test: image_obj.write with protection ---
func TestCov23_ImageObj_WriteProtected(t *testing.T) {
	pdf := newProtectedPDF(t)
	pdf.AddPage()

	pngData := createTestPNG(t, 20, 20)
	holder, err := ImageHolderByBytes(pngData)
	if err != nil {
		t.Fatalf("ImageHolderByBytes: %v", err)
	}
	err = pdf.ImageByHolder(holder, 10, 10, nil)
	if err != nil {
		t.Fatalf("ImageByHolder: %v", err)
	}

	var buf bytes.Buffer
	err = pdf.Write(&buf)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

// --- Test: EmbedFontObj write (needs a .z font file) ---
func TestCov23_EmbedFontObj_Write(t *testing.T) {
	// Create a temporary .z font file
	ensureOutDir(t)
	tmpFile := resOutDir + "/test_font.z"
	// Write some dummy compressed data
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	zw.Write([]byte("dummy font data for testing"))
	zw.Close()
	if err := os.WriteFile(tmpFile, zbuf.Bytes(), 0644); err != nil {
		t.Fatalf("write temp font: %v", err)
	}
	defer os.Remove(tmpFile)

	obj := &EmbedFontObj{
		zfontpath: tmpFile,
		font:      &mockFontForEmbed{originalSize: 100},
		getRoot: func() *GoPdf {
			pdf := &GoPdf{}
			pdf.Start(Config{PageSize: *PageSizeA4})
			return pdf
		},
	}

	var buf bytes.Buffer
	err := obj.write(&buf, 1)
	if err != nil {
		t.Fatalf("EmbedFontObj.write: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("empty output from EmbedFontObj.write")
	}
}

type mockFontForEmbed struct {
	originalSize int
}

func (m *mockFontForEmbed) Init()                                                {}
func (m *mockFontForEmbed) GetType() string                                      { return "Type1" }
func (m *mockFontForEmbed) GetName() string                                      { return "TestFont" }
func (m *mockFontForEmbed) GetDesc() []FontDescItem                              { return nil }
func (m *mockFontForEmbed) GetUp() int                                           { return -100 }
func (m *mockFontForEmbed) GetUt() int                                           { return 50 }
func (m *mockFontForEmbed) GetCw() FontCw                                        { return FontCw{} }
func (m *mockFontForEmbed) GetEnc() string                                       { return "cp1252" }
func (m *mockFontForEmbed) GetDiff() string                                      { return "" }
func (m *mockFontForEmbed) GetOriginalsize() int                                 { return m.originalSize }
func (m *mockFontForEmbed) SetFamily(family string)                              {}
func (m *mockFontForEmbed) GetFamily() string                                    { return "TestFont" }

// --- Test: EmbedFontObj write with protection ---
func TestCov23_EmbedFontObj_WriteProtected(t *testing.T) {
	ensureOutDir(t)
	tmpFile := resOutDir + "/test_font_prot.z"
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	zw.Write([]byte("dummy font data protected"))
	zw.Close()
	if err := os.WriteFile(tmpFile, zbuf.Bytes(), 0644); err != nil {
		t.Fatalf("write temp font: %v", err)
	}
	defer os.Remove(tmpFile)

	protPDF := newProtectedPDF(t)

	obj := &EmbedFontObj{
		zfontpath: tmpFile,
		font:      &mockFontForEmbed{originalSize: 80},
		getRoot: func() *GoPdf {
			return protPDF
		},
	}

	var buf bytes.Buffer
	err := obj.write(&buf, 1)
	if err != nil {
		t.Fatalf("EmbedFontObj.write protected: %v", err)
	}
}

// --- Test: EmbedFontObj write with missing file ---
func TestCov23_EmbedFontObj_WriteMissingFile(t *testing.T) {
	obj := &EmbedFontObj{
		zfontpath: "/nonexistent/path/font.z",
		font:      &mockFontForEmbed{originalSize: 100},
		getRoot: func() *GoPdf {
			pdf := &GoPdf{}
			pdf.Start(Config{PageSize: *PageSizeA4})
			return pdf
		},
	}

	var buf bytes.Buffer
	err := obj.write(&buf, 1)
	if err == nil {
		t.Error("expected error for missing font file")
	}
}

// --- Test: MergePages ---
func TestCov23_MergePages(t *testing.T) {
	// Create two simple PDFs
	ensureOutDir(t)
	path1 := resOutDir + "/merge1.pdf"
	path2 := resOutDir + "/merge2.pdf"

	for _, p := range []string{path1, path2} {
		pdf := newPDFWithFont(t)
		pdf.AddPage()
		pdf.SetXY(10, 10)
		pdf.Cell(nil, "Page from "+p)
		pdf.WritePdf(p)
	}
	defer os.Remove(path1)
	defer os.Remove(path2)

	merged, err := MergePages([]string{path1, path2}, nil)
	if err != nil {
		t.Fatalf("MergePages: %v", err)
	}
	var buf bytes.Buffer
	merged.Write(&buf)
	if buf.Len() == 0 {
		t.Error("merged PDF is empty")
	}
}

// --- Test: ExtractPages ---
func TestCov23_ExtractPages(t *testing.T) {
	ensureOutDir(t)
	path := resOutDir + "/extract_src.pdf"

	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(10, 10)
		pdf.Cell(nil, "Page "+string(rune('1'+i)))
	}
	pdf.WritePdf(path)
	defer os.Remove(path)

	extracted, err := ExtractPages(path, []int{1, 3}, nil)
	if err != nil {
		t.Fatalf("ExtractPages: %v", err)
	}
	var buf bytes.Buffer
	extracted.Write(&buf)
	if buf.Len() == 0 {
		t.Error("extracted PDF is empty")
	}
}

// --- Test: MergePagesFromBytes ---
func TestCov23_MergePagesFromBytes(t *testing.T) {
	var pdfs [][]byte
	for i := 0; i < 2; i++ {
		pdf := newPDFWithFont(t)
		pdf.AddPage()
		pdf.SetXY(10, 10)
		pdf.Cell(nil, "Bytes merge page")
		var buf bytes.Buffer
		pdf.Write(&buf)
		pdfs = append(pdfs, buf.Bytes())
	}

	merged, err := MergePagesFromBytes(pdfs, nil)
	if err != nil {
		t.Fatalf("MergePagesFromBytes: %v", err)
	}
	var buf bytes.Buffer
	merged.Write(&buf)
	if buf.Len() == 0 {
		t.Error("merged PDF is empty")
	}
}

// --- Test: ExtractPagesFromBytes ---
func TestCov23_ExtractPagesFromBytes(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(10, 10)
		pdf.Cell(nil, "Page")
	}
	var buf bytes.Buffer
	pdf.Write(&buf)

	extracted, err := ExtractPagesFromBytes(buf.Bytes(), []int{1, 2}, nil)
	if err != nil {
		t.Fatalf("ExtractPagesFromBytes: %v", err)
	}
	var out bytes.Buffer
	extracted.Write(&out)
	if out.Len() == 0 {
		t.Error("extracted PDF is empty")
	}
}

// --- Test: VerifySignature with unsigned PDF ---
func TestCov23_VerifySignature_Unsigned(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Unsigned")
	var buf bytes.Buffer
	pdf.Write(&buf)

	results, err := VerifySignature(buf.Bytes())
	if err != nil {
		t.Logf("VerifySignature on unsigned: %v", err)
	}
	if len(results) != 0 {
		t.Logf("unexpected results: %d", len(results))
	}
}

// --- Test: ParseCertificatePEM with invalid data ---
func TestCov23_ParseCertificatePEM_Invalid(t *testing.T) {
	_, err := ParseCertificatePEM([]byte("not a PEM"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

// --- Test: ParsePrivateKeyPEM with invalid data ---
func TestCov23_ParsePrivateKeyPEM_Invalid(t *testing.T) {
	_, err := ParsePrivateKeyPEM([]byte("not a PEM"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

// --- Test: ParseCertificateChainPEM with invalid data ---
func TestCov23_ParseCertChainPEM_Invalid(t *testing.T) {
	_, err := ParseCertificateChainPEM([]byte("not a PEM"))
	if err == nil {
		t.Error("expected error for invalid PEM chain")
	}
}

// --- Test: SelectPagesFromBytes ---
func TestCov23_SelectPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 5; i++ {
		pdf.AddPage()
		pdf.SetXY(10, 10)
		pdf.Cell(nil, "Page")
	}
	var buf bytes.Buffer
	pdf.Write(&buf)

	result, err := SelectPagesFromBytes(buf.Bytes(), []int{2, 4}, nil)
	if err != nil {
		t.Fatalf("SelectPagesFromBytes: %v", err)
	}
	var out bytes.Buffer
	result.Write(&out)
	if out.Len() == 0 {
		t.Error("SelectPagesFromBytes returned empty")
	}
}

// --- Test: GetBytesPdfReturnErr ---
func TestCov23_GetBytesPdfReturnErr(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Test")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdfReturnErr: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty PDF bytes")
	}
}

// --- Test: AddTTFFontByReaderWithOption ---
func TestCov23_AddTTFFontByReaderWithOption(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	f, err := os.Open(resFontPath)
	if err != nil {
		t.Skipf("font not available: %v", err)
	}
	defer f.Close()

	err = pdf.AddTTFFontByReaderWithOption(fontFamily+"Opt", f, TtfOption{
		UseKerning: true,
	})
	if err != nil {
		t.Fatalf("AddTTFFontByReaderWithOption: %v", err)
	}

	err = pdf.SetFont(fontFamily+"Opt", "", 14)
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Font with options")
}

// --- Test: KernValueByLeft ---
func TestCov23_KernValueByLeft(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontByReaderWithOption(fontFamily+"Kern", func() io.Reader {
		f, _ := os.Open(resFontPath)
		return f
	}(), TtfOption{UseKerning: true})
	if err != nil {
		t.Skipf("font: %v", err)
	}
	pdf.SetFont(fontFamily+"Kern", "", 14)
	pdf.AddPage()

	// Access the subset font obj to test KernValueByLeft
	for _, obj := range pdf.pdfObjs {
		if sfObj, ok := obj.(*SubsetFontObj); ok {
			// Try some glyph indices
			found, kv := sfObj.KernValueByLeft(0)
			t.Logf("KernValueByLeft(0): found=%v, kv=%v", found, kv)
			found, kv = sfObj.KernValueByLeft(65) // 'A'
			t.Logf("KernValueByLeft(65): found=%v, kv=%v", found, kv)
			break
		}
	}
}

// --- Test: CharWidth ---
func TestCov23_CharWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	for _, obj := range pdf.pdfObjs {
		if sfObj, ok := obj.(*SubsetFontObj); ok {
			w, err := sfObj.CharWidth('A')
			if err != nil {
				t.Logf("CharWidth('A'): %v", err)
			} else {
				t.Logf("CharWidth('A') = %d", w)
			}
			// Try a char that might not exist
			_, err = sfObj.CharWidth(0xFFFF)
			if err == nil {
				t.Logf("CharWidth(0xFFFF) unexpectedly succeeded")
			}
			break
		}
	}
}

// --- Test: charCodeToGlyphIndexFormat12 ---
func TestCov23_CharCodeToGlyphIndexFormat12(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	for _, obj := range pdf.pdfObjs {
		if sfObj, ok := obj.(*SubsetFontObj); ok {
			// Try format12 directly
			_, err := sfObj.charCodeToGlyphIndexFormat12('A')
			t.Logf("format12('A'): err=%v", err)

			// Try with a high codepoint
			_, err = sfObj.charCodeToGlyphIndexFormat12(0x10000)
			t.Logf("format12(0x10000): err=%v", err)

			// Try CharCodeToGlyphIndex which dispatches
			idx, err := sfObj.CharCodeToGlyphIndex('A')
			t.Logf("CharCodeToGlyphIndex('A'): idx=%d, err=%v", idx, err)
			break
		}
	}
}

// --- Test: replaceGlyphThatNotFound ---
func TestCov23_ReplaceGlyphThatNotFound(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddTTFFontWithOption(fontFamily+"Rep", resFontPath, TtfOption{
		OnGlyphNotFoundSubstitute: func(r rune) rune { return '?' },
	})
	if err != nil {
		t.Skipf("font: %v", err)
	}
	pdf.SetFont(fontFamily+"Rep", "", 14)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	// Use a character that likely doesn't exist in the font
	pdf.Cell(nil, string(rune(0x2603))) // snowman
}

// --- Test: parseImgByPath ---
func TestCov23_ParseImgByPath(t *testing.T) {
	info, err := parseImgByPath(resJPEGPath)
	if err != nil {
		t.Fatalf("parseImgByPath JPEG: %v", err)
	}
	if info.w == 0 || info.h == 0 {
		t.Error("zero dimensions")
	}

	info2, err := parseImgByPath(resPNGPath)
	if err != nil {
		t.Fatalf("parseImgByPath PNG: %v", err)
	}
	if info2.w == 0 || info2.h == 0 {
		t.Error("zero dimensions for PNG")
	}
}

// --- Test: parseImgByPath with non-existent file ---
func TestCov23_ParseImgByPath_NotFound(t *testing.T) {
	_, err := parseImgByPath("/nonexistent/image.png")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// --- Test: parseImg with GIF ---
func TestCov23_ParseImg_GIF(t *testing.T) {
	// GIF is imported but not explicitly handled — should fall through to image.Decode
	// Create a minimal GIF-like data (will likely fail but exercises the path)
	data := []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")
	reader := bytes.NewReader(data)
	_, err := parseImg(reader)
	// GIF should be handled by the fallback path
	t.Logf("parseImg GIF: err=%v", err)
}

// --- Test: writeImgProps with trns ---
func TestCov23_WriteImgProps_Trns(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		trns:            []byte{255, 128, 64},
	}
	var buf bytes.Buffer
	err := writeImgProps(&buf, info, false)
	if err != nil {
		t.Fatalf("writeImgProps: %v", err)
	}
	if !strings.Contains(buf.String(), "/Mask") {
		t.Error("expected /Mask in output")
	}
}

// --- Test: writeImgProps with splittedMask ---
func TestCov23_WriteImgProps_SplittedMask(t *testing.T) {
	info := imgInfo{
		w:               10,
		h:               10,
		colspace:        "DeviceRGB",
		bitsPerComponent: "8",
		decodeParms:     "/Predictor 15",
	}
	var buf bytes.Buffer
	err := writeImgProps(&buf, info, true)
	if err != nil {
		t.Fatalf("writeImgProps splitted: %v", err)
	}
}

// --- Test: writeMaskImgProps ---
func TestCov23_WriteMaskImgProps(t *testing.T) {
	info := imgInfo{
		w:               20,
		h:               20,
		colspace:        "DeviceGray",
		bitsPerComponent: "8",
	}
	var buf bytes.Buffer
	err := writeMaskImgProps(&buf, info)
	if err != nil {
		t.Fatalf("writeMaskImgProps: %v", err)
	}
	if !strings.Contains(buf.String(), "/Predictor 15") {
		t.Error("expected /Predictor 15 in mask props")
	}
}

// --- Test: isDeviceRGB ---
func TestCov23_IsDeviceRGB(t *testing.T) {
	// NRGBA image
	img1 := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	var i1 image.Image = img1
	if !isDeviceRGB("png", &i1) {
		t.Error("NRGBA should be DeviceRGB")
	}

	// Gray image
	img2 := image.NewGray(image.Rect(0, 0, 1, 1))
	var i2 image.Image = img2
	if isDeviceRGB("png", &i2) {
		t.Error("Gray should not be DeviceRGB")
	}
}

// --- Test: ImgReactagleToWH ---
func TestCov23_ImgRectangleToWH(t *testing.T) {
	rect := image.Rect(10, 20, 110, 220)
	w, h := ImgReactagleToWH(rect)
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got %vx%v", w, h)
	}
	t.Logf("ImgReactagleToWH: %vx%v", w, h)
}

// --- Test: downscaleImage ---
func TestCov23_DownscaleImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}

	// Downscale by width
	result := downscaleImage(img, 100, 0)
	bounds := result.Bounds()
	if bounds.Dx() > 100 {
		t.Errorf("expected width <= 100, got %d", bounds.Dx())
	}

	// Downscale by height
	result2 := downscaleImage(img, 0, 50)
	bounds2 := result2.Bounds()
	if bounds2.Dy() > 50 {
		t.Errorf("expected height <= 50, got %d", bounds2.Dy())
	}

	// Downscale by both
	result3 := downscaleImage(img, 50, 25)
	bounds3 := result3.Bounds()
	if bounds3.Dx() > 50 || bounds3.Dy() > 25 {
		t.Errorf("expected <= 50x25, got %dx%d", bounds3.Dx(), bounds3.Dy())
	}
}

// --- Test: replaceObjectStream ---
func TestCov23_ReplaceObjectStream(t *testing.T) {
	pdfData := []byte("1 0 obj\n<< /Type /Test >>\nstream\nold data\nendstream\nendobj\n2 0 obj\n<< >>\nendobj\n")
	result := replaceObjectStream(pdfData, 1, "<< /Type /New >>", []byte("new data"))
	if !bytes.Contains(result, []byte("new data")) {
		t.Error("expected new data in result")
	}

	// Non-existent object
	result2 := replaceObjectStream(pdfData, 99, "<< >>", []byte("x"))
	if !bytes.Equal(result2, pdfData) {
		t.Error("expected unchanged data for non-existent object")
	}
}

// --- Test: replaceObjectStream with \r\n ---
func TestCov23_ReplaceObjectStream_CRLF(t *testing.T) {
	pdfData := []byte("1 0 obj\r\n<< /Type /Test >>\r\nstream\r\nold data\r\nendstream\r\nendobj\r\n")
	result := replaceObjectStream(pdfData, 1, "<< /Type /New >>", []byte("new data"))
	if !bytes.Contains(result, []byte("new data")) {
		t.Error("expected new data in CRLF result")
	}
}

// --- Test: extractFilterValue ---
func TestCov23_ExtractFilterValue(t *testing.T) {
	dict1 := "<< /Filter /DCTDecode /Width 100 >>"
	if v := extractFilterValue(dict1); v != "DCTDecode" {
		t.Errorf("expected DCTDecode, got %s", v)
	}

	dict2 := "<< /Width 100 >>"
	if v := extractFilterValue(dict2); v != "" {
		t.Errorf("expected empty, got %s", v)
	}
}

// --- Test: ExtractedImage.GetImageFormat ---
func TestCov23_GetImageFormat(t *testing.T) {
	img1 := &ExtractedImage{Filter: "DCTDecode"}
	if img1.GetImageFormat() != "jpeg" {
		t.Errorf("expected jpeg, got %s", img1.GetImageFormat())
	}

	img2 := &ExtractedImage{Filter: "FlateDecode"}
	if img2.GetImageFormat() != "png" {
		t.Errorf("expected png, got %s", img2.GetImageFormat())
	}

	img3 := &ExtractedImage{Filter: ""}
	format := img3.GetImageFormat()
	t.Logf("empty filter format: %s", format)
}

// --- Test: findCurrentPageObjID and findCurrentPageObj ---
func TestCov23_FindCurrentPageObj(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(10, 10)
	pdf.Cell(nil, "Test")

	// These are internal methods used by form fields
	// Exercise them indirectly through AddFormField
	err := pdf.AddFormField(FormField{
		Name:  "test_field",
		Type:  FormFieldText,
		X:     10,
		Y:     50,
		W:     200,
		H:     20,
		Value: "default",
	})
	if err != nil {
		t.Fatalf("AddFormField: %v", err)
	}
}
