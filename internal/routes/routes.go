package routes

import (
	"order-management-service/internal/controller"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router, c controller.UserController, addr controller.AddressController, authMiddleware mux.MiddlewareFunc) {
	s := r.PathPrefix("/users").Subrouter()
	s.HandleFunc("/register", c.Register).Methods("POST")
	s.HandleFunc("/login", c.Login).Methods("POST")
	s.HandleFunc("/refresh-token", c.RefreshToken).Methods("POST")

	// Protected User Routes
	p := s.PathPrefix("/addresses").Subrouter()
	p.Use(authMiddleware)
	p.HandleFunc("", addr.ListAddresses).Methods("GET")
	p.HandleFunc("", addr.AddAddress).Methods("POST")
	p.HandleFunc("/{uuid}/current", addr.SetCurrent).Methods("PUT")
}

func RegisterProductRoutes(r *mux.Router, c controller.ProductController) {
	r.HandleFunc("/products", c.ListProducts).Methods("GET")
}

func RegisterOrderRoutes(r *mux.Router, c controller.OrderController, authMiddleware mux.MiddlewareFunc) {
	s := r.PathPrefix("/orders").Subrouter()
	s.Use(authMiddleware)
	s.HandleFunc("", c.CreateOrder).Methods("POST")
	s.HandleFunc("", c.ListOrders).Methods("GET")
	s.HandleFunc("/{uuid}", c.GetOrder).Methods("GET")
	s.HandleFunc("/{uuid}/cancel", c.CancelOrder).Methods("PUT")
	s.HandleFunc("/{uuid}/status", c.UpdateStatus).Methods("PUT")
}
