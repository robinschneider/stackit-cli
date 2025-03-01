package plans

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
	"github.com/stackitcloud/stackit-cli/internal/pkg/services/opensearch/client"
	"github.com/stackitcloud/stackit-cli/internal/pkg/tables"

	"github.com/spf13/cobra"
	"github.com/stackitcloud/stackit-sdk-go/services/opensearch"
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
		Use:   "plans",
		Short: "Lists all OpenSearch service plans",
		Long:  "Lists all OpenSearch service plans.",
		Args:  args.NoArgs,
		Example: examples.Build(
			examples.NewExample(
				`List all OpenSearch service plans`,
				"$ stackit opensearch plans"),
			examples.NewExample(
				`List all OpenSearch service plans in JSON format`,
				"$ stackit opensearch plans --output-format json"),
			examples.NewExample(
				`List up to 10 OpenSearch service plans`,
				"$ stackit opensearch plans --limit 10"),
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
				return fmt.Errorf("get OpenSearch service plans: %w", err)
			}
			plans := *resp.Offerings
			if len(plans) == 0 {
				projectLabel, err := projectname.GetProjectName(ctx, cmd)
				if err != nil {
					projectLabel = model.ProjectId
				}
				cmd.Printf("No plans found for project %s\n", projectLabel)
				return nil
			}

			// Truncate output
			if model.Limit != nil && len(plans) > int(*model.Limit) {
				plans = plans[:*model.Limit]
			}

			return outputResult(cmd, model.OutputFormat, plans)
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
		Limit:           limit,
	}, nil
}

func buildRequest(ctx context.Context, model *inputModel, apiClient *opensearch.APIClient) opensearch.ApiListOfferingsRequest {
	req := apiClient.ListOfferings(ctx, model.ProjectId)
	return req
}

func outputResult(cmd *cobra.Command, outputFormat string, plans []opensearch.Offering) error {
	switch outputFormat {
	case globalflags.JSONOutputFormat:
		details, err := json.MarshalIndent(plans, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal OpenSearch plans: %w", err)
		}
		cmd.Println(string(details))

		return nil
	default:
		table := tables.NewTable()
		table.SetHeader("OFFERING NAME", "ID", "NAME", "DESCRIPTION")
		for i := range plans {
			o := plans[i]
			for j := range *o.Plans {
				p := (*o.Plans)[j]
				table.AddRow(*o.Name, *p.Id, *p.Name, *p.Description)
			}
			table.AddSeparator()
		}
		table.EnableAutoMergeOnColumns(1)
		err := table.Display(cmd)
		if err != nil {
			return fmt.Errorf("render table: %w", err)
		}

		return nil
	}
}
