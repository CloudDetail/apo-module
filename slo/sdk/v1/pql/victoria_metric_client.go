package pql

type VictoriaMetricsClient struct {
	*PrometheusClient
}

func (client *VictoriaMetricsClient) BucketLabelName() string {
	return "vmrange"
}
