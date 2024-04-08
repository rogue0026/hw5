package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func GenerateCode(readFileName, writeFileName string) error {
	resultFile, err := os.Create("generated_file.go")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer resultFile.Close()

	StructsInfo, MethodsInfo := CollectAllInfo(readFileName)
	for _, methInf := range MethodsInfo {
		generatedFile, err := generateHandler(methInf, StructsInfo)
		if err != nil {
			return err
		} else {
			resultFile.WriteString(generatedFile)
		}

	}
	return nil
}

func generateHandler(methInfo MethodInfo, allStructs []StructInfo) (string, error) {

	// создаем строку, куда будем писать наш сгенерированный обработчик
	generatedFile := strings.Builder{}

	// добавляем заголовок обработчика
	generatedFile.WriteString(fmt.Sprintf("func handler%s(w http.ResponseWriter, r *http.Request) {\n", methInfo.MethodName))

	// из информаии о методе, для которого нужно сгенерировать обработчик, берем название последнего параметра для создания структуры нужного типа и создаем структуру
	nameLastMethodParameter := methInfo.Parameters[len(methInfo.Parameters)-1].Name
	generatedFile.WriteString(fmt.Sprintf("params := %s{}\n", nameLastMethodParameter))

	structInfo := StructInfo{}
	// ищем информацию о структуре, содержащей параметры, необходимые для передачи в метод
	for _, el := range allStructs {
		if el.StructName == nameLastMethodParameter {
			structInfo = el
			break
		}
	}

	// заполняем поля структуры, которая будет передаваться в метод
	for _, field := range structInfo.Fields {
		for _, tag := range field.Tags {
			for _, val := range tag.Values {
				if val.LHS == "paramname" {
					generatedFile.WriteString(fmt.Sprintf("params.%s = r.URL.Query().Get(\"%s\")\n", field.Name, val.RHS))
				}
			}
		}
		generatedFile.WriteString(fmt.Sprintf("params.%s = r.URL.Query().Get(\"%s\")\n", field.Name, strings.ToLower(field.Name)))
	}
	/*
		type CreateParams struct {
			Login  string `apivalidator:"required,min=10"`
			Name   string `apivalidator:"paramname=full_name"`
			Status string `apivalidator:"enum=user|moderator|admin,default=user"`
			Age    int    `apivalidator:"min=0,max=128"`
		}
	*/
	// делаем валидацию параметров прежде чем передавать структуру в метод
	//for _, field := range structInfo.Fields {
	//	for _, tag := range field.Tags {
	//		for _, val := range tag.Values {
	//			if val == "required" {
	//				if field.TypeName == "string" {
	//					fmt.Sprintf("if params.%s == \"\" {\n", field.Name)
	//				}
	//
	//			}
	//		}
	//	}
	//}
	return generatedFile.String(), nil
}

// required, min, max, paramname, enum, default
func GenerateValidation(s StructInfo) error {
	//validation := strings.Builder{}
	for _, f := range s.Fields {
		if f.TypeName == "string" {
			for _, tag := range f.Tags {
				for _, tagVal := range tag.Values {
					switch tagVal.LHS {
					case "default":
						fmt.Sprintf("if params.%s == \"\" {\n\tparams.%s = %s", f.Name, f.Name, tagVal.RHS)
					case "required":
						fmt.Sprintf("if params.%s == \"\" {\n\treturn ApiError{HTTPStatus: http.StatusBadRequest, Err: errors.New(\"%s must be not empty\")\n}\n\n", f.Name, strings.ToLower(f.Name))
					case "min": // %v len must be >= %v
						fmt.Sprintf("if len(params.%s) < %v {\n\treturn ApiError{HTTPStatus: http.StatusBadRequest, Err: errors.New(\"%s len must be >= %v\")}\n}", f.Name, tagVal.LHS, strings.ToLower(f.Name), tagVal.RHS)
					case "max": // %v len must be <= %v
						fmt.Sprintf("if len(params.%s) > %v {\n\treturn ApiError{HTTPStatus: http.StatusBadRequest, Err: errors.New(\"%s len must be <= %v\")}\n}", f.Name, tagVal.LHS, strings.ToLower(f.Name), tagVal.RHS)
					case "paramname":
					// todo think about it

					case "enum":
					}

				}
			}
		}
	}

	return nil
}

func universalHandler(w http.ResponseWriter, r *http.Request) {

}
