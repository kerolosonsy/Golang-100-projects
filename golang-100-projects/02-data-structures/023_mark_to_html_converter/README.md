# Project 023 — Markdown to HTML Converter (Tiny Subset)

## 1. Project Name and Number

Project **023** — `023_mark_to_html_converter`. The directory name and number must match exactly. This project builds a deterministic converter that turns a deliberately tiny subset of Markdown-flavoured text into HTML. The subset is not CommonMark and is not a roadmap toward CommonMark. It is a small, well-defined language whose rules the README pins, and the converter's behavior is fully pinned by those rules.

## 2. Project Idea

The converter reads text from an `io.Reader` and writes HTML to an `io.Writer`. The supported syntax covers six constructs only:

- ATX headings at levels 1, 2, and 3, written as `#`, `##`, `###` followed by exactly one ASCII space and at least one further character of content.
- Unordered list items beginning with `-` followed by exactly one ASCII space and at least one further character of content.
- Bold spans delimited by paired double asterisks on a single line.
- Plain paragraphs separated by blank lines.
- Plain text inside any of the above is escaped so that user-supplied HTML-significant characters become escaped literal text.
- Unsupported Markdown syntax and misnested constructs are emitted as ordinary paragraph text whose HTML-significant characters are escaped; asterisks stay as literal asterisks.

The output layout is pinned by this README: every non-empty emitted line ends with one newline, lists emit one line per opening/item/closing tag, multi-line paragraph source lines are joined into a single line with one ASCII space, and one or more consecutive blank input lines collapse to one blank output line between non-empty blocks. The converter is not a Markdown engine. It is a small, fully-tested translator for one well-defined subset.

## 3. Why This Project Now?

Projects 019 through 022 introduced streaming parsers, deterministic ordering, and validation discipline. Project 023 brings those disciplines together with a state machine: the converter must remember whether it is currently inside a list or a paragraph, must close open structures on the right line, and must close every open structure at end-of-input even when the input is malformed.

The project also introduces the discipline of escaping user content exactly once and only on user text. A naive converter either double-escapes (turning `&lt;` into `&amp;lt;`) or escapes too aggressively (turning `**` into `&#42;&#42;`). The README pins the processing order and the escape contract so the test can pin both. The project is a study in small languages, not a study in Markdown.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 023 therefore requires:

- Completion of **022** (Contact Book). Earlier projects (for example 021's streaming parsing and 019's bounded-scanner discipline) are background concepts already encountered and may inform the learner's design, but they are not additional required completions for this project.
- No prior knowledge of HTTP, databases, generics, or concurrency.

## 5. What You Must Know Before Starting

- That `bufio.Scanner` with the default line-terminator split emits one line per `Scan` call. The split strips a single trailing `\r` and a single trailing `\n`. A final line without a terminator is emitted normally.
- That `strings.Builder` accumulates output with low overhead and is the natural place to assemble the HTML.
- That HTML escaping has exactly five significant characters: `&`, `<`, `>`, `"`, and `'`. Their standard escape replacements are pinned by the HTML specification and are the same regardless of context. Asterisks (`*`), dashes (`-`), and other Markdown-flavoured characters are not HTML-significant and are not escaped.
- That a state machine for this language has a small number of states: outside any block, inside a paragraph, and inside a list. Paragraphs and lists cannot be open at the same time. A heading is a single-line state, not a wrapping state.
- That the converter's escape contract applies to user text only. The converter emits its own tags (`<h1>`, `<h2>`, `<h3>`, `<ul>`, `<li>`, `</li>`, `</ul>`, `<p>`, `</p>`, `<strong>`, `</strong>`) unescaped and writes them straight to the output. User text written between those tags is escaped exactly once.

## 6. Explanation of New Concepts

### The supported syntax

The converter recognizes exactly these constructs:

- **Level-1 heading.** A line whose first three characters are `#`, ` `, and at least one further character of content. The marker and the space are removed; the content is wrapped in `<h1>` and `</h1>`. Example: `# Title`.
- **Level-2 heading.** A line whose first four characters are `##`, ` `, and at least one further character of content. The markers and the space are removed; the content is wrapped in `<h2>` and `</h2>`. Example: `## Section`.
- **Level-3 heading.** A line whose first five characters are `###`, ` `, and at least one further character of content. The markers and the space are removed; the content is wrapped in `<h3>` and `</h3>`. Example: `### Subsection`.
- **Unordered list item.** A line whose first two characters are `-`, ` `, and at least one further character of content. The marker and the space are removed; the content is wrapped in `<li>` and `</li>`. Example: `- item`.
- **Bold span.** A pair of `**` markers on one line with non-empty content between them. The two markers are removed and the content is wrapped in `<strong>` and `</strong>`. Nested, misnested, empty-content, and unmatched pairs are not supported as bold spans.
- **Blank line.** A line that is empty or contains only whitespace. Blank lines separate paragraphs and lists.
- **Plain paragraph.** Any line that is not a heading, not a list item, and not blank. A paragraph is a run of one or more consecutive plain-paragraph lines.

