# Release Notes - v0.1.0

**Release Date:** December 2024

## Overview

Initial release of mcp-confluence, an MCP server for Confluence with safe handling of Confluence Storage Format (XHTML).

This release addresses a critical problem: AI assistants corrupting Confluence pages when editing them, especially tables and macros.

## Features

### MCP Tools

| Tool | Description |
|------|-------------|
| `confluence_read_page` | Read a page as structured content blocks |
| `confluence_read_page_xhtml` | Read a page as raw Storage Format XHTML |
| `confluence_update_page` | Update a page with structured content blocks |
| `confluence_update_page_xhtml` | Update a page with raw Storage Format XHTML |
| `confluence_create_page` | Create a new page with structured content blocks |
| `confluence_create_table` | Create a table block from structured data |
| `confluence_delete_page` | Delete a page by ID |
| `confluence_search_pages` | Search pages using CQL (Confluence Query Language) |

### Structured Blocks

The following block types are supported for creating and reading pages:

- **Paragraph** - Text paragraphs
- **Heading** - H1-H6 headings
- **Table** - Tables with headers, rows, and optional macros in cells
- **BulletList** - Unordered lists
- **NumberedList** - Ordered lists
- **Macro** - Confluence macros (status, info, code, etc.)
- **CodeBlock** - Code blocks with language
- **HorizontalRule** - Horizontal dividers

### Raw XHTML Support

For editing existing pages with complex content, the XHTML tools provide:

- **Lossless round-trip editing** - No data loss on read/modify/write cycles
- **Full attribute preservation** - Column widths, styles, IDs maintained
- **Nested content support** - Lists, bold, links inside table cells
- **Macro preservation** - `ac:` namespace macros kept intact
- **Validation** - XHTML validated before sending to Confluence API

### Storage Package

The `storage` package provides:

- **Parse** - Convert Storage Format XHTML to structured blocks
- **Render** - Convert structured blocks to valid Storage Format XHTML
- **Validate** - Check XHTML for forbidden tags, missing `<tbody>`, etc.
- **HTML Entity Support** - Full support for HTML entities (&bull;, &rarr;, etc.)

### Confluence Client

The `confluence` package provides a REST API client with:

- Basic authentication support
- Page CRUD operations
- CQL search support
- Both structured (IR) and raw XHTML methods

## Recommended Usage

| Scenario | Recommended Tool |
|----------|------------------|
| Create new page | `confluence_create_page` (blocks) |
| Read simple page | `confluence_read_page` (blocks) |
| Read complex page | `confluence_read_page_xhtml` |
| Edit existing page | `confluence_update_page_xhtml` |
| Edit tables | **Always** `confluence_update_page_xhtml` |

**Key guidance:**
- **Creating pages?** Use blocks - simpler, guaranteed valid output
- **Editing pages?** Use XHTML - preserves everything, no data loss

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CONFLUENCE_BASE_URL` | Your Confluence instance URL |
| `CONFLUENCE_USERNAME` | Your Confluence username (usually email) |
| `CONFLUENCE_API_TOKEN` | API token from Atlassian Account Settings |

### Claude Code Integration

Add to `~/.claude.json` or `.mcp.json`:

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/path/to/mcp-confluence",
      "env": {
        "CONFLUENCE_BASE_URL": "https://example.atlassian.net/wiki",
        "CONFLUENCE_USERNAME": "user@example.com",
        "CONFLUENCE_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

## Known Limitations

- OAuth authentication not yet supported (API token only)
- Attachment handling not yet implemented
- Page labels/metadata not yet supported
- Confluence Data Center/Server not tested (Cloud only)

## Links

- **Repository:** https://github.com/agentplexus/mcp-confluence
- **Documentation:** See README.md
- **Roadmap:** See ROADMAP.md
