package commands

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/cmd/argocd/commands/headless"
	argocdclient "github.com/argoproj/argo-cd/v2/pkg/apiclient"
	applicationpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/util/argo"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/io"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

// NewApplicationDeploymentCommand returns a new instance of an `argocd app deployment` command
func NewApplicationDeploymentCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	var command = &cobra.Command{
		Use:   "deployments",
		Short: "Retrieve information about an app's deployments",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
			os.Exit(1)
		},
	}
	command.AddCommand(NewApplicationDeploymentValuesCommand(clientOpts))
	return command
}

// NewApplicationDeploymentValuesCommand returns a new instance of an `argocd app deployment source` command
func NewApplicationDeploymentValuesCommand(clientOpts *argocdclient.ClientOptions) *cobra.Command {
	var output string
	var namespace string

	var command = &cobra.Command{
		Use:   "source APPNAME DEPLOYMENT",
		Short: "Retrieve the source used for a deployment",
	}

	command.Run = func(c *cobra.Command, args []string) {
		ctx := c.Context()

		if len(args) != 2 {
			c.HelpFunc()(c, args)
			os.Exit(1)
		}
		appName, appNs := argo.ParseAppQualifiedName(args[0], namespace)
		deploymentIDStr := args[1]
		deploymentID, err := strconv.ParseInt(deploymentIDStr, 10, 64)
		errors.CheckError(err)

		conn, appIf := headless.NewClientOrDie(clientOpts, c).NewApplicationClientOrDie()
		defer io.Close(conn)

		app, err := appIf.Get(ctx, &applicationpkg.ApplicationQuery{
			Name:         &appName,
			AppNamespace: &appNs,
		})
		errors.CheckError(err)

		revision, err := findRevisionHistory(app, deploymentID)
		errors.CheckError(err)

		switch output {
		case "yaml":
			yamlBytes, err := yaml.Marshal(revision.Source)
			errors.CheckError(err)
			fmt.Println(string(yamlBytes))
		case "json":
			jsonBytes, err := json.MarshalIndent(revision.Source, "", "  ")
			errors.CheckError(err)
			fmt.Println(string(jsonBytes))
		default:
			log.Fatal("Only yaml or json output formats are supported")
		}
	}

	command.Flags().StringVarP(&output, "out", "o", "yaml", "Output format. One of: yaml, json")
	command.Flags().StringVar(&namespace, "namespace", "", "Namespace")

	return command
}
