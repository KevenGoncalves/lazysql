package components

import (
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers"
	"github.com/jorgerojas26/lazysql/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ConnectionForm struct {
	*tview.Flex
	*tview.Form
	StatusText *tview.TextView
	Action     string
}

func NewConnectionForm(connectionPages *models.ConnectionPages) *ConnectionForm {
	wrapper := tview.NewFlex()

	wrapper.SetDirection(tview.FlexColumnCSS)

	addForm := tview.NewForm().
		SetFieldBackgroundColor(tcell.ColorGray).
		SetButtonBackgroundColor(tcell.ColorDefault).
		SetLabelColor(tcell.ColorWhite.TrueColor()).
		SetFieldTextColor(tcell.ColorWhite.TrueColor())

	addForm.AddInputField("Name", "", 0, nil, nil)
	addForm.AddInputField("URL", "", 0, nil, nil)
	addForm.SetBackgroundColor(tcell.ColorDefault)

	buttonsWrapper := tview.NewFlex().SetDirection(tview.FlexColumn)

	saveButton := tview.NewButton("[red]F1 [white]Save")
	saveButton.SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	buttonsWrapper.AddItem(saveButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	testButton := tview.NewButton("[red]F2 [white]Test")
	testButton.SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	buttonsWrapper.AddItem(testButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	connectButton := tview.NewButton("[red]F3 [white]Connect")
	connectButton.SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	buttonsWrapper.AddItem(connectButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	cancelButton := tview.NewButton("[red]Esc [white]Cancel")
	cancelButton.SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault))
	buttonsWrapper.AddItem(cancelButton, 0, 1, false)

	statusText := tview.NewTextView()
	statusText.SetBorderPadding(0, 1, 0, 0)
	statusText.SetBackgroundColor(tcell.ColorDefault)

	wrapper.AddItem(addForm, 0, 1, true)
	wrapper.AddItem(statusText, 2, 0, false)
	wrapper.AddItem(buttonsWrapper, 1, 0, false)

	form := &ConnectionForm{
		Flex:       wrapper,
		Form:       addForm,
		StatusText: statusText,
	}

	wrapper.SetInputCapture(form.inputCapture(connectionPages))

	return form
}

func (form *ConnectionForm) inputCapture(connectionPages *models.ConnectionPages) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			connectionPages.SwitchToPage("Connections")
		} else if event.Key() == tcell.KeyF1 || event.Key() == tcell.KeyEnter {
			connectionName := form.GetFormItem(0).(*tview.InputField).GetText()

			if connectionName == "" {
				form.StatusText.SetText("Connection name is required").SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return event
			}

			connectionString := form.GetFormItem(1).(*tview.InputField).GetText()

			parsed, err := helpers.ParseConnectionString(connectionString)

			if err != nil {
				form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
				return event
			} else {
				password, _ := parsed.User.Password()
				databases, _ := helpers.LoadConnections()
				newDatabases := make([]models.Connection, len(databases))

				switch form.Action {
				case "create":

					database := models.Connection{
						Name:     connectionName,
						Provider: parsed.Driver,
						User:     parsed.User.Username(),
						Password: password,
						Host:     parsed.Hostname(),
						Port:     parsed.Port(),
						Query:    parsed.Query().Encode(),
						DBName:   helpers.ParsedDBName(parsed.Path),
					}

					newDatabases = append(databases, database)
					err := helpers.SaveConnectionConfig(newDatabases)
					if err != nil {
						form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
						return event
					}

				case "edit":
					newDatabases = make([]models.Connection, len(databases))
					row, _ := ConnectionListTable.GetSelection()
					for i, database := range databases {
						if i == row {

							newDatabases[i].Name = connectionName
							newDatabases[i].Provider = database.Provider
							newDatabases[i].User = parsed.User.Username()
							newDatabases[i].Password, _ = parsed.User.Password()
							newDatabases[i].Host = parsed.Hostname()
							newDatabases[i].Port = parsed.Port()
							newDatabases[i].Query = parsed.Query().Encode()
							newDatabases[i].DBName = helpers.ParsedDBName(parsed.Path)
						} else {
							newDatabases[i] = database
						}
					}

					err := helpers.SaveConnectionConfig(newDatabases)
					if err != nil {
						form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
						return event

					}
				}
				ConnectionListTable.SetConnections(newDatabases)
				connectionPages.SwitchToPage("Connections")
			}
		} else if event.Key() == tcell.KeyF2 {
			connectionString := form.GetFormItem(1).(*tview.InputField).GetText()
			go form.testConnection(connectionString)
		}
		return event
	}
}

func (form *ConnectionForm) testConnection(connectionString string) {
	form.StatusText.SetText("Connecting...").SetTextColor(tcell.ColorGreen)

	db := drivers.MySQL{}

	err := db.TestConnection(connectionString)

	if err != nil {
		form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
	} else {
		form.StatusText.SetText("Connection success").SetTextColor(tcell.ColorGreen)
	}
	App.ForceDraw()
}

func (form *ConnectionForm) SetAction(action string) {
	form.Action = action
}
