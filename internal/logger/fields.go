package logger

import (
	"go.uber.org/zap"
)

func Fields(keysAndValues ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			break
		}

		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}

		val := keysAndValues[i+1]

		switch v := val.(type) {
		case string:
			fields = append(fields, zap.String(key, v))
		case int:
			fields = append(fields, zap.Int(key, v))
		case bool:
			fields = append(fields, zap.Bool(key, v))
		case float64:
			fields = append(fields, zap.Float64(key, v))
		case error:
			fields = append(fields, zap.Error(v))
		default:
			fields = append(fields, zap.Any(key, v))
		}
	}

	return fields
}
