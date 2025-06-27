package stat

type Stats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	HitRate   float64
	SizeBytes int64
}
