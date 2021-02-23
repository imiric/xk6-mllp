package mllp

import "github.com/loadimpact/k6/stats"

var (
	WriterWrites = stats.New("mllp.writer.write.count", stats.Counter)
)
