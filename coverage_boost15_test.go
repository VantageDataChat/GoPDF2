package gopdf

import (
	"bytes"
	"strings"
	"testing"
)

// ============================================================
// coverage_boost15_test.go — TestCov15_ prefix
// Targets: form_field.go, pdf_obj_id.go, svg_insert.go,
// pdf_parser.go extractNamedRefs, cache_content_text_color_cmyk,
// image_obj.go GetRect, list_cache_content.go write,
// cache_content_rotate.go write, embedded_file.go UpdateEmbeddedFile
// ============================================================

// ============================================================
// form_field.go — AddFormField, AddTextField, AddCheckbox, AddDropdown
// ============================================================

func TestCov15_AddFormField_Text(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:        FormFieldText,
		Name:        "username",
		X:           50,
		Y:           100,
		W:           200,
		H:           25,
		Value:       "Enter name",
		FontFamily:  fontFamily,
		FontSize:    12,
		HasBorder:   true,
		BorderColor: [3]uint8{0, 0, 0},
		HasFill:     true,
		FillColor:   [3]uint8{255, 255, 255},
	})
	if err != nil {
		t.Fatalf("AddFormField text: %v", err)
	}
}

func TestCov15_AddFormField_TextMultiline(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:      FormFieldText,
		Name:      "comments",
		X:         50,
		Y:         150,
		W:         200,
		H:         80,
		Multiline: true,
		MaxLen:    500,
		ReadOnly:  false,
		Required:  true,
		FontSize:  10,
		HasBorder: true,
		BorderColor: [3]uint8{128, 128, 128},
		Color:     [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("AddFormField multiline: %v", err)
	}
}

func TestCov15_AddFormField_Checkbox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:        FormFieldCheckbox,
		Name:        "agree",
		X:           50,
		Y:           200,
		W:           15,
		H:           15,
		Checked:     true,
		HasBorder:   true,
		BorderColor: [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("AddFormField checkbox: %v", err)
	}
}

func TestCov15_AddFormField_CheckboxUnchecked(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:    FormFieldCheckbox,
		Name:    "opt_out",
		X:       50,
		Y:       220,
		W:       15,
		H:       15,
		Checked: false,
	})
	if err != nil {
		t.Fatalf("AddFormField checkbox unchecked: %v", err)
	}
}

func TestCov15_AddFormField_Radio(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldRadio,
		Name: "gender",
		X:    50,
		Y:    250,
		W:    15,
		H:    15,
	})
	if err != nil {
		t.Fatalf("AddFormField radio: %v", err)
	}
}

func TestCov15_AddFormField_Choice(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:     FormFieldChoice,
		Name:     "country",
		X:        50,
		Y:        280,
		W:        200,
		H:        25,
		Options:  []string{"USA", "Canada", "UK", "Germany"},
		FontSize: 12,
		HasBorder: true,
		BorderColor: [3]uint8{0, 0, 0},
	})
	if err != nil {
		t.Fatalf("AddFormField choice: %v", err)
	}
}

func TestCov15_AddFormField_Button(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:  FormFieldButton,
		Name:  "submit",
		X:     50,
		Y:     320,
		W:     100,
		H:     30,
		Value: "Submit",
	})
	if err != nil {
		t.Fatalf("AddFormField button: %v", err)
	}
}

func TestCov15_AddFormField_Signature(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldSignature,
		Name: "sig",
		X:    50,
		Y:    360,
		W:    200,
		H:    50,
	})
	if err != nil {
		t.Fatalf("AddFormField signature: %v", err)
	}
}

func TestCov15_AddFormField_EmptyName(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldText,
		Name: "",
		X:    50,
		Y:    100,
		W:    200,
		H:    25,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCov15_AddFormField_ZeroSize(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type: FormFieldText,
		Name: "test",
		X:    50,
		Y:    100,
		W:    0,
		H:    25,
	})
	if err == nil {
		t.Fatal("expected error for zero width")
	}
}

func TestCov15_AddTextField(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddTextField("email", 50, 100, 200, 25)
	if err != nil {
		t.Fatalf("AddTextField: %v", err)
	}
}

func TestCov15_AddCheckbox(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddCheckbox("agree", 50, 100, 15, true)
	if err != nil {
		t.Fatalf("AddCheckbox: %v", err)
	}
}

func TestCov15_AddDropdown(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddDropdown("color", 50, 100, 200, 25, []string{"Red", "Green", "Blue"})
	if err != nil {
		t.Fatalf("AddDropdown: %v", err)
	}
}

