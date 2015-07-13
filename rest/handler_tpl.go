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

// handlerTemplate is the mustache template for the handler documentation.
const handlerTemplate = `
<!DOCTYPE HTML>
<html lang="en">
    <head>
        <title>Documentation - {{resource}}</title>
        <meta name="viewport"
            content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
        <link rel="stylesheet"
            href="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css">
        <script src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
        <script
            src="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js"></script>
        <style>
            .field {
                margin-bottom:0px;
                border-top:none;
                border-bottom: 1px solid #ddd;
            }

            .field:hover {
                background-color:#F8F8F8;
            }
        </style>
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
                          <li><a href="{{fileNamePrefix}}_v{{.}}.html">v{{.}}</a></li>
                          {{/versions}}
                        </ul>
                      </li>
                    </ul>
                  </div>
                </div>
            </div>

            <div class="page-header">
                <h1>{{resource}} <span class="label label-primary">v{{version}}</span></h1>
            </div>

            {{#endpoints}}
            <div class="endpoint">
                <h3><span class="label label-{{label}}">{{method}}</span> {{uri}}</h3>
                <p>{{{description}}}</p>
               
                <div class="row">
                    {{#hasInput}}
                    <div class="col-md-6">
                        <h4>Request Payload</h4>
                        <ul id="request-tabs" class="nav nav-tabs" data-tabs="request-tabs">
                            <li class="active">
                                <a href="#request-fields-{{index}}" data-toggle="tab">Fields</a>
                            </li>
                            <li><a href="#request-example-{{index}}" data-toggle="tab">Example</a></li>
                        </ul>
                        <div id="request-tab-content" class="tab-content">
                            <div class="tab-pane active" id="request-fields-{{index}}">
                                <div class="list-group">
                                    {{#inputFields}}
                                    <div class="list-group-item field">
                                        <span style="width:220px;float:left;">
                                            <strong>{{name}}</strong>
                                            <span style="display:block;color:#999;">
                                                {{required}}
                                            </span>
                                        </span>
                                        <p style="margin-left:220px;">
                                            (<em>{{type}}</em>) {{description}}
                                        </p>
                                    </div>
                                    {{/inputFields}}
                                </div>
                            </div>
                            <div class="tab-pane" id="request-example-{{index}}">
                                <div class="list-group">
                                    <div class="dl-horizontal list-group-item"
                                        style="border-top:none;">
                                        <pre>{{exampleRequest}}</pre>
                                    </div>
                                </div>
                            </div>
                        </div> 
                    </div>
                    {{/hasInput}}

                    <div class="col-md-6">
                        <h4>Response Payload</h4>
                        <ul id="response-tabs" class="nav nav-tabs" data-tabs="request-tabs">
                            <li class="active">
                                <a href="#response-fields-{{index}}" data-toggle="tab">Fields</a>
                            </li>
                            <li><a href="#response-example-{{index}}" data-toggle="tab">Example</a></li>
                        </ul>
                        <div id="response-tab-content" class="tab-content">
                            <div class="tab-pane active" id="response-fields-{{index}}">
                                <div class="list-group">
                                    {{#outputFields}}
                                    <div class="list-group-item field">
                                        <span style="width:220px;float:left;">
                                            <strong>{{name}}</strong>
                                        </span>
                                        <p style="margin-left:220px;">
                                            (<em>{{type}}</em>) {{description}}
                                        </p>
                                    </div>
                                    {{/outputFields}}
                                </div> 
                            </div>
                            <div class="tab-pane" id="response-example-{{index}}">
                                <div class="list-group">
                                    <div class="dl-horizontal list-group-item"
                                        style="border-top:none;">
                                        <pre>{{exampleResponse}}</pre>
                                    </div>
                                </div>
                            </div>
                        </div> 
                    </div>

                </div>
                {{/endpoints}}

            </div>
        </div>

        <script type="text/javascript">
            jQuery(document).ready(function ($) {
                $('#request-tabs').tab();
            });
        </script> 

    </body>

</html>
`
