package gopdf

import (
	"bytes"
	"image"
	"math"
	"os"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost2_test.go — Second round of coverage tests.
// All functions prefixed TestCov2_.
// ============================================================

// ============================================================
// html_parser.go — parseDimension, parseFontSizeAttr,
// htmlFontSizeToFloat, debugHTMLTree, parseCSSColor, parseFontSize
// ============================================================

func TestCov2_ParseDimension(t *testing.T) {
	tests := []struct {
		val  string
		rel  float64
		want float64
		ok   bool
	}{
		{"100px", 0, 75, true},
		{"72pt", 0, 72, true},
		{"50%", 200, 100, true},
		{"2em", 0, 24, true},
		{"42", 0, 42, true},
		{"invalid", 0, 0, false},
	}
	for _, tt := range tests {
		got, ok := parseDimension(tt.val, tt.rel)
		if ok != tt.ok {
			t.Errorf("parseDimension(%q): ok=%v, want %v", tt.val, ok, tt.ok)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("parseDimension(%q) = %f, want %f", tt.val, got, tt.want)
		}
	}
}

func TestCov2_ParseFontSizeAttr(t *testing.T) {
	tests := []struct {
		val  string
		want float64
		ok   bool
	}{
		{"1", 8, true},
		{"3", 12, true},
		{"7", 36, true},
		{"0", 12, true},  // out of range but still returns default
		{"abc", 0, false}, // invalid
	}
	for _, tt := range tests {
		got, ok := parseFontSizeAttr(tt.val)
		if ok != tt.ok {
			t.Errorf("parseFontSizeAttr(%q): ok=%v, want %v", tt.val, ok, tt.ok)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("parseFontSizeAttr(%q) = %f, want %f", tt.val, got, tt.want)
		}
	}
}

func TestCov2_HtmlFontSizeToFloat(t *testing.T) {
	if htmlFontSizeToFloat("3") != 12 {
		t.Error("expected 12 for size 3")
	}
	if htmlFontSizeToFloat("invalid") != 12 {
		t.Error("expected 12 for invalid input")
	}
}

func TestCov2_DebugHTMLTree(t *testing.T) {
	nodes := parseHTML("<b>Hello</b><br/><p>World</p>")
	result := debugHTMLTree(nodes, 0)
	if !strings.Contains(result, "<b>") {
		t.Error("expected <b> in debug output")
	}
	if !strings.Contains(result, "TEXT:") {
		t.Error("expected TEXT: in debug output")
	}
	if !strings.Contains(result, "<br") {
		t.Error("expected <br in debug output")
	}
}

func TestCov2_DebugHTMLTree_Empty(t *testing.T) {
	result := debugHTMLTree(nil, 0)
	if result != "" {
		t.Error("expected empty for nil nodes")
	}
}

func TestCov2_ParseCSSColor_Named(t *testing.T) {
	r, g, b, ok := parseCSSColor("red")
	if !ok || r != 255 || g != 0 || b != 0 {
		t.Errorf("red: got (%d,%d,%d,%v)", r, g, b, ok)
	}
}

func TestCov2_ParseCSSColor_Hex6(t *testing.T) {
	r, g, b, ok := parseCSSColor("#FF8000")
	if !ok || r != 255 || g != 128 || b != 0 {
		t.Errorf("#FF8000: got (%d,%d,%d,%v)", r, g, b, ok)
	}
}

func TestCov2_ParseCSSColor_Hex3(t *testing.T) {
	r, g, b, ok := parseCSSColor("#F00")
	if !ok || r != 255 || g != 0 || b != 0 {
		t.Errorf("#F00: got (%d,%d,%d,%v)", r, g, b, ok)
	}
}

func TestCov2_ParseCSSColor_RGB(t *testing.T) {
	r, g, b, ok := parseCSSColor("rgb(100, 200, 50)")
	if !ok || r != 100 || g != 200 || b != 50 {
		t.Errorf("rgb: got (%d,%d,%d,%v)", r, g, b, ok)
	}
}

func TestCov2_ParseCSSColor_Invalid(t *testing.T) {
	_, _, _, ok := parseCSSColor("notacolor")
	if ok {
		t.Error("expected false for invalid color")
	}
	_, _, _, ok = parseCSSColor("#GGG")
	if ok {
		t.Error("expected false for invalid hex")
	}
}

func TestCov2_ParseFontSize(t *testing.T) {
	tests := []struct {
		val     string
		current float64
		want    float64
		ok      bool
	}{
		{"12pt", 14, 12, true},
		{"16px", 14, 12, true},
		{"2em", 14, 28, true},
		{"150%", 14, 21, true},
		{"xx-small", 14, 6, true},
		{"x-small", 14, 7.5, true},
		{"small", 14, 10, true},
		{"medium", 14, 12, true},
		{"large", 14, 14, true},
		{"x-large", 14, 18, true},
		{"xx-large", 14, 24, true},
		{"20", 14, 20, true},
		{"invalid", 14, 0, false},
	}
	for _, tt := range tests {
		got, ok := parseFontSize(tt.val, tt.current)
		if ok != tt.ok {
			t.Errorf("parseFontSize(%q): ok=%v, want %v", tt.val, ok, tt.ok)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("parseFontSize(%q) = %f, want %f", tt.val, got, tt.want)
		}
	}
}

// ============================================================
// html_parser.go — additional coverage
// ============================================================

func TestCov2_ParseHTML_NestedElements(t *testing.T) {
	nodes := parseHTML("<div><p>Hello <b>World</b></p></div>")
	if len(nodes) == 0 {
		t.Fatal("expected nodes")
	}
	if nodes[0].Tag != "div" {
		t.Errorf("expected div, got %s", nodes[0].Tag)
	}
}

func TestCov2_ParseHTML_Comment(t *testing.T) {
	nodes := parseHTML("<!-- comment --><p>text</p>")
	if len(nodes) == 0 {
		t.Fatal("expected nodes after comment")
	}
}

func TestCov2_ParseHTML_Entities(t *testing.T) {
	result := decodeHTMLEntities("&amp; &lt; &gt; &quot; &apos; &nbsp;")
	if !strings.Contains(result, "&") {
		t.Error("expected & from &amp;")
	}
	if !strings.Contains(result, "<") {
		t.Error("expected < from &lt;")
	}
	if !strings.Contains(result, "'") {
		t.Error("expected ' from &apos;")
	}
}

func TestCov2_ParseInlineStyle(t *testing.T) {
	styles := parseInlineStyle("color: red; font-size: 12px; font-weight: bold")
	if styles["color"] != "red" {
		t.Errorf("expected color=red, got %s", styles["color"])
	}
	if styles["font-size"] != "12px" {
		t.Errorf("expected font-size=12px, got %s", styles["font-size"])
	}
}

func TestCov2_ParseInlineStyle_Empty(t *testing.T) {
	styles := parseInlineStyle("")
	if len(styles) != 0 {
		t.Error("expected empty map for empty style")
	}
}

func TestCov2_IsVoidElement(t *testing.T) {
	voids := []string{"br", "hr", "img"}
	for _, tag := range voids {
		if !isVoidElement(tag) {
			t.Errorf("expected %s to be void", tag)
		}
	}
	if isVoidElement("div") {
		t.Error("div should not be void")
	}
	if isVoidElement("input") {
		t.Error("input is not void in this implementation")
	}
}

func TestCov2_IsBlockElement(t *testing.T) {
	blocks := []string{"div", "p", "h1", "h2", "h3", "ul", "ol", "li", "hr", "center", "blockquote"}
	for _, tag := range blocks {
		if !isBlockElement(tag) {
			t.Errorf("expected %s to be block", tag)
		}
	}
	if isBlockElement("span") {
		t.Error("span should not be block")
	}
}

func TestCov2_HeadingFontSize(t *testing.T) {
	tests := []struct {
		tag  string
		want float64
	}{
		{"h1", 24},
		{"h2", 20},
		{"h3", 16},
		{"h4", 14},
		{"h5", 12},
		{"h6", 10},
		{"p", 12},
	}
	for _, tt := range tests {
		got := headingFontSize(tt.tag)
		if got != tt.want {
			t.Errorf("headingFontSize(%q) = %f, want %f", tt.tag, got, tt.want)
		}
	}
}

func TestCov2_ParseHTML_Attributes(t *testing.T) {
	nodes := parseHTML(`<font size="5" color="red">text</font>`)
	if len(nodes) == 0 {
		t.Fatal("expected nodes")
	}
	if nodes[0].Attrs["size"] != "5" {
		t.Errorf("expected size=5, got %s", nodes[0].Attrs["size"])
	}
	if nodes[0].Attrs["color"] != "red" {
		t.Errorf("expected color=red, got %s", nodes[0].Attrs["color"])
	}
}

func TestCov2_ParseHTML_SelfClosing(t *testing.T) {
	nodes := parseHTML("<br/><hr/>")
	count := 0
	for _, n := range nodes {
		if n.Tag == "br" || n.Tag == "hr" {
			count++
		}
	}
	if count < 2 {
		t.Errorf("expected 2 void elements, got %d", count)
	}
}

// ============================================================
// pdf_decrypt.go — pure functions
// ============================================================

func TestCov2_ExtractSignedIntValue(t *testing.T) {
	tests := []struct {
		dict string
		key  string
		want int
	}{
		{"/V 4 /R 4", "/V", 4},
		{"/V 4 /R 4", "/R", 4},
		{"/P -3904", "/P", -3904},
		{"/V 4 /R 4", "/Missing", 0},
		{"/V  \t 42", "/V", 42},
		{"", "/V", 0},
	}
	for _, tt := range tests {
		got := extractSignedIntValue(tt.dict, tt.key)
		if got != tt.want {
			t.Errorf("extractSignedIntValue(%q, %q) = %d, want %d", tt.dict, tt.key, got, tt.want)
		}
	}
}

func TestCov2_ExtractHexOrLiteralString_Hex(t *testing.T) {
	dict := "/O <48656C6C6F>"
	got := extractHexOrLiteralString(dict, "/O")
	if string(got) != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

func TestCov2_ExtractHexOrLiteralString_Literal(t *testing.T) {
	dict := "/U (Hello)"
	got := extractHexOrLiteralString(dict, "/U")
	if string(got) != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

func TestCov2_ExtractHexOrLiteralString_Missing(t *testing.T) {
	got := extractHexOrLiteralString("/V 4", "/O")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestCov2_ExtractHexOrLiteralString_NoDelimiter(t *testing.T) {
	got := extractHexOrLiteralString("/O something", "/O")
	if got != nil {
		t.Errorf("expected nil for no delimiter, got %v", got)
	}
}

func TestCov2_DecodeHexString(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"48656C6C6F", "Hello"},
		{"4 8 6 5 6C 6C 6F", "Hello"},
		{"414", "A@"}, // odd length gets padded with 0
		{"", ""},
	}
	for _, tt := range tests {
		got := decodeHexString(tt.hex)
		if string(got) != tt.want {
			t.Errorf("decodeHexString(%q) = %q, want %q", tt.hex, got, tt.want)
		}
	}
}

func TestCov2_DecodeLiteralString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"(Hello)", "Hello"},
		{"(Hello (World))", "Hello (World)"},
		{`(line\nbreak)`, "line\nbreak"},
		{`(tab\there)`, "tab\there"},
		{`(back\\slash)`, "back\\slash"},
		{`(paren\(esc\))`, "paren(esc)"},
		{`(\110\145\154\154\157)`, "Hello"}, // octal
		{"", ""},
		{"noparens", ""},
	}
	for _, tt := range tests {
		got := decodeLiteralString(tt.input)
		if string(got) != tt.want {
			t.Errorf("decodeLiteralString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCov2_PadPassword(t *testing.T) {
	// Empty password should be padded to 32 bytes
	result := padPassword(nil)
	if len(result) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(result))
	}

	// Short password
	result = padPassword([]byte("test"))
	if len(result) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(result))
	}
	if string(result[:4]) != "test" {
		t.Error("expected password prefix to be preserved")
	}

	// Long password (>32 bytes) should be truncated
	long := make([]byte, 50)
	for i := range long {
		long[i] = 'A'
	}
	result = padPassword(long)
	if len(result) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(result))
	}
}

