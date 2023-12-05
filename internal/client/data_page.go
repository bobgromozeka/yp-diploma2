package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
)

type DataType = int

const (
	FocusType = iota
	FocusData
	FocusView
)

const (
	DataPasswordPair DataType = iota
	DataTexts
	DataCards
	DataBin
)

type DataPage struct {
	app             *Application
	page            *tview.Pages
	currentDataType DataType
	currentFocus    int
	typesList       *tview.List
	dataList        *tview.List
	dataGrid        *tview.Grid
	entityView      *tview.Grid
	escFunc         func()
	pageCtx         context.Context
	pageCancelCtx   context.CancelFunc
	fetchingDoneCh  chan struct{}
}

func NewDataPage(app *Application) *DataPage {
	ctx, cancel := context.WithCancel(app.ctx)
	dp := &DataPage{
		app:           app,
		currentFocus:  FocusType,
		pageCtx:       ctx,
		pageCancelCtx: cancel,
	}

	dp.escFunc = func() {
		switch dp.currentFocus {
		case FocusView:
			dp.setFocusData()
		case FocusData:
			dp.setFocusTypes()
		}
	}

	return dp
}

func (p *DataPage) Render() *tview.Pages {
	p.fetchingDoneCh = p.startFetchingData(p.pageCtx)

	return p.render()
}

func (p *DataPage) render() *tview.Pages {
	p.page = pages("data")

	p.page.SetBorder(false)
	grid := tview.NewGrid()
	p.dataGrid = grid

	p.page.
		AddPage("content", grid, true, true)

	typeList := tview.NewList()
	p.typesList = typeList

	typeList.
		SetBorder(true).
		SetTitle(" data type ")

	p.gridAddDataBlock(tview.NewBox().SetBorder(true))

	typeList.
		AddItem(
			"password pairs", "", 0, func() {
				p.renderPasswordPairsList()
			},
		).
		AddItem(
			"texts", "", 0, func() {
				p.renderTextsList()
			},
		)

	actionsNote := tview.NewTextView()
	actionsNote.
		SetText("(Ctrl-A) Add new    (Esc) focus previous    ").
		SetTextColor(tcell.ColorGreen)

	emptyGrid := tview.NewGrid()
	emptyGrid.SetBorder(true)
	p.entityView = emptyGrid

	grid.
		AddItem(typeList, 0, 0, 12, 1, 0, 0, true).
		AddItem(actionsNote, 12, 0, 1, 4, 0, 0, false).
		AddItem(emptyGrid, 0, 4, 12, 2, 0, 0, false)

	grid.
		SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyCtrlA {
					p.stopFetching()
					p.app.tApp.SetRoot(NewSelectCreationPage(p.app).Render(), true)
					return nil
				}
				if event.Key() == tcell.KeyESC {
					p.escFunc()
					return nil
				}

				return event
			},
		)

	return p.page
}

func (p *DataPage) startFetchingData(ctx context.Context) chan struct{} {
	doneCh := make(chan struct{})
	go func() {
		streamClient, err := p.app.dClient.GetData(ctx, &datakeeper.GetDataRequest{})
		if err != nil {
			log.Println("stream start error: ", err)
			return
		}

		for {
			data, dataErr := streamClient.Recv()
			select {
			case <-ctx.Done():
				select {
				case <-doneCh: // clear if not empty (for any strange reason)
				default:
				}
				doneCh <- struct{}{}
				return
			default:
			}
			if dataErr != nil && !errors.Is(dataErr, io.EOF) {
				p.attachError("Get data error. Contact me in Telegram @xxSerk d^_^b: " + dataErr.Error())
				time.Sleep(time.Second * 5)
				continue
			} else if errors.Is(dataErr, io.EOF) {
				return
			} else {
				p.app.storage.PasswordPairs = mapGRPCPairsToStorage(data.PasswordPairs)
				p.app.storage.Texts = mapGRPCTextsToStorage(data.Texts)
				p.app.tApp.QueueUpdateDraw(func() {
					p.rerenderDataBlock()
					p.clearEntityView()
				})
			}
		}
	}()

	return doneCh
}

