-- Добавляем секцию (позицию по ширине) к полке
-- Существующие полки получают section = 1 (одна секция по умолчанию)
ALTER TABLE shelf ADD COLUMN section INTEGER NOT NULL DEFAULT 1;

-- Уникальность: одна секция = одна физическая ячейка (стеллаж + уровень + секция)
ALTER TABLE shelf ADD CONSTRAINT uq_shelf_slot UNIQUE (rack_id, level, section);
