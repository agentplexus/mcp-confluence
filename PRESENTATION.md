---
marp: true
theme: agentplexus
paginate: true
style: |
  @import '../agentplexus-assets-internal/agentplexus.css';
  section.section-header {
    display: flex;
    flex-direction: column;
    justify-content: center;
    text-align: center;
  }
  section.section-header h1 {
    font-size: 2.5em;
  }
---

<!-- _paginate: false -->

# mcp-confluence

## Building a Reliable MCP Server for Confluence 🔧

*How we solved the table corruption problem*

---

<!-- _class: section-header -->

# 1️⃣ The Problem

---

# Real-World Problems 🔥

When we started using existing MCP servers, we discovered:

1. 📊 **Tables** lost formatting or became invalid
2. 🧩 **Macros** with `ac:` namespaces were stripped or broken
3. 🔄 **Round-trip editing** (read → modify → write) lost data
4. 🌐 **Web UI edits** created XHTML the MCP server couldn't parse — even on pages the MCP server originally created

The root cause:

Servers were converting to/from Markdown or HTML5 internally

---

# Why Does This Happen? 🤔

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

LLMs generate HTML5 or Markdown instead → **instant corruption** 💥

---

<!-- _class: section-header -->

# 2️⃣ First Approach: Structured Blocks

---

# Structured Blocks 🧱

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

LLM produces JSON → Go renders valid XHTML ✅

---

# Structured Blocks: Results 📊

### What worked: ✅
- Creating new pages from scratch
- Simple tables, lists, headings
- Guaranteed valid XHTML output

### What didn't work: ❌
- **Complex tables** lost column widths, styles, attributes
- **Nested content** in cells (lists, bold, links) was flattened
- **Round-trip editing** still lost information

---

# The Core Issue ⚠️

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

Our block format → `{text: "Bold and links Nested list"}` 😬

---

<!-- _class: section-header -->

# 3️⃣ Second Approach: Raw XHTML

---

# Raw XHTML Tools 🛠️

Added tools that work directly with Storage Format XHTML:

| Tool | Description |
|------|-------------|
| `confluence_read_page_xhtml` | 📖 Get raw XHTML |
| `confluence_update_page_xhtml` | ✏️ Update with raw XHTML |

Let the LLM work with the actual format

---

# Raw XHTML: Results 🎯

### What worked: ✅
- **Perfect round-trip** - nothing lost
- **Complex tables** preserved exactly
- **All attributes** maintained (widths, styles, IDs)
- **Nested content** kept intact
- **Still validated** before sending to API

### Tradeoffs: ⚖️
- LLM must understand Storage Format XHTML
- More tokens in context
- Risk of LLM generating invalid XHTML

---

<!-- _class: section-header -->

# 4️⃣ Recommendations

---

# Recommendations 🧭

**Creating pages?** Use blocks ✨ → simpler, guaranteed valid
**Editing pages?** Use XHTML ✏️ → preserves everything

| Scenario | Recommended Tool |
|----------|------------------|
| Create new page | `confluence_create_page` (blocks) ✨ |
| Read simple page | `confluence_read_page` (blocks) |
| Read complex page | `confluence_read_page_xhtml` 📄 |
| Edit existing page | `confluence_update_page_xhtml` ✏️ |
| Edit tables | **Always** `confluence_update_page_xhtml` ⚠️ |

---

# Why Not Always XHTML? 🤷

Structured blocks are still valuable:

1. 🛡️ **Safer for creation** - can't produce invalid XHTML
2. 🤖 **Easier for LLMs** - JSON is more natural than XHTML
3. ⚡ **Faster & cheaper** - XHTML uses more tokens, takes longer
4. 📋 **Simpler prompts** - no need to explain Storage Format

Use the right tool for the job 🔧

---

<!-- _class: section-header -->

# 5️⃣ Takeaways

---

# Takeaways 💡

| Challenge | Solution | Lesson |
|-----------|----------|--------|
| LLMs generate wrong format | Structured blocks → valid XHTML ✅ | Work with the format, not against it 🎯 |
| Editing loses formatting | Raw XHTML tools 🔒 | Lossless round-trip is essential 🔄 |
| Tables break on round-trip | Always use XHTML for edits ⚠️ | Multiple tools > one tool 🧰 |
| No pre-flight checks | Validate before API calls ✅ | Catch errors early, not in production |

🔗 **github.com/agentplexus/mcp-confluence**

---

<!-- _class: section-header -->

# Questions? 🙋

---

<!-- _class: section-header -->

# Appendix

---

# Architecture 🏗️

```
┌─────────────────────────────────────────────────────┐
│                    MCP Server                       │
├─────────────────────────────────────────────────────┤
│  Structured Tools          │  XHTML Tools           │
│  ─────────────────         │  ────────────          │
│  read_page (blocks)        │  read_page_xhtml       │
│  update_page (blocks)      │  update_page_xhtml     │
│  create_page (blocks)      │                        │
├─────────────────────────────────────────────────────┤
│              Confluence REST API Client             │
├─────────────────────────────────────────────────────┤
│                 Storage Package                     │
│  Parse (XHTML→IR) │ Render (IR→XHTML) │ Validate    │
└─────────────────────────────────────────────────────┘
```
