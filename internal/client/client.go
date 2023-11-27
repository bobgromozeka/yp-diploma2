package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
)

type Application struct {
	tApp    *tview.Application
	uClient user.UserClient
	ctx     context.Context
}

func (a *Application) Run() error {
	return a.tApp.SetRoot(a.createAuthMenu(a.tApp), true).Run()
}

func NewApplication(ctx context.Context) *Application {
	tApp := tview.NewApplication()

	conn, connErr := grpc.Dial(":14444", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if connErr != nil {
		log.Fatalln("grpc client error: ", connErr)
	}

	client := user.NewUserClient(conn)

	return &Application{
		tApp:    tApp,
		uClient: client,
		ctx:     ctx,
	}
}

func (a *Application) createAuthMenu(app *tview.Application) *tview.Pages {
	p := pages()

	list := tview.NewList()

	list.
		AddItem(
			"Login", "", 'l', func() {
				app.SetRoot(a.createLoginForm(app), true)
			},
		).
		AddItem(
			"Register", "", 'r', func() {
				app.SetRoot(a.createRegisterForm(app), true)
			},
		)

	p.AddPage("Auth", list, true, true)

	return p
}

func (a *Application) createRegisterForm(app *tview.Application) *tview.Pages {
	p := pages()

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

					ltv := addLoadingTextView(ctx, app)
					form.AddFormItem(ltv)

					resp, respErr := a.uClient.SignUp(
						a.ctx, &user.SignUpRequest{
							Login:    login,
							Password: password,
						},
					)

					form.RemoveFormItem(form.GetFormItemIndex(ltv.GetLabel()))

					if respErr != nil {
						a.tApp.Stop()
						log.Fatalln("Server error: ", respErr)
					} else {
						if !resp.Success {
							tv := createErrorText(resp.Errors[0].Error)
							form.AddFormItem(tv)
						}
						modal := tview.NewModal()
						modal.
							SetText("Success. You can now login using your credentials").
							AddButtons([]string{"Confirm"}).
							SetDoneFunc(
								func(buttonIndex int, buttonLabel string) {
									if buttonLabel == "Confirm" {
										app.SetRoot(a.createAuthMenu(a.tApp), true)
									}
								},
							)
						p.AddPage("modal", modal, false, true)
					}
				}
			},
		)

	p.AddPage("Form", form, true, true)

	return p
}

func (a *Application) createLoginForm(app *tview.Application) *tview.Pages {
	p := pages()

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
		AddButton(
			"Confirm", func() {
				removeError()
				app.SetRoot(a.createAuthMenu(app), true)
			},
		)

	p.AddPage("Form", form, true, true)

	return p
}

func Run(ctx context.Context) error {
	return NewApplication(ctx).Run()
}

func pages() *tview.Pages {
	p := tview.NewPages()
	p.SetTitle("Auth")
	p.SetBorder(true)

	return p
}

func addLoadingTextView(ctx context.Context, app *tview.Application) *tview.TextView {
	tv := tview.NewTextView()
	tv.
		SetLabel(""). // Stupid hack. But library does not have loader and no proper way to do it
		SetText("Loading").
		SetLabelWidth(150).
		SetMaxLines(1).
		SetScrollable(false).
		SetSize(1, 0)

	go func() {
		for {
			app.QueueUpdateDraw(
				func() {
					tv.SetText("Loading.")
				},
			)
			time.Sleep(time.Millisecond * 300)
			app.QueueUpdateDraw(
				func() {
					tv.SetText("Loading..")
				},
			)
			time.Sleep(time.Millisecond * 300)
			app.QueueUpdateDraw(
				func() {
					tv.SetText("Loading...")
				},
			)
			time.Sleep(time.Millisecond * 300)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return tv
}

func createErrorText(text string) *tview.TextView {
	tv := tview.NewTextView()
	tv.
		SetLabel("error").
		SetText(text).
		SetLabelWidth(150).
		SetMaxLines(1).
		SetDynamicColors(true).
		SetScrollable(false).
		SetTextColor(tcell.ColorRed).
		SetSize(1, 0)

	return tv
}
