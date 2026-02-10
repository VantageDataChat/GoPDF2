package gopdf

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost_test.go — Tests to boost coverage for files
// with 0% or low coverage. All functions prefixed TestCov_.
// ============================================================

// ============================================================
// colorspace_convert.go — pure color conversion functions
// ============================================================

func TestCov_RGBToGray(t *testing.T) {
	// Pure white → 1.0
	g := rgbToGray(1, 1, 1)
	if math.Abs(g-1.0) > 0.001 {
		t.Errorf("white: got %f, want 1.0", g)
	}
	// Pure black → 0.0
	g = rgbToGray(0, 0, 0)
	if g != 0 {
		t.Errorf("black: got %f, want 0.0", g)
	}
	// Known luminance: pure red → 0.299
	g = rgbToGray(1, 0, 0)
	if math.Abs(g-0.299) > 0.001 {
		t.Errorf("red: got %f, want 0.299", g)
	}
	// Pure green → 0.587
	g = rgbToGray(0, 1, 0)
	if math.Abs(g-0.587) > 0.001 {
		t.Errorf("green: got %f, want 0.587", g)
	}
}

func TestCov_RGBToCMYK(t *testing.T) {
	// Pure black
	c, m, y, k := rgbToCMYK(0, 0, 0)
	if k != 1 || c != 0 || m != 0 || y != 0 {
		t.Errorf("black: got c=%f m=%f y=%f k=%f", c, m, y, k)
	}
	// Pure white
	c, m, y, k = rgbToCMYK(1, 1, 1)
	if k != 0 || c != 0 || m != 0 || y != 0 {
		t.Errorf("white: got c=%f m=%f y=%f k=%f", c, m, y, k)
	}
	// Pure red → c=0, m=1, y=1, k=0
	c, m, y, k = rgbToCMYK(1, 0, 0)
	if math.Abs(c) > 0.001 || math.Abs(m-1) > 0.001 || math.Abs(y-1) > 0.001 || math.Abs(k) > 0.001 {
		t.Errorf("red: got c=%f m=%f y=%f k=%f", c, m, y, k)
	}
}

func TestCov_CMYKToRGB(t *testing.T) {
	// Pure black (k=1)
	r, g, b := cmykToRGB(0, 0, 0, 1)
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("black: got r=%f g=%f b=%f", r, g, b)
	}
	// No ink → white
	r, g, b = cmykToRGB(0, 0, 0, 0)
	if r != 1 || g != 1 || b != 1 {
		t.Errorf("white: got r=%f g=%f b=%f", r, g, b)
	}
	// Cyan only
	r, g, b = cmykToRGB(1, 0, 0, 0)
	if math.Abs(r) > 0.001 || math.Abs(g-1) > 0.001 || math.Abs(b-1) > 0.001 {
		t.Errorf("cyan: got r=%f g=%f b=%f", r, g, b)
	}
}

func TestCov_ParseColorFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"0.5", 0.5},
		{"1.0", 1.0},
		{"0", 0},
		{"  0.75  ", 0.75},
		{"invalid", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseColorFloat(tt.input)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("parseColorFloat(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestCov_ConvertRGBOp_ToGray(t *testing.T) {
	// Non-stroking
	result := convertRGBOp(1, 0, 0, false, ColorspaceGray)
	if !strings.HasSuffix(result, "g") {
		t.Errorf("expected non-stroking gray op, got %q", result)
	}
	// Stroking
	result = convertRGBOp(1, 0, 0, true, ColorspaceGray)
	if !strings.HasSuffix(result, "G") {
		t.Errorf("expected stroking gray op, got %q", result)
	}
}

func TestCov_ConvertRGBOp_ToCMYK(t *testing.T) {
	result := convertRGBOp(1, 0, 0, false, ColorspaceCMYK)
	if !strings.HasSuffix(result, "k") {
		t.Errorf("expected non-stroking CMYK op, got %q", result)
	}
	result = convertRGBOp(1, 0, 0, true, ColorspaceCMYK)
	if !strings.HasSuffix(result, "K") {
		t.Errorf("expected stroking CMYK op, got %q", result)
	}
}

func TestCov_ConvertRGBOp_ToRGB(t *testing.T) {
	result := convertRGBOp(0.5, 0.5, 0.5, false, ColorspaceRGB)
	if !strings.HasSuffix(result, "rg") {
		t.Errorf("expected non-stroking RGB op, got %q", result)
	}
	result = convertRGBOp(0.5, 0.5, 0.5, true, ColorspaceRGB)
	if !strings.HasSuffix(result, "RG") {
		t.Errorf("expected stroking RGB op, got %q", result)
	}
}

func TestCov_ConvertCMYKOp_ToGray(t *testing.T) {
	result := convertCMYKOp(0, 0, 0, 1, false, ColorspaceGray)
	if !strings.HasSuffix(result, "g") {
		t.Errorf("expected non-stroking gray, got %q", result)
	}
	result = convertCMYKOp(0, 0, 0, 1, true, ColorspaceGray)
	if !strings.HasSuffix(result, "G") {
		t.Errorf("expected stroking gray, got %q", result)
	}
}

func TestCov_ConvertCMYKOp_ToRGB(t *testing.T) {
	result := convertCMYKOp(0, 0, 0, 0, false, ColorspaceRGB)
	if !strings.HasSuffix(result, "rg") {
		t.Errorf("expected non-stroking RGB, got %q", result)
	}
	result = convertCMYKOp(0, 0, 0, 0, true, ColorspaceRGB)
	if !strings.HasSuffix(result, "RG") {
		t.Errorf("expected stroking RGB, got %q", result)
	}
}

func TestCov_ConvertCMYKOp_ToCMYK(t *testing.T) {
	result := convertCMYKOp(0.1, 0.2, 0.3, 0.4, false, ColorspaceCMYK)
	if !strings.HasSuffix(result, "k") {
		t.Errorf("expected non-stroking CMYK, got %q", result)
	}
	result = convertCMYKOp(0.1, 0.2, 0.3, 0.4, true, ColorspaceCMYK)
	if !strings.HasSuffix(result, "K") {
		t.Errorf("expected stroking CMYK, got %q", result)
	}
}

func TestCov_ConvertGrayOp_ToRGB(t *testing.T) {
	result := convertGrayOp(0.5, false, ColorspaceRGB)
	if !strings.HasSuffix(result, "rg") {
		t.Errorf("expected non-stroking RGB, got %q", result)
	}
	result = convertGrayOp(0.5, true, ColorspaceRGB)
	if !strings.HasSuffix(result, "RG") {
		t.Errorf("expected stroking RGB, got %q", result)
	}
}

func TestCov_ConvertGrayOp_ToCMYK(t *testing.T) {
	result := convertGrayOp(0.5, false, ColorspaceCMYK)
	if !strings.HasSuffix(result, "k") {
		t.Errorf("expected non-stroking CMYK, got %q", result)
	}
	result = convertGrayOp(0.5, true, ColorspaceCMYK)
	if !strings.HasSuffix(result, "K") {
		t.Errorf("expected stroking CMYK, got %q", result)
	}
}

func TestCov_ConvertGrayOp_ToGray(t *testing.T) {
	result := convertGrayOp(0.5, false, ColorspaceGray)
	if !strings.HasSuffix(result, "g") {
		t.Errorf("expected non-stroking gray, got %q", result)
	}
	result = convertGrayOp(0.5, true, ColorspaceGray)
	if !strings.HasSuffix(result, "G") {
		t.Errorf("expected stroking gray, got %q", result)
	}
}

func TestCov_ConvertColorLine_RGB(t *testing.T) {
	// rg operator
	line := "1.0000 0.0000 0.0000 rg"
	result := convertColorLine(line, ColorspaceGray)
	if !strings.HasSuffix(result, "g") {
		t.Errorf("rg→gray: got %q", result)
	}
	// RG operator
	line = "0.0000 1.0000 0.0000 RG"
	result = convertColorLine(line, ColorspaceCMYK)
	if !strings.HasSuffix(result, "K") {
		t.Errorf("RG→CMYK: got %q", result)
	}
}

func TestCov_ConvertColorLine_CMYK(t *testing.T) {
	line := "0.0000 1.0000 1.0000 0.0000 k"
	result := convertColorLine(line, ColorspaceGray)
	if !strings.HasSuffix(result, "g") {
		t.Errorf("k→gray: got %q", result)
	}
	line = "0.0000 1.0000 1.0000 0.0000 K"
	result = convertColorLine(line, ColorspaceRGB)
	if !strings.HasSuffix(result, "RG") {
		t.Errorf("K→RGB: got %q", result)
	}
}

func TestCov_ConvertColorLine_Gray(t *testing.T) {
	line := "0.5000 g"
	result := convertColorLine(line, ColorspaceRGB)
	if !strings.HasSuffix(result, "rg") {
		t.Errorf("g→RGB: got %q", result)
	}
	line = "0.5000 G"
	result = convertColorLine(line, ColorspaceCMYK)
	if !strings.HasSuffix(result, "K") {
		t.Errorf("G→CMYK: got %q", result)
	}
}

func TestCov_ConvertColorLine_NonColor(t *testing.T) {
	line := "100 200 m"
	result := convertColorLine(line, ColorspaceGray)
	if result != line {
		t.Errorf("non-color line should be unchanged, got %q", result)
	}
}

func TestCov_ConvertColorLine_Empty(t *testing.T) {
	result := convertColorLine("", ColorspaceGray)
	if result != "" {
		t.Errorf("empty line should return empty, got %q", result)
	}
}

func TestCov_ConvertColorLine_InsufficientArgs(t *testing.T) {
	// rg needs 3 args before it, only 2 provided
	line := "0.5 0.5 rg"
	result := convertColorLine(line, ColorspaceGray)
	// Should still convert since parts[0]=0.5, parts[1]=0.5, parts[2]=rg → len(parts)=3 < 4
	// Actually with "0.5 0.5 rg", parts = ["0.5", "0.5", "rg"], len=3, needs >=4, so unchanged
	if result != line {
		t.Errorf("insufficient args should leave line unchanged, got %q", result)
	}
}

func TestCov_ConvertStreamColorspace(t *testing.T) {
	stream := []byte("1.0000 0.0000 0.0000 rg\n100 200 m\n0.5000 G\n")
	result := convertStreamColorspace(stream, ColorspaceGray)
	s := string(result)
	if strings.Contains(s, "rg") {
		t.Error("expected rg to be converted to g")
	}
	if !strings.Contains(s, "100 200 m") {
		t.Error("non-color line should be preserved")
	}
}


// ============================================================
// content_stream_clean.go — content stream cleaning functions
// ============================================================

func TestCov_SplitContentLines(t *testing.T) {
	stream := []byte("q\n1 0 0 1 0 0 cm\n  \nQ\n")
	lines := splitContentLines(stream)
	// Should skip empty lines
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			t.Error("splitContentLines should skip empty lines")
		}
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

func TestCov_SplitContentLines_Empty(t *testing.T) {
	lines := splitContentLines([]byte(""))
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty input, got %d", len(lines))
	}
}

