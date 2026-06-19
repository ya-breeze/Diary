# Feature: Assets (File Uploads)

## Purpose
How files (images and videos) are uploaded, validated, stored, served, and managed as attachments on diary entries.

## Requirements

### Requirement: Upload files while editing an entry
Files SHALL be addable to an entry via drag-and-drop or the file browser.

#### Scenario: Upload via drag-and-drop
- **GIVEN** the user is in the editor
- **WHEN** they drag one or more files onto the drop zone
- **THEN** the files are uploaded via `POST /v1/assets/batch`
- **AND** a progress bar shows upload progress (0–100%)
- **AND** on success, a markdown image tag `![](savedName)` for each file is appended to the body textarea
- **AND** the images appear in the image grid below the drop zone

#### Scenario: Upload via file browser
- **GIVEN** the user clicks "browse" in the drop zone
- **WHEN** they select one or more files
- **THEN** the same upload flow as drag-and-drop is triggered

#### Scenario: Upload accepts images and videos
- **THEN** the file input accepts `image/*` and `video/*` MIME types
- **AND** only files with allowed extensions are accepted by the server

#### Scenario: Multiple files uploaded at once
- **WHEN** the user selects 3 files in a single upload
- **THEN** all 3 are uploaded in a single batch request
- **AND** all 3 markdown image tags are appended to the body

#### Scenario: Progress bar shown during upload
- **WHEN** an upload is in progress
- **THEN** the drop zone is disabled (pointer-events: none, opacity: 50%)
- **AND** a progress indicator shows "Uploading... N%" 
- **AND** the progress bar disappears when the upload completes or fails

#### Scenario: Upload error
- **WHEN** an upload fails (network error or server error)
- **THEN** the progress bar disappears
- **AND** no image tags are added to the body
- **AND** the error is logged to the console (not shown in the UI)

### Requirement: Server-side upload validation
The server SHALL enforce limits and extension rules before saving files.

#### Scenario: No files in batch
- **WHEN** a batch upload request contains no files
- **THEN** the server returns 400

#### Scenario: Too many files in a single batch
- **WHEN** a batch upload exceeds the configured max file count
- **THEN** the server returns 413 (Request Entity Too Large)

#### Scenario: Total batch size exceeded
- **WHEN** the combined size of all files exceeds the configured batch limit
- **THEN** the server returns 413

#### Scenario: Individual file too large
- **WHEN** a single file within the batch exceeds the per-file size limit
- **THEN** the server returns 413
- **AND** any files already saved in that batch are rolled back (deleted from disk)

#### Scenario: Invalid file extension
- **WHEN** a file has an extension not in the allowed list
- **THEN** the server returns 400
- **AND** no files from the batch are saved

#### Scenario: Files are isolated per family
- **GIVEN** two families each upload a file with the same name
- **THEN** each file is stored in the family's own directory and is not accessible to the other family

### Requirement: Manage attached images in the editor
The editor SHALL show attached images and allow their removal.

#### Scenario: Images extracted from existing entry body on editor open
- **GIVEN** an entry body contains `![alt](path/to/image.jpg)` and `![](path/to/video.mp4)`
- **WHEN** the editor opens for that entry
- **THEN** both files appear in the image grid

#### Scenario: Remove an attached image
- **GIVEN** the editor has an image grid with 2 images
- **WHEN** the user clicks remove on the first image
- **THEN** the image is removed from the grid
- **AND** the corresponding `![...](...) ` markdown reference is removed from the body textarea

### Requirement: Serve assets
Assets SHALL be served at `GET /v1/assets/{path}`, scoped to the requesting family's directory.

#### Scenario: Valid asset path returns file
- **GIVEN** the family has a file at `image.jpg` in their asset directory
- **WHEN** `GET /v1/assets/image.jpg` is called
- **THEN** the file contents are returned with 200

#### Scenario: Path traversal is rejected
- **WHEN** the path contains `..` (e.g. `../other-family/secret.jpg`)
- **THEN** the server returns 400

#### Scenario: Absolute path is rejected
- **WHEN** the path is absolute (starts with `/`)
- **THEN** the server returns 400

#### Scenario: Path resolves outside family directory
- **WHEN** the cleaned path resolves to a location outside the family's asset directory
- **THEN** the server returns 400

#### Scenario: Asset not found
- **WHEN** the requested file does not exist
- **THEN** the server returns 404

#### Scenario: Path is a directory
- **WHEN** the path resolves to a directory rather than a file
- **THEN** the server returns 400
