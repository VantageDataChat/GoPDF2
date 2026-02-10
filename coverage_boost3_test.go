package gopdf

import (
	"bytes"
	"encoding/xml"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost3_test.go - Third round of coverage tests.
// All functions prefixed TestCov3_.
// ============================================================

// ============================================================
// svg_insert.go - parseSVGColor
// ============================================================

func TestCov3_ParseSVGColor_Hex6(t *testing.T) {
	c, ok := parseSVGColor("#ff8800")
	if !ok || c != [3]uint8{255, 136, 0} {
		t.Errorf("got %v %v", c, ok)
	}
}

func TestCov3_ParseSVGColor_Hex3(t *testing.T) {
	c, ok := parseSVGColor("#f80")
	if !ok || c != [3]uint8{255, 136, 0} {
		t.Errorf("got %v %v", c, ok)
	}
}

func TestCov3_ParseSVGColor_RGB(t *testing.T) {
	c, ok := parseSVGColor("rgb(10, 20, 30)")
	if !ok || c != [3]uint8{10, 20, 30} {
		t.Errorf("got %v %v", c, ok)
	}
}

func TestCov3_ParseSVGColor_Named(t *testing.T) {
	names := map[string][3]uint8{
		"black":   {0, 0, 0},
		"white":   {255, 255, 255},
		"red":     {255, 0, 0},
		"green":   {0, 128, 0},
		"blue":    {0, 0, 255},
		"yellow":  {255, 255, 0},
		"cyan":    {0, 255, 255},
		"magenta": {255, 0, 255},
		"gray":    {128, 128, 128},
		"grey":    {128, 128, 128},
		"orange":  {255, 165, 0},
		"purple":  {128, 0, 128},
	}
	for name, want := range names {
		c, ok := parseSVGColor(name)
		if !ok || c != want {
			t.Errorf("parseSVGColor(%q) = %v %v, want %v", name, c, ok, want)
		}
	}
}

func TestCov3_ParseSVGColor_Invalid(t *testing.T) {
	_, ok := parseSVGColor("notacolor")
	if ok {
		t.Error("expected false for unknown color")
	}
	_, ok = parseSVGColor("#zz")
	if ok {
		t.Error("expected false for bad hex")
	}
}

// ============================================================
// svg_insert.go - parseSVGStyle
// ============================================================

func TestCov3_ParseSVGStyle_Attrs(t *testing.T) {
	elem := svgElement{}
	attrs := map[string]string{
		"fill":         "#ff0000",
		"stroke":       "blue",
		"stroke-width": "2.5",
		"opacity":      "0.8",
	}
	parseSVGStyle(&elem, attrs)
	if !elem.hasFill || elem.fill != [3]uint8{255, 0, 0} {
		t.Errorf("fill: %v %v", elem.fill, elem.hasFill)
	}
	if !elem.hasStroke || elem.stroke != [3]uint8{0, 0, 255} {
		t.Errorf("stroke: %v %v", elem.stroke, elem.hasStroke)
	}
	if elem.strokeW != 2.5 {
		t.Errorf("strokeW = %f", elem.strokeW)
	}
	if elem.opacity != 0.8 {
		t.Errorf("opacity = %f", elem.opacity)
	}
}

func TestCov3_ParseSVGStyle_InlineStyle(t *testing.T) {
	elem := svgElement{}
	attrs := map[string]string{
		"style": "fill: rgb(10,20,30); stroke: #00ff00; stroke-width: 3; opacity: 0.5",
	}
	parseSVGStyle(&elem, attrs)
	if !elem.hasFill || elem.fill != [3]uint8{10, 20, 30} {
		t.Errorf("fill: %v %v", elem.fill, elem.hasFill)
	}
	if !elem.hasStroke || elem.stroke != [3]uint8{0, 255, 0} {
		t.Errorf("stroke: %v %v", elem.stroke, elem.hasStroke)
	}
	if elem.strokeW != 3 {
		t.Errorf("strokeW = %f", elem.strokeW)
	}
	if elem.opacity != 0.5 {
		t.Errorf("opacity = %f", elem.opacity)
	}
}

func TestCov3_ParseSVGStyle_None(t *testing.T) {
	elem := svgElement{}
	attrs := map[string]string{"fill": "none", "stroke": "none"}
	parseSVGStyle(&elem, attrs)
	if elem.hasFill {
		t.Error("fill should not be set for none")
	}
	if elem.hasStroke {
		t.Error("stroke should not be set for none")
	}
}

func TestCov3_ParseSVGStyle_InlineNone(t *testing.T) {
	elem := svgElement{}
	attrs := map[string]string{"style": "fill: none; stroke: none"}
	parseSVGStyle(&elem, attrs)
	if elem.hasFill {
		t.Error("fill should not be set")
	}
	if elem.hasStroke {
		t.Error("stroke should not be set")
	}
}

// ============================================================
// svg_insert.go - parseSVGPoints, parseSVGLength, atof
// ============================================================

func TestCov3_ParseSVGPoints(t *testing.T) {
	pts := parseSVGPoints("10,20 30,40 50,60")
	if len(pts) != 3 {
		t.Fatalf("got %d points", len(pts))
	}
	if pts[0].X != 10 || pts[0].Y != 20 {
		t.Errorf("pt0: %v", pts[0])
	}
	if pts[2].X != 50 || pts[2].Y != 60 {
		t.Errorf("pt2: %v", pts[2])
	}
}

func TestCov3_ParseSVGPoints_Odd(t *testing.T) {
	pts := parseSVGPoints("10 20 30")
	if len(pts) != 1 {
		t.Errorf("expected 1 point, got %d", len(pts))
	}
}

func TestCov3_ParseSVGLength(t *testing.T) {
	tests := []struct {
		s    string
		want float64
	}{
		{"100px", 100},
		{"72pt", 72},
		{"25.4mm", 25.4},
		{"2.54cm", 2.54},
		{"50", 50},
	}
	for _, tt := range tests {
		got := parseSVGLength(tt.s)
		if got != tt.want {
			t.Errorf("parseSVGLength(%q) = %f, want %f", tt.s, got, tt.want)
		}
	}
}

func TestCov3_Atof(t *testing.T) {
	if atof("3.14") != 3.14 {
		t.Error("atof failed")
	}
	if atof("bad") != 0 {
		t.Error("atof should return 0 for bad input")
	}
}

// ============================================================
// svg_insert.go - tokenizeSVGPath, isSVGCommand
// ============================================================

func TestCov3_TokenizeSVGPath(t *testing.T) {
	tokens := tokenizeSVGPath("M 10 20 L 30 40 Z")
	expected := []string{"M", "10", "20", "L", "30", "40", "Z"}
	if len(tokens) != len(expected) {
		t.Fatalf("got %d tokens: %v", len(tokens), tokens)
	}
	for i, tok := range tokens {
		if tok != expected[i] {
			t.Errorf("token[%d] = %q, want %q", i, tok, expected[i])
		}
	}
}

func TestCov3_TokenizeSVGPath_Negative(t *testing.T) {
	tokens := tokenizeSVGPath("M10-20L30-40")
	if len(tokens) < 4 {
		t.Fatalf("expected at least 4 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestCov3_TokenizeSVGPath_Commas(t *testing.T) {
	tokens := tokenizeSVGPath("M10,20 L30,40")
	if len(tokens) != 6 {
		t.Fatalf("got %d tokens: %v", len(tokens), tokens)
	}
}

func TestCov3_IsSVGCommand(t *testing.T) {
	cmds := "MmLlHhVvCcSsQqTtAaZz"
	for _, ch := range cmds {
		if !isSVGCommand(string(ch)) {
			t.Errorf("isSVGCommand(%q) should be true", string(ch))
		}
	}
	if isSVGCommand("X") {
		t.Error("X should not be a command")
	}
	if isSVGCommand("ML") {
		t.Error("multi-char should not be a command")
	}
}

func TestCov3_IsSVGCommandByte(t *testing.T) {
	if !isSVGCommandByte('M') {
		t.Error("M should be a command byte")
	}
	if isSVGCommandByte('X') {
		t.Error("X should not be a command byte")
	}
}

// ============================================================
// svg_insert.go - svgPaintStyle, scaleSVGPoints
// ============================================================

func TestCov3_SvgPaintStyle(t *testing.T) {
	e1 := svgElement{hasFill: true, hasStroke: true}
	if svgPaintStyle(e1) != "FD" {
		t.Error("expected FD")
	}
	e2 := svgElement{hasFill: true}
	if svgPaintStyle(e2) != "F" {
		t.Error("expected F")
	}
	e3 := svgElement{hasStroke: true}
	if svgPaintStyle(e3) != "D" {
		t.Error("expected D")
	}
	e4 := svgElement{}
	if svgPaintStyle(e4) != "D" {
		t.Error("expected D")
	}
}

func TestCov3_ScaleSVGPoints(t *testing.T) {
	pts := []Point{{X: 10, Y: 20}, {X: 30, Y: 40}}
	scaled := scaleSVGPoints(pts, 5, 10, 2, 3)
	if scaled[0].X != 25 || scaled[0].Y != 70 {
		t.Errorf("pt0: %v", scaled[0])
	}
	if scaled[1].X != 65 || scaled[1].Y != 130 {
		t.Errorf("pt1: %v", scaled[1])
	}
}

// ============================================================
// svg_insert.go - parseSVGElement
// ============================================================

func TestCov3_ParseSVGElement_Rect(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "rect"},
		Attrs: []xml.Attr{
			{Name: xml.Name{Local: "x"}, Value: "10"},
			{Name: xml.Name{Local: "y"}, Value: "20"},
			{Name: xml.Name{Local: "width"}, Value: "100"},
			{Name: xml.Name{Local: "height"}, Value: "50"},
			{Name: xml.Name{Local: "rx"}, Value: "5"},
		},
	}
	elem, ok := parseSVGElement(el)
	if !ok {
		t.Fatal("expected ok")
	}
	if elem.typ != svgRect || elem.x != 10 || elem.y != 20 || elem.w != 100 || elem.h != 50 || elem.rx != 5 {
		t.Errorf("rect values wrong: %+v", elem)
	}
}

func TestCov3_ParseSVGElement_Circle(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "circle"},
		Attrs: []xml.Attr{
			{Name: xml.Name{Local: "cx"}, Value: "50"},
			{Name: xml.Name{Local: "cy"}, Value: "60"},
			{Name: xml.Name{Local: "r"}, Value: "25"},
		},
	}
	elem, ok := parseSVGElement(el)
	if !ok || elem.typ != svgCircle || elem.cx != 50 || elem.cy != 60 || elem.r != 25 {
		t.Errorf("circle wrong: ok=%v %+v", ok, elem)
	}
}