func TestCov_ExtractOperator(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"1 0 0 1 0 0 cm", "cm"},
		{"q", "q"},
		{"Q", "Q"},
		{"0.5 g", "g"},
		{"1.0 0.0 0.0 rg", "rg"},
		{"", ""},
		{"  ", ""},
	}
	for _, tt := range tests {
		got := extractOperator(tt.line)
		if got != tt.want {
			t.Errorf("extractOperator(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestCov_RemoveRedundantStateChanges(t *testing.T) {
	lines := []string{
		"1 w",
		"2 w",   // should replace previous
		"3 w",   // should replace previous
		"q",     // non-state op, kept
		"10 Tc",
		"20 Tc", // should replace previous
	}
	result := removeRedundantStateChanges(lines)
	// Only "3 w", "q", "20 Tc" should remain
	if len(result) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(result), result)
	}
	if result[0] != "3 w" {
		t.Errorf("expected '3 w', got %q", result[0])
	}
	if result[1] != "q" {
		t.Errorf("expected 'q', got %q", result[1])
	}
	if result[2] != "20 Tc" {
		t.Errorf("expected '20 Tc', got %q", result[2])
	}
}

func TestCov_RemoveRedundantStateChanges_Empty(t *testing.T) {
	result := removeRedundantStateChanges(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestCov_RemoveEmptyQBlocks(t *testing.T) {
	lines := []string{"q", "Q"} // empty save/restore
	result := removeEmptyQBlocks(lines)
	if len(result) != 0 {
		t.Errorf("expected empty after removing q/Q pair, got %v", result)
	}
}

func TestCov_RemoveEmptyQBlocks_Nested(t *testing.T) {
	// After removing inner q/Q, outer becomes empty too
	lines := []string{"q", "q", "Q", "Q"}
	result := removeEmptyQBlocks(lines)
	if len(result) != 0 {
		t.Errorf("expected empty after nested removal, got %v", result)
	}
}

func TestCov_RemoveEmptyQBlocks_WithContent(t *testing.T) {
	lines := []string{"q", "100 200 m", "Q"}
	result := removeEmptyQBlocks(lines)
	if len(result) != 3 {
		t.Errorf("expected 3 lines (non-empty block), got %d", len(result))
	}
}

func TestCov_NormalizeWhitespace(t *testing.T) {
	lines := []string{
		"  1   0   0   1   0   0  cm  ",
		"q",
		"  hello   world  ",
	}
	result := normalizeWhitespace(lines)
	if result[0] != "1 0 0 1 0 0 cm" {
		t.Errorf("expected normalized, got %q", result[0])
	}
	if result[1] != "q" {
		t.Errorf("expected 'q', got %q", result[1])
	}
}

func TestCov_CleanContentStream(t *testing.T) {
	stream := []byte("1 w\n2 w\nq\nQ\n100 200 m\n")
	result := cleanContentStream(stream)
	s := string(result)
	// "1 w" should be removed (replaced by "2 w")
	// "q\nQ" should be removed (empty block)
	if strings.Contains(s, "1 w") {
		t.Error("redundant '1 w' should be removed")
	}
}

func TestCov_BuildCleanedDict(t *testing.T) {
	dict := "<< /Length 100 /Filter /FlateDecode >>"
	result := buildCleanedDict(dict, 200)
	if !strings.Contains(result, "/Length 200") {
		t.Errorf("expected /Length 200, got %q", result)
	}
	if !strings.Contains(result, "/FlateDecode") {
		t.Error("expected /FlateDecode to be preserved")
	}
}

func TestCov_BuildCleanedDict_AddFilter(t *testing.T) {
	dict := "<< /Length 100 >>"
	result := buildCleanedDict(dict, 50)
	if !strings.Contains(result, "/FlateDecode") {
		t.Error("expected /FlateDecode to be added")
	}
	if !strings.Contains(result, "/Length 50") {
		t.Errorf("expected /Length 50, got %q", result)
	}
}

// ============================================================
// arc.go — arc drawing functions
// ============================================================

func TestCov_WriteArcBezier(t *testing.T) {
	var buf bytes.Buffer
	// 90-degree arc from 0 to π/2
	writeArcBezier(&buf, 100, 100, 50, 0, math.Pi/2)
	s := buf.String()
	if !strings.HasSuffix(strings.TrimSpace(s), "c") {
		t.Errorf("expected Bézier curve command 'c', got %q", s)
	}
	if len(s) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestCov_WriteArcSegments_90Degrees(t *testing.T) {
	var buf bytes.Buffer
	writeArcSegments(&buf, 100, 100, 50, 0, math.Pi/2)
	s := buf.String()
	// Single segment for 90 degrees
	count := strings.Count(s, " c\n")
	if count != 1 {
		t.Errorf("expected 1 Bézier segment for 90°, got %d", count)
	}
}

func TestCov_WriteArcSegments_FullCircle(t *testing.T) {
	var buf bytes.Buffer
	writeArcSegments(&buf, 100, 100, 50, 0, 2*math.Pi)
	s := buf.String()
	// Full circle = 4 segments (each ≤90°)
	count := strings.Count(s, " c\n")
	if count != 4 {
		t.Errorf("expected 4 Bézier segments for 360°, got %d", count)
	}
}

func TestCov_WriteArcSegments_ZeroSweep(t *testing.T) {
	var buf bytes.Buffer
	writeArcSegments(&buf, 100, 100, 50, 0, 0)
	if buf.Len() != 0 {
		t.Error("expected no output for zero sweep")
	}
}

func TestCov_WriteArcSegments_NegativeSweep(t *testing.T) {
	var buf bytes.Buffer
	// endRad < startRad → should normalize to positive sweep
	writeArcSegments(&buf, 100, 100, 50, math.Pi, math.Pi/2)
	if buf.Len() == 0 {
		t.Error("expected output for negative sweep (normalized)")
	}
}

func TestCov_WriteArcSegments_SmallArc(t *testing.T) {
	var buf bytes.Buffer
	writeArcSegments(&buf, 0, 0, 10, 0, math.Pi/6) // 30 degrees
	s := buf.String()
	count := strings.Count(s, " c\n")
	if count != 1 {
		t.Errorf("expected 1 segment for 30°, got %d", count)
	}
}

// ============================================================
// strhelper.go — string helper functions
// ============================================================

func TestCov_CreateEmbeddedFontSubsetName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Arial Bold", "Arial+Bold"},
		{"Times/Roman", "Times+Roman"},
		{"NoSpaces", "NoSpaces"},
		{"Multiple Spaces Here", "Multiple+Spaces+Here"},
		{"Slash/And Space", "Slash+And+Space"},
	}
	for _, tt := range tests {
		got := CreateEmbeddedFontSubsetName(tt.input)
		if got != tt.want {
			t.Errorf("CreateEmbeddedFontSubsetName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCov_ReadShortFromByte(t *testing.T) {
	tests := []struct {
		data   []byte
		offset int
		want   int64
	}{
		{[]byte{0x00, 0x01}, 0, 1},
		{[]byte{0x00, 0xFF}, 0, 255},
		{[]byte{0x7F, 0xFF}, 0, 32767},   // max positive short
		{[]byte{0x80, 0x00}, 0, -32768},   // min negative short
		{[]byte{0xFF, 0xFF}, 0, -1},       // -1
		{[]byte{0x00, 0x00, 0x01, 0x00}, 2, 256}, // with offset
	}
	for _, tt := range tests {
		got, size := ReadShortFromByte(tt.data, tt.offset)
		if got != tt.want {
			t.Errorf("ReadShortFromByte(%v, %d) = %d, want %d", tt.data, tt.offset, got, tt.want)
		}
		if size != 2 {
			t.Errorf("ReadShortFromByte size = %d, want 2", size)
		}
	}
}

func TestCov_ReadUShortFromByte(t *testing.T) {
	tests := []struct {
		data   []byte
		offset int
		want   uint64
	}{
		{[]byte{0x00, 0x01}, 0, 1},
		{[]byte{0x00, 0xFF}, 0, 255},
		{[]byte{0xFF, 0xFF}, 0, 65535},
		{[]byte{0x80, 0x00}, 0, 32768},
		{[]byte{0x00, 0x00, 0x01, 0x00}, 2, 256},
	}
	for _, tt := range tests {
		got, size := ReadUShortFromByte(tt.data, tt.offset)
		if got != tt.want {
			t.Errorf("ReadUShortFromByte(%v, %d) = %d, want %d", tt.data, tt.offset, got, tt.want)
		}
		if size != 2 {
			t.Errorf("ReadUShortFromByte size = %d, want 2", size)
		}
	}
}

// ============================================================
// box.go — Box unit conversion
// ============================================================

func TestCov_Box_UnitsToPoints_Nil(t *testing.T) {
	var box *Box
	result := box.UnitsToPoints(UnitMM)
	if result != nil {
		t.Error("nil box should return nil")
	}
}

func TestCov_Box_UnitsToPoints_PT(t *testing.T) {
	box := &Box{Left: 10, Top: 20, Right: 30, Bottom: 40}
	result := box.UnitsToPoints(UnitPT)
	// Points → Points should be identity
	if result.Left != 10 || result.Top != 20 || result.Right != 30 || result.Bottom != 40 {
		t.Errorf("PT conversion should be identity, got %+v", result)
	}
}

func TestCov_Box_UnitsToPoints_MM(t *testing.T) {
	box := &Box{Left: 25.4, Top: 25.4, Right: 50.8, Bottom: 50.8}
	result := box.UnitsToPoints(UnitMM)
	// 25.4mm = 72pt, 50.8mm = 144pt
	if math.Abs(result.Left-72) > 0.1 {
		t.Errorf("25.4mm should be ~72pt, got %f", result.Left)
	}
	if math.Abs(result.Right-144) > 0.1 {
		t.Errorf("50.8mm should be ~144pt, got %f", result.Right)
	}
}

func TestCov_Box_UnitsToPoints_WithOverride(t *testing.T) {
	box := &Box{
		Left: 1, Top: 1, Right: 1, Bottom: 1,
		unitOverride: defaultUnitConfig{Unit: UnitIN},
	}
	// Override should take precedence over the passed unit
	result := box.UnitsToPoints(UnitMM)
	// 1 inch = 72 points
	if math.Abs(result.Left-72) > 0.1 {
		t.Errorf("1 inch should be 72pt, got %f", result.Left)
	}
}

func TestCov_Box_unitsToPoints_Nil(t *testing.T) {
	var box *Box
	result := box.unitsToPoints(defaultUnitConfig{Unit: UnitMM})
	if result != nil {
		t.Error("nil box should return nil")
	}
}

func TestCov_Box_unitsToPoints_WithOverride(t *testing.T) {
	box := &Box{
		Left: 1, Top: 1, Right: 1, Bottom: 1,
		unitOverride: defaultUnitConfig{Unit: UnitIN},
	}
	result := box.unitsToPoints(defaultUnitConfig{Unit: UnitMM})
	if math.Abs(result.Left-72) > 0.1 {
		t.Errorf("override to IN: 1 inch should be 72pt, got %f", result.Left)
	}
}

// ============================================================
// rect.go — Rect unit conversion
// ============================================================

func TestCov_Rect_PointsToUnits_Nil(t *testing.T) {
	var rect *Rect
	result := rect.PointsToUnits(UnitMM)
	if result != nil {
		t.Error("nil rect should return nil")
	}
}

func TestCov_Rect_PointsToUnits_MM(t *testing.T) {
	rect := &Rect{W: 72, H: 144}
	result := rect.PointsToUnits(UnitMM)
	// 72pt = 25.4mm, 144pt = 50.8mm
	if math.Abs(result.W-25.4) > 0.1 {
		t.Errorf("72pt should be ~25.4mm, got %f", result.W)
	}
	if math.Abs(result.H-50.8) > 0.1 {
		t.Errorf("144pt should be ~50.8mm, got %f", result.H)
	}
}

func TestCov_Rect_PointsToUnits_WithOverride(t *testing.T) {
	rect := &Rect{
		W: 72, H: 72,
		unitOverride: defaultUnitConfig{Unit: UnitIN},
	}
	result := rect.PointsToUnits(UnitMM) // override should win
	// 72pt = 1 inch
	if math.Abs(result.W-1) > 0.01 {
		t.Errorf("72pt should be 1 inch, got %f", result.W)
	}
}

func TestCov_Rect_UnitsToPoints_Nil(t *testing.T) {
	var rect *Rect
	result := rect.UnitsToPoints(UnitMM)
	if result != nil {
		t.Error("nil rect should return nil")
	}
}

func TestCov_Rect_UnitsToPoints_IN(t *testing.T) {
	rect := &Rect{W: 1, H: 2}
	result := rect.UnitsToPoints(UnitIN)
	// 1in = 72pt, 2in = 144pt
	if math.Abs(result.W-72) > 0.1 {
		t.Errorf("1in should be 72pt, got %f", result.W)
	}
	if math.Abs(result.H-144) > 0.1 {
		t.Errorf("2in should be 144pt, got %f", result.H)
	}
}

func TestCov_Rect_UnitsToPoints_WithOverride(t *testing.T) {
	rect := &Rect{
		W: 2.54, H: 2.54,
		unitOverride: defaultUnitConfig{Unit: UnitCM},
	}
	result := rect.UnitsToPoints(UnitMM) // override should win
	// 2.54cm = 72pt
	if math.Abs(result.W-72) > 0.1 {
		t.Errorf("2.54cm should be ~72pt, got %f", result.W)
	}
}

func TestCov_Rect_unitsToPoints_Nil(t *testing.T) {
	var rect *Rect
	result := rect.unitsToPoints(defaultUnitConfig{Unit: UnitMM})
	if result != nil {
		t.Error("nil rect should return nil")
	}
}

func TestCov_Rect_unitsToPoints_WithOverride(t *testing.T) {
	rect := &Rect{
		W: 1, H: 1,
		unitOverride: defaultUnitConfig{Unit: UnitIN},
	}
	result := rect.unitsToPoints(defaultUnitConfig{Unit: UnitMM})
	if math.Abs(result.W-72) > 0.1 {
		t.Errorf("override to IN: 1in should be 72pt, got %f", result.W)
	}
}


// ============================================================
// ocg.go — OCG/OCMD/LayerConfig tests
// ============================================================

func TestCov_OCG_DefaultIntent(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	ocg := pdf.AddOCG(OCG{Name: "Test"})
	if ocg.Intent != OCGIntentView {
		t.Errorf("default intent should be View, got %s", ocg.Intent)
	}
}

func TestCov_OCG_DesignIntent(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	ocg := pdf.AddOCG(OCG{Name: "Design", Intent: OCGIntentDesign})
	if ocg.Intent != OCGIntentDesign {
		t.Errorf("intent should be Design, got %s", ocg.Intent)
	}
}

func TestCov_OCMD_AddAndGet(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	l1 := pdf.AddOCG(OCG{Name: "Layer1", On: true})
	l2 := pdf.AddOCG(OCG{Name: "Layer2", On: false})

	ocmd := pdf.AddOCMD(OCMD{
		OCGs:   []OCG{l1, l2},
		Policy: OCGVisibilityAllOn,
	})
	if ocmd.Policy != OCGVisibilityAllOn {
		t.Errorf("policy should be AllOn, got %s", ocmd.Policy)
	}

	got, err := pdf.GetOCMD(ocmd)
	if err != nil {
		t.Fatalf("GetOCMD: %v", err)
	}
	if len(got.OCGs) != 2 {
		t.Errorf("expected 2 OCGs, got %d", len(got.OCGs))
	}
	if got.Policy != OCGVisibilityAllOn {
		t.Errorf("policy should be AllOn, got %s", got.Policy)
	}
}

func TestCov_OCMD_DefaultPolicy(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	l1 := pdf.AddOCG(OCG{Name: "L1", On: true})
	ocmd := pdf.AddOCMD(OCMD{OCGs: []OCG{l1}})
	if ocmd.Policy != OCGVisibilityAnyOn {
		t.Errorf("default policy should be AnyOn, got %s", ocmd.Policy)
	}
}

func TestCov_OCMD_GetInvalid(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	_, err := pdf.GetOCMD(OCMD{objIndex: -1})
	if err == nil {
		t.Error("expected error for invalid OCMD index")
	}
	_, err = pdf.GetOCMD(OCMD{objIndex: 9999})
	if err == nil {
		t.Error("expected error for out-of-range OCMD index")
	}
}

func TestCov_OCG_SetStateNotFound(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.SetOCGState("NonExistent", true)
	if err == nil {
		t.Error("expected error for non-existent OCG")
	}
}

func TestCov_OCG_SetStatesPartialError(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddOCG(OCG{Name: "Exists", On: true})
	err := pdf.SetOCGStates(map[string]bool{
		"Exists":    false,
		"NotExists": true,
	})
	if err == nil {
		t.Error("expected error when one OCG doesn't exist")
	}
}

func TestCov_OCG_SwitchLayer_Exclusive(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.AddOCG(OCG{Name: "A", On: true})
	pdf.AddOCG(OCG{Name: "B", On: true})
	pdf.AddOCG(OCG{Name: "C", On: true})

	err := pdf.SwitchLayer("B", true) // exclusive
	if err != nil {
		t.Fatalf("SwitchLayer: %v", err)
	}
	ocgs := pdf.GetOCGs()
	for _, o := range ocgs {
		if o.Name == "B" && !o.On {
			t.Error("B should be on")
		}
		if o.Name != "B" && o.On {
			t.Errorf("%s should be off in exclusive mode", o.Name)
		}
	}
}

func TestCov_OCG_SwitchLayer_NotFound(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.SwitchLayer("Ghost", false)
	if err == nil {
		t.Error("expected error for non-existent layer")
	}
}

func TestCov_OCG_LayerUIConfig(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	l1 := pdf.AddOCG(OCG{Name: "Locked", On: true})

	pdf.SetLayerUIConfig(LayerUIConfig{
		Name:      "My Config",
		Creator:   "Test",
		BaseState: "ON",
		Locked:    []OCG{l1},
	})

	cfg := pdf.GetLayerUIConfig()
	if cfg == nil {
		t.Fatal("expected non-nil UI config")
	}
	if cfg.Name != "My Config" {
		t.Errorf("name = %q, want 'My Config'", cfg.Name)
	}
	if cfg.BaseState != "ON" {
		t.Errorf("baseState = %q, want 'ON'", cfg.BaseState)
	}
	if len(cfg.Locked) != 1 {
		t.Errorf("expected 1 locked OCG, got %d", len(cfg.Locked))
	}
}

func TestCov_OCG_LayerUIConfig_Nil(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	cfg := pdf.GetLayerUIConfig()
	if cfg != nil {
		t.Error("expected nil UI config when not set")
	}
}

func TestCov_OCG_AddLayerConfig(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	l1 := pdf.AddOCG(OCG{Name: "Content", On: true})
	l2 := pdf.AddOCG(OCG{Name: "Draft", On: false})

	pdf.AddLayerConfig(LayerConfig{
		Name:    "Print",
		Creator: "Test",
		OnOCGs:  []OCG{l1},
		OffOCGs: []OCG{l2},
		Order:   []OCG{l1, l2},
	})

	configs := pdf.GetLayerConfigs()
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].Name != "Print" {
		t.Errorf("config name = %q, want 'Print'", configs[0].Name)
	}
}

func TestCov_OCGObj_Write(t *testing.T) {
	obj := ocgObj{name: "TestLayer", intent: OCGIntentDesign}
	var buf bytes.Buffer
	err := obj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "/Type /OCG") {
		t.Error("expected /Type /OCG")
	}
	if !strings.Contains(s, "(TestLayer)") {
		t.Error("expected layer name")
	}
	if !strings.Contains(s, "/Intent /Design") {
		t.Error("expected /Intent /Design")
	}
}

func TestCov_OCGObj_Write_DefaultIntent(t *testing.T) {
	obj := ocgObj{name: "Default"}
	var buf bytes.Buffer
	obj.write(&buf, 1)
	if !strings.Contains(buf.String(), "/Intent /View") {
		t.Error("expected default intent /View")
	}
}

func TestCov_OCMDObj_Write(t *testing.T) {
	obj := ocmdObj{
		ocgObjIDs: []int{1, 2, 3},
		policy:    OCGVisibilityAllOff,
	}
	var buf bytes.Buffer
	err := obj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "/Type /OCMD") {
		t.Error("expected /Type /OCMD")
	}
	if !strings.Contains(s, "/P /AllOff") {
		t.Error("expected /P /AllOff")
	}
	if !strings.Contains(s, "1 0 R") {
		t.Error("expected OCG references")
	}
}

func TestCov_OCMDObj_Write_DefaultPolicy(t *testing.T) {
	obj := ocmdObj{ocgObjIDs: []int{1}}
	var buf bytes.Buffer
	obj.write(&buf, 1)
	if !strings.Contains(buf.String(), "/P /AnyOn") {
		t.Error("expected default policy /AnyOn")
	}
}

func TestCov_OCPropertiesObj_Write(t *testing.T) {
	obj := ocPropertiesObj{
		ocgs: []ocgRef{
			{objID: 1, on: true},
			{objID: 2, on: false},
		},
	}
	var buf bytes.Buffer
	err := obj.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "/OCGs [") {
		t.Error("expected /OCGs array")
	}
	if !strings.Contains(s, "/ON [") {
		t.Error("expected /ON array")
	}
	if !strings.Contains(s, "/OFF [") {
		t.Error("expected /OFF array")
	}
}

