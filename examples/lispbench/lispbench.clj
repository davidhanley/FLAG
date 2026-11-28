(ns lispbench)

;; Keep these bodies aligned with examples/lispbench/work.flag.

(defn athlete [i]
  {:name (str "A" (mod i 50))
   :age  (+ 20 (mod i 40))
   :sex  (if (even? i) :male :female)
   :points i})

(defn athletes [n]
  (map athlete (range n)))

(defn eager-range-sum [n]
  (reduce + 0 (range n)))

(defn lazy-range-sum [n]
  (reduce + 0 (take n (range))))

(defn lazy-map-inc [n]
  (reduce + 0 (take n (map inc (range)))))

(defn ratio-scores [n]
  (let [fractions (map (fn [x] (/ 5 (+ 5 x))) (range))
        scores (map (fn [f] (* f 100)) fractions)]
    (reduce + 0 (take n scores))))

(defn keyword-gets [n]
  (let [m {:name "ADA" :age 40 :sex :female}]
    (loop [i 0 acc 0]
      (if (>= i n)
        acc
        (recur (inc i) (+ acc (if (:name m) 1 0)))))))

(defn nested-keyword [n]
  (let [m {:header {:race-name "BOSTON" :race-points 100}}]
    (loop [i 0 acc 0]
      (if (>= i n)
        acc
        (recur (inc i)
               (+ acc (if (:race-name (:header m)) 1 0)))))))

(defn assoc-scores [n]
  (let [header {:race-name "X" :race-points 100}
        xs (athletes n)]
    (count (map (fn [a i]
                  (assoc a :header header :points-scored i :overall-rank (inc i)))
                xs
                (range n)))))

(defn group-by-name [n]
  (count (group-by :name (athletes n))))

(defn mapcat-filter [n]
  (let [xs (athletes n)]
    (count (mapcat (fn [sex]
                     (filter (fn [a] (= (:sex a) sex)) xs))
                   [:male :female]))))

(defn sort-by-points [n]
  (count (sort-by :points (athletes n))))

(defn juxt-keys [n]
  (count (group-by (juxt :name :age) (athletes n))))

(defn string-pipeline [n]
  (count (map (fn [s] (clojure.string/upper-case (clojure.string/trim s)))
              (repeat n "  ada  "))))

(defn apply-max [n]
  (apply max (range n)))

(defn frs-score-shape [n]
  (let [fractions (map (fn [x] (/ 5 (+ 5 x))) (range))
        scores (map (fn [f] (* f 100)) fractions)
        xs (athletes n)
        scored (map (fn [a s r]
                      (assoc a :points-scored s :overall-rank (inc r)))
                    xs
                    scores
                    (range))]
    (count (group-by :name scored))))

(defn run-bench [name f]
  (f)
  (let [t0 (System/nanoTime)
        checksum (f)
        t1 (System/nanoTime)]
    (println (str "BENCH " name " " (- t1 t0) " " checksum))))

(defn -main [& _]
  (run-bench "eager-range-sum" (fn [] (eager-range-sum 20000)))
  (run-bench "lazy-range-sum" (fn [] (lazy-range-sum 4000)))
  (run-bench "lazy-map-inc" (fn [] (lazy-map-inc 4000)))
  (run-bench "ratio-scores" (fn [] (ratio-scores 1500)))
  (run-bench "keyword-gets" (fn [] (keyword-gets 200000)))
  (run-bench "nested-keyword" (fn [] (nested-keyword 200000)))
  (run-bench "assoc-scores" (fn [] (assoc-scores 4000)))
  (run-bench "group-by-name" (fn [] (group-by-name 4000)))
  (run-bench "mapcat-filter" (fn [] (mapcat-filter 4000)))
  (run-bench "sort-by-points" (fn [] (sort-by-points 4000)))
  (run-bench "juxt-keys" (fn [] (juxt-keys 4000)))
  (run-bench "string-pipeline" (fn [] (string-pipeline 20000)))
  (run-bench "apply-max" (fn [] (apply-max 20000)))
  (run-bench "frs-score-shape" (fn [] (frs-score-shape 1500))))

(-main)
