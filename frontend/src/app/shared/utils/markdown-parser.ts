/**
 * Extracts asset references from markdown text
 * Looks for markdown image syntax: ![alt text](filename)
 * Returns array of unique filenames (not full URLs)
 */
export function extractAssetsFromMarkdown(markdown: string): string[] {
  if (!markdown) {
    return [];
  }

  const assets: string[] = [];
  // Pattern: ![alt text](filename) or ![](filename)
  const imagePattern = /!\[([^\]]*)\]\(([^)]+)\)/g;
  let match;

  while ((match = imagePattern.exec(markdown)) !== null) {
    const path = match[2];
    // Only include if it's not already a full URL
    if (!path.startsWith("http://") && !path.startsWith("https://")) {
      assets.push(path);
    }
  }

  // Return unique assets
  return Array.from(new Set(assets));
}

/**
 * Appends asset references to markdown text
 * Adds markdown image syntax at the end of the text
 */
export function appendAssetsToMarkdown(
  markdown: string,
  assetPaths: string[]
): string {
  if (!assetPaths || assetPaths.length === 0) {
    return markdown;
  }

  // Ensure there's a newline before adding assets
  let result = markdown;
  if (result && !result.endsWith("\n")) {
    result += "\n";
  }

  // Add each asset as a markdown image reference
  for (const path of assetPaths) {
    result += `\n![](${path})`;
  }

  return result;
}