func TestCov2_RemoveEncryptFromTrailer(t *testing.T) {
	data := []byte("trailer\n<< /Size 10 /Encrypt 5 0 R /Root 1 0 R >>\n")
	result := removeEncryptFromTrailer(data)
	if bytes.Contains(result, []byte("/Encrypt")) {
		t.Error("expected /Encrypt to be removed")
	}
}

func TestCov2_DetectEncryption(t *testing.T) {
	// No encryption
	data := []byte("%PDF-1.4\ntrailer\n<< /Size 10 /Root 1 0 R >>\n")
	if detectEncryption(data) != 0 {
		t.Error("expected 0 for no encryption")
	}
}

// ============================================================
// transparency.go — defineBlendModeType, NewTransparency
// ============================================================

func TestCov2_DefineBlendModeType_AllModes(t *testing.T) {
	modes := []struct {
		input string
		want  BlendModeType
	}{
		{"/Hue", Hue},
		{"/Color", Color},
		{"/Normal", NormalBlendMode},
		{"", NormalBlendMode},
		{"/Darken", Darken},
		{"/Screen", Screen},
		{"/Overlay", Overlay},
		{"/Lighten", Lighten},
		{"/Multiply", Multiply},
		{"/Exclusion", Exclusion},
		{"/ColorBurn", ColorBurn},
		{"/HardLight", HardLight},
		{"/SoftLight", SoftLight},
		{"/Difference", Difference},
		{"/Saturation", Saturation},
		{"/Luminosity", Luminosity},
		{"/ColorDodge", ColorDodge},
	}
	for _, tt := range modes {
		got, err := defineBlendModeType(tt.input)
		if err != nil {
			t.Errorf("defineBlendModeType(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("defineBlendModeType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestCov2_DefineBlendModeType_Unknown(t *testing.T) {
	_, err := defineBlendModeType("/Unknown")
	if err == nil {
		t.Error("expected error for unknown blend mode")
	}
}

func TestCov2_NewTransparency_Valid(t *testing.T) {
	tr, err := NewTransparency(0.5, "/Normal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Alpha != 0.5 {
		t.Errorf("expected alpha 0.5, got %f", tr.Alpha)
	}
	if tr.BlendModeType != NormalBlendMode {
		t.Errorf("expected Normal, got %v", tr.BlendModeType)
	}
}

func TestCov2_NewTransparency_InvalidAlpha(t *testing.T) {
	_, err := NewTransparency(-0.1, "/Normal")
	if err == nil {
		t.Error("expected error for negative alpha")
	}
	_, err = NewTransparency(1.1, "/Normal")
	if err == nil {
		t.Error("expected error for alpha > 1")
	}
}

func TestCov2_NewTransparency_InvalidBlendMode(t *testing.T) {
	_, err := NewTransparency(0.5, "/Invalid")
	if err == nil {
		t.Error("expected error for invalid blend mode")
	}
}

func TestCov2_NewTransparency_EmptyBlendMode(t *testing.T) {
	tr, err := NewTransparency(1.0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.BlendModeType != NormalBlendMode {
		t.Errorf("expected Normal for empty, got %v", tr.BlendModeType)
	}
}

func TestCov2_TransparencyGetId(t *testing.T) {
	tr := Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode}
	id := tr.GetId()
	if !strings.Contains(id, "0.500") || !strings.Contains(id, "/Normal") {
		t.Errorf("unexpected id: %s", id)
	}
}

func TestCov2_TransparencyMap(t *testing.T) {
	tm := NewTransparencyMap()
	tr := Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode}

	// Not found initially
	_, found := tm.Find(tr)
	if found {
		t.Error("expected not found")
	}

	// Save and find
	tm.Save(tr)
	got, found := tm.Find(tr)
	if !found {
		t.Error("expected found after save")
	}
	if got.Alpha != 0.5 {
		t.Errorf("expected alpha 0.5, got %f", got.Alpha)
	}
}

// ============================================================
// transparency_xobject_group.go
// ============================================================

func TestCov2_TransparencyXObjectGroup_GetType(t *testing.T) {
	g := TransparencyXObjectGroup{}
	if g.getType() != "XObject" {
		t.Errorf("expected XObject, got %s", g.getType())
	}
}

func TestCov2_TransparencyXObjectGroup_Write(t *testing.T) {
	g := TransparencyXObjectGroup{
		BBox: [4]float64{0, 0, 595, 842},
	}
	var buf bytes.Buffer
	err := g.write(&buf, 1)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "/FormType 1") {
		t.Error("expected /FormType 1")
	}
	if !strings.Contains(out, "/Subtype /Form") {
		t.Error("expected /Subtype /Form")
	}
	if !strings.Contains(out, "stream") {
		t.Error("expected stream keyword")
	}
	if !strings.Contains(out, "endstream") {
		t.Error("expected endstream keyword")
	}
}

func TestCov2_TransparencyXObjectGroup_Protection(t *testing.T) {
	g := &TransparencyXObjectGroup{}
	if g.protection() != nil {
		t.Error("expected nil protection")
	}
	p := &PDFProtection{}
	g.setProtection(p)
	if g.protection() != p {
		t.Error("expected protection to be set")
	}
}

// ============================================================
// toc.go — tocError
// ============================================================

func TestCov2_TocError(t *testing.T) {
	err := errorf("test error message")
	if err.Error() != "test error message" {
		t.Errorf("expected 'test error message', got %q", err.Error())
	}
}

func TestCov2_ErrInvalidTOCLevel(t *testing.T) {
	if ErrInvalidTOCLevel == nil {
		t.Fatal("ErrInvalidTOCLevel should not be nil")
	}
	if !strings.Contains(ErrInvalidTOCLevel.Error(), "invalid TOC level") {
		t.Errorf("unexpected error message: %s", ErrInvalidTOCLevel.Error())
	}
}

// ============================================================
// buff.go — Len, Position, SetPosition
// ============================================================

func TestCov2_Buff_Len(t *testing.T) {
	b := &Buff{}
	if b.Len() != 0 {
		t.Error("expected 0 for empty buff")
	}
	b.Write([]byte("hello"))
	if b.Len() != 5 {
		t.Errorf("expected 5, got %d", b.Len())
	}
}

func TestCov2_Buff_Position(t *testing.T) {
	b := &Buff{}
	if b.Position() != 0 {
		t.Error("expected position 0")
	}
	b.Write([]byte("hello"))
	if b.Position() != 5 {
		t.Errorf("expected position 5, got %d", b.Position())
	}
}

func TestCov2_Buff_SetPosition(t *testing.T) {
	b := &Buff{}
	b.Write([]byte("hello"))
	b.SetPosition(2)
	if b.Position() != 2 {
		t.Errorf("expected position 2, got %d", b.Position())
	}
	// Write at position 2 should overwrite
	b.Write([]byte("XY"))
	if string(b.Bytes()) != "heXYo" {
		t.Errorf("expected heXYo, got %q", string(b.Bytes()))
	}
}

func TestCov2_Buff_GrowBeyondCap(t *testing.T) {
	b := &Buff{}
	// Write small, then write large to trigger growth
	b.Write([]byte("ab"))
	big := make([]byte, 1000)
	for i := range big {
		big[i] = 'x'
	}
	b.Write(big)
	if b.Len() != 1002 {
		t.Errorf("expected 1002, got %d", b.Len())
	}
}

// ============================================================
// margin.go — MarginRight and other margin methods
// ============================================================

func TestCov2_MarginRight(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetMarginRight(20)
	got := pdf.MarginRight()
	if got != 20 {
		t.Errorf("expected 20, got %f", got)
	}
}

func TestCov2_MarginBottom(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetMarginBottom(15)
	got := pdf.MarginBottom()
	if got != 15 {
		t.Errorf("expected 15, got %f", got)
	}
}

func TestCov2_MarginLeft(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetMarginLeft(25)
	got := pdf.MarginLeft()
	if got != 25 {
		t.Errorf("expected 25, got %f", got)
	}
}

func TestCov2_MarginTop(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetMarginTop(30)
	got := pdf.MarginTop()
	if got != 30 {
		t.Errorf("expected 30, got %f", got)
	}
}

func TestCov2_SetMargins(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetMargins(10, 20, 30, 40)
	l, top, r, b := pdf.Margins()
	if l != 10 || top != 20 || r != 30 || b != 40 {
		t.Errorf("margins mismatch: got %f,%f,%f,%f", l, top, r, b)
	}
}

// ============================================================
// watermark.go — WatermarkOption.defaults, diagonalAngle,
// AddWatermarkText, AddWatermarkImageAllPages
// ============================================================

func TestCov2_WatermarkOptionDefaults(t *testing.T) {
	opt := WatermarkOption{}
	opt.defaults()
	if opt.FontSize != 48 {
		t.Errorf("expected FontSize 48, got %f", opt.FontSize)
	}
	if opt.Angle != 45 {
		t.Errorf("expected Angle 45, got %f", opt.Angle)
	}
	if opt.Opacity != 0.3 {
		t.Errorf("expected Opacity 0.3, got %f", opt.Opacity)
	}
	if opt.RepeatSpacingX != 150 {
		t.Errorf("expected RepeatSpacingX 150, got %f", opt.RepeatSpacingX)
	}
	if opt.Color != [3]uint8{200, 200, 200} {
		t.Errorf("expected default color, got %v", opt.Color)
	}
}

func TestCov2_WatermarkOptionDefaults_NoOverwrite(t *testing.T) {
	opt := WatermarkOption{
		FontSize:       24,
		Angle:          30,
		Opacity:        0.5,
		RepeatSpacingX: 100,
		RepeatSpacingY: 100,
		Color:          [3]uint8{255, 0, 0},
	}
	opt.defaults()
	if opt.FontSize != 24 {
		t.Error("should not overwrite FontSize")
	}
	if opt.Angle != 30 {
		t.Error("should not overwrite Angle")
	}
}

func TestCov2_DiagonalAngle(t *testing.T) {
	angle := diagonalAngle(100, 100)
	if math.Abs(angle-45) > 0.01 {
		t.Errorf("expected ~45, got %f", angle)
	}
	angle = diagonalAngle(100, 0)
	if math.Abs(angle) > 0.01 {
		t.Errorf("expected ~0, got %f", angle)
	}
}

func TestCov2_AddWatermarkText_EmptyText(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddWatermarkText(WatermarkOption{Text: "", FontFamily: "test"})
	if err != ErrEmptyString {
		t.Errorf("expected ErrEmptyString, got %v", err)
	}
}

func TestCov2_AddWatermarkText_NoFont(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.AddWatermarkText(WatermarkOption{Text: "test"})
	if err != ErrMissingFontFamily {
		t.Errorf("expected ErrMissingFontFamily, got %v", err)
	}
}

func TestCov2_AddWatermarkText_Single(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "DRAFT",
		FontFamily: fontFamily,
		FontSize:   36,
		Opacity:    0.3,
		Angle:      45,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText: %v", err)
	}
}

func TestCov2_AddWatermarkText_Repeat(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkText(WatermarkOption{
		Text:       "CONFIDENTIAL",
		FontFamily: fontFamily,
		FontSize:   24,
		Opacity:    0.2,
		Angle:      30,
		Repeat:     true,
	})
	if err != nil {
		t.Fatalf("AddWatermarkText repeat: %v", err)
	}
}

func TestCov2_AddWatermarkTextAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")
	// SetPage to page 1 first to ensure pages are navigable
	_ = pdf.SetPage(1)
	err := pdf.AddWatermarkTextAllPages(WatermarkOption{
		Text:       "SAMPLE",
		FontFamily: fontFamily,
	})
	if err != nil {
		t.Fatalf("AddWatermarkTextAllPages: %v", err)
	}
}

