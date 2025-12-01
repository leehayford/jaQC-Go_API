package api

import (
	"jaQC-Go-API/utils"
)

type Aggregate struct {
	utils.Meta  `gorm:"embedded"`
	PID  int64 `gorm:"column:pid; not null" json:"pid"` // PROCESS ID
	VID  int64 `gorm:"column:vid; not null" json:"vid"` // VARIATE ID
	Code int64 `json:"code"`                 // FOR COLOR CODE ON CHART
	ADate     int64   `gorm:"a_date" json:"a_date"`

	Start int64 `json:"start"` // MIN BINARY TIME
	End   int64 `json:"end"`   // MAX BINARY TIME
	Size  int64 `json:"size"`  // SAMPLES IN CLUSTER

	Min  float32 `json:"min"`  // MIN Y IN CLUSTER
	Max  float32 `json:"max"`  // MAX Y IN CLUSTER
	Mean float32 `json:"mean"` // MEAN  Y IN CLUSTER

	Slope float32 `json:"slope"`
	Devi  float32 `json:"devi"`
	Score float32 `json:"score"`

	Valid bool `json:"valid"`
}
func (Aggregate) TableName() string { return "aggregates" }