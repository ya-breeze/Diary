package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
)

// imageRefRe matches a markdown image reference: ![alt](destination).
var imageRefRe = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)

// ComputeTagsSourceHash returns a stable hash of the tag-relevant content of an
// entry: its title, its body text with image references removed, and the sorted
// set of asset filenames the body references.
//
// Removing image references from the text component (and re-adding the assets as
// a sorted set) makes the hash invariant to reordering image references in the
// body, while still changing when text, title, or the set of referenced assets
// changes. This drives edit-triggered retagging and the backfill health check.
func ComputeTagsSourceHash(title, body string) string {
	text := imageRefRe.ReplaceAllString(body, "")

	assets := GetAssetsFromMarkdown(body)
	sorted := make([]string, len(assets))
	copy(sorted, assets)
	sort.Strings(sorted)

	h := sha256.New()
	h.Write([]byte(title))
	h.Write([]byte("\n"))
	h.Write([]byte(text))
	h.Write([]byte("\n"))
	h.Write([]byte(strings.Join(sorted, ",")))

	return hex.EncodeToString(h.Sum(nil))
}
