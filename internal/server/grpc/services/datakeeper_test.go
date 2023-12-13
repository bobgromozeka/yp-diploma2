package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bobgromozeka/yp-diploma2/internal/interfaces/datakeeper"
	"github.com/bobgromozeka/yp-diploma2/internal/server/storage"
	"github.com/bobgromozeka/yp-diploma2/pkg/testutils"
)

const Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOjF9.fXMEMQ5FGglGA64ijzTPDf-94QSY_Bv1KLcbkHmPO1s"

type MockServerStream struct {
	ctx context.Context
}

func (s *MockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (s *MockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (s *MockServerStream) SetTrailer(md metadata.MD) {

}

func (s *MockServerStream) Context() context.Context {
	return s.ctx
}

func (s *MockServerStream) SendMsg(a any) error {
	return nil
}

func (s *MockServerStream) RecvMsg(a any) error {
	return nil
}

type MockGetDataServer struct {
	grpc.ServerStream
	cancel context.CancelFunc
	Data   *datakeeper.GetDataResponse
}

func (s *MockGetDataServer) Send(m *datakeeper.GetDataResponse) error {
	s.Data = m
	s.cancel()
	return nil
}

func NewMockGetDataServer(ctx context.Context, cancel context.CancelFunc) *MockGetDataServer {
	return &MockGetDataServer{
		&MockServerStream{ctx: ctx},
		cancel,
		nil,
	}
}

func TestDataKeeperService_CreatePasswordPair_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.CreatePasswordPair(context.Background(), &datakeeper.CreatePasswordPairRequest{
		Login:    "login",
		Password: "password",
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreatePasswordPair_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreatePasswordPair(testutils.MatchContext(), gomock.Eq(1), gomock.Eq("login"), gomock.Eq("password"), gomock.Nil()).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.CreatePasswordPair(decorateCtxWithUserID(context.Background()), &datakeeper.CreatePasswordPairRequest{
		Login:    "login",
		Password: "password",
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreatePasswordPair_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreatePasswordPair(testutils.MatchContext(), gomock.Eq(1), gomock.Eq("login"), gomock.Eq("password"), gomock.Nil()).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.CreatePasswordPair(decorateCtxWithUserID(context.Background()), &datakeeper.CreatePasswordPairRequest{
		Login:    "login",
		Password: "password",
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_RemovePasswordPair_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.RemovePasswordPair(context.Background(), &datakeeper.RemovePasswordPairRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemovePasswordPair_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemovePasswordPair(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.RemovePasswordPair(decorateCtxWithUserID(context.Background()), &datakeeper.RemovePasswordPairRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemovePasswordPair_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemovePasswordPair(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.RemovePasswordPair(decorateCtxWithUserID(context.Background()), &datakeeper.RemovePasswordPairRequest{
		ID: 1,
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_CreateText_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.CreateText(context.Background(), &datakeeper.CreateTextRequest{
		Name: "login",
		Text: "password",
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreateText_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreateText(testutils.MatchContext(), gomock.Eq(1), gomock.Eq("name"), gomock.Eq("text"), gomock.Nil()).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.CreateText(decorateCtxWithUserID(context.Background()), &datakeeper.CreateTextRequest{
		Name: "name",
		Text: "text",
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreateText_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreateText(testutils.MatchContext(), gomock.Eq(1), gomock.Eq("name"), gomock.Eq("text"), gomock.Nil()).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.CreateText(decorateCtxWithUserID(context.Background()), &datakeeper.CreateTextRequest{
		Name: "name",
		Text: "text",
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_RemoveText_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveText(context.Background(), &datakeeper.RemoveTextRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemoveText_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemoveText(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveText(decorateCtxWithUserID(context.Background()), &datakeeper.RemoveTextRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemoveText_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemoveText(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveText(decorateCtxWithUserID(context.Background()), &datakeeper.RemoveTextRequest{
		ID: 1,
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_CreateCard_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.CreateCard(context.Background(), &datakeeper.CreateCardRequest{
		Name:              "name",
		Number:            "1234123412341234",
		ValidThroughYear:  24,
		ValidThroughMonth: 7,
		Cvv:               234,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreateCard_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreateCard(testutils.MatchContext(), gomock.Eq(storage.CreateCardParams{
			UserID:            1,
			Name:              "name",
			Number:            "1234123412341234",
			ValidThroughYear:  24,
			ValidThroughMonth: 7,
			Cvv:               234,
			Description:       nil,
		})).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.CreateCard(decorateCtxWithUserID(context.Background()), &datakeeper.CreateCardRequest{
		Name:              "name",
		Number:            "1234123412341234",
		ValidThroughYear:  24,
		ValidThroughMonth: 7,
		Cvv:               234,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreateCard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreateCard(testutils.MatchContext(), gomock.Eq(storage.CreateCardParams{
			UserID:            1,
			Name:              "name",
			Number:            "1234123412341234",
			ValidThroughYear:  24,
			ValidThroughMonth: 7,
			Cvv:               234,
			Description:       nil,
		})).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.CreateCard(decorateCtxWithUserID(context.Background()), &datakeeper.CreateCardRequest{
		Name:              "name",
		Number:            "1234123412341234",
		ValidThroughYear:  24,
		ValidThroughMonth: 7,
		Cvv:               234,
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_RemoveCard_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveCard(context.Background(), &datakeeper.RemoveCardRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemoveCard_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemoveCard(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveCard(decorateCtxWithUserID(context.Background()), &datakeeper.RemoveCardRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemoveCard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemoveCard(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveCard(decorateCtxWithUserID(context.Background()), &datakeeper.RemoveCardRequest{
		ID: 1,
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_CreateBin_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.CreateBin(context.Background(), &datakeeper.CreateBinRequest{
		Name: "name",
		Data: []byte("data"),
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreateBin_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreateBin(testutils.MatchContext(), gomock.Eq(1), gomock.Eq("name"), gomock.Eq([]byte("data")), gomock.Nil()).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.CreateBin(decorateCtxWithUserID(context.Background()), &datakeeper.CreateBinRequest{
		Name: "name",
		Data: []byte("data"),
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_CreateBin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		CreateBin(testutils.MatchContext(), gomock.Eq(1), gomock.Eq("name"), gomock.Eq([]byte("data")), gomock.Nil()).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.CreateBin(decorateCtxWithUserID(context.Background()), &datakeeper.CreateBinRequest{
		Name: "name",
		Data: []byte("data"),
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_RemoveBin_NoAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveBin(context.Background(), &datakeeper.RemoveBinRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemoveBin_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemoveBin(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(errors.New("random error"))

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveBin(decorateCtxWithUserID(context.Background()), &datakeeper.RemoveBinRequest{
		ID: 1,
	})

	assert.ErrorIs(t, ErrInternalServerError, err)
	assert.Nil(t, resp)
}

func TestDataKeeperService_RemoveBin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		RemoveBin(testutils.MatchContext(), gomock.Eq(1), gomock.Eq(1)).
		Return(nil)

	s := NewDataKeeperService(dks)

	resp, err := s.RemoveBin(decorateCtxWithUserID(context.Background()), &datakeeper.RemoveBinRequest{
		ID: 1,
	})

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.EmptyResponse{}, resp)
}

func TestDataKeeperService_GetData(t *testing.T) {
	decoratedCtx := decorateCtxWithUserID(context.Background())
	cancelCtx, cancel := context.WithCancel(decoratedCtx)

	gds := NewMockGetDataServer(cancelCtx, cancel)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cards := []storage.Card{
		{
			ID:                1,
			UserID:            1,
			Name:              "name",
			Number:            "1234123412341234",
			ValidThroughYear:  24,
			ValidThroughMonth: 9,
			CVV:               123,
		},
	}
	pps := []storage.PasswordPair{
		{
			ID:       1,
			UserID:   1,
			Login:    "login",
			Password: "password",
		},
	}
	texts := []storage.Text{
		{
			ID:     1,
			UserID: 1,
			Name:   "name",
			T:      "text",
		},
	}
	bins := []storage.Bin{
		{
			ID:     1,
			UserID: 1,
			Name:   "name",
			Data:   []byte("data"),
		},
	}

	dks := storage.NewMockDataKeeperStorage(ctrl)
	dks.
		EXPECT().
		GetAllCards(testutils.MatchContext(), gomock.Eq(1)).
		Return(cards, nil)
	dks.
		EXPECT().
		GetAllBins(testutils.MatchContext(), gomock.Eq(1)).
		Return(bins, nil)
	dks.
		EXPECT().
		GetAllTexts(testutils.MatchContext(), gomock.Eq(1)).
		Return(texts, nil)
	dks.
		EXPECT().
		GetAllPasswordPairs(testutils.MatchContext(), gomock.Eq(1)).
		Return(pps, nil)

	s := NewDataKeeperService(dks)

	err := s.GetData(&datakeeper.GetDataRequest{}, gds)

	assert.Nil(t, err)
	assert.Equal(t, &datakeeper.GetDataResponse{
		PasswordPairs: mapStoragePairsToGRPC(pps),
		Texts:         mapStorageTextsToGRPC(texts),
		Cards:         mapStorageCardsToGRPC(cards),
		Bins:          mapStorageBinsToGRPC(bins),
	}, gds.Data)
}

func decorateCtxWithMd(ctx context.Context) context.Context {
	md := metadata.MD{}
	md.Set("Authorization", "Bearer "+Token)
	return metadata.NewOutgoingContext(ctx, md)
}

func decorateCtxWithUserID(ctx context.Context) context.Context {
	return context.WithValue(context.Background(), "userID", 1)
}
