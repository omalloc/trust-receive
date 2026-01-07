package service

import (
	"context"
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
	pb "github.com/omalloc/trust-receive/api/receive/v1"
	"github.com/omalloc/trust-receive/internal/biz"
)

// ReceiveServiceService is the implementation of the Receive service
type ReceiveServiceService struct {
	pb.UnimplementedReceiveServiceServer

	receive *biz.ReceiveUsecase
}

// NewReceiveServiceService creates a new Receive service
func NewReceiveServiceService(receive *biz.ReceiveUsecase) *ReceiveServiceService {
	return &ReceiveServiceService{
		receive: receive,
	}
}

// Receive implements the Receive interface, receiving file info reported by clients and calling business logic for validation
func (s *ReceiveServiceService) Receive(ctx context.Context, req *pb.ReceiveRequest) (*pb.ReceiveReply, error) {
	err := s.receive.Verify(ctx, &biz.FileInfo{
		URL:          req.Url,
		Hash:         req.Hash,
		FileSize:     req.Cl,
		LastModified: req.Lm,
	})

	if err != nil {
		return nil, errors.New(http.StatusConflict, "VERIFY_FAILED", err.Error())
	}

	return &pb.ReceiveReply{Message: "ok"}, nil
}
