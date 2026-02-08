package gopdf

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

// ============================================================
// Text extraction from existing PDF files
// ============================================================

// ExtractTextFromPage extracts text from a specific page (0-based) of
// the given PDF data. Returns a list of ExtractedText items with
// position, font, and text content.
//
// This is a pure-Go PDF content stream parser that handles the most
// common text operators: BT/ET, Tf, Td, TD, Tm, T*, Tj, TJ, ', ".
//
// Example:
//
//	data, _ := os.ReadFile("input.pdf")
//	texts, _ := gopdf.ExtractTextFromPage(data, 0)
//	for _, t := range texts {
//	    fmt.Printf("(%0.f,%0.f) %s\n", t.X, t.Y, t.Text)
//	}
func ExtractTextFromPage(pdfData []byte, pageIndex int) ([]ExtractedText, error) {
	parser, err := newRawPDFParser(pdfData)
	if err != nil {
		return nil, err
	}
	if pageIndex < 0 || pageIndex >= len(parser.pages) {
		return nil, fmt.Errorf("page index %d out of range (0..%d)", pageIndex, len(parser.pages)-1)
	}
	stream := parser.getPageContentStream(pageIndex)
	if len(stream) == 0 {
		return nil, nil
	}
	page := parser.pages[pageIndex]
	fonts := buildFontMap(parser, page)
	return parseTextOperators(stream, fonts, page.mediaBox), nil
}

// ExtractTextFromAllPages extracts text from all pages.
func ExtractTextFromAllPages(pdfData []byte) (map[int][]ExtractedText, error) {
	parser, err := newRawPDFParser(pdfData)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]ExtractedText, len(parser.pages))
	for i := range parser.pages {
		stream := parser.getPageContentStream(i)
		if len(stream) == 0 {
			continue
		}
		page := parser.pages[i]
		fonts := buildFontMap(parser, page)
		texts := parseTextOperators(stream, fonts, page.mediaBox)
		if len(texts) > 0 {
			result[i] = texts
		}
	}
	return result, nil
}

// ExtractPageText extracts all text from a page as a single string.
// Convenience wrapper around ExtractTextFromPage.
func ExtractPageText(pdfData []byte, pageIndex int) (string, error) {
	texts, err := ExtractTextFromPage(pdfData, pageIndex)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	lastY := math.MaxFloat64
	for _, t := range texts {
		if math.Abs(t.Y-lastY) > 2 && lastY != math.MaxFloat64 {
			sb.WriteByte('\n')
		} else if sb.Len() > 0 && lastY != math.MaxFloat64 {
			sb.WriteByte(' ')
		}
		sb.WriteString(t.Text)
		lastY = t.Y
	}
	return sb.String(), nil
}

// fontInfo holds decoded font information for text extraction.
type fontInfo struct {
	name     string // resource name like /F1
	baseFont string
	encoding string
	toUni    map[uint16]rune // CMap: character code -> unicode
	isType0  bool
}

func buildFontMap(parser *rawPDFParser, page rawPDFPage) map[string]*fontInfo {
	fonts := make(map[string]*fontInfo)
	for name, objNum := range page.resources.fonts {
		fi := &fontInfo{name: name}
		obj, ok := parser.objects[objNum]
		if ok {
			fi.baseFont = extractName(obj.dict, "/BaseFont")
			fi.encoding = extractName(obj.dict, "/Encoding")
			fi.isType0 = strings.Contains(obj.dict, "/Type0") ||
				strings.Contains(obj.dict, "/Identity-H") ||
				strings.Contains(obj.dict, "/Identity-V")
			// Try to parse /ToUnicode CMap
			toUniRef := extractRef(obj.dict, "/ToUnicode")
			if toUniRef > 0 {
				if cmapObj, ok2 := parser.objects[toUniRef]; ok2 && cmapObj.stream != nil {
					fi.toUni = parseCMap(cmapObj.stream)
				}
			}
		}
		fonts[name] = fi
	}
	return fonts
}

// extractName extracts a /Name value from a dict string.
func extractName(dict, key string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(key) + `\s+/(\S+)`)
	m := re.FindStringSubmatch(dict)
	if m != nil {
		return m[1]
	}
	return ""
}

