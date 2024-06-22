package main

import (
	"errors"
	"io/fs"
	"net"

	"github.com/dzeromsk/nullroute"
)

func main() {
	f, err := nullroute.New()
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ip := net.IPv4(123, 123, 123, 125)
	if err := f.Add(ip); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			panic(err)
		}
		f.Delete(ip)
	}
}
