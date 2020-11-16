package gui

func (gui *Gui) validateNotInFilterMode() (bool, error) {
	if gui.State.Modes.Filtering.Active() {
		err := gui.ask(askOpts{
			title:         gui.Tr.MustExitFilterModeTitle,
			prompt:        gui.Tr.MustExitFilterModePrompt,
			handleConfirm: gui.exitFilterMode,
		})

		return false, err
	}
	return true, nil
}

func (gui *Gui) exitFilterMode() error {
	gui.State.Modes.Filtering.Path = ""
	return gui.Errors.ErrRestart
}