func TestCov15_GetFormFields(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	_ = pdf.AddTextField("f1", 50, 100, 200, 25)
	_ = pdf.AddCheckbox("f2", 50, 150, 15, false)

	fields := pdf.GetFormFields()
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
}

func TestCov15_FormFieldType_String(t *testing.T) {
	tests := []struct {
		ft   FormFieldType
		want string
	}{
		{FormFieldText, "Text"},
		{FormFieldCheckbox, "Checkbox"},
		{FormFieldRadio, "Radio"},
		{FormFieldChoice, "Choice"},
		{FormFieldButton, "Button"},
		{FormFieldSignature, "Signature"},
		{FormFieldType(99), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.ft.String()
		if got != tt.want {
			t.Errorf("FormFieldType(%d).String() = %q, want %q", tt.ft, got, tt.want)
		}
	}
}

func TestCov15_FormField_ReadOnlyRequired(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.AddFormField(FormField{
		Type:     FormFieldText,
		Name:     "readonly_field",
		X:        50,
		Y:        100,
		W:        200,
		H:        25,
		ReadOnly: true,
		Required: true,
		Value:    "Cannot edit",
	})
	if err != nil {
		t.Fatalf("AddFormField readonly+required: %v", err)
	}
}

// ============================================================
// pdf_obj_id.go — ObjID, GetObjID, GetObjByID, GetObjType
// ============================================================

func TestCov15_ObjID_Methods(t *testing.T) {
	id := ObjID(5)
	if id.Index() != 5 {
		t.Errorf("Index() = %d, want 5", id.Index())
	}
	if id.Ref() != 6 {
		t.Errorf("Ref() = %d, want 6", id.Ref())
	}
	if id.RefStr() != "6 0 R" {
		t.Errorf("RefStr() = %q, want '6 0 R'", id.RefStr())
	}
	if !id.IsValid() {
		t.Error("expected valid ObjID")
	}

	invalid := invalidObjID
	if invalid.IsValid() {
		t.Error("expected invalid ObjID")
	}
}

func TestCov15_GetObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	id := pdf.GetObjID(0)
	if !id.IsValid() {
		t.Error("expected valid ObjID for index 0")
	}

	id = pdf.GetObjID(-1)
	if id.IsValid() {
		t.Error("expected invalid ObjID for index -1")
	}

	id = pdf.GetObjID(99999)
	if id.IsValid() {
		t.Error("expected invalid ObjID for out-of-range index")
	}
}

func TestCov15_GetObjByID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	id := pdf.GetObjID(0)
	obj := pdf.GetObjByID(id)
	if obj == nil {
		t.Error("expected non-nil object")
	}

	obj = pdf.GetObjByID(invalidObjID)
	if obj != nil {
		t.Error("expected nil for invalid ObjID")
	}
}

func TestCov15_GetObjType(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	id := pdf.CatalogObjID()
	typ := pdf.GetObjType(id)
	if typ == "" {
		t.Error("expected non-empty type for catalog")
	}

	typ = pdf.GetObjType(invalidObjID)
	if typ != "" {
		t.Errorf("expected empty type for invalid ObjID, got %q", typ)
	}
}

func TestCov15_CatalogObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	id := pdf.CatalogObjID()
	if !id.IsValid() {
		t.Error("expected valid catalog ObjID")
	}
}

func TestCov15_PagesObjID(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	id := pdf.PagesObjID()
	if !id.IsValid() {
		t.Error("expected valid pages ObjID")
	}
}

// ============================================================
// svg_insert.go — ImageSVG, ImageSVGFromBytes, ImageSVGFromReader
// ============================================================

func TestCov15_ImageSVGFromBytes_Basic(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="red"/>
		<circle cx="50" cy="50" r="30" fill="blue"/>
		<line x1="0" y1="0" x2="100" y2="100" stroke="black" stroke-width="2"/>
	</svg>`)

	err := pdf.ImageSVGFromBytes(svgData, SVGOption{
		X:      50,
		Y:      50,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes: %v", err)
	}
}

func TestCov15_ImageSVGFromBytes_WidthOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="green"/>
	</svg>`)

	err := pdf.ImageSVGFromBytes(svgData, SVGOption{
		X:     50,
		Y:     50,
		Width: 200,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes width only: %v", err)
	}
}

func TestCov15_ImageSVGFromBytes_HeightOnly(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="green"/>
	</svg>`)

	err := pdf.ImageSVGFromBytes(svgData, SVGOption{
		X:      50,
		Y:      50,
		Height: 200,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromBytes height only: %v", err)
	}
}

func TestCov15_ImageSVGFromBytes_Empty(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"></svg>`)
	err := pdf.ImageSVGFromBytes(svgData, SVGOption{X: 50, Y: 50})
	if err == nil {
		t.Fatal("expected error for empty SVG")
	}
}

