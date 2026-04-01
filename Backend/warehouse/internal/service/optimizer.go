package service

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "sort"

    "warehouse/internal/config"
    "warehouse/internal/domain"
    "warehouse/internal/dto"
    "warehouse/internal/repo/shelf"
)

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

func (s *OptimizationService) OptimizeAllocations(
    ctx context.Context,
    items []dto.OptimizeItem,
) ([]*domain.AllocationResult, error) {
    s.log.Info("запуск ABC/XYZ оптимизации", "items_count", len(items))

    products := make([]*domain.Product, len(items))
    for i, item := range items {
        products[i] = &domain.Product{
            ID:   item.ProductID,
            Tags: []string{"general"},
        }
        rate, _ := s.turnover.GetTurnoverRate(item.ProductID)
        cv, _ := s.turnover.GetDemandVariability(item.ProductID)
        products[i].TurnoverRate = rate
        products[i].DemandCV = cv

        vol, _ := s.shelfRepo.GetProductVolume(ctx, item.ProductID)
        products[i].AllocatedCapacity = vol
    }

    s.ClassifyProducts(products)
    results := s.allocateItemsSequentially(ctx, products, items)

    s.log.Info("оптимизация завершена", "allocated", countAllocated(results), "total", len(results))
    return results, nil
}

func (s *OptimizationService) OptimizeAllAllocations(ctx context.Context) ([]*domain.AllocationResult, error) {
    s.log.Info("запуск ГЛОБАЛЬНОЙ перестройки склада")

    items, err := s.shelfRepo.GetStockedItems(ctx)
    if err != nil {
        return nil, fmt.Errorf("не удалось получить список товаров: %w", err)
    }

    if len(items) == 0 {
        s.log.Info("склад уже пуст, нечего оптимизировать")
        return []*domain.AllocationResult{}, nil
    }

    if err := s.shelfRepo.ClearWarehouseForOptimization(ctx); err != nil {
        return nil, fmt.Errorf("не удалось очистить склад: %w", err)
    }
    
    s.log.Info("склад очищен, начинаем расчет ABC/XYZ", "items_to_place", len(items))

    return s.OptimizeAllocations(ctx, items)
}

func (s *OptimizationService) ClassifyProducts(products []*domain.Product) {
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

    if totalValue == 0 {
        totalValue = 1
    }

    sort.Slice(values, func(i, j int) bool { return values[i].value > values[j].value })
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

func (s *OptimizationService) calculateSlottingScore(p *domain.Product) float64 {
    abcScore := 0.0
    switch p.ABCClass {
    case domain.ClassA: abcScore = 100.0
    case domain.ClassB: abcScore = 50.0
    case domain.ClassC: abcScore = 0.0
    }

    xyzScore := 0.0
    switch p.XYZClass {
    case domain.ClassZ: xyzScore = 100.0
    case domain.ClassY: xyzScore = 50.0
    case domain.ClassX: xyzScore = 0.0
    }

    return (s.cfg.WeightPriority * abcScore) + (s.cfg.WeightLevel * xyzScore)
}

func (s *OptimizationService) allocateItemsSequentially(
    ctx context.Context,
    products []*domain.Product,
    items []dto.OptimizeItem,
) []*domain.AllocationResult {
    type indexedProduct struct {
        origIdx int
        prod    *domain.Product
        score   float64
    }

    scoredProducts := make([]indexedProduct, len(products))
    for i, p := range products {
        scoredProducts[i] = indexedProduct{
            origIdx: i,
            prod:    p,
            score:   s.calculateSlottingScore(p),
        }
    }

    sort.Slice(scoredProducts, func(i, j int) bool {
        return scoredProducts[i].score > scoredProducts[j].score
    })

    results := make([]*domain.AllocationResult, len(items))

    for _, sp := range scoredProducts {
        qty := items[sp.origIdx].Quantity
        result := &domain.AllocationResult{
            ProductID: sp.prod.ID,
            Assigned:  false,
        }

        volume, err := s.shelfRepo.GetProductVolume(ctx, sp.prod.ID)
        if err != nil || volume == 0 {
            s.log.Warn("нельзя разместить товар без объема", "product_id", sp.prod.ID, "error", err)
            result.Reason = "product_has_no_volume_or_not_found"
            results[sp.origIdx] = result
            continue
        }

        // ИСПРАВЛЕНИЕ: Ищем полки с минимальным порогом (чтобы влезла хотя бы 1 шт)
        minReqCap := volume 
        shelves, err := s.shelfRepo.GetShelvesForOptimization(ctx, minReqCap)
        if err != nil || len(shelves) == 0 {
            result.Reason = "no_shelves_found"
            results[sp.origIdx] = result
            continue
        }

        remaining := qty
        var firstShelfID string
        totalPlaced := 0
        
        // ИСПРАВЛЕНИЕ: Идем по полкам и дробим партию, пока не разложим весь товар
        for _, candidate := range shelves {
            if remaining <= 0 {
                break
            }

            // Локально вычисляем свободное место (обновляем в памяти, чтобы не делать запросы в БД)
            freeSpace := candidate.MaxCapacity - candidate.UsedCapacity
            if freeSpace <= 0 {
                continue
            }

            fitQty := int(freeSpace / volume)
            if fitQty <= 0 {
                continue
            }

            actualQty := remaining
            if fitQty < remaining {
                actualQty = fitQty
            }

            err := s.shelfRepo.AllocateProduct(ctx, candidate.ID, sp.prod.ID, actualQty)
            if err == nil {
                if firstShelfID == "" {
                    firstShelfID = candidate.ID
                }
                remaining -= actualQty
                totalPlaced += actualQty
                
                // Обновляем локальное состояние полки
                candidate.UsedCapacity += float64(actualQty) * volume
            } else {
                s.log.Debug("ошибка размещения части товара", "shelf_id", candidate.ID, "error", err)
                if errors.Is(err, shelf.ErrNoSpace) {
                    continue
                }
                break 
            }
        }

        if remaining == 0 {
            result.ShelfID = firstShelfID
            result.Score = sp.score
            result.Assigned = true
            result.Reason = fmt.Sprintf("assigned_by_abc_xyz_score_%.1f", sp.score)
        } else {
            result.Reason = fmt.Sprintf("no_compatible_shelf_with_space (placed %d out of %d)", totalPlaced, qty)
        }

        results[sp.origIdx] = result
    }

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