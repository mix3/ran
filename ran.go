package ran

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/braintree/manners"
	"github.com/lestrrat/go-server-starter/listener"
)

type server struct {
	m    *manners.GracefulServer
	sigs []os.Signal
}

func New(svr *http.Server, sigs []os.Signal) *server {
	return &server{
		m:    manners.NewWithServer(svr),
		sigs: sigs,
	}
}

func Run(addr string, n http.Handler) error {
	svr := &http.Server{
		Addr:    addr,
		Handler: n,
	}
	if err := ListenAndServe(svr); err != nil {
		return err
	}
}

func ListenAndServe(s *http.Server) error {
	svr := New(s, []os.Signal{syscall.SIGTERM})
	return svr.ListenAndServe()
}

func (srv *server) newListener() (net.Listener, error) {
	listeners, err := listener.ListenAll()
	if err != nil && err != listener.ErrNoListeningTarget {
		return nil, err
	}

	var l net.Listener
	if len(listeners) == 0 {
		addr := srv.m.Addr
		if addr == "" {
			addr = ":http"
		}

		l, err = net.Listen("tcp", addr)
	} else {
		l = listeners[0]
	}

	return l, err
}

func (srv *server) ListenAndServe() error {
	l, err := srv.newListener()
	if err != nil {
		return err
	}

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, srv.sigs...)
		<-sigchan
		log.Println("Shutting down...")
		manners.Close()
	}()

	log.Println("listen and server on " + srv.m.Addr)
	return srv.m.Serve(l)
}
