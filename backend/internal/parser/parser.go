package parser

import (
        "bufio"
        "encoding/csv"
        "fmt"
        "io"
        "log"
        "os"
        "regexp"
        "strconv"
        "strings"

        "github.com/xuri/excelize/v2"
)

type ParseResult struct {
        Products []Product
        RawText  string
        Headers  []string
}

type Product struct {
        SKU   string  `json:"sku"`
        Name  string  `json:"name"`
        Price float64 `json:"price"`
        Unit  string  `json:"unit"`
}

func ParseFile(filePath string) (*ParseResult, error) {
        if strings.HasSuffix(filePath, ".csv") {
                return parseCSV(filePath)
        }
        if strings.HasSuffix(filePath, ".txt") {
                return parseTXT(filePath)
        }
        if strings.HasSuffix(filePath, ".xlsx") || strings.HasSuffix(filePath, ".xls") {
                return parseExcel(filePath)
        }
        return nil, fmt.Errorf("unsupported file type: %s", filePath)
}

func parseExcel(filePath string) (*ParseResult, error) {
        f, err := excelize.OpenFile(filePath)
        if err != nil {
                return nil, fmt.Errorf("open excel: %w", err)
        }
        defer f.Close()

        sheetName := f.GetSheetName(0)
        rows, err := f.GetRows(sheetName)
        if err != nil {
                return nil, fmt.Errorf("get rows: %w", err)
        }

        return extractProducts(rows), nil
}

func extractProducts(rows [][]string) *ParseResult {
        var products []Product
        var rawText strings.Builder
        var headers []string

        // Начинаем с 6-й строки (пропускаем шапку)
        startRow := 5
        totalRows := 0
        priceRows := 0
        skippedRows := 0

        for i := startRow; i < len(rows); i++ {
                row := rows[i]
                if len(row) < 3 {
                        continue
                }
                totalRows++

                // Цена в колонке C (индекс 2)
                priceStr := strings.TrimSpace(row[2])
                if priceStr == "" {
                        skippedRows++
                        continue
                }

                // Парсим цену
                cleanPrice := strings.ReplaceAll(priceStr, " ", "")
                cleanPrice = strings.ReplaceAll(cleanPrice, ",", ".")
                cleanPrice = strings.ReplaceAll(cleanPrice, "р", "")
                cleanPrice = strings.ReplaceAll(cleanPrice, "уб", "")
                cleanPrice = strings.ReplaceAll(cleanPrice, "₽", "")
                cleanPrice = strings.ReplaceAll(cleanPrice, "грн", "")
                cleanPrice = strings.ReplaceAll(cleanPrice, "Р", "")

                price, err := strconv.ParseFloat(cleanPrice, 64)
                if err != nil || price == 0 {
                        skippedRows++
                        continue
                }
                priceRows++

                // Название в колонке B (индекс 1)
                name := strings.TrimSpace(row[1])
                if name == "" && len(row) > 0 {
                        name = strings.TrimSpace(row[0])
                }
                if name == "" {
                        skippedRows++
                        continue
                }

                // Убираем нумерацию из названия
                name = regexp.MustCompile(`^\d+\.\d+\s*`).ReplaceAllString(name, "")
                name = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(name, "")
                name = regexp.MustCompile(`^\d+\s+`).ReplaceAllString(name, "")
                name = strings.TrimSpace(name)

                if name == "" {
                        skippedRows++
                        continue
                }

                // 🔥 АРТИКУЛ — берём из колонки A (индекс 0) ВСЕГДА
                sku := ""
                if len(row) > 0 {
                        sku = strings.TrimSpace(row[0])
                }

                // Если артикул пустой, пробуем взять из колонки B
                if sku == "" && len(row) > 1 {
                        sku = strings.TrimSpace(row[1])
                }

                products = append(products, Product{
                        SKU:   sku,
                        Name:  name,
                        Price: price,
                        Unit:  "руб",
                })

                rawText.WriteString(fmt.Sprintf("%s,%.2f,%s\n", name, price, sku))
        }

        log.Printf("📊 Статистика парсинга: Всего строк %d, с ценами %d, пропущено %d, найдено товаров %d",
                totalRows, priceRows, skippedRows, len(products))

        return &ParseResult{
                Products: products,
                RawText:  rawText.String(),
                Headers:  headers,
        }
}

func parseCSV(filePath string) (*ParseResult, error) {
        file, err := os.Open(filePath)
        if err != nil {
                return nil, err
        }
        defer file.Close()

        reader := csv.NewReader(bufio.NewReader(file))
        reader.LazyQuotes = true

        var products []Product
        var rawText strings.Builder

        h, err := reader.Read()
        if err != nil {
                return nil, err
        }
        rawText.WriteString(strings.Join(h, ",") + "\n")

        for {
                row, err := reader.Read()
                if err == io.EOF {
                        break
                }
                if err != nil {
                        continue
                }
                rawText.WriteString(strings.Join(row, ",") + "\n")

                if len(row) >= 3 {
                        var price float64
                        fmt.Sscanf(row[2], "%f", &price)
                        if price > 0 {
                                products = append(products, Product{
                                        SKU:   row[0],
                                        Name:  row[1],
                                        Price: price,
                                        Unit:  "шт",
                                })
                        }
                }
        }

        return &ParseResult{
                Products: products,
                RawText:  rawText.String(),
                Headers:  h,
        }, nil
}

func parseTXT(filePath string) (*ParseResult, error) {
        data, err := os.ReadFile(filePath)
        if err != nil {
                return nil, err
        }
        return &ParseResult{
                Products: []Product{},
                RawText:  string(data),
                Headers:  []string{},
        }, nil
}
