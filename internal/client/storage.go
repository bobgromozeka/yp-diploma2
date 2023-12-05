package client

import (
	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
)

type Storage struct {
	PasswordPairs []storage.PasswordPair
	Texts         []storage.Text
}

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
