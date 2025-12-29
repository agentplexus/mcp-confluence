package storage

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		xhtml      string
		wantErr    bool
		wantBlocks int
	}{
		{
			name:       "empty string",
			xhtml:      "",
			wantErr:    false,
			wantBlocks: 0,
		},
		{
			name:       "single paragraph",
			xhtml:      "<p>Hello, World!</p>",
			wantErr:    false,
			wantBlocks: 1,
		},
		{
			name:       "multiple paragraphs",
			xhtml:      "<p>First</p><p>Second</p><p>Third</p>",
			wantErr:    false,
			wantBlocks: 3,
		},
		{
			name:       "heading and paragraph",
			xhtml:      "<h1>Title</h1><p>Content</p>",
			wantErr:    false,
			wantBlocks: 2,
		},
		{
			name:       "all heading levels",
			xhtml:      "<h1>H1</h1><h2>H2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6>",
			wantErr:    false,
			wantBlocks: 6,
		},
		{
			name:       "bullet list",
			xhtml:      "<ul><li>Item 1</li><li>Item 2</li></ul>",
			wantErr:    false,
			wantBlocks: 1,
		},
		{
			name:       "numbered list",
			xhtml:      "<ol><li>First</li><li>Second</li></ol>",
			wantErr:    false,
			wantBlocks: 1,
		},
		{
			name:       "simple table",
			xhtml:      "<table><tbody><tr><th>Header</th></tr><tr><td>Data</td></tr></tbody></table>",
			wantErr:    false,
			wantBlocks: 1,
		},
		{
			name:       "horizontal rule",
			xhtml:      "<p>Before</p><hr/><p>After</p>",
			wantErr:    false,
			wantBlocks: 3,
		},
		{
			name:       "invalid XML",
			xhtml:      "<p>Unclosed",
			wantErr:    true,
			wantBlocks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := Parse(tt.xhtml)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(page.Blocks) != tt.wantBlocks {
				t.Errorf("Parse() blocks = %d, want %d", len(page.Blocks), tt.wantBlocks)
			}
		})
	}
}

func TestParseParagraph(t *testing.T) {
	xhtml := "<p>Hello, World!</p>"
	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(page.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
	}

	p, ok := page.Blocks[0].(*Paragraph)
	if !ok {
		t.Fatalf("Parse() block type = %T, want *Paragraph", page.Blocks[0])
	}

	if p.Text != "Hello, World!" {
		t.Errorf("Paragraph.Text = %v, want Hello, World!", p.Text)
	}
}

func TestParseHeading(t *testing.T) {
	tests := []struct {
		xhtml     string
		wantLevel int
		wantText  string
	}{
		{"<h1>Title</h1>", 1, "Title"},
		{"<h2>Subtitle</h2>", 2, "Subtitle"},
		{"<h3>Section</h3>", 3, "Section"},
		{"<h4>Subsection</h4>", 4, "Subsection"},
		{"<h5>Minor</h5>", 5, "Minor"},
		{"<h6>Smallest</h6>", 6, "Smallest"},
	}

	for _, tt := range tests {
		t.Run(tt.xhtml, func(t *testing.T) {
			page, err := Parse(tt.xhtml)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(page.Blocks) != 1 {
				t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
			}

			h, ok := page.Blocks[0].(*Heading)
			if !ok {
				t.Fatalf("Parse() block type = %T, want *Heading", page.Blocks[0])
			}

			if h.Level != tt.wantLevel {
				t.Errorf("Heading.Level = %d, want %d", h.Level, tt.wantLevel)
			}

			if h.Text != tt.wantText {
				t.Errorf("Heading.Text = %v, want %v", h.Text, tt.wantText)
			}
		})
	}
}

func TestParseBulletList(t *testing.T) {
	xhtml := "<ul><li>First</li><li>Second</li><li>Third</li></ul>"
	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(page.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
	}

	bl, ok := page.Blocks[0].(*BulletList)
	if !ok {
		t.Fatalf("Parse() block type = %T, want *BulletList", page.Blocks[0])
	}

	if len(bl.Items) != 3 {
		t.Fatalf("BulletList.Items = %d, want 3", len(bl.Items))
	}

	expected := []string{"First", "Second", "Third"}
	for i, item := range bl.Items {
		if item.Text != expected[i] {
			t.Errorf("BulletList.Items[%d].Text = %v, want %v", i, item.Text, expected[i])
		}
	}
}

func TestParseNumberedList(t *testing.T) {
	xhtml := "<ol><li>One</li><li>Two</li></ol>"
	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(page.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
	}

	nl, ok := page.Blocks[0].(*NumberedList)
	if !ok {
		t.Fatalf("Parse() block type = %T, want *NumberedList", page.Blocks[0])
	}

	if len(nl.Items) != 2 {
		t.Fatalf("NumberedList.Items = %d, want 2", len(nl.Items))
	}
}

func TestParseTable(t *testing.T) {
	xhtml := `<table><tbody>
		<tr><th>Name</th><th>Age</th></tr>
		<tr><td>Alice</td><td>30</td></tr>
		<tr><td>Bob</td><td>25</td></tr>
	</tbody></table>`

	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(page.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
	}

	table, ok := page.Blocks[0].(*Table)
	if !ok {
		t.Fatalf("Parse() block type = %T, want *Table", page.Blocks[0])
	}

	if len(table.Headers) != 2 {
		t.Errorf("Table.Headers = %d, want 2", len(table.Headers))
	}

	if len(table.Rows) != 2 {
		t.Errorf("Table.Rows = %d, want 2", len(table.Rows))
	}

	if table.Headers[0] != "Name" {
		t.Errorf("Table.Headers[0] = %v, want Name", table.Headers[0])
	}

	if table.Rows[0].Cells[0].Text != "Alice" {
		t.Errorf("Table.Rows[0].Cells[0].Text = %v, want Alice", table.Rows[0].Cells[0].Text)
	}
}

