package gopdf

import (
	"fmt"
	"strings"
)

// TextReplaceResult holds the result of a text replacement operation.
type TextReplaceResult struct {
	// PageIndex is the 0-based page index where replacement occurred.
	PageIndex int
	// Count is the number of replacements made on this page.
	Count int
	// OriginalText is the text that was searched for.
	OriginalText string
	// ReplacementText is the text that replaced the original.
	ReplacementText string
}

// ReplaceTextOptions configures text replacement behavior.
type ReplaceTextOptions struct {
	// CaseInsensitive enables case-insensitive matching.
	CaseInsensitive bool
	// MaxReplacements limits the number of replacements (0 = unlimited).
	MaxReplacements int
	// Pages limits replacement to specific pages (0-based). Empty = all pages.
	Pages []int
}

// ReplaceText searches for oldText in the PDF content streams and replaces
// it with newText. This operates on the raw content stream level and works
// best with simple text replacements where the new text has similar length.
//
// Returns the total number of replacements made across all pages.
//
// Note: This modifies the raw PDF bytes. For complex replacements involving
// different fonts or sizes, use the Content Element API instead.
//
// Example:
//
//	data, _ := os.ReadFile("input.pdf")
//	newData, results, err := gopdf.ReplaceText(data, "Draft", "Final", nil)
//	if err == nil {
//	    os.WriteFile("output.pdf", newData, 0644)
//	}
func ReplaceText(pdfData []byte, oldText, newText string, opts *ReplaceTextOptions) ([]byte, []TextReplaceResult, error) {
	if oldText == "" {
		return pdfData, nil, fmt.Errorf("oldText cannot be empty")
	}
	if opts == nil {
		opts = &ReplaceTextOptions{}
	}

	parser, err := newRawPDFParser(pdfData)
	if err != nil {
		return pdfData, nil, fmt.Errorf("parse PDF: %w", err)
	}

	var results []TextReplaceResult
	totalReplacements := 0
	result := make([]byte, len(pdfData))
	copy(result, pdfData)

	pageSet := make(map[int]bool, len(opts.Pages))
	for _, p := range opts.Pages {
		pageSet[p] = true
	}

	for pageIdx, page := range parser.pages {
		// Skip pages not in the target set.
		if len(pageSet) > 0 && !pageSet[pageIdx] {
			continue
		}

		// Check max replacements.
		if opts.MaxReplacements > 0 && totalReplacements >= opts.MaxReplacements {
			break
		}

		stream := parser.getPageContentStream(pageIdx)
		if len(stream) == 0 {
			continue
		}

		// Get the content object number for this page.
		if len(page.contents) == 0 {
			continue
		}
		contentRef := page.contents[0]

		contentObj, ok := parser.objects[contentRef]
		if !ok {
			continue
		}

		// Perform replacement in the content stream.
		streamStr := string(stream)
		count := replaceInContentStream(&streamStr, oldText, newText, opts)

		if count > 0 {
			// Update the stream in the PDF data.
			newStream := []byte(streamStr)
			result = replaceObjectStream(result, contentRef, contentObj.dict, newStream)

			results = append(results, TextReplaceResult{
				PageIndex:       pageIdx,
				Count:           count,
				OriginalText:    oldText,
				ReplacementText: newText,
			})
			totalReplacements += count
		}
	}

	if totalReplacements > 0 {
		result = rebuildXref(result)
	}

	return result, results, nil
}

// replaceInContentStream replaces text within PDF content stream operators.
// It handles both Tj and TJ operators.
func replaceInContentStream(stream *string, oldText, newText string, opts *ReplaceTextOptions) int {
	count := 0
	s := *stream

	// Strategy 1: Replace in literal strings (text)
	count += replaceLiteralStrings(&s, oldText, newText, opts)

	// Strategy 2: Replace in hex strings <hex>
	count += replaceHexStrings(&s, oldText, newText, opts)

	*stream = s
	return count
}

