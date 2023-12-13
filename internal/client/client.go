package client

import (
	"context"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/user"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
)

// Page common interface for all client pages
type Page interface {
	Render() *tview.Pages
}

// Application dependencies struct to be used in client application
type Application struct {
	tApp    *tview.Application
	uClient user.UserClient
	dClient datakeeper.DataKeeperClient
	ctx     context.Context
	storage Storage
}

// ApplicationConfig application config
type ApplicationConfig struct {
	Addr string
}

// Run starts client application and returns error if there was one
func Run(ctx context.Context, ac ApplicationConfig) error {
	return NewApplication(ctx, ac).run()
}

// run creates first page view (auth page) and renders it
func (a *Application) run() error {
	return a.tApp.SetRoot(NewAuthPage(a).Render(), true).Run()
}

// NewApplication returns pointer to new Application instance with created dependencies
func NewApplication(ctx context.Context, ac ApplicationConfig) *Application {
	tApp := tview.NewApplication()

	conn, connErr := grpc.Dial(ac.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if connErr != nil {
		log.Fatalln("grpc client error: ", connErr)
	}

	uClient := user.NewUserClient(conn)
	dClient := datakeeper.NewDataKeeperClient(conn)

	return &Application{
		tApp:    tApp,
		uClient: uClient,
		dClient: dClient,
		ctx:     ctx,
		storage: Storage{PasswordPairs: []storage.PasswordPair{}},
	}
}

// pages helper method to bootstrap common page props
func pages(title string) *tview.Pages {
	p := tview.NewPages()
	p.SetTitle("[green] Gophkeeper " + title + " ")
	p.SetBorder(true)

	return p
}

// addLoadingTextView creates text and runs mutations in goroutine for it to look like loader.
// ctx should be canceled to stop goroutine with mutations
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

// createErrorText creates text view with "error" styles
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
