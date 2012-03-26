# elasticsearch

I had a need to index a lot of documents in elasticsearch recently.
The bulk interface, in particular, was a bit difficult to get right
right away.

Now I can get pretty good update rates by batching lots of index
updates and submitting them in bulk using the `Bulk()` API.

I don't have any query interfaces in here because there are many ways
to query elasticsearch and many ways to handle results.  In most
cases, they're all pretty simple things to do with `net/http` and
`encoding/json` and I haven't figured out a way to make them easier
yet.

