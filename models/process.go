package models

type ProcessInfo struct {
	PID               int32   `json:"pid"`
	PPID              int32   `json:"ppid"`
	CPU               float64 `json:"cpu" doc:"Percentage of total machine CPU capacity used since the cursor, in the range 0-100."`
	PTicks            float64 `json:"pticks" doc:"Cumulative CPU seconds (user + system) consumed by this process."`
	MemoryPercent     float32 `json:"memoryPercent"`
	MemoryKB          uint64  `json:"memoryKB"`
	MemoryCalculation string  `json:"memoryCalculation"`
	RSSKB             uint64  `json:"rssKB"`
	RSSPercent        float32 `json:"rssPercent"`
	PSSKB             uint64  `json:"pssKB"`
	PSSPercent        float32 `json:"pssPercent"`
	Username          string  `json:"username"`
	Command           string  `json:"command"`
	FullCommand       string  `json:"fullCommand"`
	ExecutablePath    string  `json:"executablePath,omitempty"`
	ChildCount        int     `json:"childCount,omitempty" doc:"Processes merged into this row. A merged row sums its members; a value unknown for any member is absent for the row."`
	Threads           int     `json:"threads,omitempty" doc:"Thread count; merged rows sum their members."`
	Nice              *int    `json:"nice,omitempty"`

	StartTime int64    `json:"startTime,omitempty" doc:"Process start, unix seconds."`
	DiskRead  *float64 `json:"diskRead,omitempty" doc:"Disk bytes read per second since the cursor."`
	DiskWrite *float64 `json:"diskWrite,omitempty" doc:"Disk bytes written per second since the cursor."`

	Connections      *int            `json:"connections,omitempty" doc:"Open sockets with an internet remote; merged rows sum their members."`
	LocalConnections *int            `json:"localConnections,omitempty" doc:"Open sockets with a LAN, VPN or link-local remote (private, CGNAT or link-local address); loopback is never counted. Merged rows sum their members."`
	Remotes          []NetConnection `json:"remotes,omitempty" doc:"Internet and local remotes, capped at 32."`
	NetRx            *float64        `json:"netRx,omitempty" doc:"TCP payload bytes per second received from internet and local remotes since the cursor. Absent when unknown: unreadable fds, a container's own network namespace, or only UDP sockets, which have no byte counters."`
	NetTx            *float64        `json:"netTx,omitempty" doc:"TCP payload bytes sent and acknowledged, like netRx."`
	NetPartial       bool            `json:"netPartial,omitempty" doc:"True when UDP sockets (e.g. QUIC) to internet or local remotes are open too; netRx and netTx leave them out, so they are a lower bound."`

	Children []*ProcessInfo `json:"children,omitempty" doc:"With members requested, the processes of a merged row, root first, each with its own values."`
}

type NetConnection struct {
	Protocol string `json:"protocol" doc:"tcp or udp."`
	Remote   string `json:"remote" doc:"Remote address as ip:port, IPv6 bracketed."`
	Local    bool   `json:"local,omitempty" doc:"True for a LAN, VPN or link-local remote rather than an internet one."`
}

type ProcessCursorData struct {
	PID       int32   `json:"pid"`
	Ticks     float64 `json:"ticks"`
	Timestamp int64   `json:"timestamp"`
}

type ProcessListResponse struct {
	Processes []*ProcessInfo `json:"processes"`
	Cursor    string         `json:"cursor,omitempty"`
}

// ListenSocket is a socket a process listens on.
type ListenSocket struct {
	Protocol string `json:"protocol" doc:"tcp or udp."`
	Address  string `json:"address" doc:"Bind address without brackets; 0.0.0.0 or :: for any."`
	Port     int    `json:"port" doc:"Bound port."`
}

// ProcessDetail holds per-process data that is too costly to poll for every row; it is read on demand for one PID.
type ProcessDetail struct {
	PID        int32          `json:"pid"`
	ReadTotal  *uint64        `json:"readTotal,omitempty" doc:"Bytes read from storage over the process lifetime."`
	WriteTotal *uint64        `json:"writeTotal,omitempty" doc:"Bytes written to storage over the process lifetime."`
	OpenFiles  *int           `json:"openFiles,omitempty" doc:"Open file descriptors."`
	Terminal   bool           `json:"terminal,omitempty" doc:"True when the process has a controlling terminal, so it cannot be relaunched detached."`
	Cwd        string         `json:"cwd,omitempty" doc:"Working directory."`
	Listening  []ListenSocket `json:"listening,omitempty" doc:"TCP sockets in LISTEN state and unconnected UDP sockets the process owns."`

	UserTime            *float64 `json:"userTime,omitempty" doc:"CPU time spent in user mode, in seconds."`
	SystemTime          *float64 `json:"systemTime,omitempty" doc:"CPU time spent in kernel mode, in seconds."`
	VoluntarySwitches   *uint64  `json:"voluntarySwitches,omitempty" doc:"Context switches where the process gave up the CPU."`
	InvoluntarySwitches *uint64  `json:"involuntarySwitches,omitempty" doc:"Context switches forced by the scheduler."`

	RSSKB       *uint64 `json:"rssKB,omitempty" doc:"Resident memory, in KB."`
	PeakKB      *uint64 `json:"peakKB,omitempty" doc:"Peak resident memory, in KB."`
	VirtualKB   *uint64 `json:"virtualKB,omitempty" doc:"Virtual memory size, in KB."`
	MinorFaults *uint64 `json:"minorFaults,omitempty" doc:"Page faults served without disk I/O."`
	MajorFaults *uint64 `json:"majorFaults,omitempty" doc:"Page faults that needed disk I/O."`

	UID     *int   `json:"uid,omitempty" doc:"Real user ID."`
	GID     *int   `json:"gid,omitempty" doc:"Real group ID."`
	Group   string `json:"group,omitempty" doc:"Name of the real group."`
	TTY     string `json:"tty,omitempty" doc:"Controlling terminal, e.g. ttys000."`
	Session *int   `json:"session,omitempty" doc:"Session ID."`

	LaunchdLabel string `json:"launchdLabel,omitempty" doc:"launchd label of the process, e.g. com.apple.Safari, when launched via launchd."`
	BundleID     string `json:"bundleId,omitempty" doc:"Bundle identifier of the owning app, e.g. com.apple.Safari, when known."`
}
