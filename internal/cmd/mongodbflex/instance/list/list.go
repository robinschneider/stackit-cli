package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stackitcloud/stackit-cli/internal/pkg/args"
	"github.com/stackitcloud/stackit-cli/internal/pkg/errors"
	"github.com/stackitcloud/stackit-cli/internal/pkg/examples"
	"github.com/stackitcloud/stackit-cli/internal/pkg/flags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/globalflags"
	"github.com/stackitcloud/stackit-cli/internal/pkg/projectname"
	"github.com/stackitcloud/stackit-cli/internal/pkg/services/mongodbflex/client"
	"github.com/stackitcloud/stackit-cli/internal/pkg/tables"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/mongodbflex"
)

const (
	limitFlag = "limit"
)

type inputModel struct {
	*globalflags.GlobalFlagModel
	Limit *int64
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all MongoDB Flex instances",
		Long:  "Lists all MongoDB Flex instances.",
		Args:  args.NoArgs,
		Example: examples.Build(
			examples.NewExample(
				`List all MongoDB Flex instances`,
				"$ stackit mongodbflex instance list"),
			examples.NewExample(
				`List all MongoDB Flex instances in JSON format`,
				"$ stackit mongodbflex instance list --output-format json"),
			examples.NewExample(
				`List up to 10 MongoDB Flex instances`,
				"$ stackit mongodbflex instance list --limit 10"),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			model, err := parseInput(cmd)
			if err != nil {
				return err
			}

			// Configure API client
			apiClient, err := client.ConfigureClient(cmd)
			if err != nil {
				return err
			}

			// Call API
			req := buildRequest(ctx, model, apiClient)
			resp, err := req.Execute()
			if err != nil {
				return fmt.Errorf("get MongoDB Flex instances: %w", err)
			}
			if resp.Items == nil || len(*resp.Items) == 0 {
				projectLabel, err := projectname.GetProjectName(ctx, cmd)
				if err != nil {
					projectLabel = model.ProjectId
				}
				cmd.Printf("No instances found for project %s\n", projectLabel)
				return nil
			}
			instances := *resp.Items

			// Truncate output
			if model.Limit != nil && len(instances) > int(*model.Limit) {
				instances = instances[:*model.Limit]
			}

			return outputResult(cmd, model.OutputFormat, instances)
		},
	}

	configureFlags(cmd)
	return cmd
}

func configureFlags(cmd *cobra.Command) {
	cmd.Flags().Int64(limitFlag, 0, "Maximum number of entries to list")
}

func parseInput(cmd *cobra.Command) (*inputModel, error) {
	globalFlags := globalflags.Parse(cmd)
	if globalFlags.ProjectId == "" {
		return nil, &errors.ProjectIdError{}
	}

	limit := flags.FlagToInt64Pointer(cmd, limitFlag)
	if limit != nil && *limit < 1 {
		return nil, &errors.FlagValidationError{
			Flag:    limitFlag,
			Details: "must be greater than 0",
		}
	}

	return &inputModel{
		GlobalFlagModel: globalFlags,
		Limit:           flags.FlagToInt64Pointer(cmd, limitFlag),
	}, nil
}

func buildRequest(ctx context.Context, model *inputModel, apiClient *mongodbflex.APIClient) mongodbflex.ApiListInstancesRequest {
	req := apiClient.ListInstances(ctx, model.ProjectId).Tag("")
	return req
}

func outputResult(cmd *cobra.Command, outputFormat string, instances []mongodbflex.InstanceListInstance) error {
	switch outputFormat {
	case globalflags.JSONOutputFormat:
		details, err := json.MarshalIndent(instances, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal MongoDB Flex instance list: %w", err)
		}
		cmd.Println(string(details))

		return nil
	default:
		table := tables.NewTable()
		table.SetHeader("ID", "NAME", "STATUS")
		for i := range instances {
			instance := instances[i]
			table.AddRow(*instance.Id, *instance.Name, *instance.Status)
		}
		err := table.Display(cmd)
		if err != nil {
			return fmt.Errorf("render table: %w", err)
		}

		return nil
	}
}
