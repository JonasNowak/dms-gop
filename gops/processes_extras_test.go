package gops

import (
	"cmp"
	"slices"
	"testing"

	"github.com/JonasNowak/dms-gop/models"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

func TestMergeProcesses_IOAndNet(t *testing.T) {
	merge := func(rows ...*models.ProcessInfo) map[int32]*models.ProcessInfo {
		out := map[int32]*models.ProcessInfo{}
		for _, p := range mergeProcessesByExecutable(rows, false) {
			out[p.PID] = p
		}
		return out
	}

	t.Run("IO sums and unknown members make a sum unknown", func(t *testing.T) {
		summed := merge(
			&models.ProcessInfo{PID: 1, ExecutablePath: "/bin/a", DiskRead: ptr(5.0), DiskWrite: ptr(10.0), StartTime: 111},
			&models.ProcessInfo{PID: 2, PPID: 1, ExecutablePath: "/bin/a", DiskRead: ptr(2.0), DiskWrite: ptr(3.0), StartTime: 222},
		)[1]
		assert.Equal(t, 7.0, *summed.DiskRead)
		assert.Equal(t, 13.0, *summed.DiskWrite)
		assert.Equal(t, int64(111), summed.StartTime)

		unknown := merge(
			&models.ProcessInfo{PID: 1, ExecutablePath: "/bin/a", DiskRead: ptr(5.0), StartTime: 111},
			&models.ProcessInfo{PID: 2, PPID: 1, ExecutablePath: "/bin/a", DiskWrite: ptr(10.0), StartTime: 222},
		)[1]
		assert.Nil(t, unknown.DiskRead, "known parent plus unknown child is unknown")
		assert.Nil(t, unknown.DiskWrite, "unknown parent plus known child is unknown")
		assert.Equal(t, int64(111), unknown.StartTime)
	})

	t.Run("network sums, caps remotes and stays exact", func(t *testing.T) {
		remotes := func(n int, host string) []models.NetConnection {
			out := make([]models.NetConnection, n, n+10)
			for i := range out {
				out[i] = models.NetConnection{Protocol: "tcp", Remote: host}
			}
			return out
		}
		parent := &models.ProcessInfo{PID: 1, ExecutablePath: "/bin/a", Connections: ptr(30), LocalConnections: ptr(1), Remotes: remotes(30, "1.1.1.1:443"), NetRx: ptr(100.0), NetTx: ptr(10.0)}
		child := &models.ProcessInfo{PID: 2, PPID: 1, ExecutablePath: "/bin/a", Connections: ptr(5), LocalConnections: ptr(2), Remotes: remotes(5, "8.8.8.8:443"), NetRx: ptr(50.0), NetPartial: true}
		unknown := &models.ProcessInfo{PID: 3, PPID: 1, ExecutablePath: "/bin/a"}

		root := merge(parent, child, unknown)[1]
		assert.Nil(t, root.Connections, "a member with unknown sockets makes the count unknown")
		assert.Len(t, root.Remotes, maxProcRemotes)
		assert.True(t, root.NetPartial, "partial if any member is")

		known := merge(parent, child)[1]
		assert.Equal(t, 35, *known.Connections)
		assert.Equal(t, 3, *known.LocalConnections)
		assert.Equal(t, 150.0, *known.NetRx)
		assert.Nil(t, known.NetTx, "the child's unknown tx makes the sum unknown")
		assert.Equal(t, 30, *parent.Connections, "the original rows must not be written through")
		assert.Equal(t, models.NetConnection{}, parent.Remotes[:31][30], "the parent's spare capacity must not be appended into")
	})

	t.Run("members list the root first with its own values", func(t *testing.T) {
		child := &models.ProcessInfo{PID: 2, PPID: 1, ExecutablePath: "/bin/a", CPU: 3}
		parent := &models.ProcessInfo{PID: 1, ExecutablePath: "/bin/a", CPU: 1}
		lone := &models.ProcessInfo{PID: 9, ExecutablePath: "/bin/b"}
		out := map[int32]*models.ProcessInfo{}
		for _, p := range mergeProcessesByExecutable([]*models.ProcessInfo{child, parent, lone}, true) {
			out[p.PID] = p
		}
		assert.Equal(t, []*models.ProcessInfo{parent, child}, out[1].Children)
		assert.Equal(t, 4.0, out[1].CPU)
		assert.Equal(t, 1.0, parent.CPU)
		assert.Nil(t, out[9].Children, "an unmerged row has no members")
		assert.Nil(t, merge(parent, child)[1].Children, "only listed when asked")
	})
}

func TestSortProcesses_IOAndNetPutUnknownLast(t *testing.T) {
	pids := func(list []*models.ProcessInfo) []int32 {
		out := make([]int32, len(list))
		for i, p := range list {
			out[i] = p.PID
		}
		return out
	}
	rows := func() []*models.ProcessInfo {
		return []*models.ProcessInfo{
			{PID: 1},
			{PID: 2, DiskRead: ptr(1.0), DiskWrite: ptr(1.0)},
			{PID: 3},
			{PID: 4, DiskRead: ptr(0.0), DiskWrite: ptr(10.0)},
		}
	}
	t.Run("disk sort", func(t *testing.T) {
		list := rows()
		sortProcesses(list, SortByDisk)
		assert.Equal(t, []int32{4, 2}, pids(list)[:2])
		assert.ElementsMatch(t, []int32{1, 3}, pids(list)[2:])
	})

	list := []*models.ProcessInfo{
		{PID: 1},
		{PID: 2, Connections: ptr(2), NetRx: ptr(1.0), NetTx: ptr(0.0)},
		{PID: 3, Connections: ptr(3)},
		{PID: 4, Connections: ptr(2), NetRx: ptr(5.0), NetTx: ptr(0.0)},
		{PID: 5, Connections: ptr(0), NetRx: ptr(0.0), NetTx: ptr(0.0)},
	}
	sortProcesses(list, SortByNetwork)
	assert.Equal(t, []int32{3, 4, 2, 5, 1}, pids(list))
}

func netRows(nets map[int32]procNet, prevOK ...int32) []netRow {
	rows := make([]netRow, 0, len(nets))
	for pid, n := range nets {
		n.Known = true
		rows = append(rows, netRow{pid: pid, info: &models.ProcessInfo{PID: pid}, net: n, prevOK: slices.Contains(prevOK, pid)})
	}
	slices.SortFunc(rows, func(a, b netRow) int { return cmp.Compare(a.pid, b.pid) })
	return rows
}

func tcp(cookie, rx, tx uint64) tcpCounters { return tcpCounters{cookie, rx, tx} }

func dumpOf(at int64, socks ...tcpCounters) *netDump {
	d := &netDump{at: at}
	for _, s := range socks {
		d.socks = append(d.socks, dumpSocket{tcpCounters: s})
	}
	return d
}

func TestApplyNetRates(t *testing.T) {
	const now = int64(1_755_000_010_000)
	const mb = 1_000_000
	base := func(socks map[uint64]sockBaseline) netBaseline { return netBaseline{at: now - 1_000, socks: socks} }
	shared := []tcpCounters{tcp(1, 100, 100)}
	moved := []tcpCounters{tcp(1, 150, 120), tcp(2, 600, 500), tcp(3, 1000, 501)}
	kept := []tcpCounters{tcp(1, 11*mb, 5*mb), tcp(3, 1, 1), tcp(4, mb/2, mb/2+1)} // 4 is new: bytes_acked counts the SYN
	keptPrev := map[uint64]sockBaseline{1: {10 * mb, 5 * mb, 7}, 2: {7 * mb, 7 * mb, 7}, 3: {4 * mb, 4 * mb, 7}}

	tests := []struct {
		name   string
		prev   netBaseline
		cur    *netDump
		nets   map[int32]procNet
		prevOK []int32
		wantRx map[int32]*float64 // nil value = unknown
		wantTx map[int32]*float64
	}{
		{"kept, closed, backwards and new sockets", base(keptPrev), dumpOf(now, kept...), map[int32]procNet{7: {TCP: kept}}, []int32{7},
			map[int32]*float64{7: ptr(1.5 * mb)}, map[int32]*float64{7: ptr(0.5 * mb)}},
		{"all sockets closed is 0", base(keptPrev), dumpOf(now), map[int32]procNet{7: {}}, []int32{7},
			map[int32]*float64{7: ptr(0.0)}, nil},
		{"no baseline", netBaseline{}, dumpOf(now, kept...), map[int32]procNet{7: {TCP: kept}}, []int32{7},
			map[int32]*float64{7: nil}, nil},
		{"moved socket is a delta, a new holder still has a baseline per socket",
			base(map[uint64]sockBaseline{1: {100, 100, 10}, 2: {500, 500, 0}}), dumpOf(now, moved...),
			map[int32]procNet{10: {}, 20: {TCP: moved[:1]}, 30: {TCP: moved[1:]}}, []int32{10},
			map[int32]*float64{10: ptr(0.0), 20: ptr(50.0), 30: ptr(1100.0)}, map[int32]*float64{20: ptr(20.0), 30: ptr(500.0)}},
		{"a reused inode is told apart by cookie", base(map[uint64]sockBaseline{5: {1 << 30, 1 << 30, 7}}), dumpOf(now, tcp(6, 10, 11)),
			map[int32]procNet{7: {TCP: []tcpCounters{tcp(6, 10, 11)}}}, []int32{7},
			map[int32]*float64{7: ptr(10.0)}, map[int32]*float64{7: ptr(10.0)}},
		{"a shared socket counts once", base(map[uint64]sockBaseline{1: {0, 0, 4}}), dumpOf(now, shared...),
			map[int32]procNet{3: {TCP: shared}, 4: {TCP: shared}}, []int32{3, 4},
			map[int32]*float64{3: ptr(100.0), 4: ptr(0.0)}, nil},
		{"UDP only is unknown", base(map[uint64]sockBaseline{1: {0, 0, 2}}), dumpOf(now, tcp(1, 40, 0)),
			map[int32]procNet{1: {UDP: true}, 2: {UDP: true, TCP: []tcpCounters{tcp(1, 40, 0)}}, 3: {RatesUnknown: true}}, []int32{1, 2, 3},
			map[int32]*float64{1: nil, 2: ptr(40.0), 3: nil}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := netRows(tt.nets, tt.prevOK...)
			applyNetRates(tt.prev, tt.cur, rows)
			for _, r := range rows {
				if want, ok := tt.wantRx[r.pid]; ok {
					if want == nil {
						assert.Nil(t, r.info.NetRx, "rx %d", r.pid)
					} else if assert.NotNil(t, r.info.NetRx, "rx %d", r.pid) {
						assert.InDelta(t, *want, *r.info.NetRx, 1e-6, "rx %d", r.pid)
					}
				}
				if want, ok := tt.wantTx[r.pid]; ok && assert.NotNil(t, r.info.NetTx, "tx %d", r.pid) {
					assert.InDelta(t, *want, *r.info.NetTx, 1e-6, "tx %d", r.pid)
				}
			}
		})
	}
}

