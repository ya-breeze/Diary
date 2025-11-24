package utils

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
)

type imagePrefixRenderer struct {
	html.Renderer
	ImagePrefix string
}

func NewImagePrefixRenderer(imagePrefix string) *imagePrefixRenderer {
	return &imagePrefixRenderer{
		ImagePrefix: imagePrefix,
	}
}

func isVideoExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".webm", ".ogg", ".mov", ".m4v":
		return true
	default:
		return false
	}
}

func videoMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".mov":
		return "video/quicktime"
	default:
		return ""
	}
}

func (r *imagePrefixRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if img, ok := node.(*ast.Image); ok {
		dest := string(img.Destination)
		ext := strings.ToLower(filepath.Ext(dest))
		newSrc := fmt.Sprintf("%s%s", r.ImagePrefix, dest)

		if entering {
			if isVideoExtension(ext) {
				label := string(img.Title)
				if label == "" {
					label = "Embedded video"
				}
				mime := videoMimeType(ext)
				_, _ = fmt.Fprintf(w, `<br><video class="diary-image diary-video" src="%s"`+
					` controls preload="metadata" playsinline aria-label="%s">`, newSrc, template.HTMLEscapeString(label))
				if mime != "" {
					_, _ = fmt.Fprintf(w, `<source src="%s" type="%s">`, newSrc, mime)
				} else {
					_, _ = fmt.Fprintf(w, `<source src="%s">`, newSrc)
				}
			} else {
				_, _ = fmt.Fprintf(w, `<br><a href="%s"><img src="%s" alt="%s" class="diary-image"`, newSrc, newSrc, img.Title)
			}
		} else {
			if isVideoExtension(ext) {
				_, _ = io.WriteString(w, "</video><br>")
			} else {
				_, _ = io.WriteString(w, "/></a><br>")
			}
		}
		return ast.SkipChildren
	}

	// fallback to default
	return r.Renderer.RenderNode(w, node, entering)
}
