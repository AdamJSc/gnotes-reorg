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

`content.html` comprises the note's creation timestamp as well as its content. This file is parsed to create the new organisation structure outlined above.

## Running locally

```
go run main.go -i <relative_export_path> -o <relative_path_to_write_to>
```
