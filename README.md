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

* Use []byte to avoid ambiguity of copying strings, allow byte buffers ← byte buffers in DSD are reused
* Weak ref map? ← no such thing
* Can we assume no hash collisions? ← yes, it seems so.  For 10000 contexts the prob is ~1e-10 (https://preshing.com/20110504/hash-collision-probabilities/)

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

# If I was Doing This In Rust

```rust
impl Tag {
  fn new(t: &str, card: Cardinality) -> Arc<Tag> {
    let pool = Tag::pool_for_cardinality(card); // maybe from TLS?
    pool.get_or_create(t)
  }
}

impl Pool {
  fn new(..some parameters different per cardinality..) ..
  fn get_or_create(&mut self, t: &str) -> Arc<Tag> {
    let h = hash(t);
    // look for h in the existing refs, return if found
    // otherwise create a new one, keep it, return it
  }
}

// need this whole thing to have interior mutability (need to update hash and
// need to transmute repr)
pub struct TagSet {
    hash: Option<u64>,
    repr: TSRepr,
}

enum TSRepr {
    // A list of tags, possibly sorted
    Tags {
        tags: Vec<Arc<Tag>>,
        sorted: bool,
    }

    // A union of child TagSets
    Union {
        child1: Arc<TagSet>,
        child2: Arc<TagSet>,
        disjoint: bool,
    }
}

impl TagSet {
    fn hash(&self) -> Hash {
        if let Some(h) = self.hash {
            return h;
        }

        match self.repr {
            Tags { tags, .. } => {
                // iterate over tags and XOR hash together, store it
            }
            Union { child1, child2, disjoint } => {
            }
        }
    }
}
```
