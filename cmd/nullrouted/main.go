package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/dzeromsk/nullroute"
)

var (
	sockAddr = flag.String("s", "/var/run/nullrouted.sock", "Path to socket file")
	timeout  = flag.Duration("t", 10*time.Minute, "Route expiration timeout")
)

func main() {
	flag.Parse()

	log.Println("socket:", *sockAddr)
	log.Println("timeout:", *timeout)

	if err := os.RemoveAll(*sockAddr); err != nil {
		log.Fatalln(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "unix", *sockAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()

	if err := os.Chmod(*sockAddr, 0666); err != nil {
		log.Fatalln(err)
	}

	nr, err := nullroute.New()
	if err != nil {
		log.Fatalln(err)
	}
	defer nr.Close()

	go func() {
		defer stop()
		defer ln.Close()
		select {
		case <-ctx.Done():
			if err := cleanupAll(nr); err != nil {
				log.Fatalln(err)
			}
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go serve(conn, nr)
	}
}

func serve(c net.Conn, nr nullroute.NullRoute) {
	defer c.Close()
	s := bufio.NewScanner(c)
	for s.Scan() {
		ip := net.ParseIP(s.Text())
		if ip == nil {
			log.Println("Ignore:", s.Text())
			continue
		}
		if err := nr.Add(ip); err != nil {
			log.Println(err, ip)
			continue
		}
		time.AfterFunc(*timeout, func() {
			if err := nr.Delete(ip); err != nil {
				log.Println(err, ip)
			}
		})
	}
}

var nullRe = regexp.MustCompile(`^\*\s+([A-Z0-9]{8})\s+0{8}\s+0205`)

func cleanupAll(nr nullroute.NullRoute) error {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if m := nullRe.FindSubmatch(s.Bytes()); len(m) == 2 {
			if _, err := hex.Decode(m[1], m[1]); err != nil {
				return err
			}
			ip := net.IPv4(m[1][3], m[1][2], m[1][1], m[1][0])
			if err := nr.Delete(ip); err != nil {
				return err
			}
			fmt.Println("Deleted:", ip)
		}
	}

	return s.Err()
}
