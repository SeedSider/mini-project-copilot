package api

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) notImplementedError() error {
	st := status.New(codes.Unimplemented, "Not implemented yet")
	return st.Err()
}

func (s *Server) serverErrorWithDetail(err error) error {
	st := status.New(codes.Internal, err.Error())
	return st.Err()
}

func (s *Server) serverError() error {
	st := status.New(codes.Internal, "Internal Error")
	return st.Err()
}

func (s *Server) badRequestError(msg string) error {
	st := status.New(codes.InvalidArgument, msg)
	return st.Err()
}

func (s *Server) unauthorizedError() error {
	st := status.New(codes.Unauthenticated, "Unauthorized")
	return st.Err()
}

func (s *Server) conflictError(msg string) error {
	st := status.New(codes.AlreadyExists, msg)
	return st.Err()
}

func (s *Server) notFoundError(msg string) error {
	st := status.New(codes.NotFound, msg)
	return st.Err()
}
