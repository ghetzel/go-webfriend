package core

import "fmt"

type ConfigureArgs struct {
	Demo           interface{} `json:"demo"`            // null
	UserAgent      interface{} `json:"user_agent"`      // null
	ExtraHeaders   interface{} `json:"extra_headers"`   // null
	Cache          interface{} `json:"cache"`           // null
	Console        interface{} `json:"console"`         // null
	ReferrerPrefix interface{} `json:"referrer_prefix"` // null
}

// [SKIP]
// Configures various features of the Remote Debugging protocol and provides
// environment setup.
func (self *Commands) Configure(args *ConfigureArgs) error {
	return fmt.Errorf(`NI`)
}
