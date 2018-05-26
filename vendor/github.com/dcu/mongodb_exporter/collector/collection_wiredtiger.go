package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	collWTBlockManagerBlocksTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_blockmanager",
		Name:      "blocks_total",
		Help:      "The total number of blocks allocated by the WiredTiger BlockManager",
	}, []string{"ns", "type"})
)

var (
	collWTCachePages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_cache",
		Name:      "pages",
		Help:      "The current number of pages in the WiredTiger Cache",
	}, []string{"ns", "type"})
	collWTCachePagesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_cache",
		Name:      "pages_total",
		Help:      "The total number of pages read into/from the WiredTiger Cache",
	}, []string{"ns", "type"})
	collWTCacheBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_cache",
		Name:      "bytes",
		Help:      "The current size of data in the WiredTiger Cache in bytes",
	}, []string{"ns", "type"})
	collWTCacheBytesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_cache",
		Name:      "bytes_total",
		Help:      "The total number of bytes read into/from the WiredTiger Cache",
	}, []string{"ns", "type"})
	collWTCacheEvictedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_cache",
		Name:      "evicted_total",
		Help:      "The total number of pages evicted from the WiredTiger Cache",
	}, []string{"ns", "type"})
)

var (
	collWTTransactionsUpdateConflicts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_transactions",
		Name:      "update_conflicts",
		Help:      "The number of conflicts updating transactions",
	}, []string{"ns"})
)

var (
	collWTOpenCursors = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "collection_wiredtiger_session",
		Name:      "open_cursors_total",
		Help:      "The total number of cursors opened in WiredTiger",
	}, []string{"ns"})
)

// blockmanager stats
type CollWTBlockManagerStats struct {
	BlocksFreed     float64 `bson:"blocks freed"`
	BlocksAllocated float64 `bson:"blocks allocated"`
}

func (stats *CollWTBlockManagerStats) Export(ch chan<- prometheus.Metric, collection string) {
	collWTBlockManagerBlocksTotal.WithLabelValues(collection, "freed").Set(stats.BlocksFreed)
	collWTBlockManagerBlocksTotal.WithLabelValues(collection, "allocated").Set(stats.BlocksAllocated)
}

func (stats *CollWTBlockManagerStats) Describe(ch chan<- *prometheus.Desc) {
	collWTBlockManagerBlocksTotal.Describe(ch)
}

// cache stats
type CollWTCacheStats struct {
	BytesTotal        float64 `bson:"bytes currently in the cache"`
	BytesDirty        float64 `bson:"tracked dirty bytes in the cache"`
	BytesReadInto     float64 `bson:"bytes read into cache"`
	BytesWrittenFrom  float64 `bson:"bytes written from cache"`
	EvictedUnmodified float64 `bson:"unmodified pages evicted"`
	EvictedModified   float64 `bson:"modified pages evicted"`
	PagesReadInto     float64 `bson:"pages read into cache"`
	PagesWrittenFrom  float64 `bson:"pages written from cache"`
}

func (stats *CollWTCacheStats) Export(ch chan<- prometheus.Metric, collection string) {
	collWTCachePagesTotal.WithLabelValues(collection, "read").Set(stats.PagesReadInto)
	collWTCachePagesTotal.WithLabelValues(collection, "written").Set(stats.PagesWrittenFrom)
	collWTCacheBytesTotal.WithLabelValues(collection, "read").Set(stats.BytesReadInto)
	collWTCacheBytesTotal.WithLabelValues(collection, "written").Set(stats.BytesWrittenFrom)
	collWTCacheEvictedTotal.WithLabelValues(collection, "modified").Set(stats.EvictedModified)
	collWTCacheEvictedTotal.WithLabelValues(collection, "unmodified").Set(stats.EvictedUnmodified)
	collWTCacheBytes.WithLabelValues(collection, "total").Set(stats.BytesTotal)
	collWTCacheBytes.WithLabelValues(collection, "dirty").Set(stats.BytesDirty)
}

func (stats *CollWTCacheStats) Describe(ch chan<- *prometheus.Desc) {
	collWTCachePagesTotal.Describe(ch)
	collWTCacheEvictedTotal.Describe(ch)
	collWTCachePages.Describe(ch)
	collWTCacheBytes.Describe(ch)
}

// session stats
type CollWTSessionStats struct {
	Cursors float64 `bson:"open cursor count"`
}

func (stats *CollWTSessionStats) Export(ch chan<- prometheus.Metric, collection string) {
	collWTOpenCursors.WithLabelValues(collection).Set(stats.Cursors)
}

func (stats *CollWTSessionStats) Describe(ch chan<- *prometheus.Desc) {
	collWTOpenCursors.Describe(ch)
}

// transaction stats
type CollWTTransactionStats struct {
	UpdateConflicts float64 `bson:"update conflicts"`
}

func (stats *CollWTTransactionStats) Export(ch chan<- prometheus.Metric, collection string) {
	collWTTransactionsUpdateConflicts.WithLabelValues(collection).Set(stats.UpdateConflicts)
}

func (stats *CollWTTransactionStats) Describe(ch chan<- *prometheus.Desc) {
	collWTTransactionsUpdateConflicts.Describe(ch)
}

// WiredTiger stats
type CollWiredTigerStats struct {
	BlockManager *CollWTBlockManagerStats `bson:"block-manager"`
	Cache        *CollWTCacheStats        `bson:"cache"`
	Session      *CollWTSessionStats      `bson:"session"`
	Transaction  *CollWTTransactionStats  `bson:"transaction"`
}

func (stats *CollWiredTigerStats) Describe(ch chan<- *prometheus.Desc) {
	if stats.BlockManager != nil {
		stats.BlockManager.Describe(ch)
	}

	if stats.Cache != nil {
		stats.Cache.Describe(ch)
	}
	if stats.Transaction != nil {
		stats.Transaction.Describe(ch)
	}
	if stats.Session != nil {
		stats.Session.Describe(ch)
	}
}

func (stats *CollWiredTigerStats) Export(ch chan<- prometheus.Metric, collection string) {
	if stats.BlockManager != nil {
		stats.BlockManager.Export(ch, collection)
	}

	if stats.Cache != nil {
		stats.Cache.Export(ch, collection)
	}

	if stats.Transaction != nil {
		stats.Transaction.Export(ch, collection)
	}

	if stats.Session != nil {
		stats.Session.Export(ch, collection)
	}

	collWTBlockManagerBlocksTotal.Collect(ch)
	collWTCachePagesTotal.Collect(ch)
	collWTCacheBytesTotal.Collect(ch)
	collWTCacheEvictedTotal.Collect(ch)
	collWTCachePages.Collect(ch)
	collWTCacheBytes.Collect(ch)
	collWTTransactionsUpdateConflicts.Collect(ch)
	collWTOpenCursors.Collect(ch)

	collWTBlockManagerBlocksTotal.Reset()
	collWTCachePagesTotal.Reset()
	collWTCacheBytesTotal.Reset()
	collWTCacheEvictedTotal.Reset()
	collWTCachePages.Reset()
	collWTCacheBytes.Reset()
	collWTTransactionsUpdateConflicts.Reset()
	collWTOpenCursors.Reset()
}
