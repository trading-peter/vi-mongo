package view

import (
	"fmt"
	"strings"

	"github.com/kopecmaciej/mongui/internal/config"
	"github.com/kopecmaciej/mongui/internal/views/core"
	"github.com/kopecmaciej/mongui/internal/views/modals"
	"github.com/kopecmaciej/tview"
)

const (
	WelcomeView = "Welcome"
)

type Welcome struct {
	*core.BaseView
	*tview.Flex

	// Form
	form *tview.Form

	// Callbacks
	onSubmit func()
}

func NewWelcome() *Welcome {
	w := &Welcome{
		BaseView: core.NewBaseView(WelcomeView),
		Flex:     tview.NewFlex(),
		form:     tview.NewForm(),
	}

	return w
}

func (w *Welcome) Init(app *core.App) error {
	w.App = app

	w.setStyle()

	w.Render()

	return nil
}

func (w *Welcome) setStyle() {
	style := w.App.GetStyles().Welcome

	w.form.SetBorder(true)
	w.form.SetBackgroundColor(style.BackgroundColor.Color())
	w.form.SetFieldTextColor(style.FormInputColor.Color())
	w.form.SetFieldBackgroundColor(style.FormInputBackgroundColor.Color())
	w.SetBackgroundColor(style.BackgroundColor.Color())
	w.SetBorder(false)
}

func (w *Welcome) Render() {
	w.Flex.Clear()

	// easy way to center the form
	w.AddItem(tview.NewBox(), 0, 1, false)

	w.renderForm()
	w.Flex.AddItem(w.form, 80, 0, true)

	w.AddItem(tview.NewBox(), 0, 1, false)

	w.App.SetRoot(w.Flex, true)
}

func (w *Welcome) SetOnSubmitFunc(onSubmit func()) {
	w.onSubmit = onSubmit
}

func (w *Welcome) renderForm() {
	w.form.Clear(true)

	w.form.SetBorder(true)
	w.form.SetTitle(" Welcome to MongoUI ")
	w.form.SetTitleAlign(tview.AlignCenter)

	configFile, err := config.GetConfigPath()
	if err != nil {
		modals.ShowError(w.App.Pages, "Error while getting config path", err)
		return
	}
	welcomeText := "All configuration can be set in " + configFile + " file. You can also set it here."
	w.form.AddTextView("Welcome", welcomeText, 60, 2, true, false)
	w.form.AddTextView("Editor info", "Set command (vim, nano etc) or env variable ($ENV) to open editor", 60, 2, true, false)
	w.form.AddInputField("Editor", "$EDITOR", 30, nil, nil)
	w.form.AddTextView("Logs", "Requires restart if changed", 60, 1, true, false)
	w.form.AddInputField("Log File", "/tmp/mongui.log", 30, nil, nil)
	w.form.AddInputField("Log Level", "info", 30, nil, nil)
	w.form.AddCheckbox("Show connection page", true, nil)
	w.form.AddCheckbox("Show welcome page", false, nil)
	w.form.AddTextView("Show help", fmt.Sprintf("Press %s to show help", w.App.GetKeys().Global.ToggleFullScreenHelp.String()), 60, 1, true, false)

	w.form.AddButton(" Save and Connect ", func() {
		err := w.saveConfig()
		if err != nil {
			modals.ShowError(w.App.Pages, "Error while saving config", err)
			return
		}
		w.onSubmit()
	})

	w.form.AddButton(" Exit ", func() {
		w.App.Stop()
	})
}

func (w *Welcome) saveConfig() error {
	editorCmd := w.form.GetFormItemByLabel("Editor").(*tview.InputField).GetText()
	logFile := w.form.GetFormItemByLabel("Log File").(*tview.InputField).GetText()
	logLevel := w.form.GetFormItemByLabel("Log Level").(*tview.InputField).GetText()
	connPage := w.form.GetFormItemByLabel("Show connection page").(*tview.Checkbox).IsChecked()
	welcomePage := w.form.GetFormItemByLabel("Show welcome page").(*tview.Checkbox).IsChecked()

	c := w.App.GetConfig()

	splitedEditorCmd := strings.Split(editorCmd, "$")
	if len(splitedEditorCmd) > 1 {
		c.Editor.Command = ""
		c.Editor.Env = splitedEditorCmd[1]
	} else {
		c.Editor.Env = ""
		c.Editor.Command = editorCmd
	}
	c.Log.Path = logFile
	c.Log.Level = logLevel
	c.ShowConnectionPage = connPage
	c.ShowWelcomePage = welcomePage

	err := w.App.GetConfig().UpdateConfig()
	if err != nil {
		return err
	}

	return nil
}
