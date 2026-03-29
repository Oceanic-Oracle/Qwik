package dto

// StockRequest — старый запрос для ручного управления (пока оставляем для списания)
type StockRequest struct {
    ShelfID   string `json:"shelf_id"`
    ProductID string `json:"product_id"`
    Quantity  int    `json:"quantity"`
}

// AutoStockRequest — запрос на автоматическое размещение
type AutoStockRequest struct {
    ProductID string `json:"product_id"`
    Quantity  int    `json:"quantity"`
}

// StockResponse — информация по конкретной полке после размещения
type StockResponse struct {
    ShelfID      string  `json:"shelf_id"`
    ProductID    string  `json:"product_id"`
    NewQuantity  int     `json:"new_quantity"`
    UsedCapacity float64 `json:"used_capacity"`
    MaxCapacity  float64 `json:"max_capacity"`
    Message      string  `json:"message"`
}

// AutoStockResponse — полный ответ на автоматическое размещение
type AutoStockResponse struct {
    ProductID      string           `json:"product_id"`
    TotalRequested int              `json:"total_requested"`
    TotalPlaced   int              `json:"total_placed"`
    Results       []*StockResponse `json:"results"`
    Message       string           `json:"message"`
}