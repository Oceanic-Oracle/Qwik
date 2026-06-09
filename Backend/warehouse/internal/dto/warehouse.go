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
    ID           string         `json:"id"`
    Level        int            `json:"level"`
    Section      int            `json:"section"`
    MaxCapacity  float64        `json:"max_capacity"`
    UsedCapacity float64        `json:"used_capacity"`
    Products     []*ProductCell `json:"products,omitempty"`
}

type ProductCell struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Quantity int    `json:"quantity"`
}