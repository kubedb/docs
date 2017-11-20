package exporter

import (
	"math"
	"strconv"
	"time"

	"github.com/Snapbug/gomemcache/memcache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	Namespace = "memcached"
)

// Exporter collects metrics from a memcached server.
type Exporter struct {
	mc *memcache.Client

	up                    *prometheus.Desc
	uptime                *prometheus.Desc
	version               *prometheus.Desc
	bytesRead             *prometheus.Desc
	bytesWritten          *prometheus.Desc
	currentConnections    *prometheus.Desc
	maxConnections        *prometheus.Desc
	connectionsTotal      *prometheus.Desc
	currentBytes          *prometheus.Desc
	limitBytes            *prometheus.Desc
	commands              *prometheus.Desc
	items                 *prometheus.Desc
	itemsTotal            *prometheus.Desc
	evictions             *prometheus.Desc
	reclaimed             *prometheus.Desc
	malloced              *prometheus.Desc
	itemsNumber           *prometheus.Desc
	itemsAge              *prometheus.Desc
	itemsCrawlerReclaimed *prometheus.Desc
	itemsEvicted          *prometheus.Desc
	itemsEvictedNonzero   *prometheus.Desc
	itemsEvictedTime      *prometheus.Desc
	itemsEvictedUnfetched *prometheus.Desc
	itemsExpiredUnfetched *prometheus.Desc
	itemsOutofmemory      *prometheus.Desc
	itemsReclaimed        *prometheus.Desc
	itemsTailrepairs      *prometheus.Desc
	slabsChunkSize        *prometheus.Desc
	slabsChunksPerPage    *prometheus.Desc
	slabsCurrentPages     *prometheus.Desc
	slabsCurrentChunks    *prometheus.Desc
	slabsChunksUsed       *prometheus.Desc
	slabsChunksFree       *prometheus.Desc
	slabsChunksFreeEnd    *prometheus.Desc
	slabsMemRequested     *prometheus.Desc
	slabsCommands         *prometheus.Desc
}

