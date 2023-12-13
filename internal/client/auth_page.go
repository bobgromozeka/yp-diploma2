package client

import (
	"github.com/rivo/tview"
)

// AuthPage page with auth variants (sign up, sign in)
type AuthPage struct {
	app *Application
}

// NewAuthPage returns pointer of new auth page
func NewAuthPage(app *Application) *AuthPage {
	return &AuthPage{
		app: app,
	}
}

// Render building new auth page and returning tview component of it
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
