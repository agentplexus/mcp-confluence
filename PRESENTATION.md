---
marp: true
theme: agentplexus
paginate: true
style: |
  @import '../agentplexus-assets-internal/agentplexus.css';
---

# mcp-confluence

## Building a Reliable MCP Server for Confluence

---

# The Problem

AI assistants **corrupt Confluence pages** when editing them

- Tables lose formatting or become invalid
- Macros break or disappear
- Content gets mangled on round-trip

This happens with **official MCP servers** and third-party solutions

---

# Why Does This Happen?

Confluence uses **Storage Format XHTML** - not HTML5, not Markdown

```xml
<table>
  <tbody>           <!-- Required! No <thead> allowed -->
    <tr>
      <th>Name</th> <!-- Headers inside tbody -->
    </tr>
    <tr>
      <td>Alice</td>
    </tr>
  </tbody>
</table>
```

LLMs generate HTML5 or Markdown instead → **instant corruption**

---

# Real-World Failures

### What we observed:

1. **Tables edited in Atlassian Cloud** couldn't be read back correctly
2. **Official MCP servers** would corrupt pages on update
3. **Round-trip editing** (read → modify → write) lost data
4. **Macros** with `ac:` namespaces were stripped or broken

### The root cause:

Servers were converting to/from Markdown or HTML5 internally

---

# Our First Approach: Structured Blocks

Idea: Use an **Intermediate Representation (IR)** instead of raw XHTML

```go
page := &storage.Page{
    Blocks: []storage.Block{
        &storage.Heading{Level: 1, Text: "Title"},
        &storage.Table{
            Headers: []string{"Name", "Status"},
            Rows:    []storage.Row{{Cells: []storage.Cell{{Text: "Alice"}}}},
        },
    },
}
```

LLM produces JSON → Go renders valid XHTML

---

# Structured Blocks: Results

### What worked:
- Creating new pages from scratch
- Simple tables, lists, headings
- Guaranteed valid XHTML output

### What didn't work:
- **Complex tables** lost column widths, styles, attributes
- **Nested content** in cells (lists, bold, links) was flattened
- **Round-trip editing** still lost information

---

# The Core Issue

Confluence tables are **much richer** than our IR could represent:

```xml
<table data-layout="wide">
  <colgroup>
    <col style="width: 200px"/>
    <col style="width: 400px"/>
  </colgroup>
  <tbody>
    <tr>
      <td>
        <p><strong>Bold</strong> and <a href="...">links</a></p>
        <ul><li>Nested list</li></ul>
      </td>
    </tr>
  </tbody>
</table>
```

Our IR → `{text: "Bold and links Nested list"}`

---

# Second Approach: Raw XHTML Tools

Added tools that work directly with Storage Format XHTML:

| Tool | Description |
|------|-------------|
| `confluence_read_page_xhtml` | Get raw XHTML |
| `confluence_update_page_xhtml` | Update with raw XHTML |

Let the LLM work with the actual format

---

# Raw XHTML: Results

### What worked:
- **Perfect round-trip** - nothing lost
- **Complex tables** preserved exactly
- **All attributes** maintained (widths, styles, IDs)
- **Nested content** kept intact

### Tradeoffs:
- LLM must understand Storage Format XHTML
- More tokens in context
- Risk of LLM generating invalid XHTML

---

# Current Recommendation

## Use XHTML tools for **editing existing pages**

```
1. confluence_read_page_xhtml  → get raw XHTML
2. LLM modifies the XHTML string
3. confluence_update_page_xhtml → save changes
```

Preserves everything, no data loss

---

# Current Recommendation

## Use structured blocks for **creating new pages**

```
1. LLM generates JSON blocks
2. confluence_create_page → renders valid XHTML
```

Simpler, guaranteed valid output

---

# Decision Matrix

| Scenario | Recommended Tool |
|----------|------------------|
| Create new page | `confluence_create_page` (blocks) |
| Read simple page | `confluence_read_page` (blocks) |
| Read complex page | `confluence_read_page_xhtml` |
| Edit existing page | `confluence_update_page_xhtml` |
| Edit tables | **Always** `confluence_update_page_xhtml` |

---

# Why Not Always XHTML?

Structured blocks are still valuable:

1. **Safer for creation** - can't produce invalid XHTML
2. **Easier for LLMs** - JSON is more natural than XHTML
3. **Validation built-in** - catches errors before API call
4. **Simpler prompts** - no need to explain Storage Format

Use the right tool for the job

---

# Architecture

```
┌─────────────────────────────────────────────────────┐
│                    MCP Server                        │
├─────────────────────────────────────────────────────┤
│  Structured Tools          │  XHTML Tools           │
│  ─────────────────         │  ────────────          │
│  read_page (blocks)        │  read_page_xhtml       │
│  update_page (blocks)      │  update_page_xhtml     │
│  create_page (blocks)      │                        │
├─────────────────────────────────────────────────────┤
│              Confluence REST API Client              │
├─────────────────────────────────────────────────────┤
│                 Storage Package                      │
│  Parse (XHTML→IR) │ Render (IR→XHTML) │ Validate    │
└─────────────────────────────────────────────────────┘
```

---

# Key Learnings

1. **Don't fight the format** - Confluence uses XHTML, work with it
2. **Lossless round-trip matters** - users notice when formatting disappears
3. **Multiple tools > one tool** - different scenarios need different approaches
4. **Validation is essential** - catch errors before they corrupt pages

---

# Summary

| Problem | Solution |
|---------|----------|
| LLMs generate wrong format | Structured blocks → valid XHTML |
| Editing loses formatting | Raw XHTML tools preserve everything |
| Tables break on round-trip | Always use XHTML for table edits |
| Official servers corrupt data | We built our own |

---

# Links

- **Repository**: github.com/agentplexus/mcp-confluence
- **Confluence Storage Format**: [Atlassian Docs](https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html)

---

# Questions?
