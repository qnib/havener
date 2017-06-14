package havener


import (
	"net/http"
	"github.com/zpatrick/go-config"
	"fmt"
	"strings"
)


type SrvIndex struct {
	Cfg *config.Config
	RegQuery chan interface{}
}

func NewSrvIndex(cfg *config.Config, rq chan interface{}) SrvIndex {
	si := SrvIndex{
		Cfg: cfg,
		RegQuery: rq,
	}
	return si
}



func (si *SrvIndex) Handler(w http.ResponseWriter, r *http.Request) {
	req := NewRegistryRequest()
	si.RegQuery <- req
	val := <- si.RegQuery
	reg := val.(Registry)
	_ = reg
	fmt.Fprintf(w,`<body><html>
  <head>
    <title>Service Overview v%s</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/css/bootstrap.min.css">
    <script src="/static/js/jquery.min.js"></script>
    <script src="/static/js/bootstrap.min.js"></script>
  </head>
  `, Version)
	fmt.Fprintf(w, `<div class="container">
  <h2>SWARM Service Overview v%s</h2>
  <p>Services according to the SWARM service API</p>
  <table class="table table-striped">
    <thead>
      <tr>
        <th>Stack/Service</th>
        <th>Redirects</th>
      </tr>
    </thead>
    <tbody>`, Version)
	for k, lst := range reg {
		fmt.Fprintf(w, " <tr><td>%s</td><td>", k)
		links := []string{}
		for _, v := range lst {
			links = append(links, fmt.Sprintf("<a href='%s'>%s</a>", v, v))
		}
		fmt.Fprintf(w, strings.Join(links, ", "))
		fmt.Fprintln(w, "</td></tr>")
	}
	fmt.Fprintln(w, `</tbody>
  </table>
</div>
</body>
</html>`)
}