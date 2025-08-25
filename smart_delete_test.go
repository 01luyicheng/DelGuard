package main

import (
	"os"
	"testing"
)

// 标准单元测试
func TestSmartDeleteFeatures(t *testing.T) {
	// 创建测试文件
	testFiles := []string{"test_file.txt", "test_document.doc", "sample.log"}
	for _, file := range testFiles {
		err := os.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
		defer os.Remove(file)
	}

	// 测试相似度计算
	similarity := CalculateSimilarity("test_file.txt", "test_document.txt")
	if similarity < 0 || similarity > 100 {
		t.Errorf("相似度计算结果异常: %.1f%%", similarity)
	}

	// 边界测试：完全不同字符串
	simZero := CalculateSimilarity("abc", "xyz")
	if simZero > 10 {
		t.Errorf("完全不同字符串相似度过高: %.1f%%", simZero)
	}

	// 测试智能搜索
	config := DefaultSmartSearchConfig
	search := NewSmartFileSearch(config)
	results, err := search.SearchFiles("test_doc", ".")
	if err != nil {
		t.Errorf("智能搜索失败: %v", err)
	}
	if len(results) == 0 {
		t.Errorf("智能搜索未找到任何文件")
	}

	// 测试正则表达式解析
	parser, err := NewRegexParser("*.txt")
	if err != nil {
		t.Errorf("正则解析器创建失败: %v", err)
	} else {
		matches, err := parser.FindMatches(".", false)
		if err != nil {
			t.Errorf("正则匹配失败: %v", err)
		}
		// 至少能匹配到 test_file.txt
		found := false
		for _, match := range matches {
			if match == "test_file.txt" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("正则未能匹配到 test_file.txt")
		}
	}
}

// ...existing code...
