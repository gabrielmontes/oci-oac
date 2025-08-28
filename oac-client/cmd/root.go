package cmd

import (
	"fmt"
	"os"
	"strings"

	"oac-client/core/oac"

	"github.com/spf13/cobra"
)

// rootCmd is the main CLI command
var rootCmd = &cobra.Command{
	Use:   "oac <method> <path> [bodyFile]",
	Short: "OAC REST API client utility",
	Long: `OAC REST API client utility.

Examples:
  # GET a report
  oac-client GET /reports/123

  # POST a new report with JSON payload
  oac-client POST /reports payload.json

  # Update an existing report
  oac-client PUT /reports/123 update.json

Notes:
  - The bodyFile argument is mandatory for POST and PUT requests.
	`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		method := strings.ToUpper(args[0])
		path := args[1]

		var body string
		if requiresBody(method) {
			if len(args) < 3 {
				return fmt.Errorf("%s requires a body file", method)
			}
			body = args[2]
		}

		client, err := oac.NewOacClient()
		if err != nil {
			return fmt.Errorf("failed to create OAC client: %w", err)
		}

		resp, err := client.RestCall(method, path, body)
		if err != nil {
			return fmt.Errorf("error executing REST call: %w", err)
		}

		fmt.Println(resp)
		return nil
	},
}

// requiresBody returns true if the HTTP method requires a body
func requiresBody(method string) bool {
	return method == "POST" || method == "PUT"
}

// Execute runs the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// init is left empty but can be used to add subcommands if needed
func init() {}
