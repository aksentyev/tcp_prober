package exporter

import (
    "github.com/prometheus/common/log"
    "github.com/aksentyev/hubble/exportertools"
)

// Exporter collects Postgres metrics. It implements prometheus.Collector.
type TcpProbesExporter struct {
    *exportertools.BaseExporter
    Config   *Config
}

// NewExporter returns a new PostgreSQL exporter for the provided DSN.
func CreateAndRegister(config *Config) (*TcpProbesExporter, error) {
    exp := TcpProbesExporter{
        Config: config,
        BaseExporter: exportertools.NewBaseExporter("tcp_availabity", config.CacheTTL, config.Labels),
    }
    err := exportertools.Register(&exp)
    if err != nil {
        return &exp, err
    }
    return &exp, nil
}

func (e *TcpProbesExporter) Setup() error {
    e.AddCollector(NewCollector(e.Config))
    return nil
}

func (e *TcpProbesExporter) Close() (err error) {
    defer close(e.Control)

    err = exportertools.Unregister(e)

    e.Control<- true
    log.Debugf("Stop processing metric for %v", e.Labels)

    return err
}
