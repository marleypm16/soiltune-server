package services

import (
    "context"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "soiltune-consumer/internal/config"
)

type InfluxService struct {
    client influxdb2.Client
    org    string
    bucket string
}

func NewInfluxService(cfg config.InfluxConfig) (*InfluxService, error) {
    client := influxdb2.NewClient(cfg.URL, cfg.Token)
    return &InfluxService{client: client, org: cfg.Org, bucket: cfg.Bucket}, nil
}

func (s *InfluxService) Close() {
    if s != nil && s.client != nil {
        s.client.Close()
    }
}

func (s *InfluxService) Bucket() string { return s.bucket }

func (s *InfluxService) Query(ctx context.Context, flux string) ([]map[string]interface{}, error) {
    queryAPI := s.client.QueryAPI(s.org)
    result, err := queryAPI.Query(ctx, flux)
    if err != nil {
        return nil, err
    }

    var rows []map[string]interface{}
    for result.Next() {
        rec := result.Record()
        if rec == nil {
            continue
        }
        rows = append(rows, rec.Values())
    }

    if result.Err() != nil {
        return nil, result.Err()
    }

    return rows, nil
}
