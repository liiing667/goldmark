package extension

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func TestContainer(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewContainerParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewContainerHTMLRenderer()),
		),
	)
	testutil.DoTestCaseFile(markdown, "testdata/container.txt", t, testutil.ParseCliCaseArg()...)
}
