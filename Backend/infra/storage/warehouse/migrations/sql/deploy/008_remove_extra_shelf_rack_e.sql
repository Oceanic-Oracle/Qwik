-- Удаляем лишнюю 6-ю полку (уровень 6) у стеллажа Ряд Е
DELETE FROM shelf WHERE rack_id = 6 AND level = 6;
