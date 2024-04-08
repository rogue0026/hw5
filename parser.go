package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type MethodLabel struct {
	URL        string
	Auth       bool
	HTTPMethod string
}

type TagVal struct {
	LHS string
	RHS string
}

type Tag struct {
	Key    string
	Values []TagVal
}

type Field struct {
	Name     string
	TypeName string
	Tags     []Tag
}

type MethodInfo struct {
	Label      MethodLabel
	Receiver   Field
	MethodName string
	Parameters []Field
}

type StructInfo struct {
	StructName string
	Fields     []Field
}

// CollectAllInfo собирает всю интересующую нас информацию
func CollectAllInfo(filename string) ([]StructInfo, []MethodInfo) {
	fileSet := token.NewFileSet()
	root, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
		return nil, nil
	}

	allStructs := make([]StructInfo, 0)
	allMethods := make([]MethodInfo, 0)
	ast.Inspect(root, func(n ast.Node) bool {
		switch invObject := n.(type) {
		case *ast.FuncDecl:
			if invObject.Recv != nil { // если t.Recv не nil то это метод и надо проверить у него наличие специальной метки
				if invObject.Doc != nil {
					comment := invObject.Doc.List[len(invObject.Doc.List)-1].Text
					if strings.Contains(comment, "apigen") { // если метка есть, то парсим ее и парсим метод

						labelBegin := strings.Index(comment, "{")
						labelEnd := strings.Index(comment, "}")
						label := comment[labelBegin+1 : labelEnd]

						parsedMethod := ParseAPIMethod(invObject) // парсим метод (по условию структура параметров метода всегда контекст + параметр)
						parsedMethod.Label = ParseMethodLabel(label)
						allMethods = append(allMethods, parsedMethod)
					}
				}
			}
		case *ast.TypeSpec:
			stInf := ParseStruct(invObject)
			allStructs = append(allStructs, *stInf)
		}
		return true
	})

	return allStructs, allMethods
}

// ParseMethodLabel проверяет, есть ли у метода специальная метка, и в случае, если она есть, то парсит её
func ParseMethodLabel(label string) MethodLabel {
	pairs := strings.Split(strings.TrimRight(strings.TrimLeft(label, "{"), "}"), ", ")
	mlInfo := MethodLabel{}
	if len(pairs) == 3 {
		mlInfo.URL = strings.Trim(strings.Split(pairs[0], ": ")[1], `" `)
		authVal := strings.Split(pairs[1], ": ")[1]
		if authVal == "true" {
			mlInfo.Auth = true
		}
		mlInfo.HTTPMethod = strings.Trim(strings.Split(pairs[2], ": ")[1], `"'`)
	} else if len(pairs) == 2 {
		mlInfo.URL = strings.Trim(strings.Split(pairs[0], ": ")[1], `"`)
		authVal := strings.Split(pairs[1], ": ")[1]
		if authVal == "true" {
			mlInfo.Auth = true
		}
	}
	return mlInfo
}

// ParseAPIMethod получает название метода и его параметры (парсит только методы, помеченные специальной меткой)
func ParseAPIMethod(method *ast.FuncDecl) MethodInfo {
	result := MethodInfo{}
	result.MethodName = method.Name.Name
	if method.Recv != nil {
		if method.Recv.List != nil {
			varIdent := method.Recv.List[0].Names[0].Name
			if typeInfo, ok := method.Recv.List[0].Type.(*ast.StarExpr); ok {
				if typeIdentInfo, ok := typeInfo.X.(*ast.Ident); ok {
					result.Receiver.Name = varIdent
					result.Receiver.TypeName = typeIdentInfo.String()
				}
			} else if typeInfo, ok := method.Recv.List[0].Type.(*ast.Ident); ok {
				result.Receiver.Name = varIdent
				result.Receiver.TypeName = typeInfo.Name
			}
		}
	}
	for _, f := range method.Recv.List {

		if ti, ok := f.Type.(*ast.Ident); ok {
			receiverName := f.Names[0].String()
			receiverType := ti.Name
			result.Receiver.Name = receiverName
			result.Receiver.TypeName = receiverType
		}
	}

	if method.Type.Params.List != nil {
		for _, field := range method.Type.Params.List {
			fieldName := field.Names[0].String()
			if ti, ok := field.Type.(*ast.SelectorExpr); ok { // здесь пробуем получить тип поля, объявленный в другом пакете
				typeName := ti.Sel.String() // здесь должно быть название типа
				if packName, ok := ti.X.(*ast.Ident); ok {
					pName := packName.String()
					result.Parameters = append(result.Parameters,
						Field{Name: fieldName, TypeName: pName + "." + typeName})
				}
			}
			if ti, ok := field.Type.(*ast.Ident); ok {
				typename := ti.String()
				result.Parameters = append(result.Parameters,
					Field{Name: fieldName, TypeName: typename})
			}
		}
	}
	return result
}

func ParseStruct(st *ast.TypeSpec) *StructInfo {
	infoAboutStruct := StructInfo{}
	infoAboutStruct.StructName = st.Name.Name

	if t, ok := st.Type.(*ast.StructType); ok {
		for _, f := range t.Fields.List {
			fld := Field{}
			fld.Name = f.Names[0].String()
			if f.Tag != nil {
				fld.Tags = ParseTagInfo(f.Tag.Value)
			}
			if tpInf, ok := f.Type.(*ast.Ident); ok {
				fld.TypeName = tpInf.String()
			}
			infoAboutStruct.Fields = append(infoAboutStruct.Fields, fld)
		}
	}

	return &infoAboutStruct
}

func ParseTagInfo(fieldTag string) []Tag {
	tags := strings.Split(fieldTag, " ")
	result := make([]Tag, 0)
	for _, currentTag := range tags {
		currentTag = strings.ReplaceAll(currentTag, "`", "")
		tagInformation := Tag{}
		tokens := strings.Split(currentTag, ":")
		key := strings.Trim(tokens[0], " ")
		tagInformation.Key = key
		tagVal := strings.ReplaceAll(tokens[1], "\"", "")
		values := strings.Split(tagVal, ",")
		for _, v := range values {
			if strings.Contains(v, "=") {
				kv := strings.Split(v, "=")
				lhs := kv[0]
				rhs := kv[1]
				tagValue := TagVal{LHS: lhs, RHS: rhs}
				tagInformation.Values = append(tagInformation.Values, tagValue)
			} else {
				tagValue := TagVal{LHS: v, RHS: ""}
				tagInformation.Values = append(tagInformation.Values, tagValue)
			}
		}
		result = append(result, tagInformation)
	}

	return result
}
