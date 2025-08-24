package main

import (
	"fmt"
	"os"
)

// 简单的功能测试
func testSmartDeleteFeatures() {
	fmt.Println("🧪 测试DelGuard智能删除功能...")

	// 创建测试文件
	testFiles := []string{"test_file.txt", "test_document.doc", "sample.log"}
	for _, file := range testFiles {
		if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
			fmt.Printf("创建测试文件失败: %v\n", err)
			continue
		}
		defer os.Remove(file) // 清理
	}

	// 测试相似度计算
	fmt.Println("\n1. 测试字符串相似度计算:")
	similarity := CalculateSimilarity("test_file.txt", "test_document.txt")
	fmt.Printf("   'test_file.txt' 与 'test_document.txt' 相似度: %.1f%%\n", similarity)

	// 测试智能搜索
	fmt.Println("\n2. 测试智能文件搜索:")
	config := DefaultSmartSearchConfig()
	search := NewSmartFileSearch(config)

	results, err := search.SearchFiles("test_doc", ".")
	if err != nil {
		fmt.Printf("   搜索失败: %v\n", err)
	} else {
		fmt.Printf("   找到 %d 个匹配文件:\n", len(results))
		for _, result := range results {
			fmt.Printf("   - %s (相似度: %.1f%%, 类型: %s)\n",
				result.Name, result.Similarity, result.MatchType)
		}
	}

	// 测试正则表达式解析
	fmt.Println("\n3. 测试正则表达式解析:")
	parser, err := NewRegexParser("*.txt")
	if err != nil {
		fmt.Printf("   正则解析器创建失败: %v\n", err)
	} else {
		matches, err := parser.FindMatches(".", false)
		if err != nil {
			fmt.Printf("   匹配失败: %v\n", err)
		} else {
			fmt.Printf("   '*.txt' 匹配到 %d 个文件:\n", len(matches))
			for _, match := range matches {
				fmt.Printf("   - %s\n", match)
			}
		}
	}

	fmt.Println("\n✅ 智能删除功能测试完成!")
}

// 运行测试的主函数
func runTests() {
	testSmartDeleteFeatures()
}
