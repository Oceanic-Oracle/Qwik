-- Добавляем 6-й стеллаж: Ряд Е (средний приоритет, 5 полок)
INSERT INTO rack (id, aisle) VALUES (6, 'Ряд Е');

INSERT INTO shelf (rack_id, level, priority, max_capacity) VALUES
(6, 1, 75, 8.0),
(6, 2, 60, 8.0),
(6, 3, 45, 8.0),
(6, 4, 30, 8.0),
(6, 5, 15, 8.0);
