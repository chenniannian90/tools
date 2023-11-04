package openapi2word

import (
	"fmt"
	"github.com/go-courier/oas"
	"github.com/unidoc/unioffice/document"
)

func (g *GenerateOpenAPIDoc) GenerateOperationInput(operation *oas.Operation) error {
	referIDList := []string{}
	if len(operation.Parameters) > 0 {
		g.doc.AddParagraph().AddRun().AddText("输入：")
		table := g.GenTable(90, "参数名", "是否必填", "位置", "类型", "说明")

		for _, param := range operation.Parameters {
			row := table.AddRow()
			row.AddCell().AddParagraph().AddRun().AddText(param.Name)
			if param.Required {
				row.AddCell().AddParagraph().AddRun().AddText("是")
			} else {
				row.AddCell().AddParagraph().AddRun().AddText("否")
			}
			row.AddCell().AddParagraph().AddRun().AddText(string(param.In))

			switch param.Schema.Type {
			case oas.TypeArray:
				row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(string(param.Schema.Items.Type)))
				break
			default:
				if len(param.Schema.AllOf) > 0 {
					opID := param.Schema.AllOf[0].Reference.Refer.(*oas.ComponentRefer).ID
					referIDList = append(referIDList, opID)
					row.AddCell().AddParagraph().AddRun().AddText(CheckType(opID))

				} else {
					row.AddCell().AddParagraph().AddRun().AddText(CheckType(string(param.Schema.Type)))
				}
			}

			if len(param.Schema.AllOf) > 0 {
				if _, ok := param.Schema.AllOf[0].Reference.Refer.(*oas.ComponentRefer); ok {
					row.AddCell().AddParagraph().AddRun().AddText(param.Schema.AllOf[1].Description)
				} else {
					row.AddCell().AddParagraph().AddRun().AddText(param.Schema.Description)
				}
			} else {
				row.AddCell().AddParagraph().AddRun().AddText(param.Schema.Description)
			}
		}
	}
	err := g.GenerateOperationInputOuthSchemaTable(referIDList...)
	if err != nil {
		return err
	}
	if operation.RequestBody != nil {
		for contentType, mediaType := range operation.RequestBody.Content {
			switch contentType {
			case "application/json":
				err := g.GenerateOperationInputRequestBody_JSON(mediaType)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationInputRequestBody_JSON(mediaType *oas.MediaType) error {
	jms := mediaType.MediaTypeObject.Schema
	if (len(jms.AllOf)+len(jms.Properties)) == 0 && jms.Refer == nil && jms.Type == "" {
		return nil
	}
	g.doc.AddParagraph().AddRun().AddText("Body JSON 输入：")

	if len(jms.AllOf) > 0 {
		if _, ok := jms.AllOf[0].Reference.Refer.(*oas.ComponentRefer); ok {
			rID := jms.AllOf[0].Reference.Refer.(*oas.ComponentRefer).ID
			err := g.GenerateOperationInputBodyStruct(rID, nil, true)
			if err != nil {
				return err
			}

		}
	}

	if len(jms.Properties) > 0 {
		err := g.GenerateOperationInputBodyStruct("", jms, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationInputBodyStruct(referID string, schema *oas.Schema, first bool, addStructList ...string) error {
	if referID != "" {
		schema = g.openAPI.Schemas[referID]
	}

	if len(schema.Enum) > 0 {
		err := g.GenerateEnum(referID, schema)
		if err != nil {
			return err
		}
	}

	if (len(schema.AllOf)+len(schema.Properties)) == 0 && schema.Refer == nil && schema.Type != oas.TypeArray {
		return nil
	}

	if referID != "" || !first {
		g.doc.AddParagraph().AddRun().AddText(fmt.Sprintf("%s:", referID))
	}

	table := g.GenTable(80, "参数名", "是否必选", "类型", "说明")

	referIDList := []string{}
	if len(schema.Properties) > 0 {
		for key, propertie := range schema.Properties {
			row := table.AddRow()
			row.AddCell().AddParagraph().AddRun().AddText(key)
			isRequired := false
			for _, v := range schema.Required {
				if v == key {
					isRequired = true
				}
			}
			if isRequired {
				row.AddCell().AddParagraph().AddRun().AddText("是")
			} else {
				row.AddCell().AddParagraph().AddRun().AddText("否")
			}

			switch propertie.Type {
			case oas.TypeArray:
				if _, ok := propertie.Items.Refer.(*oas.ComponentRefer); ok {
					opID := propertie.Items.Refer.(*oas.ComponentRefer).ID
					referIDList = append(referIDList, opID)
					row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(opID))
				} else {
					row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(string(propertie.Items.Type)))
				}
				row.AddCell().AddParagraph().AddRun().AddText(propertie.Description)

				break

			default:
				if len(propertie.AllOf) > 0 {
					opID := propertie.AllOf[0].Reference.Refer.(*oas.ComponentRefer).ID
					referIDList = append(referIDList, opID)
					row.AddCell().AddParagraph().AddRun().AddText(CheckType(opID))
					row.AddCell().AddParagraph().AddRun().AddText(propertie.AllOf[1].Description)

				} else {
					row.AddCell().AddParagraph().AddRun().AddText(CheckType(string(propertie.Type)))
					row.AddCell().AddParagraph().AddRun().AddText(propertie.Description)
				}
			}
		}
	}

	if schema.Type == oas.TypeArray {
		err := g.GenerateOperationInputTableAddRow(table, schema)
		if err != nil {
			return err
		}
	}
	if len(schema.AllOf) > 0 {
		err := g.GenerateOperationInputTableAddRow(table, schema.AllOf[0])
		if err != nil {
			return err
		}
	}
	err := g.GenerateOperationInputOuthSchemaTable(referIDList...)
	if err != nil {
		return err
	}

	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationInputOuthSchemaTable(referIDs ...string) error {
	for _, referID := range referIDs {
		err := g.GenerateOperationInputBodyStruct(referID, nil, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GenerateOpenAPIDoc) GenerateOperationInputTableAddRow(table document.Table, schema *oas.Schema) error {
	if _, ok := schema.Reference.Refer.(*oas.ComponentRefer); ok {
		refID := schema.Reference.Refer.(*oas.ComponentRefer).ID
		return g.GenerateOperationInputTableAddRow(table, g.openAPI.Schemas[refID])
	}
	referIDList := []string{}
	if len(schema.Properties) > 0 {
		for key, propertie := range schema.Properties {

			row := table.AddRow()
			row.AddCell().AddParagraph().AddRun().AddText(key)

			isRequired := false
			for _, v := range schema.Required {
				if v == key {
					isRequired = true
				}
			}
			if isRequired {
				row.AddCell().AddParagraph().AddRun().AddText("是")
			} else {
				row.AddCell().AddParagraph().AddRun().AddText("否")
			}

			row.AddCell().AddParagraph().AddRun().AddText(CheckType(string(propertie.Type)))
			if len(propertie.AllOf) > 0 {
				if _, ok := propertie.AllOf[0].Reference.Refer.(*oas.ComponentRefer); ok {
					row.AddCell().AddParagraph().AddRun().AddText(propertie.AllOf[1].Description)
				} else {
					row.AddCell().AddParagraph().AddRun().AddText(propertie.Description)
				}
			} else {
				row.AddCell().AddParagraph().AddRun().AddText(propertie.Description)
			}
		}
	}
	switch schema.Type {
	case oas.TypeArray:
		row := table.AddRow()
		row.AddCell().AddParagraph().AddRun().AddText("")
		row.AddCell().AddParagraph().AddRun().AddText("是")

		if _, ok := schema.Items.Refer.(*oas.ComponentRefer); ok {
			opID := schema.Items.Refer.(*oas.ComponentRefer).ID
			referIDList = append(referIDList, opID)
			row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(opID))
		} else {
			row.AddCell().AddParagraph().AddRun().AddText("[]" + CheckType(string(schema.Items.Type)))
		}
		row.AddCell().AddParagraph().AddRun().AddText(schema.Description)

		break
	}

	err := g.GenerateOperationInputOuthSchemaTable(referIDList...)
	if err != nil {
		return err
	}

	return nil
}
