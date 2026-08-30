package server

import (
	"embed"

	webfriend "go.gary.cool/go-webfriend"
)

//go:embed ui/*
//go:embed ui/**
var embedded embed.FS

type Server struct {
	env *webfriend.Environment
}

func NewServer(env *webfriend.Environment) *Server {
	return &Server{
		env: env,
	}
}

func (self *Server) ListenAndServe(address string) error {
	return nil
}