A marker with no content is not a valid heading or list item. A line such as `# `, `## `, `### `, or `- ` (marker, space, then end-of-line) is unsupported syntax and is treated as a plain-paragraph line: the line is emitted as paragraph text, with HTML-significant characters escaped. The marker characters remain literal because asterisks, hashes, and dashes are not HTML-significant; the leading space and any other HTML-significant characters in the line are escaped normally.

Anything else — fenced code blocks, block quotes, ordered lists, links, images, raw HTML, tables, hard line breaks, reference-style links, horizontal rules, character escapes, indented code, four-space code blocks, nested lists, four-or-more-hash headings — is out of scope. The converter does not recognize these constructs. Their syntax characters are part of the line's text content; any HTML-significant characters among them are escaped, and the rest stays literal.

### Processing order

The converter processes input line by line. For each line, in this order:

1. Determine the line's block-level role: heading (level 1, 2, or 3), list item, blank, or paragraph text. The role is determined by the line's first characters and the current state.
2. Close any block-level structure that the new role requires closing (for example, closing a paragraph when a list item begins, or closing a list when a paragraph begins).
3. Emit the block-level opening tag (`<h1>`, `<h2>`, `<h3>`, `<ul>`, `<li>`, `<p>`) when the role requires it.
4. Process the line's inline content: detect matched bold spans and emit the line's text, applying HTML escaping to user text exactly once. The converter's own tags (`<strong>`, `</strong>`) are emitted unescaped.
5. Emit the block-level closing tag (`</h1>`, `</h2>`, `</h3>`, `</li>`, `</ul>`, `</p>`) at the right moment.

The crucial pinning: HTML escaping applies only to user text, and only once. The converter emits its own tags unescaped. Bold-span detection runs on raw text; the resulting content (whether wrapped in `<strong>` or emitted as literal text) is HTML-escaped exactly once before being written. A user-supplied `<` becomes `&lt;` once; it never becomes `&amp;lt;`. A user-supplied `&` becomes `&amp;` once. Asterisks remain asterisks.

### The block state machine

The state machine has three wrapping states:

- **Outside.** No block is open. The next non-blank line starts a new block.
- **In paragraph.** A `<p>` is open. A blank line, a heading, or a list item closes the paragraph. A non-blank, non-heading, non-list-item line continues the paragraph.
- **In list.** A `<ul>` is open. A list item continues the list; a blank line, a heading, or a paragraph line closes the list.

A paragraph and a list are never open at the same time. The state transitions are deterministic and pinned.

A heading is a single-line state, not a wrapping state. After the heading's closing tag is emitted, the state returns to `Outside`. A heading interrupts an open paragraph or list: the converter closes the open structure before emitting the heading's tag.

Two adjacent list items form one list. A blank or non-list line closes the list. A new list item after a closing non-list line opens a new list. Three consecutive list items, then a blank line, then two more list items, produce two `<ul>` elements with two separate opening and closing pairs.

### Bold-span detection rules

Bold spans are detected per line. The detection rule is:

- A line is a valid bold span if and only if that source line contains exactly one non-empty paired `**` arrangement: the line contains exactly two `**` markers, and the content between them is non-empty. In that case the two markers are removed and the matched content is wrapped in `<strong>` and `</strong>`. The matched content is HTML-escaped exactly once.
- Any other `**` arrangement on a source line is unsupported: the line is emitted as paragraph text and the `**` characters remain literal text. Asterisks are not HTML-significant and are not escaped. Specifically, the following are unsupported and produce literal text:
  - An unmatched `**` (one marker with no second marker on the same line).
  - An empty-content pair such as `****` or `** **` (two markers with nothing or whitespace between them).
  - A misnested arrangement such as `**bold **inside**` (three or more markers, with no single non-empty pair that exhausts the markers).
  - A multiple-pair arrangement such as `**a** and **b**` (two non-overlapping matched pairs on the same line).
  - Any other arrangement that does not satisfy the "exactly one non-empty paired delimiter span" rule.