func TestCov3_ParseSVGElement_Ellipse(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "ellipse"},
		Attrs: []xml.Attr{
			{Name: xml.Name{Local: "cx"}, Value: "50"},
			{Name: xml.Name{Local: "cy"}, Value: "60"},
			{Name: xml.Name{Local: "rx"}, Value: "30"},
			{Name: xml.Name{Local: "ry"}, Value: "20"},
		},
	}
	elem, ok := parseSVGElement(el)
	if !ok || elem.typ != svgEllipse {
		t.Errorf("ellipse wrong: ok=%v %+v", ok, elem)
	}
}

func TestCov3_ParseSVGElement_Line(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "line"},
		Attrs: []xml.Attr{
			{Name: xml.Name{Local: "x1"}, Value: "0"},
			{Name: xml.Name{Local: "y1"}, Value: "0"},
			{Name: xml.Name{Local: "x2"}, Value: "100"},
			{Name: xml.Name{Local: "y2"}, Value: "100"},
		},
	}
	elem, ok := parseSVGElement(el)
	if !ok || elem.typ != svgLine {
		t.Errorf("line wrong: ok=%v", ok)
	}
}

func TestCov3_ParseSVGElement_Polyline(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "polyline"},
		Attrs:   []xml.Attr{{Name: xml.Name{Local: "points"}, Value: "10,20 30,40 50,60"}},
	}
	elem, ok := parseSVGElement(el)
	if !ok || elem.typ != svgPolyline || len(elem.points) != 3 {
		t.Errorf("polyline wrong: ok=%v %+v", ok, elem)
	}
}

