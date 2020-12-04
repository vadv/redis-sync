package db

import (
	"github.com/restream/reindexer"

	schema "gitlab.diskarte.net/engineering/redis-sync"
)

type db struct {
	indexer   *reindexer.Reindexer
	namespace string
}

func Open(dsn, namespace string) (schema.DB, error) {
	r := reindexer.NewReindex(dsn)
	return &db{indexer: r, namespace: namespace},
		r.OpenNamespace(namespace, reindexer.DefaultNamespaceOptions(), schema.Message{})
}

func (r *db) Set(value *schema.Message) error {
	return r.indexer.Upsert(r.namespace, value)
}

func (r *db) Get(key string) (*schema.Message, bool) {
	result, found := r.indexer.Query(r.namespace).Where("id", reindexer.EQ, key).Get()
	if !found {
		return nil, false
	}
	return result.(*schema.Message), true
}