// parseCMap parses a ToUnicode CMap stream to build a code->rune mapping.
func parseCMap(data []byte) map[uint16]rune {
	m := make(map[uint16]rune)
	s := string(data)

	// Parse beginbfchar ... endbfchar sections
	reChar := regexp.MustCompile(`(?s)beginbfchar\s*(.*?)\s*endbfchar`)
	for _, match := range reChar.FindAllStringSubmatch(s, -1) {
		lines := strings.Split(strings.TrimSpace(match[1]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			parts := extractHexPairs(line)
			if len(parts) >= 2 {
				code := parseHex16(parts[0])
				uni := parseHex16(parts[1])
				m[code] = rune(uni)
			}
		}
	}

	// Parse beginbfrange ... endbfrange sections
	reRange := regexp.MustCompile(`(?s)beginbfrange\s*(.*?)\s*endbfrange`)
	for _, match := range reRange.FindAllStringSubmatch(s, -1) {
		lines := strings.Split(strings.TrimSpace(match[1]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			parts := extractHexPairs(line)
			if len(parts) >= 3 {
				start := parseHex16(parts[0])
				end := parseHex16(parts[1])
				uniStart := parseHex16(parts[2])
				for code := start; code <= end; code++ {
					m[code] = rune(uniStart + (code - start))
				}
			}
		}
	}

	return m
}

func extractHexPairs(line string) []string {
	re := regexp.MustCompile(`<([0-9a-fA-F]+)>`)
	matches := re.FindAllStringSubmatch(line, -1)
	var result []string
	for _, m := range matches {
		result = append(result, m[1])
	}
	return result
}

func parseHex16(s string) uint16 {
	v, _ := strconv.ParseUint(s, 16, 32)
	return uint16(v)
}

// parseTextOperators parses PDF content stream text operators.
func parseTextOperators(stream []byte, fonts map[string]*fontInfo, mediaBox [4]float64) []ExtractedText {
	tokens := tokenize(stream)
	var results []ExtractedText

	// Text state
	var (
		inText   bool
		curFont  *fontInfo
		fontSize float64
		// Text matrix components
		tmX, tmY     float64
		// Line matrix
		tlmX, tlmY   float64
		// CTM (simplified — only tracking translation)
		ctmA, ctmD   float64 = 1, 1
		ctmE, ctmF   float64
		tl            float64 // text leading
	)
	pageHeight := mediaBox[3] - mediaBox[1]

	stack := make([]float64, 0, 64)
	pushNum := func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			stack = append(stack, v)
		}
	}
	popN := func(n int) []float64 {
		if len(stack) < n {
			r := make([]float64, n)
			stack = stack[:0]
			return r
		}
		r := make([]float64, n)
		copy(r, stack[len(stack)-n:])
		stack = stack[:len(stack)-n]
		return r
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		// Try to parse as number
		if _, err := strconv.ParseFloat(tok, 64); err == nil {
			pushNum(tok)
			continue
		}

		switch tok {
		case "BT":
			inText = true
			tmX, tmY = 0, 0
			tlmX, tlmY = 0, 0
		case "ET":
			inText = false
			stack = stack[:0]
		case "Tf":
			if !inText {
				continue
			}
			args := popN(1)
			fontSize = args[0]
			// font name is the token before the size number
			// We need to look back in the original tokens
			if i >= 2 {
				fname := tokens[i-2]
				if strings.HasPrefix(fname, "/") {
					if f, ok := fonts[fname]; ok {
						curFont = f
					}
				}
			}
		case "Td":
			if !inText {
				continue
			}
			args := popN(2)
			tlmX += args[0]
			tlmY += args[1]
			tmX, tmY = tlmX, tlmY
		case "TD":
			if !inText {
				continue
			}
			args := popN(2)
			tl = -args[1]
			tlmX += args[0]
			tlmY += args[1]
			tmX, tmY = tlmX, tlmY
		case "Tm":
			if !inText {
				continue
			}
			args := popN(6)
			// args = [a b c d e f]
			tmX = args[4]
			tmY = args[5]
			tlmX, tlmY = tmX, tmY
			if args[0] != 0 {
				fontSize = math.Abs(args[0])
			}
		case "T*":
			if !inText {
				continue
			}
			tlmX += 0
			tlmY -= tl
			tmX, tmY = tlmX, tlmY
		case "TL":
			args := popN(1)
			tl = args[0]
		case "Tj":
			if !inText {
				continue
			}
			// String is the previous token (in parentheses or hex)
			if i >= 1 {
				raw := findStringBefore(tokens, i)
				text := decodeTextString(raw, curFont)
				if text != "" {
					x := ctmA*tmX + ctmE
					y := pageHeight - (ctmD*tmY + ctmF)
					results = append(results, ExtractedText{
						Text:     text,
						X:        x,
						Y:        y,
						FontName: fontDisplayName(curFont),
						FontSize: fontSize,
					})
				}
			}
		case "TJ":
			if !inText {
				continue
			}
			// TJ takes an array: [(string) num (string) ...]
			arr := findArrayBefore(tokens, i)
			var combined strings.Builder
			for _, item := range arr {
				if isStringToken(item) {
					combined.WriteString(decodeTextString(item, curFont))
				}
			}
			text := combined.String()
			if text != "" {
				x := ctmA*tmX + ctmE
				y := pageHeight - (ctmD*tmY + ctmF)
				results = append(results, ExtractedText{
					Text:     text,
					X:        x,
					Y:        y,
					FontName: fontDisplayName(curFont),
					FontSize: fontSize,
				})
			}
		case "'":
			// Move to next line and show string
			if !inText {
				continue
			}
			tlmY -= tl
			tmX, tmY = tlmX, tlmY
			if i >= 1 {
				raw := findStringBefore(tokens, i)
				text := decodeTextString(raw, curFont)
				if text != "" {
					x := ctmA*tmX + ctmE
					y := pageHeight - (ctmD*tmY + ctmF)
					results = append(results, ExtractedText{
						Text: text, X: x, Y: y,
						FontName: fontDisplayName(curFont), FontSize: fontSize,
					})
				}
			}
		case "\"":
			// aw ac string " — set word/char spacing, move to next line, show
			if !inText {
				continue
			}
			popN(2) // aw, ac
			tlmY -= tl
			tmX, tmY = tlmX, tlmY
			if i >= 1 {
				raw := findStringBefore(tokens, i)
				text := decodeTextString(raw, curFont)
				if text != "" {
					x := ctmA*tmX + ctmE
					y := pageHeight - (ctmD*tmY + ctmF)
					results = append(results, ExtractedText{
						Text: text, X: x, Y: y,
						FontName: fontDisplayName(curFont), FontSize: fontSize,
					})
				}
			}
		case "cm":
			args := popN(6)
			ctmA = args[0]
			ctmD = args[3]
			ctmE = args[4]
			ctmF = args[5]
		}
	}
	return results
}

