package gopdf

import (
	"strconv"
	"strings"
)

// HTMLBoxOption configures the behavior of InsertHTMLBox.
type HTMLBoxOption struct {
	// DefaultFontFamily is the font family used when no font is specified in HTML.
	// This font must already be added to the GoPdf instance via AddTTFFont.
	DefaultFontFamily string

	// DefaultFontSize is the default font size in points.
	DefaultFontSize float64

	// DefaultColor is the default text color (r, g, b).
	DefaultColor [3]uint8

	// LineSpacing is extra spacing between lines (in document units). Default is 0.
	LineSpacing float64

	// BoldFontFamily is the font family to use for bold text.
	// If empty, the default family is used (bold may not render if the font doesn't support it).
	BoldFontFamily string

	// ItalicFontFamily is the font family to use for italic text.
	ItalicFontFamily string

	// BoldItalicFontFamily is the font family to use for bold+italic text.
	BoldItalicFontFamily string
}

// htmlRenderState tracks the current rendering state while walking the HTML tree.
type htmlRenderState struct {
	fontFamily string
	fontSize   float64
	fontStyle  int // Regular, Bold, Italic, Underline
	colorR     uint8
	colorG     uint8
	colorB     uint8
	align      int
}

// htmlRenderer handles the rendering of HTML nodes into the PDF.
type htmlRenderer struct {
	gp      *GoPdf
	opt     HTMLBoxOption
	boxX    float64 // box left edge (units)
	boxY    float64 // box top edge (units)
	boxW    float64 // box width (units)
	boxH    float64 // box height (units)
	cursorX float64 // current X position (units)
	cursorY float64 // current Y position (units)
}

// InsertHTMLBox renders simplified HTML content into a rectangular area on the PDF.
//
// Supported HTML tags:
//   - <b>, <strong>: Bold text
//   - <i>, <em>: Italic text
//   - <u>: Underlined text
//   - <s>, <strike>, <del>: Strikethrough (rendered as underline for simplicity)
//   - <br>, <br/>: Line break
//   - <p>: Paragraph (adds vertical spacing)
//   - <h1> to <h6>: Headings with automatic font sizing
//   - <font color="..." size="..." face="...">: Font styling
//   - <span style="...">: Inline styling (color, font-size, font-family)
//   - <img src="..." width="..." height="...">: Images (src must be a local file path)
//   - <hr>: Horizontal rule
//   - <center>: Centered text
//   - <ul>, <ol>, <li>: Lists (basic bullet/number)
//   - <a href="...">: Links (rendered as colored text, link annotation added)
//   - <sub>, <sup>: Subscript/superscript (approximated with smaller font)
//
// Parameters:
//   - x, y: Top-left corner of the box (in document units)
//   - w, h: Width and height of the box (in document units)
//   - htmlStr: The HTML string to render
//   - opt: Rendering options (font families, default size, colors, etc.)
//
// Returns the Y position after the last rendered content (in document units).
func (gp *GoPdf) InsertHTMLBox(x, y, w, h float64, htmlStr string, opt HTMLBoxOption) (float64, error) {
	if opt.DefaultFontSize <= 0 {
		opt.DefaultFontSize = 12
	}
	if opt.DefaultFontFamily == "" {
		return y, ErrMissingFontFamily
	}

	nodes := parseHTML(htmlStr)

	r := &htmlRenderer{
		gp:      gp,
		opt:     opt,
		boxX:    x,
		boxY:    y,
		boxW:    w,
		boxH:    h,
		cursorX: x,
		cursorY: y,
	}

	state := htmlRenderState{
		fontFamily: opt.DefaultFontFamily,
		fontSize:   opt.DefaultFontSize,
		fontStyle:  Regular,
		colorR:     opt.DefaultColor[0],
		colorG:     opt.DefaultColor[1],
		colorB:     opt.DefaultColor[2],
		align:      Left,
	}

	if err := r.renderNodes(nodes, state); err != nil {
		return r.cursorY, err
	}

	return r.cursorY, nil
}

