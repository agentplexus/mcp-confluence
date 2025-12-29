package storage

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		xhtml   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid table with tbody",
			xhtml:   "<table><tbody><tr><th>Header</th></tr><tr><td>Data</td></tr></tbody></table>",
			wantErr: false,
		},
		{
			name:    "valid paragraph",
			xhtml:   "<p>Hello, World!</p>",
			wantErr: false,
		},
		{
			name:    "valid heading",
			xhtml:   "<h1>Title</h1>",
			wantErr: false,
		},
		{
			name:    "valid list",
			xhtml:   "<ul><li>Item 1</li><li>Item 2</li></ul>",
			wantErr: false,
		},
		{
			name:    "valid macro",
			xhtml:   `<ac:structured-macro ac:name="status"><ac:parameter ac:name="colour">Green</ac:parameter></ac:structured-macro>`,
			wantErr: false,
		},
		{
			name:    "empty string",
			xhtml:   "",
			wantErr: false,
		},
		{
			name:    "table without tbody",
			xhtml:   "<table><tr><td>Data</td></tr></table>",
			wantErr: true,
			errMsg:  "<tr> must be inside <tbody>",
		},
		{
			name:    "forbidden tag thead",
			xhtml:   "<table><thead><tr><th>Header</th></tr></thead></table>",
			wantErr: true,
			errMsg:  "forbidden tag",
		},
		{
			name:    "forbidden tag div",
			xhtml:   "<div>Content</div>",
			wantErr: true,
			errMsg:  "forbidden tag",
		},
		{
			name:    "forbidden tag span",
			xhtml:   "<p><span>Text</span></p>",
			wantErr: true,
			errMsg:  "forbidden tag",
		},
		{
			name:    "forbidden tag script",
			xhtml:   "<script>alert('xss')</script>",
			wantErr: true,
			errMsg:  "forbidden tag",
		},
		{
			name:    "forbidden tag style",
			xhtml:   "<style>.x{color:red}</style>",
			wantErr: true,
			errMsg:  "forbidden tag",
		},
		{
			name:    "forbidden tag iframe",
			xhtml:   "<iframe src=\"http://evil.com\"></iframe>",
			wantErr: true,
			errMsg:  "forbidden tag",
		},
		{
			name:    "malformed XML",
			xhtml:   "<p>Unclosed paragraph",
			wantErr: true,
			errMsg:  "invalid XML",
		},
		{
			name:    "mismatched tags",
			xhtml:   "<p>Text</div>",
			wantErr: true,
			errMsg:  "invalid XML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.xhtml)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if ve, ok := err.(*ValidationError); ok {
					if ve.Message != tt.errMsg && ve.Message[:len(tt.errMsg)] != tt.errMsg {
						// Check if error message contains expected substring
						found := false
						if len(ve.Message) >= len(tt.errMsg) {
							for i := 0; i <= len(ve.Message)-len(tt.errMsg); i++ {
								if ve.Message[i:i+len(tt.errMsg)] == tt.errMsg {
									found = true
									break
								}
							}
						}
						if !found {
							t.Errorf("Validate() error message = %v, want to contain %v", ve.Message, tt.errMsg)
						}
					}
				}
			}
		})
	}
}

func TestValidateWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		xhtml   string
		opts    ValidatorOptions
		wantErr bool
	}{
		{
			name:  "table without tbody allowed when RequireTableTbody is false",
			xhtml: "<table><tr><td>Data</td></tr></table>",
			opts: ValidatorOptions{
				RequireTableTbody: false,
				ForbiddenTags:     ForbiddenTags,
			},
			wantErr: false,
		},
		{
			name:  "allowed macro passes",
			xhtml: `<ac:structured-macro ac:name="status"><ac:parameter ac:name="colour">Green</ac:parameter></ac:structured-macro>`,
			opts: ValidatorOptions{
				RequireTableTbody: true,
				AllowedMacros:     map[string]bool{"status": true, "info": true},
				ForbiddenTags:     ForbiddenTags,
			},
			wantErr: false,
		},
		{
			name:  "disallowed macro fails",
			xhtml: `<ac:structured-macro ac:name="danger"><ac:parameter ac:name="title">Bad</ac:parameter></ac:structured-macro>`,
			opts: ValidatorOptions{
				RequireTableTbody: true,
				AllowedMacros:     map[string]bool{"status": true, "info": true},
				ForbiddenTags:     ForbiddenTags,
			},
			wantErr: true,
		},
		{
			name:  "custom forbidden tags",
			xhtml: "<custom>Content</custom>",
			opts: ValidatorOptions{
				RequireTableTbody: true,
				ForbiddenTags:     map[string]bool{"custom": true},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithOptions(tt.xhtml, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWithOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBlock(t *testing.T) {
	tests := []struct {
		name    string
		block   Block
		wantErr bool
	}{
		{
			name: "valid table block",
			block: &Table{
				Headers: []string{"A", "B"},
				Rows:    []Row{{Cells: []Cell{{Text: "1"}, {Text: "2"}}}},
			},
			wantErr: false,
		},
		{
			name:    "valid paragraph block",
			block:   &Paragraph{Text: "Hello"},
			wantErr: false,
		},
		{
			name:    "valid heading block",
			block:   &Heading{Level: 2, Text: "Title"},
			wantErr: false,
		},
		{
			name:    "invalid heading level",
			block:   &Heading{Level: 0, Text: "Bad"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlock(tt.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidXML(t *testing.T) {
	tests := []struct {
		name  string
		xhtml string
		want  bool
	}{
		{
			name:  "valid XML",
			xhtml: "<p>Hello</p>",
			want:  true,
		},
		{
			name:  "valid nested XML",
			xhtml: "<div><p>Hello</p></div>",
			want:  true,
		},
		{
			name:  "valid self-closing",
			xhtml: "<hr/>",
			want:  true,
		},
		{
			name:  "empty string",
			xhtml: "",
			want:  true,
		},
		{
			name:  "invalid unclosed tag",
			xhtml: "<p>Hello",
			want:  false,
		},
		{
			name:  "invalid mismatched tags",
			xhtml: "<p>Hello</div>",
			want:  false,
		},
		{
			name:  "invalid ampersand",
			xhtml: "<p>Hello & World</p>",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidXML(tt.xhtml)
			if got != tt.want {
				t.Errorf("IsValidXML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "error with tag",
			err:  &ValidationError{Message: "forbidden tag", Tag: "div"},
			want: "validation error: forbidden tag (tag: div)",
		},
		{
			name: "error without tag",
			err:  &ValidationError{Message: "invalid XML"},
			want: "validation error: invalid XML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustValidate(t *testing.T) {
	// Should not panic for valid XHTML
	MustValidate("<p>Valid</p>")

	// Should panic for invalid XHTML
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustValidate() did not panic for invalid XHTML")
		}
	}()
	MustValidate("<div>Invalid</div>")
}
