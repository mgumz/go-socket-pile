// go-socket-pile - listen on a socket and spawn a pile of
// client-sockets to itself. Purpose: create a lot of open sockets
// and push the system to some of its limits (open files,
// sockets, etc.). It's also useful to create a cozy environment
// for some benchmarks (e.g., test the fastest way of retrieving
// the number of open sockets)
//
// Copyright (c) 2014, Mathias Gumz <mg@2hoch5.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// The views and conclusions contained in the software and documentation are those
// of the authors and should not be interpreted as representing official policies,
// either expressed or implied, of the go-socket-pile Project.

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var one = []byte("1")

func main() {
	cli := struct {
		net      string
		addr     string
		nworkers int
		duration time.Duration
	}{net: "tcp", addr: ":12345", nworkers: 1024, duration: 1 * time.Minute}

	flag.StringVar(&cli.net, "t", cli.net, "network (tcp, udp, etc)")
	flag.StringVar(&cli.addr, "a", cli.addr, "address to listend and connect to")
	flag.IntVar(&cli.nworkers, "n", cli.nworkers, "amount of workers")
	flag.DurationVar(&cli.duration, "d", cli.duration, "duration")
	flag.Parse()

	os.Stdin.Read(nil)

	server, err := net.Listen(cli.net, cli.addr)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	go func(s net.Listener) {
		for {
			if client, err := s.Accept(); err == nil {
				go func(c net.Conn) {
					c.Read(one)
				}(client)
			}
		}
	}(server)

	wg := sync.WaitGroup{}
	wg.Add(cli.nworkers)
	for i := 0; i < cli.nworkers; i++ {
		go func() {
			dialer := net.Dialer{KeepAlive: cli.duration}
			client, err := dialer.Dial(cli.net, cli.addr)
			if err != nil {
				log.Println(err)
				return
			}
			wg.Done()
			client.Read(one)
		}()
	}

	log.Println("piled up", cli.nworkers, "clients")
	log.Println("wait for", cli.duration.String())

	<-time.After(cli.duration)

	log.Println("done.")
}