func (r *htmlRenderer) renderNodes(nodes []*htmlNode, state htmlRenderState) error {
	for _, node := range nodes {
		if r.cursorY-r.boxY >= r.boxH {
			break // exceeded box height
		}
		if err := r.renderNode(node, state); err != nil {
			return err
		}
	}
	return nil
}

func (r *htmlRenderer) renderNode(node *htmlNode, state htmlRenderState) error {
	if node.Type == htmlNodeText {
		return r.renderText(node.Text, state)
	}

	// element node
	newState := state

	switch node.Tag {
	case "b", "strong":
		newState.fontStyle |= Bold
	case "i", "em":
		newState.fontStyle |= Italic
	case "u", "ins":
		newState.fontStyle |= Underline
	case "s", "strike", "del":
		// approximate strikethrough with underline
		newState.fontStyle |= Underline
	case "br":
		r.newLine(state)
		return nil
	case "hr":
		return r.renderHR(state)
	case "p", "div":
		newState = r.applyStyleAttr(node, newState)
		if r.cursorX > r.boxX {
			r.newLine(state)
		}
		r.addVerticalSpace(state.fontSize * 0.3)
		if err := r.renderNodes(node.Children, newState); err != nil {
			return err
		}
		if r.cursorX > r.boxX {
			r.newLine(state)
		}
		r.addVerticalSpace(state.fontSize * 0.3)
		return nil
	case "h1", "h2", "h3", "h4", "h5", "h6":
		newState.fontSize = headingFontSize(node.Tag)
		newState.fontStyle |= Bold
		newState = r.applyStyleAttr(node, newState)
		if r.cursorX > r.boxX {
			r.newLine(state)
		}
		r.addVerticalSpace(newState.fontSize * 0.4)
		if err := r.renderNodes(node.Children, newState); err != nil {
			return err
		}
		if r.cursorX > r.boxX {
			r.newLine(newState)
		}
		r.addVerticalSpace(newState.fontSize * 0.3)
		return nil
	case "font":
		if color, ok := node.Attrs["color"]; ok {
			if cr, cg, cb, cok := parseCSSColor(color); cok {
				newState.colorR, newState.colorG, newState.colorB = cr, cg, cb
			}
		}
		if size, ok := node.Attrs["size"]; ok {
			if sz, sok := parseFontSizeAttr(size); sok {
				newState.fontSize = sz
			}
		}
		if face, ok := node.Attrs["face"]; ok {
			newState.fontFamily = face
		}
	case "span":
		newState = r.applyStyleAttr(node, newState)
	case "center":
		newState.align = Center
		if r.cursorX > r.boxX {
			r.newLine(state)
		}
		if err := r.renderNodes(node.Children, newState); err != nil {
			return err
		}
		if r.cursorX > r.boxX {
			r.newLine(newState)
		}
		return nil
	case "a":
		// render link text in blue with underline, then add PDF link annotation
		newState.colorR, newState.colorG, newState.colorB = 0, 0, 255
		newState.fontStyle |= Underline
		href := node.Attrs["href"]

		// record position before rendering link text
		startX := r.cursorX
		startY := r.cursorY
		lh := r.lineHeight(newState)

		if err := r.renderNodes(node.Children, newState); err != nil {
			return err
		}

		// add clickable link annotation if href is present
		if href != "" {
			endX := r.cursorX
			// convert to points for the annotation
			ax := r.gp.UnitsToPoints(startX)
			ay := r.gp.UnitsToPoints(startY)
			aw := r.gp.UnitsToPoints(endX - startX)
			ah := r.gp.UnitsToPoints(lh)
			if aw > 0 {
				r.gp.AddExternalLink(href, ax, ay, aw, ah)
			}
		}
		return nil
	case "img":
		return r.renderImage(node, state)
	case "ul", "ol":
		return r.renderList(node, state, node.Tag == "ol")
	case "li":
		// handled by renderList
	case "sub":
		newState.fontSize = state.fontSize * 0.7
	case "sup":
		newState.fontSize = state.fontSize * 0.7
	case "blockquote":
		newState = r.applyStyleAttr(node, newState)
		if r.cursorX > r.boxX {
			r.newLine(state)
		}
		oldBoxX := r.boxX
		oldBoxW := r.boxW
		indent := state.fontSize * 1.5
		r.boxX += indent
		r.boxW -= indent
		r.cursorX = r.boxX
		r.addVerticalSpace(state.fontSize * 0.3)
		if err := r.renderNodes(node.Children, newState); err != nil {
			return err
		}
		if r.cursorX > r.boxX {
			r.newLine(newState)
		}
		r.addVerticalSpace(state.fontSize * 0.3)
		r.boxX = oldBoxX
		r.boxW = oldBoxW
		r.cursorX = r.boxX
		return nil
	}

	// render children with updated state
	if err := r.renderNodes(node.Children, newState); err != nil {
		return err
	}

	return nil
}

