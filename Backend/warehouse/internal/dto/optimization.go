package dto

// OptimizeRequest — запрос на оптимизацию размещения
type OptimizeRequest struct {
    Items  []OptimizeItem `json:"items"` // Список товаров с количеством
    DryRun bool           `json:"dry_run"` // Если true — только расчёт, без записи в БД
}

// OptimizeItem — товар для оптимизации
type OptimizeItem struct {
    ProductID string `json:"product_id"`
    Quantity  int    `json:"quantity"`
}

// OptimizeResponse — ответ с результатами оптимизации
type OptimizeResponse struct {
    Results []AllocationResult `json:"results"`
    Message string             `json:"message,omitempty"`
}

// AllocationResult — результат размещения одного товара
type AllocationResult struct {
    ProductID string  `json:"product_id"`
    ShelfID   string  `json:"shelf_id,omitempty"`
    Quantity  int     `json:"quantity,omitempty"`
    Score     float64 `json:"score,omitempty"`
    Assigned  bool    `json:"assigned"`
    Reason    string  `json:"reason"`
}