# Project 013 — Time World Clock

## 1. Project Name and Number

- Project **013** — `013_time_world_clock`.
- The directory name and number must match exactly.

## 2. Project Idea

A small command-line tool that takes one fixed instant in time and the names of two IANA time zones, and shows what that same instant looks like in each of those zones. The instant is the input; the zones are how that instant is *read* for humans in different parts of the world. The program must handle a zone that the operating system does not know by reporting the unknown name and producing no conversion result, and it must surface the case where the same instant falls on a different calendar day in one of the zones.

The interesting part is the distinction between the *instant* (a single point on the timeline that does not change) and the *local representation* (the date and clock reading on a wall in a particular zone, which does change with the zone). Confusing those two is one of the most common bugs in time-handling code, so this project makes the distinction visible in both the documentation and the behavior.

## 3. Why This Project Now?

- Projects 011 and 012 introduced injected I/O, small domain logic with a declared policy, and integer-cents arithmetic.
- Project 013 keeps those habits and adds a new one: the program's behavior must be **deterministic**, even though "time" feels like it depends on when the program runs.
- The instant is provided by the caller — typed, or fixed in a test — so the answer is the same whether the program runs at noon or at midnight.

- This project is also the first time the learner meets a data source outside the program: the IANA time zone database that the operating system provides.
- The contract is "ask for a zone by name, get back its rules or an error"; the program must not assume a zone exists just because a string was typed.

## 4. Prerequisites

Per the dependency map in `plan.md`, projects 002 through 030 require only the immediately previous project. Project 013 therefore requires:

- Completion of **012** (Bill Splitter).
- No prior knowledge of HTTP, databases, or concurrency.

## 5. What You Must Know Before Starting

- The difference between an *instant* (a single point on the timeline, such as Unix epoch milliseconds) and a *local representation* (the date and clock reading on a wall).
- What an IANA time zone name is (for example `Europe/Berlin`, `Asia/Tokyo`, `America/New_York`). It is a string the operating system resolves to a set of historical and current rules about UTC offset and daylight-saving transitions.
- That the same instant can correspond to different dates in different zones, including cases where one zone shows "yesterday" and another shows "today".
- How Go's `time.Time` carries both an instant and a location, and how `In(loc)` produces a new value with the same instant but a different wall representation.
- The role of `time.LoadLocation` and the difference between a successful location and the error it returns.
- That the machine's local zone is set by the operating system; the program must not rely on it.
- The difference between `time.Time` equality via `==` and equality via `Time.Equal`. `==` compares the full struct value, including the location pointer and monotonic-clock data; `Time.Equal` asks whether two values represent the same instant. Two values can therefore be `Time.Equal` after one is rendered in another location while still not being equal with `==`.

## 6. Explanation of New Concepts

### Concepts

#### Instant versus local representation

- An instant is a single point on the timeline.
- It does not have a date or a clock reading; those are properties of how a particular zone chooses to *display* the instant.
- Two zones display the same instant with different wall readings; one zone can show "Tuesday 23:00" while another shows "Wednesday 05:00" for the same instant because they are far apart across the date line.

- A program that wants to "convert a time between zones" is really doing two things: keep the instant unchanged, and change the zone used to render it.
- The instant must not be touched by the conversion.

#### IANA time zone names

- An IANA name like `Europe/Berlin` is not a fixed UTC offset.
- It is a key into a database of rules that change over time: daylight-saving transitions, historical adjustments, and political changes. `time.LoadLocation` looks up that key on the operating system.
- If the key is unknown, it returns an error and the program must not pretend it succeeded.

#### Unknown zone handling

- A typo, a fake zone, or a zone the operating system does not carry produces an error from `time.LoadLocation`.
- The contract for this project is straightforward: **if either the source zone or the target zone is unknown, the program reports the unknown name, produces no conversion result, and exits cleanly.** The source rendering is *not* printed in that case.
- The program does not panic, does not fall back to UTC silently, and does not invent a zone.

#### Day-boundary case

- The offsets between zones can place the same instant on adjacent calendar dates.
- For example, `Asia/Tokyo` is far enough ahead of `America/Los_Angeles` that a late-evening instant in Los Angeles is already the next day in Tokyo.
- The program must surface that, not hide it: when the calendar day differs between the two zones, the output for each zone includes the date, not just the clock.

#### Calendar validation and DST scope

