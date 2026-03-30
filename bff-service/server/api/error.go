package api

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) serverError() error {
	return status.New(codes.Internal, "Internal Error").Err()
}

func (s *Server) badRequestError(msg string) error {
	return status.New(codes.InvalidArgument, msg).Err()
}

func (s *Server) unauthorizedError() error {
	return status.New(codes.Unauthenticated, "Unauthorized").Err()
}

func (s *Server) notFoundError(msg string) error {
	return status.New(codes.NotFound, msg).Err()
}

func (s *Server) conflictError(msg string) error {
	return status.New(codes.AlreadyExists, msg).Err()
}

func (s *Server) unavailableError(msg string) error {
	return status.New(codes.Unavailable, msg).Err()
}