func TestCov_OCPropertiesObj_Write_WithUIConfig(t *testing.T) {
	obj := ocPropertiesObj{
		ocgs: []ocgRef{{objID: 1, on: true}},
		uiConfig: &LayerUIConfig{
			Name:      "UI",
			Creator:   "Test",
			BaseState: "ON",
			Locked:    []OCG{{objIndex: 0}},
		},
	}
	var buf bytes.Buffer
	obj.write(&buf, 1)
	s := buf.String()
	if !strings.Contains(s, "/Name (UI)") {
		t.Error("expected /Name")
	}
	if !strings.Contains(s, "/Creator (Test)") {
		t.Error("expected /Creator")
	}
	if !strings.Contains(s, "/BaseState /ON") {
		t.Error("expected /BaseState")
	}
	if !strings.Contains(s, "/Locked [") {
		t.Error("expected /Locked")
	}
}

func TestCov_OCPropertiesObj_Write_WithLayerConfigs(t *testing.T) {
	obj := ocPropertiesObj{
		ocgs: []ocgRef{{objID: 1, on: true}},
		layerConfigs: []LayerConfig{
			{
				Name:    "Config1",
				Creator: "Test",
				OnOCGs:  []OCG{{objIndex: 0}},
				OffOCGs: []OCG{{objIndex: 1}},
				Order:   []OCG{{objIndex: 0}},
			},
		},
	}
	var buf bytes.Buffer
	obj.write(&buf, 1)
	s := buf.String()
	if !strings.Contains(s, "/Configs [") {
		t.Error("expected /Configs array")
	}
	if !strings.Contains(s, "/Name (Config1)") {
		t.Error("expected config name")
	}
}


