package gopdf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regexes for image extraction.
var (
	reFilterName  = regexp.MustCompile(`/Filter\s+/(\w+)`)
	reFilterArray = regexp.MustCompile(`/Filter\s*\[\s*/(\w+)`)
)

// ============================================================
// Image extraction from existing PDF files
// ============================================================

// ExtractImagesFromPage extracts image metadata and data from a
// specific page (0-based) of the given PDF data.
//
// Example:
//
//	data, _ := os.ReadFile("input.pdf")
//	images, _ := gopdf.ExtractImagesFromPage(data, 0)
//	for _, img := range images {
//	    fmt.Printf("%s: %dx%d %s\n", img.Name, img.Width, img.Height, img.ColorSpace)
//	}
func ExtractImagesFromPage(pdfData []byte, pageIndex int) ([]ExtractedImage, error) {
	parser, err := newRawPDFParser(pdfData)
	if err != nil {
		return nil, err
	}
	if pageIndex < 0 || pageIndex >= len(parser.pages) {
		return nil, fmt.Errorf("page index %d out of range (0..%d)", pageIndex, len(parser.pages)-1)
	}
	page := parser.pages[pageIndex]
	stream := parser.getPageContentStream(pageIndex)
	placements := parseImagePlacements(stream)
	return extractPageImages(parser, page, placements), nil
}

// ExtractImagesFromAllPages extracts images from all pages.
func ExtractImagesFromAllPages(pdfData []byte) (map[int][]ExtractedImage, error) {
	parser, err := newRawPDFParser(pdfData)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]ExtractedImage, len(parser.pages))
	for i, page := range parser.pages {
		stream := parser.getPageContentStream(i)
		placements := parseImagePlacements(stream)
		imgs := extractPageImages(parser, page, placements)
		if len(imgs) > 0 {
			result[i] = imgs
		}
	}
	return result, nil
}

// imagePlacement records where an XObject image is placed on a page.
type imagePlacement struct {
	name string  // e.g. "/Im1"
	a, b, c, d float64 // CTM components
	e, f       float64 // CTM translation
}

// parseImagePlacements finds "Do" operators in the content stream
// and extracts the CTM at each invocation point.
func parseImagePlacements(stream []byte) []imagePlacement {
	if len(stream) == 0 {
		return nil
	}
	tokens := tokenize(stream)
	var placements []imagePlacement

	// Track CTM via cm operators (simplified: last cm wins)
	ctmA, ctmB, ctmC, ctmD := 1.0, 0.0, 0.0, 1.0
	ctmE, ctmF := 0.0, 0.0

	stack := make([]float64, 0, 16)
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if v, err := strconv.ParseFloat(tok, 64); err == nil {
			stack = append(stack, v)
			continue
		}
		switch tok {
		case "cm":
			if len(stack) >= 6 {
				ctmA = stack[len(stack)-6]
				ctmB = stack[len(stack)-5]
				ctmC = stack[len(stack)-4]
				ctmD = stack[len(stack)-3]
				ctmE = stack[len(stack)-2]
				ctmF = stack[len(stack)-1]
				stack = stack[:len(stack)-6]
			}
		case "Do":
			// The name is the previous token
			if i >= 1 && strings.HasPrefix(tokens[i-1], "/") {
				placements = append(placements, imagePlacement{
					name: tokens[i-1],
					a: ctmA, b: ctmB, c: ctmC, d: ctmD,
					e: ctmE, f: ctmF,
				})
			}
			stack = stack[:0]
		case "q":
			// save graphics state — simplified, just reset stack
		case "Q":
			// restore graphics state — simplified
			ctmA, ctmB, ctmC, ctmD = 1, 0, 0, 1
			ctmE, ctmF = 0, 0
		default:
			// non-numeric, non-operator — could be a name
			if !strings.HasPrefix(tok, "/") {
				stack = stack[:0]
			}
		}
	}
	return placements
}

func extractPageImages(parser *rawPDFParser, page rawPDFPage, placements []imagePlacement) []ExtractedImage {
	var images []ExtractedImage
	pageHeight := page.mediaBox[3] - page.mediaBox[1]

	// Build placement map
	placementMap := make(map[string]imagePlacement)
	for _, p := range placements {
		placementMap[p.name] = p
	}

	for name, objNum := range page.resources.xobjs {
		obj, ok := parser.objects[objNum]
		if !ok {
			continue
		}
		// Check if it's an image XObject
		if !strings.Contains(obj.dict, "/Subtype /Image") &&
			!strings.Contains(obj.dict, "/Subtype/Image") {
			continue
		}
		img := ExtractedImage{
			Name:   name,
			ObjNum: objNum,
		}
		img.Width = extractIntValue(obj.dict, "/Width")
		img.Height = extractIntValue(obj.dict, "/Height")
		img.BitsPerComponent = extractIntValue(obj.dict, "/BitsPerComponent")
		img.ColorSpace = extractName(obj.dict, "/ColorSpace")
		img.Filter = extractFilterValue(obj.dict)
		if obj.stream != nil {
			// For FlateDecode, stream is already decompressed by parser.
			// For DCTDecode (JPEG), the raw stream IS the JPEG data.
			img.Data = obj.stream
		}
		// Apply placement info
		if p, ok := placementMap[name]; ok {
			img.X = p.e
			img.Y = pageHeight - p.f - p.d
			img.DisplayWidth = p.a
			img.DisplayHeight = p.d
		}
		images = append(images, img)
	}
	return images
}

// extractIntValue extracts an integer value for a given key from a dict string.
// Uses string search instead of regex for performance.
func extractIntValue(dict, key string) int {
	idx := strings.Index(dict, key)
	if idx < 0 {
		return 0
	}
	rest := dict[idx+len(key):]
	// Skip whitespace
	i := 0
	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t' || rest[i] == '\r' || rest[i] == '\n') {
		i++
	}
	// Read digits
	start := i
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i > start {
		v, _ := strconv.Atoi(rest[start:i])
		return v
	}
	return 0
}

func extractFilterValue(dict string) string {
	// /Filter can be a name or array — use pre-compiled regexes
	m := reFilterName.FindStringSubmatch(dict)
	if m != nil {
		return m[1]
	}
	// array form
	m2 := reFilterArray.FindStringSubmatch(dict)
	if m2 != nil {
		return m2[1]
	}
	return ""
}

// GetImageFormat returns the likely image format based on the filter.
func (img *ExtractedImage) GetImageFormat() string {
	switch img.Filter {
	case "DCTDecode":
		return "jpeg"
	case "JPXDecode":
		return "jp2"
	case "CCITTFaxDecode":
		return "tiff"
	case "FlateDecode", "":
		return "png" // raw pixel data, typically saved as PNG
	default:
		return "raw"
	}
}
