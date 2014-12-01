# What is go-rest?

The goal of go-rest is to provide a framework that makes it easy to build a flexible and (mostly) unopinionated REST API with little ceremony. More important, it offers a way to build stable, resource-oriented REST endpoints which are production-ready.

At Workiva, we require standardized REST endpoints across all our services. This makes development more fluent and consumption easier and predictable, not just through syntax but also semantics. Many of our internal web services are written in Python, often running within Google App Engine. Within the past year, however, we've begun to build more services using Go—both on and off App Engine—because it lends itself well to many of the problems we're solving. As a result, we needed a way to build RESTful web services in a fashion similar to our Python ones.

go-rest provides the tooling to build RESTful services rapidly and in a way that meets our internal API specification. Because it takes care of a lot of the boilerplate involved with building these types of services and offers some useful utilities, we're opening go-rest up to the open-source community. The remainder of this post explains go-rest in more detail, how it can be used to build REST APIs, and how we use it at Workiva.