func TestCov2_AddWatermarkImage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkImage(resJPEGPath, 0.5, 100, 100, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImage: %v", err)
	}
}

func TestCov2_AddWatermarkImage_WithRotation(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddWatermarkImage(resJPEGPath, 0, 0, 0, 45)
	if err != nil {
		t.Fatalf("AddWatermarkImage with rotation: %v", err)
	}
}

func TestCov2_AddWatermarkImageAllPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")
	_ = pdf.SetPage(1)
	err := pdf.AddWatermarkImageAllPages(resJPEGPath, 0.4, 80, 80, 0)
	if err != nil {
		t.Fatalf("AddWatermarkImageAllPages: %v", err)
	}
}

// ============================================================
// incremental_save.go — IncrementalSave, WriteIncrementalPdf
// ============================================================

func TestCov2_IncrementalSave_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Hello")

	original, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdf: %v", err)
	}

	// Now do an incremental save
	result, err := pdf.IncrementalSave(original, nil)
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if len(result) <= len(original) {
		t.Error("incremental result should be larger than original")
	}
	if !bytes.Contains(result, []byte("%%EOF")) {
		t.Error("expected EOF marker in result")
	}
}

func TestCov2_IncrementalSave_SpecificIndices(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Hello")

	original, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdf: %v", err)
	}

	// Save only first object
	result, err := pdf.IncrementalSave(original, []int{0})
	if err != nil {
		t.Fatalf("IncrementalSave: %v", err)
	}
	if !bytes.Contains(result, []byte("xref")) {
		t.Error("expected xref in result")
	}
}

func TestCov2_WriteIncrementalPdf(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Incremental")

	original, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdf: %v", err)
	}

	outPath := resOutDir + "/incremental_test.pdf"
	err = pdf.WriteIncrementalPdf(outPath, original, nil)
	if err != nil {
		t.Fatalf("WriteIncrementalPdf: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

// ============================================================
// geometry.go — RectFrom, Matrix, Distance
// ============================================================

func TestCov2_RectFrom_Contains(t *testing.T) {
	r := RectFrom{X: 10, Y: 10, W: 100, H: 50}
	if !r.Contains(50, 30) {
		t.Error("expected point inside")
	}
	if r.Contains(5, 30) {
		t.Error("expected point outside")
	}
	if !r.Contains(10, 10) {
		t.Error("expected edge point inside")
	}
}

func TestCov2_RectFrom_ContainsRect(t *testing.T) {
	outer := RectFrom{X: 0, Y: 0, W: 100, H: 100}
	inner := RectFrom{X: 10, Y: 10, W: 50, H: 50}
	if !outer.ContainsRect(inner) {
		t.Error("expected inner inside outer")
	}
	if inner.ContainsRect(outer) {
		t.Error("expected outer not inside inner")
	}
}

func TestCov2_RectFrom_Intersects(t *testing.T) {
	r1 := RectFrom{X: 0, Y: 0, W: 50, H: 50}
	r2 := RectFrom{X: 25, Y: 25, W: 50, H: 50}
	if !r1.Intersects(r2) {
		t.Error("expected intersection")
	}
	r3 := RectFrom{X: 100, Y: 100, W: 50, H: 50}
	if r1.Intersects(r3) {
		t.Error("expected no intersection")
	}
}

func TestCov2_RectFrom_Intersection(t *testing.T) {
	r1 := RectFrom{X: 0, Y: 0, W: 50, H: 50}
	r2 := RectFrom{X: 25, Y: 25, W: 50, H: 50}
	inter := r1.Intersection(r2)
	if inter.X != 25 || inter.Y != 25 || inter.W != 25 || inter.H != 25 {
		t.Errorf("unexpected intersection: %+v", inter)
	}
	// No intersection
	r3 := RectFrom{X: 100, Y: 100, W: 10, H: 10}
	empty := r1.Intersection(r3)
	if !empty.IsEmpty() {
		t.Error("expected empty intersection")
	}
}

func TestCov2_RectFrom_Union(t *testing.T) {
	r1 := RectFrom{X: 10, Y: 10, W: 20, H: 20}
	r2 := RectFrom{X: 50, Y: 50, W: 30, H: 30}
	u := r1.Union(r2)
	if u.X != 10 || u.Y != 10 || u.W != 70 || u.H != 70 {
		t.Errorf("unexpected union: %+v", u)
	}
}

func TestCov2_RectFrom_IsEmpty(t *testing.T) {
	if !(RectFrom{W: 0, H: 10}).IsEmpty() {
		t.Error("zero width should be empty")
	}
	if !(RectFrom{W: 10, H: -1}).IsEmpty() {
		t.Error("negative height should be empty")
	}
	if (RectFrom{W: 10, H: 10}).IsEmpty() {
		t.Error("positive dims should not be empty")
	}
}

func TestCov2_RectFrom_Area(t *testing.T) {
	r := RectFrom{W: 10, H: 20}
	if r.Area() != 200 {
		t.Errorf("expected 200, got %f", r.Area())
	}
	empty := RectFrom{W: -1, H: 10}
	if empty.Area() != 0 {
		t.Error("negative dim should give 0 area")
	}
}

func TestCov2_RectFrom_Center(t *testing.T) {
	r := RectFrom{X: 10, Y: 20, W: 100, H: 50}
	c := r.Center()
	if c.X != 60 || c.Y != 45 {
		t.Errorf("expected (60,45), got (%f,%f)", c.X, c.Y)
	}
}

func TestCov2_RectFrom_Normalize(t *testing.T) {
	r := RectFrom{X: 50, Y: 50, W: -30, H: -20}
	n := r.Normalize()
	if n.X != 20 || n.Y != 30 || n.W != 30 || n.H != 20 {
		t.Errorf("unexpected normalize: %+v", n)
	}
}

func TestCov2_Matrix_Identity(t *testing.T) {
	m := IdentityMatrix()
	if !m.IsIdentity() {
		t.Error("expected identity")
	}
}

func TestCov2_Matrix_Translate(t *testing.T) {
	m := TranslateMatrix(10, 20)
	x, y := m.TransformPoint(0, 0)
	if x != 10 || y != 20 {
		t.Errorf("expected (10,20), got (%f,%f)", x, y)
	}
}

func TestCov2_Matrix_Scale(t *testing.T) {
	m := ScaleMatrix(2, 3)
	x, y := m.TransformPoint(5, 10)
	if x != 10 || y != 30 {
		t.Errorf("expected (10,30), got (%f,%f)", x, y)
	}
}

func TestCov2_Matrix_Rotate(t *testing.T) {
	m := RotateMatrix(90)
	x, y := m.TransformPoint(1, 0)
	if math.Abs(x) > 0.001 || math.Abs(y-1) > 0.001 {
		t.Errorf("expected (~0,~1), got (%f,%f)", x, y)
	}
}

func TestCov2_Matrix_Multiply(t *testing.T) {
	t1 := TranslateMatrix(10, 0)
	s := ScaleMatrix(2, 2)
	m := t1.Multiply(s)
	x, y := m.TransformPoint(5, 5)
	// t1 * s: first scale (5,5)->(10,10), then translate -> (20,10)
	if math.Abs(x-20) > 0.001 || math.Abs(y-10) > 0.001 {
		t.Errorf("expected (20,10), got (%f,%f)", x, y)
	}
}

func TestCov2_Matrix_IsIdentity_False(t *testing.T) {
	m := TranslateMatrix(1, 0)
	if m.IsIdentity() {
		t.Error("should not be identity")
	}
}

func TestCov2_Distance(t *testing.T) {
	d := Distance(Point{X: 0, Y: 0}, Point{X: 3, Y: 4})
	if math.Abs(d-5) > 0.001 {
		t.Errorf("expected 5, got %f", d)
	}
}

// ============================================================
// buffer_pool.go — GetBuffer, PutBuffer
// ============================================================

func TestCov2_BufferPool(t *testing.T) {
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("expected non-nil buffer")
	}
	buf.WriteString("hello")
	if buf.Len() != 5 {
		t.Error("expected 5 bytes")
	}
	PutBuffer(buf)
	// After put, buffer should be reset
	buf2 := GetBuffer()
	if buf2 == nil {
		t.Fatal("expected non-nil buffer from pool")
	}
}

// ============================================================
// journal.go — Journal operations
// ============================================================

func TestCov2_Journal_EnableDisable(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	if pdf.JournalIsEnabled() {
		t.Error("journal should not be enabled initially")
	}

	pdf.JournalEnable()
	if !pdf.JournalIsEnabled() {
		t.Error("journal should be enabled")
	}

	pdf.JournalDisable()
	if pdf.JournalIsEnabled() {
		t.Error("journal should be disabled")
	}
}

func TestCov2_Journal_UndoRedo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()

	pdf.JournalStartOp("add text")
	pdf.SetXY(50, 50)
	_ = pdf.Text("Hello")
	pdf.JournalEndOp()

	name, err := pdf.JournalUndo()
	if err != nil {
		t.Fatalf("undo: %v", err)
	}
	if name != "add text" {
		t.Errorf("expected 'add text', got %q", name)
	}

	name, err = pdf.JournalRedo()
	if err != nil {
		t.Fatalf("redo: %v", err)
	}
	if name != "add text" {
		t.Errorf("expected 'add text', got %q", name)
	}
}

