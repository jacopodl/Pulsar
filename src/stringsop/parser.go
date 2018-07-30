package stringsop

import (
	"fmt"
	"strings"
	"io"
)

type StringsOp struct {
	options   []Option
	current   int
	separator string
}

func NewStringsOp(options []Option) *StringsOp {
	return &StringsOp{options, 0, "="}
}

func (sop *StringsOp) Next(sopts []string) (string, string, error) {
	if sop.current >= len(sop.options) {
		return "", "", io.EOF
	}

	copt := sop.options[sop.current]

	defer func() {
		sop.current++
	}()

	for _, gopt := range sopts {
		split := strings.SplitN(gopt, sop.separator, 2)
		if copt.Name == split[0] {
			if len(split) < 2 || split[1] == "" {
				return "", "", fmt.Errorf("missing argument for option: %s", split[0])
			}

			if copt.Values != nil {
				for _, val := range copt.Values {
					if val == split[1] {
						return split[0], split[1], nil
					}
				}
				return "", "", fmt.Errorf("%s requires one of these: %s", split[0], copt.Values)
			}
			return split[0], split[1], nil
		}
	}
	if copt.Default == "" {
		return "", "", fmt.Errorf("missing required option: %s", copt.Name)
	}
	return copt.Name, copt.Default, nil
}

func (sop *StringsOp) Reset() {
	sop.current = 0
}
