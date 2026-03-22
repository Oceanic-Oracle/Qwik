package dto

// OptimizeRequest — запрос на оптимизацию размещения
type OptimizeRequest struct {
	ProductIDs  []string `json:"product_ids"`   // Список ID товаров для оптимизации
	WarehouseID string   `json:"warehouse_id"`  // ID склада
	DryRun      bool     `json:"dry_run"`       // Если true — только расчёт, без записи в БД
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
	Score     float64 `json:"score,omitempty"`
	Assigned  bool    `json:"assigned"`
	Reason    string  `json:"reason"`
}