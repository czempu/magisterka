package pdf

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	idraw "image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	render "github.com/go-echarts/snapshot-chromedp/render"
	"github.com/phpdave11/gofpdf"
	qrcode "github.com/skip2/go-qrcode"
	xdraw "golang.org/x/image/draw"

	du "erpbadger/internal/domain/schedule"
)

//
// ====================== HELPERY FORMATOWANIA ======================
//

func formatMinutes(min int) string {
	sign := ""
	if min < 0 {
		sign = "-"
		min = -min
	}
	h := min / 60
	m := min % 60
	return fmt.Sprintf("%s%02d:%02d", sign, h, m)
}

func weekdayPLFromDow(dow int) string {
	// u Ciebie dow jest 1..7 (Mon..Sun)
	names := []string{"-", "Pon", "Wt", "Śr", "Czw", "Pt", "Sob", "Nd"}
	if dow < 1 || dow >= len(names) {
		return "-"
	}
	return names[dow]
}

func dateOnly(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func timeHHMMFromTS(s string) string {
	// "2025-12-05 13:48:37" -> "13:48"
	// "2025-11-01 06:00:47.000000" -> "06:00"
	if len(s) >= 16 {
		return s[11:16]
	}
	// jeśli już jest "HH:MM"
	if len(s) == 5 && s[2] == ':' {
		return s
	}
	return s
}

func safeStrPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func safeIntPtr(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func safeStrPtrTrim(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func monthKeyFromReport(r du.MonthlyReport) string {
	// prefer from_date "YYYY-MM-DD"
	if len(r.FromDate) >= 7 {
		return r.FromDate[:7]
	}
	return ""
}

type inOutPair struct {
	In  string `json:"in"`
	Out string `json:"out"`
}

func parseInOut(raw json.RawMessage) []inOutPair {
	if len(raw) == 0 {
		return nil
	}
	var arr []inOutPair
	if err := json.Unmarshal(raw, &arr); err != nil {
		// nie wywalaj PDF-a jeśli w jednym dniu JSON jest skopany
		return nil
	}
	return arr
}

func minutesBetweenTS(a, b string) int {
	// a/b: "YYYY-MM-DD HH:MM:SS" (albo z ułamkami)
	// próbujemy kilka formatów
	layouts := []string{
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	}
	var ta, tb time.Time
	var err error

	for _, l := range layouts {
		ta, err = time.ParseInLocation(l, a, time.Local)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0
	}
	for _, l := range layouts {
		tb, err = time.ParseInLocation(l, b, time.Local)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0
	}
	d := tb.Sub(ta)
	return int(d.Minutes())
}

//
// ====================== WYKRES V2 – MonthlyReport ======================
//

// GenerateChartV2 generuje wykres na bazie nowego reportu (du.MonthlyReport).
func GenerateChartV2(pngPath string, report du.MonthlyReport) error {
	bar := charts.NewBar()

	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1100px",
			Height: "420px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Plan / Wymagane / Praca (godziny)",
			Subtitle: fmt.Sprintf("%s — %s %s (#%d)", monthKeyFromReport(report), report.Name, report.Surname, report.UserID),
		}),
		charts.WithAnimation(false),
	)

	xAxis := make([]string, 0, len(report.Days))
	planned := make([]opts.BarData, 0, len(report.Days))
	required := make([]opts.BarData, 0, len(report.Days))
	work := make([]opts.BarData, 0, len(report.Days))

	for _, d := range report.Days {
		day := dateOnly(d.Day)
		label := day
		if len(label) >= 10 {
			label = label[8:10]
		}
		xAxis = append(xAxis, label)
		planned = append(planned, opts.BarData{Value: float64(d.PlanMinutes) / 60.0})
		required = append(required, opts.BarData{Value: float64(d.RequiredMinutes) / 60.0})
		work = append(work, opts.BarData{Value: float64(d.WorkMinutes) / 60.0})
	}

	bar.SetXAxis(xAxis).
		AddSeries("Plan [h]", planned).
		AddSeries("Wymagane [h]", required).
		AddSeries("Praca [h]", work)

	if err := os.MkdirAll(filepath.Dir(pngPath), 0o755); err != nil {
		return err
	}
	return render.MakeChartSnapshot(bar.RenderContent(), pngPath)
}

