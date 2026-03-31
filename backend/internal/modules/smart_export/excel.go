package smart_export

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

const sheetDefault = "Отчёт"

// GenerateXLSX builds a styled .xlsx workbook from query results.
func GenerateXLSX(qr *QueryResult, sheetName string) ([]byte, error) {
	if sheetName == "" {
		sheetName = sheetDefault
	}

	f := excelize.NewFile()
	defer f.Close()

	idx, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(idx)
	// Remove default "Sheet1" if it differs from our sheet name.
	if sheetName != "Sheet1" {
		_ = f.DeleteSheet("Sheet1")
	}

	// ---- styles ----
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "bottom", Color: "2F5496", Style: 2},
		},
	})

	dateStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt:    22, // dd.mm.yyyy hh:mm
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	currencyStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // #,##0.00
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 1, // 0
	})

	// even row stripe
	stripeStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"D9E2F3"}},
	})

	// ---- headers ----
	for colIdx, col := range qr.Columns {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		_ = f.SetCellValue(sheetName, cell, col.Label)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// ---- row height for header ----
	_ = f.SetRowHeight(sheetName, 1, 28)

	// ---- data ----
	colWidths := make([]float64, len(qr.Columns))
	for i, col := range qr.Columns {
		colWidths[i] = float64(len([]rune(col.Label))) + 2
	}

	for rowIdx, row := range qr.Rows {
		excelRow := rowIdx + 2 // 1-based, skip header
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, excelRow)
			colDef := qr.Columns[colIdx]

			displayVal := setCellTyped(f, sheetName, cell, val, colDef, dateStyle, currencyStyle, numberStyle)

			// track width
			w := float64(len([]rune(displayVal))) + 2
			if w > colWidths[colIdx] {
				colWidths[colIdx] = w
			}
		}

		// stripe even rows
		if rowIdx%2 == 1 {
			startCell, _ := excelize.CoordinatesToCellName(1, excelRow)
			endCell, _ := excelize.CoordinatesToCellName(len(qr.Columns), excelRow)
			_ = f.SetCellStyle(sheetName, startCell, endCell, stripeStyle)
		}
	}

	// ---- column widths (capped) ----
	for i, w := range colWidths {
		if w > 50 {
			w = 50
		}
		if w < 10 {
			w = 10
		}
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheetName, colName, colName, w)
	}

	// ---- freeze header row ----
	_ = f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// ---- autofilter ----
	if len(qr.Columns) > 0 && len(qr.Rows) > 0 {
		lastCol, _ := excelize.ColumnNumberToName(len(qr.Columns))
		lastRow := len(qr.Rows) + 1
		_ = f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", lastCol, lastRow), nil)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// setCellTyped writes a typed value to the cell and returns the display string for width estimation.
func setCellTyped(f *excelize.File, sheet, cell string, val any, col ColumnDef, dateStyle, currencyStyle, numberStyle int) string {
	if val == nil {
		_ = f.SetCellValue(sheet, cell, "")
		return ""
	}

	switch col.Type {
	case "date":
		switch v := val.(type) {
		case time.Time:
			if v.IsZero() {
				_ = f.SetCellValue(sheet, cell, "")
				return ""
			}
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, dateStyle)
			return v.Format("02.01.2006 15:04")
		case string:
			_ = f.SetCellValue(sheet, cell, v)
			return v
		default:
			s := fmt.Sprint(val)
			_ = f.SetCellValue(sheet, cell, s)
			return s
		}

	case "currency", "number":
		switch v := val.(type) {
		case float64:
			_ = f.SetCellValue(sheet, cell, v)
			if col.Type == "currency" {
				_ = f.SetCellStyle(sheet, cell, cell, currencyStyle)
			} else {
				_ = f.SetCellStyle(sheet, cell, cell, numberStyle)
			}
			return fmt.Sprintf("%.2f", v)
		case int64:
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, numberStyle)
			return fmt.Sprintf("%d", v)
		case int:
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, numberStyle)
			return fmt.Sprintf("%d", v)
		case []byte:
			s := string(v)
			_ = f.SetCellValue(sheet, cell, s)
			return s
		default:
			s := fmt.Sprint(val)
			_ = f.SetCellValue(sheet, cell, s)
			return s
		}

	default: // string
		switch v := val.(type) {
		case string:
			_ = f.SetCellValue(sheet, cell, v)
			return v
		case []byte:
			s := string(v)
			_ = f.SetCellValue(sheet, cell, s)
			return s
		default:
			s := fmt.Sprint(val)
			_ = f.SetCellValue(sheet, cell, s)
			return s
		}
	}
}
