---
title: Writing a lambda calculus interpreter in Rust
summary: Blazing-fast unreadably-dense mathematical objects for everyone.
time: 1692057073
---

I have studied lambda calculus at university a lot -- our [first year CS course](https://student.cs.uwaterloo.ca/~cs145/) was functional programming, there was a [second year logic course](https://cs.uwaterloo.ca/~plragde/245/) where lambda expressions were used as proofs of mathematical statements represented as types, there was a [fourth year CS course](https://student.cs.uwaterloo.ca/~cs442/W22/) which taught different programming paradigms (including functional programming) by just implementing an interpreter for a basic language in that paradigm... it's kinda becoming second nature to me at this point. Given that I've done this so often, I figured I should make a lambda calculus interpreter in Rust to become more familiar with how to use it.

### Quick introduction to lambda calculus

Implementing a lambda calculus interpreter is _really_ simple. There are three types of expressions, and two reduction operations.

The types of expressions are:

* A **variable**: this is just a string that represents some value. A variable can be "free" or "bound" -- a free variable is one that has no meaning in the current context, and a bound variable is one that, well, does have a particular meaning. You assign meanings using the next kind of expression:
* An **abstraction**: An abstraction is a package of two things -- a variable and an expression (called the body). Inside the body, the variable of the abstraction is considered "bound". You can write an abstraction as `\x.q`, where `x` is the variable and `q` is the body. If you refer to `x` inside the contents of `q`, it refers to the same `x` as the one in this abstraction, unless there's another binding to `x` deeper inside.
* An **application**: An application is also a package of two things -- an expression (rator) and another expression (rand). If the rator is an abstraction, you can "apply" the rand to it by replacing all instances of the bound variable in the abstraction with the rand. So something like `((\x.a b x) y)` can turn into `a b y`. This is called **beta-reduction**.

And that's it! These three rules define every lambda calculus expression. I even snuck in one of the reduction operations in there for good measure. The other reduction operation is **alpha-conversion**, which basically means "you can replace a bound variable of an abstraction with another one". So you can rewrite `\x.x y z` as `\n.n y z` and that would be fine.

Note that you _cannot_ rewrite it to `\y.y y z`! This is because the `y` in the body of abstraction is supposed to be free, and renaming the binding variable to `y` would bind this previously-free variable. This process of accidentally binding a free variable is called "variable capture" and something we want to avoid. In abstract theory land, we can just say we always use a "fresh" (unused) name for renaming, but while implementing a lambda calculus interpreter, this means being very careful with how you treat your variables and deal with conflicting names.

Also, if it helps, you can think of abstractions as functions and applications as calling them -- it helps my programmer brain, at least.

### Representing lambda expressions in Rust

A lambda expression can be one of three things, and each of them are treated differently but considered to be the same type of object. This is the textbook definition of a sum type, or an `enum` for Rust. So I write out the type for all my expressions as follows:

```rust
enum Expr {
    Var(String),
    Abs(String, Expr),
    App(Expr, Expr)
}
```

which is as close to a barebones syntactic representation as I can get. Looks good!

Sadly, this does not work.

The linter complains that `recursive type Expr has infinite size`. This makes sense, because Rust, as a :sparkles: low-level language :sparkles:, cares about the sizes of `enum`s and their structure in memory. A language like TypeScript, which I am more used to, would simply take the `Expr`s nested inside by reference, so each individual nested `Expr` would only take up space equivalent to the size of a reference inside the outer `Expr`. But Rust takes what I wrote to mean that the `Expr` itself should have enough space to fit an `Expr` inside it, which then itself would need enough space for an `Expr` to fit... you can see where the "infinite size" thing comes from.

Thankfully, there's an easy solution (that the linter itself suggests); boxes!

In Rust, a box is an owned reference to something on the heap. Rust cares about "ownership" of heap memory -- a chunk of heap-allocated memory is always owned by someone, and when that someone goes out of scope, the chunk of memory is deallocated. So in this case, each `Expr` would hold references to its children `Expr`s, and when the parent goes out of scope, the children get deallocated. This is pretty neat, since it does what I would normally do in C++ with a constructor/destructor, but automatically.

Now our `Expr` looks like this:

```rust
enum Expr {
    Var(String),
    Abs(String, Box<Expr>),
    App(Box<Expr>, Box<Expr>),
}
```

To see these `Expr`s in action, we can start by making them printable on the screen. The way Rust allows us to do this is using "traits". A trait is a common set of methods that you can implement for your types, and then the types can be used anywhere that trait is required. For example, by implementing the [`std::fmt::Display`](https://doc.rust-lang.org/std/fmt/trait.Display.html) trait, I can make it so that my `Expr`s are printable to the screen (or another user-facing output). Implementing a trait is as easy as writing an `impl Trait for Type {}` block in which you implement all the methods of that trait. For example:

```rust
use std::fmt::{Display, Formatter, Error};

// ...

impl Display for Expr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result<(), Error> {
        match self {
            Expr::Var(s) => write!(f, "{}", s),
            Expr::Abs(s, e) => write!(f, "λ{}.{}", s, e),
            // note: this branch is not yet fully correct
            Expr::App(u, v) => write!(f, "({} {})", u, v)
        }
    }
}
```

I really appreciate the fact that Rust strings are UTF-8, so I can use the lambda symbol `λ` for lambdas, as [Alonzo Church](https://en.wikipedia.org/wiki/Alonzo_Church) intended.

With this, we can print out our first lambda calculus expression, the identity function `\x.x`:

```rust
fn main() {
    let id = Expr::Abs(String::from("x"), Box::new(Expr::Var(String::from("x"))));
    println!("Hello lambda: {}", id);
}
```

Which results in the output:

```
$ cargo run .
warning: variant is never constructed: `App`
 --> src/main.rs:6:5
  |
6 |     App(Box<Expr>, Box<Expr>),
  |     ^^^^^^^^^^^^^^^^^^^^^^^^^
  |
  = note: `#[warn(dead_code)]` on by default

warning: `lambda` (bin "lambda") generated 1 warning
    Finished dev [unoptimized + debuginfo] target(s) in 0.00s
     Running `target/debug/lambda .`
Hello lambda: λx.x
```

There's a teensy warning about us not using function application anywhere yet, but that's fine, we'll get to that. The important thing is, we printed the identity function `λx.x`! Also note that our code for handling function applications isn't completely correct yet, but we can handle that in a minute once we realize where we're going wrong.

You'll notice that writing out these lambda expressions can get really long and complicated, what with the `Expr::Abs`-s and the `Box::new`-s. We can solve this problem by writing more code.

### Writing a parser for our expressions

It would be nice if we could write out lambda expressions and have the interpreter convert them into our `Expr` representations. To do this is called parsing, so we need to implement a parser. To make a parser, first we need to unambiguously write down a grammar -- the rules of how our lambda expressions must look.
```
expr := abstraction | application | term | "(" expr ")"
abstraction := lambda term "." expr
application := expr expr expr*
term := [A-Za-z]+
lambda := "\" | "λ"
```
These are the rules I came up with for describing the syntax I described above, along with the note that applications are left-associative -- `a b c` is interpreted as `((a b) c)`, not `(a (b c))`.

Note that one implication of these rules is that an abstraction extends as far to the right as it can, until the line ends or it encounters a closing parenthesis. This means that `(\x.x y)z` should be interpreted as the `x` and `y` being part of the abstraction's body, but not the `z`. But if we use our current `impl Display for Expr` implementation with this expression:
```rust
// warning: a mouthful
let res = Expr::App(
    Box::new(Expr::Abs(String::from("x"), 
        Box::new(Expr::App(Box::new(Expr::Var(String::from("x"))), Box::new(Expr::Var(String::from("y"))))))),
    Box::new(Expr::Var(String::from("z")))
);
println!("res: {}", res);
```
we get:
```
$ cargo run .
   Compiling lambda v0.1.0 (/home/nsood/lambda)
    Finished dev [unoptimized + debuginfo] target(s) in 0.25s
     Running `target/debug/lambda .`
res: (λx.(x y) z)
```
Well, that's not right. The output implies that `z` is part of the abstraction, which is not the case! We can fix this by writing a slightly better `impl Display for Expr`:
```rust
impl Display for Expr {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result<(), Error> {
        match self {
            Expr::Var(s) => write!(f, "{}", s),
            Expr::Abs(s, e) => write!(f, "λ{}.{}", s, e),
            Expr::App(u, v) => {
                match u.as_ref() {
                    Expr::Abs(_,_) => write!(f, "({}) ", u),
                    _ => write!(f, "{} ", u)
                }?;
                match v.as_ref() {
                    Expr::Abs(_,_) => write!(f, "({})", v),
                    Expr::App(_,_) => write!(f, "({})", v),
                    _ => write!(f, "{}", v)
                }
            }
        }
    }
}
```
Notice that this follows our rules -- since abstractions extend all the way to the right, we only need to parenthesize them if there is something to their right, so the right side doesn't "leak into" the abstraction. This can happen when abstractions are on either side of an application.

When they're on the left, it's easy to see why -- there's a right-hand-side by definition. When they're on the right, the application might be the left-hand-side of another application, so an abstraction on the right side of this application might have content to the right of it still. For example, in `(\a.\b.b a) (\x.\y.y x) c`, we need parenthesis around the `\x.\y.y x` so the `c` does not accidentally become part of it.

Also, since applications are left-associative, we also parenthesize applications that occur on the right side of another application. For example, `((x y) (z w))` can legally be written as `x y (z w)`, but not `x y z w`. With this, we get our correct answer:
```
$ cargo run .
   Compiling lambda v0.1.0 (/home/nsood/lambda)
    Finished dev [unoptimized + debuginfo] target(s) in 0.27s
     Running `target/debug/lambda .`
output: (λx.x y) z
```
> Side note: I could have also solved this problem by parenthesizing _all_ abstractions and applications, but that makes things really messy to read really quickly.

Now that we have `Display` working, we can get to writing the parser! This parser works in two stages -- tokenization and AST generation. Tokenization is the process of reading the one input string we have and breaking it up into its constituent characters. For example, `\x.x` might get broken down into "the lambda operator, a term `x`, and another term `x`". Notice that I skipped over the dot -- since it's used only for punctuation and doesn't really have a "meaning" beyond separating the lambda variable and expression, I ended up not making a token for it.

The tokens I did end up making look like this:

```rust
enum Token {
    LParen(usize),
    RParen(usize),
    Lambda(usize),
    Term(usize, String),
}
```

`LParen` and `RParen` are left and right parentheses respectively, `Lambda` is the lambda symbol, and `Term` is a string like `x` or `hello` that's used as a variable. The `usize` in all of them is used to store the position of the starting character of that token, so we can display it when printing errors. Now we can start writing a function to extract these out:

```rust
fn tokenize(input: &mut Chars) -> Vec<Token> {
    let mut res = Vec::new();
    let mut current_term = String::new();
    let mut char_position = 0;
    while let Some(ch) = input.next() {
        char_position += 1;
        let mut next_token = None;
        match ch {
            '\\' | 'λ' => next_token = Some(Token::Lambda(char_position)),
            ch if char::is_whitespace(ch) || ch == '.' => (),
            '(' => next_token = Some(Token::LParen(char_position)),
            ')' => next_token = Some(Token::RParen(char_position)),
            _ => {
                current_term.push(ch);
                continue;
            }
        }
        if current_term.len() > 0 {
            res.push(Token::Term(char_position, current_term));
            current_term = String::new();
        }
        match next_token {
            Some(token) => res.push(token),
            None => ()
        }
    }
    res
}
```
We're looping through all the characters and collecting our results as a vector of tokens, called `res`. In addition, we keep two more pieces of state. The first is that token position we talked about earlier, and that just increments by 1 in every loop iteration. The other one is a string. Notice that our "terms" can be more than one character long, so they will take multiple loop iterations to collect. So we collect them over time, and when we encounter a different token, we end our term token there.

For every character, we figure out the right kind of token it represents, and either store it in `next_token`, or for terms, append the character into the term string. Then for terms, we `continue` out of the loop iteration, since we don't want to do the next step if the term isn't complete yet. That next step being, we take the term string we've built and create a term token using it -- we want to only do this once we stop coming across characters that would be part of a term, since we're matching for terms "greedily".

Then finally, if we had stored something in `next_token`, we add that to the result as well. We need to ensure that this happens _after_ pushing the `current_term` token, because the `next_token` came after the `current_term` was finished -- we just didn't realize that `current_term` had finished until we saw `next_token`. After we go through the entire list of chars, we're done! We have a vector of tokens that we can turn into an AST.

Turning the tokens into an AST is a process I refer to as treeification. Note that I haven't seen anyone else call it that... but also, I have no idea what else you would call it. The code for treeification is more involved than tokenization, so I'll go over it in parts:

```rust
enum TreeifyError {
    UnclosedParen(usize),
    UnopenedParen(usize),
    MissingLambdaVar(usize),
    MissingLambdaBody(usize),
    EmptyExprList
}

fn treeify(tokens: &[Token]) -> Result<Expr, TreeifyError> {
    let mut i = 0;
    let mut res = Vec::new();
    while i < tokens.len() {
        match &tokens[i] {
            // ...
        }
        i += 1;
    }
    match res.into_iter().reduce(|acc, item| Expr::App(Box::new(acc), Box::new(item))) {
        Some(res) => Ok(res),
        None => Err(TreeifyError::EmptyExprList)
    }
}
```

The big picture is the same -- we go over the tokens one by one, and collect a vector of ASTs. This time we use an index through the tokens rather iterating over them, because it will prove useful to be able to change the index and go forward and backward. Finally, if there were multiple ASTs collected, we turn them into one tree by making them into applications of one onto the other -- if we collect three ASTs `T1, T2, T3`, we treat that as `T1 T2 T3 == ((T1 T2) T3)`.

A couple other things to point out: firstly, rather than accepting a `Vec<Token>`, I'm accepting a `&[Token]` -- this is called a _slice,_ and it allows me to call `treeify` with a reference to a contiguous slice of the vector. This is useful because I can recursively call `treeify` on parts of the vector without making a copy of it. The initial call to treeify just takes the largest possible slice -- the slice of all the elements of the vector. Also, I have declared an error enum, which gives you a bit of a spoiler on what problems we might run into while writing this function.

Notice that I have commented away the contents of the match expression in the loop -- that is where we decide how to deal with each incoming token. We will deal with each match arm individually:

```rust
Token::LParen(paren_idx) => {
    let mut nesting = 0;
    let mut j = i+1;
    let mut pushed_expr = false;
    while j < tokens.len() {
        match tokens[j] {
            Token::LParen(_) => {
                nesting += 1;
            }
            Token::RParen(_) => {
                if nesting == 0 {
                    let inside_expr = treeify(&tokens[i+1..=j-1])?;
                    res.push(inside_expr);
                    pushed_expr = true;
                    break;
                }
                nesting -= 1;
            },
            _ => ()
        }
        j += 1;
    }
    if !pushed_expr {
        return Err(TreeifyError::UnclosedParen(*paren_idx));
    }
    i = j;
},
```

And the first one is the most complicated one too! When we see a left parenthesis, we go into matching mode -- we run forward in the slice with another indexing variable `j`, and look for a matching right parenthesis. Parentheses can be nested, so we keep track of a "nesting number" that calculates how deeply nested we are. Whenever we come across a left paren, we increase the nesting number, and we decrease it whenever we see a right paren. If we see a right paren with the nesting number being 0, then that means we found the matching parenthesis for our original left one! If this doesn't happen, that means that our left parenthesis was never properly closed, and we return an error.

We scoop up everything between these two parentheses, and recursively call `treeify` on them. The idea here is that something inside parentheses must be a complete lambda calculus expression on its own, so we can treeify it independently, take the result, and push that as a single tree in the result.

The next match arm is quite simple:

```rust
Token::RParen(paren_idx) => {
    return Err(TreeifyError::UnopenedParen(*paren_idx));
},
```

We find and consume every right parenthesis that has a matching left one before it in the arm above -- which means that if we come across an `RParen` here, there is no corresponding `LParen`. This is an error, so we just return that.

The next one is a bit more fun:

```rust
Token::Lambda(lambda_idx) => {
    if tokens.len() <= i+2 {
        return Err(TreeifyError::MissingLambdaBody(*lambda_idx));
    }
    if let Some(Token::Term(_, term_str)) = tokens.get(i+1) {
        let rest = treeify(&tokens[i+2..])?;
        res.push(Expr::Abs(term_str.to_string(), Box::new(rest)));
        i = tokens.len();
    }
    else {
        return Err(TreeifyError::MissingLambdaVar(*lambda_idx));
    }
},
```

The lambda token represents the lambda symbol, and here we expect to find at least two tokens after it -- one term for the abstraction variable name, and at least one more token that represents the body of the abstraction. Note that because abstractions extend all the way to the right, we can assume that the entire list of remaining tokens after the abstraction variable is the body. So we scoop up all of that, recursively call treeify, and build an abstraction out of our variable and new body. If we don't find at least two tokens, or the first token isn't a term, those are both errors, so we return the corresponding errors.

The last match arm is also pretty simple:

```rust
Token::Term(_, term_str) => {
    res.push(Expr::Var(term_str.to_string()));
}
```

A lone term is just a variable, and we treat it as that.

And that's all! We're done through the hard part of parsing, and we can try out our work with a simple `main()` function that takes a line as input, parses it, and then displays the parsed output:

```rust
fn main() {
    let mut buf = String::new();
    let stdin = io::stdin();
    stdin.read_line(&mut buf).expect("could not read from stdin");
    let res = tokenize(&mut buf.chars());
    let res = treeify(&res).expect("could not parse expr");
    println!("output: {}", res);
}
```

Running this, we get:
```
$ cargo run .
   Compiling lambda v0.1.0 (/home/nsood/lambda)
    Finished dev [unoptimized + debuginfo] target(s) in 0.32s
     Running `target/debug/lambda .`
\f.\x.(f (x y z)) a
output: λf.λx.f (x y z) a
```
It works! You can tell it did something because it replaces all the backslashes with lambda symbols, and removed the extra parentheses I added in that aren't actually required.

### Implementing computation

So now we have our lambda calculus expressions as syntax trees. Now all we need to do is implement alpha conversion and beta reduction. To do that, we should first implement _substitution._

Substitution is the process of taking an AST and changing out all the instances of a particular variable with some other expression. We need this in alpha conversion because that's how we rename the variables, and in beta reduction because that's how we apply values into the abstractions. 

Let's try to write a substitution function:

```rust
// spoiler alert: mildly incorrect
fn substitute(root: Expr, var: &str, val: &Expr) -> Expr {
    match root {
        Expr::Var(v) => {
            if v == var {
                val.clone()
            } else {
                Expr::Var(v)
            }
        },
        Expr::Abs(v, body) => {
            if v == var {
                Expr::Abs(v, body)
            } else {
                Expr::Abs(v, Box::new(substitute(*body, var, val)))
            }
        },
        Expr::App(l, r) => {
            Expr::App(Box::new(substitute(*l, var, val)), Box::new(substitute(*r, var, val)))
        }
    }
}
```

At first glance, this looks correct. To substitute a variable, you should replace it if you see the same variable, replace it inside the body of an abstraction unless the variables match (ie. there is a different local binding), and replace it in both sides of an application. 

Let's go with our intuition for now, and build a very basic version of beta-reduction on top of this:

```rust
fn test_apply(exp: Expr) -> Expr {
    match exp {
        Expr::App(l, r) => {
            match *l {
                Expr::Abs(v, b) => {
                    substitute(*b, &v, &r)
                },
                _ => panic!("unimplemented")
            }
        },
        _ => panic!("also unimplemented")
    }
}

fn main() {
    // ... treeify
    let res = test_apply(res);
    // ... output
}
```

We can try a very simple application here and see that it works:

```
$ cargo run .
   Compiling lambda v0.1.0 (/home/nsood/lambda)
    Finished dev [unoptimized + debuginfo] target(s) in 0.45s
     Running `target/debug/lambda .`
(\x.x) y
output: y
```

But let's try something a bit trickier:

```
$ cargo run .
    Finished dev [unoptimized + debuginfo] target(s) in 0.00s
     Running `target/debug/lambda .`
(\a.\b.a) b
output: λb.b
```

The meaning of the abstraction changed! `\a.\b.a` is supposed to be an abstraction that takes a first argument `a`, returns an abstraction that takes a second argument `b`, and then returns the `a` -- the first argument we provided. But if we send `b` as the input to `a`, the clash in naming with the inner variable changes the meaning of the abstraction into something that returns the second argument instead.

This is because while `a` is not bound by the `\b` abstraction, it might have been bound by something else instead (in this case, our `\a` abstraction). To solve this, we want to make sure that none of the free variables in our application's rand conflict with the variable of the abstraction's body. For example, when we did `(\a.\b.a) (b c)`, the variable `b` is free in the expression `(b c)`. So while we can substitute it into the `\a`, we won't be able to do that in the `\b`, since that abstraction binds a variable that's free in the expression we'll be putting in. Instead, to take care of this conflict, we will _rename_ the `b` of `\b` to be a fresh name that does not conflict.

To do this, first we find all the free variables:

```rust
fn free_vars(e: &Expr) -> HashSet<String> {
    let mut free_vars = HashSet::new();
    let mut bound_vars = HashSet::new();
    fn recur(e: &Expr, fv: &mut HashSet<String>, bv: &mut HashSet<String>) {
        match e {
            Expr::Var(v) => {
                if !bv.contains(v) {
                    fv.insert(String::from(v));
                }
            },
            Expr::Abs(v, body) => {
                // insert returns true if thing was inserted, false if it already existed
                // if it was inserted we need to remove it, if it already existed then we don't
                let need_to_remove = bv.insert(String::from(v));
                recur(body, fv, bv);
                if need_to_remove { bv.remove(v); }
            },
            Expr::App(l, r) => {
                recur(l, fv, bv);
                recur(r, fv, bv);
            }
        }
    }
    recur(e, &mut free_vars, &mut bound_vars);
    free_vars
}
```

This is written as two functions because a recursive function was the most natural way to traverse a tree for me, and the first function is the only way I can figure to create a `HashSet` once and then populate it with the results of traversing the tree. The reason I need to pass `fv` and `bv` as parameters into the inner function is that the inner function doesn't actually capture the environment of the outer one -- the only reason to put the inner function inside is to avoid polluting the global namespace of my program.

The meat of the logic is the inner function -- it goes through every node in an AST. If it's an `App`, it just goes down into both sides. If it's an `Abs`, it only goes into the body of the loop, and adds the variable of the abstraction as a bound variable while it is in the body -- after it is done, it removes the variable (but only if that variable wasn't already bound!) so it can be free again in calls to `recur()` outside the abstraction. If it's a `Var`, we check if the variable is currently bound -- if not, we add it to the list of free variables.

Now that we can get a list of free variables, we can amend our `substitute()` function to use them:

```rust
fn substitute(root: Expr, var: &str, val: &Expr) -> Expr {
    match root {
        // ...
        Expr::Abs(v, body) => {
            if v == var {
                Expr::Abs(v, body)
            } else if free_vars(val).contains(&v) {
                let nv = disambiguate(&v);
                let nb = substitute(*body, &v, &Expr::Var(nv.clone()));
                Expr::Abs(nv, Box::new(substitute(nb, var, val)))
            } else {
                Expr::Abs(v, Box::new(substitute(*body, var, val)))
            }
        },
        // ...
    }
}
```

This is basically a translation of the fix I described above -- if the free variables of `val` contain the abstraction's variable `v`, we make a fresh variable name using the `disambiguate()` function, change the body of the abstraction to use that new variable name, and then perform the substitution in the new body as usual.

The `disambiguate()` function works by just appending a number to the variable name, which is an incrementing counter so we get a unique value every time.

```rust
thread_local!(static DISAMBIGUATE_CTR: RefCell<u64> = RefCell::new(0));

fn disambiguate(w: &str) -> String {
    DISAMBIGUATE_CTR.with(|c| {
        let mut ctr = c.borrow_mut();
        *ctr += 1;
        format!("{}_{}", w, ctr)
    })
}
```

Fun fact: Rust doesn't let you have static variables because it's not safe to access them across multiple threads, so instead we create a thread-local variable -- so `DISAMBIGUATE_CTR` technically could have the same value in different threads, but since our application is single-threaded this will not be an issue.

And now we can see that we have application working correctly!

```
$ cargo run .
   Compiling lambda v0.1.0 (/home/nsood/lambda)
    Finished dev [unoptimized + debuginfo] target(s) in 0.41s
     Running `target/debug/lambda .`
(\a.\b.a) b
output: λb_1.b
```

Applying the first abstraction changes all the `a`s to `b`s, but then our code notices the conflict and changes the `b`s of the inner abstraction to `b_1`. Success! Also, note what we just did here -- we replaced the bound variable of an abstraction with another one. We just made alpha-conversion too, without even realizing it!

Now, we just need to replace the `test_apply()` function with a function that takes _any_ lambda calculus expression and reduces it as much as possible.

It's nice to have everything in terms of logical steps that we understand, so we pull out our alpha conversion code out of the `substitute()` function into its own:

```rust
fn alpha_convert(v: &str, body: Expr) -> (String, Expr) {
    let new_var = disambiguate(v);
    let new_body = substitute(body, v, &Expr::Var(new_var.clone()));
    (new_var, new_body)
}

fn substitute(root: Expr, var: &str, val: &Expr) -> Expr {
    match root {
        // ...
        Expr::Abs(v, body) => {
            if v == var {
                Expr::Abs(v, body)
            } else if free_vars(val).contains(&v) {
                let (nv, nb) = alpha_convert(&v, *body);
                Expr::Abs(nv, Box::new(substitute(nb, var, val)))
            } else {
                Expr::Abs(v, Box::new(substitute(*body, var, val)))
            }
        },
        // ...
    }
}
```

...write a similar function for beta reduction:

```rust
fn beta_reduce(abs: Expr, val: Expr) -> Expr {
    match abs {
        Expr::Abs(var, body) => {
            substitute(*body, &var, &val)
        },
        _ => panic!("can't apply to non-abstraction")
    }
}
```

...and put it all together:

```rust
fn reduce_aoe(e: Expr) -> Expr {
    match e {
        Expr::App(l, r) => {
            let (l, r) = (reduce_aoe(*l), reduce_aoe(*r));
            match l {
                Expr::Abs(_, _) => reduce_aoe(beta_reduce(l, r)),
                _ => Expr::App(Box::new(l), Box::new(r))
            }
        },
        _ => e
    }
}

fn main() {
    let stdin = io::stdin();
    loop {
        let mut buf = String::new();
        stdin.read_line(&mut buf).expect("could not read from stdin");
        if buf.starts_with("exit") {
            println!("bye");
            break;
        }
        let res = tokenize(&mut buf.chars());
        let res = treeify(&res).expect("could not parse expr");
        let res = reduce_aoe(res);
        println!("-> {}", res);
    }
}
```

And that's it! We just built our own lambda calculus interpreter. We can try it out with some basic lambda calculus expressions.

You can represent "true" and "false" by lambdas that take two parameters and return the first and second ones respectively, and implement if conditions by applying the then-clause and else-clause to your "boolean variable". A true variable will return the first parameter (then-clause), and a false variable will return the second parameter (else-clause).

```
(\c.\t.\f. c t f) (\t.\f.t) thenclause elseclause
-> thenclause
(\c.\t.\f. c t f) (\t.\f.f) thenclause elseclause
-> elseclause
```

This form of representing a boolean value is called a Chruch boolean, and is part of a larger system of Things You Can Do With Lambda Calculus invented by Alonzo Church, called [Church encoding](https://en.wikipedia.org/wiki/Church_encoding). Natural numbers can be expressed in Church encoding as well -- as a lambda that takes two parameter and applies the first parameter to the second one N times:

```
// this is 1
\f.\x.f x
-> λf.λx.f x
// this is 2
\f.\x.f (f x)
-> λf.λx.f (f x)
// this is also 2, expressed as 1+1
(\m.\n.\f.\x.m f (n f x)) (\f.\x.f x) (\f.\x.f x)
-> λf.λx.(λf.λx.f x) f ((λf.λx.f x) f x)
```
> Side note: I don't think you'll see a more obnoxious way of saying "one plus one equals two" outside the *[Principia Mathematica](https://en.wikipedia.org/wiki/Principia_Mathematica)*.

In the third expression above, `((\a.\b.\f.\x.a f (b f x))` is a lambda that will add two Church numerals together. The two generated from adding 1 and 1 results in something that looks quite different from the 2 we wrote out directly above, but the two expressions are beta-equivalent, which means that if you provide the same parameters to them, they beta-reduce to the same thing. We can check this by reducing both of the 2s by providing them the same values for `f` and `x`:

```
(λf.λx.f (f x)) hello world
-> hello (hello world)
(λf.λx.(λf.λx.f x) f ((λf.λx.f x) f x)) hello world
-> hello (hello world)
```

Going forward, I'll pass `f` and `x` to any numbers I output so we can see them as `f (f (f x))` or such instead of a potentially hard-to-parse unsimplified form.

Also, it's annoying to have to say `\m.\n.\f.\x.m f (n f x)` every time we want to add two numbers. It would be nice if we could assign readable values and then use them in our interpreter. We want definable constants!

We're going to change our grammar to the following:

```
statement := expr | assignment
assignment := term "=" expr
expr := abstraction | application | term | "(" expr ")"
abstraction := lambda term "." expr
application := expr expr expr*
term := [A-Za-z]+
lambda := "\" | "λ"
```

Now we have the idea of a "statement" -- a statement can be either an expression, or an assignment to a variable. Expressions continue to work as before, but statements will additionally take the expressions on their right-hand side and store them as the value for the variable on the left-hand side.

We can make `Statement`s similarly to `Expr`s:

```rust
enum Statement {
    Expr(Expr),
    Assignment(String, Expr)
}

impl Display for Statement {
    fn fmt(&self, f: &mut Formatter<'_>) -> Result<(), Error> {
        match self {
            Statement::Expr(e) => write!(f, "{}", e),
            Statement::Assignment(v, e) => write!(f, "{} = {}", v, e)
        }
    }
}
```

...add the equals sign as something our tokenizer can recognize:

```rust
#[derive(Debug)]
enum Token {
    // ...
    Equals(usize)
}

fn tokenize(input: &mut Chars) -> Vec<Token> {
    // ...
    while let Some(ch) = input.next() {
        // ...
        match ch {
            // ...
            '=' => next_token = Some(Token::Equals(char_position)),
            // ...
        }
        // ...
    }
    // ...
}
```

...update `treeify` to use the new kind of token we introduced -- even though it won't have much use for it:

```rust
#[derive(Debug)]
enum TreeifyError {
    // ...
    IllegalAssignment(usize)
}

fn treeify(tokens: &[Token]) -> Result<Expr, TreeifyError> {
    // ...
    while i < tokens.len() {
        match &tokens[i] {
            // ...
            Token::Equals(equals_idx) => return Err(TreeifyError::IllegalAssignment(*equals_idx))
        }
        // ...
    }
    // ...
}
```

...and write a function to _actually_ use our new token as a "preprocessor" of sorts for `treeify`:

```rust
fn build_statement(tokens: &[Token]) -> Result<Statement, TreeifyError> {
    if tokens.len() > 2 {
        if let Some(Token::Term(_, s)) = tokens.get(0) {
            if let Some(Token::Equals(_)) = tokens.get(1) {
                let exp = treeify(&tokens[2..])?;
                return Ok(Statement::Assignment(s.clone(), exp));
            }
        }
    }
    
    let exp = treeify(tokens)?;
    return Ok(Statement::Expr(exp));
}
```

`build_statement` checks if the list of tokens we got represents an assignment -- that would be when the list starts with a term, and then an equals sign. If that's the case, it treeifies the rest of the tokens and stores the string of the term as the name of the constant that this assignment will define. Otherwise, it assumes that this is a plain expression and passes it along to `treeify`.

Notice that `treeify` never has to worry about assignment, and can always treat seeing an `Equals` as an error, which is what we did above.

Now, we can update our `main()` function to use `build_statement()` and store our computed expressions in a `HashMap` for any assignments:

```rust
fn main() {
    // ...
    let mut ctx = HashMap::<String, Expr>::new();
    loop {
        // ...
        let res = tokenize(&mut buf.chars());
        let res = build_statement(&res)
            .expect("could not parse expr");
        match res {
            Statement::Assignment(v, e) => {
                let res = reduce_aoe(e);
                println!("-> {} = {}", v, res);
                ctx.insert(v, res);
            },
            Statement::Expr(e) => {
                let res = reduce_aoe(e);
                println!("-> {}", res);
            }
        }
    }
}
```

We're storing the expressions, but we don't have any way to look them up right now. For that, we need to tell `reduce_aoe()` about the context. The below code just looks up variables while reducing them rather than doing nothing with them.

```rust
fn reduce_aoe(e: Expr, ctx: &HashMap<String, Expr>) -> Expr {
    match e {
        Expr::App(l, r) => {
            let (l, r) = (reduce_aoe(*l, ctx), reduce_aoe(*r, ctx));
            match l {
                Expr::Abs(_, _) => reduce_aoe(beta_reduce(l, r), ctx),
                _ => Expr::App(Box::new(l), Box::new(r))
            }
        },
        Expr::Var(v) => {
            if let Some(exp) = ctx.get(&v) {
                exp.clone()
            }
            else {
                Expr::Var(v)
            }
        }
        _ => e
    }
}
```

And with that, our context is ready to use!

```
$ cargo run .
   Compiling lambda v0.1.0 (/home/nsood/lambda)
    Finished dev [unoptimized + debuginfo] target(s) in 0.41s
     Running `target/debug/lambda .`
one = \f.\x.f x
-> one = λf.λx.f x
one f x
-> f x
plus = \m.\n.\f.\x.m f (n f x)
-> plus = λm.λn.λf.λx.m f (n f x)
plus one
-> λn.λf.λx.(λf.λx.f x) f (n f x)
plus one one
-> λf.λx.(λf.λx.f x) f ((λf.λx.f x) f x)
plus one one f x
-> f (f x)
```

We can even do more complex things, like multiplication:

```
multiply = \m.\n.\f.\x.m (n f) x
-> multiply = λm.λn.λf.λx.m (n f) x
two = plus one one
-> two = λf.λx.(λf.λx.f x) f ((λf.λx.f x) f x)
multiply two two f x 
-> f (f (f (f x)))
```

And if expressions using the booleans we saw before:

```
true = \t.\f.t
-> true = λt.λf.t
false = \t.\f.f
-> false = λt.λf.f
if = \c.\t.\f.c t f
-> if = λc.λt.λf.c t f
if true one two f x
-> f x
if false one two f x
-> f (f x)
```

We can define zero and an abstraction to check if something is zero:

```
zero = \f.\x.x
-> zero = λf.λx.x
isZero = \n.n (\x.false) true
-> isZero = λn.n λx.false true
isZero one
-> λt.λf.f
isZero zero
-> λt.λf.t
(\n. if (isZero n) one n) two f x
-> f (f x)
(\n. if (isZero n) one n) zero f x
-> f x
```

`isZero` works because zero is defined as `\f.\x.x` -- notice that it ignores the first argument (f) passed to it completely -- and every other number is defined as `\f.\x.f (...)` -- notice that the first thing it does is call `f` with some argument. So `isZero` takes a number and passes an `f` that always returns false -- so anything that calls `f`, any nonzero number, will return false -- and passes an `x` that is true -- so anything that returns `x` without calling `f`, like zero, will return true.

Finally, we can define the "predecessor" abstraction -- the abstraction that takes N and produces N-1 (or produces zero if N is zero, because we haven't invented negative numbers yet):

```
pred = \n.\f.\x. n (\g.\h.h (g f)) (\u.x) (\u.u)
-> pred = λn.λf.λx.n λg.λh.h (g f) λu.x λu.u
pred one
-> λf.λx.(λf.λx.f x) λg.λh.h (g f) λu.x λu.u
pred one f x
-> x
pred two f x
-> f x
```

I'm not going to go over how `pred` works -- there's induction involved, and a sketch of the proof is available on the Wikipedia page for Chruch numerals I linked above. But now that we have `pred`, `isZero`, `multiply`, and `if`, we can do something interesting:

```
fact = \n.if (isZero n) one (multiply n (fact (pred n)))
-> fact = λn.if (isZero n) one (multiply n (fact (pred n)))
fact three f x
// program hangs until I terminate it
```

...or not. We had all the ingredients to make `fact`, the factorial abstraction, work... then why didn't it?

The answer lies in the function I suggestively named `reduce_aoe`, and in particular, in what `aoe` means. It stands for "applicative order evaluation", and means that before you evaluate an application, you evaluate both the left and right sides of the application -- kinda like evaluating the argument of an abstraction before applying it. That sounds alright, right? That's how most programming languages we're familiar with do it, and those can calculate factorials just fine. But it turns out to be very not-fine when all of our programming language constructs are _also_ abstractions, and very much rely on some arguments not being evaluated.

In the particular case of `fact`, something like this might happen:

* `fact` executes with `three`
    * `if` evaluates `(isZero three)`, which is a lambda X
    * `if` evaluates `one`, which is a lambda A
    * `if` evaluates `(multiply three (fact (pred three)))`
        * `multiply` evaluates `three`, which is a lambda B
        * `multiply` evaluates `(fact (pred three))`
            * `fact` evaluates `pred three`, which is a lambda C
            * `fact` executes with C
                * `if` evaluates `(isZero C)`
                * `if` evaluates `one`, which is a lambda D
                * `if` evaluates `(multiply C (fact (prec C)))`
                    * `multiply` evaluates `C`, which is a lambda E
                    * `multiply` evalutes `(fact (pred C))`
                        * `fact` evaluates `pred C`, which is a lambda F
                            * `fact` executes with F
                                * (...)

Note that in all of this evaluation, the first if expression has not even been evaluated yet -- it is still evaluating its arguments. In fact, it will _never_ evaluate, because as part of evaluating its arguments, it needs to evaluate a recursive call to `fact`, which needs to evaluate an if expression, which needs to evaluate a recursive call to `fact`, which needs to evaluate an if expression, which (...)

This is because of applicative-order evaluation! It's extremely important that some abstractions, like `if`, do _not_ evaluate their arguments until necessary. This is fine in normal programming languages because things like `if` are special-cased and don't behave like regular functions, but lambda calculus has no notion of `if` -- it's just an abstraction that we define.

We can solve this problem by using another reduction strategy, known as "normal order evaluation", or "lazy evaluation". Lazy evaluation works by, well, being lazy about when it evaluates the argument of an application. Specifically, it won't evaluate the arugment at the time of the application, and in fact will just substitute in an un-evaluated expression. Then, the substituted argument may be evaluated at some later point, only if it needs to be.

For example, consider the expression `(\x.y) ((\x.x x)(\x.x x))`. We have a lambda that ignores its argument, followed by its argument, which is an infinite loop -- beta-reducing that argument just yields that same thing again. Applicative-order evaluation will try to reduce the argument and get stuck in the infinite loop, while lazy evaluation will completely ignore the infinite loop and halt with the correct reduction, `y`.

We can update our reduction function to be lazy, and update `main()` accordingly:

```rust
fn reduce_lazy(e: Expr, ctx: &HashMap<String, Expr>) -> Expr {
    match e {
        Expr::App(l, r) => {
            let l = reduce_lazy(*l, ctx);
            match l {
                Expr::Abs(_, _) => reduce_lazy(beta_reduce(l, *r), ctx),
                _ => Expr::App(Box::new(l), Box::new(reduce_lazy(*r, ctx)))
            }
        },
        Expr::Var(v) => {
            if let Some(exp) = ctx.get(&v) {
                exp.clone()
            }
            else {
                Expr::Var(v)
            }
        }
        _ => e
    }
}
```

Notice that the only difference is that instead of recursively reducing both the left and right sides of an application, we now only reduce the left side at first, and then only reduce the right side if necessary. With that, we can see that `fact` now works as expected!

```
fact three f x
-> f (f (f (f (f (f x)))))
```

Counting carefully, those are in fact 6 = 3! `f`s! Our lambda calculus interpreter is now complete, and should be able to simplify any lambda calculus expressions you throw at it.

### The funny startup money people

If you know anything about lambda calculus, you may have heard about fixed-point combinators, or more specifically, the Y combinator. To motivate the existence of these things, consider again how I defined `isZero`:

```
isZero = \n.n (\x.false) true
```

I was helped by the fact that I could refer to `false` and `true`. But I didn't need that help. I could have defined `isZero` as follows:

```
isZero = \n.n (\x.\t.\f.f) \t.\f.t
```

That is, I could have defined `isZero` purely in terms of lambda calculus primitives -- abstractions, applications, and variables. Now consider how I defined `fact`:

```
fact = \n.if (isZero n) one (multiply n (fact (pred n)))
```

I can try to to the same thing here, resulting in this messy (and incomplete) attempt:

```
fact = \n.(\c.\t.\f.c t f) ((\n.n (\x.\t.\f.f) \t.\f.t) n) (\f.\x.f x) ((\m.\n.\f.\x.m (n f) x) n (fact ((\n.\f.\x. n (\g.\h.h (g f)) (\u.x) (\u.u)) n)))
```

Everything seems fine... except for that one lone `fact` sitting there innocently, in the definition of `fact`. What do we replace that with? It's part of the expression, and yet it needs to be replaced by the expression. If we replace that `fact`, we'll just have another `fact` we need to replace. It seems that we can't define `fact` purely in terms of lambda calculus primitives!

Or can't we?

That's the genius of fixed-point combinators -- it gives abstractions a way to refer to themselves, so they can use recursion. The most famous fixed-point combinator I'm aware of is the Y combinator, mostly because of the [startup incubator](https://www.ycombinator.com/) named after it.

The way the Y combinator works is, if you pass it an abstraction, it will apply that abstraction to itself, as many times as necessary. We can use this, by wrapping our factorial abstraction inside another abstraction, that receives a variable standing in for the thing we can use as a recursive call. To show this in `fact`'s old form:

```
fact_y = \f.\n.if (isZero n) one (multiply n (f (pred n)))
-> fact_y = λf.λn.if (isZero n) one (multiply n (f (pred n)))
```

It's that simple! Now, we can try to use our new abstraction with the Y combinator:

```
Y = \f.(\x.f (x x)) (\x.f (x x))
-> Y = λf.(λx.f (x x)) λx.f (x x)
Y fact_y three f x
-> f (f (f (f (f (f x)))))
```

So given our suitably-abstracted `fact_y`, we find that `Y fact_y` acts identically to our original self-referentially-defined `fact`! Now we can define `fact` as follows using the Y combinator:

```
fact = (\f.(\x.f (x x)) (\x.f (x x))) (\r.\n.(\c.\t.\f.c t f) ((\n.n (\x.\t.\f.f) \t.\f.t) n) (\f.\x.f x) ((\m.\n.\f.\x.m (n f) x) n (r ((\n.\f.\x. n (\g.\h.h (g f)) (\u.x) (\u.u)) n))))
```

Now we can test that this works by pasting this definition of `fact` in a fresh interpreter, with no context, and passing `\f.\x.f (f (f x))` (three) to it:

```
$ cargo run .
    Finished dev [unoptimized + debuginfo] target(s) in 0.03s
     Running `target/debug/lambda .`
(\f.(\x.f (x x)) (\x.f (x x))) (\r.\n.(\c.\t.\f.c t f) ((\n.n (\x.\t.\f.f) \t.\f.t) n) (\f.\x.f x) ((\m.\n.\f.\x.m (n f) x) n (r ((\n.\f.\x. n (\g.\h.h (g f)) (\u.x) (\u.u)) n)))) (\f.\x.f (f (f x))) f x
-> f (f (f (f (f (f x)))))
```

Phew!

### Conclusion

This was an incredibly fun project. At the time that I'm finishing writing this, it's been over a year and a half since I've delved deep into the implementation details of lambda calculus in a university lecture, and it's enjoyable to do so every time. I also got to write my own interpreter from scratch, parser and all, which isn't something I've done before. I also discovered that Rust, with its algebraic data type-style enums and pattern matching, was a great choice to build this in -- it felt familiar, unlike Haskell or OCaml, but still had most of the powerful expressiveness that functional languages provide.

The source code of the completed interpreter is available in whole at https://github.com/tendstofortytwo/lambda -- or you can assemble it yourself from the code snippets above.
