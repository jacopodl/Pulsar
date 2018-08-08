package stringsop

import (
	"fmt"
	"strings"
)

type Option struct {
	Name        string
	Values      []string
	Default     string
	Description string
}

func (o *Option) String() string {
	var value = "..."

	if o.Values != nil {
		value = fmt.Sprintf("<%s>", strings.Join(o.Values, "|"))
	}
	return fmt.Sprintf("%s=%s", o.Name, value)
}