- Go can normalize out-of-range calendar fields instead of rejecting them.
- For example, an extra day can roll into the next month.
- This program must reject malformed and impossible calendar inputs, including dates such as `2025-02-30`, rather than silently converting a different date from the one the user supplied.

- Daylight-saving gaps and repeated wall times are important, but resolving their ambiguity is outside this project's required scope.
- Required examples and tests must use fixed instants away from DST transitions.
- The learner should still know that constructing a local wall time during a transition can select an offset in a way the application did not intend; production scheduling software needs an explicit policy for that case.

#### Determinism and the local zone

- The machine's local zone is whatever the host operating system was configured with — it can be UTC on a server, `Europe/Berlin` on a laptop, or anything else.
- A test that compares a `time.Time` against the machine's local zone is fragile: the same test passes on one machine and fails on another.
- The tests for this project must construct instants and zones explicitly; they must not call `time.Now` and they must not rely on the machine's local zone.

## 7. Learning Objective

After completing this project the learner can:

- Distinguish, in writing and in code, between an instant and a local representation.
- Construct a `time.Time` for a fixed instant from explicit year, month, day, hour, minute, second, and zone inputs, after validating the calendar fields.
- Use `time.LoadLocation` and handle the error path when the zone is unknown.
- Render the same instant in two different zones and explain why the wall readings differ.
- Recognize when a single instant produces different calendar days in the chosen zones.
- Write tests that pass on any machine because they do not depend on `time.Now` or on the host's local zone.
- Explain the difference between `time.Time` equality with `==` and equality with `Time.Equal`, and pick the right one for a given comparison.

## 8. Functional Requirements

1. Read a fixed instant: year, month, day, hour, minute, and second.
2. Read the name of the **source** IANA zone in which the instant was originally specified.
3. Read the name of the **target** IANA zone to which the user wants to convert.
4. Validate the calendar fields up front: the month, hour, minute, and second must be in range, and the year-month-day combination must describe a real date. Values such as `2025-02-30`, month 13, and hour 25 are rejected rather than normalized into another instant.
5. Construct the instant's `time.Time` in the source zone.
6. Render the instant in the source zone and in the target zone, showing date, time, and zone abbreviation.
7. If either zone name is unknown to the operating system, report the unknown name in the error, produce no conversion result, and exit cleanly.
8. If the calendar day in the target zone differs from the calendar day in the source zone, show that explicitly in the output.
9. The conversion logic must be reachable from a test that does not depend on `time.Now` or on the host's local zone.
10. The output format must be the same regardless of where the program runs.

## 9. Inputs and Outputs

### Interface Contract

#### Inputs

- Year, month, day, hour, minute, second: integers the user types. Example: `2025`, `1`, `1`, `0`, `0`, `0`.
- Source zone: an IANA name. Example: `Europe/Berlin`.
- Target zone: an IANA name. Example: `Asia/Tokyo`.

#### Outputs

- A line showing the source-zone rendering of the instant: date, clock, zone abbreviation.
- A line showing the target-zone rendering of the same instant: date, clock, zone abbreviation.
- When the dates differ between the two zones, a note that the calendar day changed.

#### Example text-only success run

The instant `2025-01-01 00:00:00` in `Europe/Berlin` is the same instant as `2025-01-01 08:00:00` in `Asia/Tokyo` (Berlin is UTC+1 in January, Tokyo is UTC+9).

```
Source (Europe/Berlin): 2025-01-01 00:00:00 CET
Target (Asia/Tokyo):   2025-01-01 08:00:00 JST
Note: same calendar day.
```

#### Example day-boundary run

The instant `2025-01-01 23:00:00` in `America/Los_Angeles` (UTC-8 in January) corresponds to `2025-01-02 08:00:00` in `Europe/Berlin` (UTC+1 in January) and `2025-01-02 16:00:00` in `Asia/Tokyo` (UTC+9 all year).

```
Source (America/Los_Angeles): 2025-01-01 23:00:00 PST
Target (Asia/Tokyo):          2025-01-02 16:00:00 JST
Note: different calendar day.
```

#### Example unknown-zone error run

```
Source zone: Mars/Olympus
Unknown time zone: Mars/Olympus
```

The same shape applies when the target zone is the one that is unknown; the program reports the unknown name and produces no further output.

#### Example invalid-calendar-field error run

```
Year: 2025
Month: 13
Month must be between 1 and 12.
```

## 10. Rules and Edge Cases

