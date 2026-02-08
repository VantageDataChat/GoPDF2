package gopdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================
// PDF raw parser — reads cross-ref table, objects, streams
// from raw PDF bytes. Used for text/image extraction.
// ============================================================

// rawPDFParser parses a raw PDF byte slice to extract objects.
type rawPDFParser struct {
	data    []byte
	objects map[int]rawPDFObject // objNum -> object
	pages   []rawPDFPage
	root    int // root catalog obj number
}

// rawPDFObject holds a parsed PDF object.
type rawPDFObject struct {
	num    int
	dict   string // dictionary content between << >>
	stream []byte // decompressed stream content (nil if not a stream)
}

// rawPDFPage holds parsed page info.
type rawPDFPage struct {
	objNum    int
	mediaBox  [4]float64
	contents  []int // content stream obj numbers
	resources rawPDFResources
}

// rawPDFResources holds resource references for a page.
type rawPDFResources struct {
	fonts  map[string]int // /F1 -> obj number
	xobjs  map[string]int // /Im1 -> obj number
}

// newRawPDFParser creates a parser for the given PDF data.
func newRawPDFParser(data []byte) (*rawPDFParser, error) {
	p := &rawPDFParser{
		data:    data,
		objects: make(map[int]rawPDFObject),
	}
	if err := p.parse(); err != nil {
		return nil, fmt.Errorf("pdf parse: %w", err)
	}
	return p, nil
}

func (p *rawPDFParser) parse() error {
	p.parseObjects()
	p.findRoot()
	p.parsePages()
	return nil
}

// parseObjects finds all "N 0 obj ... endobj" blocks.
func (p *rawPDFParser) parseObjects() {
	re := regexp.MustCompile(`(\d+)\s+0\s+obj\b`)
	matches := re.FindAllSubmatchIndex(p.data, -1)
	for _, m := range matches {
		numStr := string(p.data[m[2]:m[3]])
		num, _ := strconv.Atoi(numStr)
		objStart := m[0]
		// find endobj
		endIdx := bytes.Index(p.data[objStart:], []byte("endobj"))
		if endIdx < 0 {
			continue
		}
		objData := p.data[objStart : objStart+endIdx+6]
		obj := rawPDFObject{num: num}
		// extract dictionary
		if dictStart := bytes.Index(objData, []byte("<<")); dictStart >= 0 {
			obj.dict = extractDict(objData[dictStart:])
		}
		// extract stream
		if streamStart := bytes.Index(objData, []byte("stream")); streamStart >= 0 {
			streamData := objData[streamStart+6:]
			// skip \r\n or \n after "stream"
			if len(streamData) > 0 && streamData[0] == '\r' {
				streamData = streamData[1:]
			}
			if len(streamData) > 0 && streamData[0] == '\n' {
				streamData = streamData[1:]
			}
			if endStream := bytes.Index(streamData, []byte("endstream")); endStream >= 0 {
				raw := streamData[:endStream]
				// trim trailing whitespace
				raw = bytes.TrimRight(raw, "\r\n")
				if strings.Contains(obj.dict, "/FlateDecode") {
					if decoded, err := zlibDecompress(raw); err == nil {
						obj.stream = decoded
					} else {
						obj.stream = raw
					}
				} else {
					obj.stream = raw
				}
			}
		}
		p.objects[num] = obj
	}
}

// extractDict extracts the outermost <<...>> from data.
func extractDict(data []byte) string {
	depth := 0
	start := -1
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '<' && data[i+1] == '<' {
			if depth == 0 {
				start = i
			}
			depth++
			i++
		} else if data[i] == '>' && data[i+1] == '>' {
			depth--
			if depth == 0 {
				return string(data[start : i+2])
			}
			i++
		}
	}
	if start >= 0 {
		return string(data[start:])
	}
	return ""
}

func zlibDecompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (p *rawPDFParser) findRoot() {
	// Look for /Root N 0 R in trailer or xref stream
	re := regexp.MustCompile(`/Root\s+(\d+)\s+0\s+R`)
	m := re.FindSubmatch(p.data)
	if m != nil {
		p.root, _ = strconv.Atoi(string(m[1]))
		return
	}
}

func (p *rawPDFParser) parsePages() {
	rootObj, ok := p.objects[p.root]
	if !ok {
		return
	}
	// Find /Pages reference
	pagesRef := extractRef(rootObj.dict, "/Pages")
	if pagesRef <= 0 {
		return
	}
	p.collectPages(pagesRef)
}

func (p *rawPDFParser) collectPages(objNum int) {
	obj, ok := p.objects[objNum]
	if !ok {
		return
	}
	if strings.Contains(obj.dict, "/Type /Page\n") || strings.Contains(obj.dict, "/Type/Page") ||
		(strings.Contains(obj.dict, "/Type /Page") && !strings.Contains(obj.dict, "/Type /Pages")) {
		page := rawPDFPage{objNum: objNum}
		page.mediaBox = extractMediaBox(obj.dict)
		page.contents = extractContentRefs(obj.dict)
		page.resources = p.extractResources(obj.dict, objNum)
		p.pages = append(p.pages, page)
		return
	}
	// It's a /Pages node — recurse into /Kids
	kids := extractRefArray(obj.dict, "/Kids")
	for _, kid := range kids {
		p.collectPages(kid)
	}
}

// extractRef extracts a single "N 0 R" reference for a given key.
func extractRef(dict, key string) int {
	re := regexp.MustCompile(regexp.QuoteMeta(key) + `\s+(\d+)\s+0\s+R`)
	m := re.FindStringSubmatch(dict)
	if m != nil {
		n, _ := strconv.Atoi(m[1])
		return n
	}
	return 0
}

