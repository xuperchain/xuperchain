package acmstate

import (
	"sync"
)

type metadataInfo struct {
	metadata string
	updated  bool
}

type MetadataCache struct {
	backend MetadataReader
	m       sync.Map
}

func NewMetadataCache(backend MetadataReader) *MetadataCache {
	return &MetadataCache{
		backend: backend,
	}
}

func (cache *MetadataCache) SetMetadata(metahash MetadataHash, metadata string) error {
	cache.m.Store(metahash, &metadataInfo{updated: true, metadata: metadata})
	return nil
}

func (cache *MetadataCache) GetMetadata(metahash MetadataHash) (string, error) {
	metaInfo, err := cache.getMetadata(metahash)
	if err != nil {
		return "", err
	}

	return metaInfo.metadata, nil
}

// Syncs changes to the backend in deterministic order. Sends storage updates before updating
// the account they belong so that storage values can be taken account of in the update.
func (cache *MetadataCache) Sync(st MetadataWriter) error {
	var err error
	cache.m.Range(func(key, value interface{}) bool {
		hash := key.(MetadataHash)
		info := value.(*metadataInfo)
		if info.updated {
			err = st.SetMetadata(hash, info.metadata)
			if err != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

func (cache *MetadataCache) Reset(backend MetadataReader) {
	cache.backend = backend
	cache.m = sync.Map{}
}

// Get the cache accountInfo item creating it if necessary
func (cache *MetadataCache) getMetadata(metahash MetadataHash) (*metadataInfo, error) {
	value, ok := cache.m.Load(metahash)
	if !ok {
		metadata, err := cache.backend.GetMetadata(metahash)
		if err != nil {
			return nil, err
		}
		metaInfo := &metadataInfo{
			metadata: metadata,
		}
		cache.m.Store(metahash, metaInfo)
		return metaInfo, nil
	}
	return value.(*metadataInfo), nil
}
