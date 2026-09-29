package extension

import (
	"bytes"
	"io"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// Container directives are block level elements that wrap arbitrary block
// contents:
//
//	::: note optional title
//	Contents are parsed as normal block level elements.
//	:::
//
// A container block directive starts with a line that consists of 3 or more
// ':' characters followed by a container name and an optional title, and ends
// with a line that consists of 3 or more ':' characters only.
// Container block directives can be nested. If a container block directive is
// not closed explicitly, it is closed implicitly at the end of the document.
type containerParser struct {
}

var defaultContainerParser = &containerParser{}

func newContainerParser() parser.BlockParser {
	return defaultContainerParser
}

func (b *containerParser) Trigger() []byte {
	return []byte{':'}
}

// isContainerCloser returns true if the given line consists of 3 or more ':'
// characters followed by spaces only.
func isContainerCloser(line []byte) bool {
	i := 0
	for i < len(line) && line[i] == ':' {
		i++
	}
	return i >= 3 && util.IsBlank(line[i:])
}

// ownsCloser returns true if a closing marker on the current line belongs to
// the given container. A closing marker belongs to the innermost open
// container, and markers inside a fenced code block or an HTML block are
// treated as their contents.
func ownsCloser(node gast.Node, pc parser.Context) bool {
	blocks := pc.OpenedBlocks()
	seen := false
	for _, block := range blocks {
		if !seen {
			seen = block.Node == node
			continue
		}
		switch n := block.Node.(type) {
		case *ast.Container:
			return false
		case *gast.CodeBlock:
			if n.CodeBlockKind == gast.CodeBlockKindFenced {
				return false
			}
		case *gast.HTMLBlock:
			return false
		}
	}
	return true
}

func (b *containerParser) Open(_ gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 || pc.BlockIndent() > 3 || line[pos] != ':' {
		return nil, parser.NoChildren
	}
	i := pos
	for i < len(line) && line[i] == ':' {
		i++
	}
	if i-pos < 3 {
		return nil, parser.NoChildren
	}
	rest := line[i:]
	if len(rest) == 0 || (rest[0] != ' ' && rest[0] != '\t') {
		// A line that consists of ':' characters only is a closing marker.
		return nil, parser.NoChildren
	}
	left := 0
	for left < len(rest) && (rest[left] == ' ' || rest[left] == '\t') {
		left++
	}
	nameStart := i + left
	nameEnd := nameStart
	for nameEnd < len(line) && !util.IsSpace(line[nameEnd]) {
		nameEnd++
	}
	if nameEnd == nameStart {
		return nil, parser.NoChildren
	}
	name := line[nameStart:nameEnd]
	if bytes.IndexByte(name, ':') >= 0 {
		return nil, parser.NoChildren
	}
	titleStart := nameEnd
	for titleStart < len(line) && (line[titleStart] == ' ' || line[titleStart] == '\t') {
		titleStart++
	}
	titleEnd := len(line)
	for titleEnd > titleStart && util.IsSpace(line[titleEnd-1]) {
		titleEnd--
	}
	node := ast.NewContainer(name)
	if titleStart < titleEnd {
		offset := segment.Start - segment.Padding
		title := ast.NewContainerTitle()
		title.AppendSource(text.NewSegment(offset+titleStart, offset+titleEnd))
		title.SetPos(offset + titleStart)
		node.AppendChild(title)
	}
	reader.AdvanceToEOL()
	return node, parser.HasChildren
}

func (b *containerParser) Continue(node gast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if util.IsBlank(line) {
		return parser.Continue | parser.HasChildren
	}
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w < 4 && isContainerCloser(line[pos:]) && ownsCloser(node, pc) {
		reader.AdvanceToEOL()
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

func (b *containerParser) Close(_ gast.Node, _ text.Reader, _ parser.Context) {
	// nothing to do
}

func (b *containerParser) CanInterruptParagraph() bool {
	return true
}

func (b *containerParser) CanAcceptIndentedLine() bool {
	return false
}

type containerHTMLRendererExtension struct {
}

// NewContainerHTMLRenderer returns a new html.Extension for rendering Container nodes.
func NewContainerHTMLRenderer() html.Extension {
	return &containerHTMLRendererExtension{}
}

func (r *containerHTMLRendererExtension) RendererOptions(_ *html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderers(map[gast.NodeKind]html.NodeRenderer{
			ast.KindContainer:      html.NodeRendererFunc(r.renderContainer),
			ast.KindContainerTitle: html.NodeRendererFunc(r.renderContainerTitle),
		}),
	}
}

func (r *containerHTMLRendererExtension) renderContainer(
	writer io.Writer, _ []byte, n gast.Node, entering bool, _ renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		name := n.(*ast.Container).Name
		_, _ = w.WriteString(`<div class="container`)
		if len(name) != 0 {
			_ = w.WriteByte(' ')
			_, _ = w.Write(util.EscapeHTML(name))
		}
		_, _ = w.WriteString("\">\n")
	} else {
		_, _ = w.WriteString("</div>\n")
	}
	return gast.WalkContinue, nil
}

func (r *containerHTMLRendererExtension) renderContainerTitle(
	writer io.Writer, _ []byte, _ gast.Node, entering bool, _ renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		_, _ = w.WriteString(`<p class="container-title">`)
	} else {
		_, _ = w.WriteString("</p>\n")
	}
	return gast.WalkContinue, nil
}

type containerParserExtension struct {
}

// NewContainerParser returns a new parser.Extension for parsing container
// block directives.
func NewContainerParser() parser.Extension {
	return &containerParserExtension{}
}

func (e *containerParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithBlockParsers(
			util.Prioritized(newContainerParser(), 750),
		),
	}
}

// ContainerParser is a default [parser.Extension] for container block directives.
var ContainerParser = NewContainerParser()

// ContainerHTMLRenderer is a default [html.Extension] for container block directives.
var ContainerHTMLRenderer = NewContainerHTMLRenderer()