// tokenize splits a PDF content stream into tokens.
func tokenize(data []byte) []string {
	var tokens []string
	i := 0
	n := len(data)
	for i < n {
		// skip whitespace
		if data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n' {
			i++
			continue
		}
		// comment
		if data[i] == '%' {
			for i < n && data[i] != '\n' && data[i] != '\r' {
				i++
			}
			continue
		}
		// string literal (...)
		if data[i] == '(' {
			depth := 1
			start := i
			i++
			for i < n && depth > 0 {
				if data[i] == '\\' {
					i += 2
					continue
				}
				if data[i] == '(' {
					depth++
				} else if data[i] == ')' {
					depth--
				}
				i++
			}
			tokens = append(tokens, string(data[start:i]))
			continue
		}
		// hex string <...>
		if data[i] == '<' && (i+1 >= n || data[i+1] != '<') {
			start := i
			i++
			for i < n && data[i] != '>' {
				i++
			}
			if i < n {
				i++ // skip >
			}
			tokens = append(tokens, string(data[start:i]))
			continue
		}
		// dict << >>
		if data[i] == '<' && i+1 < n && data[i+1] == '<' {
			tokens = append(tokens, "<<")
			i += 2
			continue
		}
		if data[i] == '>' && i+1 < n && data[i+1] == '>' {
			tokens = append(tokens, ">>")
			i += 2
			continue
		}
		// array delimiters
		if data[i] == '[' {
			tokens = append(tokens, "[")
			i++
			continue
		}
		if data[i] == ']' {
			tokens = append(tokens, "]")
			i++
			continue
		}
		// regular token (name, number, operator)
		start := i
		for i < n && data[i] != ' ' && data[i] != '\t' && data[i] != '\r' &&
			data[i] != '\n' && data[i] != '(' && data[i] != ')' &&
			data[i] != '<' && data[i] != '>' && data[i] != '[' &&
			data[i] != ']' && data[i] != '%' {
			i++
		}
		if i > start {
			tokens = append(tokens, string(data[start:i]))
		}
	}
	return tokens
}

func isStringToken(s string) bool {
	return (strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")")) ||
		(strings.HasPrefix(s, "<") && strings.HasSuffix(s, ">") && !strings.HasPrefix(s, "<<"))
}

func findStringBefore(tokens []string, opIdx int) string {
	for j := opIdx - 1; j >= 0; j-- {
		if isStringToken(tokens[j]) {
			return tokens[j]
		}
		// stop if we hit another operator
		if _, err := strconv.ParseFloat(tokens[j], 64); err != nil && tokens[j] != "[" && tokens[j] != "]" {
			break
		}
	}
	return ""
}

