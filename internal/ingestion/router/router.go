package router

import "encoding/json"

type Router interface {
	Route(raw json.RawMessage) (*RoutedPayload, error)
}