func TestApplyNetRates_ClosedSocketTails(t *testing.T) {
	const now = int64(1_755_000_010_000)
	prev := netBaseline{at: now - 1_000, socks: map[uint64]sockBaseline{
		1: {100, 100, 7}, // destroyed, final counters known
		2: {100, 100, 8}, // destroyed, held by a PID that is now someone else
		3: {100, 100, 7}, // orphaned: no fd, still closing
		4: {100, 100, 0}, // destroyed, never attributed
		5: {100, 100, 7}, // destroyed, no final counters
	}}
	cur := dumpOf(now)
	cur.socks = append(cur.socks, dumpSocket{tcp(3, 130, 110), true})
	cur.tail = func(cookie uint64) (uint64, uint64, bool) {
		if cookie == 5 {
			return 0, 0, false
		}
		return 300, 250, true
	}
	rows := netRows(map[int32]procNet{7: {}, 8: {}}, 7)
	next := applyNetRates(prev, cur, rows)
	assert.Equal(t, []float64{230, 160}, []float64{*rows[0].info.NetRx, *rows[0].info.NetTx}, "tail of 1 plus the orphan's delta")
	assert.Equal(t, 0.0, *rows[1].info.NetRx, "a reused PID gets nothing")
	assert.Equal(t, map[uint64]sockBaseline{3: {130, 110, 7}}, next.socks, "an orphan keeps its holder")
}
