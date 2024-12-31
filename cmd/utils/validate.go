package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"strings"
)

type ValidatorX struct {
	Validate *validator.Validate
	uni      *ut.UniversalTranslator
	trans    ut.Translator
}

func NewValidator() *ValidatorX {
	validate := validator.New()
	zh := zh.New()
	uni := ut.New(zh, zh)
	trans, _ := uni.GetTranslator("zh")
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		label := field.Tag.Get("label")
		if label == "" {
			return field.Name
		}
		return label
	})
	v := &ValidatorX{
		Validate: validate,
		uni:      uni,
		trans:    trans,
	}
	return v
}

func (v *ValidatorX) Translate(errs validator.ValidationErrors) string {
	var errList []string
	for _, e := range errs {
		errList = append(errList, e.Translate(v.trans))
	}
	return strings.Join(errList, "|")
}

func (v *ValidatorX) ParseQuery(c *gin.Context, obj interface{}) string {
	if err := c.ShouldBindQuery(obj); err != nil {
		return "参数解析失败"
	}
	err := v.Validate.Struct(obj)
	if err != nil {
		return v.parseErrorHandler(err)
	}

	return ""
}

func (v *ValidatorX) ParseForm(c *gin.Context, obj interface{}) string {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return "参数解析失败"
	}
	err := v.Validate.Struct(obj)
	if err != nil {
		return v.parseErrorHandler(err)
	}

	return ""
}

func (v *ValidatorX) ParseJson(c *gin.Context, obj interface{}) string {
	if err := c.ShouldBindWith(obj, binding.JSON); err != nil {
		return "参数解析失败," + err.Error()
	}
	err := v.Validate.Struct(obj)
	if err != nil {
		return v.parseErrorHandler(err)
	}
	return ""
}

func (v *ValidatorX) ParseHeader(c *gin.Context, obj interface{}) string {
	if err := c.ShouldBindHeader(obj); err != nil {
		return "参数解析失败"
	}
	err := v.Validate.Struct(obj)
	if err != nil {
		return v.parseErrorHandler(err)
	}

	return ""
}

func (v *ValidatorX) parseErrorHandler(err error) string {
	var errStr string
	switch err.(type) {
	case validator.ValidationErrors:
		errStr = v.Translate(err.(validator.ValidationErrors))
	case *json.UnmarshalTypeError:
		unmarshalTypeError := err.(*json.UnmarshalTypeError)
		errStr = fmt.Errorf("%s 类型错误，期望类型 %s", unmarshalTypeError.Field, unmarshalTypeError.Type.String()).Error()
	default:
		errStr = err.Error()
	}
	return errStr
}
