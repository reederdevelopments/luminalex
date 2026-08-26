package core

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

func (a *App) ExportTableToExcel(payload ExportPayload) error {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Export"
	f.SetSheetName("Sheet1", sheetName)

	for colIdx, header := range payload.Headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	for rowIdx, row := range payload.Rows {
		for colIdx, field := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, field)
		}
	}

	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Data",
		DefaultFilename: fmt.Sprintf("%s_export.xlsx", payload.Filename),
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})

	if err != nil {
		return fmt.Errorf("dialog error: %w", err)
	}

	if filePath == "" {
		return nil
	}

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to save excel file: %w", err)
	}

	return nil
}