// ============================================================
// text_extract.go — text extraction helpers
// ============================================================

func TestCov_Tokenize(t *testing.T) {
	data := []byte("BT /F1 12 Tf (Hello) Tj ET")
	tokens := tokenize(data)
	expected := []string{"BT", "/F1", "12", "Tf", "(Hello)", "Tj", "ET"}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, e := range expected {
		if tokens[i] != e {
			t.Errorf("token[%d] = %q, want %q", i, tokens[i], e)
		}
	}
}

func TestCov_Tokenize_HexString(t *testing.T) {
	data := []byte("<48656C6C6F> Tj")
	tokens := tokenize(data)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != "<48656C6C6F>" {
		t.Errorf("hex token = %q", tokens[0])
	}
}

func TestCov_Tokenize_Comment(t *testing.T) {
	data := []byte("% this is a comment\nBT")
	tokens := tokenize(data)
	if len(tokens) != 1 || tokens[0] != "BT" {
		t.Errorf("expected [BT], got %v", tokens)
	}
}

func TestCov_Tokenize_Dict(t *testing.T) {
	data := []byte("<< /Type /Page >>")
	tokens := tokenize(data)
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != "<<" || tokens[3] != ">>" {
		t.Errorf("dict delimiters: %v", tokens)
	}
}

func TestCov_Tokenize_Array(t *testing.T) {
	data := []byte("[ 1 2 3 ]")
	tokens := tokenize(data)
	if tokens[0] != "[" || tokens[len(tokens)-1] != "]" {
		t.Errorf("array delimiters: %v", tokens)
	}
}

func TestCov_Tokenize_NestedParens(t *testing.T) {
	data := []byte("(Hello (World)) Tj")
	tokens := tokenize(data)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0] != "(Hello (World))" {
		t.Errorf("nested parens: %q", tokens[0])
	}
}