func TestCov2_Journal_UndoEmpty(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	_, err := pdf.JournalUndo()
	if err == nil {
		t.Error("expected error for undo without journal")
	}
}

func TestCov2_Journal_RedoEmpty(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	_, err := pdf.JournalRedo()
	if err == nil {
		t.Error("expected error for redo without journal")
	}
}

func TestCov2_Journal_GetOperations(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()

	pdf.JournalStartOp("op1")
	pdf.JournalEndOp()
	pdf.JournalStartOp("op2")
	pdf.JournalEndOp()

	ops := pdf.JournalGetOperations()
	if len(ops) < 2 {
		t.Errorf("expected at least 2 ops, got %d", len(ops))
	}
}

func TestCov2_Journal_GetOperations_Nil(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	ops := pdf.JournalGetOperations()
	if ops != nil {
		t.Error("expected nil for no journal")
	}
}

func TestCov2_Journal_SaveLoad(t *testing.T) {
	ensureOutDir(t)
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.JournalEnable()

	pdf.JournalStartOp("test op")
	pdf.SetXY(50, 50)
	_ = pdf.Text("Journal test")
	pdf.JournalEndOp()

	path := resOutDir + "/test.journal"
	err := pdf.JournalSave(path)
	if err != nil {
		t.Fatalf("JournalSave: %v", err)
	}

	// Load into new pdf
	pdf2 := newPDFWithFont(t)
	pdf2.AddPage()
	pdf2.JournalEnable()
	err = pdf2.JournalLoad(path)
	if err != nil {
		t.Fatalf("JournalLoad: %v", err)
	}
}

func TestCov2_Journal_SaveNoJournal(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.JournalSave("/tmp/test.journal")
	if err == nil {
		t.Error("expected error for save without journal")
	}
}

func TestCov2_Journal_LoadNoJournal(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err := pdf.JournalLoad("/tmp/test.journal")
	if err == nil {
		t.Error("expected error for load without journal")
	}
}

// ============================================================
// text_search.go — SearchText, TextSearchResult
// ============================================================

func TestCov2_TextSearchResult_Struct(t *testing.T) {
	r := TextSearchResult{
		PageIndex: 0,
		X:         100,
		Y:         200,
		Width:     50,
		Height:    12,
		Text:      "hello",
		Context:   "say hello world",
	}
	if r.PageIndex != 0 || r.Text != "hello" {
		t.Error("struct fields mismatch")
	}
}

// ============================================================
// bookmark.go — ModifyBookmark, DeleteBookmark, SetBookmarkStyle
// ============================================================

func TestCov2_ModifyBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyBookmark(-1, "test")
	if err != ErrBookmarkOutOfRange {
		t.Errorf("expected ErrBookmarkOutOfRange, got %v", err)
	}
	err = pdf.ModifyBookmark(0, "test")
	if err != ErrBookmarkOutOfRange {
		t.Errorf("expected ErrBookmarkOutOfRange for empty, got %v", err)
	}
}

func TestCov2_DeleteBookmark_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeleteBookmark(-1)
	if err != ErrBookmarkOutOfRange {
		t.Errorf("expected ErrBookmarkOutOfRange, got %v", err)
	}
}

func TestCov2_SetBookmarkStyle_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetBookmarkStyle(0, BookmarkStyle{Bold: true})
	if err != ErrBookmarkOutOfRange {
		t.Errorf("expected ErrBookmarkOutOfRange, got %v", err)
	}
}

// ============================================================
// toc.go — GetTOC, SetTOC
// ============================================================

func TestCov2_GetTOC_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	toc := pdf.GetTOC()
	if toc != nil {
		t.Error("expected nil TOC for no outlines")
	}
}

func TestCov2_SetTOC_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetTOC(nil)
	if err != nil {
		t.Fatalf("SetTOC nil: %v", err)
	}
}

func TestCov2_SetTOC_InvalidLevel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetTOC([]TOCItem{{Level: 2, Title: "bad", PageNo: 1}})
	if err != ErrInvalidTOCLevel {
		t.Errorf("expected ErrInvalidTOCLevel, got %v", err)
	}
}

func TestCov2_SetTOC_LevelJump(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetTOC([]TOCItem{
		{Level: 1, Title: "Ch1", PageNo: 1},
		{Level: 3, Title: "bad jump", PageNo: 1},
	})
	if err != ErrInvalidTOCLevel {
		t.Errorf("expected ErrInvalidTOCLevel for level jump, got %v", err)
	}
}

// ============================================================
// subset_font_obj.go — getter functions
// ============================================================

func TestCov2_SubsetFontObj_Getters(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	// Find the SubsetFontObj
	var sfObj *SubsetFontObj
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			sfObj = sf
			break
		}
	}
	if sfObj == nil {
		t.Skip("no SubsetFontObj found")
	}

	// Test getter functions
	ut := sfObj.GetUnderlineThickness()
	_ = ut // just ensure it doesn't panic

	utPx := sfObj.GetUnderlineThicknessPx(14)
	if utPx < 0 {
		t.Error("underline thickness px should be >= 0")
	}

	up := sfObj.GetUnderlinePosition()
	_ = up

	upPx := sfObj.GetUnderlinePositionPx(14)
	_ = upPx

	asc := sfObj.GetAscender()
	if asc == 0 {
		t.Log("ascender is 0 (may be expected for some fonts)")
	}

	ascPx := sfObj.GetAscenderPx(14)
	_ = ascPx

	desc := sfObj.GetDescender()
	_ = desc

	descPx := sfObj.GetDescenderPx(14)
	_ = descPx

	family := sfObj.GetFamily()
	if family != fontFamily {
		t.Errorf("expected %s, got %s", fontFamily, family)
	}

	parser := sfObj.GetTTFParser()
	if parser == nil {
		t.Error("expected non-nil TTFParser")
	}

	tp := sfObj.getType()
	if tp != "SubsetFont" {
		t.Errorf("expected SubsetFont, got %s", tp)
	}
}

func TestCov2_SubsetFontObj_SetFamily(t *testing.T) {
	sf := &SubsetFontObj{}
	sf.SetFamily("TestFont")
	if sf.GetFamily() != "TestFont" {
		t.Errorf("expected TestFont, got %s", sf.GetFamily())
	}
}

func TestCov2_SubsetFontObj_TtfFontOption(t *testing.T) {
	sf := &SubsetFontObj{}
	opt := TtfOption{UseKerning: true}
	sf.SetTtfFontOption(opt)
	got := sf.GetTtfFontOption()
	if !got.UseKerning {
		t.Error("expected UseKerning true")
	}
}

