# gnotes-reorg

## Requirements

* Golang 1.15

## About

Re-organises a directory that represents a [GNotes](https://apkpure.com/gnotes-note-notepad-memo/org.dayup.gnotes) export (a now-defunct note-taking app for Android).

Organises all notes as plain text files with the original creation date timestamp included within the filename.

## Required file structure

A standard GNotes export directory looks like this:

```
<root>
    | Other
        | <random_id_per_note>
            | content.html
```

`content.html` includes a HTML body that represents the note's creation and modified timestamps as well as its content. This file is parsed to create the new organisation structure outlined above.

## Running locally

### Clean

```
go run cmd/clean/main.go -i <relative_path_to_gnotes_export_dir> -o <relative_path_to_clean_output_dir>
```
