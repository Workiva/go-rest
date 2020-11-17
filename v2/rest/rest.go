/*
Copyright 2014 - 2015 Workiva, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package rest provides a framework that makes it easy to build a flexible and
(mostly) unopinionated REST API with little ceremony. It offers tooling for
creating stable, resource-oriented endpoints with fine-grained control over
input and output fields. The go-rest framework is platform-agnostic, meaning it
works both on- and off- App Engine, and pluggable in that it supports custom
response serializers, middleware and authentication. It also includes a utility
for generating API documentation.
*/
package rest
