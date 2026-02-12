package gopdf

import (
	"bytes"
	"fmt"
	"io"

	"github.com/phpdave11/gofpdi"
)

// pdfProbeResult holds the results of safely probing a PDF with gofpdi.
type pdfProbeResult struct {
	NumPages int
	Sizes    map[int]map[string]map[string]float64
}

// safeProbePDF creates a gofpdi importer, probes page count and sizes,
// and recovers from any panic that gofpdi may trigger on certain valid PDFs.
func safeProbePDF(pdfData []byte) (result pdfProbeResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to parse PDF: %v", r)
		}
	}()

	probe := gofpdi.NewImporter()
	rs := io.ReadSeeker(bytes.NewReader(pdfData))
	probe.SetSourceStream(&rs)

	result.NumPages = probe.GetNumPages()
	result.Sizes = probe.GetPageSizes()
	return result, nil
}

// safeProbePageCount is like safeProbePDF but only returns the page count.
func safeProbePageCount(pdfData []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to parse PDF: %v", r)
		}
	}()

	probe := gofpdi.NewImporter()
	rs := io.ReadSeeker(bytes.NewReader(pdfData))
	probe.SetSourceStream(&rs)

	n = probe.GetNumPages()
	return n, nil
}
