package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/h2non/filetype"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
)

type ParseRequest struct {
	Text string `json:"text"`
}

type ParseResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
}

type Product struct {
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
}

type FileStructure struct {
	HasHeaders bool
	NameCol    int
	PriceCol   int
	SKUCol     int
	StartRow   int
	Separator  string
	Columns    []string
}

func main() {
	app := fiber.New(fiber.Config{
		ReadTimeout:  1800 * time.Second,
		WriteTimeout: 1800 * time.Second,
		BodyLimit:    500 * 1024 * 1024,
	})
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/api/parse", func(c *fiber.Ctx) error {
		var req ParseRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		return parseTextBatched(c, req.Text)
	})

	app.Post("/api/parse-file", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer src.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
		}

		filename := file.Filename
		ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])

		var products []Product

		switch ext {
		case ".xlsx", ".xls", ".xlsm":
			products = parseExcelWithAI(fileBytes)
		case ".csv":
			products = parseCSVWithAI(fileBytes)
		case ".pdf":
			text, err := extractPDF(fileBytes)
			if err != nil {
				text = string(fileBytes)
			}
			products = parseBatch(text)
		case ".txt":
			products = parseBatch(string(fileBytes))
		default:
			kind, _ := filetype.Match(fileBytes)
			if kind.MIME.Type == "application" {
				switch kind.MIME.Subtype {
				case "vnd.openxmlformats-officedocument.spreadsheetml.sheet", "vnd.ms-excel":
					products = parseExcelWithAI(fileBytes)
				case "pdf":
					text, err := extractPDF(fileBytes)
					if err != nil {
						text = string(fileBytes)
					}
					products = parseBatch(text)
				default:
					products = parseBatch(string(fileBytes))
				}
			} else {
				products = parseBatch(string(fileBytes))
			}
		}

		// Фильтрация
		var filtered []Product
		for _, p := range products {
			if len(p.ProductName) >= 3 && p.Price > 0 {
				if p.Currency == "" {
					p.Currency = "руб"
				}
				filtered = append(filtered, p)
			}
		}

		log.Printf("Final products: %d", len(filtered))

		return c.JSON(ParseResponse{
			Products:   filtered,
			TotalFound: len(filtered),
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	app.Listen(":" + port)
}

// 🧠 AI-анализ структуры для Excel
func analyzeStructureWithAI(rows [][]string, fileType string) FileStructure {
	// Формируем образец данных (первые 15 строк)
	var sample strings.Builder
	sample.WriteString(fmt.Sprintf("Файл: %s\n", fileType))
	sample.WriteString("Первые строки:\n")

	for i, row := range rows {
		if i >= 15 {
			break
		}
		// Показываем колонки с номерами для удобства
		var cols []string
		for j, cell := range row {
			if j < 10 { // Ограничиваем колонки
				cols = append(cols, fmt.Sprintf("[%d]%s", j, cell))
			}
		}
		sample.WriteString(fmt.Sprintf("Строка %d: %s\n", i, strings.Join(cols, " | ")))
	}

	prompt := `Ты — AI-ассистент для парсинга прайс-листов.
Проанализируй структуру данных и определи:
1. Какая колонка содержит НАЗВАНИЕ товара (номер колонки с 0)
2. Какая колонка содержит ЦЕНУ (номер колонки с 0)
3. Есть ли заголовки (да/нет)
4. С какой строки начинаются данные (0 — первая строка)

Верни ТОЛЬКО JSON без пояснений:
{
  "name_col": 0,
  "price_col": 4,
  "has_headers": true,
  "start_row": 2,
  "confidence": 95
}

Данные:
` + sample.String()

	ollamaReq := map[string]interface{}{
		"model":  "llama3.2:3b",
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	jsonData, _ := json.Marshal(ollamaReq)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("AI structure analysis failed: %v", err)
		return FileStructure{NameCol: -1, PriceCol: -1}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ollamaResp map[string]interface{}
	json.Unmarshal(body, &ollamaResp)

	response, ok := ollamaResp["response"].(string)
	if !ok {
		return FileStructure{NameCol: -1, PriceCol: -1}
	}

	log.Printf("AI analysis response: %s", response)

	// Парсим JSON из ответа
	re := regexp.MustCompile(`\{[^{}]*\}`)
	jsonMatch := re.FindString(response)
	if jsonMatch == "" {
		return FileStructure{NameCol: -1, PriceCol: -1}
	}

	var result struct {
		NameCol    int  `json:"name_col"`
		PriceCol   int  `json:"price_col"`
		HasHeaders bool `json:"has_headers"`
		StartRow   int  `json:"start_row"`
		Confidence int  `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonMatch), &result); err != nil {
		log.Printf("Failed to parse AI response: %v", err)
		return FileStructure{NameCol: -1, PriceCol: -1}
	}

	log.Printf("AI analysis: NameCol=%d, PriceCol=%d, HasHeaders=%v, StartRow=%d, Confidence=%d%%",
		result.NameCol, result.PriceCol, result.HasHeaders, result.StartRow, result.Confidence)

	return FileStructure{
		NameCol:    result.NameCol,
		PriceCol:   result.PriceCol,
		HasHeaders: result.HasHeaders,
		StartRow:   result.StartRow,
	}
}

// 🧠 AI-анализ структуры для CSV
func analyzeCSVStructure(text string) FileStructure {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return FileStructure{NameCol: -1, PriceCol: -1}
	}

	// Определяем разделитель
	separators := []string{";", ",", "\t", "|"}
	bestSep := ","
	maxCount := 0

	for _, sep := range separators {
		count := strings.Count(lines[0], sep)
		if count > maxCount {
			maxCount = count
			bestSep = sep
		}
	}

	// Разбиваем строки
	var rows [][]string
	for i, line := range lines {
		if i >= 20 {
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, bestSep)
		rows = append(rows, parts)
	}

	if len(rows) == 0 {
		return FileStructure{NameCol: -1, PriceCol: -1}
	}

	// Используем AI для анализа
	structInfo := analyzeStructureWithAI(rows, "CSV")
	structInfo.Separator = bestSep

	return structInfo
}

// Парсинг Excel с AI
func parseExcelWithAI(data []byte) []Product {
	reader, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to open Excel: %v", err)
		return nil
	}
	defer reader.Close()

	sheets := reader.GetSheetList()
	if len(sheets) == 0 {
		return nil
	}

	rows, err := reader.GetRows(sheets[0])
	if err != nil || len(rows) == 0 {
		return nil
	}

	// 🧠 Анализируем структуру через AI
	structure := analyzeStructureWithAI(rows, "Excel")

	// Если AI не помог — используем fallback
	if structure.NameCol < 0 || structure.PriceCol < 0 {
		log.Printf("AI analysis failed, using fallback")
		structure = findExcelStructure(rows)
	}

	if structure.NameCol < 0 || structure.PriceCol < 0 {
		log.Printf("Structure not found, using AI batch parsing")
		var textBuilder strings.Builder
		for i, row := range rows {
			if i > 1000 {
				break
			}
			textBuilder.WriteString(strings.Join(row, " ") + "\n")
		}
		return parseBatch(textBuilder.String())
	}

	log.Printf("✅ Using structure: NameCol=%d, PriceCol=%d, StartRow=%d",
		structure.NameCol, structure.PriceCol, structure.StartRow)

	return parseExcelManual(rows, structure)
}

// Парсинг CSV с AI
func parseCSVWithAI(data []byte) []Product {
	text := string(data)

	// 🧠 Анализируем структуру через AI
	structure := analyzeCSVStructure(text)

	if structure.NameCol < 0 || structure.PriceCol < 0 {
		log.Printf("AI CSV analysis failed, using AI batch parsing")
		return parseBatch(text)
	}

	log.Printf("✅ CSV structure: NameCol=%d, PriceCol=%d, StartRow=%d, Separator='%s'",
		structure.NameCol, structure.PriceCol, structure.StartRow, structure.Separator)

	return parseCSVManual(text, structure)
}

// Ручной парсинг CSV
func parseCSVManual(text string, structure FileStructure) []Product {
	lines := strings.Split(text, "\n")
	var products []Product
	priceRegex := regexp.MustCompile(`[\d]+[.,]?\d*`)

	for i := structure.StartRow; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Split(line, structure.Separator)
		if len(parts) <= structure.NameCol || len(parts) <= structure.PriceCol {
			continue
		}

		name := strings.TrimSpace(parts[structure.NameCol])
		if name == "" || len(name) < 3 {
			continue
		}

		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, "итог") ||
			strings.Contains(nameLower, "всего") ||
			strings.Contains(nameLower, "сумма") ||
			strings.Contains(nameLower, "наименование") {
			continue
		}

		priceStr := strings.TrimSpace(parts[structure.PriceCol])
		priceStr = strings.ReplaceAll(priceStr, " ", "")
		priceStr = strings.Replace(priceStr, ",", ".", -1)

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			match := priceRegex.FindString(priceStr)
			if match != "" {
				match = strings.Replace(match, ",", ".", -1)
				price, err = strconv.ParseFloat(match, 64)
				if err != nil {
					continue
				}
			} else {
				continue
			}
		}

		if price <= 0 {
			continue
		}

		products = append(products, Product{
			ProductName: name,
			Price:       price,
			Currency:    "руб",
		})
	}

	return products
}

// Ручной парсинг Excel
func parseExcelManual(rows [][]string, structure FileStructure) []Product {
	var products []Product
	priceRegex := regexp.MustCompile(`[\d]+[.,]?\d*`)

	for i := structure.StartRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) <= structure.NameCol || len(row) <= structure.PriceCol {
			continue
		}

		name := strings.TrimSpace(row[structure.NameCol])
		if name == "" || len(name) < 3 {
			continue
		}

		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, "итог") ||
			strings.Contains(nameLower, "всего") ||
			strings.Contains(nameLower, "сумма") {
			continue
		}

		priceStr := strings.TrimSpace(row[structure.PriceCol])
		priceStr = strings.ReplaceAll(priceStr, " ", "")
		priceStr = strings.Replace(priceStr, ",", ".", -1)

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			match := priceRegex.FindString(priceStr)
			if match != "" {
				match = strings.Replace(match, ",", ".", -1)
				price, err = strconv.ParseFloat(match, 64)
				if err != nil {
					continue
				}
			} else {
				continue
			}
		}

		if price <= 0 {
			continue
		}

		products = append(products, Product{
			ProductName: name,
			Price:       price,
			Currency:    "руб",
		})
	}

	return products
}

// Fallback: поиск структуры в Excel (без AI)
func findExcelStructure(rows [][]string) FileStructure {
	structInfo := FileStructure{
		NameCol:  -1,
		PriceCol: -1,
		StartRow: 0,
	}

	headerKeywords := map[string][]string{
		"name":  {"наименование", "товар", "название", "продукт", "деталь", "запчасть", "описание", "услуга"},
		"price": {"цена", "стоимость", "сумма", "price", "руб", "розница", "опт", "стоимость"},
	}

	for rowIdx := 0; rowIdx < len(rows) && rowIdx < 15; rowIdx++ {
		row := rows[rowIdx]
		nameFound := false
		priceFound := false

		for colIdx, cell := range row {
			cellLower := strings.ToLower(strings.TrimSpace(cell))

			for _, keyword := range headerKeywords["name"] {
				if strings.Contains(cellLower, keyword) {
					structInfo.NameCol = colIdx
					nameFound = true
					break
				}
			}

			for _, keyword := range headerKeywords["price"] {
				if strings.Contains(cellLower, keyword) {
					structInfo.PriceCol = colIdx
					priceFound = true
					break
				}
			}
		}

		if nameFound && priceFound {
			structInfo.HasHeaders = true
			structInfo.StartRow = rowIdx + 1
			return structInfo
		}
	}

	if len(rows) > 0 && len(rows[0]) >= 2 {
		lastCol := len(rows[0]) - 1
		structInfo.PriceCol = lastCol
		structInfo.NameCol = 0
		structInfo.StartRow = 1
	}

	return structInfo
}

func parseTextBatched(c *fiber.Ctx, text string) error {
	products := parseBatch(text)
	if len(products) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "No products found"})
	}
	return c.JSON(ParseResponse{
		Products:   products,
		TotalFound: len(products),
	})
}

func parseBatch(text string) []Product {
	prompt := `Извлеки товары и их цены из текста.
Верни ТОЛЬКО JSON массив с объектами: product_name и price.

Важно: бери ПОЛНЫЕ названия, не обрезай их.
Если видишь структуру "артикул - название - цена", бери название, а не артикул.

Текст: ` + text

	ollamaReq := map[string]interface{}{
		"model":  "llama3.2:3b",
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.1,
		},
	}

	jsonData, _ := json.Marshal(ollamaReq)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Ollama error: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ollamaResp map[string]interface{}
	json.Unmarshal(body, &ollamaResp)

	response, ok := ollamaResp["response"].(string)
	if !ok {
		return nil
	}

	re := regexp.MustCompile(`\[\s*\{[\s\S]*\}\s*\]`)
	jsonMatch := re.FindString(response)
	if jsonMatch != "" {
		var parsed []Product
		if err := json.Unmarshal([]byte(jsonMatch), &parsed); err == nil {
			return parsed
		}
	}
	return nil
}

func extractPDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var text strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text.WriteString(content)
		text.WriteString("\n")
	}
	return text.String(), nil
}
