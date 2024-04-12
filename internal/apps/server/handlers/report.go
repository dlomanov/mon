package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"

	"github.com/dlomanov/mon/internal/apps/server/container"
	"go.uber.org/zap"
)

var reportTemplate = template.
	Must(template.New("report").Parse(`{{range $val := .}}<p>{{$val}}</p>{{end}}`))

//	@Summary		Generate a report
//	@Description	Retrieves all metrics and generates a report in HTML format.
//	@ID				generate_report
//
//	@Produce		html
//
//	@Success		200	{object}	string	"Report generated successfully"
//	@Failure		500	{object}	string	"Failed to generate report"
//
//	@Router			/report [get]
func Report(c *container.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := c.Storage.All(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			c.Logger.Error("get entities failed", zap.Error(err))
			return
		}

		result := make([]string, 0, len(values))
		for _, v := range values {
			str := fmt.Sprintf("%s: %s\n", v.String(), v.StringValue())
			result = append(result, str)
		}
		slices.Sort(result)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = reportTemplate.Execute(w, result)
		if err != nil {
			c.Logger.Error("error occurred", zap.String("error", err.Error()))
			return
		}
	}
}
