package model

const (
	AttributeHTTPURL = "http.url" // 1.x
	AttributeURLFULL = "url.full" // 2.x

	AttributeHttpPath = "http.path"

	AttributeHttpMethod        = "http.method"         // 1.x
	AttributeHttpRequestMethod = "http.request.method" // 2.x

	AttributeNetPeerIp      = "net.peer.ip"
	AttributeHTTPStatusCode = "http.status_code"
	AttributeDBStatement    = "db.statement"
	AttributeDBSystem       = "db.system"
	AttributeDBName         = "db.name"
	AttributeDBSQLTable     = "db.sql.table"
	AttributeDBOperation    = "db.operation"

	AttributeRpcSystem  = "rpc.system"
	AttributeRpcService = "rpc.service"
	AttributeRpcMethod  = "rpc.method"

	AttributeNetPeerName   = "net.peer.name"  // 1.x
	AttributeNetPeerPort   = "net.peer.port"  // 1.x
	AttributeServerAddress = "server.address" // 2.x
	AttributeServerPort    = "server.port"    // 2.x

	AttributeNetSockPeerAddr    = "net.sock.peer.addr"   // 1.x
	AttributeNetSockPeerPort    = "net.sock.peer.port"   // 1.x
	AttributeNetworkPeerAddress = "network.peer.address" // 2.x
	AttributeNetworkPeerPort    = "network.peer.port"    // 2.x

	AttributeMessageSystem          = "messaging.system"
	AttributeMessageDestination     = "messaging.destination"
	AttributeMessageDestinationName = "messaging.destination.name"

	AttributeExceptionType       = "exception.type"
	AttributeExceptionMessage    = "exception.message"
	AttributeExceptionStacktrace = "exception.stacktrace"

	AttributeApmSpanType       = "apm.span.type" // SKYWALKING / OTEL / ARMS / TINGYUN3
	AttributeApmOriginalSpanId = "apm.original.span.id"
)
