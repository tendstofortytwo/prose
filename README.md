# Prose

Prose is a blogging platform written in Go, which I am building to serve my own blog.

## Usage

Blog posts should be created in the format `title-slug.md`. Work in progress posts should be stored as `WIP-title-slug.md`. Static content should be stored in the `static/` folder, appropriately arranged.

Posts will be served as `/title-slug`, and files like `static/random/file/structure.txt` will be served as `/random/file/structure.txt`. When title slugs and static files conflict, slugs will have higher precdence.

To start the server:

	go run .

Server will be live on port 8080.