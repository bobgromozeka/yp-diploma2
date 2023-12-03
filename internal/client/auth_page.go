package client

import (
	"github.com/rivo/tview"
)

type AuthPage struct {
	app *Application
}

func NewAuthPage(app *Application) *AuthPage {
	return &AuthPage{
		app: app,
	}
}

func (p *AuthPage) Render() *tview.Pages {
	page := pages("Auth")

	list := tview.NewList()

	list.
		AddItem(
			"Login", "", 'l', func() {
				p.app.tApp.SetRoot(NewLoginPage(p.app).Render(), true)
			},
		).
		AddItem(
			"Register", "", 'r', func() {
				p.app.tApp.SetRoot(NewRegisterPage(p.app).Render(), true)
			},
		)

	page.AddPage("Auth", list, true, true)

	return page
}
