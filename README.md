# go-rest

The goal of go-rest is to provide a framework that makes it easy to build a flexible and (mostly) unopinionated REST API with little ceremony. It offers tooling for creating stable, resource-oriented endpoints with fine-grained control over input and output fields. The go-rest framework is platform-agnostic, meaning it works both on- and off- App Engine, and pluggable in that it supports custom response serializers, middleware and authentication. It also includes a utility for generating API documentation.

See the examples to get started. Additionally, the `rest` package contains a simple client implementation for consuming go-rest APIs.

## Contributing

Requirements to commit here:
  
  - Branch off master, PR back to master.
  - [gofmt](http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources) your code. Unformatted code will be rejected.
  - Follow the style suggestions found [here](https://code.google.com/p/go-wiki/wiki/CodeReviewComments).
  - Unit test coverage is required.
  - Good docstrs are required for at least exported names, and preferably for all functions.
  - Good [commit messages](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html) are required.
