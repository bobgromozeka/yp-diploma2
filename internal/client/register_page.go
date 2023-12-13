package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rivo/tview"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
)

// RegisterPage page with registration form
type RegisterPage struct {
	app *Application
}

// NewRegisterPage return pointer of RegisterPage
func NewRegisterPage(app *Application) *RegisterPage {
	return &RegisterPage{
		app: app,
	}
}

// Render creates register page and returns tview component
func (p *RegisterPage) Render() *tview.Pages {
	page := pages("Register")

	form := tview.NewForm()

	login := ""
	_ = fmt.Sprintf("%s", login)

	password := ""
	confirmPassword := ""

	confirmMx := sync.Mutex{}

	removeError := func() {
		if index := form.GetFormItemIndex("error"); index != -1 {
			form.RemoveFormItem(index)
		}
	}

	form.
		AddInputField(
			"Login", "", 150, nil, func(text string) {
				removeError()
				login = text
			},
		).
		AddPasswordField(
			"Password", "", 150, '*', func(text string) {
				removeError()
				password = text
			},
		).
		AddPasswordField(
			"Confirm password", "", 150, '*', func(text string) {
				removeError()
				confirmPassword = text
			},
		).
		AddButton(
			"Confirm", func() {
				if !confirmMx.TryLock() {
					return
				}
				defer confirmMx.Unlock()

				removeError()
				if confirmPassword != password {
					tv := createErrorText("passwords must match")

					form.AddFormItem(tv)
				} else {
					ctx, stopLoader := context.WithCancel(context.Background())
					defer stopLoader()

					ltv := addLoadingTextView(ctx, p.app.tApp)
					form.AddFormItem(ltv)

					resp, respErr := p.app.uClient.SignUp(
						p.app.ctx, &user.SignUpRequest{
							Login:    login,
							Password: password,
						},
					)

					form.RemoveFormItem(form.GetFormItemIndex(ltv.GetLabel()))

					if respErr != nil {
						p.app.tApp.Stop()
						log.Fatalln("Server error: ", respErr)
					} else {
						if !resp.Success {
							tv := createErrorText(resp.Errors[0].Error)
							form.AddFormItem(tv)
							return
						}
						modal := tview.NewModal()
						modal.
							SetText("Success. You can now login using your credentials").
							AddButtons([]string{"Confirm"}).
							SetDoneFunc(
								func(buttonIndex int, buttonLabel string) {
									if buttonLabel == "Confirm" {
										p.app.tApp.SetRoot(NewAuthPage(p.app).Render(), true)
									}
								},
							)
						page.AddPage("modal", modal, false, true)
					}
				}
			},
		)

	page.AddPage("Form", form, true, true)

	return page
}
