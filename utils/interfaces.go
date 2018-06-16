package utils

import "github.com/ghetzel/go-webfriend/scripting"

type Scopeable interface {
	Scope() *scripting.Scope
}