- **Unknown source zone**: reported by name; no conversion happens; the program exits cleanly.
- **Unknown target zone**: reported by name; no conversion happens; the program exits cleanly. The source rendering is *not* printed when the target zone is unknown.
- **Same zone for source and target**: the two lines show identical wall readings; the program does not invent a difference.
- **Day boundary crossing forward**: when the target zone is far enough ahead, the date increases. The program notes this.
- **Day boundary crossing backward**: when the target zone is far enough behind, the date decreases. The program notes this too.
- **Out-of-range calendar fields**: rejected with a clear error before any `time.Time` is constructed.
- **DST transition inputs**: deliberately outside the required scope; the required fixtures stay away from gaps and repeated wall times so the project's conversion contract remains unambiguous.
- **Local-zone reliance forbidden**: the program never calls `time.Local` implicitly; the only zones it considers are the ones the user named.

## 11. Project Constraints

- Go standard library only. No third-party time libraries.
- The conversion logic must take an explicit instant and two zone names; it must not call `time.Now`.
- All tests must construct instants and zones explicitly. A test that depends on the host's local zone or on the current wall clock is out of scope.
- The output format must be deterministic; no locale-specific date formatting.
- No persistence, no scheduling, no alarms — this project only converts a single instant.

## 12. Design Questions Before Coding

- Where will the conversion logic live — inside `main`, in a function that takes the instant and the two zones, or in a method on a struct? Which shape is easiest to test?
- How will you validate the calendar fields? A small explicit check before constructing the `time.Time`, or rely on Go's normalization? Which choice makes the user-visible behavior more honest?
- How will you express "the dates differ" — by comparing the calendar day, the year, or both? What if the year also changes?
- How will the production program pass the two zone names into the conversion? As strings that the function resolves internally, or as already-resolved `time.Location` values? If you split the resolve step from the convert step, the test can resolve stable IANA zones itself and exercise the convert step without touching `time.LoadLocation`.
- Will the output include the zone abbreviation (for example `CET`, `JST`)? How is that abbreviation produced, and what does it look like for the abbreviated and the long names?
- How will an unknown zone be reported — as a value the function returns, as a sentinel error, or as a panic caught at the top level?
- How will a test prove the conversion is correct without depending on `time.Now` or on the host's local zone?

## 13. Implementation Milestones

1. Validate the calendar fields (year, month, day, hour, minute, second) up front, before any `time.Time` is built.
2. Build the `time.Time` for the instant in the source zone.
3. Resolve the source and target zones through `time.LoadLocation`, reporting any unknown name.
4. If either zone is unknown, return or report an error and stop. Do not produce any rendering.
5. Render the instant in each zone using a fixed format string.
6. Compare the calendar days between the two renderings and add a "different calendar day" note when they differ.
7. Make the conversion logic reachable from a test that constructs an instant and zones explicitly, without calling `time.Now` or relying on `time.Local`.
8. Confirm that running the test on any machine produces the same result.

## 14. Verification Cases the Learner Must Write

### Required Cases

Each case is described in natural language. The instant and zones are constructed in the test code; no case calls `time.Now` or relies on `time.Local`.

- A winter instant in `Europe/Berlin` rendered in `Asia/Tokyo` produces the expected date and clock, and the dates are equal.
- An instant near midnight in `America/Los_Angeles` rendered in `Asia/Tokyo` produces the next calendar day in Tokyo, and the program notes the day boundary.
- An instant near midnight in `Asia/Tokyo` rendered in `America/Los_Angeles` produces the previous calendar day in Los Angeles, and the program notes the day boundary.
- Source and target zones are the same; both renderings are identical and no day-boundary note is shown.
- A typo like `Europa/Berlin` (capital `E` instead of lowercase) is reported as an unknown zone and no rendering happens.
- A nonsense zone like `Mars/Olympus` is reported as an unknown zone.
- An invalid calendar field (for example `month = 13`) is rejected before any `time.Time` is constructed; no rendering happens.
- An impossible date such as `2025-02-30` is rejected rather than normalized into March.
- The test suite as a whole does not call `time.Now` anywhere; a learner can search the test file for `time.Now` and find zero matches.

## 15. Common Mistakes to Watch For

