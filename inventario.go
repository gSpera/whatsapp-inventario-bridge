package main

// Inventario describes a service to bridge
type Inventario struct {
	// Identification name
	Name string `json:"name"`

	Username string `json:"username"`
	Password string `json:"password"`

	Addr   string `json:"addr"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}
