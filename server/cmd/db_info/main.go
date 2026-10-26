package main

import (
	"fmt"
	"os"
	"soybean-admin-go/global"
	initialize "soybean-admin-go/init"
	"soybean-admin-go/model/po"
)

func main() {
	// 初始化配置
	err := os.Setenv("CONFIG_PATH", "./config.toml")
	if err != nil {
		fmt.Println("设置环境变量失败：", err)
		return
	}
	initialize.InitAll()

	// 检查数据库连接
	if global.DB == nil {
		fmt.Println("数据库连接为空")
		return
	}

	// 打印表结构信息
	printTableInfo("l_api")
	printTableInfo("l_button")
	printTableInfo("l_users")
	printTableInfo("l_menu")
	printTableInfo("l_role")
	printTableInfo("l_user_roles")
	printTableInfo("l_role_buttons")
	printTableInfo("l_role_apis")
}

func printTableInfo(tableName string) {
	fmt.Printf("\n=== 表 %s 的结构信息 ===\n", tableName)

	// 使用原生SQL查询表结构
	if tableName == "l_user_roles" || tableName == "l_button" || tableName == "l_api" {
		// 查询表的列结构
		rows, err := global.DB.Raw("SELECT * FROM information_schema.columns WHERE table_name = ?", tableName).Rows()
		if err != nil {
			fmt.Printf("查询表结构失败: %v\n", err)
			return
		}
		defer rows.Close()

		fmt.Println("列名:")
		type Column struct {
			TableCatalog string
			TableSchema  string
			TableName    string
			ColumnName   string
			DataType     string
		}

		for rows.Next() {
			var col Column
			if err := global.DB.ScanRows(rows, &col); err != nil {
				fmt.Printf("扫描行失败: %v\n", err)
				continue
			}
			fmt.Printf("- %s (%s)\n", col.ColumnName, col.DataType)
		}
	} else {
		// 直接查询数据库中的数据
		var count int64
		global.DB.Table(tableName).Count(&count)
		fmt.Printf("记录数量: %d\n", count)

		// 如果是用户表，打印用户信息
		if tableName == "l_users" {
			var users []po.UserModel
			global.DB.Find(&users)
			for _, user := range users {
				fmt.Printf("用户ID: %d, 用户名: %s\n", user.ID, user.UserName)
			}
		}

		// 如果是角色表，打印角色信息
		if tableName == "l_role" {
			var roles []po.RoleModel
			global.DB.Find(&roles)
			for _, role := range roles {
				fmt.Printf("角色ID: %d, 角色名: %s, 角色编码: %s\n", role.ID, role.Name, role.Code)
			}
		}
	}
}
