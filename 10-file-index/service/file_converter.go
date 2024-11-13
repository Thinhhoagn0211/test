package service

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"code.sajari.com/docconv/v2"
	"github.com/fumiama/go-docx"
	"github.com/xuri/excelize/v2"
)

func readDocxContent(filePath string) (string, error) {
	readFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	fileinfo, err := readFile.Stat()
	if err != nil {
		return "", err
	}

	var content string
	size := fileinfo.Size()
	doc, err := docx.Parse(readFile, size)
	if err != nil {
		return "", err
	}
	for _, it := range doc.Document.Body.Items {
		switch it.(type) {
		case *docx.Paragraph, *docx.Table: // printable
			switch v := it.(type) {
			case *docx.Paragraph:
				content += v.String() + "\n"
			case *docx.Table:
				content += v.String()
			}
		}
	}
	return content, nil
}

func readExcelContent(filePath string) (string, error) {
	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	var contentBuilder strings.Builder
	for _, sheetName := range f.GetSheetMap() {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			return "", fmt.Errorf("failed to get rows: %w", err)
		}
		for _, row := range rows {
			for _, col := range row {
				contentBuilder.WriteString(col)
				contentBuilder.WriteString("\t")
			}
			contentBuilder.WriteString("\n")
		}
	}

	content := contentBuilder.String()
	return content, nil
}

func readPPTXContent(filePath string) (string, error) {
	// Convert .pptx file to text
	res, err := docconv.ConvertPath(filePath)
	if err != nil {
		return "", err
	}
	content := res.Body
	return content, nil
}

// IsHiddenFile checks if a file is hidden or not
func IsHiddenFile(filename string) (bool, error) {
	pointer, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}

// GetAvailableDrives returns a list of available drives on the system
func getAvailableDrives() []string {
	var drives []string
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drivePath := fmt.Sprintf("%c:\\", drive)
		if _, err := os.Stat(drivePath); !os.IsNotExist(err) {
			drives = append(drives, drivePath)
		}
	}
	return drives
}