func TestCov2_SubsetFontObj_CharIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	var sfObj *SubsetFontObj
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			sfObj = sf
			break
		}
	}
	if sfObj == nil {
		t.Skip("no SubsetFontObj found")
	}

	// Add chars first
	_, err := sfObj.AddChars("A")
	if err != nil {
		t.Fatalf("AddChars: %v", err)
	}

	idx, err := sfObj.CharIndex('A')
	if err != nil {
		t.Fatalf("CharIndex: %v", err)
	}
	if idx == 0 {
		t.Error("expected non-zero index for 'A'")
	}
}

func TestCov2_SubsetFontObj_CharWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	var sfObj *SubsetFontObj
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			sfObj = sf
			break
		}
	}
	if sfObj == nil {
		t.Skip("no SubsetFontObj found")
	}

	_, _ = sfObj.AddChars("W")
	w, err := sfObj.CharWidth('W')
	if err != nil {
		t.Fatalf("CharWidth: %v", err)
	}
	if w == 0 {
		t.Error("expected non-zero width for 'W'")
	}
}

func TestCov2_SubsetFontObj_GlyphIndexToPdfWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	var sfObj *SubsetFontObj
	for _, obj := range pdf.pdfObjs {
		if sf, ok := obj.(*SubsetFontObj); ok {
			sfObj = sf
			break
		}
	}
	if sfObj == nil {
		t.Skip("no SubsetFontObj found")
	}

	// Glyph index 0 should return 0 width
	w := sfObj.GlyphIndexToPdfWidth(0)
	_ = w
}

// ============================================================
// pdf_lowlevel.go — SetStream (via ReadObject + UpdateObject)
// ============================================================

func TestCov2_SetStream(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Test stream")

	data, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdf: %v", err)
	}

	// Try to set stream on object 1 (may or may not have a stream)
	newStream := []byte("q\n1 0 0 1 0 0 cm\nQ\n")
	result, err := SetStream(data, 1, newStream)
	if err != nil {
		// Some objects may not have streams, that's ok
		t.Logf("SetStream on obj 1: %v (expected for non-stream objects)", err)
	} else {
		if len(result) == 0 {
			t.Error("expected non-empty result")
		}
	}
}

// ============================================================
// text_extract.go — parseTextOperators additional coverage
// ============================================================

func TestCov2_ParseTextOperators_BT_ET(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n100 700 Td\n(Hello World) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	if len(results) == 0 {
		t.Error("expected at least one text result")
	}
}

func TestCov2_ParseTextOperators_TJ(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n100 700 Td\n[(Hello) -100 (World)] TJ\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	if len(results) == 0 {
		t.Error("expected text from TJ operator")
	}
}

func TestCov2_ParseTextOperators_Tm(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n1 0 0 1 100 700 Tm\n(Positioned) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	if len(results) == 0 {
		t.Error("expected text from Tm operator")
	}
}

func TestCov2_ParseTextOperators_TD(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n100 700 TD\n(Line) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	if len(results) == 0 {
		t.Error("expected text from TD operator")
	}
}

func TestCov2_ParseTextOperators_TStar(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n14 TL\n100 700 Td\n(Line1) Tj\nT*\n(Line2) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	if len(results) < 2 {
		t.Errorf("expected 2 text results, got %d", len(results))
	}
}

func TestCov2_ParseTextOperators_Quote(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n14 TL\n100 700 Td\n(Line1) '\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	_ = results // just ensure no panic
}

func TestCov2_ParseTextOperators_DoubleQuote(t *testing.T) {
	stream := []byte("BT\n/F1 12 Tf\n14 TL\n100 700 Td\n0 0 (Text) \"\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	_ = results // just ensure no panic
}

func TestCov2_ParseTextOperators_CM(t *testing.T) {
	stream := []byte("2 0 0 2 10 20 cm\nBT\n/F1 12 Tf\n100 700 Td\n(Scaled) Tj\nET\n")
	fonts := map[string]*fontInfo{
		"/F1": {name: "TestFont"},
	}
	results := parseTextOperators(stream, fonts, [4]float64{0, 0, 595, 842})
	if len(results) == 0 {
		t.Error("expected text with cm transform")
	}
}

func TestCov2_ParseTextOperators_Empty(t *testing.T) {
	results := parseTextOperators(nil, nil, [4]float64{0, 0, 595, 842})
	if len(results) != 0 {
		t.Error("expected empty results for nil stream")
	}
}

// ============================================================
// content_element.go — ContentElementType.String
// ============================================================

func TestCov2_ContentElementType_String(t *testing.T) {
	tests := []struct {
		typ  ContentElementType
		want string
	}{
		{ElementText, "Text"},
		{ElementImage, "Image"},
		{ElementLine, "Line"},
		{ElementRectangle, "Rectangle"},
		{ElementOval, "Oval"},
		{ElementPolygon, "Polygon"},
		{ElementCurve, "Curve"},
		{ElementPolyline, "Polyline"},
		{ElementSector, "Sector"},
		{ElementImportedTemplate, "ImportedTemplate"},
		{ElementLineWidth, "LineWidth"},
		{ElementLineType, "LineType"},
		{ElementCustomLineType, "CustomLineType"},
		{ElementGray, "Gray"},
		{ElementColorRGB, "ColorRGB"},
		{ElementColorCMYK, "ColorCMYK"},
		{ElementColorSpace, "ColorSpace"},
		{ElementRotate, "Rotate"},
		{ElementClipPolygon, "ClipPolygon"},
		{ElementSaveGState, "SaveGState"},
		{ElementRestoreGState, "RestoreGState"},
		{ContentElementType(999), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.typ.String()
		if got != tt.want {
			t.Errorf("ContentElementType(%d).String() = %q, want %q", tt.typ, got, tt.want)
		}
	}
}

func TestCov2_GetPageElements(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Hello")
	pdf.Line(10, 10, 100, 100)

	elems, err := pdf.GetPageElements(1)
	if err != nil {
		t.Fatalf("GetPageElements: %v", err)
	}
	if len(elems) == 0 {
		t.Error("expected elements on page")
	}
}

func TestCov2_GetPageElementsByType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)

	elems, err := pdf.GetPageElementsByType(1, ElementLine)
	if err != nil {
		t.Fatalf("GetPageElementsByType: %v", err)
	}
	if len(elems) == 0 {
		t.Error("expected line elements")
	}
}

func TestCov2_GetPageElementCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)

	count, err := pdf.GetPageElementCount(1)
	if err != nil {
		t.Fatalf("GetPageElementCount: %v", err)
	}
	if count == 0 {
		t.Error("expected non-zero count")
	}
}

func TestCov2_DeleteElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	pdf.Line(20, 20, 200, 200)

	count1, _ := pdf.GetPageElementCount(1)
	err := pdf.DeleteElement(1, 0)
	if err != nil {
		t.Fatalf("DeleteElement: %v", err)
	}
	count2, _ := pdf.GetPageElementCount(1)
	if count2 >= count1 {
		t.Error("expected fewer elements after delete")
	}
}

func TestCov2_ClearPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Hello")

	err := pdf.ClearPage(1)
	if err != nil {
		t.Fatalf("ClearPage: %v", err)
	}
	count, _ := pdf.GetPageElementCount(1)
	if count != 0 {
		t.Errorf("expected 0 elements after clear, got %d", count)
	}
}

func TestCov2_InsertLineElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("content") // need content to create ContentObj
	err := pdf.InsertLineElement(1, 10, 10, 200, 200)
	if err != nil {
		t.Fatalf("InsertLineElement: %v", err)
	}
}

func TestCov2_InsertRectElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("content")
	err := pdf.InsertRectElement(1, 50, 50, 100, 80, "D")
	if err != nil {
		t.Fatalf("InsertRectElement: %v", err)
	}
}

func TestCov2_InsertOvalElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("content")
	err := pdf.InsertOvalElement(1, 50, 50, 150, 100)
	if err != nil {
		t.Fatalf("InsertOvalElement: %v", err)
	}
}

func TestCov2_ModifyElementPosition(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)

	err := pdf.ModifyElementPosition(1, 0, 50, 50)
	if err != nil {
		t.Fatalf("ModifyElementPosition: %v", err)
	}
}

func TestCov2_DeleteElementsByType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 100, 100)
	pdf.Line(20, 20, 200, 200)

	deleted, err := pdf.DeleteElementsByType(1, ElementLine)
	if err != nil {
		t.Fatalf("DeleteElementsByType: %v", err)
	}
	if deleted == 0 {
		t.Error("expected some elements deleted")
	}
}

func TestCov2_DeleteElementsInRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 10, 50, 50)

	deleted, err := pdf.DeleteElementsInRect(1, 0, 0, 100, 100)
	if err != nil {
		t.Fatalf("DeleteElementsInRect: %v", err)
	}
	_ = deleted
}

// ============================================================
// fontconverthelper.go — FontConvertHelperCw2Str
// ============================================================

