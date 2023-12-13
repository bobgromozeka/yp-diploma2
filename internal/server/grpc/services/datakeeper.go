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

type userID = int

type subscription struct {
	// stream server stream subscription
	stream datakeeper.DataKeeper_GetDataServer

	// errChan error channel associated with stream
	errChan chan error
}

// userSubscription contains all user connections (to sync data between different client of the same user)
type userSubscription struct {
	// notifyChan channel to notify about updated data for one user
	notifyChan chan struct{}

	// subs list of different client connection of the same user
	subs map[string]subscription

	// lastData user data that will be sent to all connections
	lastData *datakeeper.GetDataResponse

	// cancelUserSub ctx cancel func associated with user
	cancelUserSub context.CancelFunc
}

// DataKeeperService implementation of gRPC data keeper service
type DataKeeperService struct {
	datakeeper.UnimplementedDataKeeperServer
	dks                  storage.DataKeeperStorage
	getDataSubscriptions map[userID]userSubscription
	subscriptionsMx      sync.RWMutex
}

func NewDataKeeperService(dks storage.DataKeeperStorage) *DataKeeperService {
	return &DataKeeperService{
		dks:                  dks,
		getDataSubscriptions: map[userID]userSubscription{},
		subscriptionsMx:      sync.RWMutex{},
	}
}

