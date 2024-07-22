package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/spf13/cobra"
)

var (
	clientServer string
	clientPort   int
)

func main() {
	root := &cobra.Command{
		Use:   "avoid",
		Short: "avoid controller",
	}

	root.PersistentFlags().StringVarP(
		&clientServer, "server", "s", "localhost", "inventory service address to use")
	root.PersistentFlags().IntVarP(
		&clientPort, "port", "p", 55554, "inventory service port to use")

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

	disconnect := &cobra.Command{
		Use:   "disconnect",
		Short: "disconnect UE from network",
	}
	root.AddCommand(disconnect)

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
		Use:   "ue <id> <lighthouse|endpoint|radio|network|relay> <value>",
		Short: "move the UE to another thing",
		Long:  "tell the UE through nebula tunnel that it needs to change values",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			MigrateUEFunc(args[0], args[1], args[2])
		},
	}
	migrate.AddCommand(MigrateUE)

	DisconnectUE := &cobra.Command{
		Use:   "ue <uuid> <endpoint>",
		Short: "disconnect UE from avoid",
		Long:  "tell the UE through nebula tunnel to disconnect",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			DisconnectUEFunc(args[0], args[1])
		},
	}
	disconnect.AddCommand(DisconnectUE)

	root.Execute()
}

func DisconnectUEFunc(name, value string) {
	req := &avoid.ActionRequest{
		Identifier: name,
		Action: &avoid.ActionMessage{
			Connection: avoid.ActionMessage_RELAY,
			Action:     avoid.ActionMessage_DISCONNECT,
		},
	}

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	// TODO: tls in WithAvoidManager
	avoid.WithAvoidManager(addr, nil, func(c avoid.AvoidManagerClient) error {
		log.Debugf("sending disconnect request: %v\n", req)
		_, err := c.Disconnect(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Disconnect Message Sent\n")

		return nil
	})
}

func MigrateUEFunc(name, typeMigrate, value string) {
	req := &avoid.ActionRequest{
		Identifier: name,
		Action: &avoid.ActionMessage{
			Connection: avoid.ActionMessage_RELAY,
			Action:     avoid.ActionMessage_MIGRATE,
		},
	}
	log.Infof("type: %s", typeMigrate)
	switch typeMigrate {
	case "lighthouse":
		req.Action.Connection = avoid.ActionMessage_LIGHTHOUSE
		break
	case "endpoint":
		req.Action.Connection = avoid.ActionMessage_ENDPOINT
		break
	case "radio":
		req.Action.Connection = avoid.ActionMessage_RADIO
		break
	case "network":
		req.Action.Connection = avoid.ActionMessage_NETWORK
		break
	case "relay":
		req.Action.Connection = avoid.ActionMessage_RELAY
		break
	default:
		log.Errorf("unknown migration type: %s\n", typeMigrate)
		return
	}

	// TODO: tls in WithAvoidManager
	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	avoid.WithAvoidManager(addr, nil, func(c avoid.AvoidManagerClient) error {
		log.Debugf("sending disconnect request: %v\n", req)
		resp, err := c.Migrate(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Migrate Message Sent: %v\n", resp)

		return nil
	})
}

func GetStatsFunc(ue string) {
	req := &avoid.StatsRequest{Name: ue}

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	// TODO: tls in WithAvoidManager
	avoid.WithAvoidManager(addr, nil, func(c avoid.AvoidManagerClient) error {
		log.Debugf("sent request: %v\n", req)
		resp, err := c.GetStats(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: nicify
		fmt.Printf("Statistics: %v\n", resp)

		return nil
	})
}

func ListConnectionsFunc() {
	req := &avoid.ListRequest{}

	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	// TODO: tls in WithAvoidManager
	avoid.WithAvoidManager(addr, nil, func(c avoid.AvoidManagerClient) error {
		log.Debugf("sent request: %v\n", req)
		resp, err := c.ListConnections(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Connections:\n")
		fmt.Printf("\tName: Uuid\t\tDuration\t\tLast Seen\n")
		for _, v := range resp.Info {
			fmt.Printf("\t%s: %v\t\t%v\n", v.Name, v.Duration, v.Lastseen)
		}

		return nil
	})
}
