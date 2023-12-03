package client

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
)

type SelectCreationPage struct {
	app *Application
}

func NewSelectCreationPage(app *Application) *SelectCreationPage {
	return &SelectCreationPage{
		app: app,
	}
}

func (p *SelectCreationPage) Render() *tview.Pages {
	page := pages("create selection")

	list := tview.NewList()
	list.
		AddItem("Create password pair", "", 'p', func() {
			p.app.tApp.SetRoot(NewCreatePasswordPairPage(p.app).Render(), true)
		})

	page.
		AddPage("list", list, true, true)

	return page
}

type CreatePasswordPairPage struct {
	app *Application
}

func NewCreatePasswordPairPage(app *Application) *CreatePasswordPairPage {
	return &CreatePasswordPairPage{
		app: app,
	}
}

func (p *CreatePasswordPairPage) Render() *tview.Pages {
	page := pages("create password pair")

	var login string
	var password string
	var description string

	form := tview.NewForm()

	form.
		AddInputField("Login", "", 100, nil, func(text string) {
			login = text
		}).
		AddInputField("Password", "", 100, nil, func(text string) {
			password = text
		}).
		AddTextArea("Description", "", 100, 10, 0, func(text string) {
			description = text
		}).
		AddTextView("ESC to return back", "", 100, 1, false, false).
		AddButton("Confirm", func() {
			_, err := p.app.dClient.CreatePasswordPair(p.app.ctx, &datakeeper.CreatePasswordPairRequest{
				Login:       login,
				Password:    password,
				Description: &description,
			})
			if err != nil {
				failureModal := tview.NewModal().
					SetText("Error has occurred. Stopping application").
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						if buttonLabel == "OK" {
							p.app.tApp.Stop()
						}
					})
				page.AddPage("failure", failureModal, true, true)
				return
			}

			successModal := tview.NewModal().
				SetText("Password pair successfully added").
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "OK" {
						p.app.tApp.SetRoot(NewDataPage(p.app).Render(), true)
					}
				})

			page.AddPage("success", successModal, true, true)
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyESC {
				p.app.tApp.SetRoot(NewDataPage(p.app).Render(), true)
				return nil
			}

			return event
		})

	page.AddPage("form", form, true, true)

	return page
}
