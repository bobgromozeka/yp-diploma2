package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/pkg/helpers"
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

// DataPage page with all views to navigate user data.
// Contains all tview components that need to be mutated to implement usable interface
// ===============================
// |         |        |          |
// |         |        |          |
// |         |        |          |
// |typesList|dataList|entityView|
// |         |        |          |
// |         |        |          |
// |         |        |          |
// ===============================
//
//	actions menu
//
// ===============================
type DataPage struct {
	// app client application dependencies structure
	app *Application

	// page current page tview component
	page *tview.Pages

	// currentDataType data type that was selected
	currentDataType DataType

	// currentFocus page part that was focused
	currentFocus int

	// typesList list with all data types
	typesList *tview.List

	// dataList list with all entities of selected type
	dataList *tview.List

	// dataGrid entityView container
	dataGrid *tview.Grid

	// entityView grid with selected entity fields
	entityView *tview.Grid

	// escFunc function that will be executed when Esc was pressed
	escFunc func()

	// pageCtx current page ctx to control concurrent processes
	pageCtx context.Context

	// pageCancelCtx pageCtx cancel func
	pageCancelCtx  context.CancelFunc
	fetchingDoneCh chan struct{}
}

// NewDataPage returns pointer to DataPage
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

// Render creates data page and returns tview component.
// Starts concurrent process of stream connection with server to receive data.
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
		).
		AddItem(
			"cards", "", 0, func() {
				p.renderCardsList()
			},
		).
		AddItem(
			"bins", "", 0, func() {
				p.renderBinsList()
			},
		)

	actionsNote := tview.NewTextView()
	actionsNote.
		SetText("(Ctrl-A) Add new    (Esc) focus previous    (Tab) focus next form field").
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

// startFetchingData server connection logic.
// Re-renders opened entities list after receiving new data.
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
				p.app.storage.Cards = mapGRPCCardsToStorage(data.Cards)
				p.app.storage.Bins = mapGRPCBinsToStorage(data.Bins)
				p.app.tApp.QueueUpdateDraw(func() {
					p.rerenderDataBlock()
					p.clearEntityView()
				})
			}
		}
	}()

	return doneCh
}

// renderPasswordPairsList method to render list of password pairs into data block
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

// renderTextsList renders list of texts into data block
func (p *DataPage) renderTextsList() {
	if p.dataList != nil {
		p.dataGrid.RemoveItem(p.dataList)
	}

	list := tview.NewList()
	list.SetBorder(true)
	list.
		SetTitle(" Texts ")

	for _, t := range p.app.storage.Texts {
		id := t.ID
		text := t.T
		shortText := text
		if len(text) > 50 {
			shortText = text[:50] + "..."
		}
		name := t.Name
		desc := t.Description
		list.
			AddItem(t.Name, shortText, 0, func() {
				p.renderTextsView(id, name, text, desc)
			})
	}

	p.dataList = list
	p.currentDataType = DataTexts
	p.gridAddDataBlock(list)

	p.setFocusData()
}

// renderCardsList renders cards list into data block
func (p *DataPage) renderCardsList() {
	if p.dataList != nil {
		p.dataGrid.RemoveItem(p.dataList)
	}

	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle(" Cards ")

	for _, c := range p.app.storage.Cards {
		card := c
		list.AddItem(c.Name, "", 0, func() {
			p.renderCardsView(card.ID, card.Name, card.Number, card.ValidThroughMonth, card.ValidThroughYear, card.CVV, card.Description)
		})
	}

	p.dataList = list
	p.currentDataType = DataCards
	p.gridAddDataBlock(list)

	p.setFocusData()
}

// renderBinsList render binary list into data block
func (p *DataPage) renderBinsList() {
	if p.dataList != nil {
		p.dataGrid.RemoveItem(p.dataList)
	}

	list := tview.NewList()
	list.SetBorder(true)
	list.SetTitle(" Bins ")

	for _, b := range p.app.storage.Bins {
		bin := b
		list.AddItem(b.Name, "", 0, func() {
			p.renderBinsView(bin.ID, bin.Name, bin.Data, bin.Description)
		})
	}

	p.dataList = list
	p.currentDataType = DataBin
	p.gridAddDataBlock(list)

	p.setFocusData()
}

// gridAddDataBlock adds primitive into dataGrid
func (p *DataPage) gridAddDataBlock(prim tview.Primitive) {
	p.dataGrid.AddItem(prim, 0, 1, 12, 3, 0, 0, false)
}

