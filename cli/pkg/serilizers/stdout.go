package serilizers

import (
	"encoding/json"
	"fmt"
)

type StdoutSerilizer struct {
}

func (s *StdoutSerilizer) Serilize(config any) error {
	j, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serilize to json: %w", err)
	}

	fmt.Println(string(j))
	return nil
}
