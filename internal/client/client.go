package client

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	token   *string
}

func (a *Application) Run() error {
	return a.tApp.SetRoot(a.createAuthMenu(), true).Run()
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

func (a *Application) createAuthMenu() *tview.Pages {
	p := pages("Auth")

	list := tview.NewList()

	list.
		AddItem(
			"Login", "", 'l', func() {
				a.tApp.SetRoot(a.createLoginForm(), true)
			},
		).
		AddItem(
			"Register", "", 'r', func() {
				a.tApp.SetRoot(a.createRegisterForm(), true)
			},
		)

	p.AddPage("Auth", list, true, true)

	return p
}

func (a *Application) createRegisterForm() *tview.Pages {
	p := pages("Register")

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

					ltv := addLoadingTextView(ctx, a.tApp)
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
							return
						}
						modal := tview.NewModal()
						modal.
							SetText("Success. You can now login using your credentials").
							AddButtons([]string{"Confirm"}).
							SetDoneFunc(
								func(buttonIndex int, buttonLabel string) {
									if buttonLabel == "Confirm" {
										a.tApp.SetRoot(a.createAuthMenu(), true)
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

func (a *Application) createLoginForm() *tview.Pages {
	p := pages("Login")

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
				if !confirmMx.TryLock() {
					return
				}
				defer confirmMx.Unlock()

				removeError()
				ctx, stopLoader := context.WithCancel(context.Background())
				defer stopLoader()

				ltv := addLoadingTextView(ctx, a.tApp)
				form.AddFormItem(ltv)

				resp, respErr := a.uClient.SignIn(
					a.ctx, &user.SignInRequest{
						Login:    login,
						Password: password,
					},
				)

				form.RemoveFormItem(form.GetFormItemIndex(ltv.GetLabel()))

				if respErr != nil {
					a.tApp.Stop()
					log.Fatalln("Server error: ", respErr)
				} else {
					if resp.Token == nil {
						tv := createErrorText(resp.Error)
						form.AddFormItem(tv)
						return
					}

					a.token = resp.Token
					a.tApp.SetRoot(a.createDataPage(), true)
				}
			},
		)

	p.AddPage("Form", form, true, true)

	return p
}

func (a *Application) createDataPage() *tview.Pages {
	p := pages("data")
	p.SetBorder(false)

	typeList := tview.NewList()
	typeList.
		SetBorder(true).
		SetTitle(" data type ")

	typesContent := map[string][]string{
		"passwords": {"pass1", "pass2"},
		"cards":     {"card1", "card2"},
	}

	dataList := tview.NewTextView()
	dataList.SetBorder(true)

	dataList.
		SetTitle(" passwords ")
	dataList.
		SetText(strings.Join(typesContent["passwords"], " - "))

	typeList.
		AddItem(
			"passwords", "", 0, func() {
				dataList.
					SetTitle(" passwords ")
				dataList.
					SetText(strings.Join(typesContent["passwords"], " - "))
			},
		).
		AddItem(
			"cards", "", 0, func() {
				dataList.
					SetTitle(" cards ")
				dataList.
					SetText(strings.Join(typesContent["cards"], " - "))
			},
		)

	actionsNote := tview.NewTextView()
	actionsNote.
		SetText("(Ctrl-A) Add new").
		SetTextColor(tcell.ColorGreen)

	grid := tview.NewGrid()
	grid.
		AddItem(typeList, 0, 0, 12, 1, 0, 0, true).
		AddItem(dataList, 0, 1, 12, 3, 0, 0, false).
		AddItem(actionsNote, 12, 0, 1, 4, 0, 0, false)

	typeList.
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyCtrlA {
					a.tApp.SetRoot(a.createAuthMenu(), true)
					return nil
				}

				return event
			},
		)

	p.
		AddPage("content", grid, true, true)

	return p
}

func Run(ctx context.Context) error {
	return NewApplication(ctx).Run()
}

func pages(title string) *tview.Pages {
	p := tview.NewPages()
	p.SetTitle("[green] Gophkeeper " + title + " ")
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
