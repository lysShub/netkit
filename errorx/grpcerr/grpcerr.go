package grpcerr

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromError copy from https://github.com/grpc/grpc-go/blob/v1.70.0/status/status.go#L96
func FromError(err error) (s *status.Status, ok bool) {
	if err == nil {
		return nil, true
	}
	type grpcstatus interface{ GRPCStatus() *status.Status }
	if gs, ok := err.(grpcstatus); ok {
		grpcStatus := gs.GRPCStatus()
		if grpcStatus == nil {
			// Error has status nil, which maps to codes.OK. There
			// is no sensible behavior for this, so we turn it into
			// an error with codes.Unknown and discard the existing
			// status.
			return status.New(codes.Unknown, err.Error()), false
		}
		return grpcStatus, true
	}
	var gs grpcstatus
	if errors.As(err, &gs) {
		s = gs.GRPCStatus()
		if s != nil {
			return s, true
		}
	}
	return status.New(codes.Unknown, err.Error()), false
}
