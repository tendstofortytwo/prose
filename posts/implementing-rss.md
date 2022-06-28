---
title: Implementing RSS for my blog (yes, this one!)
summary: The actual implementation was easy, but wiring everything in is getting harder. This code needs a refactor...
time: 1656392139
---

A few months ago, a friend of mine mentioned that they still, in the year 2021, used RSS feeds to subscribe to the blogs they liked. I was incredulous? RSS? That orange button I used to see on websites years ago? The one that returned XML if you accidentally clicked it? People still used that? People still _provided_ that?

Apparently the answer was yes. My friend also said that if I implemented RSS on my blog, they would subscribe to it. So I thought about it, and decided to do it. I forgot about it until today, but when I did remember it didn't take long to implement at all.

### Figuring out what to do

I'll be honest, I don't *really* know how people in real life implement RSS. I'm sure they all have WordPress or Jekyll blogs and there's plugins for this sort of thing. I figured that probably wasn't going to help me since this blog runs on a [bespoke mess of Go](https://github.com/tendstofortytwo/prose), so I went straight for the official RSS specification, which can be found [here](https://www.rssboard.org/rss-specification). As of writing, we were at RSS 2.0.11.

RSS stands for *Really Simple Syndication*, and you can really tell they mean it. The spec document is short and to-the-point -- it describes what an RSS XML file looks like. You make that XML file available and people can point their RSS readers to it and It Just Works. So, I just need to make that XML file happen.

### What an `rss.xml` looks like

You can go read the spec if you want the full details (highly recommend -- as I said, very short and easy to read), but I think my templated [`rss-channel.xml`](https://github.com/tendstofortytwo/prose/blob/a1ad26124c23d930fbbc4ca9e374e6ee9390aaa0/templates/rss-channel.xml) does a good job of describing a minimal structure.

```xml
<rss version="2.0">
    <channel>
        <title>{{title}}</title>
        <link>{{link}}</link>
        <description>{{description}}</description>
        <language>en-US</language>
        <pubDate>{{pubDate}}</pubDate>
        {{{items}}}
    </channel>
</rss>
```

An RSS feed consists of one singular channel, which must have a title, link, and description. I added some of the optional fields like language and publishing date since they weren't hard to figure out for me, but there's many others that I skipped over. Following these are the items, which describe individual posts in your feed. My [`rss-item.xml`](https://github.com/tendstofortytwo/prose/blob/a1ad26124c23d930fbbc4ca9e374e6ee9390aaa0/templates/rss-item.xml) is an example of that.

```xml
<item>
    <title>{{metadata.title}}</title>
    <link>{{getFullUrl slug}}</link>
    <description>{{metadata.summary}}</description>
    <author>mail@nsood.in</author>
    <pubDate>{{rssDatetime metadata.time}}</pubDate>
</item>
```

No single field in the item is mandatory -- the only requirement is that you either have a title or description (or both). I added in all the fields that I can show without changing my post storage format. Now that I knew what this file looked like, all that was left was to fill in the templates.

### Creating the feed

The code for this blog has been a lesson in premature abstraction for me. When I built it, I abstracted so many things away from each other, adding so much complexity at every step. Some of this complexity was helpful, and some of it was a hinderance. And sometimes, it turned out that even in making all the extra considerations that I did, I ran into cases where I could not simply extend my framework; I had to modify it.

The way the blog works is that posts are stored as Markdown files in `posts/`, and a file watcher watches that folder. At the start of the program, every file has its Markdown is rendered, and the HTML, along with the post metadata (title, subtitle, publish date), is stored in a `Post` structure in memory. Then whenever a file is fetched, the HTTP mux just sends the bytes that have been pregenerated for every page and stored in memory. The idea for this was that fetching things from memory would be quicker than opening/reading/closing a file -- maybe a useful optimization, but one that led to a lot of complexity.

One of the causes of complexity was that I store dates as Unix time. This was mostly because I find it a lot easier to keep track of one number that I generate from `Math.floor(Date.now() / 1000)` rather than mucking around with formatting every time. Rather than format the date manually every time I write a blog post, I just get Go to format the Unix timestamp the way I like. However, my HTML (and now XML) templates go through [`raymond`](https://github.com/aymerick/raymond) to insert data, and `raymond` just takes a `struct` and uses its fields directly. So to format the date before I display it, I have to register a "helper function" with raymond, and call that helper function in my template.

RSS expects its `pubDate`s to be in a particular format, so I had to write a new helper called `rssDatetime` to generate the RSS datetime, then abstract away the actual Unix timestamp to string generation because I realized I needed the same thing in two completely unrelated places -- in `server.go` for the channel-wide `pubDate`, and in `post.go` for the item-specific `pubDate`. Now I need to put this abstracted-away function somewhere, and I decided to put it in `server.go` -- which means that in `post.go` I call a function that doesn't appear anywhere in the file. Thanks to the magic of Go packages this works, but it represents a tangling-together of the code that is the exact thing I wanted to avoid.

Someday I'll clean up the code. But for now, the datetime generate fine and the `rss.xml` seems to generate correctly. Now all that's left is testing.

### Testing

This is the fun part -- testing isn't yet complete! I did install the `liferea` RSS reader and point it at `http://localhost:8080/rss.xml` and that worked fine, but we don't yet know how this will interact with production. Part of the reason I'm writing this blog post is to test RSS -- I'll deploy the post and check that the reader shows it as expected. Hopefully it works out fine.

Maybe I should write tests for this. hmm...

### Closing thoughts

It finally happened -- I wrote a bunch of "clever" Go code last year to make this blog and now it's hard for me to grok what I've written. It's not _too_ bad though -- I was able to get RSS to work, after all! -- but I think it's a sign that I need to refactor a lot of this code to be simpler. I don't want to abandon the all-text-files-load-from-memory thing _yet,_ we'll see if I can tame the complexity beast enough while still keeping that one thing I think is cool.

Gonna `git commit -as && git push` this now, and start refreshing the RSS reader. Fingers crossed!

**Update (June 28, 2022, 11:58AM):** Following some really helpful suggestions on this [lobste.rs](https://lobste.rs/s/lrnqsv/implementing_rss_for_my_blog_yes_this_one) thread, I made some changes to my feed and blog HTML for better compatibility and discoverability. If you're following this, highly recommend following the suggestions listed there -- like running your feed through the [W3C validator](https://validator.w3.org/feed/) or adding a `<link>` tag corresponding to your RSS feed in the HTML. Thanks to snej and carlmjohnson for the suggestions! Also, the feed works, test passed. :)
