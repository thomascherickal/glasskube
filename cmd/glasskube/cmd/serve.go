package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/glasskube/glasskube/internal/bootstrap"
	"github.com/glasskube/glasskube/internal/config"
	"github.com/glasskube/glasskube/internal/web"
	"github.com/glasskube/glasskube/pkg/client"
	"github.com/glasskube/glasskube/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"start", "ui"},
	Short:   "Open UI",
	Long:    `Start server and open the UI.`,
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var support *web.ServerConfigSupport
		cfg, err := kubeconfig.New(config.Kubeconfig)
		if err != nil {
			support = &web.ServerConfigSupport{
				KubeconfigError:           err,
				KubeconfigDefaultLocation: clientcmd.RecommendedHomeFile,
			}
			if clientcmd.IsEmptyConfig(err) {
				support.KubeconfigMissing = true
			}
		}
		if support == nil {
			isBootstrapped, err := bootstrap.IsBootstrapped(cmd.Context(), cfg)
			if !isBootstrapped || err != nil {
				support = &web.ServerConfigSupport{
					BootstrapMissing:    !isBootstrapped,
					BootstrapCheckError: err,
				}
			}
		}

		var ctx context.Context
		if cfg != nil {
			ctx, err = client.SetupContext(cmd.Context(), cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error occurred starting the webserver:\n\n%v\n", err)
				os.Exit(1)
				return
			} else {
				cmd.SetContext(ctx)
			}
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching the port number")
			os.Exit(1)
		}
		web.Port = port

		err = web.Start(ctx, support)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occurred starting the webserver:\n\n%v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8580, "Port for the webserver")
	RootCmd.AddCommand(serveCmd)
}