//
// ====================== BUILD PDF V2 – MonthlyReport ======================
//

func BuildPDFV2(outPath, qrPath, chartPNGPath string, report du.MonthlyReport) error {
	return buildWorktimePDFV2(outPath, qrPath, chartPNGPath, report)
}

func buildWorktimePDFV2(outPath, qrPath, chartPNGPath string, report du.MonthlyReport) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	fontFile := "internal/pdf/font.ttf"
	fontFamily := "Arial"
	if err := ensureFont(fontFile); err == nil {
		pdf.AddUTF8Font("Roboto", "", fontFile)
		fontFamily = "Roboto"
	} else {
		log.Printf("Błąd czcionki: %v. Używam standardowej (bez PL).", err)
	}

	pdf.SetFont(fontFamily, "", 12)
	pageW, pageH := pdf.GetPageSize()
	margin := 20.0

	// ---------- STRONA 1: OKŁADKA ----------
	pdf.AddPage()

	pdf.SetFont(fontFamily, "", 10)
	pdf.SetXY(margin, margin)
	dateStr := fmt.Sprintf("Data wygenerowania: %s", time.Now().Format("2006-01-02 15:04"))
	pdf.Cell(0, 5, dateStr)

	pdf.SetXY(margin, margin+8)
	pdf.Cell(0, 5, fmt.Sprintf("Użytkownik: %s %s , Numer Pracownika - %d", report.Name, report.Surname, report.UserID))
	pdf.Ln(6)
	pdf.SetX(margin)
	pdf.Cell(0, 5, fmt.Sprintf("Zakres: %s — %s", report.FromDate, report.ToDate))
	pdf.Ln(5)
	pdf.SetX(margin)
	pdf.Cell(0, 5, fmt.Sprintf("Miesiąc: %s", monthKeyFromReport(report)))

	centerY := pageH / 2.0
	pdf.SetFont(fontFamily, "", 40)
	title := "System czasu pracy"
	tw := pdf.GetStringWidth(title)
	pdf.SetXY((pageW-tw)/2, centerY-25)
	pdf.Cell(tw, 15, title)

	pdf.SetFont(fontFamily, "", 22)
	sub := "Raport miesięczny"
	sw := pdf.GetStringWidth(sub)
	pdf.SetXY((pageW-sw)/2, centerY+0)
	pdf.Cell(sw, 10, sub)

	pdf.SetLineWidth(0.5)
	pdf.Line(pageW/2-35, centerY+15, pageW/2+35, centerY+15)

	// ---------- STRONA 2: QR ----------
	pdf.AddPage()
	qrOpt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	pdf.RegisterImageOptions(qrPath, qrOpt)

	maxSideW := pageW - 2*margin
	maxSideH := pageH - 2*margin
	qrSide := maxSideW
	if maxSideH < qrSide {
		qrSide = maxSideH
	}
	qrSide *= 0.8
	qrX := (pageW - qrSide) / 2
	qrY := (pageH - qrSide) / 2
	pdf.ImageOptions(qrPath, qrX, qrY, qrSide, qrSide, false, qrOpt, 0, "")

	// ---------- STRONA 3: KPI / PODSUMOWANIE ----------
	addKPISummaryPage(pdf, fontFamily, report, margin)

	// ---------- STRONA 4: ALERTY ----------
	addAlertsPage(pdf, fontFamily, report, margin)

	// ---------- STRONA 5+: WEJŚCIA / WYJŚCIA ----------
	addInOutPages(pdf, fontFamily, report, margin)

	// ---------- KOLEJNA STRONA: WYKRES ----------
	addChartPage(pdf, fontFamily, chartPNGPath, margin)

	// ---------- TABELA DZIENNA ----------
	addDailyTablePages(pdf, fontFamily, report, margin)


	// ---------- SUROWE IN_OUT (NA KOŃCU) ----------
	addRawInOutPages(pdf, fontFamily, report, margin)



	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return pdf.OutputFileAndClose(outPath)
}

