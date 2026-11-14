package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"erpbadger/internal/domain/schedule"
	pdfpkg "erpbadger/internal/pdf"
)

type MonthlyPDFGenerator struct {
	baseDir string
}

func NewMonthlyPDFGenerator(baseDir string) schedule.ReportGenerator {
	if baseDir == "" {
		baseDir = filepath.Join("tmp", "raports")
	}
	return &MonthlyPDFGenerator{baseDir: baseDir}
}

func (g *MonthlyPDFGenerator) GenerateMonthlyPDF(ctx context.Context, report schedule.MonthlyReport, userIDStr, month string) (string, error) {
	raportsDir := g.baseDir
	outputDir := filepath.Join(raportsDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("błąd tworzenia folderu raportów: %w", err)
	}

	content := fmt.Sprintf("https://erpbadger.com/app/schedule/raport/verify?user=%s&month=%s", userIDStr, month)
	qrSize := 512
	quietZonePx := 20

	qrPlainOut := filepath.Join(raportsDir, "qr_plain.png")
	chartOut := filepath.Join(raportsDir, "chart.png")

	// 1. Wygeneruj QR czysty
	if err := pdfpkg.MakeQR(content, qrPlainOut, qrSize, quietZonePx, "", 0, 0); err != nil {
		return "", fmt.Errorf("błąd generowania czystego QR: %w", err)
	}

	// 2. Opcjonalne logo w QR
	logoFile := filepath.Join("internal", "pdf", "logo.png")
	qrForPDF := qrPlainOut
	if _, err := os.Stat(logoFile); err == nil {
		qrWithLogo := filepath.Join(raportsDir, "qr_logo_v3.png")
		if err := pdfpkg.MakeQR(content, qrWithLogo, qrSize, quietZonePx, logoFile, 0.20, 0.24); err == nil {
			qrForPDF = qrWithLogo
		}
	}

	// 3. Wykres KPI
	if err := pdfpkg.GenerateChartV2(chartOut, report); err != nil {
		return "", fmt.Errorf("błąd generowania wykresu: %w", err)
	}

	// 4. Budowa finalnego pliku PDF
	pdfOut := filepath.Join(outputDir, fmt.Sprintf("raport_%s_%s.pdf", userIDStr, month))
	if err := pdfpkg.BuildPDFV2(pdfOut, qrForPDF, chartOut, report); err != nil {
		return "", fmt.Errorf("błąd budowania PDF: %w", err)
	}

	return pdfOut, nil
}
