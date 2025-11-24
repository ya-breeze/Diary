{{ template "header.tpl" . }}

<div class="container-fluid diary-home-page">
    <header class="diary-page-header">
        <div class="d-flex justify-content-between align-items-center mb-2 mb-md-3">
            {{ template "date-navigation" . }}

            {{ template "layout-toggle" . }}
        </div>
    </header>

    <div class="diary-content-wrapper">
        <div class="js-disabled-message">
            <strong>Note:</strong> JavaScript is disabled. Layout toggle functionality is not available, but you can still view and navigate your diary entries.
        </div>

        <main class="diary-main-content layout-narrow" id="mainContent" role="main">
            {{ with .body }}
                <article class="diary-entry-content">
                    {{ . }}
                </article>
            {{ else }}
                <div class="diary-empty-state">
                    <p class="text-muted">No content for this date.</p>
                    <a href="/web/edit{{ if .item.Date }}?date={{ .item.Date }}{{ end }}" class="btn btn-outline-primary">
                        <i class="bi bi-pencil" aria-hidden="true"></i>
                        Create Entry
                    </a>
                </div>
            {{ end }}
        </main>
    </div>
</div>

{{ template "footer.tpl" . }}

{{ define "date-navigation" }}
<nav class="date-navigation" aria-label="Date navigation">
    {{ with .previousDate }}
        <a href="{{ addQueryParam $.CurrentURL "date" . }}"
           class="btn btn-primary me-2"
           aria-label="Go to previous date: {{ . }}"
           title="Previous date">
            <i class="bi-arrow-left-circle-fill" aria-hidden="true"></i>
            <span class="visually-hidden">Previous</span>
        </a>
    {{ end }}

    <time class="fw-bold current-date" datetime="{{ .item.Date }}">
        {{ .item.Date }}
        {{ if .item.Title }}
            -
            {{ .item.Title }}
        {{ end }}
    </time>

    {{ with .nextDate }}
        <a href="{{ addQueryParam $.CurrentURL "date" . }}"
           class="btn btn-primary ms-2"
           aria-label="Go to next date: {{ . }}"
           title="Next date">
            <i class="bi-arrow-right-circle-fill" aria-hidden="true"></i>
            <span class="visually-hidden">Next</span>
        </a>
    {{ end }}
</nav>
{{ end }}

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