func (r *htmlRenderer) applyStyleAttr(node *htmlNode, state htmlRenderState) htmlRenderState {
	styleStr, ok := node.Attrs["style"]
	if !ok {
		return state
	}
	styles := parseInlineStyle(styleStr)

	if color, ok := styles["color"]; ok {
		if cr, cg, cb, cok := parseCSSColor(color); cok {
			state.colorR, state.colorG, state.colorB = cr, cg, cb
		}
	}
	if fs, ok := styles["font-size"]; ok {
		if sz, sok := parseFontSize(fs, state.fontSize); sok {
			state.fontSize = sz
		}
	}
	if ff, ok := styles["font-family"]; ok {
		state.fontFamily = strings.Trim(ff, "'\"")
	}
	if fw, ok := styles["font-weight"]; ok {
		if fw == "bold" || fw == "700" || fw == "800" || fw == "900" {
			state.fontStyle |= Bold
		}
	}
	if fst, ok := styles["font-style"]; ok {
		if fst == "italic" {
			state.fontStyle |= Italic
		}
	}
	if td, ok := styles["text-decoration"]; ok {
		if strings.Contains(td, "underline") {
			state.fontStyle |= Underline
		}
	}
	if ta, ok := styles["text-align"]; ok {
		switch ta {
		case "center":
			state.align = Center
		case "right":
			state.align = Right
		case "left":
			state.align = Left
		}
	}
	return state
}

func (r *htmlRenderer) applyFont(state htmlRenderState) error {
	family := r.resolveFontFamily(state)
	style := state.fontStyle &^ Underline // strip underline for font lookup
	if err := r.gp.SetFontWithStyle(family, style, state.fontSize); err != nil {
		// fallback to default family
		if err2 := r.gp.SetFontWithStyle(r.opt.DefaultFontFamily, Regular, state.fontSize); err2 != nil {
			return err
		}
	}
	r.gp.SetTextColor(state.colorR, state.colorG, state.colorB)
	return nil
}

func (r *htmlRenderer) resolveFontFamily(state htmlRenderState) string {
	isBold := state.fontStyle&Bold == Bold
	isItalic := state.fontStyle&Italic == Italic

	if isBold && isItalic && r.opt.BoldItalicFontFamily != "" {
		return r.opt.BoldItalicFontFamily
	}
	if isBold && r.opt.BoldFontFamily != "" {
		return r.opt.BoldFontFamily
	}
	if isItalic && r.opt.ItalicFontFamily != "" {
		return r.opt.ItalicFontFamily
	}
	if state.fontFamily != "" {
		return state.fontFamily
	}
	return r.opt.DefaultFontFamily
}

func (r *htmlRenderer) lineHeight(state htmlRenderState) float64 {
	return state.fontSize * 1.2 / r.unitConversion() + r.opt.LineSpacing
}

func (r *htmlRenderer) unitConversion() float64 {
	// points per unit
	switch r.gp.config.Unit {
	case UnitMM:
		return conversionUnitMM
	case UnitCM:
		return conversionUnitCM
	case UnitIN:
		return conversionUnitIN
	case UnitPX:
		return conversionUnitPX
	default:
		return 1.0
	}
}

func (r *htmlRenderer) newLine(state htmlRenderState) {
	r.cursorX = r.boxX
	r.cursorY += r.lineHeight(state)
}