func (p *DataPage) renderPasswordPairsList() {
	if p.dataList != nil {
		p.dataGrid.RemoveItem(p.dataList)
	}

	list := tview.NewList()
	list.SetBorder(true)
	list.
		SetTitle(" Password pairs ")

	for _, pp := range p.app.storage.PasswordPairs {
		id := pp.ID
		login := pp.Login
		pass := pp.Password
		desc := pp.Description
		list.
			AddItem(pp.Login, "", 0, func() {
				p.renderPasswordPairsView(id, login, pass, desc)
			})
	}

	p.dataList = list
	p.currentDataType = DataPasswordPair
	p.gridAddDataBlock(p.dataList)

	p.setFocusData()
}

func (p *DataPage) renderTextsList() {
	if p.dataList != nil {
		p.dataGrid.RemoveItem(p.dataList)
	}

	list := tview.NewList()
	list.SetBorder(true)
	list.
		SetTitle(" Texts ")

	for _, t := range p.app.storage.Texts {
		text := t.T
		if len(text) > 50 {
			text = t.T[:50] + "..."
		}
		name := t.Name
		desc := t.Description
		list.
			AddItem(t.Name, text, 0, func() {
				p.renderTextsView(name, t.T, desc)
			})
	}

	p.dataList = list
	p.currentDataType = DataTexts
	p.gridAddDataBlock(list)

	p.setFocusData()
}

func (p *DataPage) gridAddDataBlock(prim tview.Primitive) {
	p.dataGrid.AddItem(prim, 0, 1, 12, 3, 0, 0, false)
}

func (p *DataPage) rerenderDataBlock() {
	switch p.currentDataType {
	case DataPasswordPair:
		p.renderPasswordPairsList()
	case DataTexts:
		p.renderTextsList()
	}
}

func (p *DataPage) attachError(text string) {
	err := tview.NewTextView()
	err.SetTextColor(tcell.ColorRed)
	err.SetText(text)

	p.dataGrid.AddItem(err, 13, 0, 1, 4, 0, 0, false)
}

func (p *DataPage) detachError() {

}

func (p *DataPage) renderPasswordPairsView(id int, login, password string, description *string) {
	g := tview.NewGrid()
	g.SetBorder(true)
	g.SetTitle(" Password ")

	loginLabel := tview.NewTextView()
	loginLabel.SetText("Login")

	loginContent := tview.NewTextView()
	loginContent.SetText(login)
	loginContent.SetBackgroundColor(tcell.ColorBlue)

	passwordLabel := tview.NewTextView()
	passwordLabel.SetText("Password")

	passwordContent := tview.NewTextView()
	passwordContent.SetText(password)
	passwordContent.SetBackgroundColor(tcell.ColorBlue)

	g.
		AddItem(loginLabel, 0, 0, 1, 1, 0, 0, false).
		AddItem(loginContent, 0, 1, 1, 3, 0, 0, false).
		AddItem(passwordLabel, 1, 0, 1, 1, 0, 0, false).
		AddItem(passwordContent, 1, 1, 1, 3, 0, 0, false)

	deleteButton := tview.NewButton("Remove")
	deleteButton.
		SetSelectedFunc(func() {
			confirmationModal := tview.NewModal()
			confirmationModal.
				SetText(fmt.Sprintf("Confirm delete of password pair with login %s ?", login)).
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Yes" {
						_, err := p.app.dClient.RemovePasswordPair(p.app.ctx, &datakeeper.RemovePasswordPairRequest{
							ID: int32(id),
						})
						if err != nil {
							p.attachError("Removing entity error. Contact me in Telegram @xxSerk d^_^b: " + err.Error())
						}

					}
					p.removeConfirmationModal()
					p.setFocusView()
				})
			p.addConfirmationModal(confirmationModal)
		})

	descriptionContent := tview.NewTextView()
	if description != nil {
		descriptionLabel := tview.NewTextView()
		descriptionLabel.SetText("Description (scrollable)")

		descriptionContent.SetText(*description)
		descriptionContent.SetBackgroundColor(tcell.ColorBlue)
		descriptionContent.SetScrollable(true)
		descriptionContent.SetFocusFunc(func() {
			descriptionLabel.SetTextColor(tcell.ColorGreen)
		})
		descriptionContent.SetBlurFunc(func() {
			descriptionLabel.SetTextColor(tcell.ColorWhite)
		})

		g.AddItem(descriptionLabel, 2, 0, 1, 1, 0, 0, false)
		g.AddItem(descriptionContent, 2, 1, 1, 3, 0, 0, true)

		descriptionContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(deleteButton)
				return nil
			}

			return event
		})
	}

	var focusButton bool
	if description == nil {
		focusButton = true
	} else {
		deleteButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(descriptionContent)
				return nil
			}

			return event
		})
	}

	g.AddItem(deleteButton, 3, 1, 1, 1, 0, 0, focusButton)

	g.SetRows(1, 1, 15, 1, 0)
	g.SetGap(1, 0)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = g
	p.dataGrid.AddItem(g, 0, 4, 12, 2, 0, 0, false)
	p.setFocusView()
}

