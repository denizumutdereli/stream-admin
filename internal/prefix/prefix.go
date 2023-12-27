package prefix

import "fmt"

type Prefix struct {
	ServicePrefixes map[string]string
	ServiceTables   map[string][]string
	TableNames      map[string]bool
}

func NewPrefixService() *Prefix {
	service := &Prefix{
		ServicePrefixes: make(map[string]string),
		ServiceTables:   make(map[string][]string),
		TableNames:      make(map[string]bool),
	}

	// TODO:config
	service.AddService("admin-action-monitoring", "administrator")
	service.AddService("admin-auth", "administrator")
	service.AddService("admin-users", "administrator")
	service.AddService("admin-user-roles", "administrator")
	service.AddService("admin-policy", "administrator")

	service.AddService("orders", "order")
	service.AddService("transactions", "transaction_manager")
	service.AddService("users", "auth")
	service.AddService("assets", "assets")
	service.AddService("file_service", "file_service")
	service.AddService("fiat_manager", "fiat_manager")

	return service
}

func (p *Prefix) AddService(serviceName, servicePrefix string) error {
	p.ServicePrefixes[serviceName] = servicePrefix
	return nil
}

func (p *Prefix) GetServicePrefix(serviceName string) (string, bool) {

	fmt.Println("Checking service:", serviceName, "Available services:", p.ServicePrefixes)

	prefix, exists := p.ServicePrefixes[serviceName]
	return prefix, exists
}

func (p *Prefix) IsTableNameExists(tableName string) bool {
	_, exists := p.TableNames[tableName]
	return exists
}

func (p *Prefix) RegisterServiceTables(servicePrefix string, tables []string) error {
	p.ServiceTables[servicePrefix] = tables
	for _, table := range tables {
		p.TableNames[table] = true
	}
	return nil
}

func (p *Prefix) GetServiceTables(serviceName string) ([]string, bool) {
	tables, exists := p.ServiceTables[serviceName]
	return tables, exists
}
