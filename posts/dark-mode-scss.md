---
title: Dark mode, and the interaction of SCSS and CSS variables
summary: Where I use an ugly hack to get the best of both worlds.
time: 1617482030
---

The CSS for this blog was created in the image of my main website, as I described in my [previous post](/hello-world). But my website didn't have a dark mode -- or at least, none that was shown as an option or automatically configured based on the user's browser settings. That's mostly down to the fact that I presented each of my web development projects with its own color scheme, but I didn't have a very good way to translate those color schemes into dark mode.

Fun fact: There *is* a dark mode "easter egg" on my main website, which you can access by typing `nox` anywhere on the page. Type `lumos` to bring back the light. This dark mode basically just inverts all the colors, sets all the colors to grey, and then re-inverts the images so the two inversions cancel out just for the images. Inverted images would be bad.

So, making a dark mode for the blog. Easy, just change all the colors, right? That's what I thought it'd be. In anticipation of this, I'd even written the stylesheet with a bunch of colors as SCSS variables right at the top, which I used for everything.

```scss
$bgColor: #fff;
$bodyColor: #555;
$strongColor: #333;
$fadeColor: #999;
$accentColor: #3498db;
```

When you want to make special styles for certain clients in CSS, you generally need to use an `@media` rule. So is the case for dark mode -- you need `prefers-color-scheme: dark`:

```scss
$bgColor: #fff;
body {
	background-color: $bgColor;
}
@media(prefers-color-scheme: dark) {
	$bgColor: #000;
	...
}
```

So I save this, refresh the page, and... nothing. The background is still as white as always. If you look this up online, the official explanation is that SCSS variables are compile-time entities, and they can't update based on runtime needs. Which is... true, but that is not what I expected to happen. I expected it to regenerate the necessary styles inside the media query. So for the above example, I would have expected:

```scss
/* hypothetical result of above, does not actually happen */
body {
	background-color: #fff;
}
@media(prefers-color-scheme: dark) {
	body {
		background-color: #000;
	}
}
```

But that doesn't work. Oh well, let's try something else. If the problem with SCSS variables is that they are not runtime entities, we can use the kinds of things that _are_ runtime entities -- CSS variables. Because our variables are now runtime entities, they can be updated by the browser on-the-fly, even in response to media queries and such. So we translate our variables into the new format, and we get:

```scss
:root {
	--bgColor: #fff;
}
body {
	background-color: var(--bgColor);
}
@media(prefers-color-scheme: dark) {
	:root {
		--bgColor: #000;
	}
}
```
And this works! To my delight, the background turns black, as expected. And so I turn all the other `$variables` into their CSS equivalent `--variables`, they all seem to work fine. At least, that's until I notice how my links look.

The way the styling for my links works right now is a medium-sized hack in itself -- the "underline" is actually a linear gradient that goes transparent-to-grey-to-transparent. If I set the size of the gradient to be one `line-height` and then align the grey part of the gradient correctly by twiddling the gradient stops, it looks like a horizontal line approximately where an underline should be. The benefit of using this over `text-decoration: underline` is that I can stylize the underline a bit, and the benefit of using this over `border-bottom` is that unlike a border I can position it closer to the text without changing the size of the element. The code for this weird background-underline, before all the dark mode stuff, looked like this:

```scss
a {
	...
	background-image: linear-gradient(
		transparent 80%, 
		lighten($fadeColor, 15%) 80%,
		lighten($fadeColor, 15%) 87.5%,
		transparent 87.5%
	);
	...
}
```

Weird and hacky as it might seem, it _worked,_ at least until I started using CSS variables. As you can see, I wanted the color of the underline to be a lighter version of my `$fadeColor` variable, so I used the `lighten()` function in SCSS. Problem is, since `lighten()` is an SCSS function, the color I put into it needs to be known at compile time -- so something that changes at runtime, like `--fadeColor`, won't work. Another, more subtle issue, is that I wouldn't want to use `lighten()` in dark mode at all -- what I actually want is a "fainter" color, which looks lighter in light mode, or darker in dark mode. So I want to use different compile-time functions depending on something that isn't known until runtime. This doesn't sound possible. What do?

The first step to solving this is to introduce a new CSS variable, called `--faintFadeColor`. This new color is set to a lighter `--fadeColor` normally, and a darker `--fadeColor` in dark mode. But wait. `--fadeColor` is a runtime entity now, so we can't pass it to `lighten()` or `darken()`! We're back where we started... or so it seems. The second step, which will fix this problem, is to define some new SCSS variables.

We don't know what `--fadeColor` will be when we compile the SCSS file, but we *do* know it will be one of two things -- either a light mode color, or a dark mode color. And we also know what these two colors are at compile-time. So we define new SCSS variables `$fadeColorLight` and `$fadeColorDark`, and we assign the CSS variable `--fadeColor` one of these two based on media queries.

This might be a bit confusing, so hopefully code will help:

```scss
$fadeColorLight: #999;
$fadeColorDark: #99b;

:root {
	--fadeColor: #{$fadeColorLight};
	--faintFadeColor: #{lighten($fadeColorLight, 15%)};
}

a {
	...
	background-image: linear-gradient(
		transparent 80%, 
		var(--faintFadeColor) 80%,
		var(--faintFadeColor) 87.5%,
		transparent 87.5%
	);
	...
}

@media(prefers-color-scheme: dark) {
	:root {
		--fadeColor: #{$fadeColorLight};
		--faintFadeColor: #{lighten($fadeColorLight, 15%)};
	}
}
```

*(Note: the weird `#{}` syntax for assigning SCSS values to CSS variables is due to a SCSS syntax incompatibility with CSS variables requiring any invalid CSS be produced literally: https://github.com/sass/sass/issues/1128)*

To make it a bit clearer as to what's happening, this is the CSS the above will compile to:

```scss
:root {
	--fadeColor: #999;
	--faintFadeColor: #bfbfbf;
}

a {
	...
	background-image: linear-gradient(
		transparent 80%, 
		var(--faintFadeColor) 80%,
		var(--faintFadeColor) 87.5%,
		transparent 87.5%
	);
	...
}

@media(prefers-color-scheme: dark) {
	:root {
		--fadeColor: #99b;
		--faintFadeColor: #6b6b9c;
	}
}
```

During the compile step, `--fadeColor` and `--faintFadeColor` are defined with the values of `$fadeColorLight` and its lightened version in the normal light-mode case (the first `:root`). These are both known at compile time so this works. Inside the media query (in the second `:root`), the variables are redefined to the values of `$fadeColorDark` and its darkened version. Both of these are also known at compile time, so this also works. 

Now, at runtime, when in light mode, the media query won't apply, and `--fadeColor` and `--faintFadeColor` will be correctly defined for light mode. In dark mode, the media query applies and both the variables are redefined to the dark mode values thanks to that. And since these variables are runtime values, the `a` that uses `--faintFadeColor` will resolve to whatever the final value turns out to be, and use the correct dark/light-mode specific color in either case.

Finally, all the colors resolve correctly, and dark mode works as expected. If your browser supports `prefers-color-scheme` (which you can check [here](https://caniuse.com/prefers-color-scheme)) and you have it set to dark mode, you should see it now! Otherwise, the blog defaults to light mode, at least for now, because if your browser is old enough to not support this, you're probably more used to seeing light mode pages anyway.

Phew, and that's that! CSS is always an interesting adventure, and figuring out what the issue was here was a really fun exercise in debugging.
