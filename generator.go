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

type MethodLabelInfo struct {
	URL    string
	Auth   bool
	Method string
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
			if v.Recv != nil {
				if v.Doc != nil {
					label := v.Doc.Text()
					if strings.Index(label, "{") != -1 { // определяем, является ли комментарий меткой для генерации кода
						label = label[strings.Index(label, "{"):]
						ParseMethodLabel(label)
					}
				}
			}
		}
	}
}

func ParseMethodLabel(label string) *MethodLabelInfo {
	pairs := strings.TrimRightstrings.TrimLeft(label, "{")
	fmt.Print(pairs)
	return nil
}