// extractRefArray extracts an array of "N 0 R" references for a given key.
func extractRefArray(dict, key string) []int {
	idx := strings.Index(dict, key)
	if idx < 0 {
		return nil
	}
	rest := dict[idx+len(key):]
	// find [ ... ]
	start := strings.Index(rest, "[")
	if start < 0 {
		return nil
	}
	end := strings.Index(rest[start:], "]")
	if end < 0 {
		return nil
	}
	arr := rest[start+1 : start+end]
	re := regexp.MustCompile(`(\d+)\s+0\s+R`)
	matches := re.FindAllStringSubmatch(arr, -1)
	var refs []int
	for _, m := range matches {
		n, _ := strconv.Atoi(m[1])
		refs = append(refs, n)
	}
	return refs
}

func extractMediaBox(dict string) [4]float64 {
	re := regexp.MustCompile(`/MediaBox\s*\[\s*([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s*\]`)
	m := re.FindStringSubmatch(dict)
	if m == nil {
		return [4]float64{0, 0, 612, 792} // default letter
	}
	var box [4]float64
	for i := 0; i < 4; i++ {
		box[i], _ = strconv.ParseFloat(m[i+1], 64)
	}
	return box
}

func extractContentRefs(dict string) []int {
	// /Contents N 0 R  or  /Contents [N 0 R M 0 R]
	idx := strings.Index(dict, "/Contents")
	if idx < 0 {
		return nil
	}
	rest := dict[idx+9:]
	rest = strings.TrimLeft(rest, " \t\r\n")
	if len(rest) > 0 && rest[0] == '[' {
		return extractRefArray(dict, "/Contents")
	}
	ref := extractRef(dict, "/Contents")
	if ref > 0 {
		return []int{ref}
	}
	return nil
}

func (p *rawPDFParser) extractResources(dict string, pageObjNum int) rawPDFResources {
	res := rawPDFResources{
		fonts: make(map[string]int),
		xobjs: make(map[string]int),
	}
	// Resources can be inline or a reference
	resDict := dict
	resRef := extractRef(dict, "/Resources")
	if resRef > 0 {
		if obj, ok := p.objects[resRef]; ok {
			resDict = obj.dict
		}
	}
	// Extract /Font << /F1 N 0 R ... >>
	p.extractNamedRefs(resDict, "/Font", res.fonts)
	// Extract /XObject << /Im1 N 0 R ... >>
	p.extractNamedRefs(resDict, "/XObject", res.xobjs)
	return res
}

func (p *rawPDFParser) extractNamedRefs(dict, key string, out map[string]int) {
	idx := strings.Index(dict, key)
	if idx < 0 {
		return
	}
	rest := dict[idx+len(key):]
	rest = strings.TrimLeft(rest, " \t\r\n")
	if len(rest) == 0 {
		return
	}
	if rest[0] == '<' && len(rest) > 1 && rest[1] == '<' {
		// inline dict
		inner := extractDict([]byte(rest))
		re := regexp.MustCompile(`/(\w+)\s+(\d+)\s+0\s+R`)
		matches := re.FindAllStringSubmatch(inner, -1)
		for _, m := range matches {
			n, _ := strconv.Atoi(m[2])
			out["/"+m[1]] = n
		}
	} else {
		// reference to another object
		re := regexp.MustCompile(`(\d+)\s+0\s+R`)
		m := re.FindStringSubmatch(rest)
		if m != nil {
			n, _ := strconv.Atoi(m[1])
			if obj, ok := p.objects[n]; ok {
				re2 := regexp.MustCompile(`/(\w+)\s+(\d+)\s+0\s+R`)
				matches := re2.FindAllStringSubmatch(obj.dict, -1)
				for _, mm := range matches {
					nn, _ := strconv.Atoi(mm[2])
					out["/"+mm[1]] = nn
				}
			}
		}
	}
}

// getPageContentStream returns the concatenated, decompressed content
// stream(s) for a page.
func (p *rawPDFParser) getPageContentStream(pageIdx int) []byte {
	if pageIdx < 0 || pageIdx >= len(p.pages) {
		return nil
	}
	page := p.pages[pageIdx]
	var buf bytes.Buffer
	for _, ref := range page.contents {
		obj, ok := p.objects[ref]
		if !ok {
			continue
		}
		if obj.stream != nil {
			buf.Write(obj.stream)
			buf.WriteByte('\n')
		}
	}
	return buf.Bytes()
}

// ============================================================
// Content stream text operator parser
// ============================================================

// ExtractedText represents a piece of text extracted from a PDF page.
type ExtractedText struct {
	// Text is the extracted text string.
	Text string
	// X is the horizontal position.
	X float64
	// Y is the vertical position.
	Y float64
	// FontName is the PDF font resource name (e.g. "/F1").
	FontName string
	// FontSize is the font size in points.
	FontSize float64
}

// ExtractedImage represents an image found on a PDF page.
type ExtractedImage struct {
	// Name is the XObject resource name (e.g. "/Im1").
	Name string
	// Width is the image width in pixels.
	Width int
	// Height is the image height in pixels.
	Height int
	// BitsPerComponent is the bits per color component.
	BitsPerComponent int
	// ColorSpace is the color space name.
	ColorSpace string
	// Filter is the compression filter.
	Filter string
	// Data is the raw (possibly compressed) image data.
	Data []byte
	// ObjNum is the PDF object number.
	ObjNum int
	// X, Y are the position on the page (from the CTM).
	X float64
	// Y position on the page.
	Y float64
	// DisplayWidth is the rendered width on the page.
	DisplayWidth float64
	// DisplayHeight is the rendered height on the page.
	DisplayHeight float64
}
