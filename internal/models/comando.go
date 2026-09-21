package models

type Command struct {
	Command *int `json:"command"`
}

func (c Command) IsValid() bool {
	return c.Command != nil && (*c.Command == 0 || *c.Command == 1)
}
