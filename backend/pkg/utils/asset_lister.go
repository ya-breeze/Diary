package utils

import (
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
)

func GetAssetsFromMarkdown(md string) []string {
	lister := &assetLister{}
	_ = markdown.ToHTML([]byte(md), nil, lister)

	return lister.Assets
}

type assetLister struct {
	html.Renderer
	Assets []string
}

func (r *assetLister) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if img, ok := node.(*ast.Image); ok && entering {
		if entering {
			r.Assets = append(r.Assets, string(img.Destination))
		}
		return ast.SkipChildren
	}

	return r.Renderer.RenderNode(w, node, entering)
}
