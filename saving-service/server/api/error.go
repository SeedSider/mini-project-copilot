package api

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) serverError() error {
	st := status.New(codes.Internal, "Internal Error")
	return st.Err()
}

func (s *Server) badRequestError(msg string) error {
	st := status.New(codes.InvalidArgument, msg)
	return st.Err()
}