func addKPISummaryPage(pdf *gofpdf.Fpdf, fontFamily string, report du.MonthlyReport, margin float64) {
	pdf.AddPage()
	pdf.SetFont(fontFamily, "", 14)
	pdf.SetXY(margin, margin-2)
	pdf.Cell(0, 8, "Podsumowanie (KPI)")

	pdf.Ln(10)
	pdf.SetFont(fontFamily, "", 10)
	pdf.SetX(margin)
	pdf.MultiCell(0, 5,
		fmt.Sprintf("Użytkownik: %s %s (#%d)\nZakres: %s — %s",
			report.Name, report.Surname, report.UserID, report.FromDate, report.ToDate),
		"", "L", false)

	pdf.Ln(4)

	// "kafelki" KPI – nadal bez bold/italic
	pageW, _ := pdf.GetPageSize()
	cardGap := 4.0
	cardW := (pageW - 2*margin - 2*cardGap) / 3
	cardH := 18.0

	type card struct {
		title string
		value string
		note  string
	}

	t := report.Totals
	cards := []card{
		{"Plan miesiąca", formatMinutes(t.MonthPlanMinutes), "plan"},
		{"Wykonane", formatMinutes(t.MonthExecMinutes), "wyk"},
		{"Cel (Nominal x FTE)", formatMinutes(t.MonthTargetMinutes), "cel"},
		{"Delta netto", formatMinutes(t.MonthNetDeltaMinutes), "net"},
		{"Wykonane - Plan", formatMinutes(t.MonthExecMinusPlanMinutes), "wyk vs plan"},
		{"Wykonane - Cel", formatMinutes(t.MonthExecMinusTargetMinutes), "wyk vs cel"},
	}

	drawCard := func(x, y float64, c card) {
		pdf.SetFillColor(245, 245, 245)
		pdf.SetDrawColor(210, 210, 210)
		pdf.Rect(x, y, cardW, cardH, "DF")
		pdf.SetXY(x+3, y+2)
		pdf.SetFont(fontFamily, "", 9)
		pdf.Cell(0, 5, c.title)
		pdf.SetXY(x+3, y+8)
		pdf.SetFont(fontFamily, "", 12)
		pdf.Cell(0, 6, c.value)
		pdf.SetXY(x+3, y+14)
		pdf.SetFont(fontFamily, "", 8)
		pdf.SetTextColor(80, 80, 80)
		pdf.Cell(0, 4, c.note)
		pdf.SetTextColor(0, 0, 0)
	}

	startX := margin
	startY := pdf.GetY() + 2
	for i, c := range cards {
		col := i % 3
		row := i / 3
		x := startX + float64(col)*(cardW+cardGap)
		y := startY + float64(row)*(cardH+cardGap)
		drawCard(x, y, c)
	}

	pdf.SetY(startY + 2*(cardH+cardGap) + 4)
	pdf.SetFont(fontFamily, "", 10)
	pdf.SetX(margin)
	//pdf.Cell(0, 6, fmt.Sprintf("Nominal missing any: %d", t.NominalMissingAny))
}

