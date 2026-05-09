package models

import (
	"errors"
)

type Segment struct {
	Base
	ProjectID   uint          `gorm:"not null;uniqueIndex:idx_segment_project_name" json:"projectId"`
	Name        string        `gorm:"not null;uniqueIndex:idx_segment_project_name" json:"name"`
	Description string        `gorm:"default:''"                                    json:"description"`
	MatchType   string        `gorm:"default:'AND'"                                 json:"matchType"`
	RuleCount   int           `gorm:"->;column:rule_count"                          json:"ruleCount"`
	Rules       []SegmentRule `gorm:"foreignKey:SegmentID"                          json:"rules,omitempty"`
}

func (Segment) TableName() string {
	return "project.segments"
}

type SegmentRule struct {
	Base
	SegmentID uint   `gorm:"not null" json:"segmentId"`
	Attribute string `gorm:"not null" json:"attribute"`
	Operator  string `gorm:"not null" json:"operator"`
	Value     string `gorm:"not null" json:"value"`
}

func (SegmentRule) TableName() string {
	return "project.segment_rules"
}

type SegmentOperator string

const (
	OpEquals     SegmentOperator = "equals"
	OpNotEquals  SegmentOperator = "not_equals"
	OpContains   SegmentOperator = "contains"
	OpStartsWith SegmentOperator = "starts_with"
	OpEndsWith   SegmentOperator = "ends_with"
	OpGTE        SegmentOperator = "gte"
	OpLTE        SegmentOperator = "lte"
	OpGT         SegmentOperator = "gt"
	OpLT         SegmentOperator = "lt"
	OpIn         SegmentOperator = "in"
	OpNotIn      SegmentOperator = "not_in"
)

var validOperators = map[SegmentOperator]bool{
	OpEquals: true, OpNotEquals: true, OpContains: true,
	OpStartsWith: true, OpEndsWith: true,
	OpGTE: true, OpLTE: true, OpGT: true, OpLT: true,
	OpIn: true, OpNotIn: true,
}

func IsValidOperator(op string) bool {
	return validOperators[SegmentOperator(op)]
}

func ValidateSegmentRules(rules []SegmentRule) error {
	for _, r := range rules {
		if r.Attribute == "" {
			return errors.New("rule attribute is required")
		}
		if r.Operator == "" {
			return errors.New("rule operator is required")
		}
		if !IsValidOperator(r.Operator) {
			return errors.New("invalid operator: " + r.Operator)
		}
		if r.Value == "" {
			return errors.New("rule value is required")
		}
	}
	return nil
}
