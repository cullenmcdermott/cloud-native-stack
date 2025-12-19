package collectors

import (
	"fmt"
	"os"
	"strings"
)

type GrubCollector struct {
}

const GrubType string = "Grub"

type GrubConfig struct {
	Key   string
	Value string
}

func (s *GrubCollector) Collect(config any) ([]Configuraion, error) {
	root := "/proc/cmdline"
	res := make([]Configuraion, 0)

	cmdline, err := os.ReadFile(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read grub config: %w", err)
	}

	params := strings.Split(string(cmdline), " ")

	for _, param := range params {
		p := strings.TrimSpace(param)
		if p == "" {
			continue
		}

		key, val := "", ""
		s := strings.Split(p, "=")
		if len(s) == 1 {
			key = s[0]
		} else if len(s) == 2 {
			key = s[0]
			val = s[1]
		} else {
			return nil, fmt.Errorf("failed to parse config %s", p)
		}

		res = append(res, Configuraion{
			Type: GrubType,
			Data: GrubConfig{
				Key:   key,
				Value: val,
			},
		})
	}

	return res, nil
}
