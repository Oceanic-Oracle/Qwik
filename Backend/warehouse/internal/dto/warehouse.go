package dto

type WarehouseMap struct {
    Racks []*RackState `json:"racks"`
}

type RackState struct {
    ID      string        `json:"id"`
    Aisle   string        `json:"aisle"`
    Shelves []*ShelfState `json:"shelves"`
}

type ShelfState struct {
    ID           string       `json:"id"`
    Level        int          `json:"level"`
    MaxCapacity  float64      `json:"max_capacity"`
    UsedCapacity float64      `json:"used_capacity"`
    Product      *ProductCell `json:"product,omitempty"`
}

type ProductCell struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Quantity int    `json:"quantity"`
}