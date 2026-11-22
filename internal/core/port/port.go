// Package port holds every interface the core depends on or offers.
package port

// Repositories bundles the driven ports, one per datasource.
type Repositories struct {
	Repo      Repo
	RedisRepo RedisRepo
	Adpt      Adapter
}
