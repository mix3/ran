package ran

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/braintree/manners"
	"github.com/lestrrat/go-server-starter/listener"
)

type Server struct {
	*manners.GracefulServer
}

func New(svr *http.Server) *Server {
	return &Server{manners.NewWithServer(svr)}
}

func Run(addr string, n http.Handler) {
	srv := New(&http.Server{Addr: addr, Handler: n})

	if err := srv.ListenAndServe(); err != nil {
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			logger := log.New(os.Stdout, "[graceful] ", 0)
			logger.Fatal(err)
		}
	}
}

func ListenAndServe(s *http.Server) error {
	svr := New(s)
	return svr.ListenAndServe()
}

func (srv *Server) newListener() (net.Listener, error) {
	listeners, err := listener.ListenAll()
	if err != nil && err != listener.ErrNoListeningTarget {
		return nil, err
	}

	var l net.Listener
	if len(listeners) == 0 {
		addr := srv.Addr
		if addr == "" {
			addr = ":http"
		}

		l, err = net.Listen("tcp", addr)
	} else {
		l = listeners[0]
	}

	return l, err
}

func (srv *Server) ListenAndServe() error {
	l, err := srv.newListener()
	if err != nil {
		return err
	}

	return srv.Serve(l)
}
