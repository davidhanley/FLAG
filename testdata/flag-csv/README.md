# flag-csv

Offline copy of [github.com/davidhanley/CSV](https://github.com/davidhanley/CSV)
for FLAG CI. Fetched by MUSE as `csv.lib`.

    {:libraries ["csv"]}

    {:namespace "app"
     :imports   ["csv.lib"]}

    (csv/read-csv-path "data.csv")

The FLAG surface is libraries/csv.lib. Host adapters live in flagbind/
(copied into generated FLAG builds). The pure Go reader is
github.com/davidhanley/CSV.
