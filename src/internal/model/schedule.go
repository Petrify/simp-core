package model

//go:generate go run github.com/objectbox/objectbox-go/cmd/objectbox-gogen

type Class struct {
	Id        uint64
	Name      string
	Abbr      string
	Majors    []string
	ChannelID string
	RoleID    string
}

func (c *Class) HasChan() bool {
	if c.ChannelID != "" {
		return true
	}
	return false
}
