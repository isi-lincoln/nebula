package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
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
	flag.IntVar(&port, "port", 55554, "avoid tunnel port")
	flag.StringVar(&server, "server", "localhost", "avoid tunnel address")
	addr = fmt.Sprintf("%s:%d", server, port)
}

func main() {
	Register := &cobra.Command{
		Use:   "register",
		Short: "register with endpoint",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			registerFunc()
		},
	}
	root.AddCommand(Register)

	HealthCheck := &cobra.Command{
		Use:   "health <id>",
		Short: "manage statistics",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			healthCheckFunc(args[0])
		},
	}
	root.AddCommand(HealthCheck)

	Watch := &cobra.Command{
		Use:   "watch <id>",
		Short: "watch on if we need to migrate",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			watchFunc(args[0])
		},
	}
	root.AddCommand(Watch)

	root.Execute()
}

const healthCheckMethod = "/grpc.health.v1.Health/Watch"

// https://github.com/grpc/grpc-go/blob/master/health/client.go
// APL 2 license
func clientHealthCheck(ctx context.Context, newStream func(string) (any, error), setConnectivityState func(connectivity.State, error), service string) error {

retryConnection:
	for {
		// Backs off if the connection has failed in some way without receiving a message in the previous retry.
		if ctx.Err() != nil {
			return nil
		}
		setConnectivityState(connectivity.Connecting, nil)
		rawS, err := newStream(healthCheckMethod)
		if err != nil {
			continue retryConnection
		}

		s, ok := rawS.(grpc.ClientStream)
		// Ideally, this should never happen. But if it happens, the server is marked as healthy for LBing purposes.
		if !ok {
			setConnectivityState(connectivity.Ready, nil)
			return fmt.Errorf("newStream returned %v (type %T); want grpc.ClientStream", rawS, rawS)
		}

		if err = s.SendMsg(&avoid.ConnectionRequest{Name: service}); err != nil && err != io.EOF {
			// Stream should have been closed, so we can safely continue to create a new stream.
			continue retryConnection
		}
		s.CloseSend()

		resp := new(avoid.ConnectionReply)
		for {
			err = s.RecvMsg(resp)

			// Reports healthy for the LBing purposes if health check is not implemented in the server.
			if status.Code(err) == codes.Unimplemented {
				setConnectivityState(connectivity.Ready, nil)
				return err
			}

			// Reports unhealthy if server's Watch method gives an error other than UNIMPLEMENTED.
			if err != nil {
				setConnectivityState(connectivity.TransientFailure, fmt.Errorf("connection active but received health check RPC error: %v", err))
				continue retryConnection
			}

			// As a message has been received, removes the need for backoff for the next retry by resetting the try count.
			if resp.Status == avoid.ConnectionReply_SERVING {
				setConnectivityState(connectivity.Ready, nil)
			} else {
				setConnectivityState(connectivity.TransientFailure, fmt.Errorf("connection active but health check failed. status=%s", resp.Status))
			}
		}
	}
}

func watchFunc(name string) {
	req := &avoid.ConnectionRequest{
		Name: name,
	}

	for {
		withAvoid(addr, func(c avoid.TunnelClient) error {

			stream, err := c.Watch(context.Background(), req)
			if err != nil {
				log.Error(err)
				log.Errorf("backoff: 5 seconds\n")
				time.Sleep(5 * time.Second)
				return err
			}

			for {
				resp, err := stream.Recv()
				if err != nil {
					log.Errorf("backoff: 5 seconds\n")
					time.Sleep(5 * time.Second)
					return err
				}

				log.Infof("%s: %s -> %s", resp.Connection, resp.Value, resp.Status)

				log.Infof("next check: 5 seconds\n")
				time.Sleep(5 * time.Second)
			}
		})
	}
}

func healthCheckFunc(ue string) {
	withAvoid(addr, func(c avoid.TunnelClient) error {
		req := &avoid.HealthRequest{Json: ue}
		log.Debugf("sent request: %v\n", req)
		resp, err := c.HealthCheck(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("HC: %v\n", resp.Json)

		return nil
	})
}

func registerFunc() {
	withAvoid(addr, func(c avoid.TunnelClient) error {
		req := &avoid.RegisterRequest{}
		log.Debugf("sent register request\n")
		resp, err := c.Register(context.TODO(), req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Token: %s\n", resp.Token)

		return nil
	})
}

func withAvoid(endpoint string, f func(avoid.TunnelClient) error) error {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to avoid service: %v", err)
	}

	client := avoid.NewTunnelClient(conn)
	defer conn.Close()

	return f(client)
}
