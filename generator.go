package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

/*
 1. пройтись по всему коду и собрать всю нужную информацию
    Если при разборе кода встретилась структура с метками apivalidator, то нужно собрать всю необходимую информацию об этой структуре
    Информацию о функции собирать в случае если у нее есть коментарий начинающийся с apigen
 2. тегов валидатора может быть несколько для каждого поля, соответственно нужно получить левую часть тега (apivalidator) и правую,
    которая может представлять собой набор значений. Значение тега в свою очередь может представлять собой пару ключ=значение или просто ключ (required например)
*/

type Label struct {
	URL    string
	Auth   bool
	Method string
}

type Field struct {
	Name string
	Type string
}

type Method struct {
	Name       string
	Parameters []Field
}

// CollectMethodsInfo собирает информацию о методах, для которых нужно сгенерировать код
func CollectMethodsInfo(filename string) {
	fileSet := token.NewFileSet()
	root, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err.Error())
		return // todo
	}
	for _, d := range root.Decls {
		switch v := d.(type) {
		case *ast.FuncDecl:
			if v.Recv != nil { // значит это метод
				if v.Doc != nil {
					label := v.Doc.Text()
					if strings.Index(label, "{") != -1 { // проверяем, есть ли в комментарии метка для генерации кода
						label = label[strings.Index(label, "{")+1 : strings.Index(label, "}")]
						_ = ParseMethodLabel(label) // todo

						methodInfo := ParseAPIMethod(v)
						fmt.Println(*methodInfo)
					}
				}
			}
		}
	}
}

func ParseAPIMethod(method *ast.FuncDecl) *Method {
	inf := &Method{}
	inf.Name = method.Name.Name
	if params := method.Type.TypeParams; params != nil {
		for _, el := range params.List {
			f := Field{}
			f.Name = el.Names[0].Name
			if TypeInfo, ok := el.Type.(*ast.Ident); ok {
				f.Type = TypeInfo.Name
			}
			inf.Parameters = append(inf.Parameters, f)
		}
	}
	return inf
}

func ParseMethodLabel(label string) *Label {
	pairs := strings.Split(strings.TrimRight(strings.TrimLeft(label, "{"), "}"), ", ")
	mlInfo := &Label{}
	if len(pairs) == 3 {
		mlInfo.URL = strings.Trim(strings.Split(pairs[0], ": ")[1], `"`)
		authVal := strings.Split(pairs[1], ": ")[1]
		if authVal == "true" {
			mlInfo.Auth = true
		}
		mlInfo.Method = strings.Trim(strings.Split(pairs[2], ": ")[1], `"'`)
	} else if len(pairs) == 2 {
		mlInfo.URL = strings.Trim(strings.Split(pairs[0], ": ")[1], `"`)
		authVal := strings.Split(pairs[1], ": ")[1]
		if authVal == "true" {
			mlInfo.Auth = true
		}
	}
	return mlInfo
}
