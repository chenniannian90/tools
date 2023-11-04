package openapi2word

import (
	"encoding/json"
	"github.com/go-courier/sqlx/v2/er"
	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
	"io/ioutil"
	"log"
	"net/http"
)

type Er struct {
	erDBList []*er.ERDatabase
	doc      *document.Document
}

func (er *Er) Document() *document.Document {
	return er.doc
}

func NewEr(urls []string) (*Er, error) {
	var erDBList []*er.ERDatabase
	for _, v := range urls {
		resp, err := http.Get(v)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		erDB := &er.ERDatabase{}
		if err := json.Unmarshal(body, erDB); err != nil {
			return nil, err
		}
		erDBList = append(erDBList, erDB)
	}

	doc := document.New()

	return &Er{
		erDBList: erDBList,
		doc:      doc,
	}, nil
}

func (er *Er) GenerateDoc() error {
	nd := er.doc.Numbering.AddDefinition()
	for _, erDB := range er.erDBList {
		er.GenerateService(nd, erDB)
	}
	if err := er.doc.Validate(); err != nil {
		log.Fatalf("error during validation: %s", err)
		return err
	}

	return nil
}

func (er *Er) GenerateService(nd document.NumberingDefinition, erDatabase *er.ERDatabase) error {
	para := er.doc.AddParagraph()
	para.SetNumberingDefinition(nd)
	para.Properties().SetHeadingLevel(2)
	para.AddRun().AddText(erDatabase.Name)

	paraRelation := er.doc.AddParagraph()
	paraRelation.SetNumberingDefinition(nd)
	paraRelation.Properties().SetHeadingLevel(3)
	paraRelation.AddRun().AddText("数据关系模型")
	er.doc.AddParagraph().AddRun().AddText("	无")

	paraDesc := er.doc.AddParagraph()
	paraDesc.SetNumberingDefinition(nd)
	paraDesc.Properties().SetHeadingLevel(3)
	paraDesc.AddRun().AddText("数据实体说明")
	er.doc.AddParagraph().AddRun().AddText("	无")

	para3 := er.doc.AddParagraph()
	para3.SetNumberingDefinition(nd)
	para3.Properties().SetHeadingLevel(3)
	para3.AddRun().AddText("数据实体属性")

	for _, v := range erDatabase.Tables {
		err := er.GenerateTable(nd, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (er *Er) GenerateTable(nd document.NumberingDefinition, erTable *er.ERTable) error {

	para := er.doc.AddParagraph()
	para.SetNumberingDefinition(nd)
	para.Properties().SetHeadingLevel(4)
	para.AddRun().AddText(erTable.Summary + "定义")

	er.doc.AddParagraph().AddRun().AddText("表名：" + erTable.Name)

	er.GenerateGrid(erTable)
	er.doc.AddParagraph()
	return nil
}

func (er *Er) GenerateGrid(erTable *er.ERTable) error {
	table := er.doc.AddTable()
	table.Properties().SetAlignment(wml.ST_JcTableCenter)
	table.Properties().SetWidthPercent(80)
	borders := table.Properties().Borders()
	borders.SetAll(wml.ST_BorderSingle, color.Auto, 1*measurement.Point)
	hdrRow := table.AddRow()

	cell := hdrRow.AddCell()
	cellPara := cell.AddParagraph()
	cellPara.Properties().SetAlignment(wml.ST_JcCenter)
	cellPara.AddRun().AddText("字段")

	cell = hdrRow.AddCell()
	cellPara = cell.AddParagraph()
	cellPara.Properties().SetAlignment(wml.ST_JcCenter)
	cellPara.AddRun().AddText("类型")

	cell = hdrRow.AddCell()
	cellPara = cell.AddParagraph()
	cellPara.Properties().SetAlignment(wml.ST_JcCenter)
	cellPara.AddRun().AddText("键")

	cell = hdrRow.AddCell()
	cellPara = cell.AddParagraph()
	cellPara.Properties().SetAlignment(wml.ST_JcCenter)
	cellPara.AddRun().AddText("说明")

	for _, param := range erTable.Cols {
		row := table.AddRow()
		row.AddCell().AddParagraph().AddRun().AddText(param.Name)
		row.AddCell().AddParagraph().AddRun().AddText(param.DataType)
		if param.Name == "f_id" {
			row.AddCell().AddParagraph().AddRun().AddText("PK")
		} else {
			row.AddCell().AddParagraph().AddRun().AddText("NK")
		}
		row.AddCell().AddParagraph().AddRun().AddText(param.Summary)
	}
	return nil
}
