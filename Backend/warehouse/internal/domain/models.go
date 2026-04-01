package domain

import (
    "github.com/twpayne/go-geom"
)

// Product представляет товар на складе с атрибутами для слоттинга
type Product struct {
    ID                string
    Name              string
    SKU               string
    AllocatedCapacity float64       // Объём в м³
    Tags              []string      // Теги совместимости: "food", "chemicals" и т.д.
    ABCClass          ABCClass      // Вычисляемое: A/B/C
    XYZClass          XYZClass      // Вычисляемое: X/Y/Z
    TurnoverRate      float64       // Оборот в единицах/день (из TurnoverProvider)
    DemandCV          float64       // Коэффициент вариации спроса для XYZ-классификации
}

// ABCClass — классификация по принципу Парето (80/20)
type ABCClass string

const (
    ClassA ABCClass = "A" // Высокий оборот: верхние 80% ценности
    ClassB ABCClass = "B" // Средний оборот: следующие 15%
    ClassC ABCClass = "C" // Низкий оборот: нижние 5%
)

// XYZClass — классификация по вариабельности спроса
type XYZClass string

const (
    ClassX XYZClass = "X" // Стабильный спрос (CV < 0.1)
    ClassY XYZClass = "Y" // Умеренная вариабельность (0.1 <= CV <= 0.5)
    ClassZ XYZClass = "Z" // Нестабильный спрос (CV > 0.5)
)

// Shelf представляет место хранения с пространственными атрибутами и ёмкостью
type Shelf struct {
    ID            string
    RackID        string
    Level         int           // Уровень стеллажа: 1 = нижний, выше = менее доступен
    Priority      float64       // Бизнес-приоритет (1.0 = лучшее место)
    MaxCapacity   float64       // Максимальная ёмкость в м³
    UsedCapacity  float64       // Текущая занятая ёмкость
    Location      geom.Point    // PostGIS-точка для расчёта расстояний
    CompatibleTags []string     // Разрешённые теги товаров (пустой = все разрешены)
}

// Rack — физическая группа полок
type Rack struct {
    ID          string
    WarehouseID string
    Location    geom.Point
}

// AllocationResult — результат оптимизации размещения
type AllocationResult struct {
    ProductID string  `json:"product_id"`
    ShelfID   string  `json:"shelf_id,omitempty"`
    Score     float64 `json:"score,omitempty"`
    Assigned  bool    `json:"assigned"`
    Reason    string  `json:"reason"`
}

// TurnoverProvider — интерфейс для получения метрик оборачиваемости
// Позволяет мокать данные в тестах и переключать источники (БД / внешний сервис)
type TurnoverProvider interface {
    GetTurnoverRate(productID string) (float64, error)
    GetDemandVariability(productID string) (float64, error)
}