func (r *htmlRenderer) addVerticalSpace(ptSpace float64) {
	r.cursorY += ptSpace / r.unitConversion()
}

func (r *htmlRenderer) remainingWidth() float64 {
	return r.boxX + r.boxW - r.cursorX
}

func (r *htmlRenderer) renderText(text string, state htmlRenderState) error {
	if err := r.applyFont(state); err != nil {
		return err
	}

	// collapse whitespace
	text = collapseWhitespace(text)
	if text == "" {
		return nil
	}

	words := splitWords(text)
	lh := r.lineHeight(state)
	spaceWidth, err := r.gp.MeasureTextWidth(" ")
	if err != nil {
		return err
	}

	for i, word := range words {
		if r.cursorY-r.boxY+lh > r.boxH {
			break // exceeded box height
		}

		wordWidth, err := r.gp.MeasureTextWidth(word)
		if err != nil {
			return err
		}

		// check if word fits on current line
		if r.cursorX > r.boxX && r.cursorX+wordWidth > r.boxX+r.boxW {
			r.newLine(state)
		}

		// if a single word is wider than the box, force-render it
		if wordWidth > r.boxW {
			// render character by character with wrapping
			if err := r.renderLongWord(word, state); err != nil {
				return err
			}
			continue
		}

		// handle alignment for new lines
		if r.cursorX == r.boxX && state.align == Center {
			lineWidth := r.measureLineWidth(words[i:], spaceWidth, state)
			if lineWidth < r.boxW {
				r.cursorX = r.boxX + (r.boxW-lineWidth)/2
			}
		} else if r.cursorX == r.boxX && state.align == Right {
			lineWidth := r.measureLineWidth(words[i:], spaceWidth, state)
			if lineWidth < r.boxW {
				r.cursorX = r.boxX + r.boxW - lineWidth
			}
		}

		r.gp.SetXY(r.cursorX, r.cursorY)

		cellOpt := CellOption{
			Align: Left | Top,
		}

		rect := &Rect{W: wordWidth, H: lh}
		if err := r.gp.CellWithOption(rect, word, cellOpt); err != nil {
			return err
		}

		r.cursorX += wordWidth

		// add space after word (except last)
		if i < len(words)-1 {
			r.cursorX += spaceWidth
		}
	}

	return nil
}

func (r *htmlRenderer) renderLongWord(word string, state htmlRenderState) error {
	lh := r.lineHeight(state)
	for _, ch := range word {
		s := string(ch)
		chWidth, err := r.gp.MeasureTextWidth(s)
		if err != nil {
			return err
		}
		if r.cursorX+chWidth > r.boxX+r.boxW && r.cursorX > r.boxX {
			r.newLine(state)
		}
		r.gp.SetXY(r.cursorX, r.cursorY)
		rect := &Rect{W: chWidth, H: lh}
		if err := r.gp.CellWithOption(rect, s, CellOption{Align: Left | Top}); err != nil {
			return err
		}
		r.cursorX += chWidth
	}
	return nil
}

func (r *htmlRenderer) measureLineWidth(words []string, spaceWidth float64, state htmlRenderState) float64 {
	total := 0.0
	for i, word := range words {
		ww, err := r.gp.MeasureTextWidth(word)
		if err != nil {
			break
		}
		if total+ww > r.boxW {
			break
		}
		total += ww
		if i < len(words)-1 {
			total += spaceWidth
		}
	}
	return total
}

func (r *htmlRenderer) renderHR(state htmlRenderState) error {
	if r.cursorX > r.boxX {
		r.newLine(state)
	}
	r.addVerticalSpace(state.fontSize * 0.3)

	y := r.cursorY + r.lineHeight(state)*0.5
	r.gp.SetStrokeColor(128, 128, 128)
	r.gp.SetLineWidth(0.5)
	r.gp.Line(r.boxX, y, r.boxX+r.boxW, y)

	r.cursorY = y + r.lineHeight(state)*0.5
	r.cursorX = r.boxX
	r.addVerticalSpace(state.fontSize * 0.3)
	return nil
}

