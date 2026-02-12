package main

import (
	"fmt"
	"os"

	gopdf "github.com/VantageDataChat/GoPDF2"
)

func main() {
	data, err := os.ReadFile("test.pdf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read file: %v\n", err)
		os.Exit(1)
	}

	// 1. 获取页数
	n, err := gopdf.GetSourcePDFPageCountFromBytes(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get page count: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Page count: %d\n", n)

	// 2. 提取全部文本
	text, err := gopdf.ExtractAllPagesText(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ExtractAllPagesText: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Extracted %d chars\n", len(text))

	// 3. 保存到文件
	if err := os.WriteFile("test_output.txt", []byte(text), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Saved to test_output.txt")

	// 4. 打印前500字符预览
	preview := text
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Printf("\n--- Preview ---\n%s\n", preview)
}