func TestCov15_ImageSVGFromBytes_Invalid(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVGFromBytes([]byte("not svg"), SVGOption{X: 50, Y: 50})
	if err == nil {
		t.Fatal("expected error for invalid SVG")
	}
}

func TestCov15_ImageSVGFromReader(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	svgData := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
		<rect x="10" y="10" width="80" height="80" fill="red"/>
	</svg>`

	err := pdf.ImageSVGFromReader(strings.NewReader(svgData), SVGOption{
		X:      50,
		Y:      50,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatalf("ImageSVGFromReader: %v", err)
	}
}

func TestCov15_ImageSVG_InvalidPath(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.ImageSVG("/nonexistent/file.svg", SVGOption{X: 50, Y: 50})
	if err == nil {
		t.Fatal("expected error for invalid SVG path")
	}
}

// ============================================================
// cache_content_text_color_cmyk.go — equal
// ============================================================

func TestCov15_CMYK_Color(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)

	// Set CMYK text color.
	pdf.SetTextColorCMYK(100, 0, 0, 0) // cyan
	_ = pdf.Text("CMYK text")

	// Set different CMYK color.
	pdf.SetTextColorCMYK(0, 100, 0, 0) // magenta
	_ = pdf.Text("Magenta text")

	// Set same CMYK color again (triggers equal check).
	pdf.SetTextColorCMYK(0, 100, 0, 0)
	_ = pdf.Text("Same magenta")
}

func TestCov15_CMYK_StrokeColor(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	pdf.SetStrokeColorCMYK(0, 0, 100, 0) // yellow stroke
	pdf.Line(10, 10, 200, 200)

	pdf.SetStrokeColorCMYK(0, 0, 0, 100) // black stroke
	pdf.Line(20, 20, 210, 210)
}

// ============================================================
// embedded_file.go — UpdateEmbeddedFile
// ============================================================

func TestCov15_UpdateEmbeddedFile(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Embedded file test")

	// Add an embedded file first.
	err := pdf.AddEmbeddedFile(EmbeddedFile{
		Name:     "test.txt",
		Content:  []byte("Hello World"),
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatalf("AddEmbeddedFile: %v", err)
	}

	// Update it.
	err = pdf.UpdateEmbeddedFile("test.txt", EmbeddedFile{
		Name:     "test.txt",
		Content:  []byte("Updated content"),
		MimeType: "text/plain",
	})
	if err != nil {
		t.Fatalf("UpdateEmbeddedFile: %v", err)
	}
}

func TestCov15_UpdateEmbeddedFile_NotFound(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()

	err := pdf.UpdateEmbeddedFile("nonexistent.txt", EmbeddedFile{
		Name:    "nonexistent.txt",
		Content: []byte("data"),
	})
	if err == nil {
		t.Fatal("expected error for non-existent embedded file")
	}
}

// ============================================================
// Write form fields to PDF output (triggers formFieldObj.write)
// ============================================================

func TestCov15_FormField_WritePDF(t *testing.T) {
	pdf := newPDFWithFont(t)
	pdf.AddPage()
	pdf.SetXY(50, 50)
	_ = pdf.Text("Form test")

	_ = pdf.AddTextField("name", 50, 100, 200, 25)
	_ = pdf.AddCheckbox("agree", 50, 150, 15, true)
	_ = pdf.AddDropdown("color", 50, 200, 200, 25, []string{"Red", "Green"})
	_ = pdf.AddFormField(FormField{
		Type: FormFieldButton, Name: "btn", X: 50, Y: 250, W: 100, H: 30,
	})
	_ = pdf.AddFormField(FormField{
		Type: FormFieldRadio, Name: "radio", X: 50, Y: 300, W: 15, H: 15,
	})
	_ = pdf.AddFormField(FormField{
		Type: FormFieldSignature, Name: "sig", X: 50, Y: 350, W: 200, H: 50,
	})

	var buf bytes.Buffer
	_, err := pdf.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo with form fields: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty PDF output")
	}
}

// ============================================================
// nullObj — write
// ============================================================

func TestCov15_NullObj_Write(t *testing.T) {
	n := nullObj{}
	var buf bytes.Buffer
	err := n.write(&buf, 1)
	if err != nil {
		t.Fatalf("nullObj.write: %v", err)
	}
	if !strings.Contains(buf.String(), "null") {
		t.Error("expected 'null' in output")
	}
	if n.getType() != "Null" {
		t.Errorf("getType() = %q, want 'Null'", n.getType())
	}
}