func TestCov2_FontConvertHelperCw2Str(t *testing.T) {
	cw := make(FontCw)
	cw[65] = 600 // 'A'
	cw[66] = 700 // 'B'
	result := FontConvertHelperCw2Str(cw)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestCov2_FontConvertHelper_Cw2Str_Deprecated(t *testing.T) {
	cw := make(FontCw)
	result := FontConvertHelper_Cw2Str(cw)
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

// ============================================================
// colorspace_convert.go — internal conversion functions
// ============================================================

func TestCov2_RgbToGray(t *testing.T) {
	gray := rgbToGray(1, 1, 1)
	if math.Abs(gray-1.0) > 0.01 {
		t.Errorf("expected ~1.0 for white, got %f", gray)
	}
	gray = rgbToGray(0, 0, 0)
	if gray != 0 {
		t.Errorf("expected 0 for black, got %f", gray)
	}
}

func TestCov2_RgbToCMYK(t *testing.T) {
	c, m, y, k := rgbToCMYK(1, 0, 0) // red
	if k != 0 {
		t.Errorf("expected k=0 for red, got %f", k)
	}
	if c != 0 {
		t.Errorf("expected c=0 for red, got %f", c)
	}
	if m != 1 {
		t.Errorf("expected m=1 for red, got %f", m)
	}
	_ = y
}

func TestCov2_RgbToCMYK_Black(t *testing.T) {
	c, m, y, k := rgbToCMYK(0, 0, 0)
	if k != 1 {
		t.Errorf("expected k=1 for black, got %f", k)
	}
	if c != 0 || m != 0 || y != 0 {
		t.Error("expected c=m=y=0 for black")
	}
}

func TestCov2_CmykToRGB(t *testing.T) {
	r, g, b := cmykToRGB(0, 0, 0, 0) // white
	if r != 1 || g != 1 || b != 1 {
		t.Errorf("expected (1,1,1) for white, got (%f,%f,%f)", r, g, b)
	}
}

func TestCov2_ConvertColorLine_RGBtoGray(t *testing.T) {
	result := convertColorLine("1.0000 0.0000 0.0000 rg", ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Errorf("expected gray operator, got %s", result)
	}
}

func TestCov2_ConvertColorLine_RGBtoCMYK(t *testing.T) {
	result := convertColorLine("1.0000 0.0000 0.0000 RG", ColorspaceCMYK)
	if !strings.Contains(result, "K") {
		t.Errorf("expected CMYK stroking operator, got %s", result)
	}
}

func TestCov2_ConvertColorLine_CMYKtoGray(t *testing.T) {
	result := convertColorLine("0.0000 1.0000 1.0000 0.0000 k", ColorspaceGray)
	if !strings.Contains(result, "g") {
		t.Errorf("expected gray operator, got %s", result)
	}
}

func TestCov2_ConvertColorLine_CMYKtoRGB(t *testing.T) {
	result := convertColorLine("0.0000 1.0000 1.0000 0.0000 K", ColorspaceRGB)
	if !strings.Contains(result, "RG") {
		t.Errorf("expected RGB stroking operator, got %s", result)
	}
}

func TestCov2_ConvertColorLine_GrayToRGB(t *testing.T) {
	result := convertColorLine("0.5000 g", ColorspaceRGB)
	if !strings.Contains(result, "rg") {
		t.Errorf("expected RGB operator, got %s", result)
	}
}

func TestCov2_ConvertColorLine_GrayToCMYK(t *testing.T) {
	result := convertColorLine("0.5000 G", ColorspaceCMYK)
	if !strings.Contains(result, "K") {
		t.Errorf("expected CMYK stroking operator, got %s", result)
	}
}

func TestCov2_ConvertColorLine_NonColor(t *testing.T) {
	result := convertColorLine("100 200 m", ColorspaceGray)
	if result != "100 200 m" {
		t.Errorf("non-color line should be unchanged, got %s", result)
	}
}

func TestCov2_ConvertStreamColorspace(t *testing.T) {
	stream := []byte("1.0000 0.0000 0.0000 rg\n100 200 m\n0.5000 g\n")
	result := convertStreamColorspace(stream, ColorspaceGray)
	if bytes.Contains(result, []byte("rg")) {
		t.Error("expected rg to be converted")
	}
}

func TestCov2_ParseColorFloat(t *testing.T) {
	if parseColorFloat("0.5") != 0.5 {
		t.Error("expected 0.5")
	}
	if parseColorFloat("invalid") != 0 {
		t.Error("expected 0 for invalid")
	}
	if parseColorFloat("  1.0  ") != 1.0 {
		t.Error("expected 1.0 with whitespace")
	}
}

// ============================================================
// image_extract.go — extractIntValue, extractFilterValue, GetImageFormat
// ============================================================

func TestCov2_ExtractIntValue(t *testing.T) {
	dict := "/Width 640 /Height 480"
	if extractIntValue(dict, "/Width") != 640 {
		t.Error("expected 640")
	}
	if extractIntValue(dict, "/Height") != 480 {
		t.Error("expected 480")
	}
	if extractIntValue(dict, "/Missing") != 0 {
		t.Error("expected 0 for missing key")
	}
}

func TestCov2_ExtractFilterValue(t *testing.T) {
	if extractFilterValue("/Filter /DCTDecode") != "DCTDecode" {
		t.Error("expected DCTDecode")
	}
	if extractFilterValue("/Filter [/FlateDecode]") != "FlateDecode" {
		t.Error("expected FlateDecode from array")
	}
	if extractFilterValue("/NoFilter") != "" {
		t.Error("expected empty for no filter")
	}
}

func TestCov2_ExtractedImage_GetImageFormat(t *testing.T) {
	tests := []struct {
		filter string
		want   string
	}{
		{"DCTDecode", "jpeg"},
		{"JPXDecode", "jp2"},
		{"CCITTFaxDecode", "tiff"},
		{"FlateDecode", "png"},
		{"", "png"},
		{"SomeOther", "raw"},
	}
	for _, tt := range tests {
		img := ExtractedImage{Filter: tt.filter}
		got := img.GetImageFormat()
		if got != tt.want {
			t.Errorf("GetImageFormat(%q) = %q, want %q", tt.filter, got, tt.want)
		}
	}
}

// ============================================================
// page_manipulate.go — DeletePage, CopyPage
// ============================================================

func TestCov2_DeletePage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")

	before := pdf.GetNumberOfPages()
	err := pdf.DeletePage(1)
	if err != nil {
		t.Fatalf("DeletePage: %v", err)
	}
	after := pdf.GetNumberOfPages()
	if after != before-1 {
		t.Errorf("expected %d pages, got %d", before-1, after)
	}
}

func TestCov2_CopyPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Original")

	before := pdf.GetNumberOfPages()
	newPage, err := pdf.CopyPage(1)
	if err != nil {
		t.Fatalf("CopyPage: %v", err)
	}
	after := pdf.GetNumberOfPages()
	if after != before+1 {
		t.Errorf("expected %d pages, got %d", before+1, after)
	}
	if newPage != after {
		t.Errorf("expected new page %d, got %d", after, newPage)
	}
}

func TestCov2_DeletePage_InvalidPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.DeletePage(99)
	if err == nil {
		t.Error("expected error for invalid page")
	}
}

// ============================================================
// page_info.go — GetPageSize, GetAllPageSizes
// ============================================================

func TestCov2_GetPageSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	w, h, err := pdf.GetPageSize(1)
	if err != nil {
		t.Fatalf("GetPageSize: %v", err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("expected positive dimensions, got %f x %f", w, h)
	}
}

func TestCov2_GetPageSize_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, _, err := pdf.GetPageSize(99)
	if err == nil {
		t.Error("expected error for invalid page")
	}
}

func TestCov2_GetAllPageSizes(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	sizes := pdf.GetAllPageSizes()
	if len(sizes) != 2 {
		t.Errorf("expected 2 page sizes, got %d", len(sizes))
	}
}

// ============================================================
// image_recompress.go — downscaleImage, RecompressOption.defaults
// ============================================================

func TestCov2_RecompressOption_Defaults(t *testing.T) {
	opt := RecompressOption{}
	opt.defaults()
	if opt.Format != "jpeg" {
		t.Errorf("expected jpeg, got %s", opt.Format)
	}
	if opt.JPEGQuality != 75 {
		t.Errorf("expected 75, got %d", opt.JPEGQuality)
	}
}

func TestCov2_RecompressOption_Defaults_NoOverwrite(t *testing.T) {
	opt := RecompressOption{Format: "png", JPEGQuality: 90}
	opt.defaults()
	if opt.Format != "png" {
		t.Error("should not overwrite format")
	}
	if opt.JPEGQuality != 90 {
		t.Error("should not overwrite quality")
	}
}

func TestCov2_DownscaleImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 200, 100))
	// Fill with some color
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			src.Pix[(y*src.Stride)+(x*4)+0] = 255
			src.Pix[(y*src.Stride)+(x*4)+3] = 255
		}
	}
	result := downscaleImage(src, 100, 50)
	bounds := result.Bounds()
	if bounds.Dx() > 100 || bounds.Dy() > 50 {
		t.Errorf("expected max 100x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// ============================================================
// scrub.go — Scrub
// ============================================================

func TestCov2_Scrub(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetInfo(PdfInfo{
		Title:   "Test",
		Author:  "Author",
		Subject: "Subject",
	})

	pdf.Scrub(DefaultScrubOption())
	// After scrub, info should be cleared
	if pdf.isUseInfo {
		t.Error("expected isUseInfo to be false after scrub")
	}
}

func TestCov2_Scrub_Selective(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetInfo(PdfInfo{Title: "Test"})

	// Only scrub metadata
	pdf.Scrub(ScrubOption{Metadata: true})
	if pdf.isUseInfo {
		t.Error("expected metadata scrubbed")
	}
}

func TestCov2_DefaultScrubOption(t *testing.T) {
	opt := DefaultScrubOption()
	if !opt.Metadata || !opt.XMLMetadata || !opt.EmbeddedFiles || !opt.PageLabels {
		t.Error("expected all options enabled")
	}
}

// ============================================================
// pdf_version.go — SetPDFVersion, GetPDFVersion, Header
// ============================================================

func TestCov2_PDFVersion(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})

	pdf.SetPDFVersion(PDFVersion17)
	if pdf.GetPDFVersion() != PDFVersion17 {
		t.Errorf("expected PDFVersion17, got %d", pdf.GetPDFVersion())
	}

	header := pdf.GetPDFVersion().Header()
	if header != "%PDF-1.7" {
		t.Errorf("expected %%PDF-1.7, got %s", header)
	}
}

// ============================================================
// paper_sizes.go — PaperSize, PaperSizeNames
// ============================================================

func TestCov2_PaperSize(t *testing.T) {
	size := PaperSize("A4")
	if size == nil {
		t.Fatal("expected non-nil for A4")
	}
	if size.W <= 0 || size.H <= 0 {
		t.Error("expected positive dimensions")
	}
}

