package form_parser

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type StageInvalidError struct {
	// Field поле
	Field string
	// ParseError ошибка парсинга
	ParseError string
}

type FormParser interface {
	SetNext(FormParser) FormParser
	Execute(str string, p any) (*StageInvalidError, error)
	Format([]string) []string
}

type Stage struct {
	// field поле
	field string
	// tagName имя тега(значение тега)
	tagName string
	// placeHolder placeholder
	placeHolder string
	// noUseFieldInFormat флаг отсутствия в форматировании
	noUseFieldInFormat bool
	// next следующая стадия
	next FormParser
}

func NewStage(title, tagName, placeHolder string, noUseFormat ...bool) *Stage {
	needFormat := len(noUseFormat) > 0

	return &Stage{
		field:              title,
		tagName:            tagName,
		placeHolder:        placeHolder,
		noUseFieldInFormat: needFormat,
	}
}

func (p *Stage) SetNext(pp FormParser) FormParser {
	p.next = pp

	return pp
}

var fieldTag = "field"

func (p *Stage) Execute(str string, pr any) (*StageInvalidError, error) {
	rgxp, err := regexp.Compile(p.field + `:\s*([^,\n]+)`)
	if err != nil {
		return nil, err
	}

	value := rgxp.FindAllString(str, -1)
	if len(value) == 0 {
		return &StageInvalidError{
			Field:      p.field,
			ParseError: "Значение не заполнено",
		}, nil
	}

	v := strings.TrimSpace(strings.TrimPrefix(value[0], p.field+":"))
	vo := reflect.ValueOf(pr)
	to := reflect.TypeOf(pr)
	for i := 0; i < vo.Elem().NumField(); i++ {
		field := vo.Elem().Field(i)
		fieldType := to.Elem().Field(i)

		val := fieldType.Tag.Get(fieldTag)
		if val == "" || p.tagName != val {
			continue
		}

		if err := set(field, v); err != nil {
			return nil, err
		}
		field.SetString(v)
		break
	}

	if p.next == nil {
		return nil, nil
	}

	return p.next.Execute(str, pr)

}

func (p *Stage) Format(str []string) []string {
	if !p.noUseFieldInFormat {
		str = append(str, p.field+": "+p.placeHolder)
	}
	if p.next == nil {
		return str
	}

	return p.next.Format(str)
}

func set(field reflect.Value, value string) error {
	switch field.Type().Kind() {
	case reflect.Pointer:
		v, ok := field.Interface().(interface {
			FromString(s string) error
		})
		if !ok {
			return fmt.Errorf(`field "%s" is not a pointer`, field.String())
		}

		err := v.FromString(value)
		if err != nil {
			return fmt.Errorf("error parse %s, %w", field.String(), err)
		}

		field.Set(reflect.ValueOf(v))
		return nil
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		if err := setBool(field, value); err != nil {
			return fmt.Errorf("error parse %s, %w", field.String(), err)
		}
		return nil
	case reflect.Int64, reflect.Int:
		if value == "" {
			return fmt.Errorf(`field "%s" is empty`, field.String())
		}

		intEnvValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", value, err)
		}

		field.SetInt(intEnvValue)
		return nil
	default:
		return fmt.Errorf("unprocessed type %s", field.Type().Kind())
	}

	return nil
}

func setBool(value reflect.Value, val string) error {
	if val == "" {
		return fmt.Errorf(`field "%s" is empty`, value.String())
	}

	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", val, err)
	}
	value.SetBool(boolValue)

	return nil
}
