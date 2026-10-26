package main

import (
	"fmt"
	"soybean-admin-go/global"
	initialize "soybean-admin-go/init"
	"soybean-admin-go/model/po"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./model/bo",                                                       //gen代码的输出目录
		ModelPkgPath:  "./model",                                                          //模型代码的输出目录
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext, //启用默认查询和链式接口
		FieldNullable: true,                                                               //允许 Null 的字段生成指针类型
	})
	// 初始化所有组件
	initialize.InitAll()
	g.UseDB(global.DB)
	//g.GenerateAllTable() //通过数据中的表生成对应的模型

	genUser(g)

	// 执行生成
	fmt.Println("正在生成 GORM GEN 代码...")
	g.Execute()
	fmt.Println("GORM GEN 代码生成完成!")
}

func genUser(g *gen.Generator) {
	fmt.Println("正在应用模型...")
	g.ApplyBasic(po.UserModel{}, po.RoleModel{}, po.MenuModel{}, po.ApiModel{}, po.ButtonModel{})
	fmt.Println("模型应用完成!")
}
