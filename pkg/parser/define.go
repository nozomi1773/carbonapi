package parser

import (
	"strings"
	"text/template"
)

type defineImpl struct {
	tpl *template.Template
}

type Define interface {
	Expand(Expr) (Expr, error)
}

// NewDefine compiles define templates
func NewDefine(defineMap map[string]string) (Define, error) {
	tpl := template.New("define")

	for name, value := range defineMap {
		t := tpl.New(name)
		_, err := t.Parse(value)
		if err != nil {
			return nil, err
		}
	}

	return &defineImpl{tpl: tpl}, nil
}

func (d *defineImpl) expandExpr(exp *expr) (*expr, error) {
	if exp == nil {
		return exp, nil
	}

	var err error

	if exp.etype == EtName || exp.etype == EtFunc {
		t := d.tpl.Lookup(exp.target)
		if t != nil {
			var b strings.Builder
			args := make([]string, len(exp.args))
			for i := 0; i < len(exp.args); i++ {
				args[i] = exp.args[i].ToString()
			}
			kwargs := make(map[string]string)
			for k, v := range exp.namedArgs {
				kwargs[k] = v.ToString()
			}
			data := map[string]interface{}{
				"argString": exp.argString,
				"args":      args,
				"kwargs":    kwargs,
			}
			err = t.Execute(&b, data)
			if err != nil {
				return exp, err
			}
			newExp, _, err := ParseExpr(b.String())
			if err != nil {
				return exp, err
			}
			exp = newExp.(*expr)
		}
	}

	for i := 0; i < len(exp.args); i++ {
		exp.args[i], err = d.expandExpr(exp.args[i])
		if err != nil {
			return exp, err
		}
	}

	for k, v := range exp.namedArgs {
		exp.namedArgs[k], err = d.expandExpr(v)
		if err != nil {
			return exp, err
		}
	}

	return exp, nil
}

func (d *defineImpl) Expand(v Expr) (Expr, error) {
	exp, ok := v.(*expr)
	if !ok {
		return v, nil
	}
	return d.expandExpr(exp)
}
