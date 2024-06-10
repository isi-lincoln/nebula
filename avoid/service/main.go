package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/slackhq/nebula/avoid/service/pkg"
	"google.golang.org/grpc"
)

var (
	Build string
)

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	tunnelPort := flag.Int("tunport", 55554, "port to configure tunnel server")
	tunnelServer := flag.String("tunserver", "0.0.0.0", "tunnel server address or interface")

	debug := flag.Bool("debug", false, "enable extra debugging")

	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	// daemon mode
	if *debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.Infof("starting avoid tunnel api: %s:%d", *tunnelServer, *tunnelPort)

	tunAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *tunnelServer, *tunnelPort))
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}

	grpcTunnelServer := grpc.NewServer()
	avoid.RegisterTunnelServer(grpcTunnelServer, &pkg.TunnelServer{Updates: make(map[string]*pkg.TunnelUpdate)})
	grpcTunnelServer.Serve(tunAddr)

	os.Exit(0)
}
