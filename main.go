package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

//自定义端口
var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests")

var (
	//Gauge 仪表盘类型
	opsQueued = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "our_company",
		Subsystem: "blob_storage",
		Name:      "ops_queued",
		Help:      "Number of blob storage operations waiting to be processed",
	})

	jobsInQueue = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "job_in_queue",
		Help: "Current number of jobs in the queue",
	}, []string{"job_type"})

	//Count 计数器类型
	taskCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "worker_pool",
		Name:      "completed_tasks_total",
		Help:      "Total number of tasks completed.",
	})

	//Summary 类型，需要提供分位点
	temps = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "pond_temperature_celsius",
		Help:       "The temperature of the frog pond.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})

	//Histogram 类型，需要提供 Bucket大小
	tempsHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "pond_temperature_histogram_celsius",
		Help:        "The temperature of the frog pond.",
		ConstLabels: nil,
		Buckets:     prometheus.LinearBuckets(20, 5, 5), // 5 个 buckets, 跨度为 5 摄氏度.
		//Buckets:     []float64{20, 25, 30, 35, 40}, //等价于这个
	})
)

func init() {
	prometheus.MustRegister(opsQueued, jobsInQueue, taskCounter, temps, tempsHistogram)
}

func main() {
	flag.Parse()
	jobsInQueue.WithLabelValues("testjob").Add(3)
	go func() {
		for true {
			//每隔一秒加 4
			opsQueued.Add(4)
			time.Sleep(time.Second)
			//计数器加一
			taskCounter.Inc()
		}
	}()
	go func() {
		// 模拟观察温度
		for i := 0; i < 1000; i++ {
			temps.Observe(30 + math.Floor(120*math.Sin(float64(i)*0.1))/10)
		}

		// 仅供示范， 让我们通过使用它的 Write 方法检查摘要的状态 (通常只在 Prometheus 内部使用).
		metric := &dto.Metric{}
		temps.Write(metric)
		fmt.Println(proto.MarshalTextString(metric))
	}()
	go func() {
		// 模拟观察温度
		for i := 0; i < 1000; i++ {
			tempsHistogram.Observe(30 + math.Floor(120*math.Sin(float64(i)*0.1))/10)
		}

		// 仅供示范， 让我们通过使用它的 Write 方法检查摘要的状态 (通常只在 Prometheus 内部使用).
		metric := &dto.Metric{}
		tempsHistogram.Write(metric)
		fmt.Println(proto.MarshalTextString(metric))
	}()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
