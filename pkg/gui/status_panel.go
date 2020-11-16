package gui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

// never call this on its own, it should only be called from within refreshCommits()
func (gui *Gui) refreshStatus() {
	gui.Mutexes.RefreshingStatusMutex.Lock()
	defer gui.Mutexes.RefreshingStatusMutex.Unlock()

	currentBranch := gui.currentBranch()
	if currentBranch == nil {
		// need to wait for branches to refresh
		return
	}
	status := ""

	if currentBranch.Pushables != "" && currentBranch.Pullables != "" {
		trackColor := color.FgYellow
		if currentBranch.Pushables == "0" && currentBranch.Pullables == "0" {
			trackColor = color.FgGreen
		} else if currentBranch.Pushables == "?" && currentBranch.Pullables == "?" {
			trackColor = color.FgRed
		}

		status = utils.ColoredString(fmt.Sprintf("↑%s↓%s ", currentBranch.Pushables, currentBranch.Pullables), trackColor)
	}

	if gui.GitCommand.WorkingTreeState() != commands.REBASE_MODE_NORMAL {
		status += utils.ColoredString(fmt.Sprintf("(%s) ", gui.GitCommand.WorkingTreeState()), color.FgYellow)
	}

	name := utils.ColoredString(currentBranch.Name, presentation.GetBranchColor(currentBranch.Name))
	repoName := utils.GetCurrentRepoName()
	status += fmt.Sprintf("%s → %s ", repoName, name)

	gui.g.Update(func(*gocui.Gui) error {
		gui.setViewContent(gui.getStatusView(), status)
		return nil
	})
}

func runeCount(str string) int {
	return len([]rune(str))
}

func cursorInSubstring(cx int, prefix string, substring string) bool {
	return cx >= runeCount(prefix) && cx < runeCount(prefix+substring)
}

func (gui *Gui) handleCheckForUpdate(g *gocui.Gui, v *gocui.View) error {
	gui.Updater.CheckForNewUpdate(gui.onUserUpdateCheckFinish, true)
	return gui.createLoaderPanel(v, gui.Tr.CheckingForUpdates)
}

func (gui *Gui) handleStatusClick(g *gocui.Gui, v *gocui.View) error {
	// TODO: move into some abstraction (status is currently not a listViewContext where a lot of this code lives)
	if gui.popupPanelFocused() {
		return nil
	}

	currentBranch := gui.currentBranch()
	if currentBranch == nil {
		// need to wait for branches to refresh
		return nil
	}

	if err := gui.switchContext(gui.Contexts.Status.Context); err != nil {
		return err
	}

	cx, _ := v.Cursor()
	upstreamStatus := fmt.Sprintf("↑%s↓%s", currentBranch.Pushables, currentBranch.Pullables)
	repoName := utils.GetCurrentRepoName()
	switch gui.GitCommand.WorkingTreeState() {
	case commands.REBASE_MODE_REBASING, commands.REBASE_MODE_MERGING:
		workingTreeStatus := fmt.Sprintf("(%s)", gui.GitCommand.WorkingTreeState())
		if cursorInSubstring(cx, upstreamStatus+" ", workingTreeStatus) {
			return gui.handleCreateRebaseOptionsMenu()
		}
		if cursorInSubstring(cx, upstreamStatus+" "+workingTreeStatus+" ", repoName) {
			return gui.handleCreateRecentReposMenu()
		}
	default:
		if cursorInSubstring(cx, upstreamStatus+" ", repoName) {
			return gui.handleCreateRecentReposMenu()
		}
	}

	return gui.handleStatusSelect()
}

func (gui *Gui) handleStatusSelect() error {
	// TODO: move into some abstraction (status is currently not a listViewContext where a lot of this code lives)
	if gui.popupPanelFocused() {
		return nil
	}

	magenta := color.New(color.FgMagenta)

	dashboardString := strings.Join(
		[]string{
			lazygitTitle(),
			"Copyright (c) 2018 Jesse Duffield",
			"Keybindings: https://github.com/jesseduffield/lazygit/blob/master/docs/keybindings",
			"Config Options: https://github.com/jesseduffield/lazygit/blob/master/docs/Config.md",
			"Tutorial: https://youtu.be/VDXvbHZYeKY",
			"Raise an Issue: https://github.com/jesseduffield/lazygit/issues",
			magenta.Sprint("Become a sponsor (github is matching all donations for 12 months): https://github.com/sponsors/jesseduffield"), // caffeine ain't free
			gui.Tr.ReleaseNotes,
		}, "\n\n")

	return gui.refreshMainViews(refreshMainOpts{
		main: &viewUpdateOpts{
			title: "",
			task:  gui.createRenderStringTask(dashboardString),
		},
	})
}

func (gui *Gui) handleOpenConfig(g *gocui.Gui, v *gocui.View) error {
	return gui.openFile(gui.Config.GetUserConfigPath())
}

func (gui *Gui) handleEditConfig(g *gocui.Gui, v *gocui.View) error {
	filename := gui.Config.GetUserConfigPath()
	return gui.editFile(filename)
}

func lazygitTitle() string {
	return `
   _                       _ _
  | |                     (_) |
  | | __ _ _____   _  __ _ _| |_
  | |/ _` + "`" + ` |_  / | | |/ _` + "`" + ` | | __|
  | | (_| |/ /| |_| | (_| | | |_
  |_|\__,_/___|\__, |\__, |_|\__|
                __/ | __/ |
               |___/ |___/       `
}

func (gui *Gui) workingTreeState() string {
	rebaseMode, _ := gui.GitCommand.RebaseMode()
	if rebaseMode != "" {
		return commands.REBASE_MODE_REBASING
	}
	merging, _ := gui.GitCommand.IsInMergeState()
	if merging {
		return commands.REBASE_MODE_MERGING
	}
	return commands.REBASE_MODE_NORMAL
}
