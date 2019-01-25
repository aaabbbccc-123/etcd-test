package cmd

import (
	"sync"
	"time"

	"go.etcd.io/etcd/pkg/transport"

	"github.com/spf13/cobra"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "etcd-test",
	Short: "A benchmark tool for etcd3",
}

var (
	endpoints    string
	totalConns   uint
	totalClients uint

	bar *pb.ProgressBar
	wg  sync.WaitGroup

	tls transport.TLSInfo

	user string

	dialTimeout time.Duration

	targetLeader bool
)

func init() {
	RootCmd.PersistentFlags().StringVar(&endpoints, "endpoints", "127.0.0.1:2379", "gRPC endpoints")
	RootCmd.PersistentFlags().UintVar(&totalConns, "conns", 1, "Total number of gRPC connections")
	RootCmd.PersistentFlags().UintVar(&totalClients, "clients", 1, "Total number of gRPC clients")

	RootCmd.PersistentFlags().StringVar(&tls.CertFile, "cert", "", "identify HTTPS client using this SSL certificate file")
	RootCmd.PersistentFlags().StringVar(&tls.KeyFile, "key", "", "identify HTTPS client using this SSL key file")
	RootCmd.PersistentFlags().StringVar(&tls.TrustedCAFile, "cacert", "", "verify certificates of HTTPS-enabled servers using this CA bundle")

	RootCmd.PersistentFlags().StringVar(&user, "user", "", "provide username[:password] and prompt if password is not supplied.")
	RootCmd.PersistentFlags().DurationVar(&dialTimeout, "dial-timeout", 0, "dial timeout for client connections")
}
