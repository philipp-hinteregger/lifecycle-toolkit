package keptnevaluation

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	klcv1alpha1 "github.com/keptn-sandbox/lifecycle-controller/operator/api/v1alpha1"
	promapi "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"net/http"
	"time"
)

type KeptnPrometheusProvider struct {
	Log logr.Logger
}

func (r *KeptnPrometheusProvider) EvaluateQuery(objective klcv1alpha1.Objective, provider klcv1alpha1.KeptnEvaluationProvider) (string, error) {
	queryTime := time.Now().UTC()
	r.Log.Info("Running query: /api/v1/query?query=" + objective.Query + "&time=" + queryTime.String())

	client, err := promapi.NewClient(promapi.Config{Address: provider.Spec.TargetServer, Client: &http.Client{}})
	api := prometheus.NewAPI(client)
	result, w, err := api.Query(
		context.Background(),
		objective.Query,
		queryTime,
		[]prometheus.Option{}...,
	)

	if err != nil {
		return "", err
	}

	if len(w) != 0 {
		r.Log.Info("Prometheus API returned warnings: " + w[0])
	}

	// check if we can cast the result to a vector, it might be another data struct which we can't process
	resultVector, ok := result.(model.Vector)
	if !ok {
		return "", fmt.Errorf("could not cast result")
	}

	// We are only allowed to return one value, if not the query may be malformed
	// we are using two different errors to give the user more information about the result
	if len(resultVector) == 0 {
		r.Log.Info("No values in query result")
		return "", fmt.Errorf("no values in query result")
	} else if len(resultVector) > 1 {
		r.Log.Info("Too many values in the query result")
		return "", fmt.Errorf("too many values in the query result")
	}
	value := resultVector[0].Value.String()
	return value, nil
}