func addAlertsPage(pdf *gofpdf.Fpdf, fontFamily string, report du.MonthlyReport, margin float64) {
	pdf.AddPage()
	pdf.SetFont(fontFamily, "", 14)
	pdf.SetXY(margin, margin-2)
	pdf.Cell(0, 8, "Alerty i odchylenia")

	pdf.Ln(10)
	pdf.SetFont(fontFamily, "", 10)

	type alertRow struct {
		day   string
		kind  string
		value string
		desc  string
	}

	var rows []alertRow
	for _, d := range report.Days {
		day := dateOnly(d.Day)

		// odpoczynek
		if strings.TrimSpace(d.RestAlert) != "" && d.RestAlert != "OK" {
			rows = append(rows, alertRow{
				day:   day,
				kind: "Odpoczynek",
				value: func() string {
					if d.RestMinutes != nil {
						return fmt.Sprintf("%d min", *d.RestMinutes)
					}
					return ""
				}(),
				desc: d.RestAlert,
			})
		}

		// spóźnienia / wcześniejsze wyjścia
		if d.LateMinutes > 0 {
			rows = append(rows, alertRow{
				day:   day,
				kind:  "Spóźnienie",
				value: fmt.Sprintf("%d min", d.LateMinutes),
				desc:  "Minuty spóźnienia > 0",
			})
		}
		if d.EarlyLeaveMinutes > 0 {
			rows = append(rows, alertRow{
				day:   day,
				kind:  "Wcześniejsze wyjście",
				value: fmt.Sprintf("%d min", d.EarlyLeaveMinutes),
				desc:  "Przedwczesne wyjście > 0",
			})
		}

		// brak normy / nominal missing
		if d.NominalMissing > 0 {
			rows = append(rows, alertRow{
				day:   day,
				kind:  "Nominal missing",
				value: fmt.Sprintf("%d min", d.NominalMissing),
				desc:  "nominal_missing > 0",
			})
		}
	}

	if len(rows) == 0 {
		pdf.SetX(margin)
		pdf.MultiCell(0, 6, "Brak alertów (OK).", "", "L", false)
		return
	}

	// stabilnie sortuj po dacie
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].day == rows[j].day {
			return rows[i].kind < rows[j].kind
		}
		return rows[i].day < rows[j].day
	})

	cols := []struct {
		h string
		w float64
	}{
		{"Data", 22},
		{"Typ", 34},
		{"Wartość", 24},
		{"Opis", 95},
	}

	rowH := 6.0

	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont(fontFamily, "", 11)
	pdf.SetX(margin)
	for _, c := range cols {
		pdf.CellFormat(c.w, rowH, c.h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont(fontFamily, "", 9)
	pdf.SetFillColor(255, 255, 255)

	for _, r := range rows {
		if pdf.GetY()+rowH+margin > 297 {
			pdf.AddPage()
			pdf.SetFont(fontFamily, "", 11)
			pdf.SetFillColor(230, 230, 230)
			pdf.SetX(margin)
			for _, c := range cols {
				pdf.CellFormat(c.w, rowH, c.h, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont(fontFamily, "", 9)
		}

		pdf.SetX(margin)
		pdf.CellFormat(cols[0].w, rowH, r.day, "1", 0, "L", false, 0, "")
		pdf.CellFormat(cols[1].w, rowH, r.kind, "1", 0, "L", false, 0, "")
		pdf.CellFormat(cols[2].w, rowH, r.value, "1", 0, "C", false, 0, "")
		pdf.CellFormat(cols[3].w, rowH, r.desc, "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}
}

func addInOutPages(pdf *gofpdf.Fpdf, fontFamily string, report du.MonthlyReport, margin float64) {
	pageW, pageH := pdf.GetPageSize()
	_ = pageW

	cols := []struct {
		h string
		w float64
	}{
		{"Data", 22},
		{"Dzień", 14},
		{"Wejście", 24},
		{"Wyjście", 24},
		{"Czas", 18},
		{"Status", 78},
	}

	rowH := 6.0

	drawHeader := func() {
		pdf.AddPage()
		pdf.SetFont(fontFamily, "", 13)
		pdf.SetXY(margin, margin-5)
		pdf.Cell(0, 7, "Wejścia / wyjścia")

		pdf.SetY(margin + 3)
		pdf.SetFont(fontFamily, "", 11)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetX(margin)
		for _, c := range cols {
			pdf.CellFormat(c.w, rowH, c.h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont(fontFamily, "", 9)
	}

	first := true

	for _, d := range report.Days {
		pairs := parseInOut(d.InOut)
		if len(pairs) == 0 {
			continue
		}
		if first {
			drawHeader()
			first = false
		}

		day := dateOnly(d.Day)
		dow := weekdayPLFromDow(d.Dow)

		for _, p := range pairs {
			if pdf.GetY()+rowH+margin > pageH {
				drawHeader()
			}

			inHH := timeHHMMFromTS(p.In)
			outHH := timeHHMMFromTS(p.Out)
			durMin := minutesBetweenTS(p.In, p.Out)
			dur := ""
			if durMin > 0 {
				dur = formatMinutes(durMin)
			}

			status := d.Status
			if strings.TrimSpace(d.RestAlert) != "" && d.RestAlert != "OK" {
				status = fmt.Sprintf("%s | ODPOCZYNEK: %s", status, d.RestAlert)
			}
			if d.LateMinutes > 0 {
				status = fmt.Sprintf("%s | SPÓŹNIENIE %d min", status, d.LateMinutes)
			}
			if d.EarlyLeaveMinutes > 0 {
				status = fmt.Sprintf("%s | PRZEDWCZESNE WYJŚCIE %d min", status, d.EarlyLeaveMinutes)
			}

			pdf.SetX(margin)

			// --- FIX: zawijanie statusu w ostatniej kolumnie ---
			statusW := cols[5].w
			lines := pdf.SplitLines([]byte(status), statusW)
			if len(lines) == 0 {
				lines = [][]byte{[]byte("")}
			}
			rowHdyn := rowH * float64(len(lines))

			// jeżeli wiersz się nie mieści, nowa strona z nagłówkiem
			if pdf.GetY()+rowHdyn+margin > pageH {
				drawHeader()
			}

			x := margin
			y := pdf.GetY()

			// kolumny 0..4 normalnie, ale z wysokością rowHdyn
			pdf.SetXY(x, y)
			pdf.CellFormat(cols[0].w, rowHdyn, day, "1", 0, "L", false, 0, "")
			x += cols[0].w

			pdf.SetXY(x, y)
			pdf.CellFormat(cols[1].w, rowHdyn, dow, "1", 0, "C", false, 0, "")
			x += cols[1].w

			pdf.SetXY(x, y)
			pdf.CellFormat(cols[2].w, rowHdyn, inHH, "1", 0, "C", false, 0, "")
			x += cols[2].w

			pdf.SetXY(x, y)
			pdf.CellFormat(cols[3].w, rowHdyn, outHH, "1", 0, "C", false, 0, "")
			x += cols[3].w

			pdf.SetXY(x, y)
			pdf.CellFormat(cols[4].w, rowHdyn, dur, "1", 0, "C", false, 0, "")
			x += cols[4].w

			// status jako MultiCell (zawija)
			pdf.SetXY(x, y)
			pdf.MultiCell(statusW, rowH, status, "1", "L", false)

			pdf.SetXY(margin, y+rowHdyn)

				}
	}

	if first {
		// brak in/out
		pdf.AddPage()
		pdf.SetFont(fontFamily, "", 13)
		pdf.SetXY(margin, margin-5)
		pdf.Cell(0, 7, "Wejścia / wyjścia")
		pdf.Ln(10)
		pdf.SetFont(fontFamily, "", 10)
		pdf.SetX(margin)
		pdf.MultiCell(0, 6, "Brak zarejestrowanych wejść/wyjść w wybranym zakresie.", "", "L", false)
	}
}

func addChartPage(pdf *gofpdf.Fpdf, fontFamily string, chartPNGPath string, margin float64) {
	pdf.AddPage()
	pdf.SetFont(fontFamily, "", 13)
	pdf.SetXY(margin, margin-5)
	pdf.Cell(0, 7, "Wykres")

	pageW, pageH := pdf.GetPageSize()
	chartOpt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	chInfo := pdf.RegisterImageOptions(chartPNGPath, chartOpt)
	if chInfo == nil {
		pdf.SetFont(fontFamily, "", 10)
		pdf.SetXY(margin, margin+10)
		pdf.MultiCell(0, 6, fmt.Sprintf("Nie udało się załadować obrazka wykresu: %s", chartPNGPath), "", "L", false)
		return
	}
	chW, chH := chInfo.Extent()
	aspect := 1.0
	if chW > 0 {
		aspect = chH / chW
	}

	targetW := (pageW - 2*margin) * 0.92
	targetH := targetW * aspect
	maxH := pageH - 2*margin - 10
	if targetH > maxH {
		targetH = maxH
		targetW = targetH / aspect
	}

	chX := (pageW - targetW) / 2
	chY := (pageH - targetH) / 2
	if chY < margin+8 {
		chY = margin + 8
	}
	pdf.ImageOptions(chartPNGPath, chX, chY, targetW, targetH, false, chartOpt, 0, "")
}

func addDailyTablePages(pdf *gofpdf.Fpdf, fontFamily string, report du.MonthlyReport, margin float64) {
	pageW, pageH := pdf.GetPageSize()
	_ = pageW

	cols := []struct {
		h string
		w float64
	}{
		{"Data", 21},
		{"D", 10},
		{"Plan", 16},
		{"Wym.", 16},
		{"Praca", 16},
		{"Przer.", 16},
		{"Noc", 14},
		{"Urlop", 16},
		{"Delta", 16},
		{"Status", 39},
	}

	rowH := 6.0

	drawHeader := func() {
		pdf.AddPage()
		pdf.SetFont(fontFamily, "", 13)
		pdf.SetXY(margin, margin-5)
		pdf.Cell(0, 7, "Tabela dzienna (skrót KPI)")

		pdf.SetY(margin + 3)
		pdf.SetFont(fontFamily, "", 10)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetX(margin)
		for _, c := range cols {
			pdf.CellFormat(c.w, rowH, c.h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont(fontFamily, "", 8)
	}

	drawHeader()

	for _, d := range report.Days {
		day := dateOnly(d.Day)
		dow := weekdayPLFromDow(d.Dow)

		status := d.Status
		if hn := safeStrPtrTrim(d.HolidayName); hn != "" {
			status = "ŚWIĘTO: " + hn
		}
		if d.SickID != nil {
			if lb := safeStrPtrTrim(d.LegalBasis); lb != "" {
				status = "CHOROBOWE (" + lb + ")"
			}
		}

		delta := d.NetDeltaMinutes
		row := []string{
			day,
			dow,
			formatMinutes(d.PlanMinutes),
			formatMinutes(d.RequiredMinutes),
			formatMinutes(d.WorkMinutes),
			formatMinutes(d.BreakMinutes),
			formatMinutes(d.NightMinutes),
			formatMinutes(d.LeaveCreditMinutes),
			formatMinutes(delta),
			status,
		}

		statusW := cols[9].w
		lines := pdf.SplitLines([]byte(row[9]), statusW)
		if len(lines) == 0 {
			lines = [][]byte{[]byte("")}
		}
		rowHdyn := rowH * float64(len(lines))

		// jeżeli cały wiersz nie mieści się na stronie – nowa strona
		if pdf.GetY()+rowHdyn+margin > pageH {
			drawHeader()
		}

		x := margin
		y := pdf.GetY()

		// kolumny 0..8 – zwykłe komórki, ale o wysokości rowHdyn
		for i := 0; i < 9; i++ {
			c := cols[i]
			align := "L"
			if i >= 2 && i <= 8 {
				align = "R"
			}
			if i == 1 {
				align = "C"
			}
			pdf.SetXY(x, y)
			pdf.CellFormat(c.w, rowHdyn, row[i], "1", 0, align, false, 0, "")
			x += c.w
		}

		// kolumna 9 (Status) – MultiCell z zawijaniem
		pdf.SetXY(x, y)
		pdf.MultiCell(statusW, rowH, row[9], "1", "L", false)

		// ustaw pozycję dokładnie na koniec wiersza tabeli (MultiCell przestawia X)
		pdf.SetXY(margin, y+rowHdyn)

		// drugi wiersz informacyjny (bez zmian)
		info := make([]string, 0, 4)
		if strings.TrimSpace(d.ScheduleHours) != "" {
			info = append(info, "Godziny: "+d.ScheduleHours)
		}
		if strings.TrimSpace(d.PlanSource) != "" || strings.TrimSpace(d.PlanType) != "" {
			info = append(info, fmt.Sprintf("Plan: %s / %s", d.PlanSource, d.PlanType))
		}
		if d.LateMinutes > 0 {
			info = append(info, fmt.Sprintf("Late: %d min", d.LateMinutes))
		}
		if d.EarlyLeaveMinutes > 0 {
			info = append(info, fmt.Sprintf("Early: %d min", d.EarlyLeaveMinutes))
		}
		if strings.TrimSpace(d.RestAlert) != "" && d.RestAlert != "OK" {
			info = append(info, "Rest: "+d.RestAlert)
		}

		if len(info) > 0 {
			if pdf.GetY()+rowH+margin > pageH {
				drawHeader()
			}
			pdf.SetX(margin)
			pdf.SetFont(fontFamily, "", 8)
			pdf.CellFormat(cols[0].w+cols[1].w, rowH, "", "1", 0, "L", false, 0, "")
			pdf.CellFormat(0, rowH, strings.Join(info, " | "), "1", 0, "L", false, 0, "")
			pdf.Ln(-1)
			pdf.SetFont(fontFamily, "", 8)
		}
	}
}

//
// ====================== CZCIONKA / QR / HELPERY ======================
//

func ensureFont(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		return nil
	}
	fmt.Println("Pobieram czcionkę Roboto dla polskich znaków...")

	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}

	url := "https://raw.githubusercontent.com/google/fonts/main/ofl/roboto/Roboto-Regular.ttf"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("błąd pobierania: %d", resp.StatusCode)
	}

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func MakeQR(content, outPath string, qrSize, quietZonePx int, logoPath string, logoRatio, logoPadPct float64) error {
	qr, err := qrcode.New(content, qrcode.Highest)
	if err != nil {
		return err
	}
	qr.BackgroundColor = color.White
	qr.ForegroundColor = color.Black
	qrImg := qr.Image(qrSize)

	finalW := qrSize + 2*quietZonePx
	finalH := finalW
	canvas := image.NewRGBA(image.Rect(0, 0, finalW, finalH))
	idraw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, idraw.Src)

	qrDst := image.Rectangle{
		Min: image.Pt(quietZonePx, quietZonePx),
		Max: image.Pt(quietZonePx+qrSize, quietZonePx+qrSize),
	}
	idraw.Draw(canvas, qrDst, qrImg, image.Point{}, idraw.Src)

	if logoPath == "" || logoRatio <= 0 {
		return savePNG(outPath, canvas)
	}

	lf, err := os.Open(logoPath)
	if err != nil {
		log.Printf("Błąd otwarcia logo %s: %v. Generuję bez logo.", logoPath, err)
		return savePNG(outPath, canvas)
	}
	defer lf.Close()
	logoSrc, err := png.Decode(lf)
	if err != nil {
		return err
	}

	targetW := int(float64(qrSize) * logoRatio)
	lb := logoSrc.Bounds()
	lw, lh := lb.Dx(), lb.Dy()
	newW, newH := fitWithin(lw, lh, targetW)

	logo := image.NewRGBA(image.Rect(0, 0, newW, newH))
	xdraw.CatmullRom.Scale(logo, logo.Bounds(), logoSrc, logoSrc.Bounds(), idraw.Over, nil)

	pad := int(float64(newW) * logoPadPct)
	bgRect := image.Rect(0, 0, newW+2*pad, newH+2*pad)
	bg := image.NewRGBA(bgRect)
	idraw.Draw(bg, bg.Bounds(), image.NewUniform(color.White), image.Point{}, idraw.Src)

	cx, cy := finalW/2, finalH/2
	bgDst := image.Rectangle{
		Min: image.Pt(cx-bgRect.Dx()/2, cy-bgRect.Dy()/2),
		Max: image.Pt(cx-bgRect.Dx()/2, cy-bgRect.Dy()/2).Add(bgRect.Size()),
	}
	logoDst := image.Rectangle{
		Min: image.Pt(cx-newW/2, cy-newH/2),
		Max: image.Pt(cx-newW/2, cy-newH/2).Add(logo.Bounds().Size()),
	}

	idraw.Draw(canvas, bgDst, bg, image.Point{}, idraw.Src)
	idraw.Draw(canvas, logoDst, logo, image.Point{}, idraw.Over)

	return savePNG(outPath, canvas)
}

func fitWithin(w, h, target int) (int, int) {
	if w == 0 || h == 0 || target <= 0 {
		return target, target
	}
	if w >= h {
		return target, int(float64(h)*float64(target)/float64(w))
	}
	return int(float64(w)*float64(target)/float64(h)), target
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}



func addRawInOutPages(pdf *gofpdf.Fpdf, fontFamily string, report du.MonthlyReport, margin float64) {
	pageW, pageH := pdf.GetPageSize()
	_ = pageW

	rowH := 6.0

	drawHeader := func() {
		pdf.AddPage()
		pdf.SetFont(fontFamily, "", 13)
		pdf.SetXY(margin, margin-5)
		pdf.Cell(0, 7, "Surowe wejścia wyjścia")

		pdf.SetY(margin + 3)
		pdf.SetFont(fontFamily, "", 10)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetX(margin)
		pdf.CellFormat(26, rowH, "Data", "1", 0, "C", true, 0, "")
		pdf.CellFormat(0, rowH, "wejścia wyjścia (JSON)", "1", 0, "C", true, 0, "")
		pdf.Ln(-1)

		pdf.SetFont(fontFamily, "", 8)
	}

	drawHeader()

	for _, d := range report.Days {
		day := dateOnly(d.Day)

		raw := ""
		if len(d.InOut) == 0 {
			raw = ""
		} else {
			raw = strings.TrimSpace(string(d.InOut))
			// jeśli MySQL zwraca "null"
			if raw == "null" {
				raw = ""
			}
		}

		// jeśli pusto, to też pokażemy (żeby było wiadomo że brak)
		if raw == "" {
			raw = "-"
		}

		// wysokość dla MultiCell (robimy kolumnę 2 jako MultiCell)
		col1W := 26.0
		col2W := (pageW - 2*margin) - col1W

		// jeżeli będziemy schodzić poza stronę — nowa strona
		// (szacujemy: min 1 linia)
		if pdf.GetY()+rowH+margin > pageH {
			drawHeader()
		}

		x := margin
		y := pdf.GetY()

		// kolumna 1 (Data)
		pdf.SetXY(x, y)
		pdf.CellFormat(col1W, rowH, day, "1", 0, "L", false, 0, "")

		// kolumna 2 (raw json) – MultiCell
		pdf.SetXY(x+col1W, y)
		// MultiCell automatycznie przesuwa Y, ale nie rysuje obramowania całej "komórki"
		startY := pdf.GetY()
		pdf.MultiCell(col2W, rowH, raw, "1", "L", false)
		endY := pdf.GetY()

		// wyrównanie: jeżeli MultiCell zrobiło więcej niż 1 linię,
		// musimy dorysować "wysokość" kolumny 1
		h := endY - startY
		if h > rowH {
			// dorysuj brakujące boki dla kolumny 1 (ponieważ miała tylko rowH)
			pdf.Rect(x, startY, col1W, h, "")
		}

		// po MultiCell jesteśmy już na nowym Y; upewnij się że X wraca
		pdf.SetX(margin)

		// jeżeli jesteśmy blisko końca — następny wpis w nowej stronie
		if pdf.GetY()+rowH+margin > pageH {
			drawHeader()
		}
	}
}

