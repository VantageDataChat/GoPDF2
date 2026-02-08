package gopdf

import (
	"fmt"
	"io"
	"strings"
)

// OCGIntent represents the intent of an Optional Content Group.
type OCGIntent string

const (
	// OCGIntentView indicates the OCG is for viewing purposes.
	OCGIntentView OCGIntent = "View"
	// OCGIntentDesign indicates the OCG is for design purposes.
	OCGIntentDesign OCGIntent = "Design"
)

// OCG represents an Optional Content Group (layer) in a PDF.
// OCGs allow content to be selectively shown or hidden.
type OCG struct {
	// Name is the display name of the layer.
	Name string
	// Intent is the visibility intent ("View" or "Design").
	Intent OCGIntent
	// On indicates whether the layer is initially visible.
	On bool
	// objIndex is the internal object index (set after adding).
	objIndex int
}

// ocgObj is the PDF object for an Optional Content Group.
type ocgObj struct {
	name   string
	intent OCGIntent
}

func (o ocgObj) init(f func() *GoPdf) {}

func (o ocgObj) getType() string {
	return "OCG"
}

func (o ocgObj) write(w io.Writer, objID int) error {
	io.WriteString(w, "<<\n")
	io.WriteString(w, "/Type /OCG\n")
	fmt.Fprintf(w, "/Name (%s)\n", escapeAnnotString(o.name))
	intent := string(o.intent)
	if intent == "" {
		intent = "View"
	}
	fmt.Fprintf(w, "/Intent /%s\n", intent)
	io.WriteString(w, ">>\n")
	return nil
}

// ocPropertiesObj is the PDF object for the /OCProperties dictionary
// in the catalog. It lists all OCGs and their default visibility.
type ocPropertiesObj struct {
	ocgs []ocgRef
}

type ocgRef struct {
	objID int // 1-based PDF object ID
	on    bool
}

func (o ocPropertiesObj) init(f func() *GoPdf) {}

func (o ocPropertiesObj) getType() string {
	return "OCProperties"
}

func (o ocPropertiesObj) write(w io.Writer, objID int) error {
	// Build OCG reference list.
	var allRefs []string
	var onRefs []string
	var offRefs []string

	for _, ref := range o.ocgs {
		r := fmt.Sprintf("%d 0 R", ref.objID)
		allRefs = append(allRefs, r)
		if ref.on {
			onRefs = append(onRefs, r)
		} else {
			offRefs = append(offRefs, r)
		}
	}

	io.WriteString(w, "<<\n")
	// /OCGs array — all OCGs in the document.
	fmt.Fprintf(w, "/OCGs [%s]\n", strings.Join(allRefs, " "))

	// /D — default configuration dictionary.
	io.WriteString(w, "/D <<\n")
	if len(onRefs) > 0 {
		fmt.Fprintf(w, "  /ON [%s]\n", strings.Join(onRefs, " "))
	}
	if len(offRefs) > 0 {
		fmt.Fprintf(w, "  /OFF [%s]\n", strings.Join(offRefs, " "))
	}
	// /Order array controls the layer panel display order.
	fmt.Fprintf(w, "  /Order [%s]\n", strings.Join(allRefs, " "))
	io.WriteString(w, ">>\n")

	io.WriteString(w, ">>\n")
	return nil
}

// AddOCG adds an Optional Content Group (layer) to the document.
// Returns the OCG for use with SetContentOCG.
//
// Example:
//
//	watermarkLayer := pdf.AddOCG(gopdf.OCG{
//	    Name:   "Watermark",
//	    Intent: gopdf.OCGIntentView,
//	    On:     true,
//	})
//	draftLayer := pdf.AddOCG(gopdf.OCG{
//	    Name:   "Draft Notes",
//	    Intent: gopdf.OCGIntentDesign,
//	    On:     false,
//	})
func (gp *GoPdf) AddOCG(ocg OCG) OCG {
	if ocg.Intent == "" {
		ocg.Intent = OCGIntentView
	}
	idx := gp.addObj(ocgObj{
		name:   ocg.Name,
		intent: ocg.Intent,
	})
	ocg.objIndex = idx
	gp.ocgs = append(gp.ocgs, ocgRef{
		objID: idx + 1,
		on:    ocg.On,
	})
	return ocg
}

// GetOCGs returns all Optional Content Groups defined in the document.
func (gp *GoPdf) GetOCGs() []OCG {
	var result []OCG
	for i, obj := range gp.pdfObjs {
		if o, ok := obj.(ocgObj); ok {
			on := true
			for _, ref := range gp.ocgs {
				if ref.objID == i+1 {
					on = ref.on
					break
				}
			}
			result = append(result, OCG{
				Name:     o.name,
				Intent:   o.intent,
				On:       on,
				objIndex: i,
			})
		}
	}
	return result
}
