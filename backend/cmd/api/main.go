package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type DiscountConfig struct {
	PrimaryPercent   float64  `json:"primary_percent"`
	SecondaryPercent float64  `json:"secondary_percent"`
	PrimaryRub       float64  `json:"primary_rub"`
	SecondaryRub     float64  `json:"secondary_rub"`
	CustomPercent    float64  `json:"custom_percent"`
	CustomRub        float64  `json:"custom_rub"`
	Categories       []string `json:"categories"`
	ApplyCustom      bool     `json:"apply_custom"`
	UsePercent       bool     `json:"use_percent"`
}

type Product struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
	NewPrice float64 `json:"new_price"`
	Manual   bool    `json:"manual"`
	RowNum   int     `json:"row_num"`
}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	tempDir := os.TempDir()
	log.Printf("📁 Temp folder: %s", tempDir)

	// 1. Upload
	app.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "file required"})
		}

		fileID := uuid.New().String()
		path := filepath.Join(tempDir, fileID+".xlsx")
		if err := c.SaveFile(file, path); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "save failed: " + err.Error()})
		}

		f, err := excelize.OpenFile(path)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer f.Close()

		rows, err := f.GetRows(f.GetSheetName(0))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		categories := []string{}
		for _, row := range rows {
			if len(row) < 1 {
				continue
			}
			a := strings.TrimSpace(row[0])
			b := ""
			if len(row) > 1 {
				b = strings.TrimSpace(row[1])
			}
			c := ""
			if len(row) > 2 {
				c = strings.TrimSpace(row[2])
			}
			if a != "" && b == "" && c == "" && strings.Contains(a, ".") {
				categories = append(categories, a)
			}
		}

		return c.JSON(fiber.Map{
			"file_id":    fileID,
			"filename":   file.Filename,
			"total":      len(rows),
			"categories": categories,
		})
	})

	// 2. Get all products
	app.Get("/api/products/:file_id", func(c *fiber.Ctx) error {
		fileID := c.Params("file_id")
		filePath := filepath.Join(tempDir, fileID+".xlsx")

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(404).JSON(fiber.Map{"error": "file not found"})
		}

		f, err := excelize.OpenFile(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer f.Close()

		rows, err := f.GetRows(f.GetSheetName(0))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Загружаем метки ручных правок
		manualPath := filepath.Join(tempDir, fileID+"_manual.json")
		manualData := make(map[int]bool)
		if data, err := os.ReadFile(manualPath); err == nil {
			json.Unmarshal(data, &manualData)
		}

		var products []Product
		for i, row := range rows {
			if len(row) < 3 {
				continue
			}
			a := strings.TrimSpace(row[0])
			b := strings.TrimSpace(row[1])
			c := strings.TrimSpace(row[2])

			if c == "" {
				continue
			}

			priceStr := strings.ReplaceAll(c, " ", "")
			priceStr = strings.ReplaceAll(priceStr, ",", ".")
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil || price == 0 {
				continue
			}

			// Пропускаем категории (A не пусто, B пусто)
			if a != "" && b == "" && strings.Contains(a, ".") {
				continue
			}

			if a == "" && b == "" {
				continue
			}

			if b == "" {
				b = a
				a = ""
			}

			products = append(products, Product{
				SKU:      a,
				Name:     b,
				Price:    price,
				NewPrice: price,
				Manual:   manualData[i],
				RowNum:   i,
			})
		}

		return c.JSON(products)
	})

	// 3. Apply discounts
	app.Post("/apply-discounts", func(c *fiber.Ctx) error {
		var req struct {
			FileID   string         `json:"file_id"`
			Discount DiscountConfig `json:"discount"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		filePath := filepath.Join(tempDir, req.FileID+".xlsx")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(404).JSON(fiber.Map{"error": "file not found"})
		}

		manualPath := filepath.Join(tempDir, req.FileID+"_manual.json")
		manualData := make(map[int]bool)
		if data, err := os.ReadFile(manualPath); err == nil {
			json.Unmarshal(data, &manualData)
		}

		// Создаём копию файла для скидок
		f, err := excelize.OpenFile(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer f.Close()

		newFilePath := filepath.Join(tempDir, req.FileID+"_discount.xlsx")
		if err := f.SaveAs(newFilePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		newFile, err := excelize.OpenFile(newFilePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer newFile.Close()

		sheet := newFile.GetSheetName(0)
		rows, err := newFile.GetRows(sheet)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		currentCategory := ""
		categoriesMap := make(map[string]bool)
		for _, cat := range req.Discount.Categories {
			categoriesMap[strings.TrimSpace(cat)] = true
		}

		var products []Product
		totalProducts := 0
		usePercent := req.Discount.UsePercent

		for i, row := range rows {
			if len(row) < 1 {
				continue
			}

			a := strings.TrimSpace(row[0])
			b := ""
			if len(row) > 1 {
				b = strings.TrimSpace(row[1])
			}
			c := ""
			if len(row) > 2 {
				c = strings.TrimSpace(row[2])
			}

			if a != "" && b == "" && c == "" && strings.Contains(a, ".") {
				currentCategory = a
				continue
			}

			if c == "" {
				continue
			}

			// Если категория не выбрана — пропускаем
			if !categoriesMap[currentCategory] {
				continue
			}

			priceStr := strings.ReplaceAll(c, " ", "")
			priceStr = strings.ReplaceAll(priceStr, ",", ".")
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil || price == 0 {
				continue
			}

			name := b
			if name == "" {
				name = a
			}
			if name == "" {
				continue
			}

			sku := a

			var discount float64 = 0
			var newPrice float64 = price

			// Если есть ручная правка — пропускаем
			if manualData[i] {
				newPrice = price
				discount = 0
			} else {
				// Приоритет: произвольная > первичный/вторичный
				if req.Discount.ApplyCustom {
					if usePercent {
						discount = req.Discount.CustomPercent
						if discount > 0 {
							newPrice = price * (1 - discount/100)
						}
					} else {
						discount = req.Discount.CustomRub
						newPrice = price - discount
					}
				} else {
					nameLower := strings.ToLower(name)
					if strings.Contains(nameLower, "первичный") {
						if usePercent {
							discount = req.Discount.PrimaryPercent
							if discount > 0 {
								newPrice = price * (1 - discount/100)
							}
						} else {
							discount = req.Discount.PrimaryRub
							newPrice = price - discount
						}
					} else if strings.Contains(nameLower, "вторичный") || strings.Contains(nameLower, "повторный") {
						if usePercent {
							discount = req.Discount.SecondaryPercent
							if discount > 0 {
								newPrice = price * (1 - discount/100)
							}
						} else {
							discount = req.Discount.SecondaryRub
							newPrice = price - discount
						}
					}
				}
			}

			if newPrice != price {
				cellName, _ := excelize.CoordinatesToCellName(3, i+1)
				newFile.SetCellValue(sheet, cellName, fmt.Sprintf("%.2f", newPrice))
				totalProducts++
			}

			product := Product{
				SKU:      sku,
				Name:     name,
				Price:    price,
				Discount: discount,
				NewPrice: newPrice,
				Manual:   manualData[i],
				RowNum:   i,
			}
			products = append(products, product)
		}

		if err := newFile.Save(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"total_products": totalProducts,
			"products":       products,
			"download_url":   "/download/" + req.FileID,
		})
	})

	// 4. Download
	app.Get("/download/:file_id", func(c *fiber.Ctx) error {
		fileID := c.Params("file_id")
		filePath := filepath.Join(tempDir, fileID+"_discount.xlsx")

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Если нет файла со скидками, скачиваем исходный
			filePath = filepath.Join(tempDir, fileID+".xlsx")
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return c.Status(404).JSON(fiber.Map{"error": "file not found"})
			}
			return c.Download(filePath, "prices.xlsx")
		}

		return c.Download(filePath, "prices_with_discount.xlsx")
	})

	// 5. Update price manually
	app.Post("/api/update-price", func(c *fiber.Ctx) error {
		var req struct {
			FileID   string  `json:"file_id"`
			RowNum   int     `json:"row_num"`
			NewPrice float64 `json:"new_price"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		// Определяем, какой файл редактировать: если есть _discount, то его, иначе исходный
		basePath := filepath.Join(tempDir, req.FileID+".xlsx")
		discountPath := filepath.Join(tempDir, req.FileID+"_discount.xlsx")

		targetPath := basePath
		if _, err := os.Stat(discountPath); err == nil {
			targetPath = discountPath
		}

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			return c.Status(404).JSON(fiber.Map{"error": "file not found"})
		}

		f, err := excelize.OpenFile(targetPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer f.Close()

		sheet := f.GetSheetName(0)
		cellName, _ := excelize.CoordinatesToCellName(3, req.RowNum+1)

		// Сохраняем с двумя знаками после запятой
		if err := f.SetCellValue(sheet, cellName, fmt.Sprintf("%.2f", req.NewPrice)); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to update cell"})
		}

		if err := f.Save(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to save"})
		}

		// Сохраняем метку ручной правки
		manualPath := filepath.Join(tempDir, req.FileID+"_manual.json")
		manualData := make(map[int]bool)
		if data, err := os.ReadFile(manualPath); err == nil {
			json.Unmarshal(data, &manualData)
		}
		manualData[req.RowNum] = true
		if jsonData, err := json.Marshal(manualData); err == nil {
			os.WriteFile(manualPath, jsonData, 0644)
		}

		return c.JSON(fiber.Map{
			"success":   true,
			"row_num":   req.RowNum,
			"new_price": req.NewPrice,
			"manual":    true,
		})
	})

	// 6. Reset manual edits
	app.Post("/api/reset-manual", func(c *fiber.Ctx) error {
		var req struct {
			FileID string `json:"file_id"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		manualPath := filepath.Join(tempDir, req.FileID+"_manual.json")
		if err := os.Remove(manualPath); err != nil && !os.IsNotExist(err) {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"success": true, "message": "Manual resets cleared"})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	log.Println("✅ Server running on :3000")
	log.Fatal(app.Listen(":3000"))
}