// NewExporter returns an initialized exporter.
func NewExporter(server string, timeout time.Duration) *Exporter {
	c := memcache.New(server)
	c.Timeout = timeout

	return &Exporter{
		mc: c,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "up"),
			"Could the memcached server be reached.",
			nil,
			nil,
		),
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "uptime_seconds"),
			"Number of seconds since the server started.",
			nil,
			nil,
		),
		version: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "version"),
			"The version of this memcached server.",
			[]string{"version"},
			nil,
		),
		bytesRead: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "read_bytes_total"),
			"Total number of bytes read by this server from network.",
			nil,
			nil,
		),
		bytesWritten: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "written_bytes_total"),
			"Total number of bytes sent by this server to network.",
			nil,
			nil,
		),
		currentConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "current_connections"),
			"Current number of open connections.",
			nil,
			nil,
		),
		maxConnections: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "max_connections"),
			"Maximum number of clients allowed.",
			nil,
			nil,
		),
		connectionsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "connections_total"),
			"Total number of connections opened since the server started running.",
			nil,
			nil,
		),
		currentBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "current_bytes"),
			"Current number of bytes used to store items.",
			nil,
			nil,
		),
		limitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "limit_bytes"),
			"Number of bytes this server is allowed to use for storage.",
			nil,
			nil,
		),
		commands: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "commands_total"),
			"Total number of all requests broken down by command (get, set, etc.) and status.",
			[]string{"command", "status"},
			nil,
		),
		items: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "current_items"),
			"Current number of items stored by this instance.",
			nil,
			nil,
		),
		itemsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "items_total"),
			"Total number of items stored during the life of this instance.",
			nil,
			nil,
		),
		evictions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "items_evicted_total"),
			"Total number of valid items removed from cache to free memory for new items.",
			nil,
			nil,
		),
		reclaimed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "items_reclaimed_total"),
			"Total number of times an entry was stored using memory from an expired entry.",
			nil,
			nil,
		),
		malloced: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "malloced_bytes"),
			"Number of bytes of memory allocated to slab pages.",
			nil,
			nil,
		),
		itemsNumber: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "current_items"),
			"Number of items currently stored in this slab class.",
			[]string{"slab"},
			nil,
		),
		itemsAge: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_age_seconds"),
			"Number of seconds the oldest item has been in the slab class.",
			[]string{"slab"},
			nil,
		),
		itemsCrawlerReclaimed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_crawler_reclaimed_total"),
			"Total number of items freed by the LRU Crawler.",
			[]string{"slab"},
			nil,
		),
		itemsEvicted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_evicted_total"),
			"Total number of times an item had to be evicted from the LRU before it expired.",
			[]string{"slab"},
			nil,
		),
		itemsEvictedNonzero: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_evicted_nonzero_total"),
			"Total number of times an item which had an explicit expire time set had to be evicted from the LRU before it expired.",
			[]string{"slab"},
			nil,
		),
		itemsEvictedTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_evicted_time_seconds"),
			"Seconds since the last access for the most recent item evicted from this class.",
			[]string{"slab"},
			nil,
		),
		itemsEvictedUnfetched: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_evicted_unfetched_total"),
			"Total nmber of items evicted and never fetched.",
			[]string{"slab"},
			nil,
		),
		itemsExpiredUnfetched: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_expired_unfetched_total"),
			"Total number of valid items evicted from the LRU which were never touched after being set.",
			[]string{"slab"},
			nil,
		),
		itemsOutofmemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_outofmemory_total"),
			"Total number of items for this slab class that have triggered an out of memory error.",
			[]string{"slab"},
			nil,
		),
		itemsReclaimed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_reclaimed_total"),
			"Total number of items reclaimed.",
			[]string{"slab"},
			nil,
		),
		itemsTailrepairs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "items_tailrepairs_total"),
			"Total number of times the entries for a particular ID need repairing.",
			[]string{"slab"},
			nil,
		),
		slabsChunkSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "chunk_size_bytes"),
			"Number of bytes allocated to each chunk within this slab class.",
			[]string{"slab"},
			nil,
		),
		slabsChunksPerPage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "chunks_per_page"),
			"Number of chunks within a single page for this slab class.",
			[]string{"slab"},
			nil,
		),
		slabsCurrentPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "current_pages"),
			"Number of pages allocated to this slab class.",
			[]string{"slab"},
			nil,
		),
		slabsCurrentChunks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "current_chunks"),
			"Number of chunks allocated to this slab class.",
			[]string{"slab"},
			nil,
		),
		slabsChunksUsed: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "chunks_used"),
			"Number of chunks allocated to an item.",
			[]string{"slab"},
			nil,
		),
		slabsChunksFree: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "chunks_free"),
			"Number of chunks not yet allocated items.",
			[]string{"slab"},
			nil,
		),
		slabsChunksFreeEnd: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "chunks_free_end"),
			"Number of free chunks at the end of the last allocated page.",
			[]string{"slab"},
			nil,
		),
		slabsMemRequested: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "mem_requested_bytes"),
			"Number of bytes of memory actual items take up within a slab.",
			[]string{"slab"},
			nil,
		),
		slabsCommands: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "slab", "commands_total"),
			"Total number of all requests broken down by command (get, set, etc.) and status per slab.",
			[]string{"slab", "command", "status"},
			nil,
		),
	}
}

