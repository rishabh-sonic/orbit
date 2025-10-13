package opensearch

import (
	"fmt"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/rishabh-sonic/orbit/pkg/config"
)

func New(cfg *config.Config) (*opensearchapi.Client, error) {
	osCfg := opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{cfg.OpenSearchHost},
		},
	}

	if cfg.OpenSearchUsername != "" {
		osCfg.Client.Username = cfg.OpenSearchUsername
		osCfg.Client.Password = cfg.OpenSearchPassword
	}

	client, err := opensearchapi.NewClient(osCfg)
	if err != nil {
		return nil, fmt.Errorf("opensearch client: %w", err)
	}
	return client, nil
}
