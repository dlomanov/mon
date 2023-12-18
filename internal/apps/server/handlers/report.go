package handlers

import (
	"fmt"
	"github.com/dlomanov/mon/internal/storage"
	"html/template"
	"net/http"
	"slices"
)

const htmlTemplate = `{{range $val := .}}<p>{{$val}}</p>{{end}}`

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
		t := template.Must(template.New("report").Parse(htmlTemplate))
		_ = t.Execute(w, result)
	}
}
