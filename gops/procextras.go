package gops

// procRemote is a remote connection for a process
type procRemote struct {
	Protocol string
	Remote   string
	Local    bool
}

type tcpCounters struct {
	Cookie  uint64
	RxBytes uint64
	TxBytes uint64
}

type procNet struct {
	// Known is false when the process's fds or the socket dump were
	// unreadable, or the process is in another network namespace with
	// interfaces besides lo.
	Known bool
	// Count is the exact number of sockets with an internet remote,
	// LocalCount those with a LAN, VPN or link-local one.
	Count      int
	LocalCount int
	Remotes    []procRemote
	// TCP lists the non-loopback TCP sockets.
	TCP []tcpCounters
	// UDP is set when a non-loopback UDP socket is open; UDP has no byte counters.
	UDP bool
	// RatesUnknown is set when a TCP socket came without byte counters, so
	// summing the rest would under-report.
	RatesUnknown bool
}

// dumpSocket is what one socket dump contributes to rates: its time (Unix ms),
// every non-loopback TCP socket with counters, and the final counters of
// sockets destroyed since, when known.
type dumpSocket struct {
	tcpCounters
	// Orphan is a socket whose last fd closed; it lingers until the
	// connection finishes closing.
	Orphan bool
}

// netDump is what one socket dump contributes to rates: its time (Unix ms),
// every non-loopback TCP socket with counters, and the final counters of
// sockets destroyed since, when known.
type netDump struct {
	at    int64
	socks []dumpSocket
	tail  func(cookie uint64) (rx, tx uint64, ok bool)
}
