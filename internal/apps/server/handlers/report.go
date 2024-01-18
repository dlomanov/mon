package handlers

import (
	"fmt"
	"github.com/dlomanov/mon/internal/storage"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"slices"
)

var reportTemplate = template.
	Must(template.New("report").Parse(`{{range $val := .}}<p>{{$val}}</p>{{end}}`))

func Report(logger *zap.Logger, db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := db.All()
		result := make([]string, len(values))
		i := 0
		for k, v := range values {
			str := fmt.Sprintf("%s: %s\n", k, v)
			result[i] = str
			i++
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
