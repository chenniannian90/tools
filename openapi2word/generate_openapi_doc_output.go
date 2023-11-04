package openapi2word

import (
	"fmt"
	"github.com/go-courier/oas"
	"github.com/unidoc/unioffice/document"
)

func (g *GenerateOpenAPIDoc) GenerateOperationOutput(response *oas.Response) error {
	for contentType, resp := range response.ResponseObject.WithContent.Content {
		schema := resp.Schema
		if (len(schema.AllOf)+len(schema.Properties)) == 0 && schema.Refer == nil && schema.Type == "" {
			return nil
		}
		g.doc.AddParagraph().AddRun().AddText("输出：")
		switch contentType {
		case "application/json":
			g.doc.AddParagraph().AddRun().AddText("    输出 JSON 格式的数据，格式如下")
			err := g.GenerateOperationOutputStruct_JSON(schema)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationOutputStruct_JSON(schema *oas.Schema) error {
	table := g.GenTable(80, "参数名", "返回类型", "说明")
	_, err := g.GenerateOperationOutputTableAddRow(table, schema)
	if err != nil {
		return err
	}

	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationOuthOutputSchemasProperties(referIDs ...string) error {
	isOut := map[string]bool{}
	for _, referID := range referIDs {
		if !isOut[referID] {
			isOut[referID] = true

			schema := g.openAPI.Schemas[referID]
			if len(schema.Enum) > 0 {
				err := g.GenerateEnum(referID, schema)
				if err != nil {
					return err
				}
			}

			if (len(schema.AllOf)+len(schema.Properties)) == 0 && schema.Refer == nil && schema.Type != oas.TypeArray {
				continue
			}

			g.doc.AddParagraph().AddRun().AddText(fmt.Sprintf("%s:", referID))

			table := g.GenTable(80, "参数名", "返回类型", "说明")

			_, err := g.GenerateOperationOutputTableAddRow(table, schema)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationOutputTableAddRow(table document.Table, schema *oas.Schema) ([]string, error) {
	referIDList := []string{}
	if _, ok := schema.Reference.Refer.(*oas.ComponentRefer); ok {
		refID := schema.Reference.Refer.(*oas.ComponentRefer).ID
		return g.GenerateOperationOutputTableAddRow(table, g.openAPI.Schemas[refID])
	}
	switch schema.Type {
	case oas.TypeArray:
		row := table.AddRow()
		row.AddCell().AddParagraph().AddRun().AddText("")
		if _, ok := schema.Items.Refer.(*oas.ComponentRefer); ok {
			opID := schema.Items.Refer.(*oas.ComponentRefer).ID
			referIDList = append(referIDList, opID)
			row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(opID))
		} else {
			row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(string(schema.Items.Type)))
		}
		row.AddCell().AddParagraph().AddRun().AddText(schema.Description)
		err := g.GenerateOperationOuthOutputSchemasProperties(referIDList...)
		if err != nil {
			return []string{}, err
		}
	}
	if len(schema.Properties) > 0 {
		for param, propertie := range schema.Properties {
			switch propertie.Type {
			case oas.TypeObject:
				for k, v := range propertie.Properties {
					row := table.AddRow()
					row.AddCell().AddParagraph().AddRun().AddText(k)
					if len(v.AllOf) > 0 {
						if _, ok := v.AllOf[0].Reference.Refer.(*oas.ComponentRefer); ok {
							opID := v.AllOf[0].Reference.Refer.(*oas.ComponentRefer).ID
							referIDList = append(referIDList, opID)
							row.AddCell().AddParagraph().AddRun().AddText(CheckType(opID))
						}
					} else {
						row.AddCell().AddParagraph().AddRun().AddText(CheckType(string(v.Type)))
					}
					row.AddCell().AddParagraph().AddRun().AddText(v.Description)
				}
				break
			case oas.TypeArray:

				row := table.AddRow()
				row.AddCell().AddParagraph().AddRun().AddText(param)

				if _, ok := propertie.Items.Refer.(*oas.ComponentRefer); ok {
					opID := propertie.Items.Refer.(*oas.ComponentRefer).ID
					referIDList = append(referIDList, opID)
					row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(opID))
				} else {
					row.AddCell().AddParagraph().AddRun().AddText(CheckType(string(propertie.Type)))
				}
				row.AddCell().AddParagraph().AddRun().AddText(propertie.Description)

				break
			default:

				row := table.AddRow()
				row.AddCell().AddParagraph().AddRun().AddText(param)
				if len(propertie.AllOf) > 0 {
					if _, ok := propertie.AllOf[0].Reference.Refer.(*oas.ComponentRefer); ok {
						rID := propertie.AllOf[0].Reference.Refer.(*oas.ComponentRefer).ID
						row.AddCell().AddParagraph().AddRun().AddText(CheckType(rID))
						referIDList = append(referIDList, rID)
						row.AddCell().AddParagraph().AddRun().AddText(propertie.AllOf[1].Description)
					}
				} else {
					row.AddCell().AddParagraph().AddRun().AddText(CheckType(string(propertie.Type)))
					row.AddCell().AddParagraph().AddRun().AddText(propertie.Description)
				}
				break
			}

		}

		err := g.GenerateOperationOuthOutputSchemasProperties(referIDList...)
		if err != nil {
			return []string{}, err
		}
	}

	if len(schema.AllOf) > 0 {
		for _, proSchema := range schema.AllOf {
			ids, err := g.GenerateOperationOutputTableAddRow(table, proSchema)
			if err != nil {
				return ids, err
			}

			referIDList = append(referIDList, ids...)
		}
	}

	return referIDList, nil
}
