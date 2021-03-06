package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func cloudwatchSend(r *result) error {
	svc := cloudwatch.New(session.New())

	metrics := []*cloudwatch.MetricDatum{}
	metrics = append(metrics, cloudwatchMetrics(r.totals, nil)...)

	for name, c := range r.queues {
		metrics = append(metrics, cloudwatchMetrics(c, []*cloudwatch.Dimension{
			{Name: aws.String("Queue"), Value: aws.String(name)},
		})...)
	}

	for name, c := range r.pipelines {
		metrics = append(metrics, cloudwatchMetrics(c, []*cloudwatch.Dimension{
			{Name: aws.String("Pipeline"), Value: aws.String(name)},
		})...)
	}

	log.Printf("Extracted %d cloudwatch metrics from results", len(metrics))

	for _, chunk := range chunkCloudwatchMetrics(10, metrics) {
		log.Printf("Submitting chunk of %d metrics to Cloudwatch", len(chunk))
		_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
			MetricData: chunk,
			Namespace:  aws.String("Buildkite"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func cloudwatchMetrics(c counts, dimensions []*cloudwatch.Dimension) []*cloudwatch.MetricDatum {
	m := []*cloudwatch.MetricDatum{}

	for k, v := range c {
		m = append(m, &cloudwatch.MetricDatum{
			MetricName: aws.String(k),
			Dimensions: dimensions,
			Value:      aws.Float64(float64(v)),
			Unit:       aws.String("Count"),
		})
	}

	return m
}

func chunkCloudwatchMetrics(size int, data []*cloudwatch.MetricDatum) [][]*cloudwatch.MetricDatum {
	var chunks = [][]*cloudwatch.MetricDatum{}
	for i := 0; i < len(data); i += size {
		end := i + size
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
