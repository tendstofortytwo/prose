# Prose

Prose is a blogging platform written in Go, which I am building to serve my own blog.

## Setup

As of July 2023, `libsass` is no longer available, so the environment running this binary must have access to `sass` on the PATH.

## Usage

Blog posts should be created in the format `title-slug.md`. Work in progress posts should be stored as `WIP-title-slug.md`. Static content should be stored in the `static/` folder, appropriately arranged.

Posts will be served as `/title-slug`, and files like `static/random/file/structure.txt` will be served as `/random/file/structure.txt`. When title slugs and static files conflict, slugs will have higher precdence. An RSS feed of the blog is available at `/rss.xml`.

To start the server:

	go run ./cmd/prose

Server will be live on port 8080.

The server can be deployed on a willing host using Docker:

	docker build -t prose .
	docker run -p 8080:8080 -it prose

## License

The code in this repository (everything other than the contents of `posts/` and `static/`) is licensed under the MIT license; see LICENSE.md. The blog posts themselves are licensed under [CC-BY-ND 4.0](https://creativecommons.org/licenses/by-nd/4.0/).
