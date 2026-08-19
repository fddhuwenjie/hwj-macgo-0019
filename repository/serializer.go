package repository

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	PrefixCategory           = "category:"
	PrefixBudgetPolicy       = "budget_policy:"
	PrefixWindow             = "window:"
	PrefixReservation        = "reservation:"
	PrefixAttempt            = "attempt:"
	PrefixResult             = "result:"
	PrefixBackoffPlan        = "backoff_plan:"
	PrefixFailureStreak      = "failure_streak:"
	PrefixSuspension         = "suspension:"
	PrefixRecoveryCredential = "recovery_credential:"
)

func cloneAggregate(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, ErrInvalidAggregate
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		cp := reflect.New(rv.Type())
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, cp.Interface()); err != nil {
			return nil, err
		}
		return cp.Elem().Interface(), nil
	}
	cp := reflect.New(rv.Elem().Type())
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, cp.Interface()); err != nil {
		return nil, err
	}
	return cp.Interface(), nil
}

func idOf(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	f := rv.FieldByName("ID")
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return f.String()
}

func versionOf(v interface{}) int64 {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName("Version")
	if !f.IsValid() || f.Kind() != reflect.Int64 {
		return 0
	}
	return f.Int()
}

func setVersion(v interface{}, version int64) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	f := rv.FieldByName("Version")
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.Int64 {
		f.SetInt(version)
	}
	tf := rv.FieldByName("UpdatedAt")
	if tf.IsValid() && tf.CanSet() && tf.Type() == reflect.TypeOf(time.Time{}) {
		tf.Set(reflect.ValueOf(time.Now().UTC()))
	}
}

func sortedKeys(records map[string]json.RawMessage, prefix string) []string {
	keys := make([]string, 0)
	for k := range records {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
