package utils

import (
	"errors"
	"log"

	"github.com/runtime-metrics-course/internal/models"
)

func ParseMetricToJSON(mType, name string, val interface{}) (*models.MetricJSON, error) {
	metric := models.MetricJSON{
		ID:    name,
		MType: mType,
	}
	switch mType {
	case models.Counter:
		valInt, ok := val.(int64)
		if !ok {
			log.Println("parse error")
			return &metric, errors.New("parse error")
		}
		metric.Delta = &valInt
	case models.Gauge:

		valFl, ok := val.(float64)
		if !ok {
			log.Println("parse error")
			return &metric, errors.New("parse error")
		}
		metric.Value = &valFl
	default:
		log.Println("parse error")
		return &metric, errors.New("parse error")
	}
	return &metric, nil
}
