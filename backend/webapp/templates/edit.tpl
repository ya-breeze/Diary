{{ template "header.tpl" . }}

<script>
$(document).ready(function () {
    {{ range .assets }}
        addImage("{{ . }}");
    {{ end }}

    // Event delegation for all dynamically added images
    $(document).on('click', 'img[id^="dynamicImg_"]', function () {
        // Delegate to extracted helper that performs the insertion
        insertAssetMarkdownFromId($(this).attr('id'));
    });

    function insertAssetMarkdownFromId(clickedId) {
        const assetName = clickedId.replace('dynamicImg_', '');
        $('#body').val(function(i, val) {
            return val + '\n![](' + assetName + ')\n';
        });
        $('#body').focus();
    }

    function addImage(name) {
        var src = $(location).attr('origin') + '/web/assets/' + name;
        const imgId = 'dynamicImg_' + name;
        $('#assets').append('<div class="card" style="width: 18rem;"><img src="' + src + '" id="' + imgId + '" class="card-img-top"></div>');
    }

    // Upload immediately when files are selected (supports single or multiple)
    $('#imageUpload').on('change', function () {
        $('#uploadError').addClass('d-none').text('');
        var fileInput = this;
        if (fileInput.files.length === 0) { return; }

        var formData = new FormData();
        if (fileInput.files.length === 1) {
            formData.append('asset', fileInput.files[0]);
            $.ajax({
                url: 'upload',
                type: 'POST',
                data: formData,
                contentType: false,
                processData: false,
                xhr: function() {
                    var xhr = $.ajaxSettings.xhr();
                    if (xhr.upload) {
                        xhr.upload.addEventListener('progress', function(e) {
                            if (e.lengthComputable) {
                                var pct = Math.round((e.loaded / e.total) * 100);
                                $('#uploadProgress').removeClass('d-none');
                                $('#uploadProgress .progress-bar').css('width', pct + '%').attr('aria-valuenow', pct).text(pct + '%');
                            }
                        }, false);
                    }
                    return xhr;
                },
                success: function (response) {
                    addImage(response);
                    insertAssetMarkdownFromId('dynamicImg_' + response);
                    $('#imageUpload').val('');
                    $('#uploadProgress').addClass('d-none');
                },
                error: function (xhr) {
                    var msg = 'Upload failed.';
                    if (xhr.responseJSON && xhr.responseJSON.error) { msg = xhr.responseJSON.error; }
                    $('#uploadError').text(msg).removeClass('d-none');
                    $('#uploadProgress').addClass('d-none');
                }
            });
        } else {
            for (var i = 0; i < fileInput.files.length; i++) {
                formData.append('assets', fileInput.files[i]);
            }
            $.ajax({
                url: 'upload-batch',
                type: 'POST',
                data: formData,
                contentType: false,
                processData: false,
                xhr: function() {
                    var xhr = $.ajaxSettings.xhr();
                    if (xhr.upload) {
                        xhr.upload.addEventListener('progress', function(e) {
                            if (e.lengthComputable) {
                                var pct = Math.round((e.loaded / e.total) * 100);
                                $('#uploadProgress').removeClass('d-none');
                                $('#uploadProgress .progress-bar').css('width', pct + '%').attr('aria-valuenow', pct).text(pct + '%');
                            }
                        }, false);
                    }
                    return xhr;
                },
                success: function (response) {
                    try {
                        var data = typeof response === 'string' ? JSON.parse(response) : response;
                        (data.files || []).forEach(function(file){
                            var name = file.savedName || file;
                            addImage(name);
                            insertAssetMarkdownFromId('dynamicImg_' + name);
                        });
                    } catch (e) { console.error('Invalid JSON', e); }
                    $('#imageUpload').val('');
                    $('#uploadProgress').addClass('d-none');
                },
                error: function (xhr) {
                    var msg = 'Batch upload failed.';
                    if (xhr.responseJSON && xhr.responseJSON.error) { msg = xhr.responseJSON.error; }
                    $('#uploadError').text(msg).removeClass('d-none');
                    $('#uploadProgress').addClass('d-none');
                }
            });
        }
    });

    // Keep upload button as a shortcut to open file picker
    $('#uploadBtn').on('click', function () {
        $('#imageUpload').click();
    });
});
</script>


<main>
    {{ with .Error }}
    <div class="alert alert-danger" role="alert">
        {{ . }}
    </div>
    {{ end }}

    <div class="row">
        <div class="col">
            <form action="/web/edit" method="POST">
                <input type="hidden" name="date" value="{{ .item.Date }}"/>
                <input type="hidden" id="user_id" value="{{ .UserID }}"/>

                <h5>{{ .item.Date }}</h5>

                <div class="mb-3">
                    <label for="title" class="form-label">Title:</label>
                    <input type="text" class="form-control" name="title" value="{{ .item.Title }}"/>
                </div>

                <div class="mb-3">
                    <label for="body" class="form-label">Body:</label>
                    <textarea class="form-control" name="body" id="body" rows="10">{{ .item.Body }}</textarea>
                </div>

                <div class="mb-3">
                    <label for="tags" class="form-label">Tags</label>
                    <input type="text" class="form-control" name="tags" value="{{ range .item.Tags }}{{.}}, {{ end }}"/>
                </div>

                <button type="submit" class="btn btn-primary">Save</button>
            </form>
        </div>
        <div class="col-3 text-center">`
            <input type="file" id="imageUpload" hidden multiple />
            <button id="uploadBtn" class="btn btn-secondary">Upload asset(s)</button>

            <div id="uploadError" class="alert alert-danger d-none mt-2" role="alert"></div>
            <div id="uploadProgress" class="progress d-none mt-2">
              <div class="progress-bar" role="progressbar" style="width: 0%;" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100">0%</div>
            </div>

            <h5>Assets</h5>
            <div id="assets">
            </div>
        </div>
    </div>
</main>

{{ template "footer.tpl" . }}