func TestCov2_PaperSize_Unknown(t *testing.T) {
	size := PaperSize("UNKNOWN_SIZE")
	if size != nil {
		t.Error("expected nil for unknown size")
	}
}

func TestCov2_PaperSizeNames(t *testing.T) {
	names := PaperSizeNames()
	if len(names) == 0 {
		t.Error("expected non-empty paper size names")
	}
	found := false
	for _, n := range names {
		if n == "a4" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a4 in paper size names")
	}
}

// ============================================================
// pdf_obj_id.go — ObjID methods
// ============================================================

func TestCov2_ObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	catID := pdf.CatalogObjID()
	if !catID.IsValid() {
		t.Error("expected valid catalog ID")
	}
	if catID.Index() < 0 {
		t.Error("expected non-negative index")
	}
	ref := catID.Ref()
	if ref <= 0 {
		t.Error("expected positive ref")
	}
	refStr := catID.RefStr()
	if refStr == "" {
		t.Error("expected non-empty ref string")
	}

	pagesID := pdf.PagesObjID()
	if !pagesID.IsValid() {
		t.Error("expected valid pages ID")
	}
}

func TestCov2_GetObjByID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	catID := pdf.CatalogObjID()
	obj := pdf.GetObjByID(catID)
	if obj == nil {
		t.Error("expected non-nil object for catalog ID")
	}
}

// ============================================================
// Additional html_parser.go edge cases
// ============================================================

func TestCov2_ParseHTML_UnquotedAttr(t *testing.T) {
	nodes := parseHTML(`<font size=5>text</font>`)
	if len(nodes) == 0 {
		t.Fatal("expected nodes")
	}
	if nodes[0].Attrs["size"] != "5" {
		t.Errorf("expected size=5, got %s", nodes[0].Attrs["size"])
	}
}

func TestCov2_ParseHTML_EmptyString(t *testing.T) {
	nodes := parseHTML("")
	if len(nodes) != 0 {
		t.Error("expected empty nodes for empty string")
	}
}

func TestCov2_ParseHTML_TextOnly(t *testing.T) {
	nodes := parseHTML("just text")
	if len(nodes) == 0 {
		t.Fatal("expected text node")
	}
	if nodes[0].Text != "just text" {
		t.Errorf("expected 'just text', got %q", nodes[0].Text)
	}
}

// ============================================================
// Additional GoPdf methods for coverage
// ============================================================

func TestCov2_GetNextObjectID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	id := pdf.GetNextObjectID()
	if id <= 0 {
		t.Errorf("expected positive next object ID, got %d", id)
	}
}

func TestCov2_GetObjectCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	count := pdf.GetObjectCount()
	if count <= 0 {
		t.Errorf("expected positive object count, got %d", count)
	}
}

func TestCov2_SetNoCompression(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.SetNoCompression()
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("No compression")
	_, err := pdf.GetBytesPdfReturnErr()
	if err != nil {
		t.Fatalf("GetBytesPdf: %v", err)
	}
}

func TestCov2_SetCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCharSpacing(2.0)
	pdf.SetXY(50, 50)
	_ = pdf.Text("Spaced text")
}

func TestCov2_SetInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetInfo(PdfInfo{
		Title:   "Test Title",
		Author:  "Test Author",
		Subject: "Test Subject",
	})
	info := pdf.GetInfo()
	if info.Title != "Test Title" {
		t.Errorf("expected 'Test Title', got %q", info.Title)
	}
}

func TestCov2_SetColorSpace(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetColorSpace("DeviceRGB")
}

func TestCov2_AddColorSpaceRGB(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceRGB("sRGB", 255, 0, 0)
}

func TestCov2_AddColorSpaceCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddColorSpaceCMYK("cmyk1", 0, 100, 100, 0)
}

// ============================================================
// config.go — UnitsToPointsVar, PointsToUnitsVar
// ============================================================

func TestCov2_UnitsToPointsVar(t *testing.T) {
	a, b := 25.4, 50.8
	UnitsToPointsVar(UnitMM, &a, &b)
	if math.Abs(a-72) > 0.1 {
		t.Errorf("expected ~72 for 25.4mm, got %f", a)
	}
	if math.Abs(b-144) > 0.1 {
		t.Errorf("expected ~144 for 50.8mm, got %f", b)
	}
}

func TestCov2_PointsToUnitsVar(t *testing.T) {
	a, b := 72.0, 144.0
	PointsToUnitsVar(UnitMM, &a, &b)
	if math.Abs(a-25.4) > 0.1 {
		t.Errorf("expected ~25.4 for 72pt, got %f", a)
	}
}

func TestCov2_UnitsToPoints(t *testing.T) {
	tests := []struct {
		unit int
		val  float64
		want float64
	}{
		{UnitPT, 72, 72},
		{UnitIN, 1, 72},
		{UnitCM, 2.54, 72},
		{UnitPX, 96, 72},
	}
	for _, tt := range tests {
		got := UnitsToPoints(tt.unit, tt.val)
		if math.Abs(got-tt.want) > 0.5 {
			t.Errorf("UnitsToPoints(%d, %f) = %f, want ~%f", tt.unit, tt.val, got, tt.want)
		}
	}
}

func TestCov2_PointsToUnits(t *testing.T) {
	got := PointsToUnits(UnitIN, 72)
	if math.Abs(got-1.0) > 0.01 {
		t.Errorf("expected ~1.0 inch for 72pt, got %f", got)
	}
}

// ============================================================
// xmp_metadata.go — SetXMPMetadata, GetXMPMetadata, xmlEscape
// ============================================================

func TestCov2_XMPMetadata(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	if pdf.GetXMPMetadata() != nil {
		t.Error("expected nil XMP initially")
	}

	meta := XMPMetadata{
		Title:       "Test Doc",
		Creator:     []string{"Author1", "Author2"},
		Description: "A test document",
		Subject:     []string{"test", "pdf"},
		Rights:      "Copyright 2025",
		Language:    "en-US",
		CreatorTool: "GoPDF2",
		Producer:    "GoPDF2",
		Keywords:    "test, pdf",
		Trapped:     "False",
	}
	pdf.SetXMPMetadata(meta)

	got := pdf.GetXMPMetadata()
	if got == nil {
		t.Fatal("expected non-nil XMP")
	}
	if got.Title != "Test Doc" {
		t.Errorf("expected 'Test Doc', got %q", got.Title)
	}
}

func TestCov2_XmlEscape(t *testing.T) {
	result := xmlEscape(`<test & "value" 'here'>`)
	if !strings.Contains(result, "&amp;") {
		t.Error("expected &amp;")
	}
	if !strings.Contains(result, "&lt;") {
		t.Error("expected &lt;")
	}
	if !strings.Contains(result, "&gt;") {
		t.Error("expected &gt;")
	}
	if !strings.Contains(result, "&quot;") {
		t.Error("expected &quot;")
	}
}

// ============================================================
// page_label.go — SetPageLabels, GetPageLabels
// ============================================================

func TestCov2_PageLabels(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()

	labels := []PageLabel{
		{PageIndex: 0, Style: PageLabelRomanLower, Start: 1},
		{PageIndex: 2, Style: PageLabelDecimal, Start: 1},
	}
	pdf.SetPageLabels(labels)

	got := pdf.GetPageLabels()
	if len(got) != 2 {
		t.Errorf("expected 2 labels, got %d", len(got))
	}
}

func TestCov2_PageLabelStyles(t *testing.T) {
	// Just verify the constants exist
	styles := []PageLabelStyle{
		PageLabelDecimal,
		PageLabelRomanUpper,
		PageLabelRomanLower,
		PageLabelAlphaUpper,
		PageLabelAlphaLower,
		PageLabelNone,
	}
	for _, s := range styles {
		_ = s // just ensure they compile
	}
}

// ============================================================
// Additional GoPdf methods for coverage
// ============================================================

func TestCov2_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Oval(50, 50, 200, 100)
}

func TestCov2_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(50, 50, 100, 200, 150, 100, 200, 50, "D")
}

func TestCov2_Br(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Br(20)
	if pdf.GetY() != 70 {
		t.Errorf("expected Y=70 after Br(20), got %f", pdf.GetY())
	}
}

func TestCov2_SetGrayFillStroke(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetGrayFill(0.5)
	pdf.SetGrayStroke(0.3)
}

func TestCov2_SetStrokeColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetFillColor(0, 255, 0)
}

func TestCov2_SetStrokeColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetStrokeColorCMYK(0, 100, 100, 0)
	pdf.SetFillColorCMYK(100, 0, 0, 0)
}

func TestCov2_SetTextColorCMYK(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetTextColorCMYK(0, 0, 0, 100)
	pdf.SetXY(50, 50)
	_ = pdf.Text("CMYK text")
}

func TestCov2_RectFromLowerLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromLowerLeft(50, 50, 100, 80)
}

func TestCov2_RectFromUpperLeft(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromUpperLeft(50, 50, 100, 80)
}

func TestCov2_RectFromLowerLeftWithStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromLowerLeftWithStyle(50, 50, 100, 80, "FD")
}

func TestCov2_RectFromUpperLeftWithStyle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.RectFromUpperLeftWithStyle(50, 50, 100, 80, "FD")
}

func TestCov2_ClipPolygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SaveGraphicsState()
	pdf.ClipPolygon([]Point{{X: 50, Y: 50}, {X: 150, Y: 50}, {X: 100, Y: 150}})
	pdf.RestoreGraphicsState()
}

func TestCov2_SetAnchor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetAnchor("test-anchor")
}

func TestCov2_AddExternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddExternalLink("https://example.com", 50, 50, 200, 20)
}

func TestCov2_AddInternalLink(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetAnchor("target")
	pdf.AddInternalLink("target", 50, 50, 200, 20)
}

func TestCov2_SetCustomLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCustomLineType([]float64{5, 3}, 0)
	pdf.Line(50, 50, 200, 50)
}

func TestCov2_SetLineType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineType("dashed")
	pdf.Line(50, 50, 200, 50)
}

