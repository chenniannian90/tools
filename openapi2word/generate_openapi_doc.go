package openapi2word

import (
	"encoding/json"
	"fmt"
	fatihcolor "github.com/fatih/color"
	"github.com/go-courier/oas"
	"github.com/unidoc/unioffice/document"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

type GenerateOpenAPIDoc struct {
	HeadingLevel int
	ServiceName  string
	URL          *url.URL
	doc          *document.Document
	openAPI      *oas.OpenAPI
}

func NewGenerateOpenAPIDoc(serviceName string, u *url.URL, headingLevel int) *GenerateOpenAPIDoc {
	return &GenerateOpenAPIDoc{
		ServiceName:  serviceName,
		HeadingLevel: headingLevel,
		URL:          u,
		openAPI:      &oas.OpenAPI{},
		doc:          document.New(),
	}
}

func (g *GenerateOpenAPIDoc) Load() {
	if g.URL == nil {
		panic(fmt.Errorf("missing spec-url or file"))
		return
	}

	if g.URL.Scheme == "file" {
		g.loadByFile()
	} else {
		g.loadBySpecURL()
	}
}

func (g *GenerateOpenAPIDoc) loadByFile() {
	data, err := ioutil.ReadFile(g.URL.Path)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, g.openAPI); err != nil {
		panic(err)
	}
}

func (g *GenerateOpenAPIDoc) loadBySpecURL() {
	hc := http.Client{}
	req, err := http.NewRequest("GET", g.URL.String(), nil)
	if err != nil {
		panic(err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(resp.Body).Decode(g.openAPI); err != nil {
		panic(err)
	}
}

func (g *GenerateOpenAPIDoc) Output(cwd string) {
	rootPath := path.Join(cwd, g.ServiceName+".docx")
	err := g.GenerateClientOpenAPIDoc(rootPath)
	if err != nil {
		panic(err)
	}

	log.Printf("generated client of %s into %s", g.ServiceName, fatihcolor.MagentaString(rootPath))
}

func (g *GenerateOpenAPIDoc) GenerateClientOpenAPIDoc(filename string) error {
	for url, path := range g.openAPI.Paths.Paths {
		for method, operation := range path.Operations.Operations {
			if operation.OperationId == "OpenAPI" || operation.OperationId == "ER" {
				continue
			}

			err := g.GenerateOperationDoc(operation, url, method)
			if err != nil {
				return err
			}
		}
	}

	if err := g.doc.Validate(); err != nil {
		log.Fatalf("error during validation: %s", err)
		return err
	}
	return g.doc.SaveToFile(filename)
}

func (g *GenerateOpenAPIDoc) GenerateOperationDoc(operation *oas.Operation, url string, method oas.HttpMethod) error {
	para := g.doc.AddParagraph()
	para.Properties().SetHeadingLevel(g.HeadingLevel)
	para.AddRun().AddText(operation.Summary)

	g.doc.AddParagraph().AddRun().AddText("方法：")
	g.doc.AddParagraph().AddRun().AddText(fmt.Sprintf("    http://ip:port%s", url))
	g.doc.AddParagraph().AddRun().AddText("请求方式：")
	g.doc.AddParagraph().AddRun().AddText(fmt.Sprintf("    %s", CheckMethod(method)))
	if len(operation.Parameters) > 0 || operation.RequestBody != nil {
		err := g.GenerateOperationInput(operation)
		if err != nil {
			return err
		}
	}

	for statusCode, response := range operation.Responses.Responses {
		switch statusCode {
		case 200, 201:
			err := g.GenerateOperationOutput(response)
			if err != nil {
				return err
			}
		}
	}

	g.doc.AddParagraph()
	return nil
}

// 生成枚举
func (g *GenerateOpenAPIDoc) GenerateEnum(enumType string, schema *oas.Schema) error {
	if _, ok := schema.SpecExtensions.Extensions["x-enum-options"].([]interface{}); !ok {
		return nil
	}
	g.doc.AddParagraph().AddRun().AddText(fmt.Sprintf("%s:", enumType))

	table := g.GenTable(80, "值", "说明")

	options := schema.SpecExtensions.Extensions["x-enum-options"].([]interface{})
	for _, option := range options {
		if _, ok := option.(map[string]interface{}); ok {
			enumKV := option.(map[string]interface{})
			row := table.AddRow()
			row.AddCell().AddParagraph().AddRun().AddText(enumKV["value"].(string))
			row.AddCell().AddParagraph().AddRun().AddText(enumKV["label"].(string))
		}

	}
	return nil
}
