package service

import (
	"context"
	"log/slog"
	"sort"
	"sync"

	"warehouse/internal/config"
	"warehouse/internal/domain"
	"warehouse/internal/repo/shelf"
)

// OptimizationService реализует 6-этапный алгоритм оптимизации склада
type OptimizationService struct {
	shelfRepo *shelf.ShelfRepo
	cfg       config.OptimizerConfig
	log       *slog.Logger
	turnover  domain.TurnoverProvider
}

func NewOptimizationService(
	shelfRepo *shelf.ShelfRepo,
	cfg config.OptimizerConfig,
	log *slog.Logger,
	turnover domain.TurnoverProvider,
) *OptimizationService {
	return &OptimizationService{
		shelfRepo: shelfRepo,
		cfg:       cfg,
		log:       log,
		turnover:  turnover,
	}
}

// OptimizeAllocations запускает полный 6-этапный алгоритм
func (s *OptimizationService) OptimizeAllocations(
	ctx context.Context,
	products []*domain.Product,
	warehouseID string,
) ([]*domain.AllocationResult, error) {
	s.log.Info("запуск оптимизации", 
		"products_count", len(products),
		"warehouse_id", warehouseID)

	// Шаг 1: Агрегация данных — обогащаем товары метриками оборачиваемости
	s.log.Debug("шаг 1: агрегация данных об оборачиваемости")
	for _, prod := range products {
		rate, err := s.turnover.GetTurnoverRate(prod.ID)
		if err != nil {
			s.log.Warn("не удалось получить оборачиваемость", "product_id", prod.ID, "error", err)
			continue
		}
		prod.TurnoverRate = rate

		cv, err := s.turnover.GetDemandVariability(prod.ID)
		if err != nil {
			s.log.Warn("не удалось получить CV спроса", "product_id", prod.ID, "error", err)
			continue
		}
		prod.DemandCV = cv
	}

	// Шаг 2: ABC-XYZ классификация
	s.log.Debug("шаг 2: классификация товаров")
	s.ClassifyProducts(products)

	// Шаги 3-4: Скоринг полок + сопоставление (объединены для эффективности)
	s.log.Debug("шаги 3-4: расчёт скоринга и размещение")
	results := s.allocateProducts(ctx, products, warehouseID)

	// Шаги 5 и 6 обрабатываются внутри allocateProducts (ограничения + конкурентность)

	s.log.Info("оптимизация завершена", 
		"allocated", countAllocated(results),
		"total", len(results))

	return results, nil
}

// ClassifyProducts назначает классы ABC и XYZ на основе порогов из конфига
// Формула: ABC — по кумулятивной ценности оборота; XYZ — по коэффициенту вариации (CV)
func (s *OptimizationService) ClassifyProducts(products []*domain.Product) {
	// ABC-классификация: сортировка по ценности (оборот * объём)
	type productValue struct {
		prod  *domain.Product
		value float64
	}
	
	values := make([]productValue, len(products))
	totalValue := 0.0
	for i, p := range products {
		val := p.TurnoverRate * p.AllocatedCapacity
		values[i] = productValue{prod: p, value: val}
		totalValue += val
	}

	// Сортировка по убыванию ценности
	sort.Slice(values, func(i, j int) bool {
		return values[i].value > values[j].value
	})

	// Назначение классов ABC по кумулятивному проценту
	cumulative := 0.0
	for i := range values {
		cumulative += values[i].value / totalValue
		if cumulative <= s.cfg.ABCCutoff {
			values[i].prod.ABCClass = domain.ClassA
		} else if cumulative <= s.cfg.BCxCutoff {
			values[i].prod.ABCClass = domain.ClassB
		} else {
			values[i].prod.ABCClass = domain.ClassC
		}
	}

	// XYZ-классификация: по коэффициенту вариации спроса (CV)
	for _, p := range products {
		if p.DemandCV < s.cfg.XYZCutoffLow {
			p.XYZClass = domain.ClassX
		} else if p.DemandCV <= s.cfg.XYZCutoffHigh {
			p.XYZClass = domain.ClassY
		} else {
			p.XYZClass = domain.ClassZ
		}
	}
}

