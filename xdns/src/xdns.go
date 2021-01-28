package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"golang.org/x/net/context"
)

func init() {
	flag.StringVar(&dnsListenAddr, "l", "", "dns proxy listen address, e.g. '127.0.0.1:53'")
	flag.StringVar(&dnsForwardVpn, "f2", "8.8.8.8,8.8.4.4", "dns forward address through vpn, e.g. '8.8.8.8'")
	flag.StringVar(&dnsForward, "f", "192.168.1.1", "dns forward address, e.g. '192.168.1.1'")
	flag.StringVar(&vpnIfname, "vpn", "l2tp-vpn", "vpn ifname, e.g. 'l2tp-vpn'")
	flag.Parse()
}

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		for {
			_, err := net.InterfaceByName(vpnIfname)
			f := false
			if err != nil {
				log.Printf("could not get interface: %s", err)
			} else {
				f = true
			}

			if f != useVpn {
				dnsSrv.book.clear()

				if !f {
					ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
					cmd := exec.CommandContext(ctx, "/etc/init.d/network", "restart")
					cmd.Run()
					cf()
				}
			}

			useVpn = f

			<-ticker.C
		}
	}()

	dnsSrv = &DNSService{}
	dnsSrv.Listen()
	fmt.Printf("end\n")
}
