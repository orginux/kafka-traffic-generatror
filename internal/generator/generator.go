// Package generator provides functionalities to generate fake datata.
package generator

import (
	"fmt"
	"log/slog"

	"github.com/brianvoe/gofakeit/v6"

	"kafka-traffic-generator/internal/config"
)

// type Fields struct {
// 	Fields []config.Field
// }

// generateFields generates a slice of gofakeit.Field based on the provided field configurations.
// TODO shoudl be a method
func GenerateFields(fieldConfigs []config.Field, logger *slog.Logger) ([]gofakeit.Field, error) {
	if len(fieldConfigs) == 0 {
		return nil, fmt.Errorf("Fields are not defined in the config file")
	}

	var fields []gofakeit.Field
	for _, fieldConfig := range fieldConfigs {
		logger.Debug("Preparing fake data", slog.String("field", fieldConfig.Name), slog.String("function", fieldConfig.Function))
		params := gofakeit.NewMapParams()
		for key, value := range fieldConfig.Params {
			logger.Debug(
				"Preparing fake data",
				slog.String("field", fieldConfig.Name),
				slog.String("function", fieldConfig.Function),
				slog.String("param", key),
			)
			params.Add(key, value)
		}

		field := gofakeit.Field{
			Name:     fieldConfig.Name,
			Function: fieldConfig.Function,
			Params:   *params,
		}

		fields = append(fields, field)
	}
	logger.Debug("Fields generated")
	return fields, nil
}