func TestCov_Tokenize_EscapedParens(t *testing.T) {
	data := []byte(`(Hello \) World) Tj`)
	tokens := tokenize(data)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestCov_IsStringToken(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"(Hello)", true},
		{"<48656C6C6F>", true},
		{"<<", false},
		{">>", false},
		{"BT", false},
		{"12", false},
	}
	for _, tt := range tests {
		got := isStringToken(tt.input)
		if got != tt.want {
			t.Errorf("isStringToken(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestCov_FindStringBefore(t *testing.T) {
	tokens := []string{"BT", "/F1", "12", "Tf", "(Hello)", "Tj"}
	result := findStringBefore(tokens, 5) // before Tj
	if result != "(Hello)" {
		t.Errorf("expected (Hello), got %q", result)
	}
}

func TestCov_FindStringBefore_NotFound(t *testing.T) {
	tokens := []string{"BT", "Tj"}
	result := findStringBefore(tokens, 1)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestCov_FindArrayBefore(t *testing.T) {
	tokens := []string{"[", "(Hello)", "10", "(World)", "]", "TJ"}
	result := findArrayBefore(tokens, 5)
	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d: %v", len(result), result)
	}
}

func TestCov_FindArrayBefore_NotFound(t *testing.T) {
	tokens := []string{"BT", "TJ"}
	result := findArrayBefore(tokens, 1)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestCov_DecodeTextString_Empty(t *testing.T) {
	result := decodeTextString("", nil)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestCov_DecodeTextString_ParenString(t *testing.T) {
	result := decodeTextString("(Hello World)", nil)
	if result != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result)
	}
}

func TestCov_DecodeTextString_HexLatin(t *testing.T) {
	// "Hi" = 48 69
	result := decodeTextString("<4869>", nil)
	if result != "Hi" {
		t.Errorf("expected 'Hi', got %q", result)
	}
}

func TestCov_DecodeTextString_HexWithCMap(t *testing.T) {
	cmap := map[uint16]rune{0x0048: 'H', 0x0069: 'i'}
	fi := &fontInfo{toUni: cmap}
	result := decodeTextString("<00480069>", fi)
	if result != "Hi" {
		t.Errorf("expected 'Hi', got %q", result)
	}
}

func TestCov_DecodeTextString_HexUTF16(t *testing.T) {
	fi := &fontInfo{isType0: true}
	// "A" = 0041 in UTF-16BE
	result := decodeTextString("<0041>", fi)
	if result != "A" {
		t.Errorf("expected 'A', got %q", result)
	}
}

func TestCov_DecodeTextString_ParenWithCMap(t *testing.T) {
	cmap := map[uint16]rune{0x41: 'X'}
	fi := &fontInfo{toUni: cmap}
	result := decodeTextString("(A)", fi) // 'A' = 0x41
	if result != "X" {
		t.Errorf("expected 'X' (mapped), got %q", result)
	}
}

func TestCov_DecodeTextString_UTF16BOM(t *testing.T) {
	// UTF-16BE BOM + "A" (0x0041)
	s := "(\xfe\xff\x00A)"
	result := decodeTextString(s, nil)
	if result != "A" {
		t.Errorf("expected 'A', got %q", result)
	}
}

func TestCov_DecodeTextString_RawPassthrough(t *testing.T) {
	result := decodeTextString("SomeRawToken", nil)
	if result != "SomeRawToken" {
		t.Errorf("expected passthrough, got %q", result)
	}
}

func TestCov_UnescapePDFString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`Hello`, "Hello"},
		{`Hello\nWorld`, "Hello\nWorld"},
		{`Tab\there`, "Tab\there"},
		{`Back\bspace`, "Back\bspace"},
		{`Form\ffeed`, "Form\ffeed"},
		{`Return\rhere`, "Return\rhere"},
		{`Paren\(here\)`, "Paren(here)"},
		{`Backslash\\here`, "Backslash\\here"},
		{`Octal\101`, "OctalA"}, // \101 = 'A'
	}
	for _, tt := range tests {
		got := unescapePDFString(tt.input)
		if got != tt.want {
			t.Errorf("unescapePDFString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCov_ParseHex16(t *testing.T) {
	tests := []struct {
		input string
		want  uint16
	}{
		{"0041", 0x41},
		{"FFFF", 0xFFFF},
		{"0000", 0},
		{"00ff", 0xFF},
	}
	for _, tt := range tests {
		got := parseHex16(tt.input)
		if got != tt.want {
			t.Errorf("parseHex16(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCov_ExtractHexPairs(t *testing.T) {
	line := "<0041> <0042>"
	pairs := extractHexPairs(line)
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0] != "0041" || pairs[1] != "0042" {
		t.Errorf("pairs = %v", pairs)
	}
}

func TestCov_ParseCMap_BfChar(t *testing.T) {
	cmap := `beginbfchar
<0041> <0042>
<0043> <0044>
endbfchar`
	m := parseCMap([]byte(cmap))
	if m[0x41] != 'B' {
		t.Errorf("0041 should map to B, got %c", m[0x41])
	}
	if m[0x43] != 'D' {
		t.Errorf("0043 should map to D, got %c", m[0x43])
	}
}

func TestCov_ParseCMap_BfRange(t *testing.T) {
	cmap := `beginbfrange
<0041> <0043> <0061>
endbfrange`
	m := parseCMap([]byte(cmap))
	// 0041→0061('a'), 0042→0062('b'), 0043→0063('c')
	if m[0x41] != 'a' {
		t.Errorf("0041 should map to 'a', got %c", m[0x41])
	}
	if m[0x42] != 'b' {
		t.Errorf("0042 should map to 'b', got %c", m[0x42])
	}
	if m[0x43] != 'c' {
		t.Errorf("0043 should map to 'c', got %c", m[0x43])
	}
}

func TestCov_FontDisplayName(t *testing.T) {
	if fontDisplayName(nil) != "" {
		t.Error("nil font should return empty")
	}
	fi := &fontInfo{baseFont: "Arial"}
	if fontDisplayName(fi) != "Arial" {
		t.Errorf("expected Arial, got %q", fontDisplayName(fi))
	}
	fi2 := &fontInfo{name: "/F1"}
	if fontDisplayName(fi2) != "/F1" {
		t.Errorf("expected /F1, got %q", fontDisplayName(fi2))
	}
}

func TestCov_DecodeHexLatin(t *testing.T) {
	result := decodeHexLatin("48656C6C6F")
	if result != "Hello" {
		t.Errorf("expected 'Hello', got %q", result)
	}
}

func TestCov_DecodeHexUTF16BE(t *testing.T) {
	// "AB" in UTF-16BE = 0041 0042
	result := decodeHexUTF16BE("00410042")
	if result != "AB" {
		t.Errorf("expected 'AB', got %q", result)
	}
}

func TestCov_DecodeHexUTF16BE_OddLength(t *testing.T) {
	// Odd hex length → falls back to latin
	result := decodeHexUTF16BE("414")
	if result == "" {
		t.Error("expected non-empty for odd-length fallback")
	}
}

func TestCov_DecodeUTF16BE(t *testing.T) {
	// "A" = 00 41
	data := []byte{0x00, 0x41}
	result := decodeUTF16BE(data)
	if result != "A" {
		t.Errorf("expected 'A', got %q", result)
	}
}

func TestCov_DecodeUTF16BE_OddBytes(t *testing.T) {
	data := []byte{0x00, 0x41, 0x00}
	result := decodeUTF16BE(data)
	// Should pad and decode
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov_DecodeHexWithCMap_TwoDigit(t *testing.T) {
	cmap := map[uint16]rune{0x41: 'X', 0x42: 'Y'}
	// len("414243") = 6, not divisible by 4 → 2-digit mode
	result := decodeHexWithCMap("414243", cmap)
	// 0x41→X, 0x42→Y, 0x43 not in cmap → raw byte 'C'
	if !strings.HasPrefix(result, "XY") {
		t.Errorf("expected prefix 'XY', got %q", result)
	}
}

func TestCov_DecodeHexWithCMap_FourDigit(t *testing.T) {
	cmap := map[uint16]rune{0x0041: 'A', 0x0042: 'B'}
	result := decodeHexWithCMap("00410042", cmap)
	if result != "AB" {
		t.Errorf("expected 'AB', got %q", result)
	}
}

func TestCov_ExtractName(t *testing.T) {
	dict := "/Type /Page /BaseFont /Arial /Encoding /WinAnsiEncoding"
	name := extractName(dict, "/BaseFont")
	if name != "Arial" {
		t.Errorf("expected 'Arial', got %q", name)
	}
	name = extractName(dict, "/NotExist")
	if name != "" {
		t.Errorf("expected empty for missing key, got %q", name)
	}
}


// ============================================================
// smask_obj.go — SMask object tests
// ============================================================

func TestCov_SMaskOptions_GetId(t *testing.T) {
	opts := SMaskOptions{
		Subtype:                       SMaskAlphaSubtype,
		TransparencyXObjectGroupIndex: 5,
	}
	id := opts.GetId()
	if !strings.Contains(id, "/Alpha") {
		t.Errorf("expected /Alpha in id, got %q", id)
	}
	if !strings.Contains(id, "5") {
		t.Errorf("expected 5 in id, got %q", id)
	}
}

func TestCov_SMask_Write_WithTransparency(t *testing.T) {
	s := SMask{
		S:                             string(SMaskLuminositySubtype),
		TransparencyXObjectGroupIndex: 3,
	}
	var buf bytes.Buffer
	err := s.write(&buf, 1)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/Type /Mask") {
		t.Error("expected /Type /Mask")
	}
	if !strings.Contains(out, "/S /Luminosity") {
		t.Error("expected /S /Luminosity")
	}
	if !strings.Contains(out, "/G 4 0 R") {
		t.Errorf("expected /G 4 0 R, got %q", out)
	}
}

func TestCov_SMask_GetType(t *testing.T) {
	s := SMask{}
	if s.getType() != "Mask" {
		t.Errorf("expected 'Mask', got %q", s.getType())
	}
}

func TestCov_SMaskMap_FindAndSave(t *testing.T) {
	m := NewSMaskMap()
	opts := SMaskOptions{
		Subtype:                       SMaskAlphaSubtype,
		TransparencyXObjectGroupIndex: 1,
	}

	// Not found initially
	_, ok := m.Find(opts)
	if ok {
		t.Error("expected not found")
	}

	// Save and find
	saved := m.Save(opts.GetId(), SMask{Index: 42})
	if saved.Index != 42 {
		t.Errorf("saved index = %d, want 42", saved.Index)
	}

	found, ok := m.Find(opts)
	if !ok {
		t.Error("expected found after save")
	}
	if found.Index != 42 {
		t.Errorf("found index = %d, want 42", found.Index)
	}
}

func TestCov_SMask_Protection(t *testing.T) {
	s := &SMask{}
	if s.protection() != nil {
		t.Error("expected nil protection initially")
	}
	p := &PDFProtection{}
	s.setProtection(p)
	if s.protection() != p {
		t.Error("expected protection to be set")
	}
}

// ============================================================
// cache_content_sector.go — sector drawing
// ============================================================

func TestCov_CacheContentSector_Write(t *testing.T) {
	c := &cacheContentSector{
		pageHeight: 842,
		cx:         200,
		cy:         300,
		r:          50,
		startDeg:   0,
		endDeg:     90,
		style:      "F",
	}
	var buf bytes.Buffer
	err := c.write(&buf, nil)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	s := buf.String()
	if !strings.HasPrefix(s, "q\n") {
		t.Error("expected q at start")
	}
	if !strings.Contains(s, "f\n") {
		t.Error("expected fill command 'f'")
	}
	if !strings.HasSuffix(strings.TrimSpace(s), "Q") {
		t.Error("expected Q at end")
	}
}

func TestCov_CacheContentSector_Write_FD(t *testing.T) {
	c := &cacheContentSector{
		pageHeight: 842,
		cx:         200,
		cy:         300,
		r:          50,
		startDeg:   0,
		endDeg:     180,
		style:      "FD",
	}
	var buf bytes.Buffer
	c.write(&buf, nil)
	if !strings.Contains(buf.String(), "b\n") {
		t.Error("expected fill+stroke command 'b'")
	}
}

func TestCov_CacheContentSector_Write_Stroke(t *testing.T) {
	c := &cacheContentSector{
		pageHeight: 842,
		cx:         200,
		cy:         300,
		r:          50,
		startDeg:   0,
		endDeg:     270,
		style:      "S",
	}
	var buf bytes.Buffer
	c.write(&buf, nil)
	if !strings.Contains(buf.String(), "s\n") {
		t.Error("expected stroke command 's'")
	}
}

func TestCov_CacheContentSector_Write_WithExtGState(t *testing.T) {
	c := &cacheContentSector{
		pageHeight: 842,
		cx:         100,
		cy:         100,
		r:          30,
		startDeg:   0,
		endDeg:     45,
		style:      "F",
		opts:       sectorOptions{extGStateIndexes: []int{0, 1}},
	}
	var buf bytes.Buffer
	c.write(&buf, nil)
	s := buf.String()
	if !strings.Contains(s, "/GS0 gs") {
		t.Error("expected /GS0 gs")
	}
	if !strings.Contains(s, "/GS1 gs") {
		t.Error("expected /GS1 gs")
	}
}

// ============================================================
// cache_content_polyline.go — polyline drawing
// ============================================================

func TestCov_CacheContentPolyline_Write(t *testing.T) {
	c := &cacheContentPolyline{
		pageHeight: 842,
		points: []Point{
			{X: 10, Y: 20},
			{X: 30, Y: 40},
			{X: 50, Y: 60},
		},
	}
	var buf bytes.Buffer
	err := c.write(&buf, nil)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, " m ") {
		t.Error("expected moveto command")
	}
	if !strings.Contains(s, " l ") {
		t.Error("expected lineto command")
	}
	if !strings.Contains(s, "S\n") {
		t.Error("expected stroke command")
	}
}

func TestCov_CacheContentPolyline_Write_TooFewPoints(t *testing.T) {
	c := &cacheContentPolyline{
		pageHeight: 842,
		points:     []Point{{X: 10, Y: 20}},
	}
	var buf bytes.Buffer
	c.write(&buf, nil)
	if buf.Len() != 0 {
		t.Error("expected no output for < 2 points")
	}
}

func TestCov_CacheContentPolyline_Write_WithExtGState(t *testing.T) {
	c := &cacheContentPolyline{
		pageHeight: 842,
		points:     []Point{{X: 0, Y: 0}, {X: 100, Y: 100}},
		opts:       polylineOptions{extGStateIndexes: []int{2}},
	}
	var buf bytes.Buffer
	c.write(&buf, nil)
	if !strings.Contains(buf.String(), "/GS2 gs") {
		t.Error("expected /GS2 gs")
	}
}

// ============================================================
// pdf_lowlevel.go — low-level PDF helpers
// ============================================================

func TestCov_ExtractDictKeyValue(t *testing.T) {
	dict := "/Type /Page /MediaBox [0 0 612 792] /Count 5"

	// Name value
	v := extractDictKeyValue(dict, "/Type")
	if v != "/Page" {
		t.Errorf("/Type = %q, want '/Page'", v)
	}

	// Array value
	v = extractDictKeyValue(dict, "/MediaBox")
	if v != "[0 0 612 792]" {
		t.Errorf("/MediaBox = %q, want '[0 0 612 792]'", v)
	}

	// Number value
	v = extractDictKeyValue(dict, "/Count")
	if v != "5" {
		t.Errorf("/Count = %q, want '5'", v)
	}

	// Missing key
	v = extractDictKeyValue(dict, "/Missing")
	if v != "" {
		t.Errorf("/Missing = %q, want ''", v)
	}
}

func TestCov_ExtractDictKeyValue_String(t *testing.T) {
	dict := "/Title (Hello World) /Author (Test)"
	v := extractDictKeyValue(dict, "/Title")
	if v != "(Hello World)" {
		t.Errorf("/Title = %q, want '(Hello World)'", v)
	}
}

func TestCov_ExtractDictKeyValue_NestedDict(t *testing.T) {
	dict := "/Resources << /Font << /F1 1 0 R >> >> /Type /Page"
	v := extractDictKeyValue(dict, "/Resources")
	if !strings.HasPrefix(v, "<<") || !strings.HasSuffix(v, ">>") {
		t.Errorf("/Resources = %q, expected nested dict", v)
	}
}

func TestCov_ExtractDictKeyValue_HexString(t *testing.T) {
	dict := "/ID <48656C6C6F>"
	v := extractDictKeyValue(dict, "/ID")
	if v != "<48656C6C6F>" {
		t.Errorf("/ID = %q, want '<48656C6C6F>'", v)
	}
}

func TestCov_ExtractDictKeyValue_Reference(t *testing.T) {
	dict := "/Font 5 0 R /Type /Page"
	v := extractDictKeyValue(dict, "/Font")
	if v != "5 0 R" {
		t.Errorf("/Font = %q, want '5 0 R'", v)
	}
}

func TestCov_SetDictKeyValue_New(t *testing.T) {
	dict := "/Type /Page"
	result := setDictKeyValue(dict, "/Count", "5")
	if !strings.Contains(result, "/Count 5") {
		t.Errorf("expected /Count 5, got %q", result)
	}
}

func TestCov_SetDictKeyValue_Replace(t *testing.T) {
	dict := "/Type /Page /Count 3"
	result := setDictKeyValue(dict, "/Count", "5")
	if !strings.Contains(result, "/Count 5") {
		t.Errorf("expected /Count 5, got %q", result)
	}
	if strings.Contains(result, "/Count 3") {
		t.Error("old value should be replaced")
	}
}

// ============================================================
// incremental_save.go — incremental save tests
// ============================================================

func TestCov_IncrementalSave_SpecificIndices(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(100, 100)
	pdf.Text("Page 1")

	var origBuf bytes.Buffer
	pdf.WriteTo(&origBuf)
	origData := origBuf.Bytes()

	// Save with specific modified indices (just index 0)
	result, err := pdf.IncrementalSave(origData, []int{0})
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) <= len(origData) {
		t.Error("incremental save should append data")
	}
	eofMarker := []byte("%%EOF")
	if !bytes.Contains(result, eofMarker) {
		t.Error("expected EOF marker")
	}
}

func TestCov_IncrementalSave_NilIndices(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Text("Test")

	var origBuf bytes.Buffer
	pdf.WriteTo(&origBuf)
	origData := origBuf.Bytes()

	// nil indices → all objects
	result, err := pdf.IncrementalSave(origData, nil)
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov_IncrementalSave_OutOfRangeIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Text("Test")

	var origBuf bytes.Buffer
	pdf.WriteTo(&origBuf)
	origData := origBuf.Bytes()

	// Out-of-range indices should be skipped gracefully
	result, err := pdf.IncrementalSave(origData, []int{-1, 99999})
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov_IncrementalSave_NoTrailingNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Text("Test")

	var origBuf bytes.Buffer
	pdf.WriteTo(&origBuf)
	origData := origBuf.Bytes()

	// Remove trailing newline if present
	if len(origData) > 0 && origData[len(origData)-1] == '\n' {
		origData = origData[:len(origData)-1]
	}

	result, err := pdf.IncrementalSave(origData, []int{0})
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}


// ============================================================
// annotation.go — annotation types and convenience methods
// ============================================================

func TestCov_AnnotationOption_Defaults(t *testing.T) {
	opt := AnnotationOption{}
	opt.defaults()
	if opt.Color != [3]uint8{255, 255, 0} {
		t.Errorf("default color should be yellow, got %v", opt.Color)
	}
	if opt.Opacity != 1.0 {
		t.Errorf("default opacity should be 1.0, got %f", opt.Opacity)
	}
	if opt.FontSize != 12 {
		t.Errorf("default font size should be 12, got %f", opt.FontSize)
	}
	if opt.BorderWidth != 1 {
		t.Errorf("default border width should be 1, got %f", opt.BorderWidth)
	}
	if opt.Stamp != StampDraft {
		t.Errorf("default stamp should be Draft, got %s", opt.Stamp)
	}
	if opt.CreationDate.IsZero() {
		t.Error("default creation date should be set")
	}
}

func TestCov_AnnotationObj_Write_AllTypes(t *testing.T) {
	types := []struct {
		name string
		opt  AnnotationOption
	}{
		{"Text", AnnotationOption{Type: AnnotText, X: 10, Y: 10, W: 24, H: 24, Title: "Author", Content: "Note", Open: true}},
		{"Highlight", AnnotationOption{Type: AnnotHighlight, X: 10, Y: 10, W: 100, H: 20, Content: "highlighted"}},
		{"Underline", AnnotationOption{Type: AnnotUnderline, X: 10, Y: 10, W: 100, H: 20}},
		{"StrikeOut", AnnotationOption{Type: AnnotStrikeOut, X: 10, Y: 10, W: 100, H: 20}},
		{"Square", AnnotationOption{Type: AnnotSquare, X: 10, Y: 10, W: 50, H: 50, InteriorColor: &[3]uint8{255, 0, 0}, Content: "box"}},
		{"Circle", AnnotationOption{Type: AnnotCircle, X: 10, Y: 10, W: 50, H: 50, InteriorColor: &[3]uint8{0, 255, 0}, Content: "circle"}},
		{"FreeText", AnnotationOption{Type: AnnotFreeText, X: 10, Y: 10, W: 200, H: 30, Content: "Free text", FontSize: 16}},
		{"Ink", AnnotationOption{Type: AnnotInk, X: 10, Y: 10, W: 100, H: 100, Content: "ink",
			InkList: [][]Point{{{X: 10, Y: 10}, {X: 50, Y: 50}}}}},
		{"Polyline", AnnotationOption{Type: AnnotPolyline, X: 10, Y: 10, W: 100, H: 100, Content: "polyline",
			Vertices: []Point{{X: 10, Y: 10}, {X: 50, Y: 50}, {X: 90, Y: 10}},
			LineEndingStyles: [2]LineEndingStyle{LineEndOpenArrow, LineEndClosedArrow}}},
		{"Polygon", AnnotationOption{Type: AnnotPolygon, X: 10, Y: 10, W: 100, H: 100, Content: "polygon",
			Vertices: []Point{{X: 10, Y: 10}, {X: 50, Y: 50}, {X: 90, Y: 10}},
			InteriorColor: &[3]uint8{0, 0, 255}}},
		{"Line", AnnotationOption{Type: AnnotLine, X: 10, Y: 10, W: 100, H: 100, Content: "line",
			LineStart: Point{X: 10, Y: 10}, LineEnd: Point{X: 100, Y: 100},
			LineEndingStyles: [2]LineEndingStyle{LineEndDiamond, LineEndSquare}}},
		{"Stamp", AnnotationOption{Type: AnnotStamp, X: 10, Y: 10, W: 100, H: 50, Stamp: StampApproved, Content: "approved"}},
		{"Squiggly", AnnotationOption{Type: AnnotSquiggly, X: 10, Y: 10, W: 100, H: 20, Content: "squiggly"}},
		{"Caret", AnnotationOption{Type: AnnotCaret, X: 10, Y: 10, W: 10, H: 20, Content: "insert here"}},
		{"FileAttachment", AnnotationOption{Type: AnnotFileAttachment, X: 10, Y: 10, W: 24, H: 24,
			FileName: "test.txt", FileData: []byte("hello"), Content: "attachment"}},
		{"Redact", AnnotationOption{Type: AnnotRedact, X: 10, Y: 10, W: 100, H: 20,
			OverlayText: "REDACTED", InteriorColor: &[3]uint8{0, 0, 0}, Content: "redacted"}},
	}

	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			obj := annotationObj{
				opt: tt.opt,
				getRoot: func() *GoPdf {
					return pdf
				},
			}
			var buf bytes.Buffer
			err := obj.write(&buf, 1)
			if err != nil {
				t.Fatalf("write %s: %v", tt.name, err)
			}
			s := buf.String()
			if !strings.Contains(s, "/Type /Annot") {
				t.Errorf("%s: missing /Type /Annot", tt.name)
			}
		})
	}
}

