package streamer

import (
	"fmt"
	"github.com/rs/jplot/graph"
	"log"
	"net"
	"net/http"
)

func Start(port int, dash *graph.Dash, width, height, refreshRate int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}

	stream := &Streamer{
		dash:        dash,
		width:       width,
		height:      height,
		refreshRate: refreshRate,
	}

	log.Println("starting server on ", port)
	er := http.Serve(listener, stream)
	if er != nil {
		return er
	}

	return nil
}
