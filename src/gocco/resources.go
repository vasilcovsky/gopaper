package gocco

var HTML = `
<!DOCTYPE html>

<html>
<head>
   <title>{{ .Title }}</title>
  <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  <link rel="stylesheet" media="all" href="http://nikhilm.github.io/gocco/gocco.css" />
  <link rel="stylesheet" href="https://highlightjs.org/static/styles/default.css" />
  <style>
  .hljs {background-color: #f5f5ff;}
  </style>
  <script src="https://highlightjs.org/static/highlight.pack.js"></script>
  <script type="text/javascript">hljs.initHighlightingOnLoad();</script>
</head>
<body>
  <div id="container">
    <div id="background"></div>
    <table cellpadding="0" cellspacing="0">
      <thead>
        <tr>
          <th class="docs">
            <h1>
                {{ .Title }}
            </h1>
          </th>
          <th class="code">
          </th>
        </tr>
      </thead>
      <tbody>
          {{ range .Sections }}
          <tr id="section-{{ .Index }}">
            <td class="docs">
              <div class="pilwrap">
                  <a class="pilcrow" href="#section-{{ .Index }}">&#182;</a>
              </div>
                {{ .DocsHTML }}
            </td>
            <td class="code">
                <pre><code class="golang">{{ .CodeHTML }}</code></pre>
            </td>
          </tr>
          {{ end }}
      </tbody>
    </table>
  </div>

</body>
</html>
`