func TestCov_AnnotationObj_Write_Opacity(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	obj := annotationObj{
		opt: AnnotationOption{
			Type:    AnnotText,
			X:       10, Y: 10, W: 24, H: 24,
			Opacity: 0.5,
			Color:   [3]uint8{255, 0, 0},
		},
		getRoot: func() *GoPdf { return pdf },
	}
	var buf bytes.Buffer
	obj.write(&buf, 1)
	if !strings.Contains(buf.String(), "/CA 0.5") {
		t.Error("expected /CA for opacity < 1")
	}
}

func TestCov_Annotation_ConvenienceMethods(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Test all convenience methods
	pdf.AddTextAnnotation(100, 100, "Author", "Note")
	pdf.AddHighlightAnnotation(100, 200, 200, 20, [3]uint8{255, 255, 0})
	pdf.AddFreeTextAnnotation(100, 300, 200, 30, "Free text", 14)
	pdf.AddInkAnnotation(100, 400, 100, 100, [][]Point{{{X: 100, Y: 400}, {X: 150, Y: 450}}}, [3]uint8{0, 0, 255})
	pdf.AddPolylineAnnotation(100, 500, 100, 50, []Point{{X: 100, Y: 500}, {X: 150, Y: 520}, {X: 200, Y: 500}}, [3]uint8{255, 0, 0})
	pdf.AddPolygonAnnotation(300, 100, 100, 100, []Point{{X: 300, Y: 100}, {X: 350, Y: 150}, {X: 400, Y: 100}}, [3]uint8{0, 255, 0})
	pdf.AddLineAnnotation(Point{X: 300, Y: 200}, Point{X: 400, Y: 300}, [3]uint8{0, 0, 0})
	pdf.AddStampAnnotation(300, 400, 100, 50, StampConfidential)
	pdf.AddSquigglyAnnotation(300, 500, 100, 20, [3]uint8{255, 128, 0})
	pdf.AddCaretAnnotation(300, 600, 10, 20, "insert")
	pdf.AddFileAttachmentAnnotation(500, 100, "test.txt", []byte("data"), "file")
	pdf.AddRedactAnnotation(500, 200, 100, 20, "REDACTED")

	annots := pdf.GetAnnotations()
	if len(annots) != 12 {
		t.Errorf("expected 12 annotations, got %d", len(annots))
	}
}