func (r *htmlRenderer) renderImage(node *htmlNode, state htmlRenderState) error {
	src, ok := node.Attrs["src"]
	if !ok || src == "" {
		return nil
	}

	// parse dimensions
	imgW := 0.0
	imgH := 0.0
	if w, ok := node.Attrs["width"]; ok {
		if v, vok := parseDimension(w, r.boxW); vok {
			imgW = v / r.unitConversion()
		}
	}
	if h, ok := node.Attrs["height"]; ok {
		if v, vok := parseDimension(h, r.boxH); vok {
			imgH = v / r.unitConversion()
		}
	}

	// default size if not specified
	if imgW == 0 && imgH == 0 {
		imgW = r.boxW * 0.5
		imgH = imgW * 0.75
	} else if imgW == 0 {
		imgW = imgH * 1.33
	} else if imgH == 0 {
		imgH = imgW * 0.75
	}

	// clamp to box width
	if imgW > r.boxW {
		ratio := r.boxW / imgW
		imgW = r.boxW
		imgH *= ratio
	}

	// check if image fits on current line, if not, new line
	if r.cursorX > r.boxX && r.cursorX+imgW > r.boxX+r.boxW {
		r.newLine(state)
	}

	// check if image fits vertically
	if r.cursorY-r.boxY+imgH > r.boxH {
		return nil // skip image if it doesn't fit
	}

	imgHolder, err := ImageHolderByPath(src)
	if err != nil {
		return err
	}

	rect := &Rect{W: imgW, H: imgH}
	if err := r.gp.ImageByHolder(imgHolder, r.cursorX, r.cursorY, rect); err != nil {
		return err
	}

	r.cursorY += imgH
	r.cursorX = r.boxX
	return nil
}

func (r *htmlRenderer) renderList(node *htmlNode, state htmlRenderState, ordered bool) error {
	if r.cursorX > r.boxX {
		r.newLine(state)
	}
	r.addVerticalSpace(state.fontSize * 0.2)

	indent := state.fontSize * 1.2 / r.unitConversion()
	counter := 0

	for _, child := range node.Children {
		if child.Type != htmlNodeElement || child.Tag != "li" {
			continue
		}
		counter++

		if r.cursorY-r.boxY+r.lineHeight(state) > r.boxH {
			break
		}

		// render bullet or number
		if err := r.applyFont(state); err != nil {
			return err
		}

		r.cursorX = r.boxX + indent*0.5
		r.gp.SetXY(r.cursorX, r.cursorY)

		var marker string
		if ordered {
			marker = string(rune('0'+counter)) + ". "
			if counter > 9 {
				marker = strconv.Itoa(counter) + ". "
			}
		} else {
			marker = "• "
		}

		markerWidth, _ := r.gp.MeasureTextWidth(marker)
		rect := &Rect{W: markerWidth, H: r.lineHeight(state)}
		if err := r.gp.CellWithOption(rect, marker, CellOption{Align: Left | Top}); err != nil {
			return err
		}

		// render list item content with indent
		oldBoxX := r.boxX
		oldBoxW := r.boxW
		r.boxX += indent
		r.boxW -= indent
		r.cursorX = r.boxX

		childState := r.applyStyleAttr(child, state)
		if err := r.renderNodes(child.Children, childState); err != nil {
			return err
		}

		if r.cursorX > r.boxX {
			r.newLine(state)
		}

		r.boxX = oldBoxX
		r.boxW = oldBoxW
		r.cursorX = r.boxX
	}

	r.addVerticalSpace(state.fontSize * 0.2)
	return nil
}

// collapseWhitespace collapses consecutive whitespace characters into a single space.
func collapseWhitespace(s string) string {
	var result strings.Builder
	inSpace := false
	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if !inSpace {
				result.WriteByte(' ')
				inSpace = true
			}
		} else {
			result.WriteRune(ch)
			inSpace = false
		}
	}
	return strings.TrimSpace(result.String())
}

// splitWords splits text into words by spaces.
func splitWords(text string) []string {
	return strings.Fields(text)
}
