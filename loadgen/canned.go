package loadgen

// DSDTagLineGenerator creates a tag-line generator that generates
// output similar to what a typical DSD client might send.
//
// As that is known right now, that means about 1000 different tag sets
// repeated very frequently.
func DSDTagLineGenerator() TagLineGenerator {
	// almost every line has an environment
	envs := NewLowCardinalityTagGenerator([][]byte{
		[]byte("env:staging"),
		[]byte("env:prod"),
		[]byte("env:dev"),
		[]byte("env:qa"),
		[]byte("env:lab"),
		[]byte("env:bench"),
		[]byte("env:rainforest"),
	})

	// some random tags that might or might not appear on a tag line
	tags := NewLowCardinalityTagGenerator([][]byte{
		[]byte("library:snorkels"),
		[]byte("library:bunny"),
		[]byte("library:bananas"),
		[]byte("library:pamplemousse"),
		[]byte("abtest:123"),
		[]byte("abtest:992"),
		[]byte("abtest:16"),
		[]byte("shard:x"),
		[]byte("shard:y"),
		[]byte("shard:z"),
	})

	var lines TagLineGenerator = NewRandomTagLineGenerator([]RandomTagLineOptions{{
		tg:      envs,
		prob:    99,
		repeats: 1,
	}, {
		tg:      tags,
		prob:    25,
		repeats: 5,
	}, {
		tg:      NewHighCardinalityTagGenerator("device", 100),
		prob:    90,
		repeats: 2,
	}})

	// duplicate those, with an active set of 100 tag lines each
	// repeated up to 1000 times
	return NewDuplicateTagLineGenerator(lines, 1000, 1000)
}
