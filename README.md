# References

 * `pkg/util/tags_builder.go` - TagBuilder
 * `pkg/aggregator/ckey/key.go` - hashing (remy)
 * `pkg/tagger/global.go` - calls to TagBuilder, cardinality
 * `pkg/tagger/utils/concat.go` - concatenate string arrays (does not dedupe)
 * `pkg/dogstatsd/intern.go` - string interner (but why, it's parsed from []byte in `pkg/dogstatsd/parse.go`)

# Notes

* Interner is hashing its input []byte and looking that up in a map
  * --> this is optimizing `string(bytes)` to minimize allocations, but the source data is already allocated
* Can't hold onto raw buffers from packets in dogstatsd, as they are reused
* Finalizers can be used to implement weak things but are _really_ tricky

# Ideas

* Use []byte to avoid ambiguity of copying strings, allow byte buffers ← byte buffers are reused
* Weak ref map? ← no such thing
* Can we assume no hash collisions?

Setting language aside, given:

```
// some set of low-cardinality tags (possibly with hints as to cardinality
// to decide whether to do things lazily or eagerly)
hostTags = TagSet("host:...", "os:linux", ..)

// tags read from a raw buffer
sampleTags = parseCommaSepratedTags(buf)

// tags potentially cached by the tagger
taggerTags = tagger.getTagsFor(something)

// combine those (possibly with some hints as to disjointness)
tags = union(hostTags, sampleTags, taggerTags)

// write to aggregator
aggregate(name, tags, host)

func aggregate(name string, tags *TagSet, host string) {
    contextHash = hash(name, tags.hash(), host)

    ..

    sendToIntake(.., tags.serialize(), ..)
}
```

And, assuming that every TagSet is eventually both hashed and serialized.

We would like to do all of this with

* minimal copies
* minimal allocations (to reduce GC churn)
* minimal scans of the tag strings
* good cache behavior
