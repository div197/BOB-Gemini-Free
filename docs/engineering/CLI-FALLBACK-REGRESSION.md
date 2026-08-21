# CLI Browser Fallback Regression — Mission 4C

The original startup goroutine declared `err` only inside the
`LaunchStudioWindow` branch, handled that error once, and then tested the
outer `err` from configuration loading. That second condition was unrelated
to the browser launch and could execute the default-browser fallback a second
time whenever the outer configuration error was non-nil.

The fix extracts one `launchStudioOrFallback` decision function. Its test
injects a failed app-mode launcher and asserts exactly one fallback invocation;
the production startup path now uses that function and no longer reuses an
out-of-scope error variable.
