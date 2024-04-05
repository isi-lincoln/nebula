package main

import (
	"context"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	clientServer string
	clientPort   int
	addr         string
	root         = &cobra.Command{
		Use:   "nebavoid",
		Short: "nebula's avoid controller",
	}
)

func init() {
	flag.IntVar(&clientPort, "port", 12345, "nebula avoid port to use")
	flag.StringVar(&clientServer, "host", "localhost", "nebula avoid host to connect to")
	addr = fmt.Sprintf("%s:%d", clientServer, clientPort)
}

func main() {
	get := &cobra.Command{
		Use:   "get",
		Short: "get nebula host information",
	}
	root.AddCommand(get)

	GetNebulaHostInfo := &cobra.Command{
		Use:   "host",
		Short: "get nebula hostinfo information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			GetNebulaHostInfoFunc()
		},
	}
	root.AddCommand(GetNebulaHostInfo)

	root.Execute()
}

func GetNebulaHostInfoFunc() {
	withAvoid(addr, func(c nebula.AvoidClient) error {
		req := &nebula.GetHostInfoRequest{}
		fmt.Printf("sent getHostInfo\n")
		resp, err := c.GetHostInfo(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)

		return nil
	})
}

func withAvoid(endpoint string, f func(nebula.AvoidClient) error) error {

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to nebula avoid service: %v", err)
	}
	client := nebula.NewAvoidClient(conn)
	defer conn.Close()

	return f(client)
}
