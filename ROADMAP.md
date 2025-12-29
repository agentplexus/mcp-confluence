# Roadmap

This document outlines planned features and improvements for go-confluence.

## Current Status (v0.1.0)

The initial release provides:

- **storage package**: IR types, render, parse, validate for Confluence Storage Format
- **confluence package**: REST API client with IR integration
- **mcpserver package**: MCP server implementation with structured tools
- **cmd/mcp-server**: Executable MCP server binary

## Short-term (v0.2.0)

### Additional Block Types

- [ ] Link blocks (`<a href="...">` and `<ri:page>`)
- [ ] Image blocks (`<ac:image>`)
- [ ] Attachment references
- [ ] Emoji support
- [ ] Panel blocks (info, note, warning, tip)
- [ ] Expand/collapse blocks

### Enhanced Table Support

- [ ] Column widths
- [ ] Row/column span (`rowspan`, `colspan`)
- [ ] Nested content in cells (lists, macros)
- [ ] Table styles/colors

### Improved Parsing

- [ ] Handle more edge cases in XHTML parsing
- [ ] Preserve unknown elements during round-trip
- [ ] Better error messages with line numbers

### API Coverage

- [ ] Labels API (add/remove/list labels)
- [ ] Attachments API (upload/download/list)
- [ ] Comments API
- [ ] Page history/versions
- [ ] Page permissions

## Medium-term (v0.3.0)

### Macro Support

- [ ] Comprehensive macro allowlist
- [ ] Macro-specific IR types for common macros:
  - [ ] `code` macro with syntax highlighting options
  - [ ] `toc` (table of contents)
  - [ ] `children` macro
  - [ ] `excerpt` macro
  - [ ] `include` macro
  - [ ] `jira` macro
- [ ] Custom macro registration

### Template System

- [ ] Page templates
- [ ] Template variables
- [ ] Template inheritance
- [ ] Common template library (meeting notes, decision records, etc.)

### Validation Enhancements

- [ ] Schema-based validation
- [ ] Custom validation rules
- [ ] Validation profiles (strict, permissive)
- [ ] Pre-flight checks before API calls

### MCP Server Improvements

- [ ] HTTP transport (in addition to stdio)
- [ ] WebSocket transport
- [ ] Authentication options (OAuth, PAT)
- [ ] Rate limiting
- [ ] Caching layer
- [ ] Batch operations

## Long-term (v1.0.0)

### Content Intelligence

- [ ] Diff/patch operations on IR
- [ ] Merge conflict detection
- [ ] Content migration tools (Markdown → Storage XHTML)

### Multi-format Support

- [ ] Export to Markdown
- [ ] Export to HTML
- [ ] Export to PDF (via rendering)
- [ ] Import from Markdown
- [ ] Import from HTML

### Enterprise Features

- [ ] Multi-instance support
- [ ] Bulk operations
- [ ] Audit logging
- [ ] Metrics/observability

### Testing & Quality

- [ ] Golden file test suite with real Confluence XHTML samples
- [ ] Fuzzing for parser robustness
- [ ] Integration tests with Confluence Cloud
- [ ] Performance benchmarks

### Documentation

- [ ] godoc improvements
- [ ] Usage examples
- [ ] Best practices guide
- [ ] MCP integration guide for Claude, GPT, etc.

## Future Considerations

### Atlas Doc Format (ADF)

- [ ] Investigate ADF support as alternative to Storage XHTML
- [ ] IR → ADF renderer
- [ ] ADF → IR parser
- [ ] Determine which format to use when

### Confluence Data Center

- [ ] Test compatibility with Confluence Data Center (on-prem)
- [ ] Handle API differences between Cloud and Data Center

### Related Atlassian Products

- [ ] Jira integration (linked issues, issue panels)
- [ ] Confluence Whiteboards
- [ ] Confluence Databases

---

## Contributing

Contributions are welcome! Please open an issue to discuss proposed features before starting work.

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes to public APIs
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible
