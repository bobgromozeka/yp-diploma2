package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rivo/tview"
	"google.golang.org/grpc/metadata"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
)

// LoginPage page with login form
type LoginPage struct {
	app *Application
}

// NewLoginPage returns pointer of LoginPage
func NewLoginPage(app *Application) *LoginPage {
	return &LoginPage{
		app: app,
	}
}

// Render creates login page and returns tview component
func (p *LoginPage) Render() *tview.Pages {
	page := pages("Login")

	form := tview.NewForm()

	login := ""
	_ = fmt.Sprintf("%s", login)

	password := ""
	_ = fmt.Sprintf("%s", password)

	removeError := func() {
		if index := form.GetFormItemIndex("error"); index != -1 {
			form.RemoveFormItem(index)
		}
	}

	confirmMx := sync.Mutex{}

	form.
		AddInputField(
			"Login", "", 100, nil, func(text string) {
				removeError()
				login = text
			},
		).
		AddPasswordField(
			"Password", "", 100, '*', func(text string) {
				removeError()
				password = text
			},
		).
		AddButton(
			"Confirm", func() {
				if !confirmMx.TryLock() {
					return
				}
				defer confirmMx.Unlock()

				removeError()
				ctx, stopLoader := context.WithCancel(context.Background())
				defer stopLoader()

				ltv := addLoadingTextView(ctx, p.app.tApp)
				form.AddFormItem(ltv)

				resp, respErr := p.app.uClient.SignIn(
					p.app.ctx, &user.SignInRequest{
						Login:    login,
						Password: password,
					},
				)

				form.RemoveFormItem(form.GetFormItemIndex(ltv.GetLabel()))

				if respErr != nil {
					p.app.tApp.Stop()
					log.Fatalln("Server error: ", respErr)
				} else {
					if resp.Token == nil {
						tv := createErrorText(resp.Error)
						form.AddFormItem(tv)
						return
					}

					p.app.ctx = metadata.AppendToOutgoingContext(p.app.ctx, "Authorization", "Bearer "+*resp.Token)
					p.app.tApp.SetRoot(NewDataPage(p.app).Render(), true)
				}
			},
		)

	page.AddPage("Form", form, true, true)

	return page
}
