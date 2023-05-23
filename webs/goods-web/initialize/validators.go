package initialize

import (
	"fmt"
	"reflect"
	"strings"
	"xrUncle/webs/goods-web/global"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translation "github.com/go-playground/validator/v10/translations/en"
	zh_translation "github.com/go-playground/validator/v10/translations/zh"
)

// InitTrans 中英文翻译器 - gin
func InitTrans(locale string) (err error) {
	// 修改 gin框架中的 validator引擎属性，实现定制
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		/*
			"LoginForm.User": "User长度必须至少为3个字符"
				-> "User" 想变成"user" 通过我们自定义的Tag`json:"user"`
		*/
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" { // JSON中的一种约束
				return ""
			}
			return name
		})
		zhT := zh.New() //中文翻译器
		enT := en.New() //英文翻译器
		// 第一个参数是备用的语言环境，后面的参数是应该支持的语言环境
		uni := ut.New(enT, zhT, enT)
		global.Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s)", locale)
		}
		switch locale {
		case "en":
			en_translation.RegisterDefaultTranslations(v, global.Trans)
		case "zh":
			zh_translation.RegisterDefaultTranslations(v, global.Trans)
		default:
			en_translation.RegisterDefaultTranslations(v, global.Trans)
		}
		return
	}
	return
}
