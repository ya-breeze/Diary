{{ template "header.tpl" . }}

<div class="container-fluid diary-search-page">
    <header class="diary-page-header">
        <div class="d-flex justify-content-between align-items-center mb-2 mb-md-3">
            <div class="search-header">
                <h1 class="h3 mb-0">Search Results</h1>
                {{ if .searchQuery }}
                    <p class="text-muted mb-0">
                        {{ if .totalCount }}
                            Found {{ .totalCount }} result{{ if ne .totalCount 1 }}s{{ end }} for "{{ .searchQuery }}"
                        {{ else }}
                            No results found for "{{ .searchQuery }}"
                        {{ end }}
                    </p>
                {{ else if .searchTags }}
                    <p class="text-muted mb-0">
                        {{ if .totalCount }}
                            Found {{ .totalCount }} result{{ if ne .totalCount 1 }}s{{ end }} with tags: {{ range $i, $tag := .searchTags }}{{ if $i }}, {{ end }}<span class="badge bg-secondary">{{ $tag }}</span>{{ end }}
                        {{ else }}
                            No results found with the specified tags
                        {{ end }}
                    </p>
                {{ else }}
                    <p class="text-muted mb-0">
                        {{ if .totalCount }}
                            Showing all {{ .totalCount }} entries
                        {{ else }}
                            No diary entries found
                        {{ end }}
                    </p>
                {{ end }}
            </div>

            {{ template "layout-toggle" . }}
        </div>
    </header>

    <div class="diary-content-wrapper">
        <div class="js-disabled-message">
            <strong>Note:</strong> JavaScript is disabled. Layout toggle functionality is not available, but you can still view and navigate your search results.
        </div>

        <main class="diary-main-content layout-narrow" id="mainContent" role="main">
            {{ if .items }}
                <div class="search-results" role="region" aria-label="Search results">
                    {{ range .items }}
                        <article class="diary-entry-card mb-4 position-relative" role="article">
                            <header class="diary-entry-header">
                                <h2 class="diary-entry-title">
                                    <a href="/?date={{ .Date }}" class="text-decoration-none stretched-link">
                                        <time datetime="{{ .Date }}" class="fw-bold">{{ .Date }}</time>
                                        {{ if .Title }}
                                            - {{ .Title }}
                                        {{ end }}
                                    </a>
                                </h2>
                                {{ if .Tags }}
                                    <div class="diary-entry-tags mt-2" aria-label="Tags">
                                        {{ range .Tags }}
                                            <span class="badge bg-secondary me-1">{{ . }}</span>
                                        {{ end }}
                                    </div>
                                {{ end }}
                            </header>
                            
                            <div class="diary-entry-preview mt-3">
                                {{ if .Body }}
                                    <div class="diary-entry-body">{{- if gt (len .Body) 300 -}}{{- snippet .Body 300 -}}... <a href="/?date={{ .Date }}" class="text-primary ms-2" aria-label="Read full entry for {{ .Date }}">Read more...</a>{{- else -}}{{- .Body -}}{{- end -}}</div>
                                {{ else }}
                                    <p class="text-muted fst-italic">No content</p>
                                {{ end }}
                            </div>
                            
                            <footer class="diary-entry-footer mt-3">
                                <div class="d-flex justify-content-between align-items-center">
                                    <a href="/web/edit?date={{ .Date }}" class="btn btn-outline-secondary btn-sm position-relative z-3">
                                        <i class="bi bi-pencil" aria-hidden="true"></i>
                                        Edit
                                    </a>
                                </div>
                            </footer>
                        </article>
                    {{ end }}
                </div>
            {{ else }}
                <div class="diary-empty-state text-center py-5">
                    <i class="bi bi-search display-1 text-muted mb-3" aria-hidden="true"></i>
                    <h2 class="h4 text-muted mb-3">No entries found</h2>
                    {{ if or .searchQuery .searchTags }}
                        <p class="text-muted mb-4">Try adjusting your search terms or browse all entries.</p>
                        <div class="d-flex gap-2 justify-content-center">
                            <a href="/web/search" class="btn btn-outline-primary">
                                <i class="bi bi-arrow-clockwise" aria-hidden="true"></i>
                                Clear Search
                            </a>
                            <a href="/web/" class="btn btn-outline-secondary">
                                <i class="bi bi-house" aria-hidden="true"></i>
                                Go Home
                            </a>
                        </div>
                    {{ else }}
                        <p class="text-muted mb-4">Start writing your first diary entry!</p>
                        <a href="/web/edit" class="btn btn-primary">
                            <i class="bi bi-pencil" aria-hidden="true"></i>
                            Create First Entry
                        </a>
                    {{ end }}
                </div>
            {{ end }}
        </main>
    </div>
</div>

{{ template "footer.tpl" . }}

{{ define "layout-toggle" }}
<div class="layout-toggle-container" role="group" aria-labelledby="layout-toggle-label">
    <span class="layout-toggle-label d-none d-md-inline" id="layout-toggle-label">Layout:</span>

    <button type="button"
            class="layout-toggle-btn"
            id="fullLayoutBtn"
            data-layout="full"
            aria-label="Switch to full width layout - images will display at 100% width"
            aria-pressed="false"
            aria-describedby="layout-status"
            title="Full Width Layout - Images at 100% width"
            tabindex="0">
        <i class="bi bi-arrows-expand" aria-hidden="true"></i>
        <span class="d-none d-lg-inline">Full</span>
    </button>

    <button type="button"
            class="layout-toggle-btn active"
            id="narrowLayoutBtn"
            data-layout="narrow"
            aria-label="Switch to narrow layout - images will display at 30% width"
            aria-pressed="true"
            aria-describedby="layout-status"
            title="Narrow Layout - Images at 30% width"
            tabindex="0">
        <i class="bi bi-arrows-collapse" aria-hidden="true"></i>
        <span class="d-none d-lg-inline">Narrow</span>
    </button>

    <div id="layout-status" class="visually-hidden" aria-live="polite" aria-atomic="true">
        Current layout: Narrow - images display at 30% width
    </div>
</div>
{{ end }}
