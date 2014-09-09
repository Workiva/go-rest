package rest

// IndexTemplate is the mustache template for the documentation index.
const IndexTemplate string = `
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
