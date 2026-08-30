
channel-some? / channel-every? drain the sender, huge efficiency loss.

Records
- Make keyword lookup fast (no reflection). Today (:field rec) walks struct fields and `flag` tags on every get.
- 
- Typed defrecord constructors must reject wrong types (e.g. `(->food 1 2 "")` for `^int` should error, not store 0). 