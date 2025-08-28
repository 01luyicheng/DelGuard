package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	// 测试各种配置文件格式
	formats := []struct {
		name      string
		filename  string
		formatter ConfigFormatter
	}{
		{"YAML", "test_config.yaml", &YAMLFormatter{}},
		{"TOML", "test_config.toml", &TOMLFormatter{}},
		{"INI", "test_config.ini", &INIFormatter{}},
		{"Properties", "test_config.properties", &PropertiesFormatter{}},
	}

	for _, format := range formats {
		fmt.Printf("测试 %s 格式...\n", format.name)
		
		// 读取文件
		content, err := ioutil.ReadFile(format.filename)
		if err != nil {
			fmt.Printf("❌ 读取 %s 文件失败: %v\n", format.filename, err)
			continue
		}

		// 验证格式
		if err := format.formatter.Validate(content); err != nil {
			fmt.Printf("❌ %s 格式验证失败: %v\n", format.name, err)
			continue
		}
		fmt.Printf("✅ %s 格式验证通过\n", format.name)

		// 解析配置
		config, err := format.formatter.Parse(content)
		if err != nil {
			fmt.Printf("❌ %s 解析失败: %v\n", format.name, err)
			continue
		}
		fmt.Printf("✅ %s 解析成功\n", format.name)

		// 验证解析结果
		if config.Version != "1.0.0" {
			fmt.Printf("❌ %s 版本解析错误: 期望 1.0.0, 实际 %s\n", format.name, config.Version)
			continue
		}
		if config.Performance.BatchSize != 100 {
			fmt.Printf("❌ %s 批处理大小解析错误: 期望 100, 实际 %d\n", format.name, config.Performance.BatchSize)
			continue
		}
		fmt.Printf("✅ %s 配置值解析正确\n", format.name)

		// 测试格式化
		formatted, err := format.formatter.Format(config)
		if err != nil {
			fmt.Printf("❌ %s 格式化失败: %v\n", format.name, err)
			continue
		}
		fmt.Printf("✅ %s 格式化成功\n", format.name)

		// 保存格式化结果
		outputFile := fmt.Sprintf("output_%s", format.filename)
		if err := ioutil.WriteFile(outputFile, formatted, 0644); err != nil {
			fmt.Printf("❌ 保存 %s 失败: %v\n", outputFile, err)
		} else {
			fmt.Printf("✅ 格式化结果已保存到 %s\n", outputFile)
		}

		fmt.Printf("✅ %s 格式测试完成\n\n", format.name)
	}

	fmt.Println("所有配置文件格式测试完成！")
}