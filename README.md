contentgremlin
==============

While I definitely don't mean any offense to Chris Webber or the rest of the Mediagoblin team, my experience has been that GNU MediaGoblin is painful to admin. It doesn't handle failure well, some of its plugins conflict badly when they ought to get along, and things just tend to go mysteriously wrong; the upgrade story isn't great, either, and the instructions for "production" deployments tend to make debugging harder.

I have no doubt that the situation is going to get better, but I'm not patient, either - and even if a new GMG comes out that makes all my dreams come true, I still need to migrate my existing instance to it, which makes me sad. Plus, if I'm going to administrate something, I'd like to understand it, or at least know that that there is thorough documentation and robust architecture.

Besides... a little friendly competition is not a bad thing!

## Self-contained

Contentgremlin is going to be designed to fit in a directory and depend on only the barest stuff outside that directory - kernel, libc, ffmpeg, etc. This is partly the benefits of Go's static linking, and partly philosophy. Wanna know where your content and database are? No surprises, no hunting. There's a db.sqlite file and a content/ directory.

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

## Reliable transcoding, or as close as it gets

Media usually needs various forms of post-upload prep - transcoding into web-friendly formats, and so on. Different media plugins will require different transcoding steps and software, but there's always a risk that it will fail for one reason or another.

Mediagoblin has been notoriously bad at dropping or breaking uploads at this stage, but Contentgremlin puts a big effort toward doing this better.

### Detection when things go wrong

We "watchdog" the external processes that are generating our output files - or rather, we periodically poll the output files, to ensure that the output time is increasing, as is the file size. If an output file goes too long without changes, we can assume that the transcoder process is hung, and should be terminated, and the transcoding process re-attempted.

### Recovery

If the transcoder fails - indicated by exit code, or by Contentgremlin having to shut down the transcoder with extreme prejudice - we delete the output file and try again, up to a configurable number of retries.

Whenever we start a job, we record in the database the input location, output location, attempt number, PID, and content id that the task is blocking. Once there are no more blocking tasks, the media becomes available on the site. This means that the recovery process can actually *span upgrades*, without interupting the transcoding processes at all!

Finally, what if Contentgremlin itself somehow fails horribly? Unlikely, thanks to the mechanisms Go gives us to write solid software, but we should design for it anyways. If Contentgremlin dies, and you restart it, the new process will be able to pick up where the old one left off, just like in an upgrade. Long-running transcoders will be unaffected, finished ones will be reaped with Wait(), and any jobs that *should* be running (but aren't) will be restarted.

In fact, this mechanism is so solid, it can survive hard reboots, where none of the original transcoders survive, nor does the orchestrating Contentgremlin process. If, in the worst case, your machine crashes, you will be able to recover everything that had fully uploaded. Unfortunately, there's nothing we can do for partial uploads.

### Garbage file collection

As a parallel part of the process, Contentgremlin cleans up any half-uploads it finds, and verifies that each finished file has a matching checksum, according to the checksums in the database.

It also moves raw uploads after transcoding finishes.

## Authentication

There are parts of Contentgremlin that will be fairly dogmatic, such as SQLite usage, but authentication is *designed* to be cleanly modular, flexible, and easily comprehensible in any scenario.

### User == uid

The core property of any user is the uid. Names and email addresses come and go, but your consistent identity persists through it all. This is what your media, and your comments, attach to.

The users table does *not* include nicknames, emails, or passwords. The default display string, is the one most recently 'touched'.

### Display names have their own table

We have a separate table for nicknames and email addresses. It has a foreign key to the users table, and a timestamp. It does not contain any secret info. It is uniquely keyed on display string.

### Passwords and other secret info have a third table

These foreign-key onto the display names table. They have a plugin-name field, and a content field. Some auth plugins will want to encrypt (or preferably, hash) the content data, but others won't, so we can't enact a single policy across this table.

### Why the separation?

This seems like more complexity, not less. But it handles every auth scenario well, that I can think to throw at it, without becoming a maze of corner cases (or otherwise incomprehensible). As long as you kinda understand what's going on under the hood, which you can inspect with any sqlite prompt, it should be obvious what the behavioral result will be.

Also keep in mind, in these scenarios, how SQLite might protect us from crazy invariants via the foreign key (and unique key) relationships.

#### Just a username

An account may have a username, but no passwords, if it is a system account, or the user intended to deactivate their account without deleting it for archive purposes.

#### Username + Password

Obviously, you can attach a password to a username, and log in with that. This will probably be the common case for anyone not using Persona or OAuth.

#### Multiple usernames, each with a different password

Less of a legit case, but could be useful for doing some stuff in 'mod context', and other stuff as yourself, without actually creating two accounts.

Note that you could only log in via Username A with Password A, and Username B with Password B, etc., assuming each username had exactly one password associated with it. You could not log in with Username A and Password B, for example.

#### Single acount, multiple passwords

Maybe someday we'll be big enough that big corporations will use Contentgremlin. IT COULD HAPPEN, DON'T JUDGE. If it does, we'll have supported a feature from Day 1 to allow multiple people access to the same account, without knowing each others' passwords.

This is also another approach you could take to site-wide moderation. Look at all this flexibility! Like a gymnast, she is.

#### Username + Email + Persona

You don't need an extra password record for Persona, because it works via email address and third-party cryptographic certs (which the client can re-supply on demand). You just need an email address, which is a type of display name (the Persona plugin detects that it is an email based on regex).

On the other hand, why present yourself as your email address? Custom handles are 'da bomb', as the kids used to say when I was a kid but spent all day twiddling bits. Display with a custom handle, log in with an email address.

#### Just an Email

Similar to the above case, but without a custom handle. For when you really can't be bothered to do much more "signing up" than logging in with Persona.

#### Old usernames

Whatever handle you use these days, you might have old ones, which should all 301 redirect to what you call yourself *now*. You could call this the "Wildstyle" use case.

As you might imagine, this is as simple as setting up password/email/whatever auth for your current username, and dropping auth for your old ones. Your alt handles will all still redirect to your current preferred one, and they'll all still belong to you!

#### No display data

Have you ever seen a reddit post where a comment was authored by `<deleted>`? Works the same way. We do provide options for nuking all the content you've ever posted, but when you just want to run off into the night anonymously, it's as simple (technically speaking) as deleting all your display names.
