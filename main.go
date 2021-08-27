package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/netutil"
	"golang.org/x/sync/errgroup"
)

type Url struct {
	Seq []string
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var errG errgroup.Group
	// ctx, cancel := context.WithCancel(context.Background())
	if r.Method == "POST" {

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading", err)
		}
		u := Url{}
		json.Unmarshal(reqBody, &u.Seq)

		if len(u.Seq) <= 20 {
			//check valid urls
			for i := 0; i < len(u.Seq); i++ {
				url := u.Seq[i]
				errG.Go(func() error {
					return func(address string) error {
						resp, err := http.Get(address)
						if err != nil {
							return err
						}
						defer resp.Body.Close()
						return nil
					}(url)
				})
			}
			//if err, stop gorutine
			if err := errG.Wait(); err != nil {
				w.Write([]byte(err.Error()))
				return
			}

			wg := sync.WaitGroup{}
			wg.Add(len(u.Seq)) //set group count

			for j := 0; j < len(u.Seq); j++ {
				go func(url string) {
					customClient := http.Client{
						Timeout: time.Second * 1,
					}
					resp, err := customClient.Get(url)
					if err != nil {
						log.Println(err)
					}
					defer resp.Body.Close()

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err)
					}
					w.Write(body)
					wg.Done()
				}(u.Seq[j])
			}
			wg.Wait()
		}
	}
}

func main() {

	router := http.NewServeMux()
	router.HandleFunc("/", handleRequest)

	srv := http.Server{
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		Handler:      router,
	}

	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatal(err)
	}
	//set maxConn
	listener = netutil.LimitListener(listener, 100)
	defer listener.Close()

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	//register signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//wait for signal
	<-c
	log.Printf("interrupted, shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) //set 5 sec, child proccess end
	defer cancel()
	//server shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v\n", err)
	}
}
