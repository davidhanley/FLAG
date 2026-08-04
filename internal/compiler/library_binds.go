package compiler

// libraryGoBinds maps host bind keys used in module :go-exports values to the
// generated runtime adapter identifier (flagrt.GoBind_*).
//
// These are NOT ambient language symbols: FLAG code only sees them after
// importing a libraries/*.lib module that re-exports them via :go-exports.
var libraryGoBinds = map[string]string{
	"async/go-run":           "GoBind_async_GoRun",
	"async/future-run":       "GoBind_async_FutureRun",
	"async/future-piped-run": "GoBind_async_FuturePipeRun",
	"async/sleep":            "GoBind_async_Sleep",
	"async/make-channel":     "GoBind_async_MakeChannel",
	"async/channel-send":     "GoBind_async_ChannelSend",
	"async/channel-receive":  "GoBind_async_ChannelReceive",
	"async/select":           "GoBind_async_Select",
	"burp/escape":            "GoBind_burp_Escape",
	"burp/html":              "GoBind_burp_Html",
	"burp/html5":             "GoBind_burp_Html5",
	"burp/raw":               "GoBind_burp_Raw",
	"csv/read-csv":           "GoBind_csv_ReadCSV",
	"csv/read-csv-lines":     "GoBind_csv_ReadCSVLines",
	"csv/read-csv-path":      "GoBind_csv_ReadCSVPath",
	"csv/read-csv-reader":    "GoBind_csv_ReadCSVReader",
}
