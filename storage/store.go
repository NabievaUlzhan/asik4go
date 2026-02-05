package storage

import (
	"errors"
	"sync"
	"time"

	"foodstore/models"
)

type Store struct {
	mu sync.RWMutex

	Users      map[int]models.User
	Products   map[int]models.Product
	Orders     map[int]models.Order
	OrderItems map[int]models.OrderItem

	nextUserID      int
	nextProductID   int
	nextOrderID     int
	nextOrderItemID int

	// For background processing (goroutine worker)
	OrderQueue chan int
}

func NewStore() *Store {
	s := &Store{
		Users:      make(map[int]models.User),
		Products:   make(map[int]models.Product),
		Orders:     make(map[int]models.Order),
		OrderItems: make(map[int]models.OrderItem),

		nextUserID:      1,
		nextProductID:   1,
		nextOrderID:     1,
		nextOrderItemID: 1,

		OrderQueue: make(chan int, 100),
	}

	// Seed user
	u := models.User{
		ID:       s.nextUserID,
		Name:     "Demo User",
		Email:    "demo@example.com",
		Password: "12345",
		Role:     "customer",
	}
	s.Users[u.ID] = u
	s.nextUserID++

	p1 := models.Product{ID: s.nextProductID, Name: "Milk", Category: "Dairy", Stock: 10, Price: 650}
	s.Products[p1.ID] = p1
	s.nextProductID++

	p2 := models.Product{ID: s.nextProductID, Name: "Bread", Category: "Bakery", Stock: 15, Price: 300}
	s.Products[p2.ID] = p2
	s.nextProductID++

	return s
}

func (s *Store) GetAllProducts() []models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]models.Product, 0, len(s.Products))
	for _, p := range s.Products {
		res = append(res, p)
	}
	return res
}

func (s *Store) CreateProduct(p models.Product) models.Product {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.ID = s.nextProductID
	s.nextProductID++
	s.Products[p.ID] = p
	return p
}

func (s *Store) UpdateProduct(id int, updated models.Product) (models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Products[id]; !ok {
		return models.Product{}, errors.New("product not found")
	}
	updated.ID = id
	s.Products[id] = updated
	return updated, nil
}

func (s *Store) DeleteProduct(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Products[id]; !ok {
		return errors.New("product not found")
	}
	delete(s.Products, id)
	return nil
}

func (s *Store) CreateOrder(userID int, items []models.OrderItem) (models.Order, []models.OrderItem, error) {
	if len(items) == 0 {
		return models.Order{}, nil, errors.New("items list is empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Users[userID]; !ok {
		return models.Order{}, nil, errors.New("user not found")
	}

	for _, it := range items {
		if it.Quantity <= 0 {
			return models.Order{}, nil, errors.New("quantity must be > 0")
		}
		p, ok := s.Products[it.ProductID]
		if !ok {
			return models.Order{}, nil, errors.New("product not found")
		}
		if p.Stock < it.Quantity {
			return models.Order{}, nil, errors.New("not enough stock for product")
		}
	}

	order := models.Order{
		ID:        s.nextOrderID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	s.nextOrderID++
	s.Orders[order.ID] = order

	createdItems := make([]models.OrderItem, 0, len(items))
	for _, it := range items {
		p := s.Products[it.ProductID]
		p.Stock -= it.Quantity
		s.Products[it.ProductID] = p

		oi := models.OrderItem{
			ID:        s.nextOrderItemID,
			OrderID:   order.ID,
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		}
		s.nextOrderItemID++
		s.OrderItems[oi.ID] = oi
		createdItems = append(createdItems, oi)
	}

	select {
	case s.OrderQueue <- order.ID:
	default:
	}

	return order, createdItems, nil
}

type OrderWithItems struct {
	Order models.Order       `json:"order"`
	Items []models.OrderItem `json:"items"`
}

func (s *Store) GetUserOrders(userID int) ([]OrderWithItems, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.Users[userID]; !ok {
		return nil, errors.New("user not found")
	}

	result := []OrderWithItems{}
	for _, o := range s.Orders {
		if o.UserID != userID {
			continue
		}

		items := []models.OrderItem{}
		for _, it := range s.OrderItems {
			if it.OrderID == o.ID {
				items = append(items, it)
			}
		}

		result = append(result, OrderWithItems{
			Order: o,
			Items: items,
		})
	}

	return result, nil
}

func (s *Store) GetAllUsers() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]models.User, 0, len(s.Users))
	for _, u := range s.Users {
		res = append(res, u)
	}
	return res
}

func (s *Store) GetUserByID(id int) (models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.Users[id]
	return u, ok
}
