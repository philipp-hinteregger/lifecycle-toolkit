package keptnevaluation

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	klcv1alpha1 "github.com/keptn-sandbox/lifecycle-controller/operator/api/v1alpha1"
	"github.com/keptn-sandbox/lifecycle-controller/operator/api/v1alpha1/common"
	promapi "github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"math"
	"net/http"
	"strconv"
	"time"
)

type KeptnPrometheusProvider struct {
	Log logr.Logger
}

func (r *KeptnPrometheusProvider) queryEvaluation(objective klcv1alpha1.Objective, provider klcv1alpha1.KeptnEvaluationProvider) *klcv1alpha1.EvaluationStatusItem {
	query := &klcv1alpha1.EvaluationStatusItem{
		Value:  "",
		Status: common.StateFailed, //setting status per default to failed
	}

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
		query.Message = err.Error()
		return query
	}

	if len(w) != 0 {
		query.Message = w[0]
		r.Log.Info("Prometheus API returned warnings: " + w[0])
	}

	// check if we can cast the result to a vector, it might be another data struct which we can't process
	resultVector, ok := result.(model.Vector)
	if !ok {
		query.Message = "could not cast result"
		return query
	}

	// We are only allowed to return one value, if not the query may be malformed
	// we are using two different errors to give the user more information about the result
	if len(resultVector) == 0 {
		r.Log.Info("No values in query result")
		query.Message = "No values in query result"
		return query
	} else if len(resultVector) > 1 {
		r.Log.Info("Too many values in the query result")
		query.Message = "Too many values in the query result"
		return query
	}

	query.Value = resultVector[0].Value.String()
	check, err := checkValue(objective, query)

	if err != nil {
		query.Message = err.Error()
		r.Log.Error(err, "Could not check query result")
	}
	if check {
		query.Status = common.StateSucceeded
	}
	return query
}

func checkValue(objective klcv1alpha1.Objective, query *klcv1alpha1.EvaluationStatusItem) (bool, error) {

	if len(query.Value) == 0 || len(objective.EvaluationTarget) == 0 {
		return false, fmt.Errorf("no values")
	}

	eval := objective.EvaluationTarget[1:]
	sign := objective.EvaluationTarget[:1]

	resultValue, err := strconv.ParseFloat(query.Value, 64)
	if err != nil || math.IsNaN(resultValue) {
		return false, err
	}

	compareValue, err := strconv.ParseFloat(eval, 64)
	if err != nil || math.IsNaN(compareValue) {
		return false, err
	}

	// choose comparator
	switch sign {
	case ">":
		return resultValue > compareValue, nil
	case "<":
		return resultValue < compareValue, nil
	default:
		return false, fmt.Errorf("invalid operator")
	}
}
