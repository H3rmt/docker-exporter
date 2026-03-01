package web

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/h3rmt/docker-exporter/internal/log"
)

import _ "embed"

//go:embed assets/chart.umd.min.js
var chartJs []byte

//go:embed assets/popper.min.js
var popperJs []byte

//go:embed assets/tippy.js
var tippyJs []byte

//go:embed assets/tippy.animations.perspective.css
var tippyPerspectiveCss []byte

//go:embed assets/tippy.themes.translucent.css
var tippTranslucentCss []byte

//go:embed assets/main.css
var css []byte

//go:embed assets/main.js
var js []byte

var assetMap = map[string][]byte{
	"main.css":                         css,
	"main.js":                          js,
	"popper.min.js":                    popperJs,
	"tippy.min.js":                     tippyJs,
	"tippy.animations.perspective.css": tippyPerspectiveCss,
	"tippy.themes.translucent.css":     tippTranslucentCss,
	"chart.umd.min.js":                 chartJs,
}

var contentTypeMap = map[string]string{
	".js":  "application/javascript; charset=utf-8",
	".css": "text/css; charset=utf-8",
}

func HandleAsset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("path")

		asset, ok := assetMap[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("404 page not found\n"))
			return
		}

		contentType := "application/octet-stream"
		for ext, ct := range contentTypeMap {
			if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
				contentType = ct
				break
			}
		}

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(asset)
	}
}

//go:embed assets/main.html
var index string

func HandleRoot() http.HandlerFunc {
	funcMap := template.FuncMap{
		"toJson": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			str := string(b)
			return str
		},
	}
	tmpl := template.Must(template.New("page").Funcs(funcMap).Parse(index))
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dataPoints := GetData()
		log.GetLogger().Log(ctx, log.LevelTrace, "data points", "dataPoints", len(dataPoints))
		totalMem, _, err := readMemInfo(ctx)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read mem info", "error", err)
		}
		_, _, _, _, cpuCount, err := readProcStat(ctx)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read cpu", "error", err)
		}

		// Prepare initial data as JSON
		type chartData struct {
			Labels        []string
			CPUData       []float64
			CPUDataUser   []float64
			CPUDataSystem []float64
			MemData       []float64
			TotalMem      uint64
			CpuCount      uint64
		}

		cd := chartData{
			Labels:        make([]string, 0),
			CPUData:       make([]float64, 0),
			CPUDataUser:   make([]float64, 0),
			CPUDataSystem: make([]float64, 0),
			MemData:       make([]float64, 0),
			TotalMem:      totalMem / 1024, // turn into KiB
			CpuCount:      cpuCount,
		}

		for i := range dataPoints {
			if !dataPoints[i].Time.IsZero() {
				cd.Labels = append(cd.Labels, dataPoints[i].Time.Format("15:04:05"))
				cd.CPUData = append(cd.CPUData, dataPoints[i].Data.CPUPercent)
				cd.CPUDataUser = append(cd.CPUDataUser, dataPoints[i].Data.CPUPercentUser)
				cd.CPUDataSystem = append(cd.CPUDataSystem, dataPoints[i].Data.CPUPercentSystem)
				cd.MemData = append(cd.MemData, dataPoints[i].Data.MemPercent)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, cd)
	}
}
