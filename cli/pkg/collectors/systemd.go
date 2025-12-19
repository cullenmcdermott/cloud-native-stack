package collectors

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

type SystemDCollector struct {
}

const SystemDType string = "SystemD"

type SystemDConfig struct {
	Unit       string
	Properties map[string]any
}

func (s *SystemDCollector) Collect(config any) ([]Configuraion, error) {
	services := []string{"containerd.service"}
	res := make([]Configuraion, 0)

	ctx := context.Background()
	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to systemd: %w", err)
	}

	for _, service := range services {
		data, err := conn.GetAllPropertiesContext(ctx, service)
		if err != nil {
			return nil, fmt.Errorf("failed to get unit properties: %w", err)
		}

		res = append(res, Configuraion{
			Type: SystemDType,
			Data: SystemDConfig{
				Unit:       service,
				Properties: data,
			},
		})
	}

	return res, nil
}
