package cmd

import (
	"fmt"
	"log"
	"os"

	"strings"

	"github.com/sema/raft/pkg/actor"
	"github.com/sema/raft/pkg/grpcserver"
	"github.com/sema/raft/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "sr",
	Short: "SR is a key value store.",
	Long: `SR is a simplistic key value store based on the raft algorithm.

The command starts a SR server node and blocks until the server exits.`,

	Run: func(cmd *cobra.Command, args []string) {
		localServerID := viper.GetString("name")
		localServerAddress := viper.GetString("listen")
		serverAddresses := viper.GetStringSlice("server")

		// This is not necessarily the most stellar interface, and error handling is lacking. Will make do for now.
		// TODO replace cobra & viper with urfave/cli and see if that provides a nicer interface
		if localServerID == "" {
			fmt.Println("--name is required")
			os.Exit(1)
		}

		if localServerAddress == "" {
			fmt.Println("--listen is required")
			os.Exit(1)
		}

		if len(serverAddresses) == 0 {
			fmt.Println("--server is required")
			os.Exit(1)
		}

		var allServers []actor.ServerID
		discovery := make(map[actor.ServerID]grpcserver.DiscoveryConfig)

		for _, serverAddress := range serverAddresses {
			parts := strings.SplitN(serverAddress, "=", 2)
			if len(parts) != 2 {
				log.Panicf("Invalid server address format for %s", serverAddress)
			}
			serverID := actor.ServerID(parts[0])
			addressRemote := parts[1]

			allServers = append(allServers, serverID)
			discovery[serverID] = grpcserver.DiscoveryConfig{
				ServerID:      serverID,
				AddressRemote: addressRemote,
			}
		}

		if _, ok := discovery[actor.ServerID(localServerID)]; !ok {
			allServers = append(allServers, actor.ServerID(localServerID))
		}

		config := actor.Config{
			Servers:                    allServers,
			LeaderElectionTimeout:      4,
			LeaderElectionTimeoutSplay: 4,
			LeaderHeartbeatFrequency:   2,
		}

		storage := actor.NewMemoryStorage()

		svr := server.NewServer(actor.ServerID(localServerID), storage, config)

		inboundServer := grpcserver.NewInboundGRPCServer(svr)
		go inboundServer.Serve(localServerAddress)
		defer inboundServer.Stop()

		outboundServer := grpcserver.NewOutboundGRPCServer(svr, discovery)
		go outboundServer.Serve()
		defer outboundServer.Stop()

		svr.Start() // Blocking call
		defer svr.Stop()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv() // Read environment variables when looking up config
	viper.SetEnvPrefix("SR")

	rootCmd.Flags().StringArray("server", []string{}, "remote server address, specified in the form {name}={ip}:{port}. Can also be set through the environment variable SR_SERVER (e.g. SR_SERVER=\"svc1=127.0.0.1:8000 svc2=127.0.0.1:8000\").")
	viper.BindPFlag("server", rootCmd.Flags().Lookup("server"))

	rootCmd.Flags().String("name", "", "name of local server Can also bet set through the environment variable SR_NAME.")
	viper.BindPFlag("name", rootCmd.Flags().Lookup("name"))

	rootCmd.Flags().String("listen", "", "address to bind local server to, specified in the form tcp:{ip}:{port}. Can also be set through the environment variable SR_LISTEN.")
	viper.BindPFlag("listen", rootCmd.Flags().Lookup("listen"))
}