// replaceLiteralStrings replaces text within PDF literal strings (...).
func replaceLiteralStrings(s *string, oldText, newText string, opts *ReplaceTextOptions) int {
	count := 0
	result := strings.Builder{}
	str := *s
	i := 0

	for i < len(str) {
		if str[i] == '(' {
			// Find matching closing paren.
			depth := 1
			start := i
			i++
			for i < len(str) && depth > 0 {
				if str[i] == '\\' {
					i += 2
					continue
				}
				if str[i] == '(' {
					depth++
				} else if str[i] == ')' {
					depth--
				}
				i++
			}
			literal := str[start:i]
			// Extract inner content.
			inner := literal[1 : len(literal)-1]

			var replaced string
			if opts.CaseInsensitive {
				replaced = caseInsensitiveReplace(inner, oldText, newText)
			} else {
				replaced = strings.ReplaceAll(inner, oldText, newText)
			}

			if replaced != inner {
				count += strings.Count(inner, oldText)
				if opts.CaseInsensitive {
					count = countCaseInsensitive(inner, oldText)
				}
				result.WriteByte('(')
				result.WriteString(replaced)
				result.WriteByte(')')
			} else {
				result.WriteString(literal)
			}
		} else {
			result.WriteByte(str[i])
			i++
		}
	}

	*s = result.String()
	return count
}

// replaceHexStrings replaces text within PDF hex strings <...>.
func replaceHexStrings(s *string, oldText, newText string, opts *ReplaceTextOptions) int {
	// Hex string replacement is more complex; for now we handle the common
	// case where hex strings encode ASCII-compatible text.
	count := 0
	str := *s
	result := strings.Builder{}
	i := 0

	for i < len(str) {
		if str[i] == '<' && i+1 < len(str) && str[i+1] != '<' {
			start := i
			i++
			for i < len(str) && str[i] != '>' {
				i++
			}
			if i < len(str) {
				i++ // skip >
			}
			hexStr := str[start:i]
			// Try to decode, replace, and re-encode.
			inner := hexStr[1 : len(hexStr)-1]
			decoded := decodeHexToASCII(inner)
			if decoded != "" && strings.Contains(decoded, oldText) {
				replaced := strings.ReplaceAll(decoded, oldText, newText)
				count += strings.Count(decoded, oldText)
				result.WriteByte('<')
				result.WriteString(encodeASCIIToHex(replaced))
				result.WriteByte('>')
			} else {
				result.WriteString(hexStr)
			}
		} else {
			result.WriteByte(str[i])
			i++
		}
	}

	if count > 0 {
		*s = result.String()
	}
	return count
}

// decodeHexToASCII decodes a hex string to ASCII if all bytes are printable.
func decodeHexToASCII(hex string) string {
	hex = strings.ReplaceAll(hex, " ", "")
	if len(hex)%2 != 0 {
		return ""
	}
	var sb strings.Builder
	for i := 0; i+1 < len(hex); i += 2 {
		b := parseHex16(hex[i : i+2])
		if b < 32 || b > 126 {
			return "" // not ASCII
		}
		sb.WriteByte(byte(b))
	}
	return sb.String()
}

// encodeASCIIToHex encodes an ASCII string to hex.
func encodeASCIIToHex(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		fmt.Fprintf(&sb, "%02X", s[i])
	}
	return sb.String()
}

// caseInsensitiveReplace performs case-insensitive string replacement.
func caseInsensitiveReplace(s, old, new string) string {
	lower := strings.ToLower(s)
	lowerOld := strings.ToLower(old)
	var result strings.Builder
	i := 0
	for i < len(s) {
		pos := strings.Index(lower[i:], lowerOld)
		if pos < 0 {
			result.WriteString(s[i:])
			break
		}
		result.WriteString(s[i : i+pos])
		result.WriteString(new)
		i += pos + len(old)
	}
	return result.String()
}

// countCaseInsensitive counts case-insensitive occurrences.
func countCaseInsensitive(s, substr string) int {
	count := 0
	lower := strings.ToLower(s)
	lowerSub := strings.ToLower(substr)
	i := 0
	for {
		pos := strings.Index(lower[i:], lowerSub)
		if pos < 0 {
			break
		}
		count++
		i += pos + len(lowerSub)
	}
	return count
}
