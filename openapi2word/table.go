package openapi2word

import (
	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
)

func (g *GenerateOpenAPIDoc) GenTable(pct float64, cellNames ...string) document.Table {
	table := g.doc.AddTable()
	table.Properties().SetAlignment(wml.ST_JcTableCenter)
	table.Properties().SetWidthPercent(80)
	borders := table.Properties().Borders()
	borders.SetAll(wml.ST_BorderSingle, color.Auto, 1*measurement.Point)
	hdrRow := table.AddRow()

	for _, cellName := range cellNames {
		cell := hdrRow.AddCell()
		cellPara := cell.AddParagraph()
		cellPara.Properties().SetAlignment(wml.ST_JcCenter)
		cellPara.AddRun().AddText(cellName)
	}

	return table
}
