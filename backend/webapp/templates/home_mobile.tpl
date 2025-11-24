{{ template "header.tpl" . }}

<main>
    {{ with .body }}
        {{ . }}
    {{ end }}
</main>

{{ template "footer.tpl" . }}
