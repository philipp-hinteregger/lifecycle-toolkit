package keptnevaluation

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	klcv1alpha1 "github.com/keptn-sandbox/lifecycle-controller/operator/api/v1alpha1"
	"io"
	"net/http"
)

type KeptnDynatraceProvider struct {
	Log    logr.Logger
	client http.Client
}

type DynatraceResponse struct {
	TotalCount int               `json:"totalCount"`
	Resolution string            `json:"resolution"`
	Result     []DynatraceResult `json:"result"`
}

type DynatraceResult struct {
	MetricId string          `json:"metricId"`
	Data     []DynatraceData `json:"data"`
}

type DynatraceData struct {
	Timestamps []int64    `json:"timestamps"`
	Values     []*float64 `json:"values"`
}

func (d *KeptnDynatraceProvider) queryEvaluation(objective *klcv1alpha1.Objective, provider klcv1alpha1.KeptnEvaluationProvider) (string, error) {
	qURL := provider.Spec.TargetServer + "/api/v2/metrics/query?metricSelector=" + objective.Query

	d.Log.Info("Running query: " + qURL)

	req, err := http.NewRequest("GET", qURL, nil)
	if err != nil {
		d.Log.Error(err, "Error while creating request")
		return "", err
	}
	req.Header.Set("Authorization", "Api-Token "+provider.Spec.SecretName)
	res, err := d.client.Do(req)
	if err != nil {
		d.Log.Error(err, "Error while creating request")
		return "", err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	result := DynatraceResponse{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		d.Log.Error(err, "Error while parsing response")
		return "", err
	}

	return fmt.Sprintf("%f", d.getSingleValue(result)), nil
}

func (d *KeptnDynatraceProvider) getSingleValue(result DynatraceResponse) float64 {
	var sum float64 = 0
	var count int64 = 0
	for _, r := range result.Result {
		for _, points := range r.Data {
			for _, v := range points.Values {
				if v != nil {
					sum += *v
					count++
				}
			}
		}
	}
	if count < 1 {
		return 0
	}
	return sum / float64(count)
}
