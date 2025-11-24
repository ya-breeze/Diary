<!DOCTYPE html>
<html lang="en" class="no-js">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ block "title" . }}My Web App{{ end }}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet"
        integrity="sha384-QWTKZyjpPEjISv5WaRU9OFeRpok6YctnYmDr5pNlyT2bRjXh0JMhjY6hW+ALEwIH" crossorigin="anonymous">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css">
    <link rel="stylesheet" href="/web/static/css/layout.css?v={{ .Timestamp }}">
    <!-- Remove no-js class as early as possible -->
    <script>
        document.documentElement.classList.remove('no-js');
        document.documentElement.classList.add('js');
    </script>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/js/bootstrap.bundle.min.js"
        integrity="sha384-YvpcrYf0tY3lHB60NNkmXc5s9fDVZLESaAA55NDzOxhy9GkcIdslK1eN7N6jIeHz"
        crossorigin="anonymous"></script>
    <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.7.1/jquery.min.js"></script>
    <script src="/web/static/js/layout-toggle.js"></script>

    <style>
        /* Default image styles - will be overridden by layout classes */
        img {
            width: 50%;
            max-width: 50%;
        }
    </style>
</head>

<body>
    <header>
        <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
            <div class="container-fluid">
                <a class="navbar-brand" href="/">Diary</a>
                <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
                    <span class="navbar-toggler-icon"></span>
                </button>
                <div class="collapse navbar-collapse" id="navbarNav">
                    <ul class="navbar-nav">
                        <li class="nav-item">
                            <a class="nav-link {{if eq .CurrentPage "home"}}active{{end}}" href="/">Home</a>
                        </li>
                        <li class="nav-item">
                            <a class="nav-link {{if eq .CurrentPage "edit"}}active{{end}}" href="/web/edit{{ if .item.Date }}?date={{ .item.Date }}{{ end }}">Edit</a>
                        </li>

                        <li class="nav-item">
                            <a class="nav-link {{if eq .CurrentPage "about"}}active{{end}}" href="/web/about">About</a>
                        </li>
                    </ul>
                    <form class="d-flex ms-auto" action="/web/search" method="GET" role="search">
                        <input class="form-control me-2"
                               type="search"
                               name="search"
                               placeholder="Search diary entries..."
                               aria-label="Search diary entries"
                               value="{{ .searchQuery }}"
                               autocomplete="off">
                        <button class="btn btn-outline-success" type="submit" aria-label="Submit search">
                            <i class="bi bi-search" aria-hidden="true"></i>
                            <span class="d-none d-md-inline">Search</span>
                        </button>
                    </form>
                    <ul class="navbar-nav ms-3">
                        <li class="nav-item">
                            <a class="nav-link" href="/web/logout">Logout</a>
                        </li>
                    </ul>
                </div>
            </div>
        </nav>
