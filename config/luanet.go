package config

// Tracks the configuration of the luanet's identity.
type Luanet struct {
	Domain      string
	Node        string
	Api         string
	ExpiresTime int64
}
