package service

import (
    "context"
    "log/slog"
    "sort"
    "sync"

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
    s.log.Info("запуск оптимизации", "items_count", len(items))

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
    }
    s.ClassifyProducts(products)

    results := s.allocateItems(ctx, products, items)

    s.log.Info("оптимизация завершена", "allocated", countAllocated(results), "total", len(results))
    return results, nil
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

func (s *OptimizationService) CalculateScore(product *domain.Product, level int, priority float64, maxShelfLevel int) float64 {
    levelScore := float64(maxShelfLevel-level+1) / float64(maxShelfLevel)
    score := s.cfg.WeightLevel*levelScore + s.cfg.WeightPriority*priority 
    if product.ABCClass == domain.ClassA && score > 0.7 {
        score *= 1.2
    }
    return score
}

func (s *OptimizationService) allocateItems(
    ctx context.Context,
    products []*domain.Product,
    items []dto.OptimizeItem,
) []*domain.AllocationResult {
    results := make([]*domain.AllocationResult, 0, len(items))
    maxLevel := 5 

    var mu sync.Mutex
    var wg sync.WaitGroup
    sem := make(chan struct{}, 10)

    for i, prod := range products {
        wg.Add(1)
        sem <- struct{}{}
        
        go func(idx int, p *domain.Product) {
            defer wg.Done()
            defer func() { <-sem }()

            qty := items[idx].Quantity
            result := &domain.AllocationResult{
                ProductID: p.ID,
                Assigned:  false,
            }

            volume, err := s.shelfRepo.GetProductVolume(ctx, p.ID)
            if err != nil || volume == 0 {
                s.log.Warn("нельзя разместить товар без объема", "product_id", p.ID, "error", err)
                result.Reason = "product_has_no_volume_or_not_found"
                mu.Lock()
                results = append(results, result)
                mu.Unlock()
                return
            }

            requiredCapacity := float64(qty) * volume
            
            // ИСПРАВЛЕНО: Ищем полки под конкретный товар, чтобы не смешивать SKU
            shelves, err := s.shelfRepo.GetAvailableShelves(ctx, requiredCapacity, p.ID)
            if err != nil || len(shelves) == 0 {
                result.Reason = "no_shelves_found"
                mu.Lock()
                results = append(results, result)
                mu.Unlock()
                return
            }

            type scoredShelf struct {
                shelf *shelf.ShelfEntity
                score float64
            }
            scored := make([]scoredShelf, 0, len(shelves))
            for _, sh := range shelves {
                score := s.CalculateScore(p, sh.Level, sh.Priority, maxLevel)
                scored = append(scored, scoredShelf{shelf: sh, score: score})
            }
            sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })

            for _, candidate := range scored {
                err := s.shelfRepo.AllocateProduct(ctx, candidate.shelf.ID, p.ID, qty)
                if err != nil {
                    s.log.Debug("резервирование не удалось", "product_id", p.ID, "shelf_id", candidate.shelf.ID, "error", err)
                    continue 
                }
                result.ShelfID = candidate.shelf.ID
                result.Score = candidate.score
                result.Assigned = true
                result.Reason = "allocated"
                break
            }

            if !result.Assigned && result.Reason == "" {
                result.Reason = "no_compatible_shelf_with_space"
            }
            mu.Lock()
            results = append(results, result)
            mu.Unlock()
        }(i, prod)
    }
    wg.Wait()
    return results
}

func countAllocated(results []*domain.AllocationResult) int {
    count := 0
    for _, r := range results { if r.Assigned { count++ } }
    return count
}