package streamer

import (
	"errors"
	"fmt"
	"github.com/rs/jplot/graph"
	"net/http"
	"syscall"
)

type Streamer struct {
	dash *graph.Dash

	width, height, refreshRate int
}

func (s *Streamer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.URL.String() == "/" {
		err := s.serveStatic(resp)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
			fmt.Println("error on serving static: ", err)
		}
		return
	}

	resp.Header().Add("Content-Type", "image/png")
	err := s.dash.Render(resp, s.width, s.height)
	if err == nil {
		return
	}

	if errors.Is(err, syscall.EPIPE) {
		return
	}

	fmt.Println("error on render: ", err)
}

func (s *Streamer) serveStatic(resp http.ResponseWriter) error {
	resp.Header().Add("Content-Type", "text/html")
	_, err := resp.Write([]byte(fmt.Sprintf(`
<html>
<head>
</head>
<body>
<img id="data" src="data.png">
</body>

<script>

var i = 0
setInterval( () => {
	i++
	document.getElementById("data").src = "data.png" + "?q=" + i
}, %d)
</script>
</html>
`, s.refreshRate)))

	return err
}
