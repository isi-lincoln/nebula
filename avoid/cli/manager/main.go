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

	disconnect := &cobra.Command{
		Use:   "disconnect",
		Short: "disconnect UE from thing",
	}
	root.AddCommand(disconnect)

	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "migrate UE from thing",
	}
	root.AddCommand(migrate)

	migrateRelay := &cobra.Command{
		Use:   "relay <ue id> <dst ip> <relay ip>",
		Short: "move the UE to another relay",
		Long:  "tell the UE through nebula tunnel that it needs to change values",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			ActionUEFunc(args[0], args[1], args[2], avoid.ActionMessage_MIGRATE)
		},
	}
	migrate.AddCommand(migrateRelay)

	disconnectUE := &cobra.Command{
		Use:   "<id> <lighthouse|endpoint|radio|network|relay>",
		Short: "move the UE to another thing",
		Long:  "tell the UE through nebula tunnel that it needs to change values",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ActionUEFunc(args[0], args[1], "", avoid.ActionMessage_DISCONNECT)
		},
	}
	disconnect.AddCommand(disconnectUE)

	root.Execute()
}

func ActionUEFunc(ueUUID, dest, changed string, am avoid.ActionMessage_Action) {

	// TODO: change the protobuf to have a standardized value set for each action
	req := &avoid.ActionRequest{
		Identifier: ueUUID,
		Action: &avoid.ActionMessage{
			Connection: avoid.ActionMessage_RELAY,
			Action:     am,
		},
		Values: []string{dest, changed},
	}
	/*
		log.Infof("type: %s", typeMigrate)
		switch typeMigrate {
		case "lighthouse":
			log.Errorf("Not implemented.")
			return
		case "endpoint":
			log.Errorf("Not implemented.")
			return
		case "radio":
			log.Errorf("Not implemented.")
			return
		case "network":
			log.Errorf("Not implemented.")
			return
		case "relay":
			req.Action.Connection = avoid.ActionMessage_RELAY
			break
		default:
			log.Errorf("unknown migration type: %s\n", typeMigrate)
			return
		}
	*/

	// TODO: tls in WithAvoidManager
	addr := fmt.Sprintf("%s:%d", clientServer, clientPort)
	avoid.WithAvoidManager(addr, nil, func(c avoid.AvoidManagerClient) error {
		log.Debugf("sending disconnect request: %v\n", req)
		resp, err := c.Action(context.TODO(), req)
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
