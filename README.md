# gnotes-reorg

## Requirements

* Golang 1.15

## About

Re-organises a directory that represents a [GNotes](https://apkpure.com/gnotes-note-notepad-memo/org.dayup.gnotes) export (a now-defunct note-taking app for Android).

Parses and sanitises raw note content, then writes to either plain text files (named by timestamp and title) or JSON files comprising additional metadata.

Notes written to JSON files can be optionally processed further by separating into custom categories and then storing as plain text files within this category hierarchy.

## Required file structure

A standard GNotes export directory looks like this:

```
<root>
    | Other
        | <random_id_per_note>
            | content.html
```

`content.html` includes a HTML body that represents the note's creation and modified timestamps as well as its content. This file is parsed to retrieve this data items as metadata for each note.

## Running locally

### Clean

The first stage is to clean the raw source files.

Either, output each note as a sanitised plain text file...

```
go run cmd/clean/main.go -txt -i <relative_path_to_gnotes_export_dir> -o ./cleaned
```

...or, output each note as a JSON payload including metadata for further processing:

```
go run cmd/clean/main.go -json -i <relative_path_to_gnotes_export_dir> -o ./cleaned
```

### Categorise

The second stage is to specify custom categories for each note and generate a manifest of categories (only applicable if previous stage has output JSON files).

```
go run cmd/categorise/main.go -i ./cleaned
```

This script will show a preview of each note in turn and prompt for a custom category to assign to the note.

The manifest will be saved as `./cleaned/manifest.json`

### Store

The third stage is to store each note as a plain text file in the hierarchy represented by the category manifest.

Either specify the filesystem destination:

```
# cleaned JSON note files and manifest must be in same directory
go run cmd/store/main.go -f -i ./cleaned
```

This command will copy the notes to `./cleaned/categorised/<category>/<note_timestamp_and_title>.txt`

Or specify the Google storage destination:

```
# cleaned JSON note files and manifest must be in same directory
go run cmd/store/main.go -g -i ./cleaned
```