func (s *DataKeeperService) CreatePasswordPair(ctx context.Context, req *datakeeper.CreatePasswordPairRequest) (*datakeeper.EmptyResponse, error) {
	userID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.CreatePasswordPair(ctx, userID, req.Login, req.Password, req.Description)
	if err != nil {
		fmt.Println("Error during creating password pair: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(userID)

	return &datakeeper.EmptyResponse{}, nil
}

// GetData
// User id is added to subscriptions map and last data for this user immediately is sent to him.
// Connection remains open and updated data is sent when notified
func (s *DataKeeperService) GetData(req *datakeeper.GetDataRequest, serv datakeeper.DataKeeper_GetDataServer) error {
	uID, ok := serv.Context().Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", serv.Context().Value(interceptors.UserID))
		return ErrInternalServerError
	}

	subErr, errChan, subUUID := s.subscribeForUpdates(uID, serv)
	if subErr != nil {
		return subErr
	}

	select {
	case <-serv.Context().Done():
		s.unsubscribeFromUpdates(uID, subUUID)
		return nil
	case err := <-errChan:
		s.unsubscribeFromUpdates(uID, subUUID)
		return err
	}
}

func (s *DataKeeperService) RemovePasswordPair(ctx context.Context, req *datakeeper.RemovePasswordPairRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.RemovePasswordPair(ctx, uID, int(req.ID))
	if err != nil {
		log.Default().Println("Error during removing password pair: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(uID)

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) CreateText(ctx context.Context, req *datakeeper.CreateTextRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.CreateText(ctx, int(uID), req.Name, req.Text, req.Description)
	if err != nil {
		log.Default().Println("Error during creating text: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(int(uID))

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) RemoveText(ctx context.Context, req *datakeeper.RemoveTextRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.RemoveText(ctx, int(uID), int(req.ID))
	if err != nil {
		log.Default().Println("Error during removing text: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(int(uID))

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) CreateCard(ctx context.Context, req *datakeeper.CreateCardRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.CreateCard(ctx, storage.CreateCardParams{
		UserID:            int(uID),
		Name:              req.Name,
		Number:            req.Number,
		ValidThroughMonth: int(req.ValidThroughMonth),
		ValidThroughYear:  int(req.ValidThroughYear),
		Cvv:               int(req.Cvv),
		Description:       req.Description,
	})
	if err != nil {
		log.Default().Println("Error during creating card: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(int(uID))

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) RemoveCard(ctx context.Context, req *datakeeper.RemoveCardRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.RemoveCard(ctx, int(uID), int(req.ID))
	if err != nil {
		log.Default().Println("Error during removing card: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(int(uID))

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) CreateBin(ctx context.Context, req *datakeeper.CreateBinRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.CreateBin(ctx, int(uID), req.Name, req.Data, req.Description)
	if err != nil {
		log.Default().Println("Error during creating bin: ", err)
		return nil, ErrInternalServerError
	}

	s.notifyUserSubs(int(uID))

	return &datakeeper.EmptyResponse{}, nil
}

func (s *DataKeeperService) RemoveBin(ctx context.Context, req *datakeeper.RemoveBinRequest) (*datakeeper.EmptyResponse, error) {
	uID, ok := ctx.Value(interceptors.UserID).(int)
	if !ok {
		fmt.Printf("No user id in context or wrong type: %t\n", ctx.Value(interceptors.UserID))
		return nil, ErrInternalServerError
	}

	err := s.dks.RemoveBin(ctx, int(uID), int(req.ID))
	if err != nil {
		log.Default().Println("Error during removing bin: ", err)
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

func mapStorageCardsToGRPC(sc []storage.Card) []*datakeeper.Card {
	gc := make([]*datakeeper.Card, len(sc))

	for i, c := range sc {
		gc[i] = &datakeeper.Card{
			ID:                int32(c.ID),
			Name:              c.Name,
			Number:            c.Number,
			ValidThroughMonth: int32(c.ValidThroughMonth),
			ValidThroughYear:  int32(c.ValidThroughYear),
			Cvv:               int32(c.CVV),
			Description:       c.Description,
		}
	}

	return gc
}

func mapStorageBinsToGRPC(sb []storage.Bin) []*datakeeper.Bin {
	gb := make([]*datakeeper.Bin, len(sb))

	for i, b := range sb {
		gb[i] = &datakeeper.Bin{
			ID:          int32(b.ID),
			Name:        b.Name,
			Data:        b.Data,
			Description: b.Description,
		}
	}

	return gb
}

// subscribeForUpdates
// Every connection gets UUID that can be used to unsubscribe from updates.
// If first connection for user was established, data will be selected for him and then sent.
// If there were connections for this user, cached data will be sent.
// This method creates goroutine to read from notifications channel (one is created for every userID)
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

// unsubscribeFromUpdates removes connection from list.
// If this connection was last for user then userSubscription will be removed.
func (s *DataKeeperService) unsubscribeFromUpdates(userID int, UUID string) {
	s.subscriptionsMx.Lock()
	defer s.subscriptionsMx.Unlock()

	delete(s.getDataSubscriptions[userID].subs, UUID)
	if len(s.getDataSubscriptions) == 0 {
		s.getDataSubscriptions[userID].cancelUserSub()
		delete(s.getDataSubscriptions, userID)
	}
}

// getUserData gets all data associated with userID
func (s *DataKeeperService) getUserData(ctx context.Context, userID int) (*datakeeper.GetDataResponse, []error) {
	var passwordPairsResult []*datakeeper.PasswordPair
	var textResults []*datakeeper.Text
	var cardResults []*datakeeper.Card
	var binResults []*datakeeper.Bin

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
		func() error {
			cards, err := s.dks.GetAllCards(ctx, userID)
			if err != nil {
				return err
			}

			cardResults = mapStorageCardsToGRPC(cards)

			return nil
		},
		func() error {
			bins, err := s.dks.GetAllBins(ctx, userID)
			if err != nil {
				return err
			}

			binResults = mapStorageBinsToGRPC(bins)

			return nil
		},
	)

	if len(errs) > 0 {
		return nil, errs
	}

	return &datakeeper.GetDataResponse{
		PasswordPairs: passwordPairsResult,
		Texts:         textResults,
		Cards:         cardResults,
		Bins:          binResults,
	}, nil
}

// notifyUserSubs sends notification into channel associated with userID
func (s *DataKeeperService) notifyUserSubs(userID int) {
	s.subscriptionsMx.RLock()
	select {
	case s.getDataSubscriptions[userID].notifyChan <- struct{}{}:
	default:
	}
	s.subscriptionsMx.RUnlock()
}
