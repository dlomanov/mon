package handlers

import (
	"fmt"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"slices"
)

var reportTemplate = template.
	Must(template.New("report").Parse(`{{range $val := .}}<p>{{$val}}</p>{{end}}`))

func Report(logger *zap.Logger, storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := storage.All()
		result := make([]string, 0, len(values))
		for _, v := range values {
			str := fmt.Sprintf("%s: %s\n", v.String(), v.StringValue())
			result = append(result, str)
		}

		slices.Sort(result)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := reportTemplate.Execute(w, result)
		if err != nil {
			logger.Error("error occurred", zap.String("error", err.Error()))
			return
		}
	}
}
