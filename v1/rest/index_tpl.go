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

package rest

// indexTemplate is the mustache template for the documentation index.
const indexTemplate = `
<!DOCTYPE HTML>
<html lang="en">
    <head>
        <title>REST API v{{version}} Documentation</title>
        <meta name="viewport"
            content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
        <link rel="stylesheet"
            href="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css">
        <script src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
        <script
            src="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js"></script>
    </head>

    <body>
        <div class="container">

            <div class="navbar navbar-default" role="navigation" style="margin-top:30px;">
                <div class="container-fluid">
                  <div class="navbar-header">
                    <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target=".navbar-collapse">
                      <span class="sr-only">Toggle navigation</span>
                      <span class="icon-bar"></span>
                      <span class="icon-bar"></span>
                      <span class="icon-bar"></span>
                    </button>
                    <a class="navbar-brand" href="index_v{{version}}.html">REST API Documentation</a>
                  </div>
                  <div class="navbar-collapse collapse">
                    <ul class="nav navbar-nav navbar-right">
                      <li class="dropdown">
                        <a href="#" class="dropdown-toggle" data-toggle="dropdown">
                            v{{version}} <span class="caret"></span>
                        </a>
                        <ul class="dropdown-menu" role="menu">
                          {{#versions}}
                              <li><a href="index_v{{.}}.html">v{{.}}</a></li>
                          {{/versions}}
                        </ul>
                      </li>
                    </ul>
                  </div>
                </div>
            </div>

            <ul>
                {{#handlers}}
                    <li><a href="{{file}}">{{name}}</a></li>
                {{/handlers}}
            </ul>
        </div>
    </body>

</html>
`