- The converter does not guess. It does not patch a missing closing marker, does not open a span without a paired close, and does not re-scan the matched content for further bold spans.

The HTML-significant characters around the literal `**` markers — for example an opening `<` or a closing `>` — are escaped once per the standard escape contract. The asterisks themselves remain literal.

### Newline behavior

The output layout is pinned and applies to every run:

- **Empty input.** Zero bytes. The output is zero bytes. Nothing is emitted.
- **Every non-empty emitted block or tag line ends with exactly one newline (`\n`).** A heading emits one line. A list emits one line for `<ul>`, one line per item (`<li>...</li>`), and one line for `</ul>`. A paragraph emits one line for `<p>...</p>`. The presence or absence of a trailing terminator on the input's final line does not change the trailing-newline behavior of the output: every non-empty emitted line ends with `\n`, including the last.
- **Multi-line paragraph source lines join with exactly one ASCII space** into a single paragraph line. A paragraph that spans two source lines becomes one `<p>...</p>` line whose content reflects that join. A blank line ends the paragraph.
- **One or more consecutive blank input lines collapse to exactly one blank output line** between non-empty blocks. The blank output line is a single `\n` that separates the previous block from the next.
- **A list with two items produces four lines** in this order: `<ul>`, `<li>first</li>`, `<li>second</li>`, `</ul>`.
- **A heading produces one line.** A level-1 heading line `# Title` becomes `<h1>Title</h1>`.

The output is byte-identical across two runs against the same input.

### The escape contract

The escape contract is pinned to the standard five-character set with the standard replacements. The exact escape strings are part of the test contract:

- `&` becomes `&amp;`.
- `<` becomes `&lt;`.
- `>` becomes `&gt;`.
- `"` becomes `&#34;`.
- `'` becomes `&#39;`.

The escape is applied to user text only, exactly once. The converter's own tags are emitted unescaped. Asterisks, hashes, dashes, and other Markdown-flavoured characters are not HTML-significant and are not escaped: they remain literal.

### Unsupported Markdown stays as escaped literal text

A line beginning with a backtick fence, a line beginning with `>`, a line beginning with `1.`, a line containing `[link](url)`, a line containing `![image](url)`, a four-space-indented line, and any other Markdown syntax the converter does not recognize: the syntax characters are part of the line's text content. HTML-significant characters in that text are escaped; asterisks, hashes, and dashes remain literal. A raw `<script>alert(1)</script>` becomes `&lt;script&gt;alert(1)&lt;/script&gt;`. The text never becomes live HTML.

## 7. Learning Objective

After completing this project the learner can:

- Define a small block-level language with a state machine and pin every transition in writing.
- Distinguish block-level structure (headings, lists, paragraphs) from inline structure (bold spans) and process them in the correct order.
- Apply HTML escaping to user text only, exactly once, with the pinned five-character escape contract; the converter's own tags are emitted unescaped.
- Emit deterministic output with the pinned newline behavior: every non-empty emitted line ends with one newline, lists emit one line per tag, paragraph source lines join with one ASCII space, and consecutive blank input lines collapse to one blank output line.
- Recognize the "exactly one non-empty paired delimiter span" rule for bold spans and refuse to guess a partial span when a line contains unmatched, empty, misnested, or multiple `**` arrangements.
- Write tests that pin every supported construct, every state transition, and every escape case including malicious raw HTML.

## 8. Functional Requirements

