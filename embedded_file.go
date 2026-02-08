package gopdf

import (
	"compress/flate"
	"fmt"
	"io"
	"time"
)

// EmbeddedFile represents a file to be embedded in the PDF.
type EmbeddedFile struct {
	Name        string    // Display name of the file
	Content     []byte    // File content
	MimeType    string    // MIME type (e.g. "application/pdf", "text/plain")
	Description string    // Optional description
	ModDate     time.Time // Modification date (default: now)
}

// embeddedFileStreamObj is the PDF stream object for the embedded file content.
type embeddedFileStreamObj struct {
	data     []byte
	mimeType string
	modDate  time.Time
}

func (e embeddedFileStreamObj) init(f func() *GoPdf) {}

func (e embeddedFileStreamObj) getType() string {
	return "EmbeddedFile"
}

func (e embeddedFileStreamObj) write(w io.Writer, objID int) error {
	// Compress the data.
	var compressed []byte
	{
		var buf []byte
		bw := &byteWriter{buf: &buf}
		zw, err := flate.NewWriter(bw, flate.DefaultCompression)
		if err != nil {
			return err
		}
		if _, err := zw.Write(e.data); err != nil {
			return err
		}
		if err := zw.Close(); err != nil {
			return err
		}
		compressed = *bw.buf
	}

	io.WriteString(w, "<<\n")
	io.WriteString(w, "/Type /EmbeddedFile\n")
	if e.mimeType != "" {
		fmt.Fprintf(w, "/Subtype /%s\n", escapeMimeType(e.mimeType))
	}
	fmt.Fprintf(w, "/Length %d\n", len(compressed))
	io.WriteString(w, "/Filter /FlateDecode\n")
	io.WriteString(w, "/Params <<\n")
	fmt.Fprintf(w, "  /Size %d\n", len(e.data))
	if !e.modDate.IsZero() {
		fmt.Fprintf(w, "  /ModDate (D:%s)\n", e.modDate.Format("20060102150405"))
	}
	io.WriteString(w, ">>\n")
	io.WriteString(w, ">>\n")
	io.WriteString(w, "stream\n")
	w.Write(compressed)
	io.WriteString(w, "\nendstream\n")
	return nil
}

// fileSpecObj is the PDF file specification object.
type fileSpecObj struct {
	name        string
	description string
	streamObjID int // 1-based object ID of the embedded file stream
}

func (f fileSpecObj) init(fn func() *GoPdf) {}

func (f fileSpecObj) getType() string {
	return "Filespec"
}

func (f fileSpecObj) write(w io.Writer, objID int) error {
	io.WriteString(w, "<<\n")
	io.WriteString(w, "/Type /Filespec\n")
	fmt.Fprintf(w, "/F (%s)\n", escapeAnnotString(f.name))
	fmt.Fprintf(w, "/UF (%s)\n", escapeAnnotString(f.name))
	if f.description != "" {
		fmt.Fprintf(w, "/Desc (%s)\n", escapeAnnotString(f.description))
	}
	fmt.Fprintf(w, "/EF << /F %d 0 R >>\n", f.streamObjID)
	io.WriteString(w, ">>\n")
	return nil
}

// byteWriter is a simple io.Writer that appends to a byte slice.
type byteWriter struct {
	buf *[]byte
}

func (bw *byteWriter) Write(p []byte) (int, error) {
	*bw.buf = append(*bw.buf, p...)
	return len(p), nil
}

// escapeMimeType converts a MIME type to a PDF name-safe string.
// e.g. "application/pdf" -> "application#2Fpdf"
func escapeMimeType(mime string) string {
	// PDF spec allows subtype names; common practice is to use the MIME subtype.
	// For simplicity, we just replace "/" with "#2F".
	result := make([]byte, 0, len(mime))
	for i := 0; i < len(mime); i++ {
		if mime[i] == '/' {
			result = append(result, '#', '2', 'F')
		} else {
			result = append(result, mime[i])
		}
	}
	return string(result)
}

// AddEmbeddedFile attaches a file to the PDF document.
// The file will appear in the PDF viewer's attachment panel.
//
// Example:
//
//	data, _ := os.ReadFile("report.csv")
//	pdf.AddEmbeddedFile(gopdf.EmbeddedFile{
//	    Name:     "report.csv",
//	    Content:  data,
//	    MimeType: "text/csv",
//	})
func (gp *GoPdf) AddEmbeddedFile(ef EmbeddedFile) error {
	if ef.Name == "" {
		return ErrEmptyString
	}
	if len(ef.Content) == 0 {
		return ErrEmptyString
	}
	if ef.ModDate.IsZero() {
		ef.ModDate = time.Now()
	}

	// Add the embedded file stream object.
	streamIdx := gp.addObj(embeddedFileStreamObj{
		data:     ef.Content,
		mimeType: ef.MimeType,
		modDate:  ef.ModDate,
	})
	streamObjID := streamIdx + 1 // PDF object IDs are 1-based

	// Add the file specification object.
	fileSpecIdx := gp.addObj(fileSpecObj{
		name:        ef.Name,
		description: ef.Description,
		streamObjID: streamObjID,
	})
	_ = fileSpecIdx

	// Track embedded files for the Names dictionary.
	if gp.embeddedFiles == nil {
		gp.embeddedFiles = make([]embeddedFileRef, 0)
	}
	gp.embeddedFiles = append(gp.embeddedFiles, embeddedFileRef{
		name:          ef.Name,
		fileSpecObjID: fileSpecIdx + 1,
	})

	return nil
}

// embeddedFileRef tracks an embedded file for the Names dictionary.
type embeddedFileRef struct {
	name          string
	fileSpecObjID int
}
