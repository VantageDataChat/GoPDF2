package gopdf

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// AnnotationType represents the type of PDF annotation.
type AnnotationType int

const (
	// AnnotText is a sticky note annotation (appears as an icon).
	AnnotText AnnotationType = iota
	// AnnotHighlight highlights existing text on the page.
	AnnotHighlight
	// AnnotUnderline underlines existing text on the page.
	AnnotUnderline
	// AnnotStrikeOut strikes out existing text on the page.
	AnnotStrikeOut
	// AnnotSquare draws a rectangle annotation.
	AnnotSquare
	// AnnotCircle draws a circle/ellipse annotation.
	AnnotCircle
	// AnnotFreeText places text directly on the page.
	AnnotFreeText
)

// AnnotationOption configures a PDF annotation.
type AnnotationOption struct {
	// Type is the annotation type.
	Type AnnotationType

	// Rect defines the annotation rectangle [x, y, w, h] in document units.
	// x, y is the top-left corner.
	X, Y, W, H float64

	// Title is the annotation title (author name for sticky notes).
	Title string

	// Content is the annotation text content.
	Content string

	// Color is the annotation color in RGB. Default: yellow (255, 255, 0).
	Color [3]uint8

	// Opacity is the annotation opacity (0.0 to 1.0). Default: 1.0.
	Opacity float64

	// Open determines if a text annotation popup is initially open.
	Open bool

	// CreationDate is the annotation creation date. Default: now.
	CreationDate time.Time

	// FontSize is used for FreeText annotations. Default: 12.
	FontSize float64
}

func (o *AnnotationOption) defaults() {
	if o.Color == [3]uint8{0, 0, 0} {
		o.Color = [3]uint8{255, 255, 0}
	}
	if o.Opacity <= 0 {
		o.Opacity = 1.0
	}
	if o.CreationDate.IsZero() {
		o.CreationDate = time.Now()
	}
	if o.FontSize <= 0 {
		o.FontSize = 12
	}
}

// annotationObj represents a PDF annotation object.
type annotationObj struct {
	opt     AnnotationOption
	getRoot func() *GoPdf
}

func (a annotationObj) init(f func() *GoPdf) {}

func (a annotationObj) getType() string {
	return "Annot"
}

