package root

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"questline/internal/engine"
	"questline/internal/ui"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show player stats and unlocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			svc, cleanup, err := openService(ctx)
			if err != nil {
				return err
			}
			defer cleanup()

			p, err := svc.PlayerRepo().GetOrCreateMain(ctx)
			if err != nil {
				return err
			}
			computedLevel := engine.LevelForTotalXP(p.XPTotal)
			nextReq := engine.XPRequiredForLevel(computedLevel + 1)
			toNext := nextReq - p.XPTotal
			if toNext < 0 {
				toNext = 0
			}

			fmt.Fprintln(cmd.OutOrStdout(), ui.Heading(ui.IconSparkle, "Player Status"))
			fmt.Fprintln(cmd.OutOrStdout(), ui.LabelValue("Level", computedLevel))
			fmt.Fprintln(cmd.OutOrStdout(), ui.LabelValue("Total XP", fmt.Sprintf("%d (next at %d, %d to go)", p.XPTotal, nextReq, toNext)))
			fmt.Fprintln(cmd.OutOrStdout(), "")

			fmt.Fprintln(cmd.OutOrStdout(), ui.H2.Render("📊 Attributes"))
			// Original 4 attributes
			fmt.Fprintf(cmd.OutOrStdout(), "- 💪 STR: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPStr), p.XPStr)
			fmt.Fprintf(cmd.OutOrStdout(), "- 🧠 INT: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPInt), p.XPInt)
			fmt.Fprintf(cmd.OutOrStdout(), "- 🧘 WIS: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPWis), p.XPWis)
			fmt.Fprintf(cmd.OutOrStdout(), "- 🎨 ART: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPArt), p.XPArt)
			// New 5 attributes
			fmt.Fprintf(cmd.OutOrStdout(), "- 🏠 HOME: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPHome), p.XPHome)
			fmt.Fprintf(cmd.OutOrStdout(), "- 🌲 OUT: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPOut), p.XPOut)
			fmt.Fprintf(cmd.OutOrStdout(), "- 📚 READ: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPRead), p.XPRead)
			fmt.Fprintf(cmd.OutOrStdout(), "- 🎬 CINEMA: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPCinema), p.XPCinema)
			fmt.Fprintf(cmd.OutOrStdout(), "- 💼 CAREER: lvl %d (xp %d)\n", engine.AttributeLevelForXP(p.XPCareer), p.XPCareer)
			fmt.Fprintln(cmd.OutOrStdout(), "")

			all, err := svc.TaskRepo().ListAll(ctx)
			if err != nil {
				return err
			}
			activeLeaf := 0
			for i := range all {
				if all[i].IsProject {
					continue
				}
				switch all[i].Status {
				case "pending", "active":
					activeLeaf++
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), ui.H2.Render("🔓 Gates"))
			fmt.Fprintf(cmd.OutOrStdout(), "- %s %d %s\n", ui.Key.Render("Max active tasks:"), engine.MaxActiveTasks(computedLevel), ui.Muted.Render(fmt.Sprintf("(currently %d)", activeLeaf)))
			maxDepth := engine.MaxSubtaskDepth(computedLevel)
			switch {
			case maxDepth == engine.SubtaskDepthUnlimited:
				fmt.Fprintln(cmd.OutOrStdout(), "- "+ui.Key.Render("Subtasks:")+" "+ui.Good.Render("enabled")+" "+ui.Muted.Render("(unlimited depth)"))
			case maxDepth == 0:
				fmt.Fprintln(cmd.OutOrStdout(), "- "+ui.Key.Render("Subtasks:")+" "+ui.Bad.Render("locked"))
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "- %s %s %s\n", ui.Key.Render("Subtasks:"), ui.Good.Render("enabled"), ui.Muted.Render(fmt.Sprintf("(max depth %d)", maxDepth)))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "- %s %s\n", ui.Key.Render("Habits:"), enabledStr(computedLevel >= engine.LevelHabits))
			fmt.Fprintf(cmd.OutOrStdout(), "- %s %s\n", ui.Key.Render("Projects:"), enabledStr(computedLevel >= engine.LevelProjects))
			// Show unlocked difficulty levels
			maxDiff := engine.MaxDifficultyForLevel(computedLevel)
			diffNames := []string{"Trivial", "Easy", "Medium", "Hard", "Epic"}
			diffUnlocked := make([]string, 0)
			for d := engine.DifficultyTrivial; d <= maxDiff; d++ {
				diffUnlocked = append(diffUnlocked, fmt.Sprintf("%d-%s", d, diffNames[d-1]))
			}
			nextDiffLevel := 0
			if maxDiff < engine.DifficultyEpic {
				nextDiffLevel = engine.DifficultyUnlockLevels[maxDiff+1]
			}
			diffInfo := ui.Good.Render(fmt.Sprintf("1-%d", maxDiff))
			if nextDiffLevel > 0 {
				diffInfo += " " + ui.Muted.Render(fmt.Sprintf("(next at L%d)", nextDiffLevel))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "- %s %s\n", ui.Key.Render("Difficulty:"), diffInfo)
			fmt.Fprintln(cmd.OutOrStdout(), "")

			// Ensure blueprint rows exist + unlocked statuses are up to date.
			if _, err := svc.EvaluateBlueprintUnlocks(ctx); err != nil {
				return err
			}

			statuses := []string{"available", "active", "completed"}
			titles := map[string]string{
				"available": "Blueprints (available)",
				"active":    "Blueprints (active)",
				"completed": "Blueprints (completed)",
			}
			for _, st := range statuses {
				list, err := svc.BlueprintRepo().ListByStatus(ctx, st)
				if err != nil {
					return err
				}
				if len(list) == 0 {
					continue
				}
				sort.Slice(list, func(i, j int) bool { return list[i].Code < list[j].Code })
				heading := titles[st]
				switch st {
				case "available":
					heading = "🟢 " + heading
				case "active":
					heading = "🟣 " + heading
				case "completed":
					heading = "🏁 " + heading
				}
				fmt.Fprintln(cmd.OutOrStdout(), ui.H2.Render(heading+":"))
				for i := range list {
					def := engine.GetBlueprintDef(list[i].Code)
					if def != nil && def.Description != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s\n", ui.Key.Render(list[i].Code), ui.Muted.Render(def.Description))
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", ui.Muted.Render(list[i].Code))
					}
				}
				fmt.Fprintln(cmd.OutOrStdout(), "")
			}

			return nil
		},
	}

	return cmd
}

func enabledStr(ok bool) string {
	if ok {
		return ui.Good.Render("enabled")
	}
	return ui.Bad.Render("locked")
}
