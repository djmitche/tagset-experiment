# References

 * `pkg/util/tags_builder.go` - TagBuilder
 * `pkg/aggregator/ckey/key.go` - hashing (remy)
 * `pkg/tagger/global.go` - calls to TagBuilder, cardinality
 * `pkg/tagger/utils/concat.go` - concatenate string arrays (does not dedupe)
 * `pkg/dogstatsd/intern.go` - string interner (but why, it's parsed from []byte in `pkg/dogstatsd/parse.go`)
 * https://github.com/DataDog/datadog-agent/pull/8822/files

# Notes

* Interner is hashing its input []byte and looking that up in a map
  * --> this is optimizing `string(bytes)` to minimize allocations, but the source data is already allocated
* Can't hold onto raw buffers from packets in dogstatsd, as they are reused
* Finalizers can be used to implement weak things but are _really_ tricky
* Will want to avoid tagset hash collisions, but ideally avoid comparing the entire tagset on every comparison.
  Maybe a secondary 64-bit hash, maybe over the serialized output?  Then serialize needs to be stable (sorted)

# Next Steps

* [DONE] I need to make a Serialize() method.
* [DONE-ISH] Tag interning, ideally avoiding locking.  thread-local storage would be super nice here (so, an intern DB per thread), but isn't available in Go.  Tags come with a cardinality level, so it may make sense to heavily intern the low-cardinality tags and lightly intern the high-cardinality tags.
* Build a tag-slice allocator around sync.Pool that can avoid all those make(..) calls in tagset.go
* Build a tagset allocator that re-uses storage for TagSet instances when they are parsed, and recognizes repeated unions of the same tagsets
* For Hash and Serialization, does Go inline the function well enough, or should I change that to a public property and warn people not to change it?
* I eagerly calculate hashes for both tags and tagsets, because it's guaranteed we'll need that.  Is the same true for serialization?  If not, is there a lock-free, threadsafe way to cache a serialization on first use?  sync.Once might do the trick -- it is lock-free on the hot path.
* Is there a better way to store tags with their hashes, without allocating a 16-byte struct separately from the byte slice?  Does that matter?
* Would it help to try to build a slab allocator for tags, allocating (say) 4k of bytes at a time and building []byte slices of that?
* [DONE] Use []byte to avoid ambiguity of copying strings, allow byte buffers ← byte buffers in DSD are reused
* [DONE] Weak ref map? ← no such thing
* Can we assume no hash collisions? ← yes, at 128 bits
* Avoid making all of this threadsafe by defining a "Universe" that contains all of the otherwise-global caches, and restricting a universe to a single goroutine at any one time (sort of like a TagBuilder, but longer-lived)
* Use 2-choice hashing with the H and L hashes, and give up and don't cache when both slots are full (very unlikely failsafe)

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
