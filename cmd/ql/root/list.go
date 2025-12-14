package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks (tree view)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			svc, cleanup, err := openService(ctx)
			if err != nil {
				return err
			}
			defer cleanup()

			tasks, err := svc.TaskRepo().ListAll(ctx)
			if err != nil {
				return err
			}
			children := map[int64][]int64{}
			roots := []int64{}
			byID := map[int64]int{}
			for i := range tasks {
				byID[tasks[i].ID] = i
				if tasks[i].ParentID == nil {
					roots = append(roots, tasks[i].ID)
					continue
				}
				pid := *tasks[i].ParentID
				children[pid] = append(children[pid], tasks[i].ID)
			}

			var render func(id int64, prefix string, isLast bool)
			render = func(id int64, prefix string, isLast bool) {
				t := tasks[byID[id]]
				branch := "├─ "
				nextPrefix := prefix + "│  "
				if isLast {
					branch = "└─ "
					nextPrefix = prefix + "   "
				}

				kind := ""
				if t.IsProject {
					kind = "[P] "
				} else if t.IsHabit {
					kind = "[H] "
				}
				line := fmt.Sprintf("%s%s%d %s%s (status=%s)", prefix, branch, t.ID, kind, t.Title, t.Status)
				fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(line))

				kids := children[id]
				for i := range kids {
					render(kids[i], nextPrefix, i == len(kids)-1)
				}
			}

			for i := range roots {
				// Render roots without the leading branch so the tree is stable.
				rootTask := tasks[byID[roots[i]]]
				kind := ""
				if rootTask.IsProject {
					kind = "[P] "
				} else if rootTask.IsHabit {
					kind = "[H] "
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%d %s%s (status=%s)\n", rootTask.ID, kind, rootTask.Title, rootTask.Status)
				kids := children[rootTask.ID]
				for j := range kids {
					render(kids[j], "", j == len(kids)-1)
				}
			}
			return nil
		},
	}

	return cmd
}