func TestCov_Annotation_GetOnPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddTextAnnotation(10, 10, "T", "C")

	annots := pdf.GetAnnotationsOnPage(1)
	if len(annots) != 1 {
		t.Errorf("expected 1 annotation on page 1, got %d", len(annots))
	}

	annots = pdf.GetAnnotationsOnPage(99)
	if annots != nil {
		t.Error("expected nil for invalid page")
	}
}

func TestCov_Annotation_DeleteOnPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddTextAnnotation(10, 10, "T", "C")

	ok := pdf.DeleteAnnotationOnPage(1, 0)
	if !ok {
		t.Error("expected successful deletion")
	}
	ok = pdf.DeleteAnnotationOnPage(1, 0)
	if ok {
		t.Error("expected failure for empty annotation list")
	}
	ok = pdf.DeleteAnnotationOnPage(99, 0)
	if ok {
		t.Error("expected failure for invalid page")
	}
}

func TestCov_Annotation_ApplyRedactions(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddRedactAnnotation(10, 10, 100, 20, "REDACTED")
	pdf.AddTextAnnotation(200, 200, "Keep", "This stays")

	removed := pdf.ApplyRedactions()
	if removed != 1 {
		t.Errorf("expected 1 redaction removed, got %d", removed)
	}
	annots := pdf.GetAnnotations()
	if len(annots) != 1 {
		t.Errorf("expected 1 remaining annotation, got %d", len(annots))
	}
}

func TestCov_Annotation_ApplyRedactionsOnPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddRedactAnnotation(10, 10, 100, 20, "R1")
	pdf.AddRedactAnnotation(10, 50, 100, 20, "R2")

	removed := pdf.ApplyRedactionsOnPage(1)
	if removed != 2 {
		t.Errorf("expected 2 redactions removed, got %d", removed)
	}

	removed = pdf.ApplyRedactionsOnPage(99)
	if removed != 0 {
		t.Error("expected 0 for invalid page")
	}
}

// ============================================================
// break_option.go
// ============================================================

func TestCov_BreakOption_HasSeparator(t *testing.T) {
	bo := BreakOption{Separator: "-"}
	if !bo.HasSeparator() {
		t.Error("expected HasSeparator true")
	}
	bo2 := BreakOption{}
	if bo2.HasSeparator() {
		t.Error("expected HasSeparator false")
	}
}

func TestCov_DefaultBreakOption(t *testing.T) {
	if DefaultBreakOption.Mode != BreakModeStrict {
		t.Error("default mode should be strict")
	}
	if DefaultBreakOption.HasSeparator() {
		t.Error("default should have no separator")
	}
}

// ============================================================
// annotation_modify.go — annotation modification
// ============================================================

func TestCov_Annotation_Modify_UpdateContent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddTextAnnotation(10, 10, "Author", "Original")

	annots := pdf.GetAnnotations()
	if len(annots) == 0 {
		t.Fatal("no annotations")
	}
	if annots[0].Option.Content != "Original" {
		t.Errorf("expected 'Original', got %q", annots[0].Option.Content)
	}
}

// ============================================================
// Additional edge cases for text_extract.go
// ============================================================

func TestCov_ParseTextOperators_BasicBT_ET(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n100 700 Td\n(Hello) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1", baseFont: "Helvetica"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected at least one text result")
	}
	if results[0].Text != "Hello" {
		t.Errorf("expected 'Hello', got %q", results[0].Text)
	}
}

func TestCov_ParseTextOperators_TJ(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n100 700 Td\n[(Hello) -10 (World)] TJ\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1", baseFont: "Helvetica"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected text result")
	}
	if !strings.Contains(results[0].Text, "Hello") || !strings.Contains(results[0].Text, "World") {
		t.Errorf("expected HelloWorld, got %q", results[0].Text)
	}
}

func TestCov_ParseTextOperators_Tm(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n1 0 0 1 200 500 Tm\n(Positioned) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected text result")
	}
}

func TestCov_ParseTextOperators_TD(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n100 700 TD\n(Line1) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected text result")
	}
}

func TestCov_ParseTextOperators_TStar(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n14 TL\n100 700 Td\n(Line1) Tj\nT*\n(Line2) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) < 2 {
		t.Fatalf("expected 2 text results, got %d", len(results))
	}
}