func TestCov3_ParseSVGElement_Polygon(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "polygon"},
		Attrs:   []xml.Attr{{Name: xml.Name{Local: "points"}, Value: "10,20 30,40 50,60"}},
	}
	elem, ok := parseSVGElement(el)
	if !ok || elem.typ != svgPolygon {
		t.Errorf("polygon wrong: ok=%v", ok)
	}
}

func TestCov3_ParseSVGElement_Path(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "path"},
		Attrs:   []xml.Attr{{Name: xml.Name{Local: "d"}, Value: "M 10 20 L 30 40 Z"}},
	}
	elem, ok := parseSVGElement(el)
	if !ok || elem.typ != svgPath || elem.pathData != "M 10 20 L 30 40 Z" {
		t.Errorf("path wrong: ok=%v %+v", ok, elem)
	}
}

func TestCov3_ParseSVGElement_Unknown(t *testing.T) {
	el := xmlElement{XMLName: xml.Name{Local: "text"}}
	_, ok := parseSVGElement(el)
	if ok {
		t.Error("expected false for unknown element")
	}
}

func TestCov3_ParseSVGElement_EmptyPolyline(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "polyline"},
		Attrs:   []xml.Attr{{Name: xml.Name{Local: "points"}, Value: ""}},
	}
	_, ok := parseSVGElement(el)
	if ok {
		t.Error("expected false for empty polyline")
	}
}

func TestCov3_ParseSVGElement_EmptyPath(t *testing.T) {
	el := xmlElement{
		XMLName: xml.Name{Local: "path"},
		Attrs:   []xml.Attr{{Name: xml.Name{Local: "d"}, Value: ""}},
	}
	_, ok := parseSVGElement(el)
	if ok {
		t.Error("expected false for empty path")
	}
}

// ============================================================
// svg_insert.go - renderSVGPath (via ImageSVGFromBytes)
// ============================================================

