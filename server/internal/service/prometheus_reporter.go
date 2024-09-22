package service

import (
	"context"
	"fmt"
	"log"
	"my_project/server/internal/model"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/spf13/viper"
)

func ReportToPrometheus(result *model.ProbeResult, timestamp int64) {
    prometheusHost := viper.GetString("prometheus.host")
    prometheusPort := viper.GetInt("prometheus.port")
    prometheusJob := viper.GetString("prometheus.job")

    uri := fmt.Sprintf("http://%s:%d/metrics/job/%s", prometheusHost, prometheusPort, prometheusJob)
    
    metrics := fmt.Sprintf("packet_loss{ip=\"%s\", timestamp=\"%d\"} %f\n", result.IP, timestamp, result.PacketLoss)
    log.Printf("Sending data to Prometheus: %s", metrics)

    hertzClient, err := client.NewClient()
    if err != nil {
        log.Fatalf("Failed to create Hertz client: %v", err)
    }

    req := &protocol.Request{}
    req.SetMethod("POST")
    req.SetRequestURI(uri)
    req.Header.Set("Content-Type", "text/plain")
    req.SetBodyString(metrics)

    resp := &protocol.Response{}
    err = hertzClient.Do(context.Background(), req, resp)
    if err != nil {
        log.Printf("Error sending data to Prometheus: %v", err)
    } else {
        log.Printf("Successfully sent data to Prometheus: %s, Response: %s", metrics, resp.Body())
    }
}
