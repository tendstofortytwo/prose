---
title: Hello, world!
summary: In my very first blog post specifically for this new platform, I describe how I created said platform.
time: 1616127842
---

"Hello, world!" as the title of my first blog post sounds almost clichéd. But I think it's a good choice, because with this blog, I'm not just starting to write -- though that is a big part of it. I'm also taking this opportunity to properly design a piece of web software. Like *properly* properly, the kind of proper you only expect from someone once they have taken a university-level course on designing and structuring programs (heh). I plan to create this blog as a fairly minimal Go-powered web server and use Markdown files to store and dynamically serve content.

### Why am I creating a blog?

There are a few reasons:

* **mathNEWS:** Over the past year, I have been involved with [mathNEWS](https://mathnews.uwaterloo.ca), the student publication of the Faculty of Mathematics at the University of Waterloo. I have been a writer for almost a year, and now I am one of the editors of the publication. mathNEWS has been amazing in bringing back the joy of writing for me. However, with it being a generally not-serious student publication, I don't feel comfortable putting some of my more personal thoughts on it. Also, since mathNEWS is more focused on print, it's not the most elegant medium for talking about code. This prompted me to create a platform of my own, focused more on the web, and solely about me. A blog!

* **Golang:** I have worked with the language a fair bit -- first at my research assistant position at the University of Waterloo, and later on as a software engineering intern at [Tailscale](https://tailscale.com). I love the language so far, and I'm looking for an excuse to write more of it. This is my excuse.

* **Cadey A. Ratio:** Semi-related to the above point -- I met them at Tailscale and really like their [blog](https://christine.website/blog). The blog, though it looks like a bunch of static web-pages loosely strung together, is actually written in Rust and dynamically serves every page fast enough for it to feel static. Great content and great software; great inspiration for me to make a blog of my own.

* **This song I found in the soundtrack of Forza Horizon 3:** I find the song Patience by Bad Suns to be pretty relaxing and inspiring -- a perfect song to listen to semi-awake, making your morning coffee. At one point, the lyrics from the song go like this:

	> I've been writing my thoughts down
	>
	> To clear my mind
	>
	> To try and figure out my brain
	>
	> To confront and set aside my pain

	I don't really have any particular "pain", but the rest of it sounds like a really good idea. Alone, a song lyric might not have been enough to motivate me. But combined with the rest of the factors all coming together in such close temporal proximity... I'm not saying it's a sign, but... ¯\\\_(ツ)_/¯

### Looks: ol' reliable, but with a few new considerations

I think my [current website](https://nsood.in) looks pretty good. A lot of the design just boils down to "get out of the way and let the projects talk for themselves", which I think is a good thing. The style is basic but it looks good and emphasizes content based on its importance without coming off as too in-your-face. I decided early on that this is the style I want to emulate, and adapt for the blog. Now it was just a [simple matter of programming (and design)](https://en.wikipedia.org/wiki/Small_matter_of_programming).

<figure>
<img src="/img/blog-init/current-site.png">
<figcaption>"Current" website as of writing, in case I change it.</figcaption>
</figure>

As editor of mathNEWS, I need to use Adobe InDesign to format the layout of each issue. While doing this, I came across a concept I hadn't seen in a long time before -- the [baseline grid](https://www.bookdesignmadesimple.com/book/baseline-grid/). The idea is that you set up horizontal lines at equal intervals down the page, and lay down your text such that each line of text rests on the baseline. It's commonly used in print design to ensure consistent alignment of things. The baseline grid is used less frequently in web design, however. Some people advocate for [using it more](https://vanseodesign.com/web-design/baseline-grids-web/) at least as a guide, so I decided to give it a try.

<figure>
<img src="/img/blog-init/baseline-grid.png">
<figcaption>My blog with the baseline grid visible.</figcaption>
</figure>

It actually looks very good, I think! Some of the individual elements might seem spaced strangely at first, but as a whole, the design seems to work very well, almost as though all the strangeness "cancels out". As someone who hasn't studied design in detail, I think this is good enough for now. Though there will probably be changes to this design in the future.

### Tech: Go with the flow

When I first had the idea to make this blog, I was considering both Go and Rust. I have programmed in Rust a little bit before. I made an [emulator](https://github.com/namansood/chip8-rust) for an old microprocessor/language called CHIP-8 with it once, and I remember liking it. That said, my recent experiences with Go have been even more positive. I don't know what it is about the language, but basically all the design decisions *feel* correct. Go seems to give you just enough control to keep things interesting, while exerting enough control to keep things from getting dangerous. I'm sure there are hidden cobwebs and dragon lairs somewhere down here, but this has been enough of a fun experience that I decided that I wanted more of it.

Given that I have heard so much about Go providing excellent standard library tools for web applications, my first instinct was to create the blog using those tools -- in this case, the `net/http` package. Immediately, however, I started running into trouble.

#### Routing

My previous backend development experience has been primarily Node.js on the Express framework. If I wanted to serve a home page at `/` and blog posts at `/title-slug`, I would do this in Express:

```js
router.get('/', function(req, res) {
	// generate and send home page
});

router.get('/:slug', function(req, res) {
	// generate and send blog post page, slug is in req.params.slug
});
```

The second route is the problem -- it seems that Go's built-in router doesn't support parameters in the URL by default. The internet seems to be torn on this issue -- half the people suggest that I use something like the [gorilla/mux](https://github.com/gorilla/mux) package, while the other half suggest [creating my own router](https://benhoyt.com/writings/go-routing/). I don't really like either solution. The first goes against the spirit of what I wanted to do when I started -- use the standard library tools to create the server, while the second seems a bit overkill -- I only need to serve a home page, a page for individual blogs, some static content, and maybe a custom error page. 

I ended up going with something similar to [Alex Wagner's approach](https://blog.merovius.de/2017/06/18/how-not-to-use-an-http-router.html). He calls it "not using an HTTP router", but I don't think that's quite correct. After all, there is a singular point of entry which then sends the request down different code paths depending on the URL and request type. That's... a router. That said, a router for a simple website can be as simple as the website itself. Here's what my router currently looks like (with some implementation details elided):

```go
func (s *server) router(res http.ResponseWriter, req *http.Request) {
	...
	s.logRequest(req)
	res = &errorCatcher{...}
	slug := req.URL.Path[1:]

	if slug == "" {
		s.homePage(res, req)
		return
	}
	...
	for _, p := range s.postList {
		if p.Slug == slug {
			s.postPage(p, res, req)
			return
		}
	}
	s.staticHandler.ServeHTTP(res, req)
}
```

That's it. No more routing code required. All I do is look at the path after the `/`. If it's empty, return the home page. If it matches a blog post slug, render that blog post. Otherwise, pass it to the static handler, which will serve a file if possible, or a 404 otherwise. I could probably clean this up a bit -- make the page handlers be standard-library-compatible, and improve the logging support (the current implementation won't be able to log the HTTP status code). These things are fixable, though, and when I fix them, they won't increase the size of the router itself.

You may notice though, that I said something about a custom error page, and the only thing that seems to be able to handle it is that `&errorCatcher{...}` wrapper. About that...

#### Error page handling

By default, Go doesn't provide any custom HTTP error handling hooks, which I found a bit weird -- surely a server-wide 404 page is a common enough ask to build that into the net/http package? But that's not too huge a problem, because Go provides something else -- interfaces.

If you're not familiar with Go interfaces, an interface is essentially a list of function signatures for methods, and any type that implements methods like these automatically implements that interface. Notably for us, `http.ResponseWriter` is an interface. You can see the complete interface [here](https://golang.org/pkg/net/http/#ResponseWriter), but the interesting part are the functions `WriteHeader()` and `Write()`. The documentation for the `ResponseWriter` states that:

> If WriteHeader is not called explicitly, the first call to Write will trigger an implicit `WriteHeader(http.StatusOK)`. Thus explicit calls to WriteHeader are mainly used to send error codes.

Which means, if anyone wants to write an HTTP error, they have to call `WriteHeader()`, which we could just create our own version of, if we make our own `ResponseWriter` that wraps the one in the standard library. Then we can intercept any error-y HTTP codes like so:

```go
func (ec *errorCatcher) WriteHeader(statusCode int) {
	if statusCode == 404 {
		// Handle custom 404 page here woohoo!
		return
	}

	if statusCode >= 400 && statusCode < 600 {
		// Handle other error pages.
		return
	}

	// If status is not an error, just pass it on to
	// the wrapped ResponseWriter.
	ec.res.WriteHeader(statusCode)
}
```

Now this is pretty cool. But recall that in my case, "handling" means serving my own HTTP error pages. Which means that I'll be calling `ec.res.Write(htmlContent)` or something of the sort. But whoever just called `WriteHeader()` doesn't know this, and they're going to be sending their own `Write()` calls with their own error page content! I don't want two error pages worth of content to show up, so I fixed this with an age-old trick: lying to the caller.

First, we fabricate the fiction:

```go
...
if statusCode == 404 {
	ec.res.Write(...)
	// If this variable is set to true, that means that a page
	// (presumably an error) has already been sent to the user,
	// and no further content should be sent.
	ec.handledError = true
	return
}
...
```

Then we sync up the stories and lie to the caller's face:

```go
func (ec *errorCatcher) Write(buf []byte) (int, error) {
	// If we have already sent a response, pretend that this
	// was successful without actually doing anything.
	if ec.handledError {
		return len(buf), nil
	}
	return ec.res.Write(buf)
}
```

Easy peasy! Custom 404 pages work perfectly now, check one out [here](/ohno/whereisthecontent). As of writing, it's nothing special, but at least I know that the ability to modify the page is there should I need it.

### Hosting: Cheap and reliable, hopefully

Choosing a hosting provider was surprisingly hard. I narrowed it down to these few options, from which I filtered due to increasingly petty reasons:

* **Linode:** Discounted due to security concerns voiced by many, many people on the internet. The Wikipedia section on [security incidents](https://en.wikipedia.org/wiki/Linode#Security_incidents) doesn't look _terrible,_ but I would rather not deal with someone with that kind of reputation if I can help it.
* **AWS EC2:** Discounted due to previous bad experiences with surprise costs. I'll probably come back to this when I'm not a cheap college student and need something that scales, but that day is not today.
* **AWS Lightsail:** Discounted due to complaints of worse performance than competitors. This is _probably_ not an immediate concern for me, but I would still prefer something that doesn't rely on the AWS burst-performance-credit mechanism, which also came back to bite me with EC2 previously.
* **DigitalOcean:** Discounted due to the fact that coupon codes are only valid for new signups and I already had an account from my time at [The Girl Code](https://thegirlcode.co).

That left **Vultr,** a company that people on the internet seemed to be mostly happy with, and that I could find fairly generous coupon codes for. I'm sold! Or I guess I will be in two months when the coupon expires.

I might move to something like DigitalOcean or AWS EC2 later down the road when I need (and can afford) the extra compute power, but for now, Vultr looks like it'll be plenty.

### Conclusion, and plans for the future

I have a blog now! Hopefully I keep writing for it. This first post took a lot of time to pen down, but part of that might be because I was writing it while creating the blog. I aim to keep myself to a similar schedule to mathNEWS, so at least one post every couple of weeks. One of my hopes is that I'll use "ooh I can get a blog post out of it" as an excuse to look into more cool new technology. I think my growth as a computer enthusiast has stagnated a bit, so hopefully this can give me the push I need.

See you soon, world!
