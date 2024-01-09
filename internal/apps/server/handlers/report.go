package handlers

import (
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/logger"
	"github.com/dlomanov/mon/internal/storage"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"slices"
)

var reportTemplate = template.
	Must(template.New("report").Parse(`{{range $val := .}}<p>{{$val}}</p>{{end}}`))

func Report(db storage.Storage) http.HandlerFunc {
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

		w.Header().Set("Content-Type", "text/html")
		err := reportTemplate.Execute(w, result)
		if err != nil {
			logger.Log.Error("error occurred", zap.String("error", err.Error()))
			return
		}
	}
}
