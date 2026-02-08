package gopdf

import (
	"fmt"
	"io"
)

// CatalogObj : catalog dictionary
type CatalogObj struct { //impl IObj
	outlinesObjID   int
	namesObjID      int // index of Names dictionary object (-1 = none)
	pageLabelsObjID int // index of PageLabels object (-1 = none)
	metadataObjID   int // index of XMP Metadata stream object (-1 = none)
}

func (c *CatalogObj) init(funcGetRoot func() *GoPdf) {
	c.outlinesObjID = -1
	c.namesObjID = -1
	c.pageLabelsObjID = -1
	c.metadataObjID = -1
}

func (c *CatalogObj) getType() string {
	return "Catalog"
}

func (c *CatalogObj) write(w io.Writer, objID int) error {
	io.WriteString(w, "<<\n")
	fmt.Fprintf(w, "  /Type /%s\n", c.getType())
	io.WriteString(w, "  /Pages 2 0 R\n")
	if c.outlinesObjID >= 0 {
		io.WriteString(w, "  /PageMode /UseOutlines\n")
		fmt.Fprintf(w, "  /Outlines %d 0 R\n", c.outlinesObjID)
	}
	if c.namesObjID >= 0 {
		fmt.Fprintf(w, "  /Names %d 0 R\n", c.namesObjID)
	}
	if c.pageLabelsObjID >= 0 {
		fmt.Fprintf(w, "  /PageLabels %d 0 R\n", c.pageLabelsObjID)
	}
	if c.metadataObjID >= 0 {
		fmt.Fprintf(w, "  /Metadata %d 0 R\n", c.metadataObjID)
	}
	io.WriteString(w, ">>\n")
	return nil
}

func (c *CatalogObj) SetIndexObjOutlines(index int) {
	c.outlinesObjID = index + 1
}

// SetIndexObjNames sets the Names dictionary object reference.
func (c *CatalogObj) SetIndexObjNames(index int) {
	c.namesObjID = index + 1
}

// SetIndexObjPageLabels sets the PageLabels object reference.
func (c *CatalogObj) SetIndexObjPageLabels(index int) {
	c.pageLabelsObjID = index + 1
}

// SetIndexObjMetadata sets the XMP Metadata stream object reference.
func (c *CatalogObj) SetIndexObjMetadata(index int) {
	c.metadataObjID = index + 1
}
