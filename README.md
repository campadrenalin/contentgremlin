contentgremlin
==============

While I definitely don't mean any offense to Chris Webber or the rest of the Mediagoblin team, my experience has been that GNU MediaGoblin is painful to admin. It doesn't handle failure well, some of its plugins conflict badly when they ought to get along, and things just tend to go mysteriously wrong; the upgrade story isn't great, either, and the instructions for "production" deployments tend to make debugging harder.

I have no doubt that the situation is going to get better, but I'm not patient, either - and even if a new GMG comes out that makes all my dreams come true, I still need to migrate my existing instance to it, which makes me sad. Plus, if I'm going to administrate something, I'd like to understand it, or at least know that that there is thorough documentation and robust architecture.

Besides... a little friendly competition is not a bad thing!

## Self-contained

Contentgremlin is going to be designed to fit in a directory and depend on only the barest stuff outside that directory - kernel, libc, etc. This is partly the benefits of Go's static linking, and partly philosophy. Wanna know where your content and database are? No surprises, no hunting. There's a db.sqlite file and a content/ directory.

At some point, *maybe*, contentgremlin will support other databases in a pluggable way. We certainly aren't burning any bridges with the architecture or SQL queries. On the other hand, SQLite is a very fast, featureful, and solid technology, and it would be surprising to see contentgremlin used beyond SQLite's capacity in any deployment. And, as all software engineers eventually learn the hard way... [YAGNI](http://en.wikipedia.org/wiki/You_aren%27t_gonna_need_it).

## Sexy upgrade process

The upgrade process is entirely automated, because the database is self-describing in its version system, and supports locking the existing process into read-only mode.

 * When you first start the server running, it stores its PID in the database. The database *is* the pidfile.
 * When you start the new version in "upgrade mode", it creates a new empty file `db_new.sqlite` alongside the existing `db.sqlite`
 * The new process flips a switch in the old DB, which puts the existing process into "maintenance mode," and does not continue until it has assurance from the old process that it may proceed untrampled.
 * The new process copies over all the data, performing any schema transformations as necessary.
 * Once the copy is complete, it sends a signal to the old process, to exit.
 * As soon as the old process has exited, the new process moves `db.sqlite` to `db_$version.sqlite` and `db_new.sqlite` to `db.sqlite`.
 * As soon as the new `db_sqlite` is in place, we can acquire the socket and start serving!

This means we have a very brief maintenance window *during which time, the old process properly operates in read-only mode*, then an even briefer outage, and then the new version is up and running. All in about the amount of time it takes to blink, thanks to kernel-level disk caching, and the fact that there is no human intervention except to start the upgrade process.

It also does not throw away your old data, so you can safely revert if things go wrong.
