Community Service
=================
First, a little backstory. Until 2007, there was something called Community
Server. It was forum software that had quite a few bugs and vulnerabilities,
including one we named "Signature Guy", where a post could contain arbitrary
HTML and include what appeared to be the next post. It even worked in forum
signatures.

Sadly, Community Server (which had a free non-commercial edition) was
discontinued. It was replaced with a piece of software that costs
[a minimum of $2100.00 per month](http://telligent.com/products/pricing-and-editions/).
You can still see the last version of Community Server running at
[The Daily WTF](http://forums.thedailywtf.com/forums/).

Another "fun" part of Community Server was the ability to add arbitrary
categories to any post you made. This led to a script that inserted a
`TagException at 0x12840d4f` tag with a random pointer to each post. Needless
to say, the database did not take it well. And neither did the web server.
Each page contained every tag ever used on the forum - twice - once
urlencoded and once with html entities.

There were a lot of less-fun aspects of Community Server as well. Posts could
never truly be deleted. There was nothing in place to stop spam (other than a
"report post" feature that did nothing but email every moderator on the site).
`logout.aspx` could be included in a post as an image. A post could contain
`<!--` and the remainder of the page would be commented out.

Bugs and vulnerabilities weren't the only things that made it hard to use
Community Server. Each page would include 5 distinct stylesheets that overrode
each other. These stylesheets were explicitly noncacheable, so when the server
was slow, a page might be in a completely different shape than what it should
normally look like. Tables within tables within tables were also common.

What's different about this one?
--------------------------------
* Any action that can be performed by a logged-in user requires a user-specific,
  action-specific token which expires after 24 hours. That means no more
  `<img src="/logout.aspx">`!
* Static content is aggressively cached. All content is gzipped.
* We use [Couchbase](http://www.couchbase.com/) as our database, which, unlike
  MSSQL Express, has no database size limit for the free version, and is
  document-based instead of row-based, allowing more complex indexes.
* All input is sanitized. There's no SQL to worry about.
* It runs on Linux, Windows, MacOS X, Plan 9, ...
