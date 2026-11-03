package model

type RuleEvaluationResult struct {
	RaisedAlarms  []Alarm
	ClearedAlarms []Alarm
}

type ReadingResult struct {
	Reading              Reading
	RuleEvaluationResult RuleEvaluationResult
}
