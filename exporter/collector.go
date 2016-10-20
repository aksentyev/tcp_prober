package exporter

import (
    "github.com/aksentyev/hubble/exportertools"
    "github.com/aksentyev/puncher/check"
)

type Collector struct {
    *Config
}

func NewCollector(config *Config) *Collector {
    return &Collector{config}
}
// Collecting metrics
func (c *Collector) Collect() ([]*exportertools.Metric, error) {
    status := check.Poke(c.Address, c.Port)

    m := exportertools.Metric {
        Name:        "port_check",
        Description: "Check port is available",
        Type:        exportertools.StringToType("GAUGE"),
        Value:       status,
        Labels:      c.Labels,
    }

    return []*exportertools.Metric{&m}, nil
}