// Describe describes all the metrics exported by the memcached exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.uptime
	ch <- e.version
	ch <- e.bytesRead
	ch <- e.bytesWritten
	ch <- e.currentConnections
	ch <- e.maxConnections
	ch <- e.connectionsTotal
	ch <- e.currentBytes
	ch <- e.limitBytes
	ch <- e.commands
	ch <- e.items
	ch <- e.itemsTotal
	ch <- e.evictions
	ch <- e.reclaimed
	ch <- e.malloced
	ch <- e.itemsNumber
	ch <- e.itemsAge
	ch <- e.itemsCrawlerReclaimed
	ch <- e.itemsEvicted
	ch <- e.itemsEvictedNonzero
	ch <- e.itemsEvictedTime
	ch <- e.itemsEvictedUnfetched
	ch <- e.itemsExpiredUnfetched
	ch <- e.itemsOutofmemory
	ch <- e.itemsReclaimed
	ch <- e.itemsTailrepairs
	ch <- e.itemsExpiredUnfetched
	ch <- e.slabsChunkSize
	ch <- e.slabsChunksPerPage
	ch <- e.slabsCurrentPages
	ch <- e.slabsCurrentChunks
	ch <- e.slabsChunksUsed
	ch <- e.slabsChunksFree
	ch <- e.slabsChunksFreeEnd
	ch <- e.slabsMemRequested
	ch <- e.slabsCommands
}