// CalculateScore вычисляет I_shelf для пары товар-полка
// Формула: I_shelf = W1 * P_level + W2 * P_priority - W3 * Distance_km
// Примечание: уровень инвертируется (ниже = лучше), поэтому используем (maxLevel - level + 1)
func (s *OptimizationService) CalculateScore(
	product *domain.Product,
	shelf *domain.Shelf,
	maxShelfLevel int,
) float64 {
	// Нормализация уровня: нижний физический уровень = более высокий приоритет
	levelScore := float64(maxShelfLevel - shelf.Level + 1) / float64(maxShelfLevel)
	
	// Priority уже нормализован [0,1]
	priorityScore := shelf.Priority
	
	// Расстояние: в продакшене использовать PostGIS ST_Distance с правильным SRID
	distance := 0.0 // Заглушка: вычисляется из shelf.Location относительно центра склада
	
	// Применяем веса из конфига
	// Формула: I_shelf = W1*P_level + W2*P_priority - W3*Distance
	score := s.cfg.WeightLevel*levelScore + 
			 s.cfg.WeightPriority*priorityScore - 
			 s.cfg.WeightDistance*distance

	// Буст скоринга для совпадения ABC-класса: товары класса A предпочитают полки с высоким скорингом
	if product.ABCClass == domain.ClassA && score > 0.7 {
		score *= 1.2
	} else if product.ABCClass == domain.ClassC && score < 0.4 {
		score *= 1.1 // Небольшой буст для заполнения низкоприоритетных полок
	}

	return score
}

// allocateProducts реализует жадный алгоритм сопоставления с проверкой ограничений
func (s *OptimizationService) allocateProducts(
	ctx context.Context,
	products []*domain.Product,
	warehouseID string,
) []*domain.AllocationResult {
	results := make([]*domain.AllocationResult, 0, len(products))
	
	// Сортировка товаров: сначала класс A, затем по оборачиваемости (жадный приоритет)
	sort.Slice(products, func(i, j int) bool {
		if products[i].ABCClass != products[j].ABCClass {
			return products[i].ABCClass < products[j].ABCClass // A < B < C лексикографически
		}
		return products[i].TurnoverRate > products[j].TurnoverRate
	})

	// Оценка максимального уровня полки для нормализации скоринга (можно кэшировать)
	maxLevel := 5 // В продакшене: запрос SELECT MAX(level) FROM shelf

	// Обработка товаров конкурентно с построчной блокировкой
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Ограничение параллельных операций с БД

	for _, prod := range products {
		wg.Add(1)
		sem <- struct{}{}
		
		go func(p *domain.Product) {
			defer wg.Done()
			defer func() { <-sem }()

			result := &domain.AllocationResult{
				ProductID: p.ID,
				Assigned:  false,
			}

			// Получение кандидатов-полок
			weights := map[string]float64{
				"level":    s.cfg.WeightLevel,
				"priority": s.cfg.WeightPriority,
				"distance": s.cfg.WeightDistance,
			}
			
			shelves, err := s.shelfRepo.GetAvailableShelves(
				ctx, p.AllocatedCapacity, p.Tags, warehouseID, weights)
			if err != nil {
				s.log.Error("ошибка получения полок", "product_id", p.ID, "error", err)
				result.Reason = "shelf_query_error"
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				return
			}

			// Расчёт скоринга и сортировка
			type scoredShelf struct {
				shelf *domain.Shelf
				score float64
			}
			scored := make([]scoredShelf, 0, len(shelves))
			for _, shelf := range shelves {
				// Шаг 5: Проверка ограничений — совместимость тегов
				compatible, err := s.shelfRepo.CheckTagCompatibility(ctx, shelf.ID, p.Tags)
				if err != nil {
					s.log.Warn("ошибка проверки тегов", "shelf_id", shelf.ID, "error", err)
					continue
				}
				if !compatible {
					continue
				}

				score := s.CalculateScore(p, shelf, maxLevel)
				scored = append(scored, scoredShelf{shelf: shelf, score: score})
			}

			// Сортировка по убыванию скоринга (лучшие полки первыми)
			sort.Slice(scored, func(i, j int) bool {
				return scored[i].score > scored[j].score
			})

			// Жадное назначение с атомарным резервированием ёмкости
			for _, candidate := range scored {
				// Шаг 6: Безопасное обновление ёмкости с блокировкой
				err := s.shelfRepo.ReserveCapacity(ctx, candidate.shelf.ID, p.AllocatedCapacity)
				if err != nil {
					s.log.Debug("резервирование не удалось", 
						"product_id", p.ID, 
						"shelf_id", candidate.shelf.ID,
						"error", err)
					continue // Пробуем следующую полку
				}

				// Успех!
				result.ShelfID = candidate.shelf.ID
				result.Score = candidate.score
				result.Assigned = true
				result.Reason = "allocated"
				s.log.Info("товар размещён",
					"product_id", p.ID,
					"shelf_id", candidate.shelf.ID,
					"score", candidate.score,
					"abc_class", p.ABCClass)
				break
			}

			if !result.Assigned && result.Reason == "" {
				result.Reason = "no_compatible_shelf"
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(prod)
	}

	wg.Wait()
	return results
}

func countAllocated(results []*domain.AllocationResult) int {
	count := 0
	for _, r := range results {
		if r.Assigned {
			count++
		}
	}
	return count
}