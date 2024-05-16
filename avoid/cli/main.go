package main

import (
	"context"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	server string
	port   int
	addr   string
	root   = &cobra.Command{
		Use:   "avoid",
		Short: "avoid controller",
	}
)

func init() {
	flag.IntVar(&port, "port", 55555, "avoid server port")
	flag.StringVar(&server, "server", "localhost", "avoid server address")
	addr = fmt.Sprintf("%s:%d", server, port)
}

func main() {
	list := &cobra.Command{
		Use:   "list",
		Short: "list endpoint related data",
	}
	root.AddCommand(list)

	get := &cobra.Command{
		Use:   "get",
		Short: "get endpoint related data",
	}
	root.AddCommand(get)

	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "tell a UE time to move on",
	}
	root.AddCommand(migrate)

	ListConnInfo := &cobra.Command{
		Use:   "conn",
		Short: "list connections associated with this endpoint",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ListConnectionsFunc()
		},
	}
	list.AddCommand(ListConnInfo)

	GetStatsUE := &cobra.Command{
		Use:   "ue <id>",
		Short: "get stats on a UE",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			GetStatsFunc(args[0])
		},
	}
	get.AddCommand(GetStatsUE)

	MigrateUE := &cobra.Command{
		Use:   "ue <id> <lighthouse|endpoint|radio|network> <value>",
		Short: "move the UE to another thing",
		Long:  "tell the UE through nebula tunnel that it needs to change values",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			MigrateUEFunc(args[0], args[1], args[2])
		},
	}
	migrate.AddCommand(MigrateUE)

	root.Execute()
}

func MigrateUEFunc(name, typeMigrate, value string) {
	req := &avoid.MigrateRequest{
		Name:    name,
		Migrate: &avoid.ConnectionReply{},
	}
	switch typeMigrate {
	case "lighthouse":
		req.Migrate.Connection = avoid.ConnectionReply_Lighthouse
		break
	case "endpoint":
		req.Migrate.Connection = avoid.ConnectionReply_Endpoint
		break
	case "radio":
		req.Migrate.Connection = avoid.ConnectionReply_Radio
		break
	case "network":
		req.Migrate.Connection = avoid.ConnectionReply_Network
		break
	default:
		log.Errorf("unknown migration type: %s\n", typeMigrate)
		return
	}

	withAvoid(addr, func(c avoid.ManagementClient) error {
		log.Debugf("sending migrate request: %v\n", req)
		resp, err := c.Migrate(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Message Received with response: %s\n", resp)

		return nil
	})
}

func GetStatsFunc(ue string) {
	withAvoid(addr, func(c avoid.ManagementClient) error {
		req := &avoid.StatsRequest{Name: ue}
		log.Debugf("sent request: %v\n", req)
		resp, err := c.GetStats(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Statistics:\n")
		for k, v := range resp.Stats {
			fmt.Printf("\t%s: %s\n", k, v)
		}

		return nil
	})
}

func ListConnectionsFunc() {
	withAvoid(addr, func(c avoid.ManagementClient) error {
		req := &avoid.ListRequest{}
		log.Debugf("sent list request\n")
		resp, err := c.ListConnections(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Connections:\n")
		fmt.Printf("\tName: Uuid\t\tDuration\t\tLast Seen\n")
		for _, v := range resp.Info {
			fmt.Printf("\t%s: %s\t\t%s\t\t%s\n", v.Name, v.Duration, v.Lastseen)
		}

		return nil
	})
}

func withAvoid(endpoint string, f func(avoid.ManagementClient) error) error {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to avoid service: %v", err)
	}

	client := avoid.NewManagementClient(conn)
	defer conn.Close()

	return f(client)
}