// Collect fetches the statistics from the configured memcached server, and
// delivers them as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	stats, err := e.mc.Stats()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		log.Errorf("Failed to collect stats from memcached: %s", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	// TODO(ts): Clean up and consolidate metric mappings.
	itemsMetrics := map[string]*prometheus.Desc{
		"crawler_reclaimed": e.itemsCrawlerReclaimed,
		"evicted":           e.itemsEvicted,
		"evicted_nonzero":   e.itemsEvictedNonzero,
		"evicted_time":      e.itemsEvictedTime,
		"evicted_unfetched": e.itemsEvictedUnfetched,
		"expired_unfetched": e.itemsExpiredUnfetched,
		"outofmemory":       e.itemsOutofmemory,
		"reclaimed":         e.itemsReclaimed,
		"tailrepairs":       e.itemsTailrepairs,
	}

	for _, t := range stats {
		s := t.Stats
		ch <- prometheus.MustNewConstMetric(e.uptime, prometheus.CounterValue, parse(s, "uptime"))
		ch <- prometheus.MustNewConstMetric(e.version, prometheus.GaugeValue, 1, s["version"])

		for _, op := range []string{"get", "delete", "incr", "decr", "cas", "touch"} {
			ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, op+"_hits"), op, "hit")
			ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, op+"_misses"), op, "miss")
		}
		ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, "cas_badval"), "cas", "badval")
		ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, parse(s, "cmd_flush"), "flush", "hit")

		// memcached includes cas operations again in cmd_set.
		set := math.NaN()
		if setCmd, err := strconv.ParseFloat(s["cmd_set"], 64); err == nil {
			if cas, casErr := sum(s, "cas_misses", "cas_hits", "cas_badval"); casErr == nil {
				set = setCmd - cas
			} else {
				log.Errorf("Failed to parse cas: %s", casErr)
			}
		} else {
			log.Errorf("Failed to parse set %q: %s", s["cmd_set"], err)
		}
		ch <- prometheus.MustNewConstMetric(e.commands, prometheus.CounterValue, set, "set", "hit")

		ch <- prometheus.MustNewConstMetric(e.currentBytes, prometheus.GaugeValue, parse(s, "bytes"))
		ch <- prometheus.MustNewConstMetric(e.limitBytes, prometheus.GaugeValue, parse(s, "limit_maxbytes"))
		ch <- prometheus.MustNewConstMetric(e.items, prometheus.GaugeValue, parse(s, "curr_items"))
		ch <- prometheus.MustNewConstMetric(e.itemsTotal, prometheus.CounterValue, parse(s, "total_items"))

		ch <- prometheus.MustNewConstMetric(e.bytesRead, prometheus.CounterValue, parse(s, "bytes_read"))
		ch <- prometheus.MustNewConstMetric(e.bytesWritten, prometheus.CounterValue, parse(s, "bytes_written"))

		ch <- prometheus.MustNewConstMetric(e.currentConnections, prometheus.GaugeValue, parse(s, "curr_connections"))
		ch <- prometheus.MustNewConstMetric(e.connectionsTotal, prometheus.CounterValue, parse(s, "total_connections"))

		ch <- prometheus.MustNewConstMetric(e.evictions, prometheus.CounterValue, parse(s, "evictions"))
		ch <- prometheus.MustNewConstMetric(e.reclaimed, prometheus.CounterValue, parse(s, "reclaimed"))

		ch <- prometheus.MustNewConstMetric(e.malloced, prometheus.GaugeValue, parse(s, "total_malloced"))

		for slab, u := range t.Items {
			slab := strconv.Itoa(slab)
			ch <- prometheus.MustNewConstMetric(e.itemsNumber, prometheus.GaugeValue, parse(u, "number"), slab)
			ch <- prometheus.MustNewConstMetric(e.itemsAge, prometheus.GaugeValue, parse(u, "age"), slab)
			for m, d := range itemsMetrics {
				if _, ok := u[m]; !ok {
					continue
				}
				ch <- prometheus.MustNewConstMetric(d, prometheus.CounterValue, parse(u, m), slab)
			}
		}

		for slab, v := range t.Slabs {
			slab := strconv.Itoa(slab)

			for _, op := range []string{"get", "delete", "incr", "decr", "cas", "touch"} {
				ch <- prometheus.MustNewConstMetric(e.slabsCommands, prometheus.CounterValue, parse(v, op+"_hits"), slab, op, "hit")
			}
			ch <- prometheus.MustNewConstMetric(e.slabsCommands, prometheus.CounterValue, parse(v, "cas_badval"), slab, "cas", "badval")

			slabSet := math.NaN()
			if slabSetCmd, err := strconv.ParseFloat(v["cmd_set"], 64); err == nil {
				if slabCas, slabCasErr := sum(v, "cas_hits", "cas_badval"); slabCasErr == nil {
					slabSet = slabSetCmd - slabCas
				} else {
					log.Errorf("Failed to parse cas: %s", slabCasErr)
				}
			} else {
				log.Errorf("Failed to parse set %q: %s", v["cmd_set"], err)
			}
			ch <- prometheus.MustNewConstMetric(e.slabsCommands, prometheus.CounterValue, slabSet, slab, "set", "hit")

			ch <- prometheus.MustNewConstMetric(e.slabsChunkSize, prometheus.GaugeValue, parse(v, "chunk_size"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsChunksPerPage, prometheus.GaugeValue, parse(v, "chunks_per_page"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsCurrentPages, prometheus.GaugeValue, parse(v, "total_pages"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsCurrentChunks, prometheus.GaugeValue, parse(v, "total_chunks"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsChunksUsed, prometheus.GaugeValue, parse(v, "used_chunks"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsChunksFree, prometheus.GaugeValue, parse(v, "free_chunks"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsChunksFreeEnd, prometheus.GaugeValue, parse(v, "free_chunks_end"), slab)
			ch <- prometheus.MustNewConstMetric(e.slabsMemRequested, prometheus.GaugeValue, parse(v, "mem_requested"), slab)
		}
	}

	statsSettings, err := e.mc.StatsSettings()
	if err != nil {
		log.Errorf("Could not query stats settings: %s", err)
	}
	for _, settings := range statsSettings {
		ch <- prometheus.MustNewConstMetric(e.maxConnections, prometheus.GaugeValue, parse(settings, "maxconns"))
	}
}

func parse(stats map[string]string, key string) float64 {
	v, err := strconv.ParseFloat(stats[key], 64)
	if err != nil {
		log.Errorf("Failed to parse %s %q: %s", key, stats[key], err)
		v = math.NaN()
	}
	return v
}

func sum(stats map[string]string, keys ...string) (float64, error) {
	s := 0.
	for _, key := range keys {
		v, err := strconv.ParseFloat(stats[key], 64)
		if err != nil {
			return math.NaN(), err
		}
		s += v
	}
	return s, nil
}
