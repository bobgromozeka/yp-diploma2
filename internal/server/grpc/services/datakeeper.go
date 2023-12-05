package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/server/grpc/interceptors"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
	"github.com/bobgromozeka/yp-diploma2/pkg/helpers/goroutines"
)

type subscription struct {
	stream  datakeeper.DataKeeper_GetDataServer
	errChan chan error
}

type userSubscription struct {
	notifyChan    chan struct{}
	subs          map[string]subscription
	lastData      *datakeeper.GetDataResponse
	cancelUserSub context.CancelFunc
}

type DataKeeperService struct {
	datakeeper.UnimplementedDataKeeperServer
	dks                  storage.DataKeeperStorage
	getDataSubscriptions map[int]userSubscription
	subscriptionsMx      sync.RWMutex
}

func NewDataKeeperService(dks storage.DataKeeperStorage) *DataKeeperService {
	return &DataKeeperService{
		dks:                  dks,
		getDataSubscriptions: map[int]userSubscription{},
		subscriptionsMx:      sync.RWMutex{},
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

	s.notifyUserSubs(int(userID))

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) GetData(req *datakeeper.GetDataRequest, serv datakeeper.DataKeeper_GetDataServer) error {
	uID, ok := serv.Context().Value(interceptors.UserID).(float64)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t", serv.Context().Value(interceptors.UserID))
		return ErrInternalServerError
	}

	subErr, errChan, subUUID := s.subscribeForUpdates(int(uID), serv)
	if subErr != nil {
		return subErr
	}

	select {
	case <-serv.Context().Done():
		s.unsubscribeFromUpdates(int(uID), subUUID)
		return nil
	case err := <-errChan:
		s.unsubscribeFromUpdates(int(uID), subUUID)
		return err
	}
}

func (s *DataKeeperService) RemovePasswordPair(ctx context.Context, req *datakeeper.RemovePasswordPairRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(float64)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.RemovePasswordPair(ctx, int(uID), int(req.ID))
	if err != nil {
		log.Default().Println("Error during removing password pair: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(int(uID))

	return &datakeeper.EmptyResponse{}, nil
}

func mapStoragePairsToGRPC(spp []storage.PasswordPair) []*datakeeper.PasswordPair {
	gpp := make([]*datakeeper.PasswordPair, len(spp))

	for i, pp := range spp {
		gpp[i] = &datakeeper.PasswordPair{
			ID:          int32(pp.ID),
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
			ID:          int32(t.ID),
			Name:        t.Name,
			Text:        t.T,
			Description: t.Description,
		}
	}

	return gt
}

func (s *DataKeeperService) subscribeForUpdates(userID int, stream datakeeper.DataKeeper_GetDataServer) (error, chan error, string) {
	s.subscriptionsMx.Lock()
	defer s.subscriptionsMx.Unlock()

	errChan := make(chan error)
	subUUID := uuid.New().String()

	if _, ok := s.getDataSubscriptions[userID]; !ok {
		userData, errs := s.getUserData(stream.Context(), userID)
		if errs != nil && len(errs) > 0 {
			log.Default().Printf("Errors during getting user %d data: %v", userID, errs)
			return ErrInternalServerError, nil, ""
		}

		ctx, cancel := context.WithCancel(context.Background())
		notifyChan := make(chan struct{}, 1)

		s.getDataSubscriptions[userID] = userSubscription{
			notifyChan:    notifyChan,
			subs:          map[string]subscription{},
			lastData:      userData,
			cancelUserSub: cancel,
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-notifyChan:
					s.subscriptionsMx.RLock()

					userData, errs := s.getUserData(ctx, userID)
					if errs != nil && len(errs) > 0 {
						for _, sub := range s.getDataSubscriptions[userID].subs {
							sub.errChan <- ErrInternalServerError
						}
						return
					}
					uSub := s.getDataSubscriptions[userID]
					uSub.lastData = userData
					s.getDataSubscriptions[userID] = uSub
					for _, sub := range s.getDataSubscriptions[userID].subs {
						_ = sub.stream.Send(userData)
					}

					s.subscriptionsMx.RUnlock()
				}
			}
		}()
	}

	sendErr := stream.Send(s.getDataSubscriptions[userID].lastData)
	if sendErr != nil {
		return ErrInternalServerError, nil, ""
	}

	subs := s.getDataSubscriptions[userID].subs
	subs[subUUID] = subscription{
		stream:  stream,
		errChan: errChan,
	}

	uSub := s.getDataSubscriptions[userID]
	uSub.subs = subs
	s.getDataSubscriptions[userID] = uSub

	return nil, errChan, subUUID
}

func (s *DataKeeperService) unsubscribeFromUpdates(userID int, UUID string) {
	s.subscriptionsMx.Lock()
	defer s.subscriptionsMx.Unlock()

	delete(s.getDataSubscriptions[userID].subs, UUID)
	if len(s.getDataSubscriptions) == 0 {
		s.getDataSubscriptions[userID].cancelUserSub()
		delete(s.getDataSubscriptions, userID)
	}
}

func (s *DataKeeperService) getUserData(ctx context.Context, userID int) (*datakeeper.GetDataResponse, []error) {
	var passwordPairsResult []*datakeeper.PasswordPair
	var textResults []*datakeeper.Text

	errs := goroutines.WaitForAll(
		func() error {
			passwordPairs, err := s.dks.GetAllPasswordPairs(ctx, userID)
			if err != nil {
				return err
			}

			passwordPairsResult = mapStoragePairsToGRPC(passwordPairs)

			return nil
		},
		func() error {
			texts, err := s.dks.GetAllTexts(ctx, userID)
			if err != nil {
				return err
			}

			textResults = mapStorageTextsToGRPC(texts)

			return nil
		},
	)

	if len(errs) > 0 {
		return nil, errs
	}

	return &datakeeper.GetDataResponse{
		PasswordPairs: passwordPairsResult,
		Texts:         textResults,
	}, nil
}

func (s *DataKeeperService) notifyUserSubs(userID int) {
	s.subscriptionsMx.RLock()
	select {
	case s.getDataSubscriptions[userID].notifyChan <- struct{}{}:
	default:
	}
	s.subscriptionsMx.RUnlock()
}
