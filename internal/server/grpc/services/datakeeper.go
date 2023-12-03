package services

import (
	"context"
	"fmt"
	"time"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc/interceptors"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
	"github.com/bobgromozeka/yp-diploma2/pkg/helpers/goroutines"
)

type DataKeeperService struct {
	datakeeper.UnimplementedDataKeeperServer
	dks storage.DataKeeperStorage
}

func NewDataKeeperService(dks storage.DataKeeperStorage) *DataKeeperService {
	return &DataKeeperService{
		dks: dks,
	}
}

func (s *DataKeeperService) CreatePasswordPair(ctx context.Context, req *datakeeper.CreatePasswordPairRequest) (*datakeeper.EmptyResponse, error) {
	userID, ok := ctx.Value(interceptors.UserID).(float64)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.CreatePasswordPair(ctx, int(userID), req.Login, req.Password, req.Description)
	if err != nil {
		fmt.Println("Error during creating password pair: ", err)
		return nil, ErrInternalServerError
	}

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) GetData(req *datakeeper.GetDataRequest, serv datakeeper.DataKeeper_GetDataServer) error {
	userID, ok := serv.Context().Value(interceptors.UserID).(float64)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t", serv.Context().Value(interceptors.UserID))
		return ErrInternalServerError
	}

	var passwordPairsResult []*datakeeper.PasswordPair
	var textResults []*datakeeper.Text

	errs := goroutines.WaitForAll(
		func() error {
			passwordPairs, err := s.dks.GetAllPasswordPairs(serv.Context(), int(userID))
			if err != nil {
				return err
			}

			passwordPairsResult = mapStoragePairsToGRPC(passwordPairs)

			return nil
		},
		func() error {
			texts, err := s.dks.GetAllTexts(serv.Context(), int(userID))
			if err != nil {
				return err
			}

			textResults = mapStorageTextsToGRPC(texts)

			return nil
		},
	)

	if len(errs) > 0 {
		fmt.Println("Error during getting data for stream: ", errs)
		return ErrInternalServerError
	}

	for {
		sendErr := serv.Send(&datakeeper.GetDataResponse{
			PasswordPairs: passwordPairsResult,
			Texts:         textResults,
		})
		if sendErr != nil {
			fmt.Println("Stream send error: ", sendErr)
			return ErrInternalServerError
		}
		select {
		case <-serv.Context().Done():
			return nil
		case <-time.NewTicker(time.Hour).C: // just to let connection be opened (will be cleared after subscriptions)
		}
	}

	return nil
}

func mapStoragePairsToGRPC(spp []storage.PasswordPair) []*datakeeper.PasswordPair {
	gpp := make([]*datakeeper.PasswordPair, len(spp))

	for i, pp := range spp {
		gpp[i] = &datakeeper.PasswordPair{
			Login:       pp.Login,
			Password:    pp.Password,
			Description: pp.Description,
		}
	}

	return gpp
}

func mapStorageTextsToGRPC(st []storage.Text) []*datakeeper.Text {
	gt := make([]*datakeeper.Text, len(st))

	for i, t := range st {
		gt[i] = &datakeeper.Text{
			Name:        t.Name,
			Text:        t.T,
			Description: t.Description,
		}
	}

	return gt
}
