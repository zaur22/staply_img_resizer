package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"staply_img_resize/resizer"
	router "staply_img_resize/router"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type Server struct {
	s          *http.Server
	stopped    chan bool
	osStopSigs chan os.Signal
	workerStop func()
}

func init() {
	viper.SetDefault("server_addr", "localhost:3000")
}

func NewServer() *Server {

	reszr := resizer.NewImgResizer()
	router := router.NewRouter(reszr)

	var server = Server{
		s: &http.Server{
			Addr:           viper.GetString("server_addr"),
			Handler:        &router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	server.workerStop = reszr.Stop
	server.osStopSigs = make(chan os.Signal, 1)
	server.stopped = make(chan bool, 1)
	signal.Notify(server.osStopSigs, os.Interrupt, syscall.SIGTERM)

	return &server
}

func (s *Server) Serve() {

	go func() {
		log.Printf("starting server at %s", s.s.Addr)
		if err := s.s.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err.Error())
		}
	}()

	go func(stop func()) {
		<-s.osStopSigs
		fmt.Println()
		fmt.Println("Stopping program...")
		stop()
		s.stopped <- true
	}(s.workerStop)

	<-s.stopped
	print("Done\n")
}