func (p *DataPage) renderTextsView(name, text string, description *string) {
	g := tview.NewGrid()
	g.SetBorder(true)
	g.SetTitle(" Text ")

	nameLabel := tview.NewTextView()
	nameLabel.SetText("Login")

	nameContent := tview.NewTextView()
	nameContent.SetText(name)
	nameContent.SetBackgroundColor(tcell.ColorBlue)

	textLabel := tview.NewTextView()
	textLabel.SetText("Text (scrollable)")

	textContent := tview.NewTextView()
	textContent.SetText(text)
	textContent.SetBackgroundColor(tcell.ColorBlue)
	textContent.SetScrollable(true)
	textContent.SetFocusFunc(func() {
		textLabel.SetTextColor(tcell.ColorGreen)
	})
	textContent.SetBlurFunc(func() {
		textLabel.SetTextColor(tcell.ColorWhite)
	})

	g.
		AddItem(nameLabel, 0, 0, 1, 1, 0, 0, false).
		AddItem(nameContent, 0, 1, 1, 3, 0, 0, false).
		AddItem(textLabel, 1, 0, 1, 1, 0, 0, false).
		AddItem(textContent, 1, 1, 1, 3, 0, 0, true)

	if description != nil {
		descriptionLabel := tview.NewTextView()
		descriptionLabel.SetText("Description (scrollable)")

		descriptionContent := tview.NewTextView()
		descriptionContent.SetText(*description)
		descriptionContent.SetBackgroundColor(tcell.ColorBlue)
		descriptionContent.SetScrollable(true)
		descriptionContent.SetFocusFunc(func() {
			descriptionLabel.SetTextColor(tcell.ColorGreen)
		})
		descriptionContent.SetBlurFunc(func() {
			descriptionLabel.SetTextColor(tcell.ColorWhite)
		})

		g.AddItem(descriptionLabel, 2, 0, 1, 1, 0, 0, false)
		g.AddItem(descriptionContent, 2, 1, 1, 3, 0, 0, false)

		textContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(descriptionContent)
				return nil
			}

			return event
		})
		descriptionContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(textContent)
				return nil
			}

			return event
		})
		descriptionContent.SetFocusFunc(func() {
			descriptionLabel.SetTextColor(tcell.ColorGreen)
		})
		descriptionContent.SetBlurFunc(func() {
			descriptionLabel.SetTextColor(tcell.ColorWhite)
		})
	}

	g.SetRows(1, 15, 15, 0)
	g.SetGap(1, 0)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = g
	p.dataGrid.AddItem(g, 0, 4, 12, 2, 0, 0, false)
	p.setFocusView()
}

func (p *DataPage) setFocusData() {
	p.currentFocus = FocusData
	p.app.tApp.SetFocus(p.dataList)
}

func (p *DataPage) setFocusView() {
	p.currentFocus = FocusView
	p.app.tApp.SetFocus(p.entityView)
}

func (p *DataPage) setFocusTypes() {
	p.currentFocus = FocusType
	p.app.tApp.SetFocus(p.typesList)
}

func (p *DataPage) stopFetching() {
	p.pageCancelCtx()
	<-p.fetchingDoneCh
}

func (p *DataPage) clearEntityView() {
	emptyGrid := tview.NewGrid()
	emptyGrid.SetBorder(true)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = emptyGrid
	p.dataGrid.AddItem(emptyGrid, 0, 4, 12, 2, 0, 0, false)
}

func (p *DataPage) addConfirmationModal(m *tview.Modal) {
	p.page.AddPage("conf_modal", m, false, true)
}

func (p *DataPage) removeConfirmationModal() {
	p.page.RemovePage("conf_modal")
}