func TestCov_ParseTextOperators_Quote(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n14 TL\n100 700 Td\n(Line1) '\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected text result from ' operator")
	}
}

func TestCov_ParseTextOperators_DoubleQuote(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n14 TL\n100 700 Td\n1 2 (Text) \"\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected text result from \" operator")
	}
}

func TestCov_ParseTextOperators_CM(t *testing.T) {
	stream := []byte("1 0 0 1 50 50 cm\nBT\n/F1 12 Tf\n100 700 Td\n(Transformed) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "/F1"},
	}
	mediaBox := [4]float64{0, 0, 612, 792}
	results := parseTextOperators(stream, fonts, mediaBox)
	if len(results) == 0 {
		t.Fatal("expected text result with cm transform")
	}
}


// ============================================================
// annotation_modify.go — ModifyAnnotation
// ============================================================

func TestCov_ModifyAnnotation_Success(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddTextAnnotation(10, 10, "Author", "Original")

	err := pdf.ModifyAnnotation(0, 0, AnnotationOption{
		Content:       "Updated",
		Title:         "NewAuthor",
		Color:         [3]uint8{255, 0, 0},
		Opacity:       0.8,
		FontSize:      16,
		BorderWidth:   2,
		X:             50,
		Y:             50,
		W:             100,
		H:             100,
		InteriorColor: &[3]uint8{0, 255, 0},
		OverlayText:   "overlay",
	})
	if err != nil {
		t.Fatalf("ModifyAnnotation: %v", err)
	}

	annots := pdf.GetAnnotationsOnPage(1)
	if annots[0].Option.Content != "Updated" {
		t.Errorf("content not updated: %q", annots[0].Option.Content)
	}
	if annots[0].Option.Title != "NewAuthor" {
		t.Errorf("title not updated: %q", annots[0].Option.Title)
	}
}

func TestCov_ModifyAnnotation_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyAnnotation(99, 0, AnnotationOption{Content: "X"})
	if err == nil {
		t.Error("expected error for invalid page")
	}
}

func TestCov_ModifyAnnotation_InvalidIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyAnnotation(0, 99, AnnotationOption{Content: "X"})
	if err == nil {
		t.Error("expected error for invalid annotation index")
	}
}

// ============================================================
// page_batch_ops.go — DeletePages, MovePage
// ============================================================

func TestCov_DeletePages_Multiple(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 5; i++ {
		pdf.AddPage()
		pdf.SetXY(100, 100)
		pdf.Text("Page " + string(rune('1'+i)))
	}
	before := pdf.GetNumberOfPages()
	err := pdf.DeletePages([]int{2, 4})
	if err != nil {
		t.Fatalf("DeletePages: %v", err)
	}
	after := pdf.GetNumberOfPages()
	if after != before-2 {
		t.Errorf("expected %d pages, got %d", before-2, after)
	}
}

func TestCov_DeletePages_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeletePages([]int{})
	if err != nil {
		t.Errorf("empty delete should succeed: %v", err)
	}
}

func TestCov_DeletePages_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeletePages([]int{99})
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

func TestCov_DeletePages_Duplicates(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(100, 100)
		pdf.Text("P")
	}
	err := pdf.DeletePages([]int{2, 2, 2})
	if err != nil {
		t.Fatalf("DeletePages with duplicates: %v", err)
	}
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov_DeletePages_AllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Text("Only page")
	err := pdf.DeletePages([]int{1})
	if err == nil {
		t.Error("expected error when deleting all pages")
	}
}

func TestCov_MovePage_Forward(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(100, 100)
		pdf.Text("P")
	}
	err := pdf.MovePage(1, 3)
	if err != nil {
		t.Fatalf("MovePage: %v", err)
	}
	if pdf.GetNumberOfPages() != 3 {
		t.Errorf("expected 3 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov_MovePage_Backward(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(100, 100)
		pdf.Text("P")
	}
	err := pdf.MovePage(3, 1)
	if err != nil {
		t.Fatalf("MovePage: %v", err)
	}
}

func TestCov_MovePage_SamePosition(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Text("P")
	err := pdf.MovePage(1, 1)
	if err != nil {
		t.Fatalf("MovePage same position: %v", err)
	}
}

func TestCov_MovePage_InvalidFrom(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.MovePage(99, 1)
	if err == nil {
		t.Error("expected error for invalid from")
	}
}

func TestCov_MovePage_InvalidTo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.MovePage(1, 99)
	if err == nil {
		t.Error("expected error for invalid to")
	}
}


// ============================================================
// page_layout.go — validPageLayout, validPageMode
// ============================================================

func TestCov_ValidPageLayout(t *testing.T) {
	valid := []string{"singlepage", "onecolumn", "twocolumnleft", "twocolumnright", "twopageleft", "twopageright"}
	for _, v := range valid {
		if !validPageLayout(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	if validPageLayout("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestCov_ValidPageMode(t *testing.T) {
	valid := []string{"usenone", "useoutlines", "usethumbs", "fullscreen", "useoc", "useattachments"}
	for _, v := range valid {
		if !validPageMode(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	if validPageMode("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

// ============================================================
// page_cropbox.go — ClearPageCropBox
// ============================================================

func TestCov_ClearPageCropBox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Text("Test")

	// Set then clear
	err := pdf.SetPageCropBox(1, Box{Left: 10, Top: 10, Right: 500, Bottom: 700})
	if err != nil {
		t.Fatalf("SetPageCropBox: %v", err)
	}
	box, err := pdf.GetPageCropBox(1)
	if err != nil || box == nil {
		t.Fatal("expected crop box to be set")
	}

	err = pdf.ClearPageCropBox(1)
	if err != nil {
		t.Fatalf("ClearPageCropBox: %v", err)
	}
	box, err = pdf.GetPageCropBox(1)
	if err != nil {
		t.Fatalf("GetPageCropBox: %v", err)
	}
	if box != nil {
		t.Error("expected nil crop box after clear")
	}
}

func TestCov_ClearPageCropBox_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ClearPageCropBox(99)
	if err == nil {
		t.Error("expected error for invalid page")
	}
}

// ============================================================
// More edge cases for existing functions
// ============================================================

func TestCov_WriteLineEndings_Default(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	obj := annotationObj{
		opt: AnnotationOption{
			Type:             AnnotLine,
			LineEndingStyles: [2]LineEndingStyle{"", ""},
		},
		getRoot: func() *GoPdf { return pdf },
	}
	var buf bytes.Buffer
	obj.writeLineEndings(&buf)
	// Both empty → defaults to None/None → no /LE written
	if strings.Contains(buf.String(), "/LE") {
		t.Error("expected no /LE for None/None")
	}
}

func TestCov_WriteLineEndings_Custom(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	obj := annotationObj{
		opt: AnnotationOption{
			Type:             AnnotLine,
			LineEndingStyles: [2]LineEndingStyle{LineEndOpenArrow, LineEndClosedArrow},
		},
		getRoot: func() *GoPdf { return pdf },
	}
	var buf bytes.Buffer
	obj.writeLineEndings(&buf)
	s := buf.String()
	if !strings.Contains(s, "/LE [/OpenArrow /ClosedArrow]") {
		t.Errorf("expected /LE with arrows, got %q", s)
	}
}

func TestCov_OCGObj_GetType(t *testing.T) {
	obj := ocgObj{}
	if obj.getType() != "OCG" {
		t.Errorf("expected 'OCG', got %q", obj.getType())
	}
}

func TestCov_OCMDObj_GetType(t *testing.T) {
	obj := ocmdObj{}
	if obj.getType() != "OCMD" {
		t.Errorf("expected 'OCMD', got %q", obj.getType())
	}
}

func TestCov_OCPropertiesObj_GetType(t *testing.T) {
	obj := ocPropertiesObj{}
	if obj.getType() != "OCProperties" {
		t.Errorf("expected 'OCProperties', got %q", obj.getType())
	}
}

func TestCov_AnnotationObj_GetType(t *testing.T) {
	obj := annotationObj{}
	if obj.getType() != "Annot" {
		t.Errorf("expected 'Annot', got %q", obj.getType())
	}
}

// ============================================================
// joinContentLines (content_stream_clean.go)
// ============================================================

func TestCov_JoinContentLines(t *testing.T) {
	lines := []string{"q", "1 0 0 1 0 0 cm", "Q"}
	result := joinContentLines(lines)
	expected := "q\n1 0 0 1 0 0 cm\nQ\n"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// ============================================================
// Roundtrip: color conversion consistency
// ============================================================

func TestCov_ColorConversion_Roundtrip_RGB_CMYK_RGB(t *testing.T) {
	// RGB → CMYK → RGB should be approximately the same
	origR, origG, origB := 0.5, 0.3, 0.8
	c, m, y, k := rgbToCMYK(origR, origG, origB)
	r, g, b := cmykToRGB(c, m, y, k)
	if math.Abs(r-origR) > 0.01 || math.Abs(g-origG) > 0.01 || math.Abs(b-origB) > 0.01 {
		t.Errorf("roundtrip failed: orig=(%.2f,%.2f,%.2f) got=(%.2f,%.2f,%.2f)", origR, origG, origB, r, g, b)
	}
}

func TestCov_ColorConversion_GrayToRGBToGray(t *testing.T) {
	gray := 0.5
	result := convertGrayOp(gray, false, ColorspaceRGB)
	// Should produce "0.5000 0.5000 0.5000 rg"
	if !strings.Contains(result, "0.5000") {
		t.Errorf("expected 0.5000 in result, got %q", result)
	}
}
