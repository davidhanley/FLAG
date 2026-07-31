package runtime

import csvlib "flag-lang/libraries/csv"

func init() {
	RegisterGoFunction("csv/read-csv", csvlib.ReadCSV)
	RegisterGoFunction("csv/read-csv-path", csvlib.ReadCSVPath)
	RegisterGoFunction("csv/read-csv-reader", csvlib.ReadCSVReader)
	RegisterGoFunction("csv/read-csv-lines", csvlib.ReadCSVLines)
}
