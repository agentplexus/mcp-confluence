package storage

import (
	"strings"
	"testing"
)

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name     string
		table    *Table
		wantErr  bool
		contains []string
	}{
		{
			name: "simple table with headers and rows",
			table: &Table{
				Headers: []string{"Name", "Age", "City"},
				Rows: []Row{
					{Cells: []Cell{{Text: "Alice"}, {Text: "30"}, {Text: "NYC"}}},
					{Cells: []Cell{{Text: "Bob"}, {Text: "25"}, {Text: "LA"}}},
				},
			},
			wantErr: false,
			contains: []string{
				"<table><tbody>",
				"<th>Name</th>",
				"<th>Age</th>",
				"<th>City</th>",
				"<td>Alice</td>",
				"<td>30</td>",
				"<td>NYC</td>",
				"<td>Bob</td>",
				"</tbody></table>",
			},
		},
		{
			name: "table with macro in cell",
			table: &Table{
				Headers: []string{"Status"},
				Rows: []Row{
					{Cells: []Cell{{Macro: &Macro{
						Name:   "status",
						Params: map[string]string{"colour": "Green", "title": "OK"},
					}}}},
				},
			},
			wantErr: false,
			contains: []string{
				`<ac:structured-macro ac:name="status">`,
				`<ac:parameter ac:name=`,
				`</ac:structured-macro>`,
			},
		},
		{
			name:     "empty table",
			table:    &Table{},
			wantErr:  false,
			contains: []string{"<table><tbody></tbody></table>"},
		},
		{
			name:     "nil table",
			table:    nil,
			wantErr:  false,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTable(tt.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("renderTable() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestRenderParagraph(t *testing.T) {
	tests := []struct {
		name    string
		p       *Paragraph
		want    string
		wantErr bool
	}{
		{
			name:    "simple paragraph",
			p:       &Paragraph{Text: "Hello, World!"},
			want:    "<p>Hello, World!</p>",
			wantErr: false,
		},
		{
			name:    "paragraph with special characters",
			p:       &Paragraph{Text: "Hello <World> & \"Friends\""},
			want:    "<p>Hello &lt;World&gt; &amp; &#34;Friends&#34;</p>",
			wantErr: false,
		},
		{
			name:    "empty paragraph",
			p:       &Paragraph{Text: ""},
			want:    "<p></p>",
			wantErr: false,
		},
		{
			name:    "nil paragraph",
			p:       nil,
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderParagraph(tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderParagraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("renderParagraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderHeading(t *testing.T) {
	tests := []struct {
		name    string
		h       *Heading
		want    string
		wantErr bool
	}{
		{
			name:    "h1",
			h:       &Heading{Level: 1, Text: "Title"},
			want:    "<h1>Title</h1>",
			wantErr: false,
		},
		{
			name:    "h3",
			h:       &Heading{Level: 3, Text: "Subtitle"},
			want:    "<h3>Subtitle</h3>",
			wantErr: false,
		},
		{
			name:    "h6",
			h:       &Heading{Level: 6, Text: "Small heading"},
			want:    "<h6>Small heading</h6>",
			wantErr: false,
		},
		{
			name:    "invalid level 0",
			h:       &Heading{Level: 0, Text: "Bad"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid level 7",
			h:       &Heading{Level: 7, Text: "Bad"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "nil heading",
			h:       nil,
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderHeading(tt.h)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderHeading() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("renderHeading() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderMacro(t *testing.T) {
	tests := []struct {
		name     string
		m        *Macro
		wantErr  bool
		contains []string
	}{
		{
			name: "status macro",
			m: &Macro{
				Name:   "status",
				Params: map[string]string{"colour": "Green", "title": "OK"},
			},
			wantErr: false,
			contains: []string{
				`<ac:structured-macro ac:name="status">`,
				`<ac:parameter ac:name=`,
				`</ac:structured-macro>`,
			},
		},
		{
			name: "info macro with body",
			m: &Macro{
				Name: "info",
				Body: "<p>This is important</p>",
			},
			wantErr: false,
			contains: []string{
				`<ac:structured-macro ac:name="info">`,
				`<ac:rich-text-body>`,
				`<p>This is important</p>`,
				`</ac:rich-text-body>`,
			},
		},
		{
			name: "code macro",
			m: &Macro{
				Name:   "code",
				Params: map[string]string{"language": "go"},
			},
			wantErr: false,
			contains: []string{
				`<ac:structured-macro ac:name="code">`,
				`<ac:parameter ac:name="language">go</ac:parameter>`,
			},
		},
		{
			name:     "nil macro",
			m:        nil,
			wantErr:  false,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderMacro(tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderMacro() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("RenderMacro() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestRenderBulletList(t *testing.T) {
	tests := []struct {
		name    string
		bl      *BulletList
		want    string
		wantErr bool
	}{
		{
			name: "simple list",
			bl: &BulletList{
				Items: []ListItem{{Text: "First"}, {Text: "Second"}, {Text: "Third"}},
			},
			want:    "<ul><li>First</li><li>Second</li><li>Third</li></ul>",
			wantErr: false,
		},
		{
			name:    "empty list",
			bl:      &BulletList{Items: []ListItem{}},
			want:    "<ul></ul>",
			wantErr: false,
		},
		{
			name:    "nil list",
			bl:      nil,
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderBulletList(tt.bl)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderBulletList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("renderBulletList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderNumberedList(t *testing.T) {
	tests := []struct {
		name    string
		nl      *NumberedList
		want    string
		wantErr bool
	}{
		{
			name: "simple list",
			nl: &NumberedList{
				Items: []ListItem{{Text: "First"}, {Text: "Second"}},
			},
			want:    "<ol><li>First</li><li>Second</li></ol>",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderNumberedList(tt.nl)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderNumberedList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("renderNumberedList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		cb       *CodeBlock
		wantErr  bool
		contains []string
	}{
		{
			name: "code block with language",
			cb: &CodeBlock{
				Language: "go",
				Code:     "func main() {}",
			},
			wantErr: false,
			contains: []string{
				`<ac:structured-macro ac:name="code">`,
				`<ac:parameter ac:name="language">go</ac:parameter>`,
				`<ac:plain-text-body><![CDATA[func main() {}]]></ac:plain-text-body>`,
			},
		},
		{
			name: "code block without language",
			cb: &CodeBlock{
				Code: "echo hello",
			},
			wantErr: false,
			contains: []string{
				`<ac:structured-macro ac:name="code">`,
				`<![CDATA[echo hello]]>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderCodeBlock(tt.cb)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderCodeBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("renderCodeBlock() = %v, want to contain %v", got, want)
				}
			}
		})
	}
}

func TestRenderBlock(t *testing.T) {
	tests := []struct {
		name    string
		block   Block
		wantErr bool
	}{
		{name: "table pointer", block: &Table{Headers: []string{"A"}}, wantErr: false},
		{name: "table value", block: Table{Headers: []string{"A"}}, wantErr: false},
		{name: "paragraph pointer", block: &Paragraph{Text: "test"}, wantErr: false},
		{name: "paragraph value", block: Paragraph{Text: "test"}, wantErr: false},
		{name: "heading pointer", block: &Heading{Level: 1, Text: "test"}, wantErr: false},
		{name: "heading value", block: Heading{Level: 1, Text: "test"}, wantErr: false},
		{name: "macro pointer", block: &Macro{Name: "test"}, wantErr: false},
		{name: "macro value", block: Macro{Name: "test"}, wantErr: false},
		{name: "bullet list pointer", block: &BulletList{}, wantErr: false},
		{name: "bullet list value", block: BulletList{}, wantErr: false},
		{name: "numbered list pointer", block: &NumberedList{}, wantErr: false},
		{name: "numbered list value", block: NumberedList{}, wantErr: false},
		{name: "code block pointer", block: &CodeBlock{Code: "x"}, wantErr: false},
		{name: "code block value", block: CodeBlock{Code: "x"}, wantErr: false},
		{name: "horizontal rule pointer", block: &HorizontalRule{}, wantErr: false},
		{name: "horizontal rule value", block: HorizontalRule{}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RenderBlock(tt.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRender(t *testing.T) {
	page := &Page{
		Blocks: []Block{
			&Heading{Level: 1, Text: "Welcome"},
			&Paragraph{Text: "This is a test page."},
			&Table{
				Headers: []string{"Col1", "Col2"},
				Rows:    []Row{{Cells: []Cell{{Text: "A"}, {Text: "B"}}}},
			},
		},
	}

	got, err := Render(page)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Verify order and content
	expected := []string{
		"<h1>Welcome</h1>",
		"<p>This is a test page.</p>",
		"<table><tbody>",
		"<th>Col1</th>",
		"<td>A</td>",
	}

	for _, want := range expected {
		if !strings.Contains(got, want) {
			t.Errorf("Render() missing expected content: %v", want)
		}
	}

	// Verify nil page
	got, err = Render(nil)
	if err != nil || got != "" {
		t.Errorf("Render(nil) = %v, %v, want empty string, nil", got, err)
	}
}
