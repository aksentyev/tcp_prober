package exporter

type Config struct {
    Address         string
    Port            string
    Labels          map[string]string
    ExporterOptions map[string]string
    CacheTTL        int
}
