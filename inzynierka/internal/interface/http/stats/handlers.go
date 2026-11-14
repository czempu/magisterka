package statshttp

import (
    "bytes"
    "github.com/labstack/echo/v4"
    "erpbadger/internal/interface/http/charts"


    uc "erpbadger/internal/usecase/stats"
)





type Handlers struct {
    svc uc.Usecase
}
func New(svc uc.Usecase) *Handlers { return &Handlers{svc: svc} }





func (h *Handlers) EmployeeStatsChart(c echo.Context) error {

	

        stats, err := h.svc.GetEmpStats(c.Request().Context())
	if err != nil {

	}
	chart := charts.PieEmployees(stats.Total, stats.Active)
    
   
    // Renderujemy do bufora ZAMIAST pisać do Writer()
    buf := bytes.NewBuffer(nil)
    if err := chart.Render(buf); err != nil {
        return err
    }

    // Zwracamy sam fragment HTML — HTMX to obsłuży i wykona <script>
    return c.HTMLBlob(200, buf.Bytes())
}

