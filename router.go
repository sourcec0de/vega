package mailbox

type RouteTable interface {
	Set(string, Storage) error
	Get(string) (Storage, bool)
}

type MemRouteTable map[string]Storage

func (ht MemRouteTable) Set(name string, st Storage) error {
	ht[name] = st
	return nil
}

func (ht MemRouteTable) Get(name string) (Storage, bool) {
	s, ok := ht[name]
	return s, ok
}

type Router struct {
	routes RouteTable
}

func testRouter() *Router {
	return &Router{make(MemRouteTable)}
}

func (r *Router) Add(name string, reg Storage) {
	r.routes.Set(name, reg)
}

func (r *Router) DiscoverEndpoint(name string) (Storage, bool) {
	reg, ok := r.routes.Get(name)

	return reg, ok
}

func (r *Router) Push(name string, body []byte) error {
	if storage, ok := r.routes.Get(name); ok {
		storage.Push(name, body)
		return nil
	}

	return ENoMailbox
}
