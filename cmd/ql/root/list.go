package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"questline/internal/ui"
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

			fmt.Fprintln(cmd.OutOrStdout(), ui.Heading(ui.IconQuest, "Quest Log"))

			tasks, err := svc.TaskRepo().ListAll(ctx)
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), ui.Muted.Render("(empty — add your first quest with: ql add \"My first task\")"))
				return nil
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

				icon := ui.KindIcon(t.IsProject, t.IsHabit)
				line := fmt.Sprintf("%s%s%s #%d %s %s", prefix, branch, icon, t.ID, t.Title, ui.Muted.Render("("+ui.StatusText(t.Status)+")"))
				fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(line))

				kids := children[id]
				for i := range kids {
					render(kids[i], nextPrefix, i == len(kids)-1)
				}
			}

			for i := range roots {
				// Render roots without the leading branch so the tree is stable.
				rootTask := tasks[byID[roots[i]]]
				icon := ui.KindIcon(rootTask.IsProject, rootTask.IsHabit)
				fmt.Fprintf(cmd.OutOrStdout(), "%s #%d %s %s\n", icon, rootTask.ID, rootTask.Title, ui.Muted.Render("("+ui.StatusText(rootTask.Status)+")"))
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
