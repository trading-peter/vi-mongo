package component

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kopecmaciej/tview"
	"github.com/kopecmaciej/vi-mongo/internal/ai"
	"github.com/kopecmaciej/vi-mongo/internal/manager"
	"github.com/kopecmaciej/vi-mongo/internal/tui/core"
)

const (
	AIPromptID = "AIPrompt"
)

type AIPrompt struct {
	*core.BaseElement
	*core.FormModal

	docKeys []string
}

func NewAIPrompt() *AIPrompt {
	formModal := core.NewFormModal()
	a := &AIPrompt{
		BaseElement: core.NewBaseElement(),
		FormModal:   formModal,
	}

	a.SetIdentifier(AIPromptID)
	a.SetAfterInitFunc(a.init)

	return a
}

func (a *AIPrompt) init() error {
	a.setLayout()
	a.setStyle()
	a.setKeybindings()

	a.handleEvents()

	return nil
}

func (a *AIPrompt) setLayout() {
	a.SetBorder(true)
	a.SetTitle("AI Prompt")
	a.SetTitleAlign(tview.AlignCenter)
	a.Form.SetBorderPadding(2, 2, 2, 2)
}

func (a *AIPrompt) setStyle() {
	styles := a.App.GetStyles()
	a.SetStyle(styles)

	a.Form.SetFieldTextColor(styles.Connection.FormInputColor.Color())
	a.Form.SetFieldBackgroundColor(styles.Connection.FormInputBackgroundColor.Color())
	a.Form.SetLabelColor(styles.Connection.FormLabelColor.Color())
}

func (a *AIPrompt) setKeybindings() {
	k := a.App.GetKeys()
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case k.Contains(k.AIPrompt.CloseModal, event.Name()):
			a.App.Pages.RemovePage(AIPromptID)
			return nil
		case k.Contains(k.AIPrompt.ClearPrompt, event.Name()):
			a.Form.GetFormItem(1).(*tview.InputField).SetText("")
			return nil
		}
		return event
	})
}

func (a *AIPrompt) IsAIPromptFocused() bool {
	if a.App.GetFocus() == a.FormModal {
		return true
	}
	if a.App.GetFocus().GetIdentifier() == a.GetIdentifier() {
		return true
	}
	return false
}

func (a *AIPrompt) handleEvents() {
	go a.HandleEvents(AIPromptID, func(event manager.EventMsg) {
		switch event.Message.Type {
		case manager.StyleChanged:
			a.setStyle()
			a.Render()
		case manager.UpdateAutocompleteKeys:
			a.docKeys = event.Message.Data.([]string)
		}
	})
}

func (a *AIPrompt) Render() {
	a.Form.Clear(true)

	models, defaultModelIndex := ai.GetAiModels()

	a.Form.AddDropDown("Model:", models, defaultModelIndex, nil).
		AddInputField("Prompt:", "", 0, nil, nil).
		AddButton("Ask LLM", a.onSubmit).
		AddButton("Apply Query", a.onApplyQuery).
		AddTextView("Response:", "", 0, 3, true, false)
}

func (a *AIPrompt) onSubmit() {
	var driver ai.AIDriver

	_, model := a.Form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()
	prompt := a.Form.GetFormItem(1).(*tview.InputField).GetText()

	gptModels, _ := ai.GetGptModels()
	anthropicModels, _ := ai.GetAnthropicModels()
	switch {
	case slices.Contains(gptModels, model):
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			a.showError("OPENAI_API_KEY not found in environment variables")
			return
		}
		driver = ai.NewOpenAIDriver(apiKey)
	case slices.Contains(anthropicModels, model):
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			a.showError("ANTHROPIC_API_KEY not found in environment variables")
			return
		}
		driver = ai.NewAnthropicDriver(apiKey)
	default:
		a.showError(fmt.Sprintf("Invalid AI model selected: %s", model))
		return
	}

	systemMessage := fmt.Sprintf(`You are an assistant helping to create MongoDB queries. 
	Respond with valid MongoDB query syntax that can be directly used in a query bar. It's
	text based query, so don't use any Javascript or other programming language.
	
	Rules:
	1. Always use proper MongoDB operators (e.g., $regex, $exists, $gt, $lt, $in).
	2. Quote values that are not numbers or booleans.
	3. Use proper formatting for regex patterns (e.g., "^pattern").
	
	Available document keys: %s
	
	If the user makes a mistake with a key name, correct it based on the available keys.
	
	Important: Respond only with the exact query, without any additional explanation, 
	Example: { name: { $regex: "^john", $options: "i" }, age: { $gt: 30 }, isActive: true }
	`, strings.Join(a.docKeys, ", "))

	driver.SetSystemMessage(systemMessage)

	response, err := driver.GetResponse(prompt, model)
	if err != nil {
		a.showError(fmt.Sprintf("Error getting response: %v", err))
		return
	}

	a.showResponse(response)
}

func (a *AIPrompt) showError(message string) {
	a.Form.GetFormItem(2).(*tview.TextView).SetText(fmt.Sprintf("Error: %s", message)).SetTextColor(tcell.ColorRed)
}

func (a *AIPrompt) showResponse(response string) {
	a.Form.GetFormItem(2).(*tview.TextView).SetText(fmt.Sprintf("Response:\n%s", response)).SetTextColor(tcell.ColorGreen)
}

func (a *AIPrompt) onApplyQuery() {
	response := a.Form.GetFormItem(2).(*tview.TextView).GetText(true)
	if response == "" {
		a.showError("No query to apply. Please submit a prompt first.")
		return
	}

	query := strings.TrimPrefix(response, "Response:\n")

	a.App.GetManager().SendTo(ContentId, manager.EventMsg{
		Sender: a.GetIdentifier(),
		Message: manager.Message{
			Type: manager.UpdateQueryBar,
			Data: query,
		},
	})

	a.App.Pages.RemovePage(AIPromptID)
}