- **Confusing the instant with the wall representation.** A common bug is to call `AddDate` or to mutate the time during conversion. The instant must be preserved.
- **Reading the host's local zone implicitly.** The program must use only the zones the user named; `time.Local` is a host-controlled value and a test must not assume it.
- **Calling `time.Now` in production or in tests.** That makes the test non-deterministic. The instant is the input.
- **Silently falling back to UTC when a zone is unknown.** The contract is to report the unknown name; silent fallback hides a user error. The source rendering must not be shown when the target zone is unknown.
- **Letting Go normalize out-of-range calendar fields.** Inputs like `month = 13` will quietly roll over into the next year if the program does not reject them up front; the user then sees a "valid" conversion for an instant they did not type.
- **Comparing instants with `==` after `In(loc)`.** The struct contains a location pointer and possibly a monotonic clock reading; `==` compares those too. Use `Time.Equal` when the comparison is "do these refer to the same instant" and use `==` (or a deliberate field-by-field check) only when the comparison is "are these the same value, location and all". Both are valid for different questions.
- **Forgetting the day-boundary note.** When the dates differ, the program must say so; otherwise the user has to do the date math by hand.

## 16. Topics and References for Study

- A Tour of Go: "Time", "Switch evaluation order".
- Effective Go: "Errors are values".
- Package documentation: `time` (`Time`, `LoadLocation`, `FixedZone`, `In`, `Format`, `Parse`, `Equal`, `Unix`).
- IANA time zone database: search for "IANA time zone database", "tzdata", `tzdata` Go documentation.
- DST mechanics: search for "DST spring forward gap", "DST fall back ambiguity", "DateTime Normalization Go".
- Concepts: search for "instant versus local time", "UTC offset", "daylight saving time", "Unix epoch".
- Test design: search for "deterministic time testing in Go".

## 17. Self-Assessment Questions

1. What is an instant, and how is it different from a wall-clock reading?
2. Why does the same instant produce different dates in `America/Los_Angeles` and `Asia/Tokyo` for some inputs?
3. What does `time.LoadLocation` return for an unknown zone, and how must the program react when *either* zone is unknown?
4. Why must tests for this project avoid `time.Now` and `time.Local`?
5. Where in the code is the day-boundary decision made, and what inputs does it compare?
6. If you stored the instant in a `time.Time` value, what carries the instant and what carries the zone?
7. Why does the program validate the calendar fields itself, instead of relying on Go to roll `month = 13` into January of the next year?
8. What is the difference between `Time.Equal` and `==` on `time.Time`, and when would you pick one over the other?

## 18. Definition of Completion

The project is complete when **all** of the following are true.

- [ ] The README's 19 sections are present in order; this file is the reference.
- [ ] Every functional requirement in section 8 is satisfied.
- [ ] Every verification case in section 14 has a corresponding test, and the tests construct instants and zones explicitly without calling `time.Now` or relying on `time.Local`.
- [ ] A learner can run the test suite on any machine and get the same results.
- [ ] The package documentation explains the distinction between an instant and a local representation, and the "round once at the boundary" rules of section 12.
- [ ] The unknown-zone error path is exercised by tests for *both* the source and the target zone, and in both cases no conversion result is produced.
- [ ] The learner can answer every self-assessment question in section 17 without re-reading the code.

## 19. Optional Extensions

At most two small extensions. Each must be cleanly separated from the required scope.

- **Multiple target zones.** Accept a list of target zones after the first one and render the instant in each. Keep the output format unchanged.

## 20. Prerequisite-Based Documentation Guide

This guide is cumulative: read the formal prerequisite documentation first, then read only the new references listed here. Shared resources are inherited instead of duplicated. Use third-party documentation for the version pinned in Section 4.

### Inherited documentation

- **Formal prerequisite:** [Project 012 — Bill Splitter](../../01-foundations/012_bill_splitter/README.md#20-prerequisite-based-documentation-guide).

Read the linked guide first. Everything introduced there—including documentation inherited from earlier prerequisites—is assumed here and intentionally not repeated.

### New documentation introduced in this project

- **API references:** [`time/tzdata`](https://pkg.go.dev/time/tzdata).
- **Standards and concept references:** [IANA Time Zone Database](https://www.iana.org/time-zones).

### Project-specific learning focus

- **Learn now:** instants versus civil time, UTC offsets, daylight-saving gaps and overlaps, location data, and deterministic clocks.
- **Verification:** Turn every case in Section 14 into a test. Reuse the testing documentation inherited from the prerequisites; if this project introduces a new testing reference, it is listed above.
