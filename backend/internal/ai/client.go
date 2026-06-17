package ai

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "log"
        "net/http"
        "regexp"
        "strings"
        "time"
)

type Client struct {
        baseURL string
        model   string
        client  *http.Client
}

func NewClient(baseURL, model string) *Client {
        return &Client{
                baseURL: baseURL,
                model:   model,
                client:  &http.Client{
                        Timeout: 300 * time.Second, // 5 минут
                },
        }
}

type Product struct {
        SKU         string  `json:"sku"`
        Name        string  `json:"name"`
        Price       float64 `json:"price"`
        Unit        string  `json:"unit"`
        Category    string  `json:"category"`
        Description string  `json:"description"`
        Manufacturer string `json:"manufacturer"`
}

type StructuredData struct {
        Columns  map[string]string `json:"columns"`
        Products []Product         `json:"products"`
}

type OllamaRequest struct {
        Model  string `json:"model"`
        Prompt string `json:"prompt"`
        Stream bool   `json:"stream"`
}

type OllamaResponse struct {
        Response string `json:"response"`
}

func (c *Client) ExtractStructuredData(text string, headers []string) (*StructuredData, error) {
        // Отправляем только первые 2000 символов для скорости
        if len(text) > 2000 {
                text = text[:2000] + "\n... (обрезано для экономии времени)"
        }
        
        prompt := buildStructuredPrompt(text, headers)
        
        reqBody := OllamaRequest{
                Model:  c.model,
                Prompt: prompt,
                Stream: false,
        }

        jsonData, err := json.Marshal(reqBody)
        if err != nil {
                return nil, err
        }

        url := c.baseURL + "/api/generate"
        resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                return nil, err
        }

        var ollamaResp OllamaResponse
        if err := json.Unmarshal(body, &ollamaResp); err != nil {
                return nil, err
        }

        log.Printf("AI raw response: %s", ollamaResp.Response)

        var result StructuredData
        if err := parseJSONResponse(ollamaResp.Response, &result); err != nil {
                return nil, err
        }

        return &result, nil
}

func buildStructuredPrompt(text string, headers []string) string {
        headerStr := strings.Join(headers, ", ")
        return fmt.Sprintf(`Ты — эксперт по парсингу прайс-листов. 
Проанализируй данные и определи структуру.

Заголовки колонок: %s

Данные:
%s

Определи, какие колонки соответствуют:
- name (название товара/услуги)
- price (цена)
- sku (артикул, если есть)
- unit (единица измерения)

Верни JSON в формате:
{
  "columns": {
    "name": "Название колонки с названием",
    "price": "Название колонки с ценой",
    "sku": "Название колонки с артикулом",
    "unit": "Название колонки с единицей измерения"
  },
  "products": [
    {"name": "...", "price": 0, "sku": "...", "unit": "..."}
  ]
}

Если колонка не найдена — оставь пустой строкой.
Верни ТОЛЬКО JSON, без пояснений.`, headerStr, text)
}

func parseJSONResponse(response string, result interface{}) error {
        re := regexp.MustCompile(`\{[\s\S]*\}`)
        match := re.FindString(response)
        if match == "" {
                return fmt.Errorf("no JSON found in response")
        }
        return json.Unmarshal([]byte(match), result)
}