func TestParseTableWithMacro(t *testing.T) {
	xhtml := `<table><tbody>
		<tr><th>Status</th></tr>
		<tr><td><ac:structured-macro ac:name="status"><ac:parameter ac:name="colour">Green</ac:parameter><ac:parameter ac:name="title">OK</ac:parameter></ac:structured-macro></td></tr>
	</tbody></table>`

	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(page.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
	}

	table, ok := page.Blocks[0].(*Table)
	if !ok {
		t.Fatalf("Parse() block type = %T, want *Table", page.Blocks[0])
	}

	if len(table.Rows) != 1 {
		t.Fatalf("Table.Rows = %d, want 1", len(table.Rows))
	}

	cell := table.Rows[0].Cells[0]
	if cell.Macro == nil {
		t.Fatal("Cell.Macro is nil, want macro")
	}

	if cell.Macro.Name != "status" {
		t.Errorf("Macro.Name = %v, want status", cell.Macro.Name)
	}

	if cell.Macro.Params["colour"] != "Green" {
		t.Errorf("Macro.Params[colour] = %v, want Green", cell.Macro.Params["colour"])
	}
}

func TestParseMacro(t *testing.T) {
	xhtml := `<ac:structured-macro ac:name="info"><ac:parameter ac:name="title">Note</ac:parameter><ac:rich-text-body>Important information</ac:rich-text-body></ac:structured-macro>`

	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(page.Blocks) != 1 {
		t.Fatalf("Parse() blocks = %d, want 1", len(page.Blocks))
	}

	macro, ok := page.Blocks[0].(*Macro)
	if !ok {
		t.Fatalf("Parse() block type = %T, want *Macro", page.Blocks[0])
	}

	if macro.Name != "info" {
		t.Errorf("Macro.Name = %v, want info", macro.Name)
	}

	if macro.Params["title"] != "Note" {
		t.Errorf("Macro.Params[title] = %v, want Note", macro.Params["title"])
	}

	if macro.Body != "Important information" {
		t.Errorf("Macro.Body = %v, want 'Important information'", macro.Body)
	}
}

func TestMultiTablePage(t *testing.T) {
	// Page with multiple tables
	xhtml := `<h1>Status Report</h1>
<p>Overview of services:</p>
<table><tbody>
<tr><th>Service</th><th>Status</th></tr>
<tr><td>Auth</td><td>OK</td></tr>
<tr><td>API</td><td>OK</td></tr>
</tbody></table>
<h2>Team Members</h2>
<table><tbody>
<tr><th>Name</th><th>Role</th></tr>
<tr><td>Alice</td><td>Lead</td></tr>
<tr><td>Bob</td><td>Dev</td></tr>
</tbody></table>
<p>End of report.</p>`

	// Parse
	page, err := Parse(xhtml)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Count tables
	tableCount := 0
	for _, block := range page.Blocks {
		if _, ok := block.(*Table); ok {
			tableCount++
		}
	}
	if tableCount != 2 {
		t.Errorf("Expected 2 tables, got %d", tableCount)
	}

	// Modify second table - add a row
	for _, block := range page.Blocks {
		if table, ok := block.(*Table); ok {
			if len(table.Headers) > 0 && table.Headers[0] == "Name" {
				table.Rows = append(table.Rows, Row{
					Cells: []Cell{{Text: "Carol"}, {Text: "QA"}},
				})
			}
		}
	}

	// Render back
	result, err := Render(page)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Validate
	if err := Validate(result); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Verify Carol was added
	if !strings.Contains(result, "Carol") {
		t.Error("Modified content (Carol) not found in rendered output")
	}
	if !strings.Contains(result, "QA") {
		t.Error("Modified content (QA) not found in rendered output")
	}

	// Verify both tables are present
	tableTagCount := strings.Count(result, "<table>")
	if tableTagCount != 2 {
		t.Errorf("Expected 2 <table> tags, got %d", tableTagCount)
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that render -> parse -> render produces consistent output
	original := &Page{
		Blocks: []Block{
			&Heading{Level: 1, Text: "Title"},
			&Paragraph{Text: "Introduction paragraph."},
			&BulletList{Items: []ListItem{{Text: "Item 1"}, {Text: "Item 2"}}},
			&Table{
				Headers: []string{"Name", "Value"},
				Rows: []Row{
					{Cells: []Cell{{Text: "Key"}, {Text: "123"}}},
				},
			},
		},
	}

	// Render to XHTML
	xhtml1, err := Render(original)
	if err != nil {
		t.Fatalf("First Render() error = %v", err)
	}

	// Parse back to Page
	parsed, err := Parse(xhtml1)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Render again
	xhtml2, err := Render(parsed)
	if err != nil {
		t.Fatalf("Second Render() error = %v", err)
	}

	// Compare
	if xhtml1 != xhtml2 {
		t.Errorf("Round trip mismatch:\nFirst:  %v\nSecond: %v", xhtml1, xhtml2)
	}

	// Validate both
	if err := Validate(xhtml1); err != nil {
		t.Errorf("First XHTML validation failed: %v", err)
	}
	if err := Validate(xhtml2); err != nil {
		t.Errorf("Second XHTML validation failed: %v", err)
	}
}