func TestCov3_ImageSVGFromBytes_Rect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<rect x="10" y="10" width="80" height="40" fill="red" stroke="blue" stroke-width="2"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Circle(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<circle cx="100" cy="100" r="50" fill="green"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Ellipse(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<ellipse cx="100" cy="100" rx="80" ry="40" fill="yellow" stroke="black"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Line(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<line x1="0" y1="0" x2="200" y2="200" stroke="black" stroke-width="2"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Polyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<polyline points="10,10 50,50 90,10 130,50" stroke="red" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Polygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<polygon points="100,10 40,198 190,78 10,78 160,198" fill="purple" stroke="black"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_Lines(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 10 L 100 10 L 100 100 L 10 100 Z" stroke="black" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_Relative(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="m 10 10 l 90 0 l 0 90 l -90 0 z" stroke="blue" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_HV(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 10 H 100 V 100 h -90 v -90 Z" stroke="green" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_Cubic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 80 C 40 10 65 10 95 80" stroke="black" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_RelCubic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 80 c 30 -70 55 -70 85 0" stroke="red" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_Quadratic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 80 Q 52 10 95 80" stroke="blue" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_Path_RelQuadratic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 80 q 42 -70 85 0" stroke="purple" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_RoundedRect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<rect x="10" y="10" width="80" height="40" rx="10" fill="orange"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ImageSVGFromBytes_StyledPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="200" height="200">
		<path d="M 10 10 L 100 100" style="stroke: red; stroke-width: 3; opacity: 0.7" fill="none"/>
	</svg>`
	err := pdf.ImageSVGFromBytes([]byte(svg), SVGOption{X: 10, Y: 10, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// content_element.go - classifyElement (all types)
// ============================================================

func TestCov3_ClassifyElement_AllTypes(t *testing.T) {
	tests := []struct {
		cache ICacheContent
		want  ContentElementType
	}{
		{&cacheContentText{}, ElementText},
		{&cacheContentImage{}, ElementImage},
		{&cacheContentLine{}, ElementLine},
		{cacheContentRectangle{}, ElementRectangle},
		{&cacheContentOval{}, ElementOval},
		{&cacheContentPolygon{}, ElementPolygon},
		{&cacheContentCurve{}, ElementCurve},
		{&cacheContentPolyline{}, ElementPolyline},
		{&cacheContentSector{}, ElementSector},
		{&cacheContentImportedTemplate{}, ElementImportedTemplate},
		{&cacheContentLineWidth{}, ElementLineWidth},
		{&cacheContentLineType{}, ElementLineType},
		{&cacheContentCustomLineType{}, ElementCustomLineType},
		{&cacheContentGray{}, ElementGray},
		{&cacheContentColorRGB{}, ElementColorRGB},
		{&cacheContentColorCMYK{}, ElementColorCMYK},
		{&cacheColorSpace{}, ElementColorSpace},
		{&cacheContentRotate{}, ElementRotate},
		{&cacheContentClipPolygon{}, ElementClipPolygon},
		{&cacheContentSaveGraphicsState{}, ElementSaveGState},
		{&cacheContentRestoreGraphicsState{}, ElementRestoreGState},
	}
	for _, tt := range tests {
		got := classifyElement(tt.cache)
		if got != tt.want {
			t.Errorf("classifyElement(%T) = %v, want %v", tt.cache, got, tt.want)
		}
	}
}

// ============================================================
// content_element.go - buildElement (all types)
// ============================================================

func TestCov3_BuildElement_Text(t *testing.T) {
	c := &cacheContentText{x: 10, y: 20, text: "hello", fontSize: 14}
	e := buildElement(0, c)
	if e.Type != ElementText || e.X != 10 || e.Y != 20 || e.Text != "hello" || e.FontSize != 14 {
		t.Errorf("text element: %+v", e)
	}
}

func TestCov3_BuildElement_Image(t *testing.T) {
	c := &cacheContentImage{x: 5, y: 15, rect: Rect{W: 100, H: 50}}
	e := buildElement(1, c)
	if e.Type != ElementImage || e.X != 5 || e.Width != 100 || e.Height != 50 {
		t.Errorf("image element: %+v", e)
	}
}

func TestCov3_BuildElement_Line(t *testing.T) {
	c := &cacheContentLine{x1: 0, y1: 0, x2: 100, y2: 100}
	e := buildElement(2, c)
	if e.Type != ElementLine || e.X != 0 || e.X2 != 100 || e.Y2 != 100 {
		t.Errorf("line element: %+v", e)
	}
}

func TestCov3_BuildElement_Rectangle(t *testing.T) {
	c := cacheContentRectangle{x: 10, y: 20, width: 80, height: 40}
	e := buildElement(3, c)
	if e.Type != ElementRectangle || e.X != 10 || e.Width != 80 || e.Height != 40 {
		t.Errorf("rect element: %+v", e)
	}
}

func TestCov3_BuildElement_Oval(t *testing.T) {
	c := &cacheContentOval{x1: 10, y1: 20, x2: 100, y2: 80}
	e := buildElement(4, c)
	if e.Type != ElementOval || e.X != 10 || e.X2 != 100 {
		t.Errorf("oval element: %+v", e)
	}
}

func TestCov3_BuildElement_Polygon(t *testing.T) {
	c := &cacheContentPolygon{points: []Point{{X: 10, Y: 20}, {X: 30, Y: 40}}}
	e := buildElement(5, c)
	if e.Type != ElementPolygon || e.X != 10 || e.Y != 20 {
		t.Errorf("polygon element: %+v", e)
	}
}

func TestCov3_BuildElement_PolygonEmpty(t *testing.T) {
	c := &cacheContentPolygon{}
	e := buildElement(5, c)
	if e.X != 0 || e.Y != 0 {
		t.Errorf("empty polygon should have 0,0: %+v", e)
	}
}

func TestCov3_BuildElement_Curve(t *testing.T) {
	c := &cacheContentCurve{x0: 10, y0: 20, x3: 100, y3: 200}
	e := buildElement(6, c)
	if e.Type != ElementCurve || e.X != 10 || e.X2 != 100 || e.Y2 != 200 {
		t.Errorf("curve element: %+v", e)
	}
}

func TestCov3_BuildElement_Polyline(t *testing.T) {
	c := &cacheContentPolyline{points: []Point{{X: 5, Y: 10}}}
	e := buildElement(7, c)
	if e.Type != ElementPolyline || e.X != 5 || e.Y != 10 {
		t.Errorf("polyline element: %+v", e)
	}
}

func TestCov3_BuildElement_PolylineEmpty(t *testing.T) {
	c := &cacheContentPolyline{}
	e := buildElement(7, c)
	if e.X != 0 || e.Y != 0 {
		t.Errorf("empty polyline should have 0,0: %+v", e)
	}
}

func TestCov3_BuildElement_Sector(t *testing.T) {
	c := &cacheContentSector{cx: 50, cy: 60, r: 30}
	e := buildElement(8, c)
	if e.Type != ElementSector || e.X != 50 || e.Y != 60 || e.Width != 30 {
		t.Errorf("sector element: %+v", e)
	}
}

func TestCov3_BuildElement_ImportedTemplate(t *testing.T) {
	c := &cacheContentImportedTemplate{tX: 100, tY: 200}
	e := buildElement(9, c)
	if e.Type != ElementImportedTemplate || e.X != 100 || e.Y != 200 {
		t.Errorf("imported template element: %+v", e)
	}
}

func TestCov3_BuildElement_LineWidth(t *testing.T) {
	c := &cacheContentLineWidth{width: 2.5}
	e := buildElement(10, c)
	if e.Type != ElementLineWidth || e.Width != 2.5 {
		t.Errorf("line width element: %+v", e)
	}
}

// ============================================================
// content_element.go - ModifyElementPosition (more branches)
// ============================================================

func TestCov3_ModifyElementPosition_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")
	err := pdf.ModifyElementPosition(1, 0, 100, 200)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Line(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 20, 100, 200)
	elems, _ := pdf.GetPageElements(1)
	lineIdx := -1
	for i, e := range elems {
		if e.Type == ElementLine {
			lineIdx = i
			break
		}
	}
	if lineIdx < 0 {
		t.Skip("no line element found")
	}
	if err := pdf.ModifyElementPosition(1, lineIdx, 50, 60); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Rect(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.InsertRectElement(1, 10, 20, 80, 40, "D")
	elems, _ := pdf.GetPageElements(1)
	rectIdx := -1
	for i, e := range elems {
		if e.Type == ElementRectangle {
			rectIdx = i
			break
		}
	}
	if rectIdx < 0 {
		t.Skip("no rect element found")
	}
	if err := pdf.ModifyElementPosition(1, rectIdx, 50, 60); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Oval(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.InsertOvalElement(1, 10, 20, 100, 80)
	elems, _ := pdf.GetPageElements(1)
	ovalIdx := -1
	for i, e := range elems {
		if e.Type == ElementOval {
			ovalIdx = i
			break
		}
	}
	if ovalIdx < 0 {
		t.Skip("no oval element found")
	}
	if err := pdf.ModifyElementPosition(1, ovalIdx, 50, 60); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Curve(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Curve(10, 20, 30, 40, 50, 60, 70, 80, "D")
	elems, _ := pdf.GetPageElements(1)
	curveIdx := -1
	for i, e := range elems {
		if e.Type == ElementCurve {
			curveIdx = i
			break
		}
	}
	if curveIdx < 0 {
		t.Skip("no curve element found")
	}
	if err := pdf.ModifyElementPosition(1, curveIdx, 100, 200); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Polygon(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pts := []Point{{X: 10, Y: 20}, {X: 50, Y: 20}, {X: 30, Y: 60}}
	pdf.Polygon(pts, "D")
	elems, _ := pdf.GetPageElements(1)
	polyIdx := -1
	for i, e := range elems {
		if e.Type == ElementPolygon {
			polyIdx = i
			break
		}
	}
	if polyIdx < 0 {
		t.Skip("no polygon element found")
	}
	if err := pdf.ModifyElementPosition(1, polyIdx, 100, 200); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(100, 100, 50, 0, 90, "FD")
	elems, _ := pdf.GetPageElements(1)
	secIdx := -1
	for i, e := range elems {
		if e.Type == ElementSector {
			secIdx = i
			break
		}
	}
	if secIdx < 0 {
		t.Skip("no sector element found")
	}
	if err := pdf.ModifyElementPosition(1, secIdx, 200, 200); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyElementPosition_Unsupported(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetLineWidth(2)
	elems, _ := pdf.GetPageElements(1)
	lwIdx := -1
	for i, e := range elems {
		if e.Type == ElementLineWidth {
			lwIdx = i
			break
		}
	}
	if lwIdx < 0 {
		t.Skip("no line width element found")
	}
	err := pdf.ModifyElementPosition(1, lwIdx, 10, 10)
	if err == nil {
		t.Error("expected error for unsupported element type")
	}
}

// ============================================================
// gopdf.go - convertNumericToFloat64
// ============================================================

func TestCov3_ConvertNumericToFloat64(t *testing.T) {
	tests := []struct {
		val  interface{}
		want float64
		ok   bool
	}{
		{float32(3.14), 3.140000104904175, true},
		{float64(2.71), 2.71, true},
		{int(42), 42, true},
		{int8(8), 8, true},
		{int16(16), 16, true},
		{int32(32), 32, true},
		{int64(64), 64, true},
		{uint(10), 10, true},
		{uint8(8), 8, true},
		{uint16(16), 16, true},
		{uint32(32), 32, true},
		{uint64(64), 64, true},
		{"string", 0, false},
		{nil, 0, false},
	}
	for _, tt := range tests {
		got, err := convertNumericToFloat64(tt.val)
		if tt.ok && err != nil {
			t.Errorf("convertNumericToFloat64(%T) unexpected error: %v", tt.val, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("convertNumericToFloat64(%T) expected error", tt.val)
		}
		if tt.ok && got != tt.want {
			t.Errorf("convertNumericToFloat64(%T) = %f, want %f", tt.val, got, tt.want)
		}
	}
}

// ============================================================
// gopdf.go - IsFitMultiCellWithNewline
// ============================================================

func TestCov3_IsFitMultiCellWithNewline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	rect := &Rect{W: 200, H: 100}
	ok, h, err := pdf.IsFitMultiCellWithNewline(rect, "Hello\nWorld")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected fit")
	}
	if h <= 0 {
		t.Error("expected positive height")
	}
}

func TestCov3_IsFitMultiCellWithNewline_TooSmall(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	rect := &Rect{W: 200, H: 1}
	ok, _, err := pdf.IsFitMultiCellWithNewline(rect, "Hello\nWorld\nLine3\nLine4\nLine5")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected not fit for tiny rect")
	}
}

// ============================================================
// pixmap_render.go - RenderOption.defaults, drawRectOnImage
// ============================================================

func TestCov3_RenderOption_Defaults(t *testing.T) {
	opt := RenderOption{}
	opt.defaults()
	if opt.DPI != 72 {
		t.Errorf("DPI = %f, want 72", opt.DPI)
	}
	if opt.Background == nil {
		t.Error("Background should not be nil")
	}
}

func TestCov3_RenderOption_CustomDPI(t *testing.T) {
	opt := RenderOption{DPI: 150, Background: color.Black}
	opt.defaults()
	if opt.DPI != 150 {
		t.Errorf("DPI = %f, want 150", opt.DPI)
	}
}

func TestCov3_DrawRectOnImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	drawRectOnImage(img, 10, 10, 50, 30, 1.0, c)
	got := img.RGBAAt(20, 10)
	if got != c {
		t.Errorf("pixel at (20,10) = %v, want %v", got, c)
	}
	got = img.RGBAAt(10, 20)
	if got != c {
		t.Errorf("pixel at (10,20) = %v, want %v", got, c)
	}
}

func TestCov3_DrawRectOnImage_OutOfBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	c := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	drawRectOnImage(img, -10, -10, 200, 200, 1.0, c)
}

// ============================================================
// pixmap_render.go - renderContentStream
// ============================================================

func TestCov3_RenderContentStream_Basic(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	stream := []byte("1 0 0 RG 10 10 80 40 re S")
	parser := &rawPDFParser{objects: map[int]rawPDFObject{}}
	page := rawPDFPage{mediaBox: [4]float64{0, 0, 200, 200}}
	renderContentStream(img, stream, parser, page, 1.0, 200)
}

func TestCov3_RenderContentStream_Colors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	stream := []byte("0.5 0.5 0.5 RG 0.2 0.3 0.4 rg 0.7 G 0.3 g 10 10 50 30 re f")
	parser := &rawPDFParser{objects: map[int]rawPDFObject{}}
	page := rawPDFPage{mediaBox: [4]float64{0, 0, 200, 200}}
	renderContentStream(img, stream, parser, page, 1.0, 200)
}

func TestCov3_RenderContentStream_CTM(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	stream := []byte("2 0 0 2 10 20 cm 10 10 50 30 re S")
	parser := &rawPDFParser{objects: map[int]rawPDFObject{}}
	page := rawPDFPage{mediaBox: [4]float64{0, 0, 200, 200}}
	renderContentStream(img, stream, parser, page, 1.0, 200)
}

func TestCov3_RenderContentStream_PathOps(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	stream := []byte("q 10 20 m 100 200 l S Q 10 10 50 30 re f 10 10 50 30 re B 10 10 50 30 re n W")
	parser := &rawPDFParser{objects: map[int]rawPDFObject{}}
	page := rawPDFPage{mediaBox: [4]float64{0, 0, 200, 200}}
	renderContentStream(img, stream, parser, page, 1.0, 200)
}

func TestCov3_RenderContentStream_FillVariants(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	stream := []byte("10 10 50 30 re F 10 10 50 30 re f* 10 10 50 30 re B* 10 10 50 30 re b 10 10 50 30 re b* 10 10 50 30 re s W*")
	parser := &rawPDFParser{objects: map[int]rawPDFObject{}}
	page := rawPDFPage{mediaBox: [4]float64{0, 0, 200, 200}}
	renderContentStream(img, stream, parser, page, 1.0, 200)
}

// ============================================================
// form_field.go - FormFieldType.String, AddFormField, etc.
// ============================================================

func TestCov3_FormFieldType_String(t *testing.T) {
	tests := []struct {
		ft   FormFieldType
		want string
	}{
		{FormFieldText, "Text"},
		{FormFieldCheckbox, "Checkbox"},
		{FormFieldRadio, "Radio"},
		{FormFieldChoice, "Choice"},
		{FormFieldButton, "Button"},
		{FormFieldSignature, "Signature"},
		{FormFieldType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.ft.String(); got != tt.want {
			t.Errorf("FormFieldType(%d).String() = %q, want %q", tt.ft, got, tt.want)
		}
	}
}

func TestCov3_AddFormField_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "username",
		X: 50, Y: 100, W: 200, H: 25,
		Value: "test", FontFamily: fontFamily, FontSize: 12,
		HasBorder: true, BorderColor: [3]uint8{0, 0, 0},
		HasFill: true, FillColor: [3]uint8{255, 255, 255},
		Multiline: true, MaxLen: 100,
		ReadOnly: true, Required: true,
		Color: [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	fields := pdf.GetFormFields()
	if len(fields) != 1 || fields[0].Name != "username" {
		t.Errorf("fields: %+v", fields)
	}
}

func TestCov3_AddFormField_Checkbox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldCheckbox, Name: "agree",
		X: 50, Y: 100, W: 20, H: 20, Checked: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddFormField_CheckboxUnchecked(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldCheckbox, Name: "agree2",
		X: 50, Y: 100, W: 20, H: 20, Checked: false,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddFormField_Radio(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldRadio, Name: "option1",
		X: 50, Y: 100, W: 20, H: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddFormField_Choice(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldChoice, Name: "country",
		X: 50, Y: 100, W: 200, H: 25,
		Options: []string{"USA", "Canada", "UK"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddFormField_Button(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldButton, Name: "submit",
		X: 50, Y: 100, W: 100, H: 30,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddFormField_Signature(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{
		Type: FormFieldSignature, Name: "sig",
		X: 50, Y: 100, W: 200, H: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddFormField_NoName(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{Type: FormFieldText, X: 50, Y: 100, W: 200, H: 25})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCov3_AddFormField_BadSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.AddFormField(FormField{Type: FormFieldText, Name: "f1", X: 50, Y: 100, W: 0, H: 25})
	if err == nil {
		t.Error("expected error for zero width")
	}
}

func TestCov3_AddTextField(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.AddTextField("name", 50, 100, 200, 25); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddCheckbox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.AddCheckbox("agree", 50, 100, 20, true); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_AddDropdown(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.AddDropdown("country", 50, 100, 200, 25, []string{"A", "B"}); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_FormFieldObj_Write(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddFormField(FormField{
		Type: FormFieldText, Name: "t1",
		X: 10, Y: 10, W: 100, H: 20,
		Value: "hello", FontSize: 14,
		HasBorder: true, BorderColor: [3]uint8{0, 0, 0},
		HasFill: true, FillColor: [3]uint8{255, 255, 255},
		Color: [3]uint8{0, 0, 0},
	})
	pdf.AddFormField(FormField{
		Type: FormFieldCheckbox, Name: "c1",
		X: 10, Y: 40, W: 20, H: 20, Checked: true,
		HasBorder: true, BorderColor: [3]uint8{0, 0, 0},
	})
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================
// ocg.go - AddOCG, GetOCGs, SetOCGState, SetOCGStates, SwitchLayer
// ============================================================

func TestCov3_AddOCG(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	ocg := pdf.AddOCG(OCG{Name: "Layer1", On: true})
	if ocg.Name != "Layer1" {
		t.Error("name mismatch")
	}
	ocgs := pdf.GetOCGs()
	if len(ocgs) != 1 {
		t.Errorf("expected 1 OCG, got %d", len(ocgs))
	}
}

func TestCov3_AddOCG_DefaultIntent(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	ocg := pdf.AddOCG(OCG{Name: "Layer2"})
	if ocg.Intent != OCGIntentView {
		t.Errorf("expected View intent, got %v", ocg.Intent)
	}
}

func TestCov3_SetOCGState(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "L1", On: true})
	if err := pdf.SetOCGState("L1", false); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_SetOCGState_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetOCGState("nonexistent", true); err == nil {
		t.Error("expected error for nonexistent OCG")
	}
}

func TestCov3_SetOCGStates(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "A", On: true})
	pdf.AddOCG(OCG{Name: "B", On: false})
	if err := pdf.SetOCGStates(map[string]bool{"A": false, "B": true}); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_SetOCGStates_Error(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SetOCGStates(map[string]bool{"missing": true}); err == nil {
		t.Error("expected error")
	}
}

func TestCov3_SwitchLayer(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.AddOCG(OCG{Name: "X", On: true})
	pdf.AddOCG(OCG{Name: "Y", On: true})
	if err := pdf.SwitchLayer("X", true); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_SwitchLayer_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.SwitchLayer("nope", false); err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// page_batch_ops.go - DeletePages, MovePage
// ============================================================

func TestCov3_DeletePages(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 5; i++ {
		pdf.AddPage()
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "page")
	}
	if err := pdf.DeletePages([]int{2, 4}); err != nil {
		t.Fatal(err)
	}
	if pdf.GetNumberOfPages() != 3 {
		t.Errorf("expected 3 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov3_DeletePages_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.DeletePages([]int{}); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_DeletePages_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.DeletePages([]int{5}); err == nil {
		t.Error("expected error for out of range")
	}
}

func TestCov3_DeletePages_All(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.DeletePages([]int{1}); err == nil {
		t.Error("expected error for deleting all pages")
	}
}

func TestCov3_DeletePages_Duplicates(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "page")
	}
	if err := pdf.DeletePages([]int{2, 2, 2}); err != nil {
		t.Fatal(err)
	}
	if pdf.GetNumberOfPages() != 2 {
		t.Errorf("expected 2 pages, got %d", pdf.GetNumberOfPages())
	}
}

func TestCov3_MovePage(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "page")
	}
	if err := pdf.MovePage(3, 1); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_MovePage_SamePos(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "page")
	if err := pdf.MovePage(1, 1); err != nil {
		t.Fatal(err)
	}
}

func TestCov3_MovePage_OutOfRange(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	if err := pdf.MovePage(0, 1); err == nil {
		t.Error("expected error")
	}
	if err := pdf.MovePage(1, 5); err == nil {
		t.Error("expected error")
	}
}

func TestCov3_MovePage_Forward(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 4; i++ {
		pdf.AddPage()
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "page")
	}
	if err := pdf.MovePage(1, 3); err != nil {
		t.Fatal(err)
	}
}

// ============================================================
// annotation_modify.go - ModifyAnnotation
// ============================================================

func TestCov3_ModifyAnnotation_NoPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.ModifyAnnotation(0, 0, AnnotationOption{Content: "test"})
	if err == nil {
		t.Error("expected error for no page")
	}
}

func TestCov3_ModifyAnnotation_BadIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	err := pdf.ModifyAnnotation(0, 99, AnnotationOption{Content: "test"})
	if err == nil {
		t.Error("expected error for bad index")
	}
}

// ============================================================
// gc.go - GarbageCollect with GCDedup
// ============================================================

func TestCov3_GarbageCollect_None(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	removed := pdf.GarbageCollect(GCNone)
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestCov3_GarbageCollect_Compact(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "page")
	}
	pdf.DeletePage(2)
	removed := pdf.GarbageCollect(GCCompact)
	if removed < 1 {
		t.Errorf("expected at least 1 removed, got %d", removed)
	}
}

func TestCov3_GarbageCollect_Dedup(t *testing.T) {
	pdf := newPDFWithFont(t)
	for i := 0; i < 3; i++ {
		pdf.AddPage()
		pdf.SetXY(50, 50)
		pdf.Cell(nil, "page")
	}
	pdf.DeletePage(2)
	removed := pdf.GarbageCollect(GCDedup)
	if removed < 1 {
		t.Errorf("expected at least 1 removed, got %d", removed)
	}
}

func TestCov3_GarbageCollect_NoNulls(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	removed := pdf.GarbageCollect(GCCompact)
	if removed != 0 {
		t.Errorf("expected 0 removed (no nulls), got %d", removed)
	}
}

func TestCov3_GetLiveObjectCount(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	count := pdf.GetLiveObjectCount()
	total := pdf.GetObjectCount()
	if count <= 0 || count > total {
		t.Errorf("live=%d total=%d", count, total)
	}
}

// ============================================================
// html_insert.go - collapseWhitespace, splitWords, renderLongWord
// ============================================================

func TestCov3_CollapseWhitespace(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"  hello   world  ", "hello world"},
		{"no\nnewlines\there", "no newlines here"},
		{"", ""},
		{"single", "single"},
	}
	for _, tt := range tests {
		got := collapseWhitespace(tt.in)
		if got != tt.want {
			t.Errorf("collapseWhitespace(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCov3_SplitWords(t *testing.T) {
	words := splitWords("hello world  foo")
	if len(words) != 3 {
		t.Errorf("expected 3 words, got %d: %v", len(words), words)
	}
}

func TestCov3_SplitWords_Empty(t *testing.T) {
	words := splitWords("")
	if len(words) != 0 {
		t.Errorf("expected 0 words, got %d", len(words))
	}
}

func TestCov3_InsertHTMLBox_LongWord(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	longWord := strings.Repeat("W", 200)
	_, err := pdf.InsertHTMLBox(50, 50, 100, 500, longWord, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_InsertHTMLBox_Image(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	html := `<img src="nonexistent.jpg" width="100" height="50"/>`
	pdf.InsertHTMLBox(50, 50, 300, 300, html, HTMLBoxOption{
		DefaultFontFamily: fontFamily,
	})
}

// ============================================================
// gopdf.go - misc methods
// ============================================================

func TestCov3_Polyline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pts := []Point{{X: 10, Y: 10}, {X: 50, Y: 50}, {X: 90, Y: 10}}
	pdf.Polyline(pts)
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov3_Sector(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Sector(100, 100, 50, 0, 90, "FD")
	pdf.Sector(100, 100, 50, 0, 360, "D")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov3_Rectangle_Rounded(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Rectangle(50, 50, 200, 150, "FD", 10, 12)
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov3_SplitTextWithWordWrap_Short(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	lines, err := pdf.SplitTextWithWordWrap("hi", 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestCov3_SplitTextWithWordWrap_LongWord(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	long := strings.Repeat("X", 200)
	lines, err := pdf.SplitTextWithWordWrap(long, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

func TestCov3_AddHeader_AddFooter(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddHeader(func() {
		pdf.SetXY(50, 10)
		pdf.Cell(nil, "Header")
	})
	pdf.AddFooter(func() {
		pdf.SetXY(50, 800)
		pdf.Cell(nil, "Footer")
	})
	pdf.AddPage()
	pdf.SetXY(50, 100)
	pdf.Cell(nil, "Body")
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov3_SetInfo_GetInfo(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetInfo(PdfInfo{
		Title: "Test", Author: "Author",
		Subject: "Subject", Creator: "Creator", Producer: "Producer",
	})
	info := pdf.GetInfo()
	if info.Title != "Test" || info.Author != "Author" {
		t.Errorf("info: %+v", info)
	}
}

func TestCov3_SetNoCompression(t *testing.T) {
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	pdf.SetNoCompression()
	pdf.AddPage()
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov3_SetCharSpacing(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetCharSpacing(2.0)
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "spaced")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	if buf.Len() == 0 {
		t.Error("expected output")
	}
}

func TestCov3_GetNextObjectID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	id := pdf.GetNextObjectID()
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

// ============================================================
// content_element.go - ReplaceElement, InsertElementAt, ModifyTextElement
// ============================================================

func TestCov3_ReplaceElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 20, 100, 200)
	newLine := &cacheContentLine{x1: 50, y1: 60, x2: 150, y2: 260}
	_ = pdf.ReplaceElement(1, 0, newLine)
}

func TestCov3_ReplaceElement_BadPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.ReplaceElement(99, 0, &cacheContentLine{})
	if err == nil {
		t.Error("expected error for bad page")
	}
}

func TestCov3_InsertElementAt(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.Line(10, 20, 100, 200)
	newLine := &cacheContentLine{x1: 50, y1: 60, x2: 150, y2: 260}
	_ = pdf.InsertElementAt(1, 0, newLine)
}

func TestCov3_InsertElementAt_BadPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.InsertElementAt(99, 0, &cacheContentLine{})
	if err == nil {
		t.Error("expected error for bad page")
	}
}

func TestCov3_ModifyTextElement(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "original")
	err := pdf.ModifyTextElement(1, 0, "modified")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCov3_ModifyTextElement_BadPage(t *testing.T) {
	pdf := newPDFWithFont(t)
	err := pdf.ModifyTextElement(99, 0, "test")
	if err == nil {
		t.Error("expected error")
	}
}

// ============================================================
// pixmap_render.go - RenderPageToImage (integration)
// ============================================================

func TestCov3_RenderPageToImage_InvalidPDF(t *testing.T) {
	_, err := RenderPageToImage([]byte("not a pdf"), 0, RenderOption{})
	if err == nil {
		t.Error("expected error for invalid PDF")
	}
}

func TestCov3_RenderPageToImage_BadIndex(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	pdf.Cell(nil, "test")
	var buf bytes.Buffer
	pdf.WriteTo(&buf)
	_, err := RenderPageToImage(buf.Bytes(), 99, RenderOption{})
	if err == nil {
		t.Error("expected error for bad page index")
	}
}

func TestCov3_RenderAllPagesToImages_InvalidPDF(t *testing.T) {
	// RenderAllPagesToImages may or may not error on invalid data
	// depending on how the parser handles it. Just ensure no panic.
	_, _ = RenderAllPagesToImages([]byte("not a pdf"), RenderOption{})
}

// Unused import guard
var _ = math.Pi
var _ = image.Black
var _ = strings.Contains
var _ = color.Black
