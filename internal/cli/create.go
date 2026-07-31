package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newCreateCmd() *cobra.Command {
	var (
		file   string
		from   string
		export string
		start  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a container from a YAML spec or an existing container",
		Long: "Create a container. Provide a spec with --file <yaml>, or clone an\n" +
			"existing container's config with --from <container>. Use --export to\n" +
			"print the equivalent `docker run` command or compose fragment instead\n" +
			"of creating anything.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if (file == "") == (from == "") {
				return fmt.Errorf("gantry: provide exactly one of --file or --from")
			}
			return withClient(cmd.Context(), func(ctx context.Context, cli docker.Client) error {
				spec, err := loadSpec(ctx, cli, file, from)
				if err != nil {
					return err
				}
				switch export {
				case "":
					id, err := cli.CreateContainer(ctx, spec, start)
					if err != nil {
						return err
					}
					fmt.Fprintln(cmd.OutOrStdout(), id)
				case "run":
					fmt.Fprintln(cmd.OutOrStdout(), docker.SpecToDockerRun(spec))
				case "compose":
					fmt.Fprint(cmd.OutOrStdout(), docker.SpecToCompose(spec))
				default:
					return fmt.Errorf("gantry: unknown --export %q (want run|compose)", export)
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "YAML spec file")
	cmd.Flags().StringVar(&from, "from", "", "Existing container to clone config from")
	cmd.Flags().StringVar(&export, "export", "", "Print the spec as run|compose instead of creating")
	cmd.Flags().BoolVar(&start, "start", false, "Start the container after creating it")
	return cmd
}

func loadSpec(ctx context.Context, cli docker.Client, file, from string) (docker.CreateSpec, error) {
	if from != "" {
		return cli.SpecFromContainer(ctx, from)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return docker.CreateSpec{}, fmt.Errorf("gantry: reading spec file: %w", err)
	}
	var spec docker.CreateSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return docker.CreateSpec{}, fmt.Errorf("gantry: parsing spec file: %w", err)
	}
	if spec.Image == "" {
		return docker.CreateSpec{}, fmt.Errorf("gantry: spec is missing 'image'")
	}
	return spec, nil
}