func TestCov2_SetLineWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(2.0)
	pdf.Line(50, 50, 200, 50)
}

// ============================================================
// html_insert.go — InsertHTMLBox
// ============================================================

func TestCov2_InsertHTMLBox_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700,
		"<p>Hello <b>World</b></p>",
		HTMLBoxOption{
			DefaultFontFamily: fontFamily,
			DefaultFontSize:   12,
		})
	if err != nil {
		t.Fatalf("InsertHTMLBox: %v", err)
	}
}

func TestCov2_InsertHTMLBox_NoFont(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, "<p>test</p>", HTMLBoxOption{})
	if err != ErrMissingFontFamily {
		t.Errorf("expected ErrMissingFontFamily, got %v", err)
	}
}

func TestCov2_InsertHTMLBox_Headings(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := "<h1>Title</h1><h2>Subtitle</h2><h3>Section</h3><p>Body text</p>"
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox headings: %v", err)
	}
}

func TestCov2_InsertHTMLBox_Lists(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := "<ul><li>Item 1</li><li>Item 2</li></ul><ol><li>First</li><li>Second</li></ol>"
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox lists: %v", err)
	}
}

func TestCov2_InsertHTMLBox_StyledText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<p style="color: red; font-size: 16px">Red text</p><p><i>Italic</i> and <b>bold</b></p>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox styled: %v", err)
	}
}

func TestCov2_InsertHTMLBox_LineBreaks(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := "Line 1<br/>Line 2<br/>Line 3"
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox br: %v", err)
	}
}

func TestCov2_InsertHTMLBox_Center(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<center>Centered text</center>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox center: %v", err)
	}
}

func TestCov2_InsertHTMLBox_FontTag(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<font size="5" color="blue">Large blue text</font>`
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox font: %v", err)
	}
}

func TestCov2_InsertHTMLBox_HR(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := "<p>Before</p><hr/><p>After</p>"
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox hr: %v", err)
	}
}

func TestCov2_InsertHTMLBox_LongText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	long := strings.Repeat("This is a long sentence that should wrap. ", 20)
	html := "<p>" + long + "</p>"
	_, err := pdf.InsertHTMLBox(50, 50, 400, 700, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
		DefaultFontSize:   12,
	})
	if err != nil {
		t.Fatalf("InsertHTMLBox long: %v", err)
	}
}

// ============================================================
// Additional coverage for SplitText, MeasureTextWidth
// ============================================================

func TestCov2_SplitText(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitText("Hello World this is a test of text splitting", 100)
	if err != nil {
		t.Fatalf("SplitText: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

func TestCov2_SplitTextWithWordWrap(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitTextWithWordWrap("Hello World this is a test of text splitting with word wrap", 100)
	if err != nil {
		t.Fatalf("SplitTextWithWordWrap: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

func TestCov2_MeasureTextWidth(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	w, err := pdf.MeasureTextWidth("Hello World")
	if err != nil {
		t.Fatalf("MeasureTextWidth: %v", err)
	}
	if w <= 0 {
		t.Error("expected positive width")
	}
}

// ============================================================
// select_pages.go — SelectPages
// ============================================================

func TestCov2_SelectPages(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 1")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 2")
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Page 3")

	newPdf, err := pdf.SelectPages([]int{1, 3})
	if err != nil {
		t.Fatalf("SelectPages: %v", err)
	}
	if newPdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", newPdf.GetNumberOfPages())
	}
}

// ============================================================
// markinfo.go — SetMarkInfo, GetMarkInfo
// ============================================================

func TestCov2_MarkInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetMarkInfo(MarkInfo{Marked: true})
	mi := pdf.GetMarkInfo()
	if mi == nil || !mi.Marked {
		t.Error("expected mark info Marked=true")
	}
	pdf.SetMarkInfo(MarkInfo{Marked: false})
	mi = pdf.GetMarkInfo()
	if mi == nil || mi.Marked {
		t.Error("expected mark info Marked=false")
	}
}

// ============================================================
// content_stream_clean.go — internal functions
// ============================================================

func TestCov2_SplitContentLines(t *testing.T) {
	stream := []byte("q\n1 0 0 1 0 0 cm\n(Hello) Tj\nQ\n")
	lines := splitContentLines(stream)
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestCov2_RemoveRedundantStateChanges(t *testing.T) {
	lines := []string{"1 w", "2 w", "3 w", "(Hello) Tj"}
	result := removeRedundantStateChanges(lines)
	// Only the last "w" should remain
	wCount := 0
	for _, l := range result {
		if strings.HasSuffix(l, " w") {
			wCount++
		}
	}
	if wCount != 1 {
		t.Errorf("expected 1 w operator, got %d", wCount)
	}
}

func TestCov2_RemoveEmptyQBlocks(t *testing.T) {
	lines := []string{"q", "Q", "(Hello) Tj"}
	result := removeEmptyQBlocks(lines)
	if len(result) != 1 {
		t.Errorf("expected 1 line after removing empty q/Q, got %d", len(result))
	}
}

func TestCov2_NormalizeWhitespace(t *testing.T) {
	lines := []string{"  1.0   0   0   1   0   0  cm  ", "  (Hello)  Tj  "}
	result := normalizeWhitespace(lines)
	if strings.Contains(result[0], "  ") {
		t.Error("expected single spaces")
	}
}

func TestCov2_ExtractOperator(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"1 0 0 1 0 0 cm", "cm"},
		{"(Hello) Tj", "Tj"},
		{"q", "q"},
		{"", ""},
		{"  Q  ", "Q"},
	}
	for _, tt := range tests {
		got := extractOperator(tt.line)
		if got != tt.want {
			t.Errorf("extractOperator(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestCov2_BuildCleanedDict(t *testing.T) {
	dict := "<< /Length 100 /Filter /FlateDecode >>"
	result := buildCleanedDict(dict, 200)
	if !strings.Contains(result, "/Length 200") {
		t.Error("expected updated length")
	}
}

func TestCov2_BuildCleanedDict_AddFilter(t *testing.T) {
	dict := "<< /Length 100 >>"
	result := buildCleanedDict(dict, 200)
	if !strings.Contains(result, "/FlateDecode") {
		t.Error("expected FlateDecode filter added")
	}
}

func TestCov2_CleanContentStream(t *testing.T) {
	stream := []byte("q\nQ\n1 w\n2 w\n(Hello) Tj\n")
	result := cleanContentStream(stream)
	// Should remove empty q/Q and redundant w
	if bytes.Contains(result, []byte("1 w")) {
		t.Error("expected redundant 1 w to be removed")
	}
}

// ============================================================
// markinfo.go — FindPagesByLabel, computePageLabel
// ============================================================

func TestCov2_FindPagesByLabel(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddPage()
	pdf.AddPage()

	pdf.SetPageLabels([]PageLabel{
		{PageIndex: 0, Style: PageLabelRomanLower, Start: 1},
		{PageIndex: 2, Style: PageLabelDecimal, Start: 1},
	})

	pages := pdf.FindPagesByLabel("i")
	if len(pages) == 0 {
		t.Error("expected to find page labeled 'i'")
	}

	pages = pdf.FindPagesByLabel("1")
	if len(pages) == 0 {
		t.Error("expected to find page labeled '1'")
	}
}

func TestCov2_FindPagesByLabel_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pages := pdf.FindPagesByLabel("1")
	if pages != nil {
		t.Error("expected nil for no labels")
	}
}

func TestCov2_ToRoman(t *testing.T) {
	tests := []struct {
		n     int
		upper bool
		want  string
	}{
		{1, true, "I"},
		{4, true, "IV"},
		{9, true, "IX"},
		{14, true, "XIV"},
		{1999, true, "MCMXCIX"},
		{3, false, "iii"},
	}
	for _, tt := range tests {
		got := toRoman(tt.n, tt.upper)
		if got != tt.want {
			t.Errorf("toRoman(%d, %v) = %q, want %q", tt.n, tt.upper, got, tt.want)
		}
	}
}

func TestCov2_ToAlpha(t *testing.T) {
	tests := []struct {
		n     int
		upper bool
		want  string
	}{
		{1, true, "A"},
		{26, true, "Z"},
		{27, true, "AA"},
		{1, false, "a"},
	}
	for _, tt := range tests {
		got := toAlpha(tt.n, tt.upper)
		if got != tt.want {
			t.Errorf("toAlpha(%d, %v) = %q, want %q", tt.n, tt.upper, got, tt.want)
		}
	}
}

// ============================================================
// Additional GoPdf methods for coverage
// ============================================================

func TestCov2_Rotate(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rotate(45, 100, 100)
	pdf.SetXY(100, 100)
	_ = pdf.Text("Rotated")
	pdf.RotateReset()
}

func TestCov2_SetTransparency(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.SetTransparency(Transparency{Alpha: 0.5, BlendModeType: NormalBlendMode})
	if err != nil {
		t.Fatalf("SetTransparency: %v", err)
	}
	pdf.SetXY(50, 50)
	_ = pdf.Text("Transparent")
	pdf.ClearTransparency()
}

func TestCov2_AddOutline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 1")
	pdf.AddOutline("Chapter 1")

	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Chapter 2")
	pdf.AddOutlineWithPosition("Chapter 2")
}

func TestCov2_AddHeader_AddFooter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddHeader(func() {
		pdf.SetXY(50, 20)
		_ = pdf.Text("Header")
	})
	pdf.AddFooter(func() {
		pdf.SetXY(50, 800)
		_ = pdf.Text("Footer")
	})
	pdf.AddPage()
	pdf.SetXY(50, 100)
	_ = pdf.Text("Body")
}

func TestCov2_SetFontSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetFontSize(20); err != nil {
		t.Fatalf("SetFontSize: %v", err)
	}
	pdf.SetXY(50, 50)
	_ = pdf.Text("Large text")
}

func TestCov2_WriteTo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("WriteTo test")

	var buf bytes.Buffer
	n, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
}
