package client

import (
	"strconv"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
)

// SelectCreationPage page with list of new entities
type SelectCreationPage struct {
	app *Application
}

// NewSelectCreationPage returns pointer to SelectCreationPage
func NewSelectCreationPage(app *Application) *SelectCreationPage {
	return &SelectCreationPage{
		app: app,
	}
}

// Render adds all needed items to page and returns tview component
func (p *SelectCreationPage) Render() *tview.Pages {
	page := pages("create selection")

	list := tview.NewList()
	list.
		AddItem("Create password pair", "", 'p', func() {
			p.app.tApp.SetRoot(NewCreatePasswordPairPage(p.app).Render(), true)
		}).
		AddItem("Create text", "", 't', func() {
			p.app.tApp.SetRoot(NewCreateTextPage(p.app).Render(), true)
		}).
		AddItem("Create card", "", 'c', func() {
			p.app.tApp.SetRoot(NewCreateCardPage(p.app).Render(), true)
		}).
		AddItem("Create bin", "", 'b', func() {
			p.app.tApp.SetRoot(NewCreateBinPage(p.app).Render(), true)
		})

	page.
		AddPage("list", list, true, true)

	return page
}

// CreatePasswordPairPage page with form to create password pair
type CreatePasswordPairPage struct {
	app *Application
}

// NewCreatePasswordPairPage returns pointer to CreatePasswordPairPage
func NewCreatePasswordPairPage(app *Application) *CreatePasswordPairPage {
	return &CreatePasswordPairPage{
		app: app,
	}
}

// Render creates password pair page and returns tview component
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

// CreateTextPage page with form to create text
type CreateTextPage struct {
	app *Application
}

// NewCreateTextPage returns pointer to CreateTextPage
func NewCreateTextPage(app *Application) *CreateTextPage {
	return &CreateTextPage{
		app: app,
	}
}

// Render creates text page and return tview component
func (p *CreateTextPage) Render() *tview.Pages {
	page := pages("create text")

	var name string
	var text string
	var description string

	form := tview.NewForm()

	form.
		AddInputField("Name", "", 100, nil, func(text string) {
			name = text
		}).
		AddTextArea("Text", "", 100, 10, 0, func(t string) {
			text = t
		}).
		AddTextArea("Description", "", 100, 10, 0, func(text string) {
			description = text
		}).
		AddTextView("ESC to return back", "", 100, 1, false, false).
		AddButton("Confirm", func() {
			_, err := p.app.dClient.CreateText(p.app.ctx, &datakeeper.CreateTextRequest{
				Name:        name,
				Text:        text,
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
				SetText("Text was successfully added").
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

// CreateCardPage page with form to create new card
type CreateCardPage struct {
	app *Application
}

// NewCreateCardPage return pointer to CreateCardPage
func NewCreateCardPage(app *Application) *CreateCardPage {
	return &CreateCardPage{
		app: app,
	}
}

// Render creates card page and return tview component
func (p *CreateCardPage) Render() *tview.Pages {
	page := pages("create card")

	var name string
	var number string
	var vtm int
	var vty int
	var cvv int
	var description string

	form := tview.NewForm()

	form.
		AddInputField("Name", "", 100, nil, func(text string) {
			name = text
		}).
		AddInputField("Number", "", 100, func(textToCheck string, lastChar rune) bool {
			return unicode.IsDigit(lastChar) && len(textToCheck) < 16
		}, func(t string) {
			number = t
		}).
		AddInputField("Month", "", 10, func(textToCheck string, lastChar rune) bool {
			if !unicode.IsDigit(lastChar) {
				return false
			}

			num, err := strconv.Atoi(textToCheck)
			if err != nil {
				return false
			}

			return num >= 1 && num <= 12
		}, func(text string) {
			num, err := strconv.Atoi(text)
			if err == nil {
				vtm = num
			}
		}).
		AddInputField("Year", "", 10, func(textToCheck string, lastChar rune) bool {
			if !unicode.IsDigit(lastChar) {
				return false
			}

			num, err := strconv.Atoi(textToCheck)
			if err != nil {
				return false
			}

			return num >= 0 && num <= 99
		}, func(text string) {
			num, err := strconv.Atoi(text)
			if err == nil {
				vty = num
			}
		}).
		AddInputField("CVV", "", 15, func(textToCheck string, lastChar rune) bool {
			if !unicode.IsDigit(lastChar) {
				return false
			}

			num, err := strconv.Atoi(textToCheck)
			if err != nil {
				return false
			}

			return num <= 999
		}, func(text string) {
			num, err := strconv.Atoi(text)
			if err == nil {
				cvv = num
			}
		}).
		AddTextArea("Description", "", 100, 10, 0, func(text string) {
			description = text
		}).
		AddTextView("ESC to return back", "", 100, 1, false, false).
		AddButton("Confirm", func() {
			_, err := p.app.dClient.CreateCard(p.app.ctx, &datakeeper.CreateCardRequest{
				Name:              name,
				Number:            number,
				ValidThroughMonth: uint32(vtm),
				ValidThroughYear:  uint32(vty),
				Cvv:               uint32(cvv),
				Description:       &description,
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
				SetText("Card was successfully added").
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

// CreateBinPage page with form to create entry with binary data
type CreateBinPage struct {
	app *Application
}

// NewCreateBinPage returns pointer to CreateBinPage
func NewCreateBinPage(app *Application) *CreateBinPage {
	return &CreateBinPage{
		app: app,
	}
}

// Render creates new bin page and returns tview component
func (p *CreateBinPage) Render() *tview.Pages {
	page := pages("create bin")

	var name string
	var data []byte
	var description string

	form := tview.NewForm()

	form.
		AddInputField("Name", "", 100, nil, func(text string) {
			name = text
		}).
		AddTextArea("Data", "", 100, 10, 0, func(text string) {
			data = []byte(text)
		}).
		AddTextArea("Description", "", 100, 10, 0, func(text string) {
			description = text
		}).
		AddTextView("ESC to return back", "", 100, 1, false, false).
		AddButton("Confirm", func() {
			_, err := p.app.dClient.CreateBin(p.app.ctx, &datakeeper.CreateBinRequest{
				Name:        name,
				Data:        data,
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
				SetText("Text was successfully added").
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
