package main

import (
    "github.com/aksentyev/hubble/hubble"
    "github.com/aksentyev/hubble/backend/consul"
    "github.com/aksentyev/hubble/exportertools"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/common/log"

    "github.com/aksentyev/tcp_prober/exporter"
    "github.com/aksentyev/tcp_prober/util"

    "flag"
    "fmt"

    "net/http"
    // _ "net/http/pprof"
)

// landingPage contains the HTML served at '/'.
// TODO: Make this nicer and more informative.
var landingPage = []byte(`<html>
<head><title>exporter</title></head>
<body>
<h1>metric exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)


var (
    d *hubble.Dispatcher
    esURI string
)

var (
    consulURL = flag.String(
        "consul.url", "consul.service.consul:8500",
        "Consul url",
    )
    consulDC = flag.String(
        "consul.dc", "staging",
        "Consul datacenter",
    )
    consulTag = flag.String(
        "consul.tag", "",
        "Look for services that have the tag specified.",
    )
    listenAddress = flag.String(
        "listen", ":9115",
        "Address to listen on for web interface and telemetry.",
    )
    metricPath = flag.String(
        "web.telemetry-path", "/metrics",
        "Path under which to expose exporter.",
    )
    updateInterval = flag.Int(
        "update-interval", 120,
        "Update interval in seconds",
    )
    scrapeInterval = flag.Int(
        "scrape-interval", 60,
        "Scrape interval in seconds",
    )
    showVersion = flag.Bool(
        "version", false,
        "Show versions and exit",
    )
)

func setup() {
    config := consul.DefaultConfig()
    config.Address = *consulURL
    config.Datacenter = *consulDC

    client, err := consul.New(config)
    if err != nil {
        panic(err)
    }

    kv := consul.NewKV(client)
    h := hubble.New(client, kv, *consulTag)

    filterCB := func(list []*hubble.Service) []*hubble.Service {
        if *consulTag == "" {
            return list
        }

        var servicesForMonitoring []*hubble.Service
        for _, svc := range list {
            if util.IncludesStr(svc.Tags, *consulTag) && !util.IncludesStr(svc.Tags, "udp") {
                servicesForMonitoring = append(servicesForMonitoring, svc)
            }
        }
        return servicesForMonitoring
    }

    cb := func() (list []*hubble.ServiceAtomic, err error) {
        services, err := h.Services(filterCB)
        if err != nil {
            return list, err
        }
        for _, svc := range services {
            for _, el := range svc.MakeAtomic(nil) {
                list = append(list, el)
            }
        }
        return list, err
    }

    d = hubble.NewDispatcher(*updateInterval)
    go d.Run(cb)
}

func printVersions(){
    fmt.Printf("exporter: %v\n", exporter.VERSION)
    fmt.Printf("hubble: %v\n", hubble.VERSION)
    fmt.Printf("exportertools: %v\n", exportertools.VERSION)
    fmt.Printf("consul backend: %v\n", consul.VERSION)
}

func listenAndRegister() {
    for svc := range d.ToRegister {
        config := exporter.Config{
            Address:         svc.Address,
            Port:            svc.Port,
            Labels:          svc.ExtraLabels,
            ExporterOptions: svc.ExporterOptions,
            CacheTTL:        *scrapeInterval,
        }
        exp, err := exporter.CreateAndRegister(&config)
        if err == nil {
            d.Register(svc, exp)
            log.Infof("Registered %v %v", svc.Name, svc.Address)
        } else {
            log.Warnf("Register was failed for service %v %v %v", svc.Name, svc.Address, err)
            exp.Close()
        }
    }
}

func getProto(svc *hubble.Service) string {
    if util.IncludesStr(svc.Tags, "udp") {
        return "udp"
    }
    return "tcp"
}

func listenAndUnregister() {
    for m := range d.ToUnregister {
        for h, svc := range m {
            exporter := d.Exporters[h].(*exporter.TcpProbesExporter)
            err := exporter.Close()
            if err != nil {
                log.Warnf("Unregister() for %v %v returned %v:", svc.Name, svc.Address, err)
            } else {
                log.Infof("Unregister service %v %v", svc.Name, svc.Address)
            }
            d.UnregisterWithHash(h)
        }
    }
}

func main(){
    flag.Parse()

    if *showVersion {
        printVersions()
        return
    }

    setup()
    go listenAndRegister()
    go listenAndUnregister()

    http.Handle(*metricPath, prometheus.Handler())
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write(landingPage)
    })
    log.Infof("Starting Server: %s", *listenAddress)
    log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