func (a annotationObj) write(w io.Writer, objID int) error {
	a.opt.defaults()

	pageH := a.getRoot().config.PageSize.H

	// Convert from top-left origin to PDF bottom-left origin.
	x1 := a.opt.X
	y1 := pageH - a.opt.Y
	x2 := a.opt.X + a.opt.W
	y2 := pageH - (a.opt.Y + a.opt.H)

	// Escape content string.
	content := escapeAnnotString(a.opt.Content)
	title := escapeAnnotString(a.opt.Title)

	// Color components as floats (0-1).
	cr := float64(a.opt.Color[0]) / 255.0
	cg := float64(a.opt.Color[1]) / 255.0
	cb := float64(a.opt.Color[2]) / 255.0

	io.WriteString(w, "<<\n")
	io.WriteString(w, "/Type /Annot\n")

	switch a.opt.Type {
	case AnnotText:
		fmt.Fprintf(w, "/Subtype /Text\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		fmt.Fprintf(w, "/Contents (%s)\n", content)
		if title != "" {
			fmt.Fprintf(w, "/T (%s)\n", title)
		}
		fmt.Fprintf(w, "/C [%.4f %.4f %.4f]\n", cr, cg, cb)
		if a.opt.Open {
			io.WriteString(w, "/Open true\n")
		}
		io.WriteString(w, "/Name /Comment\n")

	case AnnotHighlight:
		fmt.Fprintf(w, "/Subtype /Highlight\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		// QuadPoints for highlight area.
		fmt.Fprintf(w, "/QuadPoints [%.2f %.2f %.2f %.2f %.2f %.2f %.2f %.2f]\n",
			x1, y1, x2, y1, x1, y2, x2, y2)
		fmt.Fprintf(w, "/C [%.4f %.4f %.4f]\n", cr, cg, cb)
		if content != "" {
			fmt.Fprintf(w, "/Contents (%s)\n", content)
		}

	case AnnotUnderline:
		fmt.Fprintf(w, "/Subtype /Underline\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		fmt.Fprintf(w, "/QuadPoints [%.2f %.2f %.2f %.2f %.2f %.2f %.2f %.2f]\n",
			x1, y1, x2, y1, x1, y2, x2, y2)
		fmt.Fprintf(w, "/C [%.4f %.4f %.4f]\n", cr, cg, cb)

	case AnnotStrikeOut:
		fmt.Fprintf(w, "/Subtype /StrikeOut\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		fmt.Fprintf(w, "/QuadPoints [%.2f %.2f %.2f %.2f %.2f %.2f %.2f %.2f]\n",
			x1, y1, x2, y1, x1, y2, x2, y2)
		fmt.Fprintf(w, "/C [%.4f %.4f %.4f]\n", cr, cg, cb)

	case AnnotSquare:
		fmt.Fprintf(w, "/Subtype /Square\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		fmt.Fprintf(w, "/C [%.4f %.4f %.4f]\n", cr, cg, cb)
		if content != "" {
			fmt.Fprintf(w, "/Contents (%s)\n", content)
		}

	case AnnotCircle:
		fmt.Fprintf(w, "/Subtype /Circle\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		fmt.Fprintf(w, "/C [%.4f %.4f %.4f]\n", cr, cg, cb)
		if content != "" {
			fmt.Fprintf(w, "/Contents (%s)\n", content)
		}

	case AnnotFreeText:
		fmt.Fprintf(w, "/Subtype /FreeText\n")
		fmt.Fprintf(w, "/Rect [%.2f %.2f %.2f %.2f]\n", x1, y2, x2, y1)
		fmt.Fprintf(w, "/Contents (%s)\n", content)
		fmt.Fprintf(w, "/DA (/Helv %.0f Tf %.4f %.4f %.4f rg)\n",
			a.opt.FontSize, cr, cg, cb)
		io.WriteString(w, "/DS (font: Helvetica,sans-serif)\n")
	}

	// Opacity via CA entry.
	if a.opt.Opacity < 1.0 {
		fmt.Fprintf(w, "/CA %.4f\n", a.opt.Opacity)
	}

	// Border.
	io.WriteString(w, "/Border [0 0 1]\n")

	// Flags: Print (bit 3).
	io.WriteString(w, "/F 4\n")

	io.WriteString(w, ">>\n")
	return nil
}

// escapeAnnotString escapes special characters in PDF annotation strings.
func escapeAnnotString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// AddAnnotation adds an annotation to the current page.
//
// Supported annotation types:
//   - AnnotText: Sticky note (comment icon)
//   - AnnotHighlight: Highlight markup
//   - AnnotUnderline: Underline markup
//   - AnnotStrikeOut: Strikeout markup
//   - AnnotSquare: Rectangle shape
//   - AnnotCircle: Circle/ellipse shape
//   - AnnotFreeText: Text directly on the page
//
// Example:
//
//	pdf.AddAnnotation(gopdf.AnnotationOption{
//	    Type:    gopdf.AnnotText,
//	    X:       100,
//	    Y:       100,
//	    W:       24,
//	    H:       24,
//	    Title:   "Reviewer",
//	    Content: "Please check this section.",
//	    Color:   [3]uint8{255, 255, 0},
//	})
func (gp *GoPdf) AddAnnotation(opt AnnotationOption) {
	gp.UnitsToPointsVar(&opt.X, &opt.Y, &opt.W, &opt.H)

	page := gp.pdfObjs[gp.curr.IndexOfPageObj].(*PageObj)
	objIdx := gp.addObj(annotationObj{
		opt: opt,
		getRoot: func() *GoPdf {
			return gp
		},
	})
	page.LinkObjIds = append(page.LinkObjIds, objIdx+1)
}

// AddTextAnnotation is a convenience method for adding a sticky note annotation.
func (gp *GoPdf) AddTextAnnotation(x, y float64, title, content string) {
	gp.AddAnnotation(AnnotationOption{
		Type:    AnnotText,
		X:       x,
		Y:       y,
		W:       24,
		H:       24,
		Title:   title,
		Content: content,
	})
}

// AddHighlightAnnotation is a convenience method for highlighting a rectangular area.
func (gp *GoPdf) AddHighlightAnnotation(x, y, w, h float64, color [3]uint8) {
	gp.AddAnnotation(AnnotationOption{
		Type:  AnnotHighlight,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
		Color: color,
	})
}

// AddFreeTextAnnotation is a convenience method for adding text directly on the page.
func (gp *GoPdf) AddFreeTextAnnotation(x, y, w, h float64, text string, fontSize float64) {
	gp.AddAnnotation(AnnotationOption{
		Type:     AnnotFreeText,
		X:        x,
		Y:        y,
		W:        w,
		H:        h,
		Content:  text,
		FontSize: fontSize,
	})
}
