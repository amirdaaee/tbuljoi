package bot

import (
	"sync"
	"time"

	"github.com/amirdaaee/tbuljoi/internal/settings"
	cache "github.com/patrickmn/go-cache"
)

var _afCache *cache.Cache
var aFCacheOnce sync.Once
var _afRelaxCache *cache.Cache
var aFRelaxCacheOnce sync.Once

func GetAFCache() *cache.Cache {
	aFCacheOnce.Do(func() {
		_afCache = cache.New(cache.NoExpiration, 10*time.Minute)
	})
	return _afCache
}
func GetAFRelaxCache() *cache.Cache {
	aFRelaxCacheOnce.Do(func() {
		_afRelaxCache = cache.New(time.Duration(settings.Config().AFRelax)*time.Second, 1*time.Second)
	})
	return _afRelaxCache
}