1. The converter reads from an `io.Reader` and writes to an `io.Writer`. Production wires a file or standard input; tests wire a `strings.Reader` and a `bytes.Buffer`.
2. The supported block-level constructs are: `#` level-1 headings with content, `##` level-2 headings with content, `###` level-3 headings with content, `- ` list items with content, blank lines, and plain paragraphs. A marker with no content (`# `, `## `, `### `, `- `) is unsupported syntax and is treated as a plain-paragraph line.
3. A line is a level-1 heading if and only if its first three characters are `#`, ` `, and at least one further character of content. The hash and the space are removed; the content is wrapped in `<h1>` and `</h1>` on one emitted line.
4. A line is a level-2 heading if and only if its first four characters are `##`, ` `, and at least one further character of content. The hashes and the space are removed; the content is wrapped in `<h2>` and `</h2>` on one emitted line.
5. A line is a level-3 heading if and only if its first five characters are `###`, ` `, and at least one further character of content. The hashes and the space are removed; the content is wrapped in `<h3>` and `</h3>` on one emitted line.
6. A line is a list item if and only if its first two characters are `-`, ` `, and at least one further character of content. The `-` and the space are removed; the content is wrapped in `<li>` and `</li>` on one emitted line.
7. A blank line (empty or whitespace-only) separates paragraphs and lists. One or more consecutive blank input lines collapse to one blank output line between non-empty blocks. Blank lines never appear inside a heading or a list item.
8. A paragraph is a run of one or more consecutive non-blank, non-heading, non-list-item lines. The lines are joined with exactly one ASCII space, and the resulting text is wrapped in `<p>` and `</p>` on one emitted line.
9. Two or more consecutive list items form one list. The list emits one `<ul>` line, then one `<li>...</li>` line per item, then one `</ul>` line. A blank or non-list line between two list items closes the previous list and a subsequent list item opens a new list.
10. Headings interrupt paragraphs and lists. A heading closes any open paragraph or list before emitting its own tag.
11. Paragraphs and lists are never open at the same time.
12. The converter closes every open block-level structure at end-of-input.
13. Bold spans are detected per line using `**` markers. A line is a valid bold span if and only if it contains exactly one non-empty paired `**` arrangement on that source line. In that case the matched content is wrapped in `<strong>` and `</strong>`. Any other arrangement — unmatched, empty-content, misnested, multiple-pair, or otherwise unsupported — remains literal text; no `<strong>` is emitted for it. The converter does not guess.
14. HTML escaping applies only to user text and is applied exactly once. The escaped characters and their replacements are pinned: `&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`, `"` → `&#34;`, `'` → `&#39;`. The converter's own tags are emitted unescaped. Asterisks, hashes, dashes, and other non-HTML-significant characters remain literal.
15. The output layout is pinned: every non-empty emitted line ends with exactly one newline; lists emit one line per opening, item, and closing tag; paragraph source lines join with one ASCII space; consecutive blank input lines collapse to one blank output line.
16. The output is byte-identical across two runs against the same input.

## 9. Inputs and Outputs

### Inputs

- A stream of UTF-8 text through an `io.Reader`. The text may contain any combination of supported syntax, plain text, raw HTML, Markdown syntax the converter does not support, blank lines, and trailing or missing newlines.

### Outputs

- HTML written to an injected `io.Writer`. The HTML contains only the tags listed in section 8. Raw HTML from the input never appears as live HTML in the output. Output is deterministic.

### Example text-only success run

Input:
```
# Title

Hello **world**.

- one
- two

Second paragraph.
```

Output:
```
<h1>Title</h1>
<p>Hello <strong>world</strong>.</p>
<ul>
<li>one</li>
<li>two</li>
</ul>
<p>Second paragraph.</p>
```

### Example text-only escaped-raw-HTML run

Input:
```
<script>alert(1)</script>
```

Output:
```
<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>
```

### Example text-only unmatched-bold run

Input:
```
This **has no closing marker.
```

Output:
```
<p>This **has no closing marker.</p>
```

(The unmatched `**` is literal text and is not escaped, because asterisks are not HTML-significant. No `<strong>` is emitted.)

## 10. Rules and Edge Cases

