package datacache

import "context"

type DataCache interface {
	Get(ctx context.Context, name string) (interface{}, error)
	List(ctx context.Context) ([]interface{}, error)
}