// rerenderDataBlock render currently active data block with new data
func (p *DataPage) rerenderDataBlock() {
	switch p.currentDataType {
	case DataPasswordPair:
		p.renderPasswordPairsList()
	case DataTexts:
		p.renderTextsList()
	case DataCards:
		p.renderCardsList()
	case DataBin:
		p.renderBinsList()
	}
}

// attachError creates error with specified text and adds it under actions menu of the page
func (p *DataPage) attachError(text string) {
	err := tview.NewTextView()
	err.SetTextColor(tcell.ColorRed)
	err.SetText(text)

	p.dataGrid.AddItem(err, 13, 0, 1, 4, 0, 0, false)
}

// renderPasswordPairsView render password pair view into entityView
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

// renderTextsView render texts view into entityView
func (p *DataPage) renderTextsView(id int, name, text string, description *string) {
	g := tview.NewGrid()
	g.SetBorder(true)
	g.SetTitle(" Text ")

	nameLabel := tview.NewTextView()
	nameLabel.SetText("Name")

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

	deleteButton := tview.NewButton("Remove")
	deleteButton.
		SetSelectedFunc(func() {
			confirmationModal := tview.NewModal()
			confirmationModal.
				SetText(fmt.Sprintf("Confirm delete of text with name %s ?", name)).
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Yes" {
						_, err := p.app.dClient.RemoveText(p.app.ctx, &datakeeper.RemoveTextRequest{
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
				p.app.tApp.SetFocus(deleteButton)
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

	var focusButton bool
	if description == nil {
		focusButton = true
	} else {
		deleteButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(textContent)
				return nil
			}

			return event
		})
	}

	g.AddItem(deleteButton, 3, 1, 1, 1, 0, 0, focusButton)

	g.SetRows(1, 15, 15, 1, 0)
	g.SetGap(1, 0)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = g
	p.dataGrid.AddItem(g, 0, 4, 12, 2, 0, 0, false)
	p.setFocusView()
}

// renderCardsView render cards view into entityView
func (p *DataPage) renderCardsView(id int, name, number string, validThroughMonth, validThroughYear, cvv int, description *string) {
	g := tview.NewGrid()
	g.SetBorder(true)
	g.SetTitle(" Text ")

	nameLabel := tview.NewTextView()
	nameLabel.SetText("Name")

	nameContent := tview.NewTextView()
	nameContent.SetText(name)
	nameContent.SetBackgroundColor(tcell.ColorBlue)

	numberLabel := tview.NewTextView()
	numberLabel.SetText("Number")

	numberContent := tview.NewTextView()
	numberContent.SetText(number)
	numberContent.SetBackgroundColor(tcell.ColorBlue)

	vtmLabel := tview.NewTextView()
	vtmLabel.SetText("     M")

	vtmContent := tview.NewTextView()
	vtmContent.SetText(helpers.PadLeft(strconv.Itoa(validThroughMonth), '0', 2))
	vtmContent.SetBackgroundColor(tcell.ColorBlue)

	vtyLabel := tview.NewTextView()
	vtyLabel.SetText("     Y")

	vtyContent := tview.NewTextView()
	vtyContent.SetText(strconv.Itoa(validThroughYear))
	vtyContent.SetBackgroundColor(tcell.ColorBlue)

	g.
		AddItem(nameLabel, 0, 0, 1, 1, 0, 0, false).
		AddItem(nameContent, 0, 1, 1, 3, 0, 0, false).
		AddItem(numberLabel, 1, 0, 1, 1, 0, 0, false).
		AddItem(numberContent, 1, 1, 1, 3, 0, 0, false).
		AddItem(vtmLabel, 2, 0, 1, 1, 0, 0, false).
		AddItem(vtmContent, 2, 1, 1, 1, 0, 0, false).
		AddItem(vtyLabel, 2, 2, 1, 1, 0, 0, false).
		AddItem(vtyContent, 2, 3, 1, 1, 0, 0, false)

	deleteButton := tview.NewButton("Remove")
	deleteButton.
		SetSelectedFunc(func() {
			confirmationModal := tview.NewModal()
			confirmationModal.
				SetText(fmt.Sprintf("Confirm delete of card with name %s ?", name)).
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Yes" {
						_, err := p.app.dClient.RemoveCard(p.app.ctx, &datakeeper.RemoveCardRequest{
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

		g.AddItem(descriptionLabel, 3, 0, 1, 1, 0, 0, false)
		g.AddItem(descriptionContent, 3, 1, 1, 3, 0, 0, true)

		descriptionContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(deleteButton)
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

	g.AddItem(deleteButton, 4, 1, 1, 1, 0, 0, focusButton)

	g.SetRows(1, 1, 1, 15, 1, 0)
	g.SetGap(1, 0)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = g
	p.dataGrid.AddItem(g, 0, 4, 12, 2, 0, 0, false)
	p.setFocusView()
}

// renderBinsView render binary view into entityView
func (p *DataPage) renderBinsView(id int, name string, data []byte, description *string) {
	g := tview.NewGrid()
	g.SetBorder(true)
	g.SetTitle(" Text ")

	nameLabel := tview.NewTextView()
	nameLabel.SetText("Name")

	nameContent := tview.NewTextView()
	nameContent.SetText(name)
	nameContent.SetBackgroundColor(tcell.ColorBlue)

	dataLabel := tview.NewTextView()
	dataLabel.SetText("Data (scrollable)")

	dataContent := tview.NewTextView()
	dataContent.SetText(string(data))
	dataContent.SetBackgroundColor(tcell.ColorBlue)
	dataContent.SetScrollable(true)
	dataContent.SetFocusFunc(func() {
		dataLabel.SetTextColor(tcell.ColorGreen)
	})
	dataContent.SetBlurFunc(func() {
		dataLabel.SetTextColor(tcell.ColorWhite)
	})

	g.
		AddItem(nameLabel, 0, 0, 1, 1, 0, 0, false).
		AddItem(nameContent, 0, 1, 1, 3, 0, 0, false).
		AddItem(dataLabel, 1, 0, 1, 1, 0, 0, false).
		AddItem(dataContent, 1, 1, 1, 3, 0, 0, true)

	deleteButton := tview.NewButton("Remove")
	deleteButton.
		SetSelectedFunc(func() {
			confirmationModal := tview.NewModal()
			confirmationModal.
				SetText(fmt.Sprintf("Confirm delete of bin with name %s ?", name)).
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Yes" {
						_, err := p.app.dClient.RemoveBin(p.app.ctx, &datakeeper.RemoveBinRequest{
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
		g.AddItem(descriptionContent, 2, 1, 1, 3, 0, 0, false)

		dataContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(descriptionContent)
				return nil
			}

			return event
		})
		descriptionContent.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(deleteButton)
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

	var focusButton bool
	if description == nil {
		focusButton = true
	} else {
		deleteButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				p.app.tApp.SetFocus(dataContent)
				return nil
			}

			return event
		})
	}

	g.AddItem(deleteButton, 3, 1, 1, 1, 0, 0, focusButton)

	g.SetRows(1, 15, 15, 1, 0)
	g.SetGap(1, 0)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = g
	p.dataGrid.AddItem(g, 0, 4, 12, 2, 0, 0, false)
	p.setFocusView()
}

// setFocusData sets current focus and app focus to dataList
func (p *DataPage) setFocusData() {
	p.currentFocus = FocusData
	p.app.tApp.SetFocus(p.dataList)
}

// setFocusView sets current focus and app focus to entityView
func (p *DataPage) setFocusView() {
	p.currentFocus = FocusView
	p.app.tApp.SetFocus(p.entityView)
}

// setFocusTypes sets current focus and app focus to typesList
func (p *DataPage) setFocusTypes() {
	p.currentFocus = FocusType
	p.app.tApp.SetFocus(p.typesList)
}

// stopFetching cancels page context and waits for data fetching stop signal to prevent leaks
func (p *DataPage) stopFetching() {
	p.pageCancelCtx()
	<-p.fetchingDoneCh
}

// clearEntityView clears entityView
func (p *DataPage) clearEntityView() {
	emptyGrid := tview.NewGrid()
	emptyGrid.SetBorder(true)

	p.dataGrid.RemoveItem(p.entityView)
	p.entityView = emptyGrid
	p.dataGrid.AddItem(emptyGrid, 0, 4, 12, 2, 0, 0, false)
}

// addConfirmationModal adds specified modal as named modal
func (p *DataPage) addConfirmationModal(m *tview.Modal) {
	p.page.AddPage("conf_modal", m, false, true)
}

// removeConfirmationModal removes named modal
func (p *DataPage) removeConfirmationModal() {
	p.page.RemovePage("conf_modal")
}
