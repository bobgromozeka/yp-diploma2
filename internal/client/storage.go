package client

import (
	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
)

// Storage client storage structure
type Storage struct {
	PasswordPairs []storage.PasswordPair
	Texts         []storage.Text
	Cards         []storage.Card
	Bins          []storage.Bin
}

// mapGRPCPairsToStorage maps password pairs grpc type to storage type
func mapGRPCPairsToStorage(gpp []*datakeeper.PasswordPair) []storage.PasswordPair {
	var spp []storage.PasswordPair

	if gpp == nil {
		return spp
	}

	spp = make([]storage.PasswordPair, len(gpp))

	for i, pp := range gpp {
		spp[i] = storage.PasswordPair{
			ID:          int(pp.ID),
			Login:       pp.Login,
			Password:    pp.Password,
			Description: pp.Description,
		}
	}

	return spp
}

// mapGRPCTextsToStorage maps texts grpc type to storage type
func mapGRPCTextsToStorage(gt []*datakeeper.Text) []storage.Text {
	var st []storage.Text

	if gt == nil {
		return st
	}

	st = make([]storage.Text, len(gt))

	for i, t := range gt {
		st[i] = storage.Text{
			ID:          int(t.ID),
			Name:        t.Name,
			T:           t.Text,
			Description: t.Description,
		}
	}

	return st
}

// mapGRPCCardsToStorage maps cards grpc type to storage type
func mapGRPCCardsToStorage(gc []*datakeeper.Card) []storage.Card {
	var sc []storage.Card

	if gc == nil {
		return sc
	}

	sc = make([]storage.Card, len(gc))

	for i, c := range gc {
		sc[i] = storage.Card{
			ID:                int(c.ID),
			Name:              c.Name,
			Number:            c.Number,
			ValidThroughMonth: int(c.ValidThroughMonth),
			ValidThroughYear:  int(c.ValidThroughYear),
			CVV:               int(c.Cvv),
			Description:       c.Description,
		}
	}

	return sc
}

// mapGRPCBinsToStorage maps binaries grpc type to storage type
func mapGRPCBinsToStorage(gb []*datakeeper.Bin) []storage.Bin {
	var sb []storage.Bin

	if gb == nil {
		return sb
	}

	sb = make([]storage.Bin, len(gb))

	for i, b := range gb {
		sb[i] = storage.Bin{
			ID:          int(b.ID),
			Name:        b.Name,
			Data:        b.Data,
			Description: b.Description,
		}
	}

	return sb
}
