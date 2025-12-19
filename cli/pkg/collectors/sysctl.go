package collectors

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type SysctlCollector struct {
}

const SysctlType string = "Sysctl"

type SysctlConfig struct {
	Key   string
	Value string
}

func (s *SysctlCollector) Collect(config any) ([]Configuraion, error) {
	root := "/proc/sys"
	res := make([]Configuraion, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk dir: %w", err)
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(path, "/proc/sys/net") {
			return nil
		}

		c, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("failed to read path: %+v\n", err)
		}

		res = append(res, Configuraion{
			Type: SysctlType,
			Data: SysctlConfig{
				Key:   path,
				Value: strings.TrimSpace(string(c)),
			},
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to capture sysctl config: %w", err)
	}

	return res, nil
}