- **Empty input.** Zero bytes. The output is zero bytes. Nothing is emitted.
- **Whitespace-only input.** All whitespace. Treated as blank lines. The output is zero bytes.
- **Single line, no markup.** A line of plain text becomes one `<p>...</p>` line. The text is HTML-escaped.
- **Heading followed by paragraph.** A heading line followed by a paragraph line produces a heading line and a paragraph line in that order, with one blank output line between them when the input had a blank line.
- **List followed by paragraph.** A list line followed by a paragraph line produces the list block and the paragraph block, with the list closed before the paragraph begins. The list emits `<ul>`, each item, and `</ul>` on separate lines; the paragraph emits one `<p>...</p>` line.
- **Paragraph followed by list.** A paragraph line followed by a list item line produces the paragraph block, then the list block. The paragraph emits one `<p>...</p>` line; the list emits `<ul>`, each item, and `</ul>` on separate lines.
- **Adjacent lists with blank between.** Two lists separated by a blank input line produce two separate `<ul>` blocks. The blank line closes the first list; the next list item opens a new list. The output has two `<ul>` blocks with one blank output line between them.
- **List closed at EOF.** A list whose last item is the final input line is closed with `</ul>` at end-of-input. The output ends with the `</ul>` line and the trailing newline.
- **Paragraph closed at EOF.** A paragraph whose last line is the final input line is closed with `</p>` at end-of-input. The output ends with the `</p>` line and the trailing newline.
- **Marker with no content.** A line such as `# `, `## `, `### `, or `- ` (marker, space, then end-of-line) is unsupported syntax and becomes a plain-paragraph line containing the literal marker characters. The leading space is part of the text; HTML-significant characters in the line (if any) are escaped normally.
- **Bold span with non-empty content.** `**word**` on a single line produces `<strong>word</strong>` within its block. HTML-significant characters inside the matched content are escaped once.
- **Bold span with empty content.** `****` or `** **` on a single line is not a bold span. The `**` characters remain literal text.
- **Unmatched bold.** A line with one `**` and no closing `**` is paragraph text. The `**` remains literal. No `<strong>` is emitted.
- **Misnested bold.** A line with three or more `**` markers such as `**bold **inside**` is not a bold span. The line is paragraph text. All `**` characters remain literal. No `<strong>` is emitted. The converter does not guess a partial span.
- **Multiple bold spans per line.** A line with two non-overlapping matched pairs, for example `**a** and **b**`, is not a bold span under the project's "exactly one non-empty paired delimiter span" rule. The line is paragraph text. The `**` characters remain literal. No `<strong>` is emitted.
- **Raw HTML.** A line containing `<`, `>`, `&`, `"`, or `'` inside any block has those characters escaped to `&lt;`, `&gt;`, `&amp;`, `&#34;`, `&#39;` respectively. Raw `<script>`, raw `<img>`, raw `<a href="...">` all become escaped literal text.
- **Ampersand first.** When a user `&` appears next to a sequence that looks like an entity (for example `&lt;`), the `&` is escaped first to `&amp;`, producing `&amp;lt;`. The double-escape trap (`&amp;amp;`) does not occur because the second pass never runs.
- **Unsupported Markdown.** A line beginning with `>`, `1.`, four spaces of indentation, or a fenced code block (```` ``` ````) is treated as paragraph text. The syntax characters are part of the text; HTML-significant characters among them are escaped.
- **Trailing newline in input.** A final newline does not change the output's trailing-newline behavior. Every non-empty emitted line ends with one newline.
- **No trailing newline in input.** A final line without a terminator is processed normally. The output still ends with one newline on the last emitted line.
- **Carriage returns.** A `\r\n` line ending is normalized by the scanner; the field contents are unaffected.

## 11. Project Constraints

- Go standard library only. No third-party Markdown libraries, no HTML libraries beyond `html.EscapeString` if the learner chooses to use it, no regular expressions.
- The supported syntax is the small set in section 8. A marker with no content is unsupported and becomes plain-paragraph text. No other Markdown syntax is recognized.
- The escape contract is the standard five-character HTML escape with the pinned replacements (`&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`, `"` → `&#34;`, `'` → `&#39;`). The escape applies to user text only and is applied exactly once. The converter's own tags are emitted unescaped.
- Asterisks, hashes, dashes, and other non-HTML-significant characters are not escaped. They remain literal text.
- The output layout is pinned: every non-empty emitted line ends with exactly one newline; lists emit one line per opening, item, and closing tag; paragraph source lines join with one ASCII space; consecutive blank input lines collapse to one blank output line.
- The output is deterministic across runs on the same machine.
- The converter does not load the whole input into memory before processing. A test that drives the converter through a custom incremental `io.Reader` confirms streaming behavior.
- The output is written to an injected `io.Writer`. The converter does not write to standard output or standard error directly.

## 12. Design Questions Before Coding

- Where does the block-level state machine live? As a small type inside the converter package, as a function with explicit state parameters, or as a set of mutually recursive functions? Which choice keeps every transition visible in one place?
- How is the role of a line determined? Through a sequence of `HasPrefix` checks, through a small classifier function, or through a switch? Which choice keeps the priority (heading before list before paragraph) obvious?
- How is HTML escaping applied? Through a wrapper that escapes every piece of user text written, through a final pass that re-escapes the whole output (forbidden), or through disciplined escaping at every emission point? Which choice keeps "exactly once on user text only" easy to test?
- How is the closing of structures managed? Through a deferred end-of-input pass, through a `closeAll` helper called at every transition, or through a stack of open structures? Which choice matches the small state space here?
- How is the output layout pinned? Through constants for tag placement, through a build helper, or through a single format function? Which choice makes the test's expected output stable across runs?
- How is the streaming test built? Through a custom `io.Reader` that yields the input across many small `Read` calls, or through a `bytes.Reader`? Which choice confirms that the converter never reads everything into memory before emitting HTML?

## 13. Implementation Milestones

1. Decide the package layout. Keep `main` as a thin wrapper that opens the file and writes the output. The converter package owns the state machine, the inline processing, and the escaping.
2. Pin the supported syntax and the unsupported-syntax rule (marker with no content becomes paragraph text) as named rules. Pin the HTML escape replacements as named constants matching the standard five-character set.
3. Build a `bufio.Scanner` over the reader. Configure the buffer with explicit headroom before any `Scan` call.
4. Iterate lines. For each line, determine its block-level role and update the state machine: close any open structure the new role requires closing, then open the new structure.
5. Emit the block-level opening tag (`<h1>`, `<h2>`, `<h3>`, `<ul>`, `<li>`, `<p>`) on its own line when the role requires it.
6. Process the line's inline content: detect matched bold spans, emit the line's text and the `<strong>`/`</strong>` tags, and apply HTML escaping to user text exactly once. The converter's own tags are emitted unescaped.
7. Emit the block-level closing tag (`</h1>`, `</h2>`, `</h3>`, `</li>`, `</ul>`, `</p>`) on its own line at the right moment. Lists emit one line per opening, item, and closing tag.
8. Implement the end-of-input pass. After the scanner finishes, close every still-open structure. The final line of the output ends with one newline.
9. Wire `main`. Read from a file path (or standard input), run the converter, write to standard output or a file path. Errors go to standard error.
10. Add tests for every verification case in section 14, including the malicious-raw-HTML case, the unmatched-bold case, the list-closed-at-EOF case, the streaming case, and the determinism case.

## 14. Verification Cases the Learner Must Write

Each case is described in natural language. Tests drive the converter through an `io.Reader` and assert on the `io.Writer` output.

### Supported constructs

- A level-1 heading line becomes `<h1>title</h1>` followed by a newline, with the content HTML-escaped.
- A level-2 heading line becomes `<h2>title</h2>` followed by a newline.
- A level-3 heading line becomes `<h3>title</h3>` followed by a newline.
- A list of two items produces four lines: `<ul>`, `<li>first</li>`, `<li>second</li>`, `</ul>`. Each line ends with a newline.
- A single paragraph line becomes `<p>text</p>` followed by a newline, with the text HTML-escaped.
- A bold span with non-empty content (`**word**`) becomes `<strong>word</strong>` within its block, with the matched content HTML-escaped.
- A bold span with HTML-escaped content (`**<x>**`) becomes `<strong>&lt;x&gt;</strong>` — the `**` are recognized as markers, the `<` and `>` are escaped once.
- Two non-overlapping matched pairs on the same line (for example `**a** and **b**`) is not a bold span under the project's "exactly one non-empty paired delimiter span" rule. The line is paragraph text. The `**` characters remain literal. No `<strong>` is emitted.

### Marker with no content

- A line containing `# ` (hash, space, end-of-line) is unsupported syntax and becomes a paragraph line containing the literal `# ` and any HTML-significant characters escaped. The leading space is part of the text.
- A line containing `## ` is treated the same way: literal `## ` text in a paragraph.
- A line containing `### ` is treated the same way: literal `### ` text in a paragraph.
- A line containing `- ` is treated the same way: literal `- ` text in a paragraph.

### Adjacent transitions

- A paragraph line followed by a heading line produces a paragraph line, one blank output line, then the heading line. The paragraph's `</p>` and the heading's `<h1>` are emitted in that order.
- A paragraph line followed by a list item line produces a paragraph line, one blank output line, then the list block (`<ul>`, items, `</ul>`). The paragraph's `</p>` and the list's `<ul>` are emitted in that order.
- A list item line followed by a paragraph line produces the list block, one blank output line, then the paragraph line. The list's `</ul>` and the paragraph's `<p>` are emitted in that order.
- A heading line followed by a list item line produces the heading line, one blank output line, then the list block.
- Three consecutive list items form one list. A blank input line between two list items closes the first list; the next list item opens a new list. The output has two `<ul>` blocks with one blank output line between them.

### Blank and empty input

- Zero bytes of input produces zero bytes of output.
- Whitespace-only input produces zero bytes of output.
- A single blank input line between two paragraphs produces one blank output line between the two `<p>...</p>` lines.
- Two or more consecutive blank input lines between two paragraphs produce one blank output line between them (the consecutive blank lines collapse).

### Plain text

- A line of plain text with no markup becomes one `<p>...</p>` line. The text is HTML-escaped.
- A multi-line paragraph (lines joined without blank input lines between them) becomes one `<p>...</p>` line whose content joins the source lines with one ASCII space.

### Malicious raw HTML

- A line containing `<script>alert(1)</script>` becomes a paragraph line with `<` escaped to `&lt;` and `>` escaped to `&gt;`. The output never contains a live `<script>` tag.
- A line containing the four characters `&lt;` (that is, `&`, `l`, `t`, `;`) becomes a paragraph line with `&` escaped to `&amp;`, producing `&amp;lt;`. The double-escape trap (`&amp;amp;`) does not occur.
- A line containing an attribute-quoted string like `" onerror="alert(1)` becomes paragraph text with `"` escaped to `&#34;` and `<` (if any) escaped to `&lt;`.

### Ampersands and quotes (pinned exact entities)

- An ampersand in plain text is escaped to `&amp;`. The double-escaping trap (`&amp;amp;`) does not occur.
- A `<` in plain text is escaped to `&lt;`.
- A `>` in plain text is escaped to `&gt;`.
- A double quote in plain text is escaped to `&#34;`.
- A single quote in plain text is escaped to `&#39;`.

### Unmatched bold

- A line with one `**` and no closing `**` is paragraph text. The `**` remains literal. No `<strong>` is emitted.
- A line with `****` (two adjacent pairs, no content) is paragraph text. The `****` remains literal.
- A line with `** ` (marker followed by only whitespace before end-of-line) is paragraph text. No `<strong>` is emitted.
- A line `**bold **inside**` (three or more markers, no single non-empty pair that exhausts the markers) is paragraph text. All `**` characters remain literal. The converter does not guess a partial span. No `<strong>` is emitted.
- A line `**a** and **b**` (two non-overlapping matched pairs) is paragraph text under the "exactly one non-empty paired delimiter span" rule. The `**` characters remain literal. No `<strong>` is emitted.

### Unsupported syntax

- A line beginning with `>` (blockquote syntax) is paragraph text. The `>` remains literal.
- A line beginning with `1. ` (ordered list syntax) is paragraph text. The `1.` remains literal.
- A line beginning with four spaces (indented code block) is paragraph text. The leading spaces are part of the text.
- A fenced code block delimited by ```` ``` ```` is paragraph text on every line.
- A line containing `[link](url)` is paragraph text. The brackets and parentheses remain literal.

### List closed at EOF

- A list whose last item is the final input line is closed with `</ul>` on its own line at end-of-input.
- A paragraph whose last line is the final input line is closed with `</p>` on its own line at end-of-input.
- A paragraph and a list are never open at the same time at end-of-input.

### Deterministic output

- Two runs against the same input produce byte-identical output.
- The newline layout, the placement of tags, and the per-line ordering are all pinned by the test.

### Process

- An integration test runs the compiled binary against a temporary file with a mix of supported and unsupported syntax and confirms the exit code is zero and the output is on standard output.
- An integration test runs the compiled binary against a temporary file containing raw HTML and confirms the output contains escaped literal text and no live HTML tags.

## 15. Common Mistakes to Watch For

- **Double-escaping.** Escaping the entire output again after constructing it turns `&lt;` into `&amp;lt;`. User text is escaped once, at the right moment; the converter's own tags are never escaped.
- **Forgetting to escape.** Emitting raw user text without escaping turns `<script>` into live script tags. Every piece of user text is escaped, always.
- **Escaping the converter's own tags.** Tags like `<h1>`, `<p>`, `<strong>` are emitted unescaped. Escaping them produces `&lt;h1&gt;` instead of `<h1>`, which is wrong.
- **Escaping asterisks.** Asterisks are not HTML-significant. Treating `**` as if it needed escaping produces `&#42;&#42;` and breaks the literal-text contract for unmatched and misnested markers.
- **Treating unmatched `**` as a bold opening.** A line with one `**` and no closing marker is paragraph text. The marker is literal; no `<strong>` is emitted.
- **Treating empty-content `**` as a bold span.** `****` and `** **` are not bold spans. The markers remain literal.
- **Guessing a bold span from misnested or multiple markers.** The "exactly one non-empty paired delimiter span" rule means a line with three or more `**` markers, or with two non-overlapping matched pairs, is paragraph text. The converter does not guess a partial span.
- **Re-scanning bold-span content for nested bold.** Nested bold is not supported. The text inside a matched pair is processed for HTML escaping and wrapped in `<strong>`; it is not re-scanned for further bold markers.
- **Forgetting to close open structures at EOF.** An open paragraph or list at end-of-input must be closed by an explicit end-of-input pass.
- **Joining all lists into one `<ul>`.** Two lists separated by a blank input line are two `<ul>` blocks, not one. The first list is closed when the blank line is processed.
- **Treating a marker with no content as a valid empty heading or list item.** `# `, `## `, `### `, and `- ` are unsupported syntax and become plain-paragraph text containing the literal marker characters.
- **Outputting unescaped user content under a heading.** Heading content is also user text and must be HTML-escaped.
- **Outputting non-deterministic ordering.** Tag placement, newline layout, and per-line order must be deterministic. The test pins the layout.
- **Loading the whole input into memory before emitting output.** The streaming test confirms correctness across many small `Read` calls.
- **Writing the output to standard output from inside the converter.** The converter writes to an injected `io.Writer`. Hard-coding `os.Stdout` makes the converter un-testable.

## 16. Topics and References for Study

- A Tour of Go: "Methods", "Errors", "Reading files".
- Effective Go: "Data", "Errors".
- Package documentation: `bufio` (`Scanner`, `Scanner.Buffer`, `Scanner.Scan`, `Scanner.Err`, `ScanLines`), `strings` (`Builder`, `HasPrefix`, `TrimSpace`), `io` (`Reader`, `Writer`), `html` (`EscapeString`).
- HTML escape contract: search for "HTML entity escape", "OWASP HTML escaping", "Go html.EscapeString".
- Markdown subset design: search for "CommonMark vs subset", "tiny Markdown converter", "Markdown subset state machine".

## 17. Self-Assessment Questions

1. Why is the supported syntax a tiny subset pinned by the README before the converter is written, instead of grown toward CommonMark during implementation?
2. Why is HTML escaping applied to user text only, exactly once, with the pinned five-character set?
3. Why does unmatched `**` remain literal text instead of being escaped or guessed as half of a bold span?
4. Why are two lists separated by a blank input line emitted as two `<ul>` blocks, not one?
5. Why is a marker with no content treated as paragraph text instead of as a valid empty heading or list item?
6. Why is the state machine designed so paragraphs and lists are never open at the same time?
7. Why must every open block-level structure be closed at end-of-input, and what would the test pin if the converter forgot?
8. Why does raw `<script>` in the input become escaped literal text, never live HTML?
9. Why must the output layout be pinned to specific line-end and tag-placement rules instead of left as learner choice?
10. Why must the streaming test use a custom incremental `io.Reader` rather than a `bytes.Reader`, and what would a `bytes.Reader` fail to prove?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- The README's 19 sections are present in order; this file is the reference.
- Every functional requirement in section 8 is satisfied.
- Every verification case in section 14 has a corresponding test.
- The state machine has the three wrapping states described in section 6, with the transitions documented and pinned by tests. Paragraphs and lists are never open at the same time.
- HTML escaping applies to user text only and is applied exactly once. The escape contract is `&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`, `"` → `&#34;`, `'` → `&#39;`. The converter's own tags are emitted unescaped. Asterisks, hashes, and dashes remain literal.
- A marker with no content (`# `, `## `, `### `, `- `) is unsupported syntax and becomes a paragraph line containing the literal marker characters.
- Unmatched and misnested `**` remain literal text; no `<strong>` is emitted for them.
- Every open block-level structure is closed at end-of-input.
- Raw HTML in the input never appears as live HTML in the output.
- The output layout is pinned: every non-empty emitted line ends with one newline; lists emit one line per opening, item, and closing tag; paragraph source lines join with one ASCII space; consecutive blank input lines collapse to one blank output line.
- Two runs against the same input produce byte-identical output.
- The package documentation states the supported syntax, the unsupported-syntax rule, the escape contract, the newline behavior, and the inline-processing rules.
- The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Italic spans.** Add a single-asterisk italic syntax (`*text*`) using the same per-line pairing rule as bold. Italic and bold do not nest; bold markers inside italic text remain literal text. Do not add underscore italic, three-asterisk bold-italic, or any other inline construct.
- **Ordered list items.** Add a `1. ` list marker that produces `<ol>`, `<li>...</li>` lines, and `</ol>` on separate lines instead of `<ul>` and `</ul>`. The number is fixed at `1.`; the converter does not renumber. Do not add other ordered-list markers, nesting, or a start-attribute.