func findArrayBefore(tokens []string, opIdx int) []string {
	// Walk backwards to find matching [ ... ]
	end := -1
	start := -1
	for j := opIdx - 1; j >= 0; j-- {
		if tokens[j] == "]" {
			end = j
		}
		if tokens[j] == "[" {
			start = j
			break
		}
	}
	if start >= 0 && end > start {
		return tokens[start+1 : end]
	}
	return nil
}

func fontDisplayName(fi *fontInfo) string {
	if fi == nil {
		return ""
	}
	if fi.baseFont != "" {
		return fi.baseFont
	}
	return fi.name
}

// decodeTextString decodes a PDF string token to a Go string.
func decodeTextString(raw string, fi *fontInfo) string {
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "<") && strings.HasSuffix(raw, ">") {
		// Hex string
		hex := raw[1 : len(raw)-1]
		hex = strings.ReplaceAll(hex, " ", "")
		hex = strings.ReplaceAll(hex, "\n", "")
		hex = strings.ReplaceAll(hex, "\r", "")
		if fi != nil && fi.toUni != nil {
			return decodeHexWithCMap(hex, fi.toUni)
		}
		if fi != nil && fi.isType0 {
			return decodeHexUTF16BE(hex)
		}
		return decodeHexLatin(hex)
	}
	if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") {
		inner := raw[1 : len(raw)-1]
		inner = unescapePDFString(inner)
		// Check for UTF-16BE BOM
		if len(inner) >= 2 && inner[0] == '\xfe' && inner[1] == '\xff' {
			return decodeUTF16BE([]byte(inner[2:]))
		}
		if fi != nil && fi.toUni != nil {
			// Try CMap mapping byte by byte
			var sb strings.Builder
			for j := 0; j < len(inner); j++ {
				code := uint16(inner[j])
				if r, ok := fi.toUni[code]; ok {
					sb.WriteRune(r)
				} else {
					sb.WriteByte(inner[j])
				}
			}
			return sb.String()
		}
		return inner
	}
	return raw
}

func decodeHexWithCMap(hex string, cmap map[uint16]rune) string {
	var sb strings.Builder
	// Try 4-digit codes first (CID fonts), fall back to 2-digit
	if len(hex)%4 == 0 {
		for i := 0; i+3 < len(hex); i += 4 {
			code := parseHex16(hex[i : i+4])
			if r, ok := cmap[code]; ok {
				sb.WriteRune(r)
			} else if code > 0 {
				sb.WriteRune(rune(code))
			}
		}
		return sb.String()
	}
	for i := 0; i+1 < len(hex); i += 2 {
		code := parseHex16(hex[i : i+2])
		if r, ok := cmap[code]; ok {
			sb.WriteRune(r)
		} else {
			sb.WriteByte(byte(code))
		}
	}
	return sb.String()
}

func decodeHexUTF16BE(hex string) string {
	if len(hex)%4 != 0 {
		// fall back to latin
		return decodeHexLatin(hex)
	}
	var codes []uint16
	for i := 0; i+3 < len(hex); i += 4 {
		codes = append(codes, parseHex16(hex[i:i+4]))
	}
	return string(utf16.Decode(codes))
}

func decodeHexLatin(hex string) string {
	var sb strings.Builder
	for i := 0; i+1 < len(hex); i += 2 {
		b := parseHex16(hex[i : i+2])
		sb.WriteByte(byte(b))
	}
	return sb.String()
}

func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		data = append(data, 0)
	}
	var codes []uint16
	for i := 0; i+1 < len(data); i += 2 {
		codes = append(codes, uint16(data[i])<<8|uint16(data[i+1]))
	}
	return string(utf16.Decode(codes))
}

func unescapePDFString(s string) string {
	var buf bytes.Buffer
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(':
				buf.WriteByte('(')
			case ')':
				buf.WriteByte(')')
			case '\\':
				buf.WriteByte('\\')
			default:
				// octal
				if s[i] >= '0' && s[i] <= '7' {
					oct := string(s[i])
					i++
					for j := 0; j < 2 && i < len(s) && s[i] >= '0' && s[i] <= '7'; j++ {
						oct += string(s[i])
						i++
					}
					v, _ := strconv.ParseUint(oct, 8, 8)
					buf.WriteByte(byte(v))
					continue
				}
				buf.WriteByte(s[i])
			}
		} else {
			buf.WriteByte(s[i])
		}
		i++
	}
	return buf.String()